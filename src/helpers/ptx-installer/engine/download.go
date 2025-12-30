package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadFile downloads a file from URL to the specified filepath
func DownloadFile(filepath string, url string) error {
	fmt.Printf("📥 Downloading from: %s\n", url)

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file with progress
	size := resp.ContentLength
	if size > 0 {
		fmt.Printf("📦 File size: %.2f MB\n", float64(size)/(1024*1024))
	}

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Downloaded: %.2f MB\n", float64(written)/(1024*1024))
	return nil
}

// DownloadFileWithProperFilename downloads a file and determines the proper filename
func DownloadFileWithProperFilename(url string, cacheDir string) (string, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Get the response to check headers
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Extract filename from response
	filename := ExtractFilenameFromResponse(resp)
	if filename == "" {
		// Fallback: extract from URL
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			filename = parts[len(parts)-1]
		}
		if filename == "" {
			filename = "download"
		}
	}

	// Create full path
	filepath := filepath.Join(cacheDir, filename)

	// Create file
	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Download
	fmt.Printf("📥 Downloading: %s\n", filename)
	size := resp.ContentLength
	if size > 0 {
		fmt.Printf("📦 Size: %.2f MB\n", float64(size)/(1024*1024))
	}

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Downloaded: %s (%.2f MB)\n", filename, float64(written)/(1024*1024))
	return filepath, nil
}

// ExtractFilenameFromResponse extracts filename from HTTP response
func ExtractFilenameFromResponse(resp *http.Response) string {
	// Try Content-Disposition header first
	if disposition := resp.Header.Get("Content-Disposition"); disposition != "" {
		if filename := ParseContentDisposition(disposition); filename != "" {
			return filename
		}
	}

	// Try to extract from URL
	if resp.Request != nil && resp.Request.URL != nil {
		path := resp.Request.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			if filename != "" && strings.Contains(filename, ".") {
				return filename
			}
		}
	}

	// Generate from content type
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		return GenerateFilenameFromContentType(contentType)
	}

	return ""
}

// ParseContentDisposition parses Content-Disposition header
func ParseContentDisposition(disposition string) string {
	// Look for filename= or filename*=
	parts := strings.Split(disposition, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			filename = strings.Trim(filename, "\"'")
			return filename
		}
		if strings.HasPrefix(part, "filename*=") {
			// Handle RFC 5987 encoded filenames
			filename := strings.TrimPrefix(part, "filename*=")
			if idx := strings.Index(filename, "''"); idx != -1 {
				filename = filename[idx+2:]
			}
			filename = strings.Trim(filename, "\"'")
			return filename
		}
	}
	return ""
}

// GenerateFilenameFromContentType generates a filename based on content type
func GenerateFilenameFromContentType(contentType string) string {
	// Remove parameters
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	contentType = strings.TrimSpace(contentType)

	// Map common content types to extensions
	extensions := map[string]string{
		"application/zip":             ".zip",
		"application/x-tar":           ".tar",
		"application/gzip":            ".gz",
		"application/x-gzip":          ".gz",
		"application/x-compressed":    ".tar.gz",
		"application/x-msdownload":    ".exe",
		"application/vnd.ms-cab":      ".cab",
		"application/octet-stream":    ".bin",
	}

	if ext, ok := extensions[contentType]; ok {
		return "download" + ext
	}

	return "download"
}
