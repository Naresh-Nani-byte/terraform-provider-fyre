// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
)

func TestFixedIntervalStrategyNext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		strategy := newFixedIntervalStrategy(5 * time.Second)

		for i := range 10 {
			got, err := strategy.next(context.Background())
			require.NoErrorf(t, err, "expected 5 seconds on attempt %d, got: %d", i, got)
			require.Equal(t, 5*time.Second, got)
		}
	})
}

func TestAdaptiveIntervalStrategyNext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		strategy := &adaptiveIntervalStrategy{
			phases: []pollingPhase{
				{Duration: time.Minute, Interval: 10 * time.Second},
				{Duration: 0, Interval: 30 * time.Second},
			},
			startTime: now.Add(-30 * time.Second),
		}

		got, err := strategy.next(context.Background())
		require.NoError(t, err, "expected no error in first phase")
		require.Equal(t, 10*time.Second, got, "expected 10s in first phase")

		strategy.startTime = now.Add(-2 * time.Minute)

		got, err = strategy.next(context.Background())
		require.NoError(t, err, "expected no error in final phase")
		require.Equal(t, 30*time.Second, got, "expected 30s in final phase")
	})
}

func TestNewVMBuildPollingStrategy(t *testing.T) {
	t.Parallel()

	strategy, ok := newVMBuildPollingStrategy().(*adaptiveIntervalStrategy)
	require.True(t, ok, "expected adaptive strategy type")

	require.Len(t, strategy.phases, 3, "expected 3 phases")

	require.Equal(t, time.Minute, strategy.phases[0].Interval, "expected one minute for initial poll")
	require.Equal(t, 3*time.Minute, strategy.phases[1].Duration, "unexpected second phase duration")
	require.Equal(t, time.Minute, strategy.phases[1].Interval, "unexpected second phase interval")
	require.Equal(t, time.Duration(0), strategy.phases[2].Duration, "unexpected third phase duration")
	require.Equal(t, 30*time.Second, strategy.phases[2].Interval, "unexpected third phase interval")
}

func TestParseBlockedTime(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		details   *client.Error_Details
		want      time.Time
		wantFound bool
	}{
		"nil details": {},
		"plain string standard timestamp": {
			details:   mustErrorDetailsString(t, "user alice (id 1) blocked at 2026-05-06 06:54:42 until 2026-05-06 10:54:42 due to too many requests"),
			want:      time.Date(2026, 5, 6, 10, 54, 42, 0, time.UTC),
			wantFound: true,
		},
		"plain string rfc3339 timestamp": {
			details:   mustErrorDetailsString(t, "user alice (id 1) blocked at 2026-05-06T06:54:42Z until 2026-05-06T10:54:42Z due to too many requests"),
			want:      time.Date(2026, 5, 6, 10, 54, 42, 0, time.UTC),
			wantFound: true,
		},
		"plain string invalid fractional timestamp not matched by pattern": {
			details:   mustErrorDetailsString(t, "user alice (id 1) blocked at 2026-05-06T06:54:42.123456789Z until 2026-05-06T10:54:42.987654321Z due to too many requests"),
			wantFound: false,
		},
		"object details with errors array": {
			details:   mustErrorDetailsObject(t, []string{"validation failed", "user alice (id 1) blocked at 2026-05-06 06:54:42 until 2026-05-06 10:54:42 due to too many requests"}),
			want:      time.Date(2026, 5, 6, 10, 54, 42, 0, time.UTC),
			wantFound: true,
		},
		"non blocking message": {
			details:   mustErrorDetailsString(t, "some other validation error"),
			wantFound: false,
		},
		"blocking message missing due to suffix": {
			details:   mustErrorDetailsString(t, "user alice (id 1) blocked at 2026-05-06 06:54:42 until 2026-05-06 10:54:42"),
			wantFound: false,
		},
		"blocking message invalid until timestamp": {
			details:   mustErrorDetailsString(t, "user alice (id 1) blocked at 2026-05-06 06:54:42 until not-a-time due to too many requests"),
			wantFound: false,
		},
		"empty object details": {
			details:   mustErrorDetailsObject(t, nil),
			wantFound: false,
		},
		"raw json fallback does not match pattern": {
			details:   mustRawErrorDetails(t, map[string]any{"message": "blocked elsewhere"}),
			wantFound: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, found := parseBlockedTime(tc.details)
			require.Equal(t, tc.wantFound, found, "unexpected found value")
			if !tc.wantFound {
				require.True(t, got.IsZero(), "expected zero time when not found")
				return
			}
			require.True(t, got.Equal(tc.want), "expected %v, got %v", tc.want, got)
		})
	}
}

