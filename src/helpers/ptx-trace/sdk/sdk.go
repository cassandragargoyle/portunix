/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package sdk

import (
	"sync"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-trace/alerts"
	"portunix.ai/portunix/src/helpers/ptx-trace/models"
	"portunix.ai/portunix/src/helpers/ptx-trace/pii"
	"portunix.ai/portunix/src/helpers/ptx-trace/sampling"
	"portunix.ai/portunix/src/helpers/ptx-trace/storage"
)

// Session represents an active tracing session
type Session struct {
	session *models.Session
	storage *storage.Storage
	traceID string
	masker  *pii.Masker
	sampler *sampling.Sampler
	mu      sync.Mutex

	// Alerting
	alertManager    *alerts.Manager
	alertingEnabled bool
	recentEvents    []*models.TraceEvent // Buffer for alert evaluation
	maxRecentEvents int                  // Max events to keep for evaluation
	firedAlerts     []*alerts.Alert      // Alerts fired during session
}

// SessionOption is a function that configures a session
type SessionOption func(*Session)

// WithSource sets the data source
func WithSource(sourceType string, files ...string) SessionOption {
	return func(s *Session) {
		s.session.SetSource(sourceType, files)
	}
}

// WithDestination sets the data destination
func WithDestination(destType, url, table string) SessionOption {
	return func(s *Session) {
		s.session.SetDestination(destType, url, table)
	}
}

// WithPIIMasking enables PII masking
func WithPIIMasking(enabled bool) SessionOption {
	return func(s *Session) {
		if s.session.Config != nil {
			s.session.Config.PIIMasking = enabled
		}
	}
}

// WithSampling sets the sampling rate (0.0 to 1.0)
func WithSampling(rate float64) SessionOption {
	return func(s *Session) {
		if s.session.Config != nil {
			s.session.Config.SamplingRate = rate
		}
	}
}

// WithTags adds tags to the session
func WithTags(tags ...string) SessionOption {
	return func(s *Session) {
		s.session.Tags = append(s.session.Tags, tags...)
	}
}

// WithAlerting enables real-time alerting during the session
func WithAlerting(enabled bool) SessionOption {
	return func(s *Session) {
		s.alertingEnabled = enabled
	}
}

// WithAlertConfig sets a custom alert configuration path
func WithAlertConfig(configPath string) SessionOption {
	return func(s *Session) {
		s.alertingEnabled = true
		// Config will be loaded during session initialization
		if manager, err := alerts.NewManager(configPath); err == nil {
			s.alertManager = manager
		}
	}
}

// WithAlertManager sets a pre-configured alert manager
func WithAlertManager(manager *alerts.Manager) SessionOption {
	return func(s *Session) {
		s.alertingEnabled = true
		s.alertManager = manager
	}
}

