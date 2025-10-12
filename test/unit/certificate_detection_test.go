package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"portunix.ai/app/system"
)

func TestDetectCertificateBundle(t *testing.T) {
	// Create a temporary certificate file for testing
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca-certificates.crt")
	
	certContent := `-----BEGIN CERTIFICATE-----
MIIDQTCCAimgAwIBAgITBmyfz5m/jAo54vB4ikPmljZbyjANBgkqhkiG9w0BAQsF
ADA5MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6
b24gUm9vdCBDQSAxMB4XDTE1MDUyNjAwMDAwMFoXDTM4MDExNzAwMDAwMFowOTEL
MAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJv
b3QgQ0EgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALJ4gHHKeNXj
ca9HgFB0fW7Y14h29Jlo91ghYPl0hAEvrAIthtOgQ3pOsqTQNroBvo3bSMgHFzZM
9O6II8c+6zf1tRn4SWiw3te5djgdYZ6k/oI2peVKVuRF4fn9tBb6dNqcmzU5L/qw
IFAGbHrQgLKm+a/sRxmPUDgH3KKHOVj4utWp+UhnMJbulHheb4mjUcAwhmahRWa6
VOujw5H5SNz/0egwLX0tdHA114gk957EWW67c4cX8jJGKLhD+rcdqsq08p8kDi1L
93FcXmn/6pUCyziKrlA4b9v7LWIbxcceVOF34GfID5yHI9Y/QCB/IIDEgEw+OyQm
jgSubJrIqg0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMC
AYYwHQYDVR0OBBYEFIQYzIU07LwMlJQuCFmcx7IQTgoIMA0GCSqGSIb3DQEBCwUA
A4IBAQCY8jdaQZChGsV2USggNiMOruYou6r4lK5IpDB/G/wkjUu0yKGX9rbxenDI
U5PMCCjjmCXPI6T53iHTfIuJruydjsw2hUwsqdkPAJGFZqnGsJiW4ueKLhBBmKAb
Q1kJiKP3KQImQhSAgjZGZeFZqEKvN9Rnlw0W8Fw3UUNWW2Qmh+Vh8Ia8yAQ7cXBj
-----END CERTIFICATE-----`
	
	err := os.WriteFile(certPath, []byte(certContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test certificate file: %v", err)
	}
	
	// Test with custom cert paths
	certInfo, err := system.DetectCertificateBundleWithPaths([]string{certPath})
	if err != nil {
		t.Fatalf("DetectCertificateBundle failed: %v", err)
	}
	
	if !certInfo.Available {
		t.Error("Expected certificate to be available")
	}
	
	if certInfo.Path != certPath {
		t.Errorf("Expected path %s, got %s", certPath, certInfo.Path)
	}
	
	if certInfo.Size == 0 {
		t.Error("Expected certificate file size > 0")
	}
	
	// Check that ModTime is recent (within last minute)
	if time.Since(certInfo.ModTime) > time.Minute {
		t.Error("Expected recent modification time")
	}
}

func TestDetectCertificateBundle_NotFound(t *testing.T) {
	// Test with non-existent paths
	certInfo, err := system.DetectCertificateBundleWithPaths([]string{"/nonexistent/path"})
	if err != nil {
		t.Fatalf("DetectCertificateBundle failed: %v", err)
	}
	
	if certInfo.Available {
		t.Error("Expected certificate to be unavailable")
	}
	
	if certInfo.Path != "" {
		t.Errorf("Expected empty path, got %s", certInfo.Path)
	}
	
	if certInfo.Size != 0 {
		t.Errorf("Expected size 0, got %d", certInfo.Size)
	}
}

func TestCertificateInfo_JSON(t *testing.T) {
	certInfo := system.CertificateInfo{
		Available:    true,
		Path:         "/etc/ssl/certs/ca-certificates.crt",
		Size:         1024,
		ModTime:      time.Now(),
		HTTPSWorking: true,
	}
	
	// Test that struct can be marshaled to JSON (for system info output)
	data, err := certInfo.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal CertificateInfo to JSON: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Expected non-empty JSON data")
	}
}

func TestGenerateCertificateInstallCommand(t *testing.T) {
	tests := []struct {
		pkgManager string
		expected   []string
	}{
		{"apt-get", []string{"apt-get", "install", "-y", "ca-certificates", "curl"}},
		{"apt", []string{"apt-get", "install", "-y", "ca-certificates", "curl"}},
		{"yum", []string{"yum", "install", "-y", "ca-certificates", "curl"}},
		{"dnf", []string{"dnf", "install", "-y", "ca-certificates", "curl"}},
		{"apk", []string{"apk", "add", "ca-certificates", "curl"}},
		{"unknown", []string{"echo", "Unknown package manager for certificate installation"}},
	}
	
	for _, test := range tests {
		cmd := system.GenerateCertificateInstallCommand(test.pkgManager)
		
		if len(cmd) != len(test.expected) {
			t.Errorf("For %s: expected %d args, got %d", test.pkgManager, len(test.expected), len(cmd))
			continue
		}
		
		for i, arg := range cmd {
			if arg != test.expected[i] {
				t.Errorf("For %s: expected arg[%d] = %s, got %s", test.pkgManager, i, test.expected[i], arg)
			}
		}
	}
}

func TestGenerateCertificateUpdateCommand(t *testing.T) {
	tests := []struct {
		pkgManager string
		expected   []string
	}{
		{"apt-get", []string{"update-ca-certificates"}},
		{"apt", []string{"update-ca-certificates"}},
		{"yum", []string{"update-ca-trust"}},
		{"dnf", []string{"update-ca-trust"}},
		{"apk", []string{"update-ca-certificates"}},
		{"unknown", []string{"echo", "Unknown package manager for certificate update"}},
	}
	
	for _, test := range tests {
		cmd := system.GenerateCertificateUpdateCommand(test.pkgManager)
		
		if len(cmd) != len(test.expected) {
			t.Errorf("For %s: expected %d args, got %d", test.pkgManager, len(test.expected), len(cmd))
			continue
		}
		
		for i, arg := range cmd {
			if arg != test.expected[i] {
				t.Errorf("For %s: expected arg[%d] = %s, got %s", test.pkgManager, i, test.expected[i], arg)
			}
		}
	}
}