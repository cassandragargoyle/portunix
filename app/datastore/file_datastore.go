package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// FileDatastore implements DatastoreInterface using local filesystem
type FileDatastore struct {
	basePath           string
	format             string
	createDirectories  bool
	backupEnabled      bool
	mu                 sync.RWMutex
	stats              *Stats
	startTime          time.Time
}

// FileEntry represents a stored file entry with metadata
type FileEntry struct {
	Key       string                 `json:"key" yaml:"key"`
	Value     interface{}            `json:"value" yaml:"value"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" yaml:"updated_at"`
}

// NewFileDatastore creates a new file-based datastore
func NewFileDatastore() *FileDatastore {
	return &FileDatastore{
		stats: &Stats{
			TotalKeys:   0,
			TotalSize:   0,
			Collections: make(map[string]int64),
			Performance: &PerformanceStats{
				AverageReadTime:  0,
				AverageWriteTime: 0,
				OperationsPerSec: 0,
				ErrorRate:        0,
			},
			LastUpdated: time.Now(),
		},
		startTime: time.Now(),
	}
}

// Initialize initializes the file datastore with configuration
func (f *FileDatastore) Initialize(ctx context.Context, config map[string]interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Set default values
	f.basePath = "~/.portunix/data"
	f.format = "yaml"
	f.createDirectories = true
	f.backupEnabled = false

	// Parse configuration
	if basePath, ok := config["base_path"].(string); ok {
		f.basePath = expandPath(basePath)
	}
	if format, ok := config["format"].(string); ok {
		f.format = format
	}
	if createDirs, ok := config["create_directories"].(bool); ok {
		f.createDirectories = createDirs
	}
	if backup, ok := config["backup_enabled"].(bool); ok {
		f.backupEnabled = backup
	}

	// Create base directory if it doesn't exist
	if f.createDirectories {
		if err := os.MkdirAll(f.basePath, 0755); err != nil {
			return fmt.Errorf("failed to create base directory: %w", err)
		}
	}

	// Initialize stats
	f.updateStats()

	return nil
}

