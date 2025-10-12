package guid

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

var (
	// UUID format regex pattern (8-4-4-4-12)
	uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// GenerateRandom generates a random UUID v4 following RFC 4122 standard
func GenerateRandom() (string, error) {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Set version (4) and variant bits according to RFC 4122
	// Version 4: bits 12-15 of time_hi_and_version field = 0100
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	// Variant: bits 6-7 of clock_seq_hi_and_reserved field = 10
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	// Format as standard UUID string (8-4-4-4-12)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16]), nil
}

// GenerateFromStrings generates a deterministic UUID based on two input strings
// Same input strings will always produce the same UUID
func GenerateFromStrings(str1, str2 string) (string, error) {
	if str1 == "" && str2 == "" {
		return "", fmt.Errorf("at least one input string must be non-empty")
	}

	// Concatenate strings with separator to avoid collision
	// Example: ("ab", "cd") vs ("a", "bcd") should produce different results
	combined := str1 + "|" + str2

	// Generate SHA-256 hash
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hash := hasher.Sum(nil)

	// Use first 16 bytes of hash for UUID
	bytes := hash[:16]

	// Set version (5) and variant bits according to RFC 4122
	// Version 5: bits 12-15 of time_hi_and_version field = 0101
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	// Variant: bits 6-7 of clock_seq_hi_and_reserved field = 10
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	// Format as standard UUID string (8-4-4-4-12)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16]), nil
}

// Validate checks if the provided string is a valid UUID format
func Validate(uuid string) bool {
	if uuid == "" {
		return false
	}
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// IsValid is an alias for Validate for better API consistency
func IsValid(uuid string) bool {
	return Validate(uuid)
}

// Format ensures a UUID string is in lowercase format
func Format(uuid string) string {
	if !Validate(uuid) {
		return uuid // Return as-is if invalid
	}
	return strings.ToLower(uuid)
}