// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// requestPoller handles polling of Fyre API requests with configurable behavior.
type requestPoller struct {
	api          *client.ClientWithResponses
	requestID    string
	site         string
	operation    string
	strategy     pollingStrategy
	errorHandler responseErrorHandler
	initialDelay time.Duration
}

// pollerOption is a functional option for configuring a requestPoller.
type pollerOption func(*requestPoller)

// withInitialDelay sets an initial delay before the first poll attempt.
func withInitialDelay(delay time.Duration) pollerOption {
	return func(p *requestPoller) {
		p.initialDelay = delay
	}
}

// newRequestPoller creates a new request poller with the given configuration.
func newRequestPoller(
	api *client.ClientWithResponses,
	requestID, site, operation string,
	strategy pollingStrategy,
	opts ...pollerOption,
) *requestPoller {
	p := &requestPoller{
		api:          api,
		requestID:    requestID,
		site:         site,
		operation:    operation,
		strategy:     strategy,
		errorHandler: &defaultResponseErrorHandler{},
		initialDelay: 0,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// poll executes the polling loop until completion, failure, or timeout.
func (p *requestPoller) poll(ctx context.Context) error {
	if p.requestID == "" {
		return fmt.Errorf("no request_id was provided")
	}

	if p.strategy == nil {
		return fmt.Errorf("polling strategy is required")
	}

	// Apply initial delay if configured
	if p.initialDelay > 0 {
		tflog.Debug(ctx, "Waiting before first poll attempt", map[string]any{
			"request_id":    p.requestID,
			"operation":     p.operation,
			"initial_delay": p.initialDelay.String(),
		})

		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: context cancelled during initial delay: %w", p.operation, ctx.Err())
		case <-time.After(p.initialDelay):
			// Continue to polling
		}
	}

	statusSiteParam := client.GetRequestStatusParamsSite(p.site)
	var statusResp *client.GetRequestStatusResponse
	startTime := time.Now()
	var lastErr error
	var consecutiveErrors int

	// checkRequestCompleted checks whether or not the request has completed.
	checkRequestCompleted := func() (time.Duration, error) {
		checkCtx, checkCancel := context.WithTimeout(ctx, 10*time.Second)
		defer checkCancel()

		var err error
		statusResp, err = p.api.GetRequestStatusWithResponse(checkCtx, p.requestID, &client.GetRequestStatusParams{
			Site: &statusSiteParam,
		})
		if err != nil {
			return 0, err
		}

		if statusResp == nil {
			return 0, fmt.Errorf("no status response returned")
		}

		statusCode := statusResp.StatusCode()
		if statusCode >= 400 && statusCode < 500 {
			var apiError *client.Error
			switch statusCode {
			case 401:
				apiError = statusResp.JSON401
			case 404:
				apiError = statusResp.JSON404
			}

			shouldRetry, waitDuration := p.errorHandler.ShouldRetry(statusCode, apiError, consecutiveErrors)

			logFields := map[string]any{
				"operation":   p.operation,
				"request_id":  p.requestID,
				"status_code": statusCode,
				"will_retry":  shouldRetry,
			}

			if apiError != nil {
				if apiError.Message != nil {
					logFields["error_message"] = *apiError.Message
				}
				if apiError.Details != nil {
					if detailsStr, detailsErr := extractDetailsString(apiError.Details); detailsErr == nil {
						logFields["error_details"] = detailsStr
					}
				}
			}

			if waitDuration > 0 {
				logFields["wait_duration"] = waitDuration.String()
			}

			tflog.Warn(ctx, "HTTP 4xx error during polling", logFields)

			if !shouldRetry {
				return 0, fmt.Errorf("non-retryable error: status %d", statusCode)
			}

			consecutiveErrors++
			return waitDuration, fmt.Errorf("retryable error: status %d", statusCode)
		}

		consecutiveErrors = 0

		if statusCode != 200 {
			return 0, fmt.Errorf("waiting for 200, got %d", statusCode)
		}

		if statusResp.JSON200 == nil {
			return 0, fmt.Errorf("no status response payload returned")
		}

		if statusResp.JSON200.Request == nil {
			if statusResp.JSON200.Status == nil {
				return 0, fmt.Errorf("response missing status field")
			}

			rs := *statusResp.JSON200.Status
			if rs != "success" {
				return 0, fmt.Errorf("request status was not yet successful, got %s", rs)
			}

			tflog.Debug(ctx, "Operation completed without job tracking", map[string]any{
				"request_id": p.requestID,
				"operation":  p.operation,
				"status":     "success",
			})

			return 0, nil
		}

		if err := checkRequestCompletion(statusResp.JSON200, p.operation, p.requestID); err != nil {
			return 0, err
		}

		if statusResp.JSON200.Request.CompletionPercent == nil {
			return 0, fmt.Errorf("no status response completion percentage reported")
		}

		if pct := *statusResp.JSON200.Request.CompletionPercent; pct != 100 {
			return 0, fmt.Errorf("waiting for request completion percentage to be 100, got %d", pct)
		}

		return 0, nil
	}

	waitForNextAttempt := func(waitDuration time.Duration) error {
		if waitDuration <= 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			if _, ok := ctx.Deadline(); ok {
				return fmt.Errorf("%s: timed out after %v waiting for request to complete: %w", p.operation, time.Since(startTime), lastErr)
			}
			return fmt.Errorf("%s: context cancelled: %w", p.operation, ctx.Err())
		case <-time.After(waitDuration):
			return nil
		}
	}

	// Do our first attempt
	waitDuration, err := checkRequestCompleted()
	if err == nil {
		return nil
	}

	lastErr = err
	tflog.Debug(ctx, "Initial poll check incomplete, starting polling", map[string]any{
		"request_id": p.requestID,
		"operation":  p.operation,
		"error":      err.Error(),
	})

	for {
		// Retry according to either our errorHandler's waitDuration or the polling
		// strategy.
		if waitDuration <= 0 {
			var strategyErr error
			waitDuration, strategyErr = p.strategy.Next(ctx)
			if strategyErr != nil {
				return fmt.Errorf("%s: %w", p.operation, strategyErr)
			}
		}

		if err := waitForNextAttempt(waitDuration); err != nil {
			return err
		}

		waitDuration, lastErr = checkRequestCompleted()
		if lastErr == nil {
			return nil
		}

		tflog.Debug(ctx, "poll check for request completion failed", map[string]any{
			"request_id": p.requestID,
			"operation":  p.operation,
			"elapsed":    time.Since(startTime).String(),
			"error":      lastErr.Error(),
			"response":   statusResp,
		})
	}
}

// checkRequestCompletion checks if a request has completed or failed.
func checkRequestCompletion(status *client.RequestStatus, operation, requestID string) error {
	if status.Request == nil {
		return nil
	}

	request := status.Request

	if request.Failed != nil && *request.Failed != "0" {
		errorMsg := fmt.Sprintf("%s: request failed with %s failed tasks (request_id: %s)", operation, *request.Failed, requestID)
		if request.LastStatus != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *request.LastStatus)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	if status.Status != nil && *status.Status == "error" {
		errorMsg := fmt.Sprintf("%s: request failed (request_id: %s)", operation, requestID)
		if request.LastStatus != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *request.LastStatus)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}
