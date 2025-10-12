package system

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// CertificateInfo contains information about system CA certificate bundle
type CertificateInfo struct {
	Available    bool      `json:"available"`
	Path         string    `json:"path,omitempty"`
	Size         int64     `json:"size,omitempty"`
	ModTime      time.Time `json:"mod_time,omitempty"`
	HTTPSWorking bool      `json:"https_working"`
}

// MarshalJSON implements custom JSON marshaling for CertificateInfo
func (c CertificateInfo) MarshalJSON() ([]byte, error) {
	type Alias CertificateInfo
	return json.Marshal(&struct {
		ModTime string `json:"mod_time,omitempty"`
		Alias
	}{
		ModTime: c.ModTime.Format(time.RFC3339),
		Alias:   (Alias)(c),
	})
}

// DetectCertificateBundle detects the system CA certificate bundle
func DetectCertificateBundle() (CertificateInfo, error) {
	// Default certificate paths for different systems
	certPaths := []string{
		"/etc/ssl/certs/ca-certificates.crt",        // Ubuntu/Debian
		"/etc/pki/tls/certs/ca-bundle.crt",         // RHEL/CentOS
		"/etc/ssl/ca-bundle.pem",                   // openSUSE
		"/usr/local/share/certs/ca-root-nss.crt",  // FreeBSD
		"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem", // Modern RHEL/CentOS
		"/system/etc/security/cacerts",             // Android (if applicable)
	}
	
	return DetectCertificateBundleWithPaths(certPaths)
}

// DetectCertificateBundleWithPaths detects certificate bundle with custom paths (for testing)
func DetectCertificateBundleWithPaths(certPaths []string) (CertificateInfo, error) {
	var certInfo CertificateInfo
	
	// Check each certificate path
	for _, path := range certPaths {
		if fileExists(path) {
			stat, err := os.Stat(path)
			if err == nil && !stat.IsDir() && stat.Size() > 0 {
				certInfo.Available = true
				certInfo.Path = path
				certInfo.Size = stat.Size()
				certInfo.ModTime = stat.ModTime()
				
				// Test HTTPS connectivity
				certInfo.HTTPSWorking = testHTTPSConnectivity()
				
				break
			}
		}
	}
	
	return certInfo, nil
}

// testHTTPSConnectivity tests if HTTPS connections work
func testHTTPSConnectivity() bool {
	// Test with a reliable HTTPS endpoint
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	testURLs := []string{
		"https://go.dev/dl/",
		"https://www.google.com",
		"https://github.com",
	}
	
	for _, url := range testURLs {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
	}
	
	return false
}

// GenerateCertificateInstallCommand generates command to install CA certificates
func GenerateCertificateInstallCommand(pkgManager string) []string {
	switch pkgManager {
	case "apt", "apt-get":
		return []string{"apt-get", "install", "-y", "ca-certificates", "curl"}
	case "yum":
		return []string{"yum", "install", "-y", "ca-certificates", "curl"}
	case "dnf":
		return []string{"dnf", "install", "-y", "ca-certificates", "curl"}
	case "apk":
		return []string{"apk", "add", "ca-certificates", "curl"}
	case "pacman":
		return []string{"pacman", "-S", "--noconfirm", "ca-certificates", "curl"}
	case "zypper":
		return []string{"zypper", "install", "-y", "ca-certificates", "curl"}
	default:
		return []string{"echo", "Unknown package manager for certificate installation"}
	}
}

// GenerateCertificateUpdateCommand generates command to update CA certificates
func GenerateCertificateUpdateCommand(pkgManager string) []string {
	switch pkgManager {
	case "apt", "apt-get":
		return []string{"update-ca-certificates"}
	case "yum", "dnf":
		return []string{"update-ca-trust"}
	case "apk":
		return []string{"update-ca-certificates"}
	case "pacman":
		return []string{"trust", "extract-compat"}
	case "zypper":
		return []string{"update-ca-certificates"}
	default:
		return []string{"echo", "Unknown package manager for certificate update"}
	}
}

// GeneratePackageUpdateCommand generates command to update package manager
func GeneratePackageUpdateCommand(pkgManager string) []string {
	switch pkgManager {
	case "apt", "apt-get":
		return []string{"apt-get", "update", "-y"}
	case "yum":
		return []string{"yum", "makecache"}
	case "dnf":
		return []string{"dnf", "makecache"}
	case "apk":
		return []string{"apk", "update"}
	case "pacman":
		return []string{"pacman", "-Sy"}
	case "zypper":
		return []string{"zypper", "refresh"}
	default:
		return []string{"echo", "Unknown package manager for update"}
	}
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}