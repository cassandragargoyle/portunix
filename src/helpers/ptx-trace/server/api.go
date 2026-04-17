/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"portunix.ai/portunix/src/helpers/ptx-trace/index"
	"portunix.ai/portunix/src/helpers/ptx-trace/models"
	"portunix.ai/portunix/src/helpers/ptx-trace/storage"
)

// Server represents the dashboard HTTP server
type Server struct {
	storage *storage.Storage
	index   *index.Index
	port    int
	host    string
	hub     *Hub
}

// NewServer creates a new dashboard server
func NewServer(store *storage.Storage, idx *index.Index, host string, port int) *Server {
	return &Server{
		storage: store,
		index:   idx,
		port:    port,
		host:    host,
		hub:     NewHub(),
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.hub.Run()

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/sessions/", s.handleSession)
	mux.HandleFunc("/api/stats", s.handleStats)

	// WebSocket
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Dashboard (static files)
	mux.HandleFunc("/", s.handleDashboard)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	fmt.Printf("Starting dashboard server at http://%s\n", addr)

	return http.ListenAndServe(addr, s.corsMiddleware(mux))
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleSessions returns list of sessions
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.storage.ListSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, sessions)
}

// handleSession handles /api/sessions/:id and /api/sessions/:id/events etc.
func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	sessionID := parts[0]

	if len(parts) == 1 {
		// DELETE /api/sessions/:id - delete session
		if r.Method == http.MethodDelete {
			s.handleDeleteSession(w, r, sessionID)
			return
		}

		// GET /api/sessions/:id - session detail
		session, err := s.storage.LoadSession(sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		s.jsonResponse(w, session)
		return
	}

	subPath := parts[1]

	switch subPath {
	case "events":
		s.handleSessionEvents(w, r, sessionID)
	case "stats":
		s.handleSessionStats(w, r, sessionID)
	case "errors":
		s.handleSessionErrors(w, r, sessionID)
	case "timeline":
		s.handleSessionTimeline(w, r, sessionID)
	default:
		http.Error(w, "Unknown endpoint", http.StatusNotFound)
	}
}

// handleDeleteSession handles DELETE /api/sessions/:id
func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	// Verify session exists
	_, err := s.storage.LoadSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Delete the session
	err = s.storage.DeleteSession(sessionID)
	if err != nil {
		http.Error(w, "Failed to delete session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast deletion event to WebSocket clients
	s.BroadcastEvent(map[string]interface{}{
		"type":       "session_deleted",
		"session_id": sessionID,
	})

	// Return success
	s.jsonResponse(w, map[string]interface{}{
		"success":    true,
		"session_id": sessionID,
		"message":    "Session deleted successfully",
	})
}

// handleSessionEvents returns events for a session
func (s *Server) handleSessionEvents(w http.ResponseWriter, r *http.Request, sessionID string) {
	// Parse query parameters
	query := r.URL.Query()

	filter := &storage.EventFilter{}

	if op := query.Get("operation"); op != "" {
		filter.Operation = op
	}
	if level := query.Get("level"); level != "" {
		filter.Level = models.Level(level)
	}
	if status := query.Get("status"); status != "" {
		filter.Status = status
	}
	if limit := query.Get("limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &filter.Limit)
	}
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	events, err := s.storage.ReadEvents(sessionID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, events)
}

// handleSessionStats returns stats for a session
func (s *Server) handleSessionStats(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.storage.LoadSession(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	s.jsonResponse(w, session.Stats)
}

// handleSessionErrors returns grouped errors for a session
func (s *Server) handleSessionErrors(w http.ResponseWriter, r *http.Request, sessionID string) {
	// Ensure index is up to date
	s.rebuildIndexIfNeeded(sessionID)

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	groups, err := s.index.GetErrorGroups(sessionID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, groups)
}

// handleSessionTimeline returns timeline data for visualization
func (s *Server) handleSessionTimeline(w http.ResponseWriter, r *http.Request, sessionID string) {
	events, err := s.storage.ReadEvents(sessionID, &storage.EventFilter{Limit: 1000})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Group events by minute for timeline
	timeline := make(map[string]*TimelinePoint)

	for _, event := range events {
		key := event.Timestamp.Format("2006-01-02T15:04")

		if timeline[key] == nil {
			timeline[key] = &TimelinePoint{
				Time:    key,
				Success: 0,
				Error:   0,
				Warning: 0,
			}
		}

		switch event.Level {
		case "error":
			timeline[key].Error++
		case "warning":
			timeline[key].Warning++
		default:
			timeline[key].Success++
		}
	}

	// Convert to sorted slice
	var points []*TimelinePoint
	for _, p := range timeline {
		points = append(points, p)
	}

	s.jsonResponse(w, points)
}

// TimelinePoint represents a point on the timeline
type TimelinePoint struct {
	Time    string `json:"time"`
	Success int    `json:"success"`
	Error   int    `json:"error"`
	Warning int    `json:"warning"`
}

// handleStats returns global statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.storage.ListSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats := &GlobalStats{
		TotalSessions: len(sessions),
		ActiveCount:   0,
		TotalEvents:   0,
		TotalErrors:   0,
	}

	for _, session := range sessions {
		if session.Status == "active" {
			stats.ActiveCount++
		}
		if session.Stats != nil {
			stats.TotalEvents += session.Stats.TotalEvents
			stats.TotalErrors += session.Stats.ByStatus["error"]
		}
	}

	s.jsonResponse(w, stats)
}

// GlobalStats represents global statistics
type GlobalStats struct {
	TotalSessions int   `json:"total_sessions"`
	ActiveCount   int   `json:"active_count"`
	TotalEvents   int64 `json:"total_events"`
	TotalErrors   int64 `json:"total_errors"`
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ServeWs(s.hub, w, r)
}

// handleDashboard serves the dashboard HTML
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		// Serve static assets
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}
	}

	// Disable caching for development
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

// BroadcastEvent sends an event to all connected WebSocket clients
func (s *Server) BroadcastEvent(event interface{}) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	s.hub.Broadcast(data)
}

// jsonResponse sends a JSON response
func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// rebuildIndexIfNeeded rebuilds index if session is not indexed
func (s *Server) rebuildIndexIfNeeded(sessionID string) {
	if sessionID == "" {
		return
	}

	_, err := s.index.GetSessionStats(sessionID)
	if err != nil {
		session, _ := s.storage.LoadSession(sessionID)
		events, _ := s.storage.ReadEvents(sessionID, nil)
		if session != nil && events != nil {
			s.index.RebuildSessionIndex(sessionID, events, session)
		}
	}
}

// parseLevel converts string to models.Level
func parseLevel(s string) models.Level {
	return models.Level(s)
}
