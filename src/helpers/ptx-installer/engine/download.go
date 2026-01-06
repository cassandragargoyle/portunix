package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProgressWriter wraps an io.Writer to display download progress
type ProgressWriter struct {
	Total      int64
	Downloaded int64
	LastUpdate time.Time
	StartTime  time.Time
}

// Write implements io.Writer interface with progress display
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Downloaded += int64(n)

	// Update progress every 100ms to avoid too frequent updates
	if time.Since(pw.LastUpdate) > 100*time.Millisecond {
		pw.displayProgress()
		pw.LastUpdate = time.Now()
	}

	return n, nil
}

// displayProgress shows current download progress
func (pw *ProgressWriter) displayProgress() {
	if pw.Total <= 0 {
		// Unknown size - just show downloaded amount
		fmt.Printf("\r⏳ Downloaded: %s", formatBytes(pw.Downloaded))
		return
	}

	percent := float64(pw.Downloaded) / float64(pw.Total) * 100
	elapsed := time.Since(pw.StartTime).Seconds()

	// Calculate speed and ETA
	var speedStr, etaStr string
	if elapsed > 0 {
		speed := float64(pw.Downloaded) / elapsed
		speedStr = fmt.Sprintf("%s/s", formatBytes(int64(speed)))

		if speed > 0 {
			remaining := float64(pw.Total-pw.Downloaded) / speed
			etaStr = formatDuration(remaining)
		}
	}

	// Progress bar (20 chars wide)
	barWidth := 20
	filled := int(percent / 100 * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r⏳ [%s] %.1f%% (%s / %s) %s %s   ",
		bar, percent,
		formatBytes(pw.Downloaded), formatBytes(pw.Total),
		speedStr, etaStr)
}

// Finish completes the progress display
func (pw *ProgressWriter) Finish() {
	if pw.Total > 0 {
		elapsed := time.Since(pw.StartTime).Seconds()
		speed := float64(pw.Downloaded) / elapsed
		fmt.Printf("\r✅ Downloaded: %s in %s (%s/s)                    \n",
			formatBytes(pw.Downloaded),
			formatDuration(elapsed),
			formatBytes(int64(speed)))
	} else {
		fmt.Printf("\r✅ Downloaded: %s                    \n", formatBytes(pw.Downloaded))
	}
}

// formatDuration formats seconds to human-readable duration
func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	minutes := int(seconds) / 60
	secs := int(seconds) % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// DownloadFile downloads a file from URL to the specified filepath with progress
func DownloadFile(destPath string, url string) error {
	fmt.Printf("📥 Downloading from: %s\n", url)

	// Create the file
	out, err := os.Create(destPath)
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

	// Show file size
	size := resp.ContentLength
	if size > 0 {
		fmt.Printf("📦 Size: %s\n", formatBytes(size))
	}

	// Create progress writer
	progress := &ProgressWriter{
		Total:     size,
		StartTime: time.Now(),
		LastUpdate: time.Now(),
	}

	// Download with progress
	_, err = io.Copy(out, io.TeeReader(resp.Body, progress))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	progress.Finish()
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
	destPath := filepath.Join(cacheDir, filename)

	// Create file
	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Show download info
	fmt.Printf("📥 Downloading: %s\n", filename)
	size := resp.ContentLength
	if size > 0 {
		fmt.Printf("📦 Size: %s\n", formatBytes(size))
	}

	// Create progress writer
	progress := &ProgressWriter{
		Total:      size,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}

	// Download with progress
	_, err = io.Copy(out, io.TeeReader(resp.Body, progress))
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	progress.Finish()
	return destPath, nil
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
