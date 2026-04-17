/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package alerts

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-trace/models"
)

// Evaluator evaluates alert conditions against session data
type Evaluator struct {
	rules []*Rule
}

// NewEvaluator creates a new alert evaluator
func NewEvaluator(config *AlertConfig) (*Evaluator, error) {
	rules := make([]*Rule, 0, len(config.Alerts.Rules))

	for _, rc := range config.Alerts.Rules {
		rule, err := parseRule(rc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rule '%s': %w", rc.Name, err)
		}
		rules = append(rules, rule)
	}

	return &Evaluator{rules: rules}, nil
}

// EvaluationContext contains data for evaluating conditions
type EvaluationContext struct {
	Session      *models.Session
	Events       []*models.TraceEvent
	WindowStart  time.Time
	WindowEnd    time.Time
	ErrorCount   int
	WarningCount int
	SuccessCount int
	TotalCount   int
	SlowOpsCount int
	ErrorRate    float64
	AvgDuration  float64
}

// Alert represents a fired alert
type Alert struct {
	Rule      *Rule
	Timestamp time.Time
	Context   *EvaluationContext
	Value     float64
	Message   string
}

// Evaluate evaluates all rules against the context
func (e *Evaluator) Evaluate(ctx *EvaluationContext) []*Alert {
	var alerts []*Alert

	for _, rule := range e.rules {
		if alert := e.evaluateRule(rule, ctx); alert != nil {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// EvaluateSession creates context from session and evaluates
func (e *Evaluator) EvaluateSession(session *models.Session, events []*models.TraceEvent) []*Alert {
	ctx := buildContext(session, events)
	return e.Evaluate(ctx)
}

func (e *Evaluator) evaluateRule(rule *Rule, ctx *EvaluationContext) *Alert {
	// Check cooldown
	if !rule.LastFired.IsZero() && time.Since(rule.LastFired) < rule.Cooldown {
		return nil
	}

	// Get value based on condition type
	value := getConditionValue(rule.Condition, ctx)

	// Evaluate condition
	if !evaluateCondition(rule.Condition, value) {
		return nil
	}

	// Rule matched - create alert
	rule.LastFired = time.Now()
	rule.FireCount++

	message := rule.Message
	if message == "" {
		message = fmt.Sprintf("Alert: %s - %s %.2f (threshold: %.2f)",
			rule.Name, rule.Condition.Type, value, rule.Condition.Threshold)
	}

	return &Alert{
		Rule:      rule,
		Timestamp: time.Now(),
		Context:   ctx,
		Value:     value,
		Message:   message,
	}
}

func buildContext(session *models.Session, events []*models.TraceEvent) *EvaluationContext {
	ctx := &EvaluationContext{
		Session:   session,
		Events:    events,
		WindowEnd: time.Now(),
	}

	if session.Stats != nil {
		ctx.ErrorCount = int(session.Stats.ByStatus["error"])
		ctx.WarningCount = int(session.Stats.ByStatus["warning"])
		ctx.SuccessCount = int(session.Stats.ByStatus["success"])
		ctx.TotalCount = int(session.Stats.TotalEvents)
	}

	// Calculate metrics from events
	var totalDuration int64
	slowThreshold := int64(1000) // 1ms in microseconds

	for _, event := range events {
		if event.Level == models.LevelError {
			ctx.ErrorCount++
		}
		if event.DurationUS > slowThreshold {
			ctx.SlowOpsCount++
		}
		totalDuration += event.DurationUS
	}

	if len(events) > 0 {
		ctx.TotalCount = len(events)
		ctx.AvgDuration = float64(totalDuration) / float64(len(events))
	}

	if ctx.TotalCount > 0 {
		ctx.ErrorRate = float64(ctx.ErrorCount) / float64(ctx.TotalCount)
	}

	return ctx
}

func getConditionValue(cond Condition, ctx *EvaluationContext) float64 {
	switch cond.Type {
	case ConditionErrorRate:
		return ctx.ErrorRate
	case ConditionSlowOps:
		return float64(ctx.SlowOpsCount)
	case ConditionErrorCount:
		return float64(ctx.ErrorCount)
	case ConditionEventCount:
		return float64(ctx.TotalCount)
	case ConditionDuration:
		return ctx.AvgDuration
	default:
		return 0
	}
}

func evaluateCondition(cond Condition, value float64) bool {
	switch cond.Operator {
	case ">":
		return value > cond.Threshold
	case ">=":
		return value >= cond.Threshold
	case "<":
		return value < cond.Threshold
	case "<=":
		return value <= cond.Threshold
	case "==":
		return value == cond.Threshold
	case "!=":
		return value != cond.Threshold
	default:
		return false
	}
}

func parseRule(rc RuleConfig) (*Rule, error) {
	condition, err := parseCondition(rc.Condition)
	if err != nil {
		return nil, err
	}

	window, err := ParseDuration(rc.Window)
	if err != nil {
		return nil, fmt.Errorf("invalid window: %w", err)
	}

	cooldown, err := ParseDuration(rc.Cooldown)
	if err != nil {
		cooldown = 5 * time.Minute // default cooldown
	}

	return &Rule{
		Name:      rc.Name,
		Condition: condition,
		Window:    window,
		Severity:  ParseSeverity(rc.Severity),
		Channels:  rc.Channels,
		Cooldown:  cooldown,
		Message:   rc.Message,
	}, nil
}

// parseCondition parses condition string like "error_rate > 0.05"
func parseCondition(s string) (Condition, error) {
	// Pattern: metric operator value
	pattern := regexp.MustCompile(`^\s*(\w+)\s*(>=|<=|>|<|==|!=)\s*([\d.]+)\s*$`)
	matches := pattern.FindStringSubmatch(s)

	if len(matches) != 4 {
		return Condition{}, fmt.Errorf("invalid condition format: %s", s)
	}

	metric := strings.ToLower(matches[1])
	operator := matches[2]
	threshold, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return Condition{}, fmt.Errorf("invalid threshold: %s", matches[3])
	}

	condType := parseConditionType(metric)

	return Condition{
		Type:      condType,
		Operator:  operator,
		Threshold: threshold,
	}, nil
}

func parseConditionType(s string) ConditionType {
	switch s {
	case "error_rate", "errorrate":
		return ConditionErrorRate
	case "slow_ops", "slowops", "slow_operations":
		return ConditionSlowOps
	case "error_count", "errorcount", "errors":
		return ConditionErrorCount
	case "event_count", "eventcount", "events":
		return ConditionEventCount
	case "duration", "avg_duration":
		return ConditionDuration
	default:
		return ConditionCustomMetric
	}
}

// GetRules returns all configured rules
func (e *Evaluator) GetRules() []*Rule {
	return e.rules
}

// ResetCooldowns resets all cooldown timers
func (e *Evaluator) ResetCooldowns() {
	for _, rule := range e.rules {
		rule.LastFired = time.Time{}
	}
}
