/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
 
package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.BaseDir = dir
	return cfg
}

func TestDefaultCacheDir(t *testing.T) {
	dir := DefaultCacheDir()
	if dir == "" {
		t.Error("DefaultCacheDir returned empty string")
	}
}

func TestCacheKey(t *testing.T) {
	key1 := CacheKey("https://example.com/file.tar.gz")
	key2 := CacheKey("https://example.com/file.tar.gz")
	key3 := CacheKey("https://example.com/other.tar.gz")

	if key1 != key2 {
		t.Error("Same input should produce same key")
	}
	if key1 == key3 {
		t.Error("Different inputs should produce different keys")
	}
	if len(key1) != 64 {
		t.Errorf("Key should be 64 hex chars, got %d", len(key1))
	}
}

func TestEnsureDirs(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	if err := mgr.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	for _, cat := range AllCategories() {
		dir := mgr.CategoryDir(cat)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Category dir %s not created", dir)
		}
	}

	locksDir := filepath.Join(cfg.BaseDir, "locks")
	if _, err := os.Stat(locksDir); os.IsNotExist(err) {
		t.Error("Locks directory not created")
	}
}

func TestStoreAndGet(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	// Create a temporary file to cache
	tmpFile := filepath.Join(t.TempDir(), "test-package.tar.gz")
	if err := os.WriteFile(tmpFile, []byte("test content data"), 0644); err != nil {
		t.Fatal(err)
	}

	key := CacheKey("https://example.com/test-package.tar.gz")

	// Store
	meta, err := mgr.Store(CategoryDownloads, key, tmpFile, "https://example.com/test-package.tar.gz")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	if meta == nil {
		t.Fatal("Store returned nil metadata")
	}
	if meta.Source != "https://example.com/test-package.tar.gz" {
		t.Errorf("Source mismatch: %s", meta.Source)
	}
	if meta.Filename != "test-package.tar.gz" {
		t.Errorf("Filename mismatch: %s", meta.Filename)
	}

	// Get
	path, gotMeta, err := mgr.Get(CategoryDownloads, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if path == "" {
		t.Fatal("Get returned empty path")
	}
	if gotMeta == nil {
		t.Fatal("Get returned nil metadata")
	}
	if gotMeta.Source != meta.Source {
		t.Error("Metadata mismatch after Get")
	}

	// Verify cached file content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Cannot read cached file: %v", err)
	}
	if string(data) != "test content data" {
		t.Errorf("Cached content mismatch: %s", string(data))
	}
}

func TestGetNonExistent(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	path, meta, err := mgr.Get(CategoryDownloads, "nonexistent")
	if err != nil {
		t.Fatalf("Get should not error for missing key: %v", err)
	}
	if path != "" || meta != nil {
		t.Error("Get should return empty for missing key")
	}
}

func TestGetExpired(t *testing.T) {
	cfg := testConfig(t)
	// Set very short TTL
	cfg.Categories[CategoryDownloads] = CategoryConfig{MaxSize: 100 << 20, TTL: 1 * time.Millisecond}
	mgr := NewManagerWithConfig(cfg)

	tmpFile := filepath.Join(t.TempDir(), "expired.bin")
	os.WriteFile(tmpFile, []byte("data"), 0644)

	key := CacheKey("expired-test")
	_, err := mgr.Store(CategoryDownloads, key, tmpFile, "expired-test")
	if err != nil {
		t.Fatal(err)
	}

	// Wait for expiry
	time.Sleep(5 * time.Millisecond)

	path, meta, err := mgr.Get(CategoryDownloads, key)
	if err != nil {
		t.Fatal(err)
	}
	if path != "" || meta != nil {
		t.Error("Expired entry should not be returned")
	}
}

func TestRemove(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	tmpFile := filepath.Join(t.TempDir(), "removable.bin")
	os.WriteFile(tmpFile, []byte("data"), 0644)

	key := CacheKey("remove-test")
	mgr.Store(CategoryDownloads, key, tmpFile, "remove-test")

	if err := mgr.Remove(CategoryDownloads, key); err != nil {
		t.Fatal(err)
	}

	path, _, _ := mgr.Get(CategoryDownloads, key)
	if path != "" {
		t.Error("Entry should be removed")
	}
}

