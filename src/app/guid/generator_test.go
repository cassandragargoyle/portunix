package guid

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateRandom(t *testing.T) {
	// Test basic generation
	uuid1, err := GenerateRandom()
	if err != nil {
		t.Fatalf("GenerateRandom failed: %v", err)
	}

	if !Validate(uuid1) {
		t.Errorf("Generated UUID is not valid: %s", uuid1)
	}

	// Test uniqueness - generate multiple UUIDs and ensure they're different
	uuids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		uuid, err := GenerateRandom()
		if err != nil {
			t.Fatalf("GenerateRandom failed on iteration %d: %v", i, err)
		}

		if uuids[uuid] {
			t.Errorf("Duplicate UUID generated: %s", uuid)
		}
		uuids[uuid] = true

		if !Validate(uuid) {
			t.Errorf("Invalid UUID generated: %s", uuid)
		}
	}

	// Test UUID version (should be 4)
	uuid, _ := GenerateRandom()
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		t.Errorf("UUID should have 5 parts, got %d", len(parts))
	}

	// Check version nibble (should be 4)
	versionChar := parts[2][0]
	if versionChar != '4' {
		t.Errorf("Expected version 4, got version %c", versionChar)
	}

	// Check variant nibble (should be 8, 9, a, or b)
	variantChar := parts[3][0]
	if variantChar != '8' && variantChar != '9' && variantChar != 'a' && variantChar != 'b' {
		t.Errorf("Invalid variant nibble: %c", variantChar)
	}
}

func TestGenerateFromStrings(t *testing.T) {
	testCases := []struct {
		str1     string
		str2     string
		expected bool // whether it should succeed
		name     string
	}{
		{"hello", "world", true, "basic strings"},
		{"", "world", true, "empty first string"},
		{"hello", "", true, "empty second string"},
		{"", "", false, "both empty strings"},
		{"test", "123", true, "string and number"},
		{"special!@#", "chars$%^", true, "special characters"},
		{"unicode🚀", "test", true, "unicode characters"},
		{"very long string that contains many characters", "another long string", true, "long strings"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uuid, err := GenerateFromStrings(tc.str1, tc.str2)

			if tc.expected {
				if err != nil {
					t.Errorf("GenerateFromStrings failed for valid input: %v", err)
				}
				if !Validate(uuid) {
					t.Errorf("Generated UUID is not valid: %s", uuid)
				}

				// Test deterministic behavior - same inputs should produce same output
				uuid2, err2 := GenerateFromStrings(tc.str1, tc.str2)
				if err2 != nil {
					t.Errorf("Second generation failed: %v", err2)
				}
				if uuid != uuid2 {
					t.Errorf("Deterministic generation failed: %s != %s", uuid, uuid2)
				}

				// Check version (should be 5 for name-based UUIDs)
				parts := strings.Split(uuid, "-")
				if len(parts) != 5 {
					t.Errorf("UUID should have 5 parts, got %d", len(parts))
				}
				versionChar := parts[2][0]
				if versionChar != '5' {
					t.Errorf("Expected version 5, got version %c", versionChar)
				}
			} else {
				if err == nil {
					t.Errorf("GenerateFromStrings should have failed for invalid input")
				}
			}
		})
	}
}

func TestGenerateFromStringsCollisionResistance(t *testing.T) {
	// Test that different input combinations produce different UUIDs
	testPairs := [][]string{
		{"ab", "cd"},
		{"a", "bcd"},
		{"abc", "d"},
		{"", "abcd"},
		{"abcd", ""},
		{"hello", "world"},
		{"world", "hello"},
	}

	uuids := make(map[string]string)
	for i, pair := range testPairs {
		uuid, err := GenerateFromStrings(pair[0], pair[1])
		if err != nil && !(pair[0] == "" && pair[1] == "") {
			t.Errorf("Unexpected error for pair %d (%s, %s): %v", i, pair[0], pair[1], err)
			continue
		}
		if err == nil {
			key := pair[0] + "|" + pair[1]
			if existing, exists := uuids[uuid]; exists {
				t.Errorf("Collision detected: (%s) and (%s) both produce UUID %s", key, existing, uuid)
			}
			uuids[uuid] = key
		}
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		name     string
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true, "valid UUID"},
		{"550E8400-E29B-41D4-A716-446655440000", true, "valid UUID uppercase"},
		{"550e8400-e29b-41d4-a716-44665544000", false, "too short"},
		{"550e8400-e29b-41d4-a716-4466554400000", false, "too long"},
		{"550e8400e29b41d4a716446655440000", false, "missing hyphens"},
		{"550e8400-e29b-41d4-a716-44665544000g", false, "invalid character"},
		{"", false, "empty string"},
		{"not-a-uuid", false, "completely invalid"},
		{"550e8400-e29b-41d4-a716", false, "missing last segment"},
		{"550e8400-e29b-41d4-a716-", false, "missing last segment value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Validate(tc.input)
			if result != tc.expected {
				t.Errorf("Validate(%s) = %v, expected %v", tc.input, result, tc.expected)
			}

			// Test IsValid alias
			resultAlias := IsValid(tc.input)
			if resultAlias != tc.expected {
				t.Errorf("IsValid(%s) = %v, expected %v", tc.input, resultAlias, tc.expected)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		name     string
	}{
		{
			"550E8400-E29B-41D4-A716-446655440000",
			"550e8400-e29b-41d4-a716-446655440000",
			"uppercase to lowercase",
		},
		{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440000",
			"already lowercase",
		},
		{
			"invalid-uuid",
			"invalid-uuid",
			"invalid UUID returned as-is",
		},
		{
			"",
			"",
			"empty string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Format(tc.input)
			if result != tc.expected {
				t.Errorf("Format(%s) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRandom()
		if err != nil {
			b.Fatalf("GenerateRandom failed: %v", err)
		}
	}
}

func BenchmarkGenerateFromStrings(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateFromStrings("test-string-1", "test-string-2")
		if err != nil {
			b.Fatalf("GenerateFromStrings failed: %v", err)
		}
	}
}

func BenchmarkValidate(b *testing.B) {
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(uuid)
	}
}

// Test performance requirement (sub-millisecond)
func TestPerformanceRequirement(t *testing.T) {
	iterations := 1000

	// Test random generation performance
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := GenerateRandom()
		if err != nil {
			t.Fatalf("GenerateRandom failed: %v", err)
		}
	}
	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	if avgDuration > time.Millisecond {
		t.Errorf("Random generation too slow: average %v per operation (requirement: < 1ms)", avgDuration)
	}

	// Test deterministic generation performance
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := GenerateFromStrings("test", "performance")
		if err != nil {
			t.Fatalf("GenerateFromStrings failed: %v", err)
		}
	}
	duration = time.Since(start)
	avgDuration = duration / time.Duration(iterations)

	if avgDuration > time.Millisecond {
		t.Errorf("Deterministic generation too slow: average %v per operation (requirement: < 1ms)", avgDuration)
	}
}