// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// retryWithBackoff executes a function with exponential backoff retry logic.
// It retries up to maxRetries times for transient errors (5xx, network issues).
// Does NOT retry on 4xx client errors as these indicate permanent request problems.
// Uses exponential backoff: 1s, 2s, 4s between retries.
// The function fn should return an error if the operation failed.
func retryWithBackoff(ctx context.Context, operation string, maxRetries int, fn func() error) error {
	var lastErr error
	var actualAttempts int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		actualAttempts = attempt + 1
		err := fn()

		// Success case
		if err == nil {
			return nil
		}

		// Store error for final error message
		lastErr = errors.Join(lastErr, err)

		// Don't retry on 4xx client errors - these are permanent request problems
		// Check if error message contains "status 4" which indicates a 400-level error
		if strings.Contains(err.Error(), "status 4") {
			tflog.Debug(ctx, "Not retrying 4xx client error", map[string]any{
				"operation": operation,
				"error":     err.Error(),
				"attempts":  actualAttempts,
			})
			return fmt.Errorf("%s: %w", operation, lastErr)
		}

		// Don't retry on last attempt
		if attempt == maxRetries {
			break
		}

		// Exponential backoff: 1s, 2s, 4s
		backoffDuration := time.Duration(1<<uint(attempt)) * time.Second
		tflog.Debug(ctx, "Retrying operation after error", map[string]any{
			"operation":   operation,
			"attempt":     attempt + 1,
			"max_retries": maxRetries,
			"backoff":     backoffDuration.String(),
			"error":       err.Error(),
		})

		select {
		case <-ctx.Done():
			return fmt.Errorf("%s failed: context cancelled during retry backoff", operation)
		case <-time.After(backoffDuration):
			// Continue to next retry
		}
	}

	// All retries exhausted, return wrapped error
	return fmt.Errorf("%s failed after %d attempts: %w", operation, actualAttempts, lastErr)
}
