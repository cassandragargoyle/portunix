package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorageSetAndGet(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	storage, err := NewStorage("test-store", "")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test set
	err = storage.Set("test-key", "test-value", "Test Label", nil)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test get
	value, err := storage.Get("test-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != "test-value" {
		t.Errorf("Get: got %q, want %q", value, "test-value")
	}

	// Test get full credential
	cred, err := storage.GetCredential("test-key")
	if err != nil {
		t.Fatalf("GetCredential failed: %v", err)
	}
	if cred.Name != "test-key" {
		t.Errorf("Credential name: got %q, want %q", cred.Name, "test-key")
	}
	if cred.Label != "Test Label" {
		t.Errorf("Credential label: got %q, want %q", cred.Label, "Test Label")
	}
	if cred.Value != "test-value" {
		t.Errorf("Credential value: got %q, want %q", cred.Value, "test-value")
	}
}

func TestStorageUpdate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	storage, err := NewStorage("test-store", "")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Set initial value
	err = storage.Set("update-key", "initial-value", "Initial", nil)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Update value
	err = storage.Set("update-key", "updated-value", "Updated", nil)
	if err != nil {
		t.Fatalf("Set update failed: %v", err)
	}

	// Verify update
	value, err := storage.Get("update-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != "updated-value" {
		t.Errorf("Get after update: got %q, want %q", value, "updated-value")
	}
}

func TestStorageDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	storage, err := NewStorage("test-store", "")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Set value
	err = storage.Set("delete-key", "delete-value", "", nil)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify it exists
	exists, _ := storage.Exists("delete-key")
	if !exists {
		t.Error("Credential should exist before delete")
	}

	// Delete
	err = storage.Delete("delete-key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	exists, _ = storage.Exists("delete-key")
	if exists {
		t.Error("Credential should not exist after delete")
	}

	// Get should fail
	_, err = storage.Get("delete-key")
	if err == nil {
		t.Error("Get should fail after delete")
	}
}

func TestStorageDeleteNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	storage, err := NewStorage("test-store", "")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Delete non-existent should fail
	err = storage.Delete("non-existent-key")
	if err == nil {
		t.Error("Delete non-existent should fail")
	}
}

func TestStorageList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	storage, err := NewStorage("test-store", "")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Add multiple credentials
	storage.Set("key1", "value1", "Label 1", nil)
	storage.Set("key2", "value2", "Label 2", nil)
	storage.Set("key3", "value3", "", nil)

	// List credentials
	creds, err := storage.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(creds) != 3 {
		t.Errorf("List: got %d credentials, want 3", len(creds))
	}

	// Verify values are not exposed in list
	for _, cred := range creds {
		if cred.Value != "" {
			t.Errorf("List should not expose values, got: %s", cred.Value)
		}
	}
}

func TestStoragePasswordProtected(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	password := "test-password-123"

	// Create storage with password
	storage, err := NewStorage("secure-store", password)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Set value
	err = storage.Set("secure-key", "secure-value", "", nil)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get with same password should work
	value, err := storage.Get("secure-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != "secure-value" {
		t.Errorf("Get: got %q, want %q", value, "secure-value")
	}

	// Get with wrong password should fail
	wrongStorage, err := NewStorage("secure-store", "wrong-password")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	_, err = wrongStorage.Get("secure-key")
	if err == nil {
		t.Error("Get with wrong password should fail")
	}
}

func TestM365Storage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	storage, err := NewM365Storage("")
	if err != nil {
		t.Fatalf("Failed to create M365 storage: %v", err)
	}

	// Test set raw JSON
	jsonData := `{"accessToken":"abc123","refreshToken":"xyz789","expiresAt":1234567890,"tokenType":"Bearer"}`
	err = storage.SetRawM365Data(jsonData)
	if err != nil {
		t.Fatalf("SetRawM365Data failed: %v", err)
	}

	// Test get raw JSON
	retrieved, err := storage.GetRawM365Data()
	if err != nil {
		t.Fatalf("GetRawM365Data failed: %v", err)
	}
	if retrieved != jsonData {
		t.Errorf("GetRawM365Data: got %q, want %q", retrieved, jsonData)
	}

	// Verify file location
	expectedPath := filepath.Join(tmpDir, ".portunix", ".portunix-m365-tokens.enc")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("M365 token file not created at expected location: %s", expectedPath)
	}
}

func TestM365StorageInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	storage, err := NewM365Storage("")
	if err != nil {
		t.Fatalf("Failed to create M365 storage: %v", err)
	}

	// Test invalid JSON
	err = storage.SetRawM365Data("not valid json")
	if err == nil {
		t.Error("SetRawM365Data should fail for invalid JSON")
	}
}

func TestListStores(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	// Create multiple stores
	storage1, _ := NewStorage("store1", "")
	storage1.Set("key1", "value1", "", nil)

	storage2, _ := NewStorage("store2", "")
	storage2.Set("key2", "value2", "", nil)

	// List stores
	stores, err := ListStores()
	if err != nil {
		t.Fatalf("ListStores failed: %v", err)
	}

	if len(stores) != 2 {
		t.Errorf("ListStores: got %d stores, want 2", len(stores))
	}

	// Check store names
	foundStore1 := false
	foundStore2 := false
	for _, store := range stores {
		if store == "store1" {
			foundStore1 = true
		}
		if store == "store2" {
			foundStore2 = true
		}
	}
	if !foundStore1 {
		t.Error("ListStores: store1 not found")
	}
	if !foundStore2 {
		t.Error("ListStores: store2 not found")
	}
}

func TestCreateAndDeleteStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	// Create store
	err = CreateStore("new-store", "")
	if err != nil {
		t.Fatalf("CreateStore failed: %v", err)
	}

	// Verify it exists
	stores, _ := ListStores()
	found := false
	for _, store := range stores {
		if store == "new-store" {
			found = true
		}
	}
	if !found {
		t.Error("Created store not found in list")
	}

	// Create duplicate should fail
	err = CreateStore("new-store", "")
	if err == nil {
		t.Error("CreateStore duplicate should fail")
	}

	// Delete store
	err = DeleteStoreByName("new-store")
	if err != nil {
		t.Fatalf("DeleteStoreByName failed: %v", err)
	}

	// Verify it's gone
	stores, _ = ListStores()
	for _, store := range stores {
		if store == "new-store" {
			t.Error("Deleted store still in list")
		}
	}
}

func TestFilePermissions(t *testing.T) {
	// Skip on Windows as permissions work differently
	if os.PathSeparator == '\\' {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "ptx-credential-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
	}()

	storage, err := NewStorage("permission-test", "")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	storage.Set("key", "value", "", nil)

	// Check file permissions
	storePath := filepath.Join(tmpDir, ".portunix", "credentials", "permission-test.enc")
	info, err := os.Stat(storePath)
	if err != nil {
		t.Fatalf("Failed to stat store file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Store file permissions: got %o, want 0600", mode)
	}
}
