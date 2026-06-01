// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp-forge/terraform-provider-fyre/internal/client"
)

// blockedUntilPattern is the regex for parsing the blocked time out of 400
// response details. Hopefully we never hit these because our operation polling
// strategies are correct but in the event we hit a hard block we can attempt
// to wait until we are not blocked.
var blockedUntilPattern = regexp.MustCompile(`blocked at\s+[\d\-:TtZz+\s]+\s+until\s+([\d\-:TtZz+\s]+?)\s+due to`)

// parseBlockedTime extracts the "until" timestamp from blocking error messages.
func parseBlockedTime(details *client.Error_Details) (time.Time, bool) {
	if details == nil {
		return time.Time{}, false
	}

	detailsStr, err := extractDetailsString(details)
	if err != nil {
		return time.Time{}, false
	}

	matches := blockedUntilPattern.FindStringSubmatch(detailsStr)
	if len(matches) != 2 {
		return time.Time{}, false
	}

	untilStr := strings.TrimSpace(matches[1])
	formats := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, parseErr := time.Parse(format, untilStr); parseErr == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

// extractDetailsString converts Error_Details union to a string.
func extractDetailsString(details *client.Error_Details) (string, error) {
	if details == nil {
		return "", fmt.Errorf("details is nil")
	}

	if str, err := details.AsErrorDetails0(); err == nil {
		return string(str), nil
	}

	if obj, err := details.AsErrorDetails1(); err == nil && obj.Errors != nil && len(*obj.Errors) > 0 {
		return strings.Join(*obj.Errors, "\n"), nil
	}

	raw, err := json.Marshal(details)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
