package system

import (
	"testing"
)

// BenchmarkGetSystemInfo measures the performance of the full system info retrieval
func BenchmarkGetSystemInfo(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := GetSystemInfo()
		if err != nil {
			b.Fatalf("GetSystemInfo failed: %v", err)
		}
	}
}

// BenchmarkGetDockerVersion measures the performance of Docker version retrieval
func BenchmarkGetDockerVersion(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetDockerVersion()
	}
}

// BenchmarkGetPodmanVersion measures the performance of Podman version retrieval
func BenchmarkGetPodmanVersion(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetPodmanVersion()
	}
}

// TestGetSystemInfo verifies that system info can be retrieved without error
func TestGetSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	// Verify basic fields are populated
	if info.OS == "" {
		t.Error("OS field is empty")
	}
	if info.Architecture == "" {
		t.Error("Architecture field is empty")
	}
	if info.Hostname == "" {
		t.Error("Hostname field is empty")
	}
}

// TestCheckCondition tests the condition checking functionality
func TestCheckCondition(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	// Test that the current OS condition returns true
	switch info.OS {
	case "Windows":
		if !CheckCondition(info, "windows") {
			t.Error("CheckCondition for windows should return true on Windows")
		}
	case "Linux":
		if !CheckCondition(info, "linux") {
			t.Error("CheckCondition for linux should return true on Linux")
		}
	case "Darwin":
		if !CheckCondition(info, "macos") {
			t.Error("CheckCondition for macos should return true on macOS")
		}
	}
}
