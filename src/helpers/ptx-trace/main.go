package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"portunix.ai/portunix/src/helpers/ptx-trace/alerts"
	"portunix.ai/portunix/src/helpers/ptx-trace/export"
	"portunix.ai/portunix/src/helpers/ptx-trace/index"
	"portunix.ai/portunix/src/helpers/ptx-trace/models"
	"portunix.ai/portunix/src/helpers/ptx-trace/sdk"
	"portunix.ai/portunix/src/helpers/ptx-trace/server"
	"portunix.ai/portunix/src/helpers/ptx-trace/storage"
)

var version = "dev"

// Root command
var rootCmd = &cobra.Command{
	Use:   "portunix",
	Short: "Portunix Trace Helper",
	Long: `Portunix Trace (ptx-trace) is a universal tracing system for software development.
It captures, stores, and visualizes operations during data processing, ETL/ELT pipelines,
API calls, build processes, and any software workflows requiring debugging and analysis.

This helper is invoked by the main portunix dispatcher when running
'portunix trace' commands.`,
	Version: version,
}

// trace command group
var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Universal tracing system for software development",
	Long: `PTX-TRACE is a universal tracing system optimized for debugging,
AI analysis, and monitoring software development workflows.

Features:
  - Hierarchical JSON records with full context
  - Indexed search, filters, and drill-down
  - Export optimized for Claude/GPT analysis
  - Real-time monitoring and timeline
  - Replay mode for reproducibility`,
}

// trace start - create new session
var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a new trace session",
	Long: `Start a new trace session for capturing transformation events.

Examples:
  portunix trace start "import-customers"
  portunix trace start "etl-pipeline" --source data.csv --tag production
  portunix trace start "backup" --pii-mask`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Get flags
		source, _ := cmd.Flags().GetString("source")
		destination, _ := cmd.Flags().GetString("destination")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		sampling, _ := cmd.Flags().GetFloat64("sampling")
		piiMask, _ := cmd.Flags().GetBool("pii-mask")
		alertsEnabled, _ := cmd.Flags().GetBool("alerts")
		alertsConfig, _ := cmd.Flags().GetString("alerts-config")

		// Check for existing active session
		existingSession, _ := sdk.GetActiveSession()
		if existingSession != nil {
			fmt.Fprintf(os.Stderr, "Error: Session '%s' is already active\n", existingSession.Name())
			fmt.Fprintf(os.Stderr, "Use 'portunix trace end' to close it first\n")
			os.Exit(1)
		}

		// Build options
		var opts []sdk.SessionOption

		if source != "" {
			opts = append(opts, sdk.WithSource("file", source))
		}

		if destination != "" {
			parts := strings.SplitN(destination, "://", 2)
			if len(parts) == 2 {
				opts = append(opts, sdk.WithDestination(parts[0], destination, ""))
			}
		}

		if len(tags) > 0 {
			opts = append(opts, sdk.WithTags(tags...))
		}

		if sampling < 1.0 {
			opts = append(opts, sdk.WithSampling(sampling))
		}

		if piiMask {
			opts = append(opts, sdk.WithPIIMasking(true))
		}

		// Enable alerting if requested
		if alertsEnabled {
			if alertsConfig != "" {
				opts = append(opts, sdk.WithAlertConfig(alertsConfig))
			} else {
				opts = append(opts, sdk.WithAlerting(true))
			}
		}

		// Create session
		session, err := sdk.NewSession(name, opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create session: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Session started: %s\n", session.ID())
		fmt.Printf("Name: %s\n", session.Name())
		if session.IsAlertingEnabled() {
			fmt.Println("Alerting: enabled (real-time)")
		}
	},
}

// trace end - end active session
var endCmd = &cobra.Command{
	Use:   "end",
	Short: "End the active trace session",
	Long: `End the currently active trace session.

Examples:
  portunix trace end
  portunix trace end --status failed
  portunix trace end --summary`,
	Run: func(cmd *cobra.Command, args []string) {
		status, _ := cmd.Flags().GetString("status")
		showSummary, _ := cmd.Flags().GetBool("summary")
		showAlerts, _ := cmd.Flags().GetBool("show-alerts")

		session, err := sdk.GetActiveSession()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if session == nil {
			fmt.Fprintf(os.Stderr, "Error: No active session\n")
			os.Exit(1)
		}

		// Get alerts before ending session
		firedAlerts := session.GetFiredAlerts()

		// Determine status
		var sessionStatus models.SessionStatus
		switch status {
		case "failed":
			sessionStatus = models.SessionStatusFailed
		case "cancelled":
			sessionStatus = models.SessionStatusCancelled
		default:
			sessionStatus = models.SessionStatusCompleted
		}

		if err := session.End(sessionStatus); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to end session: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Session ended: %s\n", session.ID())
		fmt.Printf("Status: %s\n", sessionStatus)

		// Show alert summary
		if len(firedAlerts) > 0 {
			fmt.Printf("Alerts fired: %d\n", len(firedAlerts))
		}

		if showSummary {
			// Load full session for stats
			store, _ := storage.NewStorage()
			fullSession, _ := store.LoadSession(session.ID())
			if fullSession != nil && fullSession.Stats != nil {
				fmt.Printf("\nSummary:\n")
				fmt.Printf("  Total events: %d\n", fullSession.Stats.TotalEvents)
				fmt.Printf("  By status: %v\n", fullSession.Stats.ByStatus)
			}
		}

		// Show detailed alert information
		if showAlerts && len(firedAlerts) > 0 {
			fmt.Printf("\nAlerts:\n")
			for i, a := range firedAlerts {
				severityIcon := getSeverityIcon(string(a.Rule.Severity))
				fmt.Printf("  %d. %s %s [%s]\n", i+1, severityIcon, a.Rule.Name, a.Rule.Severity)
				fmt.Printf("     Value: %.2f, Threshold: %.2f\n", a.Value, a.Rule.Condition.Threshold)
				fmt.Printf("     Time: %s\n", a.Timestamp.Format("15:04:05"))
			}
		}
	},
}

