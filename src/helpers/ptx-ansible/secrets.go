package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SecretStore represents a secure secret storage system
type SecretStore struct {
	Type       string                 `json:"type"`        // "file", "env", "vault", "aws", "azure"
	Config     map[string]interface{} `json:"config"`      // Store-specific configuration
	Encryption *EncryptionConfig      `json:"encryption"`  // Encryption settings
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Enabled   bool   `json:"enabled"`
	Algorithm string `json:"algorithm"` // "aes-256-gcm"
	KeySource string `json:"key_source"` // "env", "file", "prompt"
	KeyPath   string `json:"key_path,omitempty"`
}

// SecretManager handles secure secret operations
type SecretManager struct {
	stores          map[string]*SecretStore
	defaultStore    string
	encryptionKey   []byte
	auditMgr        *AuditManager
}

// SecretValue represents an encrypted secret
type SecretValue struct {
	Value       string            `json:"value"`
	Encrypted   bool              `json:"encrypted"`
	Store       string            `json:"store"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
}

// NewSecretManager creates a new secret manager
func NewSecretManager(auditMgr *AuditManager) *SecretManager {
	return &SecretManager{
		stores:   make(map[string]*SecretStore),
		auditMgr: auditMgr,
	}
}

// LoadSecretStores loads secret store configurations
func (sm *SecretManager) LoadSecretStores(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default configuration
		return sm.createDefaultConfig(configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read secret config: %v", err)
	}

	var config struct {
		DefaultStore string                    `json:"default_store"`
		Stores       map[string]*SecretStore   `json:"stores"`
		Encryption   *EncryptionConfig         `json:"encryption"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse secret config: %v", err)
	}

	sm.stores = config.Stores
	sm.defaultStore = config.DefaultStore

	// Initialize encryption if enabled
	if config.Encryption != nil && config.Encryption.Enabled {
		if err := sm.initializeEncryption(config.Encryption); err != nil {
			return fmt.Errorf("failed to initialize encryption: %v", err)
		}
	}

	sm.auditMgr.LogSystemEvent(AuditLevelInfo, "secret.store.loaded", "system", "local", map[string]interface{}{
		"stores": len(sm.stores),
	})

	return nil
}

// createDefaultConfig creates a default secret store configuration
func (sm *SecretManager) createDefaultConfig(configPath string) error {
	defaultConfig := map[string]interface{}{
		"default_store": "file",
		"stores": map[string]*SecretStore{
			"file": {
				Type: "file",
				Config: map[string]interface{}{
					"path": "~/.portunix/secrets",
				},
				Encryption: &EncryptionConfig{
					Enabled:   true,
					Algorithm: "aes-256-gcm",
					KeySource: "env",
				},
			},
			"env": {
				Type: "env",
				Config: map[string]interface{}{
					"prefix": "PTX_SECRET_",
				},
			},
		},
		"encryption": &EncryptionConfig{
			Enabled:   true,
			Algorithm: "aes-256-gcm",
			KeySource: "env",
		},
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0700)

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// initializeEncryption sets up encryption for the secret manager
func (sm *SecretManager) initializeEncryption(config *EncryptionConfig) error {
	switch config.KeySource {
	case "env":
		keyStr := os.Getenv("PTX_ENCRYPTION_KEY")
		if keyStr == "" {
			return fmt.Errorf("PTX_ENCRYPTION_KEY environment variable not set")
		}

		// Derive 32-byte key from string
		hash := sha256.Sum256([]byte(keyStr))
		sm.encryptionKey = hash[:]

	case "file":
		if config.KeyPath == "" {
			return fmt.Errorf("key_path required for file key source")
		}

		keyData, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to read encryption key file: %v", err)
		}

		// Derive 32-byte key
		hash := sha256.Sum256(keyData)
		sm.encryptionKey = hash[:]

	default:
		return fmt.Errorf("unsupported key source: %s", config.KeySource)
	}

	return nil
}

// ProcessSecretsInPlaybook processes secrets in a .ptxbook file
func (sm *SecretManager) ProcessSecretsInPlaybook(ptxbook *PtxbookFile, context *ExecutionContext) error {
	if ptxbook.Spec.Secrets == nil {
		return nil // No secrets to process
	}

	// Access control check would be implemented here

	// Process secret references in variables
	if err := sm.processVariableSecrets(ptxbook.Spec.Variables, context); err != nil {
		return fmt.Errorf("failed to process variable secrets: %v", err)
	}

	// Process secret references in packages
	if ptxbook.Spec.Portunix != nil {
		for i := range ptxbook.Spec.Portunix.Packages {
			if err := sm.processPackageSecrets(&ptxbook.Spec.Portunix.Packages[i], context); err != nil {
				return fmt.Errorf("failed to process package secrets: %v", err)
			}
		}
	}

	sm.auditMgr.LogSystemEvent(AuditLevelInfo, "secret.references.processed", "system", "local", map[string]interface{}{
		"playbook": ptxbook.Metadata.Name,
	})

	return nil
}

// ProcessSecretReferences processes secret references in a .ptxbook file
func (sm *SecretManager) ProcessSecretReferences(ptxbook *PtxbookFile) error {
	// Create execution context
	context := &ExecutionContext{
		User:        "system",
		Environment: "local",
	}

	// Check if secrets section exists
	if ptxbook.Spec.Secrets == nil {
		return nil // No secrets to process
	}

	// Access control check would be implemented here

	// Process secret references in variables
	if err := sm.processVariableSecrets(ptxbook.Spec.Variables, context); err != nil {
		return fmt.Errorf("failed to process variable secrets: %v", err)
	}

	// Process secret references in Ansible playbooks
	if ptxbook.Spec.Ansible != nil {
		for _, playbook := range ptxbook.Spec.Ansible.Playbooks {
			if err := sm.processVariableSecrets(playbook.Vars, context); err != nil {
				return fmt.Errorf("failed to process playbook secrets: %v", err)
			}
		}
	}

	sm.auditMgr.LogSystemEvent(AuditLevelInfo, "secret.references.processed", "system", "local", map[string]interface{}{
		"playbook": ptxbook.Metadata.Name,
	})

	return nil
}

// processVariableSecrets processes secret references in variables
func (sm *SecretManager) processVariableSecrets(variables map[string]interface{}, context *ExecutionContext) error {
	for key, value := range variables {
		if strValue, ok := value.(string); ok {
			if resolved, err := sm.resolveSecretReference(strValue, context); err != nil {
				return fmt.Errorf("failed to resolve secret in variable %s: %v", key, err)
			} else if resolved != strValue {
				variables[key] = resolved
			}
		}
	}
	return nil
}

// processPackageSecrets processes secret references in package configuration
func (sm *SecretManager) processPackageSecrets(pkg *PtxbookPackage, context *ExecutionContext) error {
	// Process package name
	if resolved, err := sm.resolveSecretReference(pkg.Name, context); err != nil {
		return err
	} else {
		pkg.Name = resolved
	}

	// Process variant
	if resolved, err := sm.resolveSecretReference(pkg.Variant, context); err != nil {
		return err
	} else {
		pkg.Variant = resolved
	}

	// Process package-specific variables
	if pkg.Vars != nil {
		for key, value := range pkg.Vars {
			if strValue, ok := value.(string); ok {
				if resolved, err := sm.resolveSecretReference(strValue, context); err != nil {
					return fmt.Errorf("failed to resolve secret in package var %s: %v", key, err)
				} else if resolved != strValue {
					pkg.Vars[key] = resolved
				}
			}
		}
	}

	return nil
}

// resolveSecretReference resolves secret references in format {{ secret:store:key }}
func (sm *SecretManager) resolveSecretReference(input string, context *ExecutionContext) (string, error) {
	// Pattern for secret references: {{ secret:store:key }} or {{ secret:key }}
	secretPattern := regexp.MustCompile(`\{\{\s*secret:([^:}]+):([^}]+)\s*\}\}|\{\{\s*secret:([^}]+)\s*\}\}`)

	return secretPattern.ReplaceAllStringFunc(input, func(match string) string {
		matches := secretPattern.FindStringSubmatch(match)

		var store, key string
		if matches[1] != "" && matches[2] != "" {
			// Format: {{ secret:store:key }}
			store = strings.TrimSpace(matches[1])
			key = strings.TrimSpace(matches[2])
		} else if matches[3] != "" {
			// Format: {{ secret:key }}
			store = sm.defaultStore
			key = strings.TrimSpace(matches[3])
		} else {
			return match // Return original if no match
		}

		// Retrieve secret value
		value, err := sm.GetSecret(store, key, context)
		if err != nil {
			// Log error but don't expose in template
			sm.auditMgr.LogSecretAccess(context.User, context.Environment, store, key, false)
			return fmt.Sprintf("{{ SECRET_ERROR:%s }}", key)
		}

		sm.auditMgr.LogSecretAccess(context.User, context.Environment, store, key, true)

		return value
	}), nil
}

// GetSecret retrieves a secret from the specified store
func (sm *SecretManager) GetSecret(storeName, key string, context *ExecutionContext) (string, error) {
	store, exists := sm.stores[storeName]
	if !exists {
		return "", fmt.Errorf("secret store '%s' not found", storeName)
	}

	switch store.Type {
	case "file":
		return sm.getFileSecret(store, key, context)
	case "env":
		return sm.getEnvSecret(store, key, context)
	case "vault":
		return sm.getVaultSecret(store, key, context)
	default:
		return "", fmt.Errorf("unsupported secret store type: %s", store.Type)
	}
}

// getFileSecret retrieves a secret from file-based storage
func (sm *SecretManager) getFileSecret(store *SecretStore, key string, context *ExecutionContext) (string, error) {
	basePath, ok := store.Config["path"].(string)
	if !ok {
		return "", fmt.Errorf("invalid file store configuration: missing path")
	}

	// Expand home directory
	if strings.HasPrefix(basePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		basePath = filepath.Join(home, basePath[2:])
	}

	secretFile := filepath.Join(basePath, fmt.Sprintf("%s.secret", key))

	if _, err := os.Stat(secretFile); os.IsNotExist(err) {
		return "", fmt.Errorf("secret '%s' not found", key)
	}

	data, err := os.ReadFile(secretFile)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file: %v", err)
	}

	// Decrypt if necessary
	if store.Encryption != nil && store.Encryption.Enabled && sm.encryptionKey != nil {
		decrypted, err := sm.decrypt(data)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt secret: %v", err)
		}
		return string(decrypted), nil
	}

	return string(data), nil
}

// getEnvSecret retrieves a secret from environment variables
func (sm *SecretManager) getEnvSecret(store *SecretStore, key string, context *ExecutionContext) (string, error) {
	prefix, ok := store.Config["prefix"].(string)
	if !ok {
		prefix = "PTX_SECRET_"
	}

	envKey := prefix + strings.ToUpper(key)
	value := os.Getenv(envKey)

	if value == "" {
		return "", fmt.Errorf("environment secret '%s' not found", envKey)
	}

	return value, nil
}

// getVaultSecret retrieves a secret from HashiCorp Vault
func (sm *SecretManager) getVaultSecret(store *SecretStore, key string, context *ExecutionContext) (string, error) {
	// Placeholder for HashiCorp Vault integration
	// In a full implementation, this would use the Vault API
	return "", fmt.Errorf("HashiCorp Vault integration not yet implemented")
}

// SetSecret stores a secret in the specified store
func (sm *SecretManager) SetSecret(storeName, key, value string, context *ExecutionContext) error {
	store, exists := sm.stores[storeName]
	if !exists {
		return fmt.Errorf("secret store '%s' not found", storeName)
	}

	// Check permissions
	// Access control check would be implemented here

	switch store.Type {
	case "file":
		return sm.setFileSecret(store, key, value, context)
	case "env":
		return fmt.Errorf("environment store is read-only")
	case "vault":
		return sm.setVaultSecret(store, key, value, context)
	default:
		return fmt.Errorf("unsupported secret store type: %s", store.Type)
	}
}

