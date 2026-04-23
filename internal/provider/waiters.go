// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// pollRequestStatus polls a Fyre API request until completion, failure, or timeout.
func pollRequestStatus(ctx context.Context,
	api *client.ClientWithResponses,
	requestID, site, operation string,
	timeout, pollInterval time.Duration,
) error {
	if requestID == "" {
		return fmt.Errorf("no request_id was provided")
	}

	// Create context with deadline
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	statusSiteParam := client.GetRequestStatusParamsSite(site)
	var statusResp *client.GetRequestStatusResponse

	checkRequestCompleted := func() error {
		var err error

		checkCtx, checkCancel := context.WithDeadline(pollCtx, time.Now().Add(2*time.Second))
		defer checkCancel()
		statusResp, err = api.GetRequestStatusWithResponse(checkCtx, requestID, &client.GetRequestStatusParams{
			Site: &statusSiteParam,
		})
		if err != nil {
			return err
		}

		if statusResp == nil {
			return fmt.Errorf("no status response returned")
		}

		if statusResp.StatusCode() != 200 {
			return fmt.Errorf("waiting for 200, got %d", statusResp.StatusCode())
		}

		if statusResp.JSON200 == nil {
			return fmt.Errorf("no status response payload returned")
		}

		// Handle near-synchronous operations that complete without a completion
		// percentage. If status is "success" but no Request object is returned
		// then we can assume we don't need to wait for 100% completion.
		if statusResp.JSON200.Request == nil {
			if statusResp.JSON200.Status == nil {
				return fmt.Errorf("response missing status field")
			}

			rs := *statusResp.JSON200.Status
			if rs != "success" {
				return fmt.Errorf("request status was not yet successful, got %s", rs)
			}

			tflog.Debug(ctx, "Operation was successful with no job to track to 100%", map[string]any{
				"request_id": requestID,
				"operation":  operation,
				"status":     "success",
			})

			return nil
		}

		if err := checkRequestCompletion(statusResp.JSON200, operation, requestID); err != nil {
			return err
		}

		if statusResp.JSON200.Request.CompletionPercent == nil {
			return fmt.Errorf("no status response completion percentage reported")
		}

		if pct := *statusResp.JSON200.Request.CompletionPercent; pct != 100 {
			return fmt.Errorf("waiting for request completion percentage to be 100, got %d", pct)
		}

		return nil
	}

	// Poll until completion, failure, or timeout
	var err error
	for {
		select {
		case <-pollCtx.Done():
			return fmt.Errorf("%s: timed out after %v waiting for request to complete: %w, %v", operation, timeout, err, spew.Sdump(statusResp))
		case <-ticker.C:
			err = checkRequestCompleted()
			if err == nil {
				return nil
			}

			tflog.Debug(ctx, "poll check for request completion failed", map[string]any{
				"request_id": requestID,
				"error":      err.Error(),
				"response":   statusResp,
			})
			continue
		}
	}
}

// checkRequestCompletion checks if a request has completed or failed.
func checkRequestCompletion(status *client.RequestStatus, operation, requestID string) error {
	if status.Request == nil {
		return nil
	}

	request := status.Request

	// Check if any tasks failed
	if request.Failed != nil && *request.Failed != "0" {
		errorMsg := fmt.Sprintf("%s: request failed with %s failed tasks (request_id: %s)", operation, *request.Failed, requestID)
		if request.LastStatus != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *request.LastStatus)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	// Check for overall error status
	if status.Status != nil && *status.Status == "error" {
		errorMsg := fmt.Sprintf("%s: request failed (request_id: %s)", operation, requestID)
		if request.LastStatus != nil {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *request.LastStatus)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}