func TestListEntries(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	// Store two entries
	for _, name := range []string{"pkg-a.tar.gz", "pkg-b.tar.gz"} {
		tmpFile := filepath.Join(t.TempDir(), name)
		os.WriteFile(tmpFile, []byte("data for "+name), 0644)
		key := CacheKey(name)
		mgr.Store(CategoryDownloads, key, tmpFile, "https://example.com/"+name)
	}

	entries, err := mgr.ListEntries(CategoryDownloads)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestGetInfo(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	tmpFile := filepath.Join(t.TempDir(), "info-test.bin")
	os.WriteFile(tmpFile, []byte("info test data"), 0644)
	mgr.Store(CategoryDownloads, CacheKey("info-test"), tmpFile, "info-test")

	info, err := mgr.GetInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.TotalFiles != 2 { // data file + .meta.json
		t.Errorf("Expected 2 files, got %d", info.TotalFiles)
	}
	if info.TotalSize == 0 {
		t.Error("Total size should be > 0")
	}
	if !info.Enabled {
		t.Error("Cache should be enabled")
	}
}

func TestClean(t *testing.T) {
	cfg := testConfig(t)
	cfg.Categories[CategoryDownloads] = CategoryConfig{MaxSize: 100 << 20, TTL: 1 * time.Millisecond}
	mgr := NewManagerWithConfig(cfg)

	tmpFile := filepath.Join(t.TempDir(), "clean-test.bin")
	os.WriteFile(tmpFile, []byte("clean test"), 0644)
	mgr.Store(CategoryDownloads, CacheKey("clean-test"), tmpFile, "clean-test")

	time.Sleep(5 * time.Millisecond)

	result, err := mgr.Clean()
	if err != nil {
		t.Fatal(err)
	}
	if result.RemovedFiles == 0 {
		t.Error("Should have removed expired entries")
	}
}

func TestPurge(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	tmpFile := filepath.Join(t.TempDir(), "purge-test.bin")
	os.WriteFile(tmpFile, []byte("purge test"), 0644)
	mgr.Store(CategoryDownloads, CacheKey("purge-test"), tmpFile, "purge-test")

	result, err := mgr.Purge()
	if err != nil {
		t.Fatal(err)
	}
	if result.FreedBytes == 0 {
		t.Error("Purge should have freed bytes")
	}

	// Verify cache is empty
	entries, _ := mgr.ListEntries(CategoryDownloads)
	if len(entries) != 0 {
		t.Error("Cache should be empty after purge")
	}
}

func TestRemoveBySource(t *testing.T) {
	cfg := testConfig(t)
	mgr := NewManagerWithConfig(cfg)

	// Store entries with different sources
	for _, name := range []string{"nodejs-v20.tar.gz", "python-3.13.tar.gz", "nodejs-v22.tar.gz"} {
		tmpFile := filepath.Join(t.TempDir(), name)
		os.WriteFile(tmpFile, []byte("data"), 0644)
		mgr.Store(CategoryDownloads, CacheKey(name), tmpFile, "https://example.com/"+name)
	}

	result, err := mgr.RemoveBySource("nodejs")
	if err != nil {
		t.Fatal(err)
	}
	if result.RemovedFiles != 2 {
		t.Errorf("Expected 2 removed, got %d", result.RemovedFiles)
	}

	entries, _ := mgr.ListEntries(CategoryDownloads)
	if len(entries) != 1 {
		t.Errorf("Expected 1 remaining entry, got %d", len(entries))
	}
}

func TestDisabledCache(t *testing.T) {
	cfg := testConfig(t)
	cfg.Enabled = false
	mgr := NewManagerWithConfig(cfg)

	tmpFile := filepath.Join(t.TempDir(), "disabled.bin")
	os.WriteFile(tmpFile, []byte("data"), 0644)

	meta, err := mgr.Store(CategoryDownloads, "key", tmpFile, "src")
	if err != nil {
		t.Fatal(err)
	}
	if meta != nil {
		t.Error("Store should return nil when cache is disabled")
	}

	path, _, _ := mgr.Get(CategoryDownloads, "key")
	if path != "" {
		t.Error("Get should return empty when cache is disabled")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1536 * 1024 * 1024, "1.50 GB"},
	}

	for _, tt := range tests {
		got := FormatSize(tt.input)
		if got != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{-1 * time.Hour, "expired"},
		{30 * time.Minute, "30m"},
		{2 * time.Hour, "2h 0m"},
		{25 * time.Hour, "1d 1h"},
		{48*time.Hour + 30*time.Minute, "2d 0h"},
	}

	for _, tt := range tests {
		got := FormatDuration(tt.input)
		if got != tt.expected {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
