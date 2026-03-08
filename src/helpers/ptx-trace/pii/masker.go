package pii

import (
	"regexp"
	"strings"
)

// Masker handles PII data masking
type Masker struct {
	enabled  bool
	patterns []*Pattern
}

// Pattern defines a PII pattern to mask
type Pattern struct {
	Name        string
	Regex       *regexp.Regexp
	Mask        string
	PreserveLast int // Preserve last N characters
}

// NewMasker creates a new PII masker with default patterns
func NewMasker(enabled bool) *Masker {
	m := &Masker{
		enabled:  enabled,
		patterns: defaultPatterns(),
	}
	return m
}

// defaultPatterns returns the default PII patterns
func defaultPatterns() []*Pattern {
	return []*Pattern{
		{
			Name:  "email",
			Regex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			Mask:  "***@***.***",
		},
		{
			Name:  "phone_international",
			Regex: regexp.MustCompile(`\+[0-9]{1,3}[\s-]?[0-9]{3,4}[\s-]?[0-9]{3,4}[\s-]?[0-9]{3,4}`),
			Mask:  "+**-***-***-****",
		},
		{
			Name:  "phone_czech",
			Regex: regexp.MustCompile(`\+420[\s-]?[0-9]{3}[\s-]?[0-9]{3}[\s-]?[0-9]{3}`),
			Mask:  "+420-***-***-***",
		},
		{
			Name:  "phone_us",
			Regex: regexp.MustCompile(`\(?[0-9]{3}\)?[\s.-]?[0-9]{3}[\s.-]?[0-9]{4}`),
			Mask:  "(***) ***-****",
		},
		{
			Name:         "credit_card",
			Regex:        regexp.MustCompile(`\b[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}\b`),
			Mask:         "****-****-****-",
			PreserveLast: 4,
		},
		{
			Name:  "ssn",
			Regex: regexp.MustCompile(`\b[0-9]{3}-[0-9]{2}-[0-9]{4}\b`),
			Mask:  "***-**-****",
		},
		{
			Name:  "ipv4",
			Regex: regexp.MustCompile(`\b[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\b`),
			Mask:  "***.***.***.***",
		},
		{
			Name:  "api_key",
			Regex: regexp.MustCompile(`(?i)(api[_-]?key|apikey|api_secret|secret_key)[\s]*[=:]\s*["']?([a-zA-Z0-9_-]{16,})["']?`),
			Mask:  "$1=***REDACTED***",
		},
		{
			Name:  "password_field",
			Regex: regexp.MustCompile(`(?i)(password|passwd|pwd)[\s]*[=:]\s*["']?([^"'\s]+)["']?`),
			Mask:  "$1=***REDACTED***",
		},
	}
}

// IsEnabled returns whether masking is enabled
func (m *Masker) IsEnabled() bool {
	return m.enabled
}

// Enable enables PII masking
func (m *Masker) Enable() {
	m.enabled = true
}

// Disable disables PII masking
func (m *Masker) Disable() {
	m.enabled = false
}

// MaskString masks PII in a string
func (m *Masker) MaskString(s string) string {
	if !m.enabled || s == "" {
		return s
	}

	result := s
	for _, p := range m.patterns {
		if p.PreserveLast > 0 {
			// Handle patterns that preserve last N chars (like credit cards)
			result = p.Regex.ReplaceAllStringFunc(result, func(match string) string {
				clean := strings.ReplaceAll(strings.ReplaceAll(match, " ", ""), "-", "")
				if len(clean) >= p.PreserveLast {
					return p.Mask + clean[len(clean)-p.PreserveLast:]
				}
				return p.Mask
			})
		} else {
			result = p.Regex.ReplaceAllString(result, p.Mask)
		}
	}

	return result
}

// MaskMap masks PII in a map of strings
func (m *Masker) MaskMap(data map[string]interface{}) map[string]interface{} {
	if !m.enabled || data == nil {
		return data
	}

	result := make(map[string]interface{})
	for k, v := range data {
		// Check if key itself suggests sensitive data
		keyLower := strings.ToLower(k)
		if m.isSensitiveKey(keyLower) {
			result[k] = "***REDACTED***"
			continue
		}

		// Mask string values
		switch val := v.(type) {
		case string:
			result[k] = m.MaskString(val)
		case map[string]interface{}:
			result[k] = m.MaskMap(val)
		default:
			result[k] = v
		}
	}

	return result
}

// isSensitiveKey checks if a key name suggests sensitive data
func (m *Masker) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "passwd", "pwd", "secret",
		"api_key", "apikey", "api-key",
		"token", "auth_token", "access_token",
		"credit_card", "creditcard", "card_number",
		"ssn", "social_security",
		"private_key", "privatekey",
	}

	for _, sk := range sensitiveKeys {
		if strings.Contains(key, sk) {
			return true
		}
	}
	return false
}

// AddPattern adds a custom PII pattern
func (m *Masker) AddPattern(name, pattern, mask string, preserveLast int) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	m.patterns = append(m.patterns, &Pattern{
		Name:         name,
		Regex:        regex,
		Mask:         mask,
		PreserveLast: preserveLast,
	})

	return nil
}

// RemovePattern removes a pattern by name
func (m *Masker) RemovePattern(name string) {
	var filtered []*Pattern
	for _, p := range m.patterns {
		if p.Name != name {
			filtered = append(filtered, p)
		}
	}
	m.patterns = filtered
}

// PatternNames returns the names of all registered patterns
func (m *Masker) PatternNames() []string {
	var names []string
	for _, p := range m.patterns {
		names = append(names, p.Name)
	}
	return names
}
