/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package export

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-trace/index"
	"portunix.ai/portunix/src/helpers/ptx-trace/models"
	"portunix.ai/portunix/src/helpers/ptx-trace/storage"
)

// AIExportOptions configures the AI export
type AIExportOptions struct {
	Focus         string // "errors", "slow", "all"
	MaxTokens     int    // Approximate token limit
	IncludeSample bool   // Include sample data
	MaxErrors     int    // Max error groups to include
	MaxEvents     int    // Max events to include
}

// DefaultAIExportOptions returns default options
func DefaultAIExportOptions() *AIExportOptions {
	return &AIExportOptions{
		Focus:         "errors",
		MaxTokens:     4000,
		IncludeSample: true,
		MaxErrors:     10,
		MaxEvents:     20,
	}
}

// AIExporter exports trace data in AI-friendly markdown format
type AIExporter struct {
	storage *storage.Storage
	index   *index.Index
}

// NewAIExporter creates a new AI exporter
func NewAIExporter(store *storage.Storage, idx *index.Index) *AIExporter {
	return &AIExporter{
		storage: store,
		index:   idx,
	}
}

// Export generates AI-friendly markdown for a session
func (e *AIExporter) Export(sessionID string, opts *AIExportOptions) (string, error) {
	if opts == nil {
		opts = DefaultAIExportOptions()
	}

	session, err := e.storage.LoadSession(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to load session: %w", err)
	}

	var sb strings.Builder

	// Header
	e.writeHeader(&sb, session)

	// Quick summary
	e.writeSummary(&sb, session)

	// Error analysis (if focus includes errors)
	if opts.Focus == "errors" || opts.Focus == "all" {
		if err := e.writeErrorAnalysis(&sb, sessionID, opts); err != nil {
			return "", err
		}
	}

	// Slow operations (if focus includes slow)
	if opts.Focus == "slow" || opts.Focus == "all" {
		if err := e.writeSlowOperations(&sb, sessionID, opts); err != nil {
			return "", err
		}
	}

	// Operation statistics
	e.writeOperationStats(&sb, session)

	// Sample events (if requested)
	if opts.IncludeSample {
		if err := e.writeSampleEvents(&sb, sessionID, opts); err != nil {
			return "", err
		}
	}

	// Recommendations
	e.writeRecommendations(&sb, session)

	return sb.String(), nil
}

func (e *AIExporter) writeHeader(sb *strings.Builder, session *models.Session) {
	sb.WriteString("# Data Processing Session Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Session ID**: `%s`\n", session.ID))
	sb.WriteString(fmt.Sprintf("**Name**: %s\n", session.Name))
	sb.WriteString(fmt.Sprintf("**Date**: %s\n", session.StartedAt.Format("2006-01-02 15:04:05")))
	if len(session.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags**: %s\n", strings.Join(session.Tags, ", ")))
	}
	sb.WriteString("\n---\n\n")
}

func (e *AIExporter) writeSummary(sb *strings.Builder, session *models.Session) {
	sb.WriteString("## Quick Summary\n\n")

	duration := session.Duration()
	sb.WriteString(fmt.Sprintf("- **Status**: %s\n", e.formatStatus(session.Status)))
	sb.WriteString(fmt.Sprintf("- **Duration**: %s\n", e.formatDuration(duration)))

	if session.Stats != nil {
		total := session.Stats.TotalEvents
		errors := session.Stats.ByStatus["error"]
		warnings := session.Stats.ByStatus["warning"]
		success := session.Stats.ByStatus["success"]

		sb.WriteString(fmt.Sprintf("- **Total Events**: %d\n", total))

		if total > 0 {
			successRate := float64(success) / float64(total) * 100
			sb.WriteString(fmt.Sprintf("- **Success Rate**: %.2f%%\n", successRate))
		}

		if errors > 0 {
			sb.WriteString(fmt.Sprintf("- **Errors**: %d\n", errors))
		}
		if warnings > 0 {
			sb.WriteString(fmt.Sprintf("- **Warnings**: %d\n", warnings))
		}
	}

	sb.WriteString("\n")
}

