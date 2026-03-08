package sampling

import (
	"math/rand"
	"sync"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-trace/models"
)

// Sampler determines which events should be recorded
type Sampler struct {
	config *Config
	rng    *rand.Rand
	mu     sync.Mutex
}

// Config defines sampling configuration
type Config struct {
	// DefaultRate is the default sampling rate (0.0 to 1.0)
	DefaultRate float64

	// AlwaysLogErrors ensures all error events are logged (rate = 1.0)
	AlwaysLogErrors bool

	// AlwaysLogSlow ensures all slow operations are logged
	AlwaysLogSlow bool

	// SlowThresholdUS defines the threshold for slow operations in microseconds
	SlowThresholdUS int64

	// Rules defines custom sampling rules
	Rules []Rule
}

// Rule defines a conditional sampling rule
type Rule struct {
	// Condition to match
	Condition RuleCondition

	// Rate to apply when condition matches (0.0 to 1.0)
	Rate float64
}

// RuleCondition defines conditions for a rule
type RuleCondition struct {
	// Level matches specific log levels
	Level models.Level

	// Operation matches specific operation names
	Operation string

	// HasTag matches events with specific tags
	HasTag string

	// MinDuration matches events with duration >= this value (microseconds)
	MinDuration int64
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultRate:     1.0, // Log everything by default
		AlwaysLogErrors: true,
		AlwaysLogSlow:   true,
		SlowThresholdUS: 10000, // 10ms
		Rules:           []Rule{},
	}
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig() *Config {
	return &Config{
		DefaultRate:     0.1, // 10% sampling
		AlwaysLogErrors: true,
		AlwaysLogSlow:   true,
		SlowThresholdUS: 5000, // 5ms
		Rules: []Rule{
			// Always log warnings
			{
				Condition: RuleCondition{Level: models.LevelWarning},
				Rate:      1.0,
			},
			// Always log debug tagged events
			{
				Condition: RuleCondition{HasTag: "debug"},
				Rate:      1.0,
			},
		},
	}
}

// NewSampler creates a new sampler with the given configuration
func NewSampler(config *Config) *Sampler {
	if config == nil {
		config = DefaultConfig()
	}

	return &Sampler{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample determines if an event should be recorded
func (s *Sampler) ShouldSample(event *models.TraceEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Always log errors if configured
	if s.config.AlwaysLogErrors && event.Level == models.LevelError {
		return true
	}

	// Always log slow operations if configured
	if s.config.AlwaysLogSlow && event.DurationUS >= s.config.SlowThresholdUS {
		return true
	}

	// Check custom rules
	for _, rule := range s.config.Rules {
		if s.matchesCondition(event, rule.Condition) {
			return s.shouldSampleWithRate(rule.Rate)
		}
	}

	// Apply default rate
	return s.shouldSampleWithRate(s.config.DefaultRate)
}

// matchesCondition checks if an event matches a rule condition
func (s *Sampler) matchesCondition(event *models.TraceEvent, cond RuleCondition) bool {
	// Check level
	if cond.Level != "" && event.Level != cond.Level {
		return false
	}

	// Check operation
	if cond.Operation != "" && event.Operation.Name != cond.Operation {
		return false
	}

	// Check tag
	if cond.HasTag != "" {
		found := false
		for _, tag := range event.Tags {
			if tag == cond.HasTag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check duration
	if cond.MinDuration > 0 && event.DurationUS < cond.MinDuration {
		return false
	}

	return true
}

// shouldSampleWithRate returns true based on probability
func (s *Sampler) shouldSampleWithRate(rate float64) bool {
	if rate >= 1.0 {
		return true
	}
	if rate <= 0.0 {
		return false
	}
	return s.rng.Float64() < rate
}

// GetConfig returns the current configuration
func (s *Sampler) GetConfig() *Config {
	return s.config
}

// SetDefaultRate updates the default sampling rate
func (s *Sampler) SetDefaultRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.DefaultRate = rate
}

// AddRule adds a custom sampling rule
func (s *Sampler) AddRule(condition RuleCondition, rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Rules = append(s.config.Rules, Rule{
		Condition: condition,
		Rate:      rate,
	})
}
