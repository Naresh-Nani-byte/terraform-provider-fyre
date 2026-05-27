// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// retryWithBackoff executes a function with configurable retry logic.
func retryWithBackoff(
	ctx context.Context,
	operation string,
	maxRetries int,
	errorHandler responseErrorHandler,
	fn func() (*http.Response, *client.Error, error),
) error {
	if errorHandler == nil {
		errorHandler = &defaultResponseErrorHandler{}
	}

	var lastErr error
	var actualAttempts int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		actualAttempts = attempt + 1
		resp, apiError, err := fn()

		if err == nil && (resp == nil || (resp.StatusCode >= 200 && resp.StatusCode < 300)) {
			return nil
		}

		if err != nil {
			lastErr = errors.Join(lastErr, err)
		} else if resp != nil {
			lastErr = errors.Join(lastErr, fmt.Errorf("status %d", resp.StatusCode))
		}

		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}

		shouldRetry, waitDuration := errorHandler.ShouldRetry(statusCode, apiError, attempt)

		if statusCode >= 400 && statusCode < 500 {
			logFields := map[string]any{
				"operation":   operation,
				"status_code": statusCode,
				"attempt":     actualAttempts,
				"max_retries": maxRetries + 1,
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

			tflog.Warn(ctx, "HTTP 4xx error encountered", logFields)
		}

		if !shouldRetry || attempt == maxRetries {
			break
		}

		tflog.Debug(ctx, "Retrying operation after error", map[string]any{
			"operation":   operation,
			"attempt":     actualAttempts,
			"max_retries": maxRetries + 1,
			"backoff":     waitDuration.String(),
		})

		select {
		case <-ctx.Done():
			return fmt.Errorf("%s failed: reached deadline during retry backoff", operation)
		case <-time.After(waitDuration):
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operation, actualAttempts, lastErr)
}
