package export

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"portunix.ai/portunix/src/helpers/ptx-trace/models"
	pb "portunix.ai/portunix/src/helpers/ptx-trace/proto"
	"portunix.ai/portunix/src/helpers/ptx-trace/storage"
)

// FulltextExportOptions configures fulltext export
type FulltextExportOptions struct {
	Host           string // Fulltext plugin host (default: localhost)
	Port           int    // Fulltext plugin port (default: 50051)
	IndexName      string // Target index name
	BatchSize      int    // Number of documents per batch
	IncludeSession bool   // Also index session metadata
	Timeout        int    // Connection timeout in seconds
}

// DefaultFulltextExportOptions returns default options
func DefaultFulltextExportOptions() *FulltextExportOptions {
	return &FulltextExportOptions{
		Host:           "localhost",
		Port:           50051,
		IndexName:      "trace_events",
		BatchSize:      100,
		IncludeSession: true,
		Timeout:        30,
	}
}

// FulltextExporter exports trace data to fulltext search engine
type FulltextExporter struct {
	storage *storage.Storage
}

// NewFulltextExporter creates a new fulltext exporter
func NewFulltextExporter(store *storage.Storage) *FulltextExporter {
	return &FulltextExporter{
		storage: store,
	}
}

// FulltextExportResult contains export statistics
type FulltextExportResult struct {
	SessionID       string
	EventsIndexed   int
	EventsFailed    int
	Duration        time.Duration
	IndexName       string
	FulltextBackend string
}

