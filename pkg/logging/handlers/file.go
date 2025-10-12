package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileHandler handles file output with rotation support
type FileHandler struct {
	config   FileConfig
	file     *os.File
	mutex    sync.Mutex
	size     int64
	rotateAt int64
}

// FileConfig holds configuration for file output
type FileConfig struct {
	FilePath   string // Path to log file
	MaxSize    int64  // Maximum size in bytes before rotation
	MaxAge     int    // Maximum age in days
	MaxBackups int    // Maximum number of backup files
	Compress   bool   // Compress rotated files
	JSONFormat bool   // Use JSON format
	Append     bool   // Append to existing file
}

// DefaultFileConfig returns default file configuration
func DefaultFileConfig(filePath string) FileConfig {
	return FileConfig{
		FilePath:   filePath,
		MaxSize:    100 * 1024 * 1024, // 100 MB
		MaxAge:     30,                 // 30 days
		MaxBackups: 10,
		Compress:   false,
		JSONFormat: true,
		Append:     true,
	}
}

// NewFileHandler creates a new file handler
func NewFileHandler(config FileConfig) (*FileHandler, error) {
	handler := &FileHandler{
		config:   config,
		rotateAt: config.MaxSize,
	}

	if err := handler.openFile(); err != nil {
		return nil, err
	}

	return handler, nil
}

// openFile opens the log file for writing
func (h *FileHandler) openFile() error {
	// Ensure directory exists
	dir := filepath.Dir(h.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}

	// Determine open flags
	flags := os.O_CREATE | os.O_WRONLY
	if h.config.Append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	// Open file
	file, err := os.OpenFile(h.config.FilePath, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", h.config.FilePath, err)
	}

	// Get current file size
	if stat, err := file.Stat(); err == nil {
		h.size = stat.Size()
	}

	h.file = file
	return nil
}

// Write implements io.Writer interface
func (h *FileHandler) Write(p []byte) (n int, err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check if rotation is needed
	if h.shouldRotate(len(p)) {
		if err := h.rotate(); err != nil {
			// Log rotation failed, but continue writing to current file
			fmt.Fprintf(os.Stderr, "Log rotation failed: %v\n", err)
		}
	}

	// Write to file
	n, err = h.file.Write(p)
	if err != nil {
		return n, err
	}

	h.size += int64(n)
	return n, nil
}

// shouldRotate checks if log rotation is needed
func (h *FileHandler) shouldRotate(writeSize int) bool {
	if h.config.MaxSize <= 0 {
		return false
	}
	return h.size+int64(writeSize) > h.rotateAt
}

