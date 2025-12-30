package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheFileName = ".pft-cache.json"

// CacheEntry represents a cached feedback item state
type CacheEntry struct {
	ID         string    `json:"id"`
	ExternalID string    `json:"external_id,omitempty"`
	Title      string    `json:"title"`
	Hash       string    `json:"hash"`
	SyncedAt   time.Time `json:"synced_at"`
	FilePath   string    `json:"file_path,omitempty"`
}

// SyncCache manages local cache of synchronized items
type SyncCache struct {
	Version   string                `json:"version"`
	UpdatedAt time.Time             `json:"updated_at"`
	Entries   map[string]CacheEntry `json:"entries"`
	filePath  string
}

// NewSyncCache creates a new sync cache
func NewSyncCache(projectDir string) *SyncCache {
	return &SyncCache{
		Version:  "1.0",
		Entries:  make(map[string]CacheEntry),
		filePath: filepath.Join(projectDir, cacheFileName),
	}
}

// Load reads the cache from disk
func (c *SyncCache) Load() error {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file yet
		}
		return fmt.Errorf("failed to read cache: %w", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse cache: %w", err)
	}

	return nil
}

// Save writes the cache to disk
func (c *SyncCache) Save() error {
	c.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize cache: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// Get retrieves a cache entry by ID
func (c *SyncCache) Get(id string) (CacheEntry, bool) {
	entry, ok := c.Entries[id]
	return entry, ok
}

// Set adds or updates a cache entry
func (c *SyncCache) Set(entry CacheEntry) {
	if c.Entries == nil {
		c.Entries = make(map[string]CacheEntry)
	}
	c.Entries[entry.ID] = entry
}

// Delete removes a cache entry
func (c *SyncCache) Delete(id string) {
	delete(c.Entries, id)
}

// Clear removes all cache entries
func (c *SyncCache) Clear() {
	c.Entries = make(map[string]CacheEntry)
}

// GetAll returns all cache entries
func (c *SyncCache) GetAll() []CacheEntry {
	entries := make([]CacheEntry, 0, len(c.Entries))
	for _, entry := range c.Entries {
		entries = append(entries, entry)
	}
	return entries
}

// HasChanged checks if an item has changed since last sync
func (c *SyncCache) HasChanged(item *FeedbackItem) bool {
	entry, ok := c.Entries[item.ID]
	if !ok {
		return true // New item, not in cache
	}

	currentHash := hashItem(item)
	return currentHash != entry.Hash
}

// RecordSync records a successful sync for an item
func (c *SyncCache) RecordSync(item *FeedbackItem) {
	entry := CacheEntry{
		ID:         item.ID,
		ExternalID: item.ExternalID,
		Title:      item.Title,
		Hash:       hashItem(item),
		SyncedAt:   time.Now(),
		FilePath:   item.FilePath,
	}
	c.Set(entry)
}

// hashItem creates a simple hash of item content for change detection
func hashItem(item *FeedbackItem) string {
	// Simple hash using title + description + status
	content := item.Title + "|" + item.Description + "|" + item.Status
	var hash uint32
	for _, c := range content {
		hash = hash*31 + uint32(c)
	}
	return fmt.Sprintf("%08x", hash)
}

// GetSyncStats returns statistics about the cache
func (c *SyncCache) GetSyncStats() (total, synced, unsynced int) {
	total = len(c.Entries)
	for _, entry := range c.Entries {
		if entry.ExternalID != "" {
			synced++
		} else {
			unsynced++
		}
	}
	return total, synced, unsynced
}

// PrintCacheStatus displays cache status
func (c *SyncCache) PrintCacheStatus() {
	total, synced, unsynced := c.GetSyncStats()

	fmt.Printf("📦 Cache Status: %s\n", c.filePath)
	fmt.Printf("   Version: %s\n", c.Version)
	fmt.Printf("   Last updated: %s\n", c.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Total entries: %d\n", total)
	fmt.Printf("   Synced: %d\n", synced)
	fmt.Printf("   Unsynced: %d\n", unsynced)
}

// CleanupOrphans removes cache entries for files that no longer exist
func (c *SyncCache) CleanupOrphans() int {
	removed := 0
	for id, entry := range c.Entries {
		if entry.FilePath != "" {
			if _, err := os.Stat(entry.FilePath); os.IsNotExist(err) {
				delete(c.Entries, id)
				removed++
			}
		}
	}
	return removed
}

// FindUnsyncedItems returns local items that haven't been synced
func (c *SyncCache) FindUnsyncedItems(items []*FeedbackItem) []*FeedbackItem {
	var unsynced []*FeedbackItem
	for _, item := range items {
		entry, ok := c.Entries[item.ID]
		if !ok || entry.ExternalID == "" {
			unsynced = append(unsynced, item)
		}
	}
	return unsynced
}

// FindModifiedItems returns items that have changed since last sync
func (c *SyncCache) FindModifiedItems(items []*FeedbackItem) []*FeedbackItem {
	var modified []*FeedbackItem
	for _, item := range items {
		if c.HasChanged(item) {
			modified = append(modified, item)
		}
	}
	return modified
}