func (e *AIExporter) writeErrorAnalysis(sb *strings.Builder, sessionID string, opts *AIExportOptions) error {
	groups, err := e.index.GetErrorGroups(sessionID, opts.MaxErrors)
	if err != nil {
		// Index might not exist, try to rebuild
		e.rebuildIndex(sessionID)
		groups, err = e.index.GetErrorGroups(sessionID, opts.MaxErrors)
		if err != nil {
			return nil // Skip error analysis if index fails
		}
	}

	if len(groups) == 0 {
		return nil
	}

	sb.WriteString("## Error Analysis\n\n")

	for i, g := range groups {
		sb.WriteString(fmt.Sprintf("### %d. %s (%d occurrences)\n\n", i+1, g.ErrorCode, g.Count))
		sb.WriteString(fmt.Sprintf("**Message**: %s\n\n", g.ErrorMessage))
		sb.WriteString(fmt.Sprintf("**Severity**: %s\n", g.Severity))
		sb.WriteString(fmt.Sprintf("**Affected Operations**: %s\n", strings.Join(g.Operations, ", ")))
		sb.WriteString(fmt.Sprintf("**First Seen**: %s\n", g.FirstSeen.Format("15:04:05")))
		sb.WriteString(fmt.Sprintf("**Last Seen**: %s\n", g.LastSeen.Format("15:04:05")))

		// Add pattern analysis
		pattern := e.analyzeErrorPattern(g)
		if pattern != "" {
			sb.WriteString(fmt.Sprintf("\n**Pattern**: %s\n", pattern))
		}

		// Add suggestion
		suggestion := e.suggestFix(g)
		if suggestion != "" {
			sb.WriteString(fmt.Sprintf("\n**Suggested Fix**: %s\n", suggestion))
		}

		sb.WriteString("\n")
	}

	return nil
}

func (e *AIExporter) writeSlowOperations(sb *strings.Builder, sessionID string, opts *AIExportOptions) error {
	// Query for slow operations (duration > 1ms)
	filter := &index.QueryFilter{
		MinDuration: 1000, // 1ms in microseconds
		Limit:       10,
	}

	results, err := e.index.QueryEvents(sessionID, filter)
	if err != nil {
		return nil // Skip if query fails
	}

	if len(results) == 0 {
		return nil
	}

	sb.WriteString("## Slow Operations\n\n")
	sb.WriteString("| Operation | Duration | Status | Time |\n")
	sb.WriteString("|-----------|----------|--------|------|\n")

	for _, r := range results {
		durationMs := float64(r.DurationUS) / 1000
		sb.WriteString(fmt.Sprintf("| %s | %.2fms | %s | %s |\n",
			r.OperationName,
			durationMs,
			r.Status,
			r.Timestamp.Format("15:04:05"),
		))
	}

	sb.WriteString("\n")
	return nil
}

func (e *AIExporter) writeOperationStats(sb *strings.Builder, session *models.Session) {
	if session.Stats == nil || len(session.Stats.ByOperation) == 0 {
		return
	}

	sb.WriteString("## Operation Statistics\n\n")
	sb.WriteString("| Operation | Count | Avg Duration | Errors |\n")
	sb.WriteString("|-----------|-------|--------------|--------|\n")

	// Sort operations by count
	type opStat struct {
		name  string
		stats *models.OperationStats
	}
	var ops []opStat
	for name, stats := range session.Stats.ByOperation {
		ops = append(ops, opStat{name, stats})
	}
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].stats.Count > ops[j].stats.Count
	})

	for _, op := range ops {
		avgUs := op.stats.AvgDuration
		avgMs := avgUs / 1000
		var avgStr string
		if avgMs >= 1 {
			avgStr = fmt.Sprintf("%.2fms", avgMs)
		} else {
			avgStr = fmt.Sprintf("%.0fus", avgUs)
		}

		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %d |\n",
			op.name,
			op.stats.Count,
			avgStr,
			op.stats.TotalErrors,
		))
	}

	sb.WriteString("\n")
}

func (e *AIExporter) writeSampleEvents(sb *strings.Builder, sessionID string, opts *AIExportOptions) error {
	// Get sample error events
	filter := &storage.EventFilter{
		Level: models.LevelError,
		Limit: 5,
	}

	events, err := e.storage.ReadEvents(sessionID, filter)
	if err != nil || len(events) == 0 {
		return nil
	}

	sb.WriteString("## Sample Error Events\n\n")

	for i, event := range events {
		sb.WriteString(fmt.Sprintf("### Event %d: %s\n\n", i+1, event.Operation.Name))
		sb.WriteString("```json\n")

		// Write simplified event
		sb.WriteString(fmt.Sprintf("{\n"))
		sb.WriteString(fmt.Sprintf("  \"operation\": \"%s\",\n", event.Operation.Name))
		sb.WriteString(fmt.Sprintf("  \"timestamp\": \"%s\",\n", event.Timestamp.Format(time.RFC3339)))

		if event.Input != nil && len(event.Input.Fields) > 0 {
			sb.WriteString(fmt.Sprintf("  \"input\": %v,\n", e.formatFields(event.Input.Fields)))
		}

		if event.Error != nil {
			sb.WriteString(fmt.Sprintf("  \"error\": {\n"))
			sb.WriteString(fmt.Sprintf("    \"code\": \"%s\",\n", event.Error.Code))
			sb.WriteString(fmt.Sprintf("    \"message\": \"%s\"\n", event.Error.Message))
			sb.WriteString(fmt.Sprintf("  }\n"))
		}

		sb.WriteString("}\n")
		sb.WriteString("```\n\n")
	}

	return nil
}

