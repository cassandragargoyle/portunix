/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package models

import (
	"time"
)

// SchemaVersion is the current trace event schema version
const SchemaVersion = 1

// TraceEvent represents a single trace event in the system
type TraceEvent struct {
	Version   int    `json:"_v"`
	ID        string `json:"id"`
	TraceID   string `json:"trace_id"`
	ParentID  string `json:"parent_id,omitempty"`
	SessionID string `json:"session_id"`

	Timestamp  time.Time `json:"timestamp"`
	DurationUS int64     `json:"duration_us,omitempty"`

	Operation Operation `json:"operation"`
	Input     *DataInfo `json:"input,omitempty"`
	Output    *DataInfo `json:"output,omitempty"`

	Context  map[string]interface{} `json:"context,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Level    Level                  `json:"level"`
	Error    *ErrorInfo             `json:"error,omitempty"`
	Recovery *RecoveryInfo          `json:"recovery,omitempty"`

	Performance *PerformanceInfo  `json:"performance,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Operation describes the operation being traced
type Operation struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
	Version  string `json:"version,omitempty"`
}

// DataInfo contains input or output data information
type DataInfo struct {
	Fields   map[string]interface{} `json:"fields,omitempty"`
	Source   *SourceInfo            `json:"source,omitempty"`
	Checksum string                 `json:"checksum,omitempty"`
	Status   string                 `json:"status,omitempty"`
}

// SourceInfo describes the data source
type SourceInfo struct {
	Type   string `json:"type"`
	File   string `json:"file,omitempty"`
	Row    int    `json:"row,omitempty"`
	Column string `json:"column,omitempty"`
	URL    string `json:"url,omitempty"`
	Table  string `json:"table,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Category   string                 `json:"category,omitempty"`
	Severity   Severity               `json:"severity"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Stack      string                 `json:"stack,omitempty"`
}

// RecoveryInfo describes recovery attempt
type RecoveryInfo struct {
	Attempted bool   `json:"attempted"`
	Strategy  string `json:"strategy"`
	Success   bool   `json:"success"`
}

// PerformanceInfo contains performance metrics
type PerformanceInfo struct {
	CPUUS       int64 `json:"cpu_us,omitempty"`
	MemoryBytes int64 `json:"memory_bytes,omitempty"`
	Allocations int   `json:"allocations,omitempty"`
}

// Level represents the log level
type Level string

const (
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// Severity represents error severity
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// NewTraceEvent creates a new trace event with defaults
func NewTraceEvent(sessionID, traceID, opType, opName string) *TraceEvent {
	return &TraceEvent{
		Version:   SchemaVersion,
		ID:        GenerateEventID(),
		TraceID:   traceID,
		SessionID: sessionID,
		Timestamp: time.Now().UTC(),
		Operation: Operation{
			Type: opType,
			Name: opName,
		},
		Level:    LevelInfo,
		Metadata: make(map[string]string),
	}
}

// SetInput sets the input data
func (e *TraceEvent) SetInput(fields map[string]interface{}) *TraceEvent {
	if e.Input == nil {
		e.Input = &DataInfo{}
	}
	e.Input.Fields = fields
	return e
}

// SetOutput sets the output data with status
func (e *TraceEvent) SetOutput(fields map[string]interface{}, status string) *TraceEvent {
	if e.Output == nil {
		e.Output = &DataInfo{}
	}
	e.Output.Fields = fields
	e.Output.Status = status
	return e
}

// SetSource sets the data source information
func (e *TraceEvent) SetSource(sourceType, file string, row int, column string) *TraceEvent {
	if e.Input == nil {
		e.Input = &DataInfo{}
	}
	e.Input.Source = &SourceInfo{
		Type:   sourceType,
		File:   file,
		Row:    row,
		Column: column,
	}
	return e
}

// SetError sets error information
func (e *TraceEvent) SetError(code, message string, severity Severity) *TraceEvent {
	e.Level = LevelError
	e.Error = &ErrorInfo{
		Code:     code,
		Message:  message,
		Severity: severity,
	}
	return e
}

// AddTag adds a tag to the event
func (e *TraceEvent) AddTag(tag string) *TraceEvent {
	e.Tags = append(e.Tags, tag)
	return e
}

// SetDuration sets the duration in microseconds
func (e *TraceEvent) SetDuration(durationUS int64) *TraceEvent {
	e.DurationUS = durationUS
	return e
}

// SetParent sets the parent event ID
func (e *TraceEvent) SetParent(parentID string) *TraceEvent {
	e.ParentID = parentID
	return e
}
