package cmd

import (
	"testing"
)

func TestIsEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid env vars
		{"simple env var", "GOOS=linux", true},
		{"env var with number", "CGO_ENABLED=0", true},
		{"env var with underscore", "MY_VAR=value", true},
		{"env var starting with underscore", "_VAR=value", true},
		{"empty value", "GOOS=", true},
		{"value with equals", "LDFLAGS=-X main.version=1.0.0", true},
		{"value with spaces", "MSG=hello world", true},
		{"multiple underscores", "MY_LONG_VAR_NAME=test", true},

		// Invalid env vars
		{"lowercase", "goos=linux", false},
		{"mixed case key", "GoOs=linux", false},
		{"no equals sign", "GOOS", false},
		{"starts with number", "1VAR=value", false},
		{"command", "go", false},
		{"path", "/usr/bin/go", false},
		{"flag", "-o", false},
		{"flag with value", "--output=file", false},
		{"relative path", "./myapp", false},
		{"dot in key", "MY.VAR=value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEnvVar(tt.input)
			if result != tt.expected {
				t.Errorf("IsEnvVar(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseEnvVars(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedEnv    map[string]string
		expectedCmdIdx int
	}{
		{
			name:           "single env var",
			args:           []string{"GOOS=linux", "go", "build"},
			expectedEnv:    map[string]string{"GOOS": "linux"},
			expectedCmdIdx: 1,
		},
		{
			name:           "multiple env vars",
			args:           []string{"GOOS=linux", "GOARCH=amd64", "go", "build"},
			expectedEnv:    map[string]string{"GOOS": "linux", "GOARCH": "amd64"},
			expectedCmdIdx: 2,
		},
		{
			name:           "three env vars",
			args:           []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=arm64", "go", "build", "-o", "output"},
			expectedEnv:    map[string]string{"CGO_ENABLED": "0", "GOOS": "linux", "GOARCH": "arm64"},
			expectedCmdIdx: 3,
		},
		{
			name:           "no env vars",
			args:           []string{"go", "build"},
			expectedEnv:    map[string]string{},
			expectedCmdIdx: 0,
		},
		{
			name:           "empty value",
			args:           []string{"GOOS=", "go", "version"},
			expectedEnv:    map[string]string{"GOOS": ""},
			expectedCmdIdx: 1,
		},
		{
			name:           "value with equals sign",
			args:           []string{"LDFLAGS=-X main.version=1.0.0", "go", "build"},
			expectedEnv:    map[string]string{"LDFLAGS": "-X main.version=1.0.0"},
			expectedCmdIdx: 1,
		},
		{
			name:           "only env vars no command",
			args:           []string{"GOOS=linux", "GOARCH=amd64"},
			expectedEnv:    map[string]string{"GOOS": "linux", "GOARCH": "amd64"},
			expectedCmdIdx: 2,
		},
		{
			name:           "empty args",
			args:           []string{},
			expectedEnv:    map[string]string{},
			expectedCmdIdx: 0,
		},
		{
			name:           "complex ldflags",
			args:           []string{"GOOS=linux", "go", "build", "-ldflags", "-X main.version=1.0.0 -X main.commit=abc123"},
			expectedEnv:    map[string]string{"GOOS": "linux"},
			expectedCmdIdx: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars, cmdIdx := ParseEnvVars(tt.args)

			// Check command index
			if cmdIdx != tt.expectedCmdIdx {
				t.Errorf("ParseEnvVars() cmdIdx = %v, want %v", cmdIdx, tt.expectedCmdIdx)
			}

			// Check env vars count
			if len(envVars) != len(tt.expectedEnv) {
				t.Errorf("ParseEnvVars() env count = %v, want %v", len(envVars), len(tt.expectedEnv))
			}

			// Check each env var
			for key, expectedVal := range tt.expectedEnv {
				if actualVal, ok := envVars[key]; !ok {
					t.Errorf("ParseEnvVars() missing key %q", key)
				} else if actualVal != expectedVal {
					t.Errorf("ParseEnvVars() env[%q] = %q, want %q", key, actualVal, expectedVal)
				}
			}
		})
	}
}

func TestParseEnvVarsCommandExtraction(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedCmd string
	}{
		{
			name:        "go command",
			args:        []string{"GOOS=linux", "go", "build"},
			expectedCmd: "go",
		},
		{
			name:        "echo command",
			args:        []string{"MY_VAR=test", "echo", "hello"},
			expectedCmd: "echo",
		},
		{
			name:        "no env vars",
			args:        []string{"go", "version"},
			expectedCmd: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmdIdx := ParseEnvVars(tt.args)

			if cmdIdx >= len(tt.args) {
				t.Errorf("ParseEnvVars() cmdIdx %d out of range for args %v", cmdIdx, tt.args)
				return
			}

			actualCmd := tt.args[cmdIdx]
			if actualCmd != tt.expectedCmd {
				t.Errorf("Command at index %d = %q, want %q", cmdIdx, actualCmd, tt.expectedCmd)
			}
		})
	}
}
