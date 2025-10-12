package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"portunix.ai/portunix/pkg/logging"
)

// CorrelationID header names
const (
	CorrelationIDHeader    = "X-Correlation-ID"
	RequestIDHeader        = "X-Request-ID"
	TraceIDHeader          = "X-Trace-ID"
)

// CorrelationIDGenerator generates correlation IDs
type CorrelationIDGenerator interface {
	Generate() string
}

// DefaultGenerator is the default correlation ID generator
type DefaultGenerator struct{}

// Generate creates a new correlation ID
func (g *DefaultGenerator) Generate() string {
	// Generate a random 12-byte ID and encode as hex
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// UUIDGenerator generates UUID-like correlation IDs
type UUIDGenerator struct{}

// Generate creates a UUID-like correlation ID
func (g *UUIDGenerator) Generate() string {
	// Generate a simple UUID-like string
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("uuid_%d", time.Now().UnixNano())
	}

	// Format as UUID
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

// TimestampGenerator generates timestamp-based correlation IDs
type TimestampGenerator struct {
	prefix string
}

// NewTimestampGenerator creates a timestamp-based generator
func NewTimestampGenerator(prefix string) *TimestampGenerator {
	return &TimestampGenerator{prefix: prefix}
}

// Generate creates a timestamp-based correlation ID
func (g *TimestampGenerator) Generate() string {
	timestamp := time.Now().UnixNano()
	if g.prefix != "" {
		return fmt.Sprintf("%s_%d", g.prefix, timestamp)
	}
	return fmt.Sprintf("%d", timestamp)
}

// CorrelationMiddleware provides correlation ID middleware for HTTP
type CorrelationMiddleware struct {
	generator   CorrelationIDGenerator
	headerNames []string
	logger      logging.Logger
}

// NewCorrelationMiddleware creates a new correlation middleware
func NewCorrelationMiddleware(generator CorrelationIDGenerator) *CorrelationMiddleware {
	if generator == nil {
		generator = &DefaultGenerator{}
	}

	return &CorrelationMiddleware{
		generator: generator,
		headerNames: []string{
			CorrelationIDHeader,
			RequestIDHeader,
			TraceIDHeader,
		},
		logger: logging.GetLogger("correlation"),
	}
}

// WithHeaderNames sets custom header names to check for correlation ID
func (m *CorrelationMiddleware) WithHeaderNames(headers ...string) *CorrelationMiddleware {
	m.headerNames = headers
	return m
}

// HTTPMiddleware returns an HTTP middleware function
func (m *CorrelationMiddleware) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to extract correlation ID from headers
			correlationID := m.extractCorrelationID(r)

			// Generate new ID if not found
			if correlationID == "" {
				correlationID = m.generator.Generate()
				m.logger.Debug("Generated new correlation ID", "correlation_id", correlationID)
			} else {
				m.logger.Debug("Using existing correlation ID", "correlation_id", correlationID)
			}

			// Add correlation ID to response header
			w.Header().Set(CorrelationIDHeader, correlationID)

			// Add correlation ID to request context
			ctx := logging.WithCorrelationID(r.Context(), correlationID)
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// extractCorrelationID tries to extract correlation ID from request headers
func (m *CorrelationMiddleware) extractCorrelationID(r *http.Request) string {
	for _, headerName := range m.headerNames {
		if id := r.Header.Get(headerName); id != "" {
			return id
		}
	}
	return ""
}

// CorrelationContext provides correlation ID management for non-HTTP contexts
type CorrelationContext struct {
	generator CorrelationIDGenerator
	logger    logging.Logger
}

// NewCorrelationContext creates a new correlation context manager
func NewCorrelationContext(generator CorrelationIDGenerator) *CorrelationContext {
	if generator == nil {
		generator = &DefaultGenerator{}
	}

	return &CorrelationContext{
		generator: generator,
		logger:    logging.GetLogger("correlation"),
	}
}

// NewContext creates a new context with a correlation ID
func (c *CorrelationContext) NewContext(parent context.Context) context.Context {
	correlationID := c.generator.Generate()
	c.logger.Debug("Created new correlation context", "correlation_id", correlationID)
	return logging.WithCorrelationID(parent, correlationID)
}

// NewContextWithID creates a new context with a specific correlation ID
func (c *CorrelationContext) NewContextWithID(parent context.Context, correlationID string) context.Context {
	c.logger.Debug("Created correlation context with ID", "correlation_id", correlationID)
	return logging.WithCorrelationID(parent, correlationID)
}

// PropagateToChild creates a child context that inherits the correlation ID
func (c *CorrelationContext) PropagateToChild(parent context.Context) context.Context {
	if correlationID, exists := logging.CorrelationIDFromContext(parent); exists {
		return logging.WithCorrelationID(context.Background(), correlationID)
	}
	return c.NewContext(context.Background())
}

// CorrelationIDPropagator handles correlation ID propagation for different scenarios
type CorrelationIDPropagator struct {
	context   *CorrelationContext
	separator string
	maxDepth  int
}

// NewCorrelationIDPropagator creates a new propagator
func NewCorrelationIDPropagator(generator CorrelationIDGenerator) *CorrelationIDPropagator {
	return &CorrelationIDPropagator{
		context:   NewCorrelationContext(generator),
		separator: ".",
		maxDepth:  10,
	}
}

// WithSeparator sets the separator for hierarchical correlation IDs
func (p *CorrelationIDPropagator) WithSeparator(sep string) *CorrelationIDPropagator {
	p.separator = sep
	return p
}

// WithMaxDepth sets the maximum depth for hierarchical correlation IDs
func (p *CorrelationIDPropagator) WithMaxDepth(depth int) *CorrelationIDPropagator {
	p.maxDepth = depth
	return p
}

// CreateChild creates a child context with hierarchical correlation ID
func (p *CorrelationIDPropagator) CreateChild(parent context.Context, operation string) context.Context {
	parentID, exists := logging.CorrelationIDFromContext(parent)
	if !exists {
		// No parent correlation ID, create new one
		return p.context.NewContext(parent)
	}

	// Check depth to prevent infinite nesting
	depth := strings.Count(parentID, p.separator)
	if depth >= p.maxDepth {
		// Use parent ID as-is if too deep
		return logging.WithCorrelationID(parent, parentID)
	}

	// Generate child ID
	childSuffix := p.generateChildSuffix()
	childID := fmt.Sprintf("%s%s%s_%s", parentID, p.separator, operation, childSuffix)

	return logging.WithCorrelationID(parent, childID)
}

// generateChildSuffix generates a short suffix for child correlation IDs
func (p *CorrelationIDPropagator) generateChildSuffix() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSpan creates a span context for operation tracing
func (p *CorrelationIDPropagator) CreateSpan(parent context.Context, operation string) (context.Context, func()) {
	spanStart := time.Now()
	childCtx := p.CreateChild(parent, operation)

	logger := logging.LoggerFromContext(childCtx)
	logger.Debug("Span started", "operation", operation, "start_time", spanStart)

	finishFunc := func() {
		duration := time.Since(spanStart)
		logger.Debug("Span finished", "operation", operation, "duration", duration)
	}

	return childCtx, finishFunc
}

// Global correlation manager
var (
	defaultGenerator   CorrelationIDGenerator = &DefaultGenerator{}
	defaultMiddleware  = NewCorrelationMiddleware(defaultGenerator)
	defaultContext     = NewCorrelationContext(defaultGenerator)
	defaultPropagator  = NewCorrelationIDPropagator(defaultGenerator)
)

// Helper functions for easy usage

// HTTPMiddleware returns the default HTTP correlation middleware
func HTTPMiddleware() func(http.Handler) http.Handler {
	return defaultMiddleware.HTTPMiddleware()
}

// NewCorrelationID generates a new correlation ID
func NewCorrelationID() string {
	return defaultGenerator.Generate()
}

// NewContextWithCorrelationID creates a context with a new correlation ID
func NewContextWithCorrelationID(parent context.Context) context.Context {
	return defaultContext.NewContext(parent)
}

// NewContextWithID creates a context with a specific correlation ID
func NewContextWithID(parent context.Context, correlationID string) context.Context {
	return defaultContext.NewContextWithID(parent, correlationID)
}

// CreateChildContext creates a child context with hierarchical correlation ID
func CreateChildContext(parent context.Context, operation string) context.Context {
	return defaultPropagator.CreateChild(parent, operation)
}

// CreateSpan creates a span for operation tracing
func CreateSpan(parent context.Context, operation string) (context.Context, func()) {
	return defaultPropagator.CreateSpan(parent, operation)
}

// SetGlobalGenerator sets the global correlation ID generator
func SetGlobalGenerator(generator CorrelationIDGenerator) {
	defaultGenerator = generator
	defaultMiddleware = NewCorrelationMiddleware(generator)
	defaultContext = NewCorrelationContext(generator)
	defaultPropagator = NewCorrelationIDPropagator(generator)
}