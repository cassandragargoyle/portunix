package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// EntryMeta holds metadata for a cached item
type EntryMeta struct {
	Key       string    `json:"key"`
	Category  Category  `json:"category"`
	Source     string    `json:"source"`     // original URL or identifier
	Filename  string    `json:"filename"`   // original filename
	Size      int64     `json:"size"`
	Checksum  string    `json:"checksum"`   // SHA256 checksum
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired returns true if the entry has passed its expiration time
func (e *EntryMeta) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}

// Get retrieves a cached file path by key and category.
// Returns the file path and metadata if found and not expired,
// or empty string if not cached or expired.
func (m *Manager) Get(cat Category, key string) (string, *EntryMeta, error) {
	if !m.config.Enabled {
		return "", nil, nil
	}

	meta, err := m.loadMeta(cat, key)
	if err != nil {
		return "", nil, nil // not found
	}

	if meta.IsExpired() {
		// Clean up expired entry
		_ = m.Remove(cat, key)
		return "", nil, nil
	}

	filePath := m.entryFilePath(cat, key, meta.Filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Metadata exists but file is missing, clean up
		_ = m.removeMeta(cat, key)
		return "", nil, nil
	}

	return filePath, meta, nil
}

// Store copies a file into the cache under the given category and key
func (m *Manager) Store(cat Category, key string, sourcePath string, source string) (*EntryMeta, error) {
	if !m.config.Enabled {
		return nil, nil
	}

	if err := m.EnsureDirs(); err != nil {
		return nil, err
	}

	srcInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("source file not accessible: %w", err)
	}

	filename := filepath.Base(sourcePath)
	destPath := m.entryFilePath(cat, key, filename)

	// Create entry directory
	entryDir := m.entryDir(cat, key)
	if err := os.MkdirAll(entryDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create entry directory: %w", err)
	}

	// Copy file to cache
	if err := copyFile(sourcePath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy to cache: %w", err)
	}

	// Calculate checksum
	checksum, err := fileChecksum(destPath)
	if err != nil {
		checksum = "" // non-fatal
	}

	// Determine TTL
	catCfg, ok := m.config.Categories[cat]
	ttl := DefaultTTL
	if ok && catCfg.TTL > 0 {
		ttl = catCfg.TTL
	}

	meta := &EntryMeta{
		Key:       key,
		Category:  cat,
		Source:    source,
		Filename:  filename,
		Size:      srcInfo.Size(),
		Checksum:  checksum,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	if err := m.saveMeta(cat, key, meta); err != nil {
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return meta, nil
}

// Remove removes a cached entry by key and category
func (m *Manager) Remove(cat Category, key string) error {
	entryDir := m.entryDir(cat, key)
	return os.RemoveAll(entryDir)
}

// ListEntries returns all cache entries for a given category
func (m *Manager) ListEntries(cat Category) ([]*EntryMeta, error) {
	catDir := m.CategoryDir(cat)
	entries, err := os.ReadDir(catDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []*EntryMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := m.loadMeta(cat, entry.Name())
		if err != nil {
			continue // skip entries without valid metadata
		}
		result = append(result, meta)
	}

	return result, nil
}

// ListAllEntries returns all cache entries across all categories
func (m *Manager) ListAllEntries() (map[Category][]*EntryMeta, error) {
	result := make(map[Category][]*EntryMeta)
	for _, cat := range AllCategories() {
		entries, err := m.ListEntries(cat)
		if err != nil {
			continue
		}
		if len(entries) > 0 {
			result[cat] = entries
		}
	}
	return result, nil
}

// entryDir returns the directory for a specific cache entry
func (m *Manager) entryDir(cat Category, key string) string {
	return filepath.Join(m.CategoryDir(cat), key)
}

// entryFilePath returns the full path for a cached file
func (m *Manager) entryFilePath(cat Category, key string, filename string) string {
	return filepath.Join(m.entryDir(cat, key), filename)
}

// metaFilePath returns the path for the metadata JSON file
func (m *Manager) metaFilePath(cat Category, key string) string {
	return filepath.Join(m.entryDir(cat, key), ".meta.json")
}

// saveMeta saves entry metadata as JSON
func (m *Manager) saveMeta(cat Category, key string, meta *EntryMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.metaFilePath(cat, key), data, 0644)
}

// loadMeta loads entry metadata from JSON
func (m *Manager) loadMeta(cat Category, key string) (*EntryMeta, error) {
	data, err := os.ReadFile(m.metaFilePath(cat, key))
	if err != nil {
		return nil, err
	}

	var meta EntryMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// removeMeta removes only the metadata file
func (m *Manager) removeMeta(cat Category, key string) error {
	return os.Remove(m.metaFilePath(cat, key))
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// fileChecksum calculates SHA256 checksum of a file
func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := newSHA256()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
