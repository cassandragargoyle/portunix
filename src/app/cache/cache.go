package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Category represents a cache category with its own storage limits
type Category string

const (
	CategoryDownloads Category = "downloads"
	CategoryHTTP      Category = "http"
	CategoryBuilds    Category = "builds"
	CategoryMetadata  Category = "metadata"
)

// AllCategories returns all available cache categories
func AllCategories() []Category {
	return []Category{CategoryDownloads, CategoryHTTP, CategoryBuilds, CategoryMetadata}
}

// DefaultMaxSize is the default maximum cache size in bytes (1GB)
const DefaultMaxSize int64 = 1 << 30

// DefaultTTL is the default time-to-live for cache entries
const DefaultTTL = 7 * 24 * time.Hour

// CategoryConfig holds configuration for a single cache category
type CategoryConfig struct {
	MaxSize int64         `json:"max_size"`
	TTL     time.Duration `json:"ttl"`
}

// Config holds cache configuration
type Config struct {
	Enabled    bool                        `json:"enabled"`
	BaseDir    string                      `json:"directory"`
	MaxSize    int64                       `json:"max_size"`
	Categories map[Category]CategoryConfig `json:"categories"`
}

// DefaultConfig returns the default cache configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled: true,
		BaseDir: DefaultCacheDir(),
		MaxSize: DefaultMaxSize,
		Categories: map[Category]CategoryConfig{
			CategoryDownloads: {MaxSize: 500 << 20, TTL: 7 * 24 * time.Hour},
			CategoryHTTP:      {MaxSize: 100 << 20, TTL: 1 * time.Hour},
			CategoryBuilds:    {MaxSize: 300 << 20, TTL: 24 * time.Hour},
			CategoryMetadata:  {MaxSize: 50 << 20, TTL: 1 * time.Hour},
		},
	}
}

// LoadConfig loads cache configuration from environment variables,
// falling back to defaults
func LoadConfig() *Config {
	cfg := DefaultConfig()

	if dir := os.Getenv("PORTUNIX_CACHE_DIR"); dir != "" {
		cfg.BaseDir = dir
	}

	if disabled := os.Getenv("PORTUNIX_CACHE_DISABLED"); disabled == "true" || disabled == "1" {
		cfg.Enabled = false
	}

	return cfg
}

// DefaultCacheDir returns the platform-specific default cache directory
func DefaultCacheDir() string {
	switch runtime.GOOS {
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "portunix", "Cache")
		}
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			return filepath.Join(userProfile, ".cache", "portunix")
		}
		return filepath.Join("C:\\", "Users", "Public", ".cache", "portunix")
	default:
		// Linux/macOS: respect XDG_CACHE_HOME
		if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
			return filepath.Join(xdgCache, "portunix")
		}
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".cache", "portunix")
		}
		return filepath.Join(os.TempDir(), "portunix-cache")
	}
}

// Manager provides centralized cache management
type Manager struct {
	config *Config
}

// NewManager creates a new cache manager
func NewManager() *Manager {
	return &Manager{
		config: LoadConfig(),
	}
}

// NewManagerWithConfig creates a new cache manager with explicit config
func NewManagerWithConfig(cfg *Config) *Manager {
	return &Manager{config: cfg}
}

// IsEnabled returns whether caching is enabled
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}

// BaseDir returns the base cache directory
func (m *Manager) BaseDir() string {
	return m.config.BaseDir
}

// CategoryDir returns the directory path for a specific cache category
func (m *Manager) CategoryDir(cat Category) string {
	return filepath.Join(m.config.BaseDir, string(cat))
}

// EnsureDirs creates the cache directory structure if it doesn't exist
func (m *Manager) EnsureDirs() error {
	for _, cat := range AllCategories() {
		dir := m.CategoryDir(cat)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory %s: %w", dir, err)
		}
	}

	// Create locks directory
	locksDir := filepath.Join(m.config.BaseDir, "locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		return fmt.Errorf("failed to create locks directory: %w", err)
	}

	return nil
}

// CacheKey generates a SHA256-based cache key from a URL or identifier
func CacheKey(identifier string) string {
	h := sha256.Sum256([]byte(identifier))
	return hex.EncodeToString(h[:])
}

// Info holds statistics about the cache
type Info struct {
	BaseDir    string                    `json:"base_dir"`
	TotalSize  int64                     `json:"total_size"`
	TotalFiles int                       `json:"total_files"`
	MaxSize    int64                     `json:"max_size"`
	Enabled    bool                      `json:"enabled"`
	Categories map[Category]CategoryInfo `json:"categories"`
}

// CategoryInfo holds statistics for a single cache category
type CategoryInfo struct {
	Dir     string `json:"dir"`
	Size    int64  `json:"size"`
	Files   int    `json:"files"`
	MaxSize int64  `json:"max_size"`
}

// GetInfo returns information about the cache
func (m *Manager) GetInfo() (*Info, error) {
	info := &Info{
		BaseDir:    m.config.BaseDir,
		MaxSize:    m.config.MaxSize,
		Enabled:    m.config.Enabled,
		Categories: make(map[Category]CategoryInfo),
	}

	for _, cat := range AllCategories() {
		catDir := m.CategoryDir(cat)
		catCfg := m.config.Categories[cat]
		catInfo := CategoryInfo{
			Dir:     catDir,
			MaxSize: catCfg.MaxSize,
		}

		size, files, err := dirStats(catDir)
		if err == nil {
			catInfo.Size = size
			catInfo.Files = files
		}

		info.Categories[cat] = catInfo
		info.TotalSize += catInfo.Size
		info.TotalFiles += catInfo.Files
	}

	return info, nil
}

// dirStats calculates total size and file count for a directory
func dirStats(dir string) (int64, int, error) {
	var totalSize int64
	var totalFiles int

	err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible files
		}
		if !fi.IsDir() {
			totalSize += fi.Size()
			totalFiles++
		}
		return nil
	})

	return totalSize, totalFiles, err
}