func TestExtractDetailsString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		details *client.Error_Details
		want    string
		wantErr bool
	}{
		"nil details": {
			wantErr: true,
		},
		"plain string details": {
			details: mustErrorDetailsString(t, "single error"),
			want:    "single error",
		},
		"object details with errors": {
			details: mustErrorDetailsObject(t, []string{"first", "second"}),
			want:    "first\nsecond",
		},
		"object details without errors falls back to raw json": {
			details: mustRawErrorDetails(t, map[string]any{"foo": "bar"}),
			want:    `{"foo":"bar"}`,
		},
		"object details with empty errors falls back to raw json": {
			details: mustRawErrorDetails(t, map[string]any{"errors": []string{}}),
			want:    `{"errors":[]}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := extractDetailsString(tc.details)
			if tc.wantErr {
				require.Error(t, err, "expected error")
				return
			}
			require.NoError(t, err, "expected no error")
			require.Equal(t, tc.want, got, "unexpected details string")
		})
	}
}

func TestDefaultResponseErrorHandlerShouldRetry(t *testing.T) {
	t.Parallel()

	handler := &defaultResponseErrorHandler{}

	tests := []struct {
		name        string
		statusCode  int
		attempt     int
		shouldRetry bool
		wait        time.Duration
	}{
		{name: "400 first attempt", statusCode: 400, attempt: 0, shouldRetry: true, wait: time.Minute},
		{name: "400 second attempt", statusCode: 400, attempt: 1, shouldRetry: false, wait: 0},
		{name: "404 first attempt", statusCode: 404, attempt: 0, shouldRetry: true, wait: time.Minute},
		{name: "405 first attempt", statusCode: 405, attempt: 0, shouldRetry: true, wait: time.Minute},
		{name: "425 first attempt", statusCode: 425, attempt: 0, shouldRetry: true, wait: time.Minute},
		{name: "429", statusCode: 429, attempt: 0, shouldRetry: true, wait: 10 * time.Minute},
		{name: "403", statusCode: 403, attempt: 0, shouldRetry: false, wait: 0},
		{name: "500 first attempt", statusCode: 500, attempt: 0, shouldRetry: true, wait: time.Minute},
		{name: "500 fourth attempt", statusCode: 500, attempt: 3, shouldRetry: false, wait: 0},
		{name: "network error", statusCode: 0, attempt: 2, shouldRetry: false, wait: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shouldRetry, wait := handler.shouldRetry(tc.statusCode, nil, tc.attempt)
			require.Equal(t, tc.shouldRetry, shouldRetry, "unexpected shouldRetry value")
			require.Equal(t, tc.wait, wait, "unexpected wait duration")
		})
	}
}

func TestDefaultResponseErrorHandlerBlockingOverride(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		handler := &defaultResponseErrorHandler{}

		until := time.Now().Add(2 * time.Minute).UTC().Format("2006-01-02 15:04:05")
		var details client.Error_Details
		if err := details.FromErrorDetails0("user alice (id 1) blocked at 2026-05-06 06:54:42 until " + until + " due to too many requests"); err != nil {
			t.Fatalf("failed to build details: %v", err)
		}

		apiErr := &client.Error{Details: &details}
		shouldRetry, wait := handler.shouldRetry(400, apiErr, 0)
		require.True(t, shouldRetry, "expected retry for blocking 400")
		require.GreaterOrEqual(t, wait, time.Minute, "expected blocking wait >= 1m")
		require.LessOrEqual(t, wait, 3*time.Minute, "expected blocking wait <= 3m")

		shouldRetry, wait = handler.shouldRetry(403, apiErr, 0)
		require.False(t, shouldRetry, "expected no retry for non-retryable status")
		require.Equal(t, time.Duration(0), wait, "expected zero wait for non-retryable status")
	})
}

func TestRetryWithBackoffRetriesRetryable400(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := retryWithBackoff(ctx, "test operation", 1, nil, func() (*http.Response, *client.Error, error) {
			attempts++
			if attempts == 1 {
				return &http.Response{StatusCode: 400}, nil, assertiveError("status 400")
			}
			return &http.Response{StatusCode: 200}, nil, nil
		})
		require.NoError(t, err, "expected success after retry")
		require.Equal(t, 2, attempts, "expected 2 attempts")
	})
}

func mustErrorDetailsString(t *testing.T, value string) *client.Error_Details {
	t.Helper()

	var details client.Error_Details
	if err := details.FromErrorDetails0(value); err != nil {
		t.Fatalf("failed to build string details: %v", err)
	}

	return &details
}

func mustErrorDetailsObject(t *testing.T, errors []string) *client.Error_Details {
	t.Helper()

	var details client.Error_Details
	payload := client.ErrorDetails1{}
	if errors != nil {
		payload.Errors = &errors
	}
	if err := details.FromErrorDetails1(payload); err != nil {
		t.Fatalf("failed to build object details: %v", err)
	}

	return &details
}

func mustRawErrorDetails(t *testing.T, value any) *client.Error_Details {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal raw details: %v", err)
	}

	var details client.Error_Details
	if err := details.UnmarshalJSON(raw); err != nil {
		t.Fatalf("failed to unmarshal raw details: %v", err)
	}

	return &details
}

type assertiveError string

func (e assertiveError) Error() string {
	return string(e)
}