// setFileSecret stores a secret in file-based storage
func (sm *SecretManager) setFileSecret(store *SecretStore, key, value string, context *ExecutionContext) error {
	basePath, ok := store.Config["path"].(string)
	if !ok {
		return fmt.Errorf("invalid file store configuration: missing path")
	}

	// Expand home directory
	if strings.HasPrefix(basePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}
		basePath = filepath.Join(home, basePath[2:])
	}

	// Ensure directory exists
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return fmt.Errorf("failed to create secret directory: %v", err)
	}

	secretFile := filepath.Join(basePath, fmt.Sprintf("%s.secret", key))

	var data []byte
	if store.Encryption != nil && store.Encryption.Enabled && sm.encryptionKey != nil {
		// Encrypt the value
		encrypted, err := sm.encrypt([]byte(value))
		if err != nil {
			return fmt.Errorf("failed to encrypt secret: %v", err)
		}
		data = encrypted
	} else {
		data = []byte(value)
	}

	if err := os.WriteFile(secretFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write secret file: %v", err)
	}

	sm.auditMgr.LogSystemEvent(AuditLevelInfo, "secret.stored", context.User, context.Environment, map[string]interface{}{
		"store": "file",
		"key":   key,
	})

	return nil
}

// setVaultSecret stores a secret in HashiCorp Vault
func (sm *SecretManager) setVaultSecret(store *SecretStore, key, value string, context *ExecutionContext) error {
	// Placeholder for HashiCorp Vault integration
	return fmt.Errorf("HashiCorp Vault integration not yet implemented")
}

// encrypt encrypts data using AES-GCM
func (sm *SecretManager) encrypt(plaintext []byte) ([]byte, error) {
	if sm.encryptionKey == nil {
		return nil, fmt.Errorf("encryption key not available")
	}

	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode as base64 for safe storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return []byte(encoded), nil
}

// decrypt decrypts data using AES-GCM
func (sm *SecretManager) decrypt(ciphertext []byte) ([]byte, error) {
	if sm.encryptionKey == nil {
		return nil, fmt.Errorf("encryption key not available")
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(decoded) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid ciphertext length")
	}

	nonce := decoded[:gcm.NonceSize()]
	ciphertext = decoded[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// ClearSecretFromMemory securely clears secret from memory
func (sm *SecretManager) ClearSecretFromMemory(secret string) {
	// Overwrite string memory with zeros
	// Note: In Go, strings are immutable, so this is limited in effectiveness
	// For production, use byte slices and explicit memory clearing
	_ = secret // Placeholder for secure memory clearing
}

// ValidateSecretConfiguration validates secret store configuration
func (sm *SecretManager) ValidateSecretConfiguration() error {
	if len(sm.stores) == 0 {
		return fmt.Errorf("no secret stores configured")
	}

	if sm.defaultStore == "" {
		return fmt.Errorf("no default secret store specified")
	}

	if _, exists := sm.stores[sm.defaultStore]; !exists {
		return fmt.Errorf("default secret store '%s' not found", sm.defaultStore)
	}

	return nil
}

// ExecutionContext provides context for secret operations
type ExecutionContext struct {
	User        string
	Role        string
	Environment string
	RequestID   string
	Timestamp   string
}

// SecretConfig represents the configuration for the secret management system
type SecretConfig struct {
	DefaultStore  string                    `json:"default_store"`
	Stores        map[string]*SecretStore   `json:"stores"`
	Encryption    *EncryptionConfig         `json:"encryption"`
}

// GetDefaultSecretConfig returns the default secret management configuration
func GetDefaultSecretConfig() *SecretConfig {
	homeDir, _ := os.UserHomeDir()
	secretsDir := filepath.Join(homeDir, ".portunix", "secrets")

	return &SecretConfig{
		DefaultStore: "file",
		Stores: map[string]*SecretStore{
			"file": {
				Type: "file",
				Config: map[string]interface{}{
					"path": secretsDir,
				},
				Encryption: &EncryptionConfig{
					Enabled:   true,
					Algorithm: "aes-256-gcm",
					KeySource: "env",
				},
			},
			"env": {
				Type: "env",
				Config: map[string]interface{}{},
				Encryption: &EncryptionConfig{
					Enabled:   false,
					Algorithm: "none",
					KeySource: "none",
				},
			},
		},
		Encryption: &EncryptionConfig{
			Enabled:   true,
			Algorithm: "aes-256-gcm",
			KeySource: "env",
		},
	}
}