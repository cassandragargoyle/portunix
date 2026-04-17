package system

import (
	"testing"
)

func TestParseDockerVersionOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard docker version output",
			input:    "Docker version 27.5.1, build 9f9e405",
			expected: "27.5.1",
		},
		{
			name:     "older docker version format",
			input:    "Docker version 24.0.7, build afdd53b",
			expected: "24.0.7",
		},
		{
			name:     "version with newline",
			input:    "Docker version 28.3.2, build abc1234\n",
			expected: "28.3.2",
		},
		{
			name:     "version with leading whitespace",
			input:    "  Docker version 26.1.0, build xyz789  ",
			expected: "26.1.0",
		},
		{
			name:     "empty output",
			input:    "",
			expected: "",
		},
		{
			name:     "no version number present",
			input:    "Docker not found",
			expected: "",
		},
		{
			name:     "must not return ersion from word version",
			input:    "Docker version something-wrong",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDockerVersionOutput(tt.input)
			if result != tt.expected {
				t.Errorf("parseDockerVersionOutput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