// NewSession creates a new tracing session
func NewSession(name string, opts ...SessionOption) (*Session, error) {
	store, err := storage.NewStorage()
	if err != nil {
		return nil, err
	}

	session := models.NewSession(name)

	s := &Session{
		session:         session,
		storage:         store,
		traceID:         models.GenerateTraceID(),
		masker:          pii.NewMasker(false),     // Disabled by default
		sampler:         sampling.NewSampler(nil), // Default config
		maxRecentEvents: 100,                      // Keep last 100 events for alert evaluation
		recentEvents:    make([]*models.TraceEvent, 0, 100),
		firedAlerts:     make([]*alerts.Alert, 0),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Enable masker if PII masking is configured
	if s.session.Config != nil && s.session.Config.PIIMasking {
		s.masker.Enable()
	}

	// Update sampler rate from session config
	if s.session.Config != nil && s.session.Config.SamplingRate < 1.0 {
		s.sampler.SetDefaultRate(s.session.Config.SamplingRate)
	}

	// Initialize alerting if enabled but no manager provided
	if s.alertingEnabled && s.alertManager == nil {
		if manager, err := alerts.NewManager(""); err == nil {
			s.alertManager = manager
		}
	}

	// Persist session
	if err := store.CreateSession(session); err != nil {
		return nil, err
	}

	return s, nil
}

// LoadSession loads an existing session
func LoadSession(sessionID string) (*Session, error) {
	store, err := storage.NewStorage()
	if err != nil {
		return nil, err
	}

	session, err := store.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Initialize masker and sampler based on session config
	piiEnabled := session.Config != nil && session.Config.PIIMasking
	samplingRate := 1.0
	if session.Config != nil {
		samplingRate = session.Config.SamplingRate
	}

	sampler := sampling.NewSampler(nil)
	sampler.SetDefaultRate(samplingRate)

	return &Session{
		session:         session,
		storage:         store,
		traceID:         models.GenerateTraceID(),
		masker:          pii.NewMasker(piiEnabled),
		sampler:         sampler,
		maxRecentEvents: 100,
		recentEvents:    make([]*models.TraceEvent, 0, 100),
		firedAlerts:     make([]*alerts.Alert, 0),
	}, nil
}

// GetActiveSession returns the currently active session
func GetActiveSession() (*Session, error) {
	store, err := storage.NewStorage()
	if err != nil {
		return nil, err
	}

	session, err := store.GetActiveSession()
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Initialize masker and sampler based on session config
	piiEnabled := session.Config != nil && session.Config.PIIMasking
	samplingRate := 1.0
	if session.Config != nil {
		samplingRate = session.Config.SamplingRate
	}

	sampler := sampling.NewSampler(nil)
	sampler.SetDefaultRate(samplingRate)

	return &Session{
		session:         session,
		storage:         store,
		traceID:         models.GenerateTraceID(),
		masker:          pii.NewMasker(piiEnabled),
		sampler:         sampler,
		maxRecentEvents: 100,
		recentEvents:    make([]*models.TraceEvent, 0, 100),
		firedAlerts:     make([]*alerts.Alert, 0),
	}, nil
}

// ID returns the session ID
func (s *Session) ID() string {
	return s.session.ID
}

// Name returns the session name
func (s *Session) Name() string {
	return s.session.Name
}

// Close ends the session
func (s *Session) Close() error {
	return s.End(models.SessionStatusCompleted)
}

// End ends the session with a specific status
func (s *Session) End(status models.SessionStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.session.End(status)
	return s.storage.UpdateSession(s.session)
}

// Start begins a new traced operation
func (s *Session) Start(operationName string) *Operation {
	return s.StartWithType("transform", operationName)
}

// StartWithType begins a new traced operation with a specific type
func (s *Session) StartWithType(operationType, operationName string) *Operation {
	event := models.NewTraceEvent(s.session.ID, s.traceID, operationType, operationName)

	return &Operation{
		event:     event,
		session:   s,
		startTime: time.Now(),
	}
}

// Trace creates a traced operation with fluent API
func (s *Session) Trace(operationName string) *OperationBuilder {
	return &OperationBuilder{
		session:       s,
		operationName: operationName,
		operationType: "transform",
		tags:          []string{},
	}
}

// writeEvent writes an event and updates session stats
func (s *Session) writeEvent(event *models.TraceEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check sampling - always sample errors, slow ops, and based on rate
	if s.sampler != nil && !s.sampler.ShouldSample(event) {
		// Event sampled out - still update stats but don't write
		s.updateStats(event)
		// Still add to recent events for alerting (even if not persisted)
		s.addToRecentEvents(event)
		// Evaluate alerts even for sampled-out events
		s.evaluateAlertsUnlocked()
		// Persist session to save stats even for sampled-out events
		return s.storage.UpdateSession(s.session)
	}

	// Apply PII masking if enabled
	if s.masker != nil && s.masker.IsEnabled() {
		s.maskEventData(event)
	}

	// Write event to storage
	if err := s.storage.WriteEvent(s.session.ID, event); err != nil {
		return err
	}

	// Update statistics
	s.updateStats(event)

	// Add to recent events buffer for alerting
	s.addToRecentEvents(event)

	// Evaluate alerts in real-time
	s.evaluateAlertsUnlocked()

	// Persist updated session
	return s.storage.UpdateSession(s.session)
}

// updateStats updates session statistics for an event
func (s *Session) updateStats(event *models.TraceEvent) {
	status := "success"
	if event.Output != nil && event.Output.Status != "" {
		status = event.Output.Status
	}
	if event.Level == models.LevelError {
		status = "error"
	}

	s.session.IncrementEventCount(status, event.Level, event.Operation.Name, event.DurationUS)
}

// Operation represents a traced operation
type Operation struct {
	event     *models.TraceEvent
	session   *Session
	startTime time.Time
	parentID  string
}

// Input sets input data
func (o *Operation) Input(key string, value interface{}) *Operation {
	if o.event.Input == nil {
		o.event.Input = &models.DataInfo{Fields: make(map[string]interface{})}
	}
	o.event.Input.Fields[key] = value
	return o
}

// Output sets output data
func (o *Operation) Output(key string, value interface{}) *Operation {
	if o.event.Output == nil {
		o.event.Output = &models.DataInfo{Fields: make(map[string]interface{})}
	}
	o.event.Output.Fields[key] = value
	return o
}

// Source sets the data source
func (o *Operation) Source(source models.SourceInfo) *Operation {
	if o.event.Input == nil {
		o.event.Input = &models.DataInfo{}
	}
	o.event.Input.Source = &source
	return o
}

// Tag adds a tag
func (o *Operation) Tag(tag string) *Operation {
	o.event.Tags = append(o.event.Tags, tag)
	return o
}

// Error records an error
func (o *Operation) Error(err error, severity models.Severity) *Operation {
	o.event.SetError("E_OPERATION_FAILED", err.Error(), severity)
	return o
}

// ErrorWithCode records an error with a specific code
func (o *Operation) ErrorWithCode(code, message string, severity models.Severity) *Operation {
	o.event.SetError(code, message, severity)
	return o
}

// Recovery records a recovery attempt
func (o *Operation) Recovery(strategy string, success bool) *Operation {
	o.event.Recovery = &models.RecoveryInfo{
		Attempted: true,
		Strategy:  strategy,
		Success:   success,
	}
	return o
}

// Success marks the operation as successful
func (o *Operation) Success() *Operation {
	if o.event.Output == nil {
		o.event.Output = &models.DataInfo{}
	}
	o.event.Output.Status = "success"
	return o
}

// Context adds context information
func (o *Operation) Context(key string, value interface{}) *Operation {
	if o.event.Context == nil {
		o.event.Context = make(map[string]interface{})
	}
	o.event.Context[key] = value
	return o
}

// End finalizes and records the operation
func (o *Operation) End() error {
	duration := time.Since(o.startTime)
	o.event.SetDuration(duration.Microseconds())

	if o.parentID != "" {
		o.event.SetParent(o.parentID)
	}

	return o.session.writeEvent(o.event)
}

// Child creates a child operation
func (o *Operation) Child(operationName string) *Operation {
	child := o.session.Start(operationName)
	child.parentID = o.event.ID
	child.event.TraceID = o.event.TraceID
	return child
}

// OperationBuilder provides a fluent API for creating operations
type OperationBuilder struct {
	session       *Session
	operationName string
	operationType string
	tags          []string
	input         map[string]interface{}
	source        *models.SourceInfo
	context       map[string]interface{}
	ruleID        string
	ruleVersion   string
}

// WithType sets the operation type
func (b *OperationBuilder) WithType(opType string) *OperationBuilder {
	b.operationType = opType
	return b
}

// Input sets input data
func (b *OperationBuilder) Input(key string, value interface{}) *OperationBuilder {
	if b.input == nil {
		b.input = make(map[string]interface{})
	}
	b.input[key] = value
	return b
}

// Source sets the data source
func (b *OperationBuilder) Source(source models.SourceInfo) *OperationBuilder {
	b.source = &source
	return b
}

// Tag adds a tag
func (b *OperationBuilder) Tag(tags ...string) *OperationBuilder {
	b.tags = append(b.tags, tags...)
	return b
}

// WithRule sets rule information
func (b *OperationBuilder) WithRule(ruleID, version string) *OperationBuilder {
	b.ruleID = ruleID
	b.ruleVersion = version
	return b
}

// Execute runs the operation with a function
func (b *OperationBuilder) Execute(fn func(ctx Context) error) error {
	op := b.session.StartWithType(b.operationType, b.operationName)

	// Apply builder settings
	for _, tag := range b.tags {
		op.Tag(tag)
	}

	if b.input != nil {
		for k, v := range b.input {
			op.Input(k, v)
		}
	}

	if b.source != nil {
		op.Source(*b.source)
	}

	if b.ruleID != "" {
		op.Context("rule_id", b.ruleID)
		op.Context("rule_version", b.ruleVersion)
	}

	// Create context
	ctx := &operationContext{op: op}

	// Execute function
	err := fn(ctx)
	if err != nil {
		op.Error(err, models.SeverityMedium)
	} else {
		op.Success()
	}

	return op.End()
}

// Context is passed to traced functions
type Context interface {
	Output(key string, value interface{})
	Tag(tag string)
	Context(key string, value interface{})
}

type operationContext struct {
	op *Operation
}

func (c *operationContext) Output(key string, value interface{}) {
	c.op.Output(key, value)
}

func (c *operationContext) Tag(tag string) {
	c.op.Tag(tag)
}

func (c *operationContext) Context(key string, value interface{}) {
	c.op.Context(key, value)
}

// maskEventData applies PII masking to event data
func (s *Session) maskEventData(event *models.TraceEvent) {
	if s.masker == nil {
		return
	}

	// Mask input fields
	if event.Input != nil && event.Input.Fields != nil {
		event.Input.Fields = s.masker.MaskMap(event.Input.Fields)
	}

	// Mask output fields
	if event.Output != nil && event.Output.Fields != nil {
		event.Output.Fields = s.masker.MaskMap(event.Output.Fields)
	}

	// Mask context
	if event.Context != nil {
		event.Context = s.masker.MaskMap(event.Context)
	}

	// Mask error message
	if event.Error != nil {
		event.Error.Message = s.masker.MaskString(event.Error.Message)
	}
}

// addToRecentEvents adds an event to the recent events buffer
// Must be called with lock held
func (s *Session) addToRecentEvents(event *models.TraceEvent) {
	if !s.alertingEnabled || s.alertManager == nil {
		return
	}

	s.recentEvents = append(s.recentEvents, event)

	// Trim buffer if too large
	if len(s.recentEvents) > s.maxRecentEvents {
		s.recentEvents = s.recentEvents[len(s.recentEvents)-s.maxRecentEvents:]
	}
}

// evaluateAlertsUnlocked evaluates alerts against current session state
// Must be called with lock held
func (s *Session) evaluateAlertsUnlocked() {
	if !s.alertingEnabled || s.alertManager == nil {
		return
	}

	// Evaluate alerts against session and recent events
	newAlerts, _ := s.alertManager.ProcessSession(s.session, s.recentEvents)

	// Track fired alerts
	if len(newAlerts) > 0 {
		s.firedAlerts = append(s.firedAlerts, newAlerts...)
	}
}

// EvaluateAlerts manually triggers alert evaluation
func (s *Session) EvaluateAlerts() []*alerts.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.alertingEnabled || s.alertManager == nil {
		return nil
	}

	newAlerts, _ := s.alertManager.ProcessSession(s.session, s.recentEvents)
	if len(newAlerts) > 0 {
		s.firedAlerts = append(s.firedAlerts, newAlerts...)
	}
	return newAlerts
}

// GetFiredAlerts returns all alerts fired during this session
func (s *Session) GetFiredAlerts() []*alerts.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.firedAlerts
}

// GetAlertCount returns the number of alerts fired during this session
func (s *Session) GetAlertCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.firedAlerts)
}

// IsAlertingEnabled returns whether alerting is enabled for this session
func (s *Session) IsAlertingEnabled() bool {
	return s.alertingEnabled
}

// SetAlertManager sets or replaces the alert manager
func (s *Session) SetAlertManager(manager *alerts.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alertManager = manager
	s.alertingEnabled = manager != nil
}

// EnableAlerting enables alerting with default configuration
func (s *Session) EnableAlerting() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.alertManager == nil {
		manager, err := alerts.NewManager("")
		if err != nil {
			return err
		}
		s.alertManager = manager
	}
	s.alertingEnabled = true
	return nil
}

// DisableAlerting disables alerting
func (s *Session) DisableAlerting() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alertingEnabled = false
}
