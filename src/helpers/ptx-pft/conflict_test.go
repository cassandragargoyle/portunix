package main

import (
	"testing"
	"time"
)

func TestNewConflictDetector(t *testing.T) {
	detector := NewConflictDetector(ConflictLocal)

	if detector.Resolution != ConflictLocal {
		t.Errorf("Expected resolution ConflictLocal, got %s", detector.Resolution)
	}
}

func TestDetectConflictNoConflict(t *testing.T) {
	detector := NewConflictDetector(ConflictLocal)

	local := &FeedbackItem{
		ID:          "UC001",
		Title:       "Same Title",
		Description: "Same Description",
		Status:      "open",
	}
	remote := &FeedbackItem{
		ID:          "UC001",
		Title:       "Same Title",
		Description: "Same Description",
		Status:      "open",
	}

	conflict := detector.DetectConflict(local, remote)

	if conflict != nil {
		t.Error("Should not detect conflict for identical items")
	}
}

func TestDetectConflictTitleDiff(t *testing.T) {
	detector := NewConflictDetector(ConflictLocal)

	local := &FeedbackItem{
		ID:          "UC001",
		Title:       "Local Title",
		Description: "Same Description",
		Status:      "open",
	}
	remote := &FeedbackItem{
		ID:          "UC001",
		Title:       "Remote Title",
		Description: "Same Description",
		Status:      "open",
	}

	conflict := detector.DetectConflict(local, remote)

	if conflict == nil {
		t.Fatal("Should detect title conflict")
	}
	if conflict.ItemID != "UC001" {
		t.Errorf("Expected item ID 'UC001', got '%s'", conflict.ItemID)
	}
	if conflict.Reason != "modified fields: title" {
		t.Errorf("Unexpected reason: %s", conflict.Reason)
	}
}

func TestDetectConflictMultipleFields(t *testing.T) {
	detector := NewConflictDetector(ConflictLocal)

	local := &FeedbackItem{
		ID:          "UC001",
		Title:       "Local Title",
		Description: "Local Description",
		Status:      "open",
	}
	remote := &FeedbackItem{
		ID:          "UC001",
		Title:       "Remote Title",
		Description: "Remote Description",
		Status:      "closed",
	}

	conflict := detector.DetectConflict(local, remote)

	if conflict == nil {
		t.Fatal("Should detect multiple field conflicts")
	}
	if conflict.Reason != "modified fields: title, description, status" {
		t.Errorf("Unexpected reason: %s", conflict.Reason)
	}
}

func TestDetectConflictNilItems(t *testing.T) {
	detector := NewConflictDetector(ConflictLocal)

	if detector.DetectConflict(nil, nil) != nil {
		t.Error("Should return nil for nil items")
	}
	if detector.DetectConflict(&FeedbackItem{}, nil) != nil {
		t.Error("Should return nil when remote is nil")
	}
	if detector.DetectConflict(nil, &FeedbackItem{}) != nil {
		t.Error("Should return nil when local is nil")
	}
}

func TestResolveConflictLocal(t *testing.T) {
	detector := NewConflictDetector(ConflictLocal)

	conflict := &SyncConflict{
		ItemID:     "UC001",
		LocalItem:  FeedbackItem{Title: "Local Version"},
		RemoteItem: FeedbackItem{Title: "Remote Version"},
		Reason:     "title modified",
	}

	resolved, winner, err := detector.ResolveConflict(conflict)

	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if winner != "local" {
		t.Errorf("Expected winner 'local', got '%s'", winner)
	}
	if resolved.Title != "Local Version" {
		t.Errorf("Expected local version, got '%s'", resolved.Title)
	}
	if conflict.Resolution != "kept local version" {
		t.Errorf("Unexpected resolution: %s", conflict.Resolution)
	}
}

func TestResolveConflictRemote(t *testing.T) {
	detector := NewConflictDetector(ConflictRemote)

	conflict := &SyncConflict{
		ItemID:     "UC001",
		LocalItem:  FeedbackItem{Title: "Local Version"},
		RemoteItem: FeedbackItem{Title: "Remote Version"},
	}

	resolved, winner, err := detector.ResolveConflict(conflict)

	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if winner != "remote" {
		t.Errorf("Expected winner 'remote', got '%s'", winner)
	}
	if resolved.Title != "Remote Version" {
		t.Errorf("Expected remote version, got '%s'", resolved.Title)
	}
}

