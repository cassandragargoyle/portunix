package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSyncCache(t *testing.T) {
	cache := NewSyncCache("/tmp/test-project")

	if cache.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", cache.Version)
	}
	if cache.Entries == nil {
		t.Error("Entries map should be initialized")
	}
	if cache.filePath != "/tmp/test-project/.pft-cache.json" {
		t.Errorf("Unexpected file path: %s", cache.filePath)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	entry := CacheEntry{
		ID:         "UC001",
		ExternalID: "42",
		Title:      "Test Entry",
		Hash:       "abc123",
		SyncedAt:   time.Now(),
	}

	cache.Set(entry)

	retrieved, ok := cache.Get("UC001")
	if !ok {
		t.Error("Should find cached entry")
	}
	if retrieved.Title != "Test Entry" {
		t.Errorf("Expected title 'Test Entry', got '%s'", retrieved.Title)
	}
	if retrieved.ExternalID != "42" {
		t.Errorf("Expected external ID '42', got '%s'", retrieved.ExternalID)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	cache.Set(CacheEntry{ID: "UC001", Title: "Test"})
	cache.Delete("UC001")

	_, ok := cache.Get("UC001")
	if ok {
		t.Error("Entry should be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	cache.Set(CacheEntry{ID: "UC001", Title: "Test 1"})
	cache.Set(CacheEntry{ID: "UC002", Title: "Test 2"})
	cache.Clear()

	if len(cache.Entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(cache.Entries))
	}
}

func TestCacheGetAll(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	cache.Set(CacheEntry{ID: "UC001", Title: "Test 1"})
	cache.Set(CacheEntry{ID: "UC002", Title: "Test 2"})
	cache.Set(CacheEntry{ID: "UC003", Title: "Test 3"})

	all := cache.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(all))
	}
}

func TestCacheSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create and save cache
	cache1 := NewSyncCache(tmpDir)
	cache1.Set(CacheEntry{
		ID:         "UC001",
		ExternalID: "100",
		Title:      "Persisted Entry",
		Hash:       "hash123",
		SyncedAt:   time.Now(),
	})

	if err := cache1.Save(); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}

	// Verify file exists
	cachePath := filepath.Join(tmpDir, ".pft-cache.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("Cache file should exist after save")
	}

	// Load cache in new instance
	cache2 := NewSyncCache(tmpDir)
	if err := cache2.Load(); err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	entry, ok := cache2.Get("UC001")
	if !ok {
		t.Error("Should find loaded entry")
	}
	if entry.Title != "Persisted Entry" {
		t.Errorf("Expected 'Persisted Entry', got '%s'", entry.Title)
	}
}

func TestCacheLoadNonexistent(t *testing.T) {
	cache := NewSyncCache("/nonexistent/path")
	err := cache.Load()

	// Should not error on nonexistent file
	if err != nil {
		t.Errorf("Load should not error on nonexistent file: %v", err)
	}
}

func TestCacheHasChanged(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	item := &FeedbackItem{
		ID:          "UC001",
		Title:       "Original Title",
		Description: "Original Description",
		Status:      "open",
	}

	// New item - should report changed
	if !cache.HasChanged(item) {
		t.Error("New item should report as changed")
	}

	// Record sync
	cache.RecordSync(item)

	// Same item - should not report changed
	if cache.HasChanged(item) {
		t.Error("Same item should not report as changed")
	}

	// Modified item - should report changed
	item.Title = "Modified Title"
	if !cache.HasChanged(item) {
		t.Error("Modified item should report as changed")
	}
}

func TestCacheRecordSync(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	item := &FeedbackItem{
		ID:          "UC001",
		ExternalID:  "42",
		Title:       "Test Item",
		Description: "Description",
		Status:      "open",
		FilePath:    "/path/to/file.md",
	}

	cache.RecordSync(item)

	entry, ok := cache.Get("UC001")
	if !ok {
		t.Error("Should find synced entry")
	}
	if entry.ExternalID != "42" {
		t.Errorf("Expected external ID '42', got '%s'", entry.ExternalID)
	}
	if entry.FilePath != "/path/to/file.md" {
		t.Errorf("Expected file path '/path/to/file.md', got '%s'", entry.FilePath)
	}
	if entry.Hash == "" {
		t.Error("Hash should not be empty")
	}
}

func TestCacheGetSyncStats(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	cache.Set(CacheEntry{ID: "UC001", ExternalID: "1"})
	cache.Set(CacheEntry{ID: "UC002", ExternalID: "2"})
	cache.Set(CacheEntry{ID: "UC003", ExternalID: ""}) // Not synced

	total, synced, unsynced := cache.GetSyncStats()

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
	if synced != 2 {
		t.Errorf("Expected synced 2, got %d", synced)
	}
	if unsynced != 1 {
		t.Errorf("Expected unsynced 1, got %d", unsynced)
	}
}

func TestCacheCleanupOrphans(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(tmpDir, "UC001-real.md")
	os.WriteFile(realFile, []byte("test"), 0644)

	cache := NewSyncCache(tmpDir)
	cache.Set(CacheEntry{ID: "UC001", FilePath: realFile})
	cache.Set(CacheEntry{ID: "UC002", FilePath: filepath.Join(tmpDir, "UC002-deleted.md")})

	removed := cache.CleanupOrphans()

	if removed != 1 {
		t.Errorf("Expected 1 orphan removed, got %d", removed)
	}

	_, ok := cache.Get("UC001")
	if !ok {
		t.Error("Real file entry should remain")
	}

	_, ok = cache.Get("UC002")
	if ok {
		t.Error("Orphan entry should be removed")
	}
}

func TestCacheFindUnsyncedItems(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	cache.Set(CacheEntry{ID: "UC001", ExternalID: "1"}) // Synced

	items := []*FeedbackItem{
		{ID: "UC001", Title: "Synced"},
		{ID: "UC002", Title: "Not synced"},
		{ID: "UC003", Title: "Also not synced"},
	}

	unsynced := cache.FindUnsyncedItems(items)

	if len(unsynced) != 2 {
		t.Errorf("Expected 2 unsynced items, got %d", len(unsynced))
	}
}

func TestCacheFindModifiedItems(t *testing.T) {
	cache := NewSyncCache("/tmp/test")

	original := &FeedbackItem{
		ID:          "UC001",
		Title:       "Original",
		Description: "Desc",
		Status:      "open",
	}
	cache.RecordSync(original)

	items := []*FeedbackItem{
		{ID: "UC001", Title: "Modified", Description: "Desc", Status: "open"},
		{ID: "UC002", Title: "New", Description: "New desc", Status: "open"},
	}

	modified := cache.FindModifiedItems(items)

	if len(modified) != 2 {
		t.Errorf("Expected 2 modified items (1 changed, 1 new), got %d", len(modified))
	}
}

func TestHashItem(t *testing.T) {
	item1 := &FeedbackItem{
		Title:       "Test",
		Description: "Description",
		Status:      "open",
	}
	item2 := &FeedbackItem{
		Title:       "Test",
		Description: "Description",
		Status:      "open",
	}
	item3 := &FeedbackItem{
		Title:       "Different",
		Description: "Description",
		Status:      "open",
	}

	hash1 := hashItem(item1)
	hash2 := hashItem(item2)
	hash3 := hashItem(item3)

	if hash1 != hash2 {
		t.Error("Same items should have same hash")
	}
	if hash1 == hash3 {
		t.Error("Different items should have different hash")
	}
	if len(hash1) != 8 {
		t.Errorf("Hash should be 8 characters, got %d", len(hash1))
	}
}