// Export exports a session to fulltext search engine
func (e *FulltextExporter) Export(sessionID string, opts *FulltextExportOptions) (*FulltextExportResult, error) {
	if opts == nil {
		opts = DefaultFulltextExportOptions()
	}

	startTime := time.Now()

	// Connect to fulltext plugin
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opts.Timeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to fulltext plugin at %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewFulltextServiceClient(conn)

	// Health check
	healthCtx, healthCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer healthCancel()

	healthResp, err := client.HealthCheck(healthCtx, &pb.HealthCheckRequest{})
	if err != nil {
		return nil, fmt.Errorf("fulltext plugin health check failed: %w", err)
	}
	if !healthResp.Healthy {
		return nil, fmt.Errorf("fulltext plugin is not healthy: %s", healthResp.Message)
	}

	// Ensure index exists
	ensureCtx, ensureCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ensureCancel()

	_, err = client.EnsureIndex(ensureCtx, &pb.EnsureIndexRequest{
		IndexName: opts.IndexName,
		Settings: map[string]string{
			"type": "trace_events",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ensure index: %w", err)
	}

	// Load session
	session, err := e.storage.LoadSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Index session metadata if requested
	if opts.IncludeSession {
		if err := e.indexSession(client, session, opts); err != nil {
			return nil, fmt.Errorf("failed to index session: %w", err)
		}
	}

	// Load and index events
	events, err := e.storage.ReadEvents(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	indexed, failed, err := e.indexEvents(client, sessionID, events, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to index events: %w", err)
	}

	return &FulltextExportResult{
		SessionID:       sessionID,
		EventsIndexed:   indexed,
		EventsFailed:    failed,
		Duration:        time.Since(startTime),
		IndexName:       opts.IndexName,
		FulltextBackend: healthResp.Backend,
	}, nil
}

func (e *FulltextExporter) indexSession(client pb.FulltextServiceClient, session *models.Session, opts *FulltextExportOptions) error {
	// Build searchable content from session
	var contentParts []string
	contentParts = append(contentParts, fmt.Sprintf("Session: %s", session.Name))

	if len(session.Tags) > 0 {
		contentParts = append(contentParts, fmt.Sprintf("Tags: %s", strings.Join(session.Tags, ", ")))
	}

	if session.Source != nil {
		contentParts = append(contentParts, fmt.Sprintf("Source: %s", session.Source.Type))
		if len(session.Source.Files) > 0 {
			contentParts = append(contentParts, fmt.Sprintf("Files: %s", strings.Join(session.Source.Files, ", ")))
		}
	}

	if session.Destination != nil {
		contentParts = append(contentParts, fmt.Sprintf("Destination: %s %s", session.Destination.Type, session.Destination.Table))
	}

	// Build metadata
	metadata := map[string]string{
		"type":       "session",
		"session_id": session.ID,
		"name":       session.Name,
		"status":     string(session.Status),
		"started_at": session.StartedAt.Format(time.RFC3339),
	}

	if session.EndedAt != nil {
		metadata["ended_at"] = session.EndedAt.Format(time.RFC3339)
		metadata["duration_ms"] = fmt.Sprintf("%d", session.EndedAt.Sub(session.StartedAt).Milliseconds())
	}

	if len(session.Tags) > 0 {
		metadata["tags"] = strings.Join(session.Tags, ",")
	}

	if session.Stats != nil {
		metadata["total_events"] = fmt.Sprintf("%d", session.Stats.TotalEvents)
		metadata["error_count"] = fmt.Sprintf("%d", session.Stats.ByStatus["error"])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.IndexDocument(ctx, &pb.IndexRequest{
		Id:          fmt.Sprintf("session_%s", session.ID),
		Content:     strings.Join(contentParts, "\n"),
		Metadata:    metadata,
		ContentType: "session",
		IndexName:   opts.IndexName,
	})

	return err
}

func (e *FulltextExporter) indexEvents(client pb.FulltextServiceClient, sessionID string, events []*models.TraceEvent, opts *FulltextExportOptions) (int, int, error) {
	if len(events) == 0 {
		return 0, 0, nil
	}

	totalIndexed := 0
	totalFailed := 0

	// Process in batches
	for i := 0; i < len(events); i += opts.BatchSize {
		end := i + opts.BatchSize
		if end > len(events) {
			end = len(events)
		}

		batch := events[i:end]
		indexed, failed, err := e.indexBatch(client, sessionID, batch, opts)
		if err != nil {
			return totalIndexed, totalFailed, err
		}

		totalIndexed += indexed
		totalFailed += failed
	}

	return totalIndexed, totalFailed, nil
}

func (e *FulltextExporter) indexBatch(client pb.FulltextServiceClient, sessionID string, events []*models.TraceEvent, opts *FulltextExportOptions) (int, int, error) {
	documents := make([]*pb.IndexRequest, 0, len(events))

	for _, event := range events {
		doc := e.eventToDocument(sessionID, event, opts)
		documents = append(documents, doc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.IndexBatch(ctx, &pb.IndexBatchRequest{
		Documents: documents,
		IndexName: opts.IndexName,
	})
	if err != nil {
		return 0, len(events), fmt.Errorf("batch index failed: %w", err)
	}

	return int(resp.IndexedCount), int(resp.FailedCount), nil
}

func (e *FulltextExporter) eventToDocument(sessionID string, event *models.TraceEvent, opts *FulltextExportOptions) *pb.IndexRequest {
	// Build searchable content
	var contentParts []string

	contentParts = append(contentParts, fmt.Sprintf("Operation: %s", event.Operation.Name))
	if event.Operation.Category != "" {
		contentParts = append(contentParts, fmt.Sprintf("Category: %s", event.Operation.Category))
	}

	// Add input fields to content
	if event.Input != nil && len(event.Input.Fields) > 0 {
		for k, v := range event.Input.Fields {
			contentParts = append(contentParts, fmt.Sprintf("Input %s: %v", k, v))
		}
	}

	// Add output fields to content
	if event.Output != nil && len(event.Output.Fields) > 0 {
		for k, v := range event.Output.Fields {
			contentParts = append(contentParts, fmt.Sprintf("Output %s: %v", k, v))
		}
	}

	// Add error info
	if event.Error != nil {
		contentParts = append(contentParts, fmt.Sprintf("Error: %s - %s", event.Error.Code, event.Error.Message))
	}

	// Add tags
	if len(event.Tags) > 0 {
		contentParts = append(contentParts, fmt.Sprintf("Tags: %s", strings.Join(event.Tags, ", ")))
	}

	// Build metadata
	metadata := map[string]string{
		"type":           "event",
		"event_id":       event.ID,
		"session_id":     sessionID,
		"trace_id":       event.TraceID,
		"operation_name": event.Operation.Name,
		"operation_type": event.Operation.Type,
		"level":          string(event.Level),
		"timestamp":      event.Timestamp.Format(time.RFC3339),
	}

	if event.ParentID != "" {
		metadata["parent_id"] = event.ParentID
	}

	if event.Operation.Category != "" {
		metadata["operation_category"] = event.Operation.Category
	}

	if event.DurationUS > 0 {
		metadata["duration_us"] = fmt.Sprintf("%d", event.DurationUS)
	}

	if event.Output != nil && event.Output.Status != "" {
		metadata["status"] = event.Output.Status
	}

	if event.Error != nil {
		metadata["has_error"] = "true"
		metadata["error_code"] = event.Error.Code
		metadata["error_severity"] = string(event.Error.Severity)
	}

	if len(event.Tags) > 0 {
		metadata["tags"] = strings.Join(event.Tags, ",")
	}

	// Add context as JSON
	if len(event.Context) > 0 {
		if contextJSON, err := json.Marshal(event.Context); err == nil {
			metadata["context"] = string(contextJSON)
		}
	}

	return &pb.IndexRequest{
		Id:          event.ID,
		Content:     strings.Join(contentParts, "\n"),
		Metadata:    metadata,
		ContentType: "trace_event",
		IndexName:   opts.IndexName,
	}
}

// Search searches the fulltext index for trace events
func (e *FulltextExporter) Search(query string, opts *FulltextExportOptions) ([]*pb.SearchResult, int64, error) {
	if opts == nil {
		opts = DefaultFulltextExportOptions()
	}

	// Connect to fulltext plugin
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opts.Timeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to connect to fulltext plugin: %w", err)
	}
	defer conn.Close()

	client := pb.NewFulltextServiceClient(conn)

	searchCtx, searchCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer searchCancel()

	resp, err := client.Search(searchCtx, &pb.SearchRequest{
		Query:     query,
		IndexName: opts.IndexName,
		Limit:     100,
		Highlight: true,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search failed: %w", err)
	}

	if !resp.Success {
		return nil, 0, fmt.Errorf("search failed: %s", resp.Message)
	}

	return resp.Results, resp.TotalHits, nil
}

// GetStats returns fulltext index statistics
func (e *FulltextExporter) GetStats(opts *FulltextExportOptions) (*pb.GetStatsResponse, error) {
	if opts == nil {
		opts = DefaultFulltextExportOptions()
	}

	// Connect to fulltext plugin
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opts.Timeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to fulltext plugin: %w", err)
	}
	defer conn.Close()

	client := pb.NewFulltextServiceClient(conn)

	statsCtx, statsCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer statsCancel()

	return client.GetStats(statsCtx, &pb.GetStatsRequest{
		IndexName: opts.IndexName,
	})
}