// trace event - add event to active session
var eventCmd = &cobra.Command{
	Use:   "event <operation>",
	Short: "Add a trace event to the active session",
	Long: `Add a trace event to the currently active session.

Examples:
  portunix trace event "normalize_phone" --input "phone=+420777123456" --status success
  portunix trace event "validate_email" --input "email=test@example.com" --error "Invalid format"
  portunix trace event "transform" --duration 1500`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := args[0]

		// Get flags
		inputStr, _ := cmd.Flags().GetString("input")
		outputStr, _ := cmd.Flags().GetString("output")
		status, _ := cmd.Flags().GetString("status")
		duration, _ := cmd.Flags().GetInt64("duration")
		errorMsg, _ := cmd.Flags().GetString("error")
		tags, _ := cmd.Flags().GetStringSlice("tag")

		session, err := sdk.GetActiveSession()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if session == nil {
			fmt.Fprintf(os.Stderr, "Error: No active session. Use 'portunix trace start' first.\n")
			os.Exit(1)
		}

		// Create operation
		op := session.Start(operation)

		// Parse and set input
		if inputStr != "" {
			for _, kv := range strings.Split(inputStr, ",") {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					op.Input(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}
		}

		// Parse and set output
		if outputStr != "" {
			for _, kv := range strings.Split(outputStr, ",") {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					op.Output(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}
		}

		// Set tags
		for _, tag := range tags {
			op.Tag(tag)
		}

		// Handle error or success
		if errorMsg != "" {
			op.ErrorWithCode("E_CLI_ERROR", errorMsg, models.SeverityMedium)
		} else if status == "success" || status == "" {
			op.Success()
		}

		// Set duration if provided (otherwise End() calculates it)
		if duration > 0 {
			op.Context("manual_duration", true)
		}

		if err := op.End(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to record event: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Event recorded: %s\n", operation)
	},
}

// trace sessions - list sessions
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List trace sessions",
	Long: `List all trace sessions.

Examples:
  portunix trace sessions
  portunix trace sessions --limit 10
  portunix trace sessions --status active
  portunix trace sessions --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		statusFilter, _ := cmd.Flags().GetString("status")
		format, _ := cmd.Flags().GetString("format")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		sessions, err := store.ListSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Filter by status
		if statusFilter != "" {
			var filtered []*models.Session
			for _, s := range sessions {
				if string(s.Status) == statusFilter {
					filtered = append(filtered, s)
				}
			}
			sessions = filtered
		}

		// Apply limit
		if limit > 0 && len(sessions) > limit {
			sessions = sessions[:limit]
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found")
			return
		}

		if format == "json" {
			data, _ := json.MarshalIndent(sessions, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tEVENTS\tSTARTED\tDURATION")

		for _, s := range sessions {
			duration := s.Duration().Round(time.Second)
			eventCount := int64(0)
			if s.Stats != nil {
				eventCount = s.Stats.TotalEvents
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
				s.ID,
				s.Name,
				s.Status,
				eventCount,
				s.StartedAt.Format("2006-01-02 15:04:05"),
				duration,
			)
		}
		w.Flush()
	},
}

// trace view - view events
var viewCmd = &cobra.Command{
	Use:   "view [session-id]",
	Short: "View trace events",
	Long: `View trace events from a session.

Examples:
  portunix trace view                              # View active session
  portunix trace view ses_2026-01-27_import       # View specific session
  portunix trace view --operation normalize_phone
  portunix trace view --status error --limit 50`,
	Run: func(cmd *cobra.Command, args []string) {
		operation, _ := cmd.Flags().GetString("operation")
		status, _ := cmd.Flags().GetString("status")
		levelStr, _ := cmd.Flags().GetString("level")
		tag, _ := cmd.Flags().GetString("tag")
		limit, _ := cmd.Flags().GetInt("limit")
		format, _ := cmd.Flags().GetString("format")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			// Use active session
			session, _ := store.GetActiveSession()
			if session != nil {
				sessionID = session.ID
			} else {
				// Use most recent session
				sessions, _ := store.ListSessions()
				if len(sessions) > 0 {
					sessionID = sessions[0].ID
				}
			}
		}

		if sessionID == "" {
			fmt.Fprintf(os.Stderr, "Error: No session found\n")
			os.Exit(1)
		}

		// Build filter
		filter := &storage.EventFilter{
			Operation: operation,
			Status:    status,
			Tag:       tag,
			Limit:     limit,
		}

		if levelStr != "" {
			filter.Level = models.Level(levelStr)
		}

		events, err := store.ReadEvents(sessionID, filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(events) == 0 {
			fmt.Println("No events found")
			return
		}

		if format == "json" {
			data, _ := json.MarshalIndent(events, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Pretty output
		for i, e := range events {
			levelIcon := getLevelIcon(e.Level)
			fmt.Printf("%s [%s] %s\n", levelIcon, e.Timestamp.Format("15:04:05.000"), e.Operation.Name)

			if e.Input != nil && len(e.Input.Fields) > 0 {
				fmt.Printf("   Input: %v\n", e.Input.Fields)
			}

			if e.Output != nil && len(e.Output.Fields) > 0 {
				fmt.Printf("   Output: %v\n", e.Output.Fields)
			}

			if e.Error != nil {
				fmt.Printf("   Error: %s - %s\n", e.Error.Code, e.Error.Message)
			}

			if e.DurationUS > 0 {
				fmt.Printf("   Duration: %dμs\n", e.DurationUS)
			}

			if i < len(events)-1 {
				fmt.Println()
			}
		}
	},
}

// trace stats - show session statistics
var statsCmd = &cobra.Command{
	Use:   "stats [session-id]",
	Short: "Show session statistics",
	Long: `Show statistics for a trace session.

Examples:
  portunix trace stats
  portunix trace stats ses_2026-01-27_import
  portunix trace stats --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			// Use active or most recent session
			session, _ := store.GetActiveSession()
			if session != nil {
				sessionID = session.ID
			} else {
				sessions, _ := store.ListSessions()
				if len(sessions) > 0 {
					sessionID = sessions[0].ID
				}
			}
		}

		if sessionID == "" {
			fmt.Fprintf(os.Stderr, "Error: No session found\n")
			os.Exit(1)
		}

		session, err := store.LoadSession(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if format == "json" {
			data, _ := json.MarshalIndent(session, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Pretty output
		fmt.Printf("Session: %s\n", session.ID)
		fmt.Printf("Name: %s\n", session.Name)
		fmt.Printf("Status: %s\n", session.Status)
		fmt.Printf("Started: %s\n", session.StartedAt.Format("2006-01-02 15:04:05"))

		if session.EndedAt != nil {
			fmt.Printf("Ended: %s\n", session.EndedAt.Format("2006-01-02 15:04:05"))
		}

		fmt.Printf("Duration: %s\n", session.Duration().Round(time.Millisecond))

		if len(session.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(session.Tags, ", "))
		}

		if session.Stats != nil {
			fmt.Printf("\nStatistics:\n")
			fmt.Printf("  Total Events: %d\n", session.Stats.TotalEvents)

			if len(session.Stats.ByStatus) > 0 {
				fmt.Printf("  By Status:\n")
				for status, count := range session.Stats.ByStatus {
					fmt.Printf("    %s: %d\n", status, count)
				}
			}

			if len(session.Stats.ByLevel) > 0 {
				fmt.Printf("  By Level:\n")
				for level, count := range session.Stats.ByLevel {
					fmt.Printf("    %s: %d\n", level, count)
				}
			}

			if len(session.Stats.ByOperation) > 0 {
				fmt.Printf("  By Operation:\n")
				for op, stats := range session.Stats.ByOperation {
					fmt.Printf("    %s: %d calls, avg %.2fμs\n", op, stats.Count, stats.AvgDuration)
				}
			}
		}
	},
}

// trace query - SQL-like query on events
var queryCmd = &cobra.Command{
	Use:   "query <sql>",
	Short: "Query events using SQL-like syntax",
	Long: `Query trace events using SQL-like syntax.

Examples:
  portunix trace query "operation = 'validate_email' AND level = 'error'"
  portunix trace query "duration_us > 1000" --session ses_2026-01-27_import
  portunix trace query "error_code = 'E_VALIDATION'" --limit 50`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		queryStr := args[0]
		sessionID, _ := cmd.Flags().GetString("session")
		limit, _ := cmd.Flags().GetInt("limit")
		format, _ := cmd.Flags().GetString("format")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		idx, err := index.NewIndex(store.GetBaseDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to open index: %v\n", err)
			os.Exit(1)
		}
		defer idx.Close()

		// Parse simple query syntax
		filter := parseQueryString(queryStr)
		filter.Limit = limit

		results, err := idx.QueryEvents(sessionID, filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No matching events found")
			return
		}

		if format == "json" {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tOPERATION\tLEVEL\tSTATUS\tDURATION\tERROR")

		for _, r := range results {
			errorInfo := ""
			if r.HasError {
				errorInfo = r.ErrorCode
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%dμs\t%s\n",
				r.Timestamp.Format("15:04:05.000"),
				r.OperationName,
				r.Level,
				r.Status,
				r.DurationUS,
				errorInfo,
			)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d events\n", len(results))
	},
}

// trace errors - show grouped errors
var errorsCmd = &cobra.Command{
	Use:   "errors [session-id]",
	Short: "Show grouped errors with analysis",
	Long: `Show grouped errors for a session with patterns and suggestions.

Examples:
  portunix trace errors
  portunix trace errors ses_2026-01-27_import --limit 20
  portunix trace errors --group`,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		format, _ := cmd.Flags().GetString("format")
		withContext, _ := cmd.Flags().GetBool("with-context")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			session, _ := store.GetActiveSession()
			if session != nil {
				sessionID = session.ID
			} else {
				sessions, _ := store.ListSessions()
				if len(sessions) > 0 {
					sessionID = sessions[0].ID
				}
			}
		}

		idx, err := index.NewIndex(store.GetBaseDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to open index: %v\n", err)
			os.Exit(1)
		}
		defer idx.Close()

		// First, ensure the session is indexed
		rebuildIndexIfNeeded(store, idx, sessionID)

		groups, err := idx.GetErrorGroups(sessionID, limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(groups) == 0 {
			fmt.Println("No errors found")
			return
		}

		if format == "json" {
			data, _ := json.MarshalIndent(groups, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Pretty output
		fmt.Printf("Error Analysis for session: %s\n", sessionID)
		fmt.Printf("Found %d unique error patterns\n\n", len(groups))

		for i, g := range groups {
			severityIcon := getSeverityIcon(g.Severity)
			fmt.Printf("%s Error #%d: %s\n", severityIcon, i+1, g.ErrorCode)
			fmt.Printf("   Message: %s\n", g.ErrorMessage)
			fmt.Printf("   Count: %d occurrences\n", g.Count)
			fmt.Printf("   Operations: %s\n", strings.Join(g.Operations, ", "))
			fmt.Printf("   First seen: %s\n", g.FirstSeen.Format("2006-01-02 15:04:05"))
			fmt.Printf("   Last seen: %s\n", g.LastSeen.Format("2006-01-02 15:04:05"))

			if withContext {
				// Get sample events for this error
				filter := &index.QueryFilter{
					ErrorCode: g.ErrorCode,
					Limit:     3,
				}
				samples, _ := idx.QueryEvents(sessionID, filter)
				if len(samples) > 0 {
					fmt.Printf("   Sample events:\n")
					for _, s := range samples {
						fmt.Printf("     - %s: %s\n", s.Timestamp.Format("15:04:05"), s.OperationName)
					}
				}
			}

			fmt.Println()
		}
	},
}

// trace index - manage index
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage trace index",
	Long:  `Manage the SQLite index for trace sessions and events.`,
}

// trace serve - start dashboard server
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the trace dashboard web server",
	Long: `Start an HTTP server with a web dashboard for viewing and analyzing trace sessions.

The dashboard provides:
  - Real-time event monitoring via WebSocket
  - Session list with statistics
  - Event filtering and search
  - Error analysis and grouping

Examples:
  portunix trace serve
  portunix trace serve --port 8080
  portunix trace serve --host 0.0.0.0 --port 3000`,
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		idx, err := index.NewIndex(store.GetBaseDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to open index: %v\n", err)
			os.Exit(1)
		}
		defer idx.Close()

		srv := server.NewServer(store, idx, host, port)
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// trace index rebuild
var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild [session-id]",
	Short: "Rebuild index for a session",
	Long: `Rebuild the SQLite index for a session from NDJSON files.

Examples:
  portunix trace index rebuild ses_2026-01-27_import
  portunix trace index rebuild --all`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		idx, err := index.NewIndex(store.GetBaseDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to open index: %v\n", err)
			os.Exit(1)
		}
		defer idx.Close()

		if all {
			sessions, err := store.ListSessions()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			for _, session := range sessions {
				fmt.Printf("Rebuilding index for %s...\n", session.ID)
				rebuildSessionIndex(store, idx, session.ID)
			}
			fmt.Printf("Rebuilt index for %d sessions\n", len(sessions))
		} else if len(args) > 0 {
			sessionID := args[0]
			fmt.Printf("Rebuilding index for %s...\n", sessionID)
			rebuildSessionIndex(store, idx, sessionID)
			fmt.Println("Index rebuilt successfully")
		} else {
			fmt.Fprintf(os.Stderr, "Error: Specify session ID or use --all\n")
			os.Exit(1)
		}
	},
}

