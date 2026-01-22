package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// Directory and file names
	credentialDir     = "credentials"
	defaultStoreName  = "default"
	storeExtension    = ".enc"
	passwordMarker    = ".password-protected"
	m365TokenFile     = ".portunix-m365-tokens.enc"
	portunixDir       = ".portunix"
	currentVersion    = 1
)

// Credential represents a single stored credential
type Credential struct {
	Name     string            `json:"name"`
	Label    string            `json:"label,omitempty"`
	Value    string            `json:"value"`
	Created  time.Time         `json:"created"`
	Updated  time.Time         `json:"updated"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CredentialStore represents the credential store structure
type CredentialStore struct {
	Version     int          `json:"version"`
	Credentials []Credential `json:"credentials"`
}

// Storage handles credential storage operations
type Storage struct {
	baseDir   string
	storeName string
	password  string
	crypto    *CryptoService
}

// NewStorage creates a new storage instance
func NewStorage(storeName string, password string) (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, portunixDir, credentialDir)

	// Ensure base directory exists with proper permissions
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// Generate appropriate seed
	var seed string
	if password != "" {
		seed, err = GeneratePasswordProtectedSeed(password)
	} else {
		seed, err = GenerateDefaultSeed()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	return &Storage{
		baseDir:   baseDir,
		storeName: storeName,
		password:  password,
		crypto:    NewCryptoService(seed),
	}, nil
}

// NewM365Storage creates a storage instance for M365 compatibility mode
func NewM365Storage(password string) (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, portunixDir)

	// Ensure base directory exists with proper permissions
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create portunix directory: %w", err)
	}

	// Generate M365-compatible seed
	var seed string
	if password != "" {
		info, err := GetOSInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to get OS info: %w", err)
		}
		// M365 password format: "{hostname}|{username}|{os_name}|{home_dir}|portunix-m365|pw:{password}"
		seed = fmt.Sprintf("%s|%s|%s|%s|portunix-m365|pw:%s",
			info.Hostname, info.Username, info.OSName, info.HomeDir, password)
	} else {
		seed, err = GenerateDefaultM365Seed()
		if err != nil {
			return nil, fmt.Errorf("failed to generate M365 seed: %w", err)
		}
	}

	return &Storage{
		baseDir:   baseDir,
		storeName: "m365",
		password:  password,
		crypto:    NewCryptoService(seed),
	}, nil
}

// getStorePath returns the path to the store file
func (s *Storage) getStorePath() string {
	if s.storeName == "m365" {
		return filepath.Join(s.baseDir, m365TokenFile)
	}
	return filepath.Join(s.baseDir, s.storeName+storeExtension)
}

// Load loads the credential store from disk
func (s *Storage) Load() (*CredentialStore, error) {
	storePath := s.getStorePath()

	// Check if store exists
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		// Return empty store
		return &CredentialStore{
			Version:     currentVersion,
			Credentials: []Credential{},
		}, nil
	}

	// Read encrypted data
	encryptedData, err := os.ReadFile(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read store file: %w", err)
	}

	// Decrypt data
	decrypted, err := s.crypto.Decrypt(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt store: %w", err)
	}

	// Parse JSON
	var store CredentialStore
	if err := json.Unmarshal([]byte(decrypted), &store); err != nil {
		return nil, fmt.Errorf("failed to parse store: %w", err)
	}

	return &store, nil
}

// Save saves the credential store to disk
func (s *Storage) Save(store *CredentialStore) error {
	storePath := s.getStorePath()

	// Serialize to JSON
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize store: %w", err)
	}

	// Encrypt data
	encrypted, err := s.crypto.Encrypt(string(data))
	if err != nil {
		return fmt.Errorf("failed to encrypt store: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(storePath, []byte(encrypted), 0600); err != nil {
		return fmt.Errorf("failed to write store file: %w", err)
	}

	return nil
}

// Set adds or updates a credential
func (s *Storage) Set(name, value, label string, metadata map[string]string) error {
	store, err := s.Load()
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	// Find existing credential
	found := false
	for i := range store.Credentials {
		if store.Credentials[i].Name == name {
			store.Credentials[i].Value = value
			store.Credentials[i].Updated = now
			if label != "" {
				store.Credentials[i].Label = label
			}
			if metadata != nil {
				store.Credentials[i].Metadata = metadata
			}
			found = true
			break
		}
	}

	// Add new credential if not found
	if !found {
		cred := Credential{
			Name:     name,
			Label:    label,
			Value:    value,
			Created:  now,
			Updated:  now,
			Metadata: metadata,
		}
		store.Credentials = append(store.Credentials, cred)
	}

	return s.Save(store)
}

// Get retrieves a credential value by name
func (s *Storage) Get(name string) (string, error) {
	store, err := s.Load()
	if err != nil {
		return "", err
	}

	for _, cred := range store.Credentials {
		if cred.Name == name {
			return cred.Value, nil
		}
	}

	return "", fmt.Errorf("credential not found: %s", name)
}

// GetCredential retrieves a full credential by name
func (s *Storage) GetCredential(name string) (*Credential, error) {
	store, err := s.Load()
	if err != nil {
		return nil, err
	}

	for _, cred := range store.Credentials {
		if cred.Name == name {
			return &cred, nil
		}
	}

	return nil, fmt.Errorf("credential not found: %s", name)
}

// Delete removes a credential by name
func (s *Storage) Delete(name string) error {
	store, err := s.Load()
	if err != nil {
		return err
	}

	found := false
	newCreds := make([]Credential, 0, len(store.Credentials))
	for _, cred := range store.Credentials {
		if cred.Name != name {
			newCreds = append(newCreds, cred)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("credential not found: %s", name)
	}

	store.Credentials = newCreds
	return s.Save(store)
}

// List returns all credentials (without values)
func (s *Storage) List() ([]Credential, error) {
	store, err := s.Load()
	if err != nil {
		return nil, err
	}

	// Return credentials without values for security
	result := make([]Credential, len(store.Credentials))
	for i, cred := range store.Credentials {
		result[i] = Credential{
			Name:     cred.Name,
			Label:    cred.Label,
			Created:  cred.Created,
			Updated:  cred.Updated,
			Metadata: cred.Metadata,
			// Value is intentionally omitted
		}
	}

	return result, nil
}

// Exists checks if a credential exists
func (s *Storage) Exists(name string) (bool, error) {
	store, err := s.Load()
	if err != nil {
		return false, err
	}

	for _, cred := range store.Credentials {
		if cred.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// StoreExists checks if the store file exists
func (s *Storage) StoreExists() bool {
	_, err := os.Stat(s.getStorePath())
	return err == nil
}

// DeleteStore deletes the entire store file
func (s *Storage) DeleteStore() error {
	storePath := s.getStorePath()
	if err := os.Remove(storePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete store: %w", err)
	}
	return nil
}

// M365Tokens represents the M365 token structure for compatibility
type M365Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
	TokenType    string `json:"tokenType"`
}

// GetM365Tokens retrieves M365 tokens in legacy format
func (s *Storage) GetM365Tokens() (*M365Tokens, error) {
	store, err := s.Load()
	if err != nil {
		return nil, err
	}

	// Look for M365 token structure in store
	tokens := &M365Tokens{}
	for _, cred := range store.Credentials {
		switch cred.Name {
		case "accessToken":
			tokens.AccessToken = cred.Value
		case "refreshToken":
			tokens.RefreshToken = cred.Value
		case "expiresAt":
			// Parse as int64
			fmt.Sscanf(cred.Value, "%d", &tokens.ExpiresAt)
		case "tokenType":
			tokens.TokenType = cred.Value
		}
	}

	if tokens.AccessToken == "" {
		return nil, fmt.Errorf("no M365 tokens found")
	}

	return tokens, nil
}

// SetM365Tokens stores M365 tokens in legacy format
func (s *Storage) SetM365Tokens(tokens *M365Tokens) error {
	store, err := s.Load()
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	// Helper to set or update a credential in the store
	setCredential := func(name, value string) {
		found := false
		for i := range store.Credentials {
			if store.Credentials[i].Name == name {
				store.Credentials[i].Value = value
				store.Credentials[i].Updated = now
				found = true
				break
			}
		}
		if !found {
			store.Credentials = append(store.Credentials, Credential{
				Name:    name,
				Value:   value,
				Created: now,
				Updated: now,
			})
		}
	}

	setCredential("accessToken", tokens.AccessToken)
	setCredential("refreshToken", tokens.RefreshToken)
	setCredential("expiresAt", fmt.Sprintf("%d", tokens.ExpiresAt))
	setCredential("tokenType", tokens.TokenType)

	return s.Save(store)
}

// GetRawM365Data retrieves raw M365 token data as JSON string
// This is used for compatibility with Java implementation that stores
// the entire token object as a single encrypted JSON string
func (s *Storage) GetRawM365Data() (string, error) {
	storePath := s.getStorePath()

	// Check if store exists
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return "", fmt.Errorf("M365 token file not found")
	}

	// Read encrypted data
	encryptedData, err := os.ReadFile(storePath)
	if err != nil {
		return "", fmt.Errorf("failed to read M365 token file: %w", err)
	}

	// Decrypt data
	decrypted, err := s.crypto.Decrypt(string(encryptedData))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt M365 tokens: %w", err)
	}

	return decrypted, nil
}

// SetRawM365Data stores raw JSON data for M365 tokens
// This is used for compatibility with Java implementation
func (s *Storage) SetRawM365Data(jsonData string) error {
	// Validate JSON
	var temp interface{}
	if err := json.Unmarshal([]byte(jsonData), &temp); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	// Encrypt data
	encrypted, err := s.crypto.Encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt M365 tokens: %w", err)
	}

	storePath := s.getStorePath()

	// Write to file with restricted permissions
	if err := os.WriteFile(storePath, []byte(encrypted), 0600); err != nil {
		return fmt.Errorf("failed to write M365 token file: %w", err)
	}

	return nil
}

// ListStores returns a list of all available credential stores
func ListStores() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	credDir := filepath.Join(homeDir, portunixDir, credentialDir)

	// Check if directory exists
	if _, err := os.Stat(credDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(credDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials directory: %w", err)
	}

	var stores []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == storeExtension {
			stores = append(stores, name[:len(name)-len(storeExtension)])
		}
	}

	return stores, nil
}

// CreateStore creates a new empty credential store
func CreateStore(storeName string, password string) error {
	storage, err := NewStorage(storeName, password)
	if err != nil {
		return err
	}

	if storage.StoreExists() {
		return fmt.Errorf("store already exists: %s", storeName)
	}

	store := &CredentialStore{
		Version:     currentVersion,
		Credentials: []Credential{},
	}

	if err := storage.Save(store); err != nil {
		return err
	}

	// Create password marker if password-protected
	if password != "" {
		markerPath := filepath.Join(storage.baseDir, storeName+passwordMarker)
		if err := os.WriteFile(markerPath, []byte{}, 0600); err != nil {
			return fmt.Errorf("failed to create password marker: %w", err)
		}
	}

	return nil
}

// DeleteStoreByName deletes a credential store by name
func DeleteStoreByName(storeName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, portunixDir, credentialDir)

	// Delete store file
	storePath := filepath.Join(baseDir, storeName+storeExtension)
	if err := os.Remove(storePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete store: %w", err)
	}

	// Delete password marker if exists
	markerPath := filepath.Join(baseDir, storeName+passwordMarker)
	os.Remove(markerPath) // Ignore error if marker doesn't exist

	return nil
}

// IsPasswordProtected checks if a store is password-protected
func IsPasswordProtected(storeName string) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to get home directory: %w", err)
	}

	markerPath := filepath.Join(homeDir, portunixDir, credentialDir, storeName+passwordMarker)
	_, err = os.Stat(markerPath)
	return err == nil, nil
}
