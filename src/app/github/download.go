package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadManager handles downloading files with progress reporting
type DownloadManager struct {
	client *http.Client
}

// ProgressCallback is called during download to report progress
type ProgressCallback func(downloaded, total int64, percentage int)

// NewDownloadManager creates a new download manager
func NewDownloadManager() *DownloadManager {
	return &DownloadManager{
		client: &http.Client{
			Timeout: 30 * time.Minute, // Long timeout for large files
		},
	}
}

// DownloadFile downloads a file from URL to destination
func (d *DownloadManager) DownloadFile(url, dest string) error {
	return d.DownloadWithProgress(url, dest, nil)
}

// DownloadWithProgress downloads a file with progress reporting
func (d *DownloadManager) DownloadWithProgress(url, dest string, progress ProgressCallback) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create the HTTP request
	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create destination file
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Get total file size
	total := resp.ContentLength

	// Create progress reader if callback provided
	var reader io.Reader = resp.Body
	if progress != nil && total > 0 {
		reader = &progressReader{
			reader:   resp.Body,
			total:    total,
			callback: progress,
		}
	}

	// Copy data
	_, err = io.Copy(out, reader)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// DownloadAsset downloads a GitHub release asset
func (d *DownloadManager) DownloadAsset(asset *Asset, dest string) error {
	return d.DownloadFile(asset.DownloadURL, dest)
}

// DownloadAssetWithProgress downloads a GitHub release asset with progress
func (d *DownloadManager) DownloadAssetWithProgress(asset *Asset, dest string, progress ProgressCallback) error {
	return d.DownloadWithProgress(asset.DownloadURL, dest, progress)
}

// progressReader wraps an io.Reader to provide progress callbacks
type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	callback   ProgressCallback
	lastUpdate time.Time
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.downloaded += int64(n)

	// Call progress callback every 100ms to avoid too frequent updates
	now := time.Now()
	if pr.callback != nil && (now.Sub(pr.lastUpdate) > 100*time.Millisecond || err == io.EOF) {
		percentage := int(float64(pr.downloaded) / float64(pr.total) * 100)
		pr.callback(pr.downloaded, pr.total, percentage)
		pr.lastUpdate = now
	}

	return n, err
}

// FormatBytes formats bytes into human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Simple progress bar for console output
func SimpleProgressBar(downloaded, total int64, percentage int) {
	const barWidth = 50
	pos := percentage * barWidth / 100
	
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < pos {
			bar += "="
		} else if i == pos {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %d%% (%s/%s)", percentage, FormatBytes(downloaded), FormatBytes(total))
	
	fmt.Printf("\r%s", bar)
	if percentage >= 100 {
		fmt.Println() // New line when complete
	}
}