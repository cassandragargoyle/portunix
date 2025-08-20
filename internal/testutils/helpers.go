package testutils

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T, prefix string) string {
	t.Helper()
	
	tempDir, err := os.MkdirTemp("", prefix)
	require.NoError(t, err, "Failed to create temp directory")
	
	// Cleanup on test completion
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	
	return tempDir
}

// CreateTempFile creates a temporary file with content for testing
func CreateTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	
	if dir == "" {
		dir = CreateTempDir(t, "testfile")
	}
	
	filePath := filepath.Join(dir, name)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create temp file")
	
	return filePath
}

// SetupTestCache creates a test cache directory structure
func SetupTestCache(t *testing.T) string {
	t.Helper()
	
	cacheDir := CreateTempDir(t, "portunix-test-cache")
	
	// Create subdirectories
	subdirs := []string{
		"docker",
		"install",
		"python-embeddable",
		"openssh",
		"notepadplusplus",
	}
	
	for _, subdir := range subdirs {
		err := os.MkdirAll(filepath.Join(cacheDir, subdir), 0755)
		require.NoError(t, err, "Failed to create cache subdir %s", subdir)
	}
	
	return cacheDir
}

// MockDockerfile creates a mock Dockerfile for testing
func MockDockerfile(baseImage string) string {
	return `FROM ` + baseImage + `

# Test Dockerfile
RUN echo "This is a test Dockerfile"
WORKDIR /workspace
CMD ["echo", "Hello from test container"]
`
}

// MockPackageJSON creates a mock package.json for testing
func MockPackageJSON() string {
	return `{
  "packages": ["docker", "python", "java"],
  "variants": {
    "python": ["3.11", "3.12"],
    "java": ["11", "17", "21"]
  },
  "test": true
}`
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// SkipIfNoDocker skips the test if Docker is not available
func SkipIfNoDocker(t *testing.T) {
	t.Helper()
	
	if !IsDockerAvailable() {
		t.Skip("Docker is not available, skipping test")
	}
}

// IsDockerAvailable checks if Docker is available for testing
func IsDockerAvailable() bool {
	// Try to run a simple Docker command
	return os.Getenv("DOCKER_HOST") != "" || 
		   fileExists("/var/run/docker.sock") ||
		   fileExists("//./pipe/docker_engine") // Windows named pipe
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestLogger provides a simple logger for tests
type TestLogger struct {
	t *testing.T
}

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t: t}
}

func (l *TestLogger) Log(msg string) {
	l.t.Helper()
	l.t.Logf("[TEST] %s", msg)
}

func (l *TestLogger) Logf(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Logf("[TEST] "+format, args...)
}

func (l *TestLogger) Error(msg string) {
	l.t.Helper()
	l.t.Errorf("[ERROR] %s", msg)
}

func (l *TestLogger) Errorf(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Errorf("[ERROR] "+format, args...)
}

// AssertEventuallyEqual checks that a value eventually equals expected within timeout
func AssertEventuallyEqual(t *testing.T, expected interface{}, actual func() interface{}, timeout time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			t.Errorf("Timeout: expected %v to equal %v within %v", actual(), expected, timeout)
			if len(msgAndArgs) > 0 {
				t.Errorf("Additional info: %v", msgAndArgs...)
			}
			return
		case <-ticker.C:
			if actual() == expected {
				return
			}
		}
	}
}

// MockSystemInfo creates a mock system info for testing
func MockSystemInfo(os, version, distribution string) map[string]interface{} {
	return map[string]interface{}{
		"os":      os,
		"version": version,
		"distribution": distribution,
		"architecture": "amd64",
	}
}

// CreateTestFixture creates a test fixture file with given content
func CreateTestFixture(t *testing.T, fixturePath, content string) {
	t.Helper()
	
	dir := filepath.Dir(fixturePath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err, "Failed to create fixture directory")
	
	err = os.WriteFile(fixturePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create fixture file")
	
	t.Cleanup(func() {
		os.Remove(fixturePath)
	})
}

// RandomString generates a random string for testing
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// FreePort finds a free port for testing
func FreePort() int {
	// Simple implementation - in real tests you might want something more robust
	return 8080 + int(time.Now().UnixNano()%1000)
}

// SetEnvForTest sets an environment variable for the duration of the test
func SetEnvForTest(t *testing.T, key, value string) {
	t.Helper()
	
	oldValue := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err, "Failed to set environment variable %s", key)
	
	t.Cleanup(func() {
		if oldValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, oldValue)
		}
	})
}