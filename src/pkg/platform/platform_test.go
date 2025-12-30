package platform

import (
	"os"
	"runtime"
	"testing"
)

func TestGetOS(t *testing.T) {
	os := GetOS()
	validOS := []string{"windows", "linux", "darwin", "windows_sandbox"}

	found := false
	for _, valid := range validOS {
		if os == valid {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("GetOS() returned unexpected value: %s", os)
	}
}

func TestGetArchitecture(t *testing.T) {
	arch := GetArchitecture()
	validArchs := []string{"x64", "x86", "arm64", "arm"}

	found := false
	for _, valid := range validArchs {
		if arch == valid {
			found = true
			break
		}
	}

	if !found {
		// Allow unknown architectures for forward compatibility
		t.Logf("GetArchitecture() returned non-standard value: %s (this may be OK)", arch)
	}
}

func TestGetPlatform(t *testing.T) {
	platform := GetPlatform()

	// Platform should be in format "os-arch"
	if len(platform) < 3 {
		t.Errorf("GetPlatform() returned invalid format: %s", platform)
	}

	// Should contain a dash
	if !contains(platform, "-") {
		t.Errorf("GetPlatform() missing separator: %s", platform)
	}
}

func TestIsWindows(t *testing.T) {
	result := IsWindows()
	expected := runtime.GOOS == "windows"

	if result != expected {
		t.Errorf("IsWindows() = %v, want %v", result, expected)
	}
}

func TestIsLinux(t *testing.T) {
	result := IsLinux()
	expected := runtime.GOOS == "linux"

	if result != expected {
		t.Errorf("IsLinux() = %v, want %v", result, expected)
	}
}

func TestIsDarwin(t *testing.T) {
	result := IsDarwin()
	expected := runtime.GOOS == "darwin"

	if result != expected {
		t.Errorf("IsDarwin() = %v, want %v", result, expected)
	}
}

func TestIsWindowsSandbox(t *testing.T) {
	// Save original value
	original := os.Getenv("PORTUNIX_SANDBOX")
	defer os.Setenv("PORTUNIX_SANDBOX", original)

	// Test with sandbox env set
	os.Setenv("PORTUNIX_SANDBOX", "windows")
	if !IsWindowsSandbox() {
		t.Error("IsWindowsSandbox() should return true when PORTUNIX_SANDBOX=windows")
	}

	// Test without sandbox env
	os.Unsetenv("PORTUNIX_SANDBOX")
	if IsWindowsSandbox() {
		t.Error("IsWindowsSandbox() should return false when PORTUNIX_SANDBOX is not set")
	}
}

func TestGetSudoPrefix(t *testing.T) {
	prefix := GetSudoPrefix()

	// Should either be empty or "sudo "
	if prefix != "" && prefix != "sudo " {
		t.Errorf("GetSudoPrefix() returned unexpected value: %q", prefix)
	}

	// On Windows, should always be empty
	if runtime.GOOS == "windows" && prefix != "" {
		t.Errorf("GetSudoPrefix() should return empty string on Windows, got: %q", prefix)
	}
}

func TestIsRunningAsRoot(t *testing.T) {
	// Just verify it doesn't panic and returns a bool
	result := IsRunningAsRoot()
	t.Logf("IsRunningAsRoot() = %v", result)

	// On Windows, should always return false (for now)
	if runtime.GOOS == "windows" && result {
		t.Error("IsRunningAsRoot() should return false on Windows (simplified implementation)")
	}
}

func TestIsSudoAvailable(t *testing.T) {
	// Just verify it doesn't panic and returns a bool
	result := IsSudoAvailable()
	t.Logf("IsSudoAvailable() = %v", result)

	// On Windows, should always return false
	if runtime.GOOS == "windows" && result {
		t.Error("IsSudoAvailable() should return false on Windows")
	}
}

func TestCanWriteToDirectory(t *testing.T) {
	// Test with a directory we should be able to create/write
	tempDir := t.TempDir()

	if !CanWriteToDirectory(tempDir) {
		t.Errorf("CanWriteToDirectory(%s) should return true", tempDir)
	}
}

func TestIsUserDirectory(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Unable to get home directory")
	}

	// Home directory itself should return true
	if !IsUserDirectory(homeDir) {
		t.Errorf("IsUserDirectory(%s) should return true", homeDir)
	}

	// Root directory should return false
	if IsUserDirectory("/") {
		t.Error("IsUserDirectory(\"/\") should return false")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