// rotate performs log rotation
func (h *FileHandler) rotate() error {
	// Close current file
	if h.file != nil {
		h.file.Close()
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupPath := fmt.Sprintf("%s.%s", h.config.FilePath, timestamp)

	// Rename current file to backup
	if err := os.Rename(h.config.FilePath, backupPath); err != nil {
		// If rename fails, try to continue with new file
		fmt.Fprintf(os.Stderr, "Failed to rename log file for rotation: %v\n", err)
	}

	// Compress backup if configured
	if h.config.Compress {
		go func() {
			if err := h.compressFile(backupPath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to compress rotated log: %v\n", err)
			}
		}()
	}

	// Clean up old backups
	go h.cleanupOldBackups()

	// Open new file
	h.size = 0
	return h.openFile()
}

// compressFile compresses a file using gzip (placeholder for now)
func (h *FileHandler) compressFile(filePath string) error {
	// TODO: Implement gzip compression
	// For now, just rename with .gz extension
	gzPath := filePath + ".gz"
	return os.Rename(filePath, gzPath)
}

// cleanupOldBackups removes old backup files based on MaxAge and MaxBackups
func (h *FileHandler) cleanupOldBackups() {
	if h.config.MaxBackups <= 0 && h.config.MaxAge <= 0 {
		return
	}

	dir := filepath.Dir(h.config.FilePath)
	baseName := filepath.Base(h.config.FilePath)

	// Find all backup files
	files, err := filepath.Glob(filepath.Join(dir, baseName+".*"))
	if err != nil {
		return
	}

	// Sort files by modification time (newest first)
	fileInfos := make([]fileInfo, 0, len(files))
	for _, file := range files {
		if stat, err := os.Stat(file); err == nil {
			fileInfos = append(fileInfos, fileInfo{
				path:    file,
				modTime: stat.ModTime(),
			})
		}
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(fileInfos)-1; i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.Before(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	now := time.Now()
	for i, info := range fileInfos {
		shouldDelete := false

		// Check MaxBackups
		if h.config.MaxBackups > 0 && i >= h.config.MaxBackups {
			shouldDelete = true
		}

		// Check MaxAge
		if h.config.MaxAge > 0 {
			age := now.Sub(info.modTime)
			if age > time.Duration(h.config.MaxAge)*24*time.Hour {
				shouldDelete = true
			}
		}

		if shouldDelete {
			os.Remove(info.path)
		}
	}
}

// fileInfo holds file information for sorting
type fileInfo struct {
	path    string
	modTime time.Time
}

// Close closes the file handler
func (h *FileHandler) Close() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.file != nil {
		err := h.file.Close()
		h.file = nil
		return err
	}
	return nil
}

// Sync flushes any buffered data to disk
func (h *FileHandler) Sync() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.file != nil {
		return h.file.Sync()
	}
	return nil
}

// GetFilePath returns the current log file path
func (h *FileHandler) GetFilePath() string {
	return h.config.FilePath
}

// GetSize returns the current file size
func (h *FileHandler) GetSize() int64 {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.size
}

// FileHandlerBuilder provides a fluent interface for building file handlers
type FileHandlerBuilder struct {
	config FileConfig
}

// NewFileBuilder creates a new file handler builder
func NewFileBuilder(filePath string) *FileHandlerBuilder {
	return &FileHandlerBuilder{
		config: DefaultFileConfig(filePath),
	}
}

// WithMaxSize sets the maximum file size before rotation
func (b *FileHandlerBuilder) WithMaxSize(maxSize int64) *FileHandlerBuilder {
	b.config.MaxSize = maxSize
	return b
}

// WithMaxSizeMB sets the maximum file size in megabytes
func (b *FileHandlerBuilder) WithMaxSizeMB(maxSizeMB int) *FileHandlerBuilder {
	b.config.MaxSize = int64(maxSizeMB) * 1024 * 1024
	return b
}

// WithMaxAge sets the maximum age for log files
func (b *FileHandlerBuilder) WithMaxAge(maxAge int) *FileHandlerBuilder {
	b.config.MaxAge = maxAge
	return b
}

// WithMaxBackups sets the maximum number of backup files
func (b *FileHandlerBuilder) WithMaxBackups(maxBackups int) *FileHandlerBuilder {
	b.config.MaxBackups = maxBackups
	return b
}

// WithCompress enables compression for rotated files
func (b *FileHandlerBuilder) WithCompress(compress bool) *FileHandlerBuilder {
	b.config.Compress = compress
	return b
}

// WithJSON enables JSON format output
func (b *FileHandlerBuilder) WithJSON(json bool) *FileHandlerBuilder {
	b.config.JSONFormat = json
	return b
}

// WithAppend sets append mode
func (b *FileHandlerBuilder) WithAppend(append bool) *FileHandlerBuilder {
	b.config.Append = append
	return b
}

// Build creates the file handler
func (b *FileHandlerBuilder) Build() (*FileHandler, error) {
	return NewFileHandler(b.config)
}

// Helper functions for common configurations

// NewRotatingFileHandler creates a file handler with rotation
func NewRotatingFileHandler(filePath string, maxSizeMB, maxAge, maxBackups int) (*FileHandler, error) {
	return NewFileBuilder(filePath).
		WithMaxSizeMB(maxSizeMB).
		WithMaxAge(maxAge).
		WithMaxBackups(maxBackups).
		WithCompress(true).
		Build()
}

// NewSimpleFileHandler creates a simple file handler without rotation
func NewSimpleFileHandler(filePath string) (*FileHandler, error) {
	return NewFileBuilder(filePath).
		WithMaxSize(0). // Disable rotation
		Build()
}

// NewDailyFileHandler creates a file handler suitable for daily logs
func NewDailyFileHandler(filePath string) (*FileHandler, error) {
	return NewFileBuilder(filePath).
		WithMaxSizeMB(50).
		WithMaxAge(7).
		WithMaxBackups(7).
		Build()
}