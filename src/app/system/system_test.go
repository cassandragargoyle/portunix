package system

import (
	"os"
	"runtime"
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

// BenchmarkCheckCapabilities measures the aggregate capabilities-detection phase
// (PowerShell, Docker/Podman versions+daemon, compose, admin, certs, virtualization).
func BenchmarkCheckCapabilities(b *testing.B) {
	info, err := GetSystemInfo()
	if err != nil {
		b.Fatalf("GetSystemInfo failed: %v", err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		info.Capabilities = &Capabilities{}
		checkCapabilities(info, SystemInfoOptions{})
	}
}

// BenchmarkDetectEnvironment measures environment detection
// (sandbox/docker/wsl/VM heuristics).
func BenchmarkDetectEnvironment(b *testing.B) {
	info, err := GetSystemInfo()
	if err != nil {
		b.Fatalf("GetSystemInfo failed: %v", err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		info.Environment = info.Environment[:0]
		info.Variant = ""
		detectEnvironment(info)
	}
}

// BenchmarkCheckVirtualization measures virtualization-backend detection
// (QEMU, VirtualBox, libvirt, KVM, hardware virt).
func BenchmarkCheckVirtualization(b *testing.B) {
	info, err := GetSystemInfo()
	if err != nil {
		b.Fatalf("GetSystemInfo failed: %v", err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		info.Capabilities.VirtualizationInfo = nil
		checkVirtualizationCapabilities(info)
	}
}

// BenchmarkCheckHardwareVirt measures hardware-virtualization probe in isolation.
// On Linux this reads /proc/cpuinfo; on Windows it uses native API or PowerShell.
func BenchmarkCheckHardwareVirt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = checkHardwareVirtualization()
	}
}

// BenchmarkAdminCheck measures the admin/root privilege check.
// On Unix this is a single syscall (geteuid); on Windows it uses native API
// or `net session` fallback.
func BenchmarkAdminCheck(b *testing.B) {
	b.ReportAllocs()
	if runtime.GOOS == "windows" {
		for i := 0; i < b.N; i++ {
			if nativeIsAdmin != nil {
				_ = nativeIsAdmin()
			} else {
				// Fallback path is intentionally not exercised here:
				// it spawns `net session` and would dominate the benchmark.
				_ = false
			}
		}
		return
	}
	for i := 0; i < b.N; i++ {
		_ = os.Geteuid() == 0
	}
}

// BenchmarkVMDetection measures the VM-classification subset of environment
// detection (for Linux this is mostly /proc/version + virt heuristics; for
// Windows this exercises the wmic computersystem path or the native module).
func BenchmarkVMDetection(b *testing.B) {
	info, err := GetSystemInfo()
	if err != nil {
		b.Fatalf("GetSystemInfo failed: %v", err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		info.Environment = info.Environment[:0]
		info.Variant = ""
		detectEnvironment(info)
	}
}

// BenchmarkCertificateBundle measures certificate-bundle detection
// (file system probes on Linux, store probes on Windows).
func BenchmarkCertificateBundle(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = DetectCertificateBundle()
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