// trace export - export trace data
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export trace data",
	Long:  `Export trace data in various formats for analysis and integration.`,
}

// trace export ai - export for AI analysis
var exportAICmd = &cobra.Command{
	Use:   "ai [session-id]",
	Short: "Export session for AI analysis (Claude, GPT)",
	Long: `Export a trace session in markdown format optimized for AI analysis.

The export includes:
  - Quick summary with statistics
  - Error analysis with patterns and suggestions
  - Slow operation identification
  - Sample events for context
  - Actionable recommendations

Examples:
  portunix trace export ai
  portunix trace export ai ses_2026-01-27_import
  portunix trace export ai --focus errors --output analysis.md`,
	Run: func(cmd *cobra.Command, args []string) {
		focus, _ := cmd.Flags().GetString("focus")
		output, _ := cmd.Flags().GetString("output")
		maxTokens, _ := cmd.Flags().GetInt("max-tokens")
		includeSample, _ := cmd.Flags().GetBool("include-samples")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		idx, err := index.NewIndex(store.GetBaseDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to open index: %v\n", err)
			os.Exit(1)
		}
		defer idx.Close()

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			sessions, _ := store.ListSessions()
			if len(sessions) > 0 {
				sessionID = sessions[0].ID
			}
		}

		if sessionID == "" {
			fmt.Fprintf(os.Stderr, "Error: No session found\n")
			os.Exit(1)
		}

		exporter := export.NewAIExporter(store, idx)
		opts := &export.AIExportOptions{
			Focus:         focus,
			MaxTokens:     maxTokens,
			IncludeSample: includeSample,
			MaxErrors:     10,
			MaxEvents:     20,
		}

		result, err := exporter.Export(sessionID, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if output != "" {
			if err := os.WriteFile(output, []byte(result), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Exported to %s\n", output)
		} else {
			fmt.Println(result)
		}
	},
}

// trace export file - export to file
var exportFileCmd = &cobra.Command{
	Use:   "file [session-id]",
	Short: "Export session to file (JSON, CSV)",
	Long: `Export a trace session to a file in JSON or CSV format.

Examples:
  portunix trace export file ses_2026-01-27_import --format json
  portunix trace export file --format csv --output events.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		eventsOnly, _ := cmd.Flags().GetBool("events-only")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			sessions, _ := store.ListSessions()
			if len(sessions) > 0 {
				sessionID = sessions[0].ID
			}
		}

		if sessionID == "" {
			fmt.Fprintf(os.Stderr, "Error: No session found\n")
			os.Exit(1)
		}

		events, err := store.ReadEvents(sessionID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var data []byte
		if format == "csv" {
			data = eventsToCSV(events)
		} else {
			if eventsOnly {
				data, _ = json.MarshalIndent(events, "", "  ")
			} else {
				session, _ := store.LoadSession(sessionID)
				exportData := map[string]interface{}{
					"session": session,
					"events":  events,
				}
				data, _ = json.MarshalIndent(exportData, "", "  ")
			}
		}

		if output != "" {
			if err := os.WriteFile(output, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Exported %d events to %s\n", len(events), output)
		} else {
			fmt.Println(string(data))
		}
	},
}

// trace export db - export to database
var exportDBCmd = &cobra.Command{
	Use:   "db [session-id]",
	Short: "Export session to PostgreSQL or MySQL database",
	Long: `Export a trace session to a SQL database (PostgreSQL or MySQL).

The exporter will:
  - Auto-detect database type from connection string
  - Create tables if they don't exist (with --create-table)
  - Support insert, upsert, or replace modes
  - Export both session metadata and events

Connection string formats:
  PostgreSQL: postgres://user:pass@host:5432/dbname
  MySQL:      user:pass@tcp(host:3306)/dbname

Examples:
  portunix trace export db --connection "postgres://user:pass@localhost/traces"
  portunix trace export db ses_2026-01-27_import --connection "postgres://..." --mode upsert
  portunix trace export db --connection "user:pass@tcp(localhost:3306)/traces" --table my_traces`,
	Run: func(cmd *cobra.Command, args []string) {
		connStr, _ := cmd.Flags().GetString("connection")
		table, _ := cmd.Flags().GetString("table")
		mode, _ := cmd.Flags().GetString("mode")
		driver, _ := cmd.Flags().GetString("driver")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		createTable, _ := cmd.Flags().GetBool("create-table")
		noSession, _ := cmd.Flags().GetBool("no-session")

		if connStr == "" {
			fmt.Fprintf(os.Stderr, "Error: --connection is required\n")
			os.Exit(1)
		}

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			sessions, _ := store.ListSessions()
			if len(sessions) > 0 {
				sessionID = sessions[0].ID
			}
		}

		if sessionID == "" {
			fmt.Fprintf(os.Stderr, "Error: No session found\n")
			os.Exit(1)
		}

		exporter := export.NewDatabaseExporter(store)
		opts := &export.DatabaseExportOptions{
			ConnectionString: connStr,
			Driver:           driver,
			Table:            table,
			Mode:             mode,
			BatchSize:        batchSize,
			CreateTable:      createTable,
			IncludeSession:   !noSession,
		}

		fmt.Printf("Exporting session %s to database...\n", sessionID)

		result, err := exporter.Export(sessionID, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Export completed:\n")
		fmt.Printf("  Session: %s\n", result.SessionID)
		fmt.Printf("  Events exported: %d\n", result.EventsExported)
		fmt.Printf("  Table: %s\n", result.Table)
		fmt.Printf("  Driver: %s\n", result.Driver)
		fmt.Printf("  Duration: %s\n", result.Duration.Round(time.Millisecond))
	},
}

// eventsToCSV converts events to CSV format
func eventsToCSV(events []*models.TraceEvent) []byte {
	var sb strings.Builder
	sb.WriteString("timestamp,operation,level,status,duration_us,error_code,error_message\n")

	for _, e := range events {
		status := ""
		if e.Output != nil {
			status = e.Output.Status
		}
		errorCode := ""
		errorMsg := ""
		if e.Error != nil {
			errorCode = e.Error.Code
			errorMsg = strings.ReplaceAll(e.Error.Message, ",", ";")
		}

		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%d,%s,%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Operation.Name,
			e.Level,
			status,
			e.DurationUS,
			errorCode,
			errorMsg,
		))
	}

	return []byte(sb.String())
}

// parseQueryString parses a simple query string into a filter
func parseQueryString(query string) *index.QueryFilter {
	filter := &index.QueryFilter{}

	// Simple parser for key = 'value' AND key2 = 'value2' syntax
	parts := strings.Split(strings.ToLower(query), " and ")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "operation") && strings.Contains(part, "=") {
			filter.Operation = extractValue(part)
		} else if strings.Contains(part, "level") && strings.Contains(part, "=") {
			filter.Level = extractValue(part)
		} else if strings.Contains(part, "status") && strings.Contains(part, "=") {
			filter.Status = extractValue(part)
		} else if strings.Contains(part, "error_code") && strings.Contains(part, "=") {
			filter.ErrorCode = extractValue(part)
			filter.HasError = true
		} else if strings.Contains(part, "has_error") {
			filter.HasError = true
		} else if strings.Contains(part, "tag") && strings.Contains(part, "=") {
			filter.Tag = extractValue(part)
		} else if strings.Contains(part, "duration_us") && strings.Contains(part, ">") {
			// Parse duration_us > N
			val := extractValue(part)
			var dur int64
			fmt.Sscanf(val, "%d", &dur)
			filter.MinDuration = dur
		}
	}

	return filter
}

// extractValue extracts value from 'key = value' or "key = 'value'"
func extractValue(part string) string {
	// Try to find quoted value
	if idx := strings.Index(part, "'"); idx != -1 {
		end := strings.LastIndex(part, "'")
		if end > idx {
			return part[idx+1 : end]
		}
	}

	// Try to find unquoted value after =
	if idx := strings.Index(part, "="); idx != -1 {
		return strings.TrimSpace(part[idx+1:])
	}

	// Try to find value after >
	if idx := strings.Index(part, ">"); idx != -1 {
		return strings.TrimSpace(part[idx+1:])
	}

	return ""
}

// rebuildIndexIfNeeded checks if index needs rebuilding and does it
func rebuildIndexIfNeeded(store *storage.Storage, idx *index.Index, sessionID string) {
	if sessionID == "" {
		return
	}

	// Check if session is in index
	_, err := idx.GetSessionStats(sessionID)
	if err != nil {
		// Session not in index, rebuild
		rebuildSessionIndex(store, idx, sessionID)
	}
}

// rebuildSessionIndex rebuilds the index for a session
func rebuildSessionIndex(store *storage.Storage, idx *index.Index, sessionID string) {
	session, err := store.LoadSession(sessionID)
	if err != nil {
		return
	}

	events, err := store.ReadEvents(sessionID, nil)
	if err != nil {
		return
	}

	idx.RebuildSessionIndex(sessionID, events, session)
}

// getSeverityIcon returns an icon for error severity
func getSeverityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	default:
		return "🟢"
	}
}

// Helper function for level icons
func getLevelIcon(level models.Level) string {
	switch level {
	case models.LevelError:
		return "❌"
	case models.LevelWarning:
		return "⚠️"
	case models.LevelDebug:
		return "🔍"
	default:
		return "✓"
	}
}

// trace alerts - alert management
var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage alerting system",
	Long: `Manage the alerting system for trace sessions.

The alerting system monitors trace sessions and sends notifications
when conditions are met (high error rates, slow operations, etc.).

Examples:
  portunix trace alerts rules        # List configured rules
  portunix trace alerts history      # View alert history
  portunix trace alerts test         # Test alerts against a session`,
}

// trace alerts rules - list rules
var alertsRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List configured alert rules",
	Long: `List all configured alert rules with their conditions and channels.

Examples:
  portunix trace alerts rules
  portunix trace alerts rules --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		configPath, _ := cmd.Flags().GetString("config")

		config, err := alerts.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if !config.Alerts.Enabled {
			fmt.Println("Alerting is disabled")
			return
		}

		if format == "json" {
			data, _ := json.MarshalIndent(config.Alerts.Rules, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCONDITION\tSEVERITY\tWINDOW\tCOOLDOWN\tCHANNELS")

		for _, r := range config.Alerts.Rules {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				r.Name,
				r.Condition,
				r.Severity,
				r.Window,
				r.Cooldown,
				strings.Join(r.Channels, ", "),
			)
		}
		w.Flush()

		fmt.Printf("\nConfigured channels: %d\n", len(config.Alerts.Channels))
		for name, ch := range config.Alerts.Channels {
			fmt.Printf("  - %s (%s)\n", name, ch.Type)
		}
	},
}

// trace alerts history - view alert history
var alertsHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "View alert history",
	Long: `View the history of fired alerts.

Examples:
  portunix trace alerts history
  portunix trace alerts history --limit 20
  portunix trace alerts history --severity critical
  portunix trace alerts history --session ses_2026-01-27_import`,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		severity, _ := cmd.Flags().GetString("severity")
		sessionID, _ := cmd.Flags().GetString("session")
		format, _ := cmd.Flags().GetString("format")

		history, err := alerts.NewAlertHistory("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
			os.Exit(1)
		}

		var records []*alerts.AlertRecord
		if sessionID != "" {
			records = history.GetBySession(sessionID)
		} else if severity != "" {
			records = history.GetBySeverity(alerts.Severity(severity), limit)
		} else {
			records = history.GetRecent(limit)
		}

		if len(records) == 0 {
			fmt.Println("No alerts found")
			return
		}

		if format == "json" {
			data, _ := json.MarshalIndent(records, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tSEVERITY\tRULE\tVALUE\tSESSION")

		for _, r := range records {
			session := r.SessionName
			if session == "" {
				session = r.SessionID
			}
			if len(session) > 20 {
				session = session[:20] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%s\n",
				r.Timestamp.Format("2006-01-02 15:04:05"),
				getSeverityIcon(string(r.Severity))+" "+string(r.Severity),
				r.RuleName,
				r.Value,
				session,
			)
		}
		w.Flush()

		fmt.Printf("\nTotal alerts: %d\n", history.Count())
	},
}

// trace alerts stats - show alert statistics
var alertsStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show alert statistics",
	Long: `Show statistics about fired alerts.

Examples:
  portunix trace alerts stats
  portunix trace alerts stats --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")

		history, err := alerts.NewAlertHistory("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
			os.Exit(1)
		}

		stats := history.GetStats()

		if format == "json" {
			data, _ := json.MarshalIndent(stats, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("Alert Statistics\n")
		fmt.Printf("================\n\n")
		fmt.Printf("Total alerts: %d\n", stats.Total)
		fmt.Printf("Alerts today: %d\n", stats.AlertsToday)
		fmt.Printf("Alerts this week: %d\n", stats.AlertsThisWeek)

		if stats.LastAlert != nil {
			fmt.Printf("Last alert: %s\n", stats.LastAlert.Format("2006-01-02 15:04:05"))
		}

		if len(stats.BySeverity) > 0 {
			fmt.Printf("\nBy Severity:\n")
			for sev, count := range stats.BySeverity {
				fmt.Printf("  %s %s: %d\n", getSeverityIcon(sev), sev, count)
			}
		}

		if len(stats.ByRule) > 0 {
			fmt.Printf("\nBy Rule:\n")
			for rule, count := range stats.ByRule {
				fmt.Printf("  %s: %d\n", rule, count)
			}
		}
	},
}

// trace alerts test - test alerts against session
var alertsTestCmd = &cobra.Command{
	Use:   "test [session-id]",
	Short: "Test alert rules against a session",
	Long: `Test alert rules against a session without sending notifications.

This command evaluates all alert rules against the specified session
and shows which alerts would fire.

Examples:
  portunix trace alerts test
  portunix trace alerts test ses_2026-01-27_import
  portunix trace alerts test --config /path/to/alerts.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			sessions, _ := store.ListSessions()
			if len(sessions) > 0 {
				sessionID = sessions[0].ID
			}
		}

		if sessionID == "" {
			fmt.Fprintf(os.Stderr, "Error: No session found\n")
			os.Exit(1)
		}

		// Load session and events
		session, err := store.LoadSession(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading session: %v\n", err)
			os.Exit(1)
		}

		events, err := store.ReadEvents(sessionID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading events: %v\n", err)
			os.Exit(1)
		}

		// Load config and create evaluator
		config, err := alerts.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		evaluator, err := alerts.NewEvaluator(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating evaluator: %v\n", err)
			os.Exit(1)
		}

		// Reset cooldowns for testing
		evaluator.ResetCooldowns()

		// Evaluate
		firedAlerts := evaluator.EvaluateSession(session, events)

		fmt.Printf("Testing alerts for session: %s\n", sessionID)
		fmt.Printf("Session name: %s\n", session.Name)
		fmt.Printf("Events count: %d\n\n", len(events))

		if len(firedAlerts) == 0 {
			fmt.Println("No alerts would fire for this session.")
			fmt.Println("\nRules evaluated:")
			for _, r := range evaluator.GetRules() {
				fmt.Printf("  - %s: condition not met\n", r.Name)
			}
			return
		}

		fmt.Printf("Alerts that would fire (%d):\n\n", len(firedAlerts))
		for i, a := range firedAlerts {
			severityIcon := getSeverityIcon(string(a.Rule.Severity))
			fmt.Printf("%d. %s %s [%s]\n", i+1, severityIcon, a.Rule.Name, a.Rule.Severity)
			fmt.Printf("   Condition: %s %s %.2f\n", a.Rule.Condition.Type, a.Rule.Condition.Operator, a.Rule.Condition.Threshold)
			fmt.Printf("   Actual value: %.2f\n", a.Value)
			fmt.Printf("   Message: %s\n", a.Message)
			fmt.Printf("   Channels: %s\n\n", strings.Join(a.Rule.Channels, ", "))
		}
	},
}

// trace export fulltext - export to fulltext search engine
var exportFulltextCmd = &cobra.Command{
	Use:   "fulltext [session-id]",
	Short: "Export session to fulltext search engine",
	Long: `Export a trace session to a fulltext search engine via the fulltext plugin.

The fulltext plugin must be running and accessible via gRPC.
Events are indexed with their operation names, input/output fields, errors, and tags.

Examples:
  portunix trace export fulltext
  portunix trace export fulltext ses_2026-01-27_import
  portunix trace export fulltext --host localhost --port 50051
  portunix trace export fulltext --index my-traces`,
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		indexName, _ := cmd.Flags().GetString("index")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		noSession, _ := cmd.Flags().GetBool("no-session")
		timeout, _ := cmd.Flags().GetInt("timeout")

		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Determine session ID
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			sessions, _ := store.ListSessions()
			if len(sessions) > 0 {
				sessionID = sessions[0].ID
			}
		}

		if sessionID == "" {
			fmt.Fprintf(os.Stderr, "Error: No session found\n")
			os.Exit(1)
		}

		exporter := export.NewFulltextExporter(store)
		opts := &export.FulltextExportOptions{
			Host:           host,
			Port:           port,
			IndexName:      indexName,
			BatchSize:      batchSize,
			IncludeSession: !noSession,
			Timeout:        timeout,
		}

		fmt.Printf("Exporting session %s to fulltext search engine...\n", sessionID)
		fmt.Printf("Connecting to %s:%d...\n", host, port)

		result, err := exporter.Export(sessionID, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Export completed:\n")
		fmt.Printf("  Session: %s\n", result.SessionID)
		fmt.Printf("  Events indexed: %d\n", result.EventsIndexed)
		if result.EventsFailed > 0 {
			fmt.Printf("  Events failed: %d\n", result.EventsFailed)
		}
		fmt.Printf("  Index: %s\n", result.IndexName)
		fmt.Printf("  Backend: %s\n", result.FulltextBackend)
		fmt.Printf("  Duration: %s\n", result.Duration.Round(time.Millisecond))
	},
}

// trace alerts clear - clear alert history
var alertsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear alert history",
	Long: `Clear all alert history.

Examples:
  portunix trace alerts clear`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Print("Are you sure you want to clear all alert history? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled")
				return
			}
		}

		history, err := alerts.NewAlertHistory("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
			os.Exit(1)
		}

		if err := history.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing history: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Alert history cleared")
	},
}

func init() {
	// Add trace command to root
	rootCmd.AddCommand(traceCmd)

	// Add subcommands to trace
	traceCmd.AddCommand(startCmd)
	traceCmd.AddCommand(endCmd)
	traceCmd.AddCommand(eventCmd)
	traceCmd.AddCommand(sessionsCmd)
	traceCmd.AddCommand(viewCmd)
	traceCmd.AddCommand(statsCmd)
	traceCmd.AddCommand(queryCmd)
	traceCmd.AddCommand(errorsCmd)
	traceCmd.AddCommand(indexCmd)
	traceCmd.AddCommand(serveCmd)
	traceCmd.AddCommand(exportCmd)

	// Add subcommands to index
	indexCmd.AddCommand(indexRebuildCmd)

	// Add subcommands to export
	exportCmd.AddCommand(exportAICmd)
	exportCmd.AddCommand(exportFileCmd)
	exportCmd.AddCommand(exportDBCmd)
	exportCmd.AddCommand(exportFulltextCmd)

	// Add alerts command
	traceCmd.AddCommand(alertsCmd)

	// Add subcommands to alerts
	alertsCmd.AddCommand(alertsRulesCmd)
	alertsCmd.AddCommand(alertsHistoryCmd)
	alertsCmd.AddCommand(alertsStatsCmd)
	alertsCmd.AddCommand(alertsTestCmd)
	alertsCmd.AddCommand(alertsClearCmd)

	// start command flags
	startCmd.Flags().StringP("source", "s", "", "Source data (file or URL)")
	startCmd.Flags().StringP("destination", "d", "", "Destination (connection string)")
	startCmd.Flags().StringSliceP("tag", "t", []string{}, "Add tag (can be repeated)")
	startCmd.Flags().Float64("sampling", 1.0, "Sampling rate (0.0-1.0)")
	startCmd.Flags().Bool("pii-mask", false, "Enable PII masking")
	startCmd.Flags().Bool("alerts", false, "Enable real-time alerting during session")
	startCmd.Flags().String("alerts-config", "", "Path to custom alert configuration file")

	// end command flags
	endCmd.Flags().String("status", "completed", "Session status: completed, failed, cancelled")
	endCmd.Flags().Bool("summary", false, "Show session summary")
	endCmd.Flags().Bool("show-alerts", false, "Show detailed alert information")

	// event command flags
	eventCmd.Flags().String("input", "", "Input data (key=value,key2=value2)")
	eventCmd.Flags().String("output", "", "Output data (key=value,key2=value2)")
	eventCmd.Flags().String("status", "", "Event status: success, error, warning")
	eventCmd.Flags().Int64("duration", 0, "Duration in microseconds")
	eventCmd.Flags().String("error", "", "Error message")
	eventCmd.Flags().StringSliceP("tag", "t", []string{}, "Add tag")

	// sessions command flags
	sessionsCmd.Flags().IntP("limit", "n", 0, "Limit number of results")
	sessionsCmd.Flags().String("status", "", "Filter by status")
	sessionsCmd.Flags().StringP("format", "f", "table", "Output format: table, json")

	// view command flags
	viewCmd.Flags().StringP("operation", "o", "", "Filter by operation name")
	viewCmd.Flags().String("status", "", "Filter by status")
	viewCmd.Flags().StringP("level", "l", "", "Filter by level: debug, info, warning, error")
	viewCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	viewCmd.Flags().IntP("limit", "n", 100, "Limit number of results")
	viewCmd.Flags().StringP("format", "f", "pretty", "Output format: pretty, json")

	// stats command flags
	statsCmd.Flags().StringP("format", "f", "pretty", "Output format: pretty, json")

	// query command flags
	queryCmd.Flags().StringP("session", "s", "", "Session ID to query")
	queryCmd.Flags().IntP("limit", "n", 100, "Limit number of results")
	queryCmd.Flags().StringP("format", "f", "table", "Output format: table, json")

	// errors command flags
	errorsCmd.Flags().IntP("limit", "n", 20, "Limit number of error groups")
	errorsCmd.Flags().StringP("format", "f", "pretty", "Output format: pretty, json")
	errorsCmd.Flags().Bool("with-context", false, "Include sample events for each error")

	// index rebuild command flags
	indexRebuildCmd.Flags().Bool("all", false, "Rebuild index for all sessions")

	// serve command flags
	serveCmd.Flags().String("host", "localhost", "Host to bind the server")
	serveCmd.Flags().IntP("port", "p", 8765, "Port to run the server")

	// export ai command flags
	exportAICmd.Flags().String("focus", "errors", "Export focus: errors, slow, all")
	exportAICmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	exportAICmd.Flags().Int("max-tokens", 4000, "Approximate token limit")
	exportAICmd.Flags().Bool("include-samples", true, "Include sample events")

	// export file command flags
	exportFileCmd.Flags().StringP("format", "f", "json", "Output format: json, csv")
	exportFileCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	exportFileCmd.Flags().Bool("events-only", false, "Export only events (no session metadata)")

	// export db command flags
	exportDBCmd.Flags().StringP("connection", "c", "", "Database connection string (required)")
	exportDBCmd.Flags().StringP("table", "t", "trace_events", "Target table name")
	exportDBCmd.Flags().StringP("mode", "m", "insert", "Insert mode: insert, upsert, replace")
	exportDBCmd.Flags().String("driver", "", "Database driver: postgres, mysql (auto-detected if empty)")
	exportDBCmd.Flags().Int("batch-size", 100, "Number of records per batch")
	exportDBCmd.Flags().Bool("create-table", true, "Create table if not exists")
	exportDBCmd.Flags().Bool("no-session", false, "Skip session metadata export")

	// export fulltext command flags
	exportFulltextCmd.Flags().String("host", "localhost", "Fulltext plugin host")
	exportFulltextCmd.Flags().IntP("port", "p", 50051, "Fulltext plugin gRPC port")
	exportFulltextCmd.Flags().StringP("index", "i", "trace_events", "Target index name")
	exportFulltextCmd.Flags().Int("batch-size", 100, "Number of documents per batch")
	exportFulltextCmd.Flags().Bool("no-session", false, "Skip session metadata indexing")
	exportFulltextCmd.Flags().Int("timeout", 30, "Connection timeout in seconds")

	// alerts rules command flags
	alertsRulesCmd.Flags().StringP("format", "f", "table", "Output format: table, json")
	alertsRulesCmd.Flags().StringP("config", "c", "", "Path to alert config file")

	// alerts history command flags
	alertsHistoryCmd.Flags().IntP("limit", "n", 50, "Limit number of results")
	alertsHistoryCmd.Flags().String("severity", "", "Filter by severity: critical, high, medium, low")
	alertsHistoryCmd.Flags().StringP("session", "s", "", "Filter by session ID")
	alertsHistoryCmd.Flags().StringP("format", "f", "table", "Output format: table, json")

	// alerts stats command flags
	alertsStatsCmd.Flags().StringP("format", "f", "pretty", "Output format: pretty, json")

	// alerts test command flags
	alertsTestCmd.Flags().StringP("config", "c", "", "Path to alert config file")

	// alerts clear command flags
	alertsClearCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	// Version template
	rootCmd.SetVersionTemplate("portunix trace version {{.Version}}\n")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