func TestResolveConflictTimestamp(t *testing.T) {
	detector := NewConflictDetector(ConflictTimestamp)

	now := time.Now()
	older := now.Add(-1 * time.Hour)

	conflict := &SyncConflict{
		ItemID:     "UC001",
		LocalItem:  FeedbackItem{Title: "Newer Local", UpdatedAt: now.Format(time.RFC3339)},
		RemoteItem: FeedbackItem{Title: "Older Remote", UpdatedAt: older.Format(time.RFC3339)},
	}

	resolved, winner, err := detector.ResolveConflict(conflict)

	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if winner != "local" {
		t.Errorf("Expected winner 'local' (newer), got '%s'", winner)
	}
	if resolved.Title != "Newer Local" {
		t.Errorf("Expected newer local version")
	}
}

func TestResolveConflictTimestampRemoteNewer(t *testing.T) {
	detector := NewConflictDetector(ConflictTimestamp)

	now := time.Now()
	older := now.Add(-1 * time.Hour)

	conflict := &SyncConflict{
		ItemID:     "UC001",
		LocalItem:  FeedbackItem{Title: "Older Local", UpdatedAt: older.Format(time.RFC3339)},
		RemoteItem: FeedbackItem{Title: "Newer Remote", UpdatedAt: now.Format(time.RFC3339)},
	}

	resolved, winner, err := detector.ResolveConflict(conflict)

	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}
	if winner != "remote" {
		t.Errorf("Expected winner 'remote' (newer), got '%s'", winner)
	}
	if resolved.Title != "Newer Remote" {
		t.Errorf("Expected newer remote version")
	}
}

func TestResolveConflictManual(t *testing.T) {
	detector := NewConflictDetector(ConflictManual)

	conflict := &SyncConflict{
		ItemID:     "UC001",
		LocalItem:  FeedbackItem{Title: "Local"},
		RemoteItem: FeedbackItem{Title: "Remote"},
	}

	_, _, err := detector.ResolveConflict(conflict)

	if err == nil {
		t.Error("Manual resolution should return error requiring user input")
	}
	if err.Error() != "manual resolution required" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestResolveConflictNilConflict(t *testing.T) {
	detector := NewConflictDetector(ConflictLocal)

	_, _, err := detector.ResolveConflict(nil)

	if err == nil {
		t.Error("Should error on nil conflict")
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"2024-01-15T10:30:00Z", true},
		{"2024-01-15", true},
		{"", false},
		{"invalid", false},
	}

	for _, tc := range tests {
		result := parseTimestamp(tc.input)
		isValid := !result.IsZero()
		if isValid != tc.expected {
			t.Errorf("parseTimestamp(%q) valid=%v, expected %v", tc.input, isValid, tc.expected)
		}
	}
}

func TestDetectAllConflicts(t *testing.T) {
	localItems := []*FeedbackItem{
		{ID: "1", ExternalID: "ext1", Title: "Local 1", Description: "Desc", Status: "open"},
		{ID: "2", ExternalID: "ext2", Title: "Different", Description: "Desc", Status: "open"},
		{ID: "3", Title: "No match", Description: "Desc", Status: "open"},
	}

	remoteItems := []FeedbackItem{
		{ID: "ext1", Title: "Local 1", Description: "Desc", Status: "open"},     // No conflict
		{ID: "ext2", Title: "Remote 2", Description: "Desc", Status: "open"},    // Title conflict
		{ID: "ext3", Title: "Remote 3", Description: "Desc", Status: "closed"},  // No local match
	}

	conflicts := DetectAllConflicts(localItems, remoteItems)

	if len(conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(conflicts))
	}
	if len(conflicts) > 0 && conflicts[0].ItemID != "2" {
		t.Errorf("Expected conflict for item '2', got '%s'", conflicts[0].ItemID)
	}
}

func TestBuildConflictReason(t *testing.T) {
	tests := []struct {
		title, desc, status bool
		expected            string
	}{
		{true, false, false, "modified fields: title"},
		{false, true, false, "modified fields: description"},
		{false, false, true, "modified fields: status"},
		{true, true, false, "modified fields: title, description"},
		{true, true, true, "modified fields: title, description, status"},
	}

	for _, tc := range tests {
		result := buildConflictReason(tc.title, tc.desc, tc.status)
		if result != tc.expected {
			t.Errorf("buildConflictReason(%v,%v,%v) = %q, expected %q",
				tc.title, tc.desc, tc.status, result, tc.expected)
		}
	}
}
