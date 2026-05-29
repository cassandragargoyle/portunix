/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProgressCallback is called periodically with current download progress.
type ProgressCallback func(downloaded, total int64, percent int)

// Options controls the download behaviour.
type Options struct {
	URL         string
	Dest        string
	Resume      bool
	Token       string // optional GitHub token for private assets
	ExpectedSHA string // optional SHA-256 (hex) verified after completion
	Progress    ProgressCallback
	Timeout     time.Duration
	UserAgent   string
}

// Result describes a completed download.
type Result struct {
	Path    string
	Size    int64
	SHA256  string
	Resumed bool
}

// Download fetches opts.URL into opts.Dest. If opts.Resume is true and a
// partial file exists at Dest, the download is continued via Range header.
func Download(opts Options) (*Result, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("download: URL required")
	}
	if opts.Dest == "" {
		return nil, fmt.Errorf("download: destination required")
	}
	if err := os.MkdirAll(filepath.Dir(opts.Dest), 0o755); err != nil {
		return nil, fmt.Errorf("create destination dir: %w", err)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Minute
	}
	client := &http.Client{Timeout: timeout}

	var existing int64
	if opts.Resume {
		if info, err := os.Stat(opts.Dest); err == nil {
			existing = info.Size()
		}
	}

	req, err := http.NewRequest(http.MethodGet, opts.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if opts.Token != "" && strings.Contains(opts.URL, "api.github.com") {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
		req.Header.Set("Accept", "application/octet-stream")
	}
	if opts.UserAgent != "" {
		req.Header.Set("User-Agent", opts.UserAgent)
	} else {
		req.Header.Set("User-Agent", "ptx-github")
	}
	if existing > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existing))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	resumed := false
	switch resp.StatusCode {
	case http.StatusOK:
		// server ignored or didn't honor Range — restart from zero
		existing = 0
	case http.StatusPartialContent:
		resumed = existing > 0
	default:
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}

	flags := os.O_WRONLY | os.O_CREATE
	if resumed {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	out, err := os.OpenFile(opts.Dest, flags, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open destination: %w", err)
	}
	defer out.Close()

	total := resp.ContentLength
	if resumed && total > 0 {
		total += existing
	}

	var hasher hash.Hash
	var writer io.Writer = out
	if opts.ExpectedSHA != "" && !resumed {
		// Hash only the bytes we wrote during this run.
		// We can't verify a resumed download, so caller must rerun without resume.
		hasher = sha256.New()
		writer = io.MultiWriter(out, hasher)
	}

	pr := &progressReader{
		reader:     resp.Body,
		downloaded: existing,
		total:      total,
		callback:   opts.Progress,
	}

	if _, err := io.Copy(writer, pr); err != nil {
		return nil, fmt.Errorf("copy body: %w", err)
	}

	finalSize := pr.downloaded
	res := &Result{Path: opts.Dest, Size: finalSize, Resumed: resumed}

	if hasher != nil {
		got := hex.EncodeToString(hasher.Sum(nil))
		res.SHA256 = got
		if !strings.EqualFold(got, opts.ExpectedSHA) {
			_ = os.Remove(opts.Dest)
			return res, fmt.Errorf("checksum mismatch: expected %s, got %s", opts.ExpectedSHA, got)
		}
	}
	return res, nil
}

// VerifyFile computes SHA-256 of a file and compares it against expected (hex).
func VerifyFile(path, expected string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if expected != "" && !strings.EqualFold(got, expected) {
		return got, fmt.Errorf("checksum mismatch: expected %s, got %s", expected, got)
	}
	return got, nil
}

// FormatBytes formats bytes into a human readable suffix.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// SimpleProgressBar prints a console progress bar suitable for ProgressCallback.
func SimpleProgressBar(downloaded, total int64, percent int) {
	const barWidth = 40
	if total <= 0 {
		fmt.Fprintf(os.Stderr, "\rDownloaded %s", FormatBytes(downloaded))
		return
	}
	pos := percent * barWidth / 100
	bar := strings.Repeat("=", pos)
	if pos < barWidth {
		bar += ">"
		bar += strings.Repeat(" ", barWidth-pos-1)
	}
	fmt.Fprintf(os.Stderr, "\r[%s] %3d%% (%s/%s)",
		bar, percent, FormatBytes(downloaded), FormatBytes(total))
	if percent >= 100 {
		fmt.Fprintln(os.Stderr)
	}
}

type progressReader struct {
	reader     io.Reader
	downloaded int64
	total      int64
	callback   ProgressCallback
	lastPct    int
	lastTick   time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)
	if pr.callback != nil {
		now := time.Now()
		percent := 0
		if pr.total > 0 {
			percent = int(float64(pr.downloaded) / float64(pr.total) * 100)
		}
		if err == io.EOF || percent != pr.lastPct || now.Sub(pr.lastTick) > 100*time.Millisecond {
			pr.callback(pr.downloaded, pr.total, percent)
			pr.lastPct = percent
			pr.lastTick = now
		}
	}
	return n, err
}
