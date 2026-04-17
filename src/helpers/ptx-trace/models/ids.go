/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateEventID generates a unique event ID
// Format: evt_<uuid without dashes>
func GenerateEventID() string {
	id := uuid.New()
	return "evt_" + strings.ReplaceAll(id.String(), "-", "")[:24]
}

// GenerateTraceID generates a unique trace ID
// Format: trc_<uuid without dashes>
func GenerateTraceID() string {
	id := uuid.New()
	return "trc_" + strings.ReplaceAll(id.String(), "-", "")[:24]
}

// GenerateSessionID generates a session ID based on name and date
// Format: ses_<date>_<sanitized_name>
func GenerateSessionID(name string) string {
	date := time.Now().UTC().Format("2006-01-02")
	sanitized := sanitizeName(name)
	return fmt.Sprintf("ses_%s_%s", date, sanitized)
}

// sanitizeName converts a name to a URL-safe identifier
func sanitizeName(name string) string {
	// Convert to lowercase and replace spaces with dashes
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")

	// Remove any characters that are not alphanumeric or dash
	var cleaned strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}

	// Limit length
	s := cleaned.String()
	if len(s) > 50 {
		s = s[:50]
	}

	return s
}
