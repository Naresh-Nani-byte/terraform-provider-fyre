// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"
)

// pollingStrategy defines how to poll for operation completion.
type pollingStrategy interface {
	// Next returns the duration to wait before the next poll attempt.
	Next(ctx context.Context) (time.Duration, error)
}

// fixedIntervalStrategy polls at a fixed interval.
type fixedIntervalStrategy struct {
	interval  time.Duration
	startTime time.Time
	attempt   int
}

// adaptiveIntervalStrategy changes polling frequency over time.
type adaptiveIntervalStrategy struct {
	phases    []pollingPhase
	startTime time.Time
	attempt   int
}

// pollingPhase defines a polling interval for a period of time.
type pollingPhase struct {
	Duration time.Duration
	Interval time.Duration
}

// newVMBuildPollingStrategy creates an adaptive strategy for VM build operations.
func newVMBuildPollingStrategy() pollingStrategy {
	return &adaptiveIntervalStrategy{
		phases: []pollingPhase{
			{Duration: 0, Interval: time.Minute},
			{Duration: 3 * time.Minute, Interval: 1 * time.Minute},
			{Duration: 0, Interval: 30 * time.Second},
		},
	}
}

// newFixedIntervalStrategy creates a strategy that polls at a fixed interval.
func newFixedIntervalStrategy(interval time.Duration) pollingStrategy {
	return &fixedIntervalStrategy{
		interval: interval,
	}
}

// Next returns the next fixed polling interval.
func (s *fixedIntervalStrategy) Next(ctx context.Context) (time.Duration, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	if s.interval < 0 {
		return 0, fmt.Errorf("poll interval must not be negative, got %s", s.interval)
	}

	if s.startTime.IsZero() {
		s.startTime = time.Now()
	}

	s.attempt++
	return s.interval, nil
}

// Next returns the next adaptive polling interval.
func (s *adaptiveIntervalStrategy) Next(ctx context.Context) (time.Duration, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	if len(s.phases) == 0 {
		return 0, fmt.Errorf("polling strategy exhausted: no phases configured")
	}

	if s.startTime.IsZero() {
		s.startTime = time.Now()
	}

	elapsed := time.Since(s.startTime)
	var cumulativeDuration time.Duration

	for idx, phase := range s.phases {
		if phase.Duration < 0 {
			return 0, fmt.Errorf("polling strategy exhausted: phase %d has negative duration %s", idx, phase.Duration)
		}
		if phase.Interval < 0 {
			return 0, fmt.Errorf("polling strategy exhausted: phase %d has negative interval %s", idx, phase.Interval)
		}

		if phase.Duration == 0 {
			s.attempt++
			return phase.Interval, nil
		}

		cumulativeDuration += phase.Duration
		if elapsed < cumulativeDuration {
			s.attempt++
			return phase.Interval, nil
		}
	}

	return 0, fmt.Errorf("polling strategy exhausted: no phase found for elapsed time %v (attempt %d)", elapsed, s.attempt)
}
