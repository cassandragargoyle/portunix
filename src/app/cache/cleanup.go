package cache

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"os"
	"sort"
	"time"
)

// CleanResult holds results of a cleanup operation
type CleanResult struct {
	RemovedFiles int   `json:"removed_files"`
	FreedBytes   int64 `json:"freed_bytes"`
}

// Clean removes expired and invalid cache entries
func (m *Manager) Clean() (*CleanResult, error) {
	result := &CleanResult{}

	for _, cat := range AllCategories() {
		entries, err := m.ListEntries(cat)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsExpired() {
				result.FreedBytes += entry.Size
				result.RemovedFiles++
				_ = m.Remove(entry.Category, entry.Key)
				continue
			}

			// Verify file still exists
			filePath := m.entryFilePath(cat, entry.Key, entry.Filename)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				result.RemovedFiles++
				_ = m.Remove(cat, entry.Key)
			}
		}
	}

	return result, nil
}

// Purge removes all cache contents
func (m *Manager) Purge() (*CleanResult, error) {
	info, err := m.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache info: %w", err)
	}

	result := &CleanResult{
		RemovedFiles: info.TotalFiles,
		FreedBytes:   info.TotalSize,
	}

	for _, cat := range AllCategories() {
		catDir := m.CategoryDir(cat)
		if err := os.RemoveAll(catDir); err != nil {
			return result, fmt.Errorf("failed to purge %s: %w", cat, err)
		}
	}

	return result, nil
}

// RemoveBySource removes all cache entries matching a source pattern
func (m *Manager) RemoveBySource(pattern string) (*CleanResult, error) {
	result := &CleanResult{}

	for _, cat := range AllCategories() {
		entries, err := m.ListEntries(cat)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if matchSource(entry.Source, pattern) || matchSource(entry.Filename, pattern) {
				result.FreedBytes += entry.Size
				result.RemovedFiles++
				_ = m.Remove(cat, entry.Key)
			}
		}
	}

	return result, nil
}

// EnforceSize removes oldest entries until total size is under the limit
func (m *Manager) EnforceSize() (*CleanResult, error) {
	result := &CleanResult{}

	// Enforce per-category limits
	for _, cat := range AllCategories() {
		catCfg, ok := m.config.Categories[cat]
		if !ok {
			continue
		}

		entries, err := m.ListEntries(cat)
		if err != nil || len(entries) == 0 {
			continue
		}

		var catSize int64
		for _, e := range entries {
			catSize += e.Size
		}

		if catSize <= catCfg.MaxSize {
			continue
		}

		// Sort by creation time (oldest first)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].CreatedAt.Before(entries[j].CreatedAt)
		})

		// Remove oldest entries until under limit
		for _, entry := range entries {
			if catSize <= catCfg.MaxSize {
				break
			}
			catSize -= entry.Size
			result.FreedBytes += entry.Size
			result.RemovedFiles++
			_ = m.Remove(cat, entry.Key)
		}
	}

	return result, nil
}

// matchSource checks if a source string contains the pattern (case-insensitive substring)
func matchSource(source, pattern string) bool {
	if pattern == "" {
		return false
	}
	// Simple substring match
	return containsIgnoreCase(source, pattern)
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)

	return len(sLower) >= len(substrLower) && contains(sLower, substrLower)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// newSHA256 returns a new SHA256 hash
func newSHA256() hash.Hash {
	return sha256.New()
}

// FormatSize formats bytes into human-readable string
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatDuration formats a duration into human-readable string
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "expired"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}
