// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"time"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
)

// responseErrorHandler determines retry behavior for HTTP errors.
type responseErrorHandler interface {
	shouldRetry(statusCode int, errorDetails *client.Error, attemptNumber int) (shouldRetry bool, waitDuration time.Duration)
}

// defaultResponseErrorHandler implements standard retry logic.
type defaultResponseErrorHandler struct{}

// shouldRetry returns whether an HTTP error should be retried and how long to wait.
func (h *defaultResponseErrorHandler) shouldRetry(statusCode int, errorDetails *client.Error, attemptNumber int) (bool, time.Duration) {
	switch {
	case statusCode < 300:
		// Don't retry non-errors
		return false, 0
	case statusCode >= 300 && statusCode < 400:
		// Do exponential retries up to three times for redirections
		if attemptNumber < 3 {
			return true, time.Duration(1<<uint(attemptNumber)) * time.Second
		}
		return false, 0
	case statusCode >= 400 && statusCode < 500:
		// If we've received a 400 error we might have a temporary ban. If we are
		// and the error is retryable we'll wait for until our temporary ban has
		// expired.
		if errorDetails != nil && errorDetails.Details != nil {
			if blockUntil, ok := parseBlockedTime(errorDetails.Details); ok {
				waitDuration := time.Until(blockUntil.Add(time.Second))
				if waitDuration > 0 && isRetryable4XXStatusCode(statusCode) {
					return true, waitDuration
				}
			}
		}

		// Handle various 4XX's differently.
		switch {
		case isRetryable4XXStatusCode(statusCode):
			// Retry 429 after 10 minutes. We don't want to retry sooner so that we
			// don't get a temporary ban for too many requests.
			if statusCode == 429 {
				return true, 10 * time.Minute
			}

			// Retry other 400's one more time after a minute.
			if attemptNumber == 0 {
				return true, time.Minute
			}
			return false, 0
		default:
			return false, 0
		}
	// Retry 500's up to three times. Wait a minute in between retries. While that
	// is certainly a long time it beats resources failing because of the
	// occasional service blip.
	case statusCode >= 500 && statusCode < 600:
		if attemptNumber < 3 {
			return true, time.Minute
		}
		return false, 0
	default:
		return false, 0
	}
}

func isRetryable4XXStatusCode(statusCode int) bool {
	switch statusCode {
	case 400, 404, 405, 425, 429:
		return true
	default:
		return false
	}
}