func (e *AIExporter) writeRecommendations(sb *strings.Builder, session *models.Session) {
	sb.WriteString("## Recommendations\n\n")

	recommendations := e.generateRecommendations(session)
	for i, rec := range recommendations {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}

	if len(recommendations) == 0 {
		sb.WriteString("No specific recommendations. Session completed successfully.\n")
	}

	sb.WriteString("\n---\n")
	sb.WriteString("*Generated by PTX-TRACE for AI analysis*\n")
}

func (e *AIExporter) formatStatus(status models.SessionStatus) string {
	switch status {
	case models.SessionStatusCompleted:
		return "Completed"
	case models.SessionStatusFailed:
		return "Failed"
	case models.SessionStatusCancelled:
		return "Cancelled"
	case models.SessionStatusActive:
		return "Active (in progress)"
	default:
		return string(status)
	}
}

func (e *AIExporter) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func (e *AIExporter) formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return "{}"
	}

	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("\"%s\": \"%v\"", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (e *AIExporter) analyzeErrorPattern(g *index.ErrorGroup) string {
	msg := strings.ToLower(g.ErrorMessage)

	if strings.Contains(msg, "invalid") && strings.Contains(msg, "format") {
		return "Data format validation failures - check input data quality"
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "timed out") {
		return "Timeout issues - consider increasing timeout or optimizing operation"
	}
	if strings.Contains(msg, "connection") || strings.Contains(msg, "refused") {
		return "Connection issues - check network connectivity and service availability"
	}
	if strings.Contains(msg, "permission") || strings.Contains(msg, "denied") {
		return "Permission issues - verify credentials and access rights"
	}
	if strings.Contains(msg, "not found") || strings.Contains(msg, "missing") {
		return "Missing data or resources - verify data completeness"
	}

	return ""
}

func (e *AIExporter) suggestFix(g *index.ErrorGroup) string {
	code := strings.ToLower(g.ErrorCode)

	if strings.Contains(code, "validation") {
		return "Add input validation rules or data cleaning step before processing"
	}
	if strings.Contains(code, "parse") {
		return "Check data format and encoding, add error handling for malformed data"
	}
	if strings.Contains(code, "timeout") {
		return "Increase timeout values or implement retry logic with exponential backoff"
	}
	if strings.Contains(code, "auth") {
		return "Verify authentication credentials and token expiration"
	}

	return ""
}

func (e *AIExporter) generateRecommendations(session *models.Session) []string {
	var recs []string

	if session.Stats == nil {
		return recs
	}

	// High error rate
	if session.Stats.TotalEvents > 0 {
		errorRate := float64(session.Stats.ByStatus["error"]) / float64(session.Stats.TotalEvents)
		if errorRate > 0.1 {
			recs = append(recs, fmt.Sprintf("High error rate (%.1f%%) - investigate root cause and add validation", errorRate*100))
		}
	}

	// Check for slow operations
	for name, stats := range session.Stats.ByOperation {
		if stats.AvgDuration > 10000 { // > 10ms average
			recs = append(recs, fmt.Sprintf("Operation '%s' is slow (avg %.1fms) - consider optimization", name, stats.AvgDuration/1000))
		}
	}

	// Session duration
	if session.Status == models.SessionStatusFailed {
		recs = append(recs, "Session failed - review error logs and implement proper error handling")
	}

	return recs
}

func (e *AIExporter) rebuildIndex(sessionID string) {
	session, err := e.storage.LoadSession(sessionID)
	if err != nil {
		return
	}

	events, err := e.storage.ReadEvents(sessionID, nil)
	if err != nil {
		return
	}

	e.index.RebuildSessionIndex(sessionID, events, session)
}