// Store stores data to a file
func (f *FileDatastore) Store(ctx context.Context, key string, value interface{}, metadata map[string]interface{}) error {
	start := time.Now()
	defer func() {
		f.updatePerformanceStats("write", time.Since(start))
	}()

	f.mu.Lock()
	defer f.mu.Unlock()

	filePath := f.getFilePath(key)
	
	// Create directory if needed
	if f.createDirectories {
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create file entry
	entry := FileEntry{
		Key:       key,
		Value:     value,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Check if file exists to determine if this is an update
	if _, err := os.Stat(filePath); err == nil {
		// File exists, load existing entry to preserve creation time
		if existingEntry, err := f.loadEntry(filePath); err == nil {
			entry.CreatedAt = existingEntry.CreatedAt
		}
	}

	// Backup existing file if enabled
	if f.backupEnabled {
		if err := f.backupFile(filePath); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to backup file %s: %v\n", filePath, err)
		}
	}

	// Write file
	if err := f.writeEntry(filePath, &entry); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	f.updateStats()
	return nil
}

// Retrieve retrieves data from a file
func (f *FileDatastore) Retrieve(ctx context.Context, key string, filter map[string]interface{}) (interface{}, error) {
	start := time.Now()
	defer func() {
		f.updatePerformanceStats("read", time.Since(start))
	}()

	f.mu.RLock()
	defer f.mu.RUnlock()

	filePath := f.getFilePath(key)
	
	entry, err := f.loadEntry(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load entry: %w", err)
	}

	// Apply filter if provided
	if filter != nil && !f.matchesFilter(entry, filter) {
		return nil, fmt.Errorf("entry does not match filter")
	}

	return entry.Value, nil
}

// Query queries files based on criteria
func (f *FileDatastore) Query(ctx context.Context, criteria QueryCriteria) ([]QueryResult, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var results []QueryResult
	
	// Determine search path based on collection
	searchPath := f.basePath
	if criteria.Collection != "" {
		searchPath = filepath.Join(f.basePath, criteria.Collection)
	}

	// Walk through all files
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil // Skip directories
		}

		// Skip non-data files
		if !f.isDataFile(path) {
			return nil
		}

		entry, err := f.loadEntry(path)
		if err != nil {
			return nil // Skip files that can't be loaded
		}

		// Apply filter
		if criteria.Filter != nil && !f.matchesFilter(entry, criteria.Filter) {
			return nil
		}

		// Convert file path back to key
		key := f.pathToKey(path)
		
		results = append(results, QueryResult{
			Key:      key,
			Value:    entry.Value,
			Metadata: entry.Metadata,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}

	// Apply sorting, limit, and offset
	results = f.applySorting(results, criteria.Sort)
	results = f.applyPagination(results, criteria.Limit, criteria.Offset)

	return results, nil
}

// Delete deletes a file
func (f *FileDatastore) Delete(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	filePath := f.getFilePath(key)
	
	// Backup before deletion if enabled
	if f.backupEnabled {
		if err := f.backupFile(filePath); err != nil {
			fmt.Printf("Warning: failed to backup file before deletion %s: %v\n", filePath, err)
		}
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	f.updateStats()
	return nil
}

// List lists all keys matching a pattern
func (f *FileDatastore) List(ctx context.Context, pattern string) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var keys []string
	
	err := filepath.Walk(f.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		if !f.isDataFile(path) {
			return nil
		}

		key := f.pathToKey(path)
		if matchPattern(pattern, key) {
			keys = append(keys, key)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return keys, nil
}

// Health returns health status of the file datastore
func (f *FileDatastore) Health(ctx context.Context) (*HealthStatus, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Check if base directory is accessible
	_, err := os.Stat(f.basePath)
	healthy := err == nil

	status := "healthy"
	message := "File datastore is operational"
	if !healthy {
		status = "unhealthy"
		message = fmt.Sprintf("Base directory not accessible: %v", err)
	}

	return &HealthStatus{
		Healthy:   healthy,
		Status:    status,
		Message:   message,
		Uptime:    time.Since(f.startTime),
		LastCheck: time.Now(),
		Metrics: map[string]string{
			"base_path":   f.basePath,
			"format":      f.format,
			"total_keys":  fmt.Sprintf("%d", f.stats.TotalKeys),
			"total_size":  fmt.Sprintf("%d", f.stats.TotalSize),
		},
	}, nil
}

// Stats returns statistics about the file datastore
func (f *FileDatastore) Stats(ctx context.Context) (*Stats, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	f.updateStats()
	return f.stats, nil
}

// Close closes the file datastore (no-op for file system)
func (f *FileDatastore) Close(ctx context.Context) error {
	return nil
}

// Helper methods

func (f *FileDatastore) getFilePath(key string) string {
	// Replace path separators in key to create safe file path
	safePath := strings.ReplaceAll(key, "/", string(os.PathSeparator))
	
	// Add file extension based on format
	ext := "." + f.format
	if !strings.HasSuffix(safePath, ext) {
		safePath += ext
	}
	
	return filepath.Join(f.basePath, safePath)
}

func (f *FileDatastore) pathToKey(path string) string {
	// Convert file path back to key
	relPath, _ := filepath.Rel(f.basePath, path)
	
	// Remove file extension
	ext := filepath.Ext(relPath)
	if ext != "" {
		relPath = strings.TrimSuffix(relPath, ext)
	}
	
	// Convert path separators back to forward slashes
	return strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
}

func (f *FileDatastore) isDataFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yaml" || ext == ".yml" || ext == ".json"
}

func (f *FileDatastore) writeEntry(filePath string, entry *FileEntry) error {
	var data []byte
	var err error

	switch f.format {
	case "json":
		data, err = json.MarshalIndent(entry, "", "  ")
	case "yaml", "yml":
		data, err = yaml.Marshal(entry)
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (f *FileDatastore) loadEntry(filePath string) (*FileEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var entry FileEntry
	ext := filepath.Ext(filePath)
	
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &entry)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &entry)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return &entry, err
}

func (f *FileDatastore) matchesFilter(entry *FileEntry, filter map[string]interface{}) bool {
	// Simple filter matching - can be enhanced later
	for key, expectedValue := range filter {
		if key == "metadata" {
			// Check metadata filters
			if metadataFilter, ok := expectedValue.(map[string]interface{}); ok {
				for metaKey, metaValue := range metadataFilter {
					if entry.Metadata == nil {
						return false
					}
					if actualValue, exists := entry.Metadata[metaKey]; !exists || actualValue != metaValue {
						return false
					}
				}
			}
		}
		// Add more filter types as needed
	}
	return true
}

func (f *FileDatastore) applySorting(results []QueryResult, sort map[string]int) []QueryResult {
	// TODO: Implement sorting
	return results
}

func (f *FileDatastore) applyPagination(results []QueryResult, limit, offset int) []QueryResult {
	if offset > 0 && offset < len(results) {
		results = results[offset:]
	}
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}
	return results
}

func (f *FileDatastore) updateStats() {
	// Count files and calculate total size
	var totalKeys int64
	var totalSize int64
	collections := make(map[string]int64)

	filepath.Walk(f.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !f.isDataFile(path) {
			return nil
		}

		totalKeys++
		totalSize += info.Size()

		// Count collections (subdirectories)
		relPath, _ := filepath.Rel(f.basePath, path)
		dir := filepath.Dir(relPath)
		if dir != "." {
			collections[dir]++
		}

		return nil
	})

	f.stats.TotalKeys = totalKeys
	f.stats.TotalSize = totalSize
	f.stats.Collections = collections
	f.stats.LastUpdated = time.Now()
}

func (f *FileDatastore) updatePerformanceStats(operation string, duration time.Duration) {
	// TODO: Implement proper performance tracking with moving averages
	if operation == "read" {
		f.stats.Performance.AverageReadTime = duration
	} else if operation == "write" {
		f.stats.Performance.AverageWriteTime = duration
	}
}

func (f *FileDatastore) backupFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // No file to backup
	}

	backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")
	
	input, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, input, 0644)
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}