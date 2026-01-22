package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Cryptographic parameters - must match Java TokenStorage.java
	keyLength      = 32    // AES-256 (256 bits = 32 bytes)
	ivLength       = 12    // GCM standard IV length
	gcmTagLength   = 16    // 128-bit authentication tag
	pbkdf2Iterations = 65536 // Balance security/performance
)

// CryptoService handles encryption and decryption of credentials
type CryptoService struct {
	key []byte
}

// NewCryptoService creates a new crypto service with the given seed
func NewCryptoService(seed string) *CryptoService {
	key := deriveKey(seed)
	return &CryptoService{key: key}
}

// deriveKey derives an AES-256 key from seed using PBKDF2-HMAC-SHA256
// This must match the Java implementation exactly:
// seed = "{hostname}|{username}|{os_name}|{home_dir}|portunix-credential"
// salt = seed.getBytes()
// key = PBKDF2(seed, salt, 65536, 32, SHA256)
func deriveKey(seed string) []byte {
	seedBytes := []byte(seed)
	// In Java implementation, salt = seed.getBytes()
	salt := seedBytes
	return pbkdf2.Key(seedBytes, salt, pbkdf2Iterations, keyLength, sha256.New)
}

// Encrypt encrypts plaintext using AES-256-GCM
// Returns base64 encoded: IV[12 bytes] || Ciphertext || GCM_AuthTag[16 bytes]
func (cs *CryptoService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(cs.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random IV
	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt: result is ciphertext || authTag
	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)

	// Combine: IV || ciphertext || authTag
	result := make([]byte, ivLength+len(ciphertext))
	copy(result[:ivLength], iv)
	copy(result[ivLength:], ciphertext)

	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt decrypts base64 encoded ciphertext using AES-256-GCM
// Expects format: Base64(IV[12 bytes] || Ciphertext || GCM_AuthTag[16 bytes])
func (cs *CryptoService) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(data) < ivLength+gcmTagLength {
		return "", fmt.Errorf("invalid ciphertext: too short")
	}

	block, err := aes.NewCipher(cs.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract IV and ciphertext
	iv := data[:ivLength]
	ciphertext := data[ivLength:]

	// Decrypt
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptWithIV encrypts plaintext using a specific IV (for testing/compatibility)
func (cs *CryptoService) EncryptWithIV(plaintext string, iv []byte) (string, error) {
	if len(iv) != ivLength {
		return "", fmt.Errorf("invalid IV length: expected %d, got %d", ivLength, len(iv))
	}

	block, err := aes.NewCipher(cs.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt with provided IV
	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)

	// Combine: IV || ciphertext || authTag
	result := make([]byte, ivLength+len(ciphertext))
	copy(result[:ivLength], iv)
	copy(result[ivLength:], ciphertext)

	return base64.StdEncoding.EncodeToString(result), nil
}

// SecureWipe overwrites sensitive data in memory
func SecureWipe(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// GenerateSeed generates the seed string for key derivation
// Format: "{hostname}|{username}|{os_name}|{home_dir}|portunix-credential"
func GenerateSeed(hostname, username, osName, homeDir, suffix string) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", hostname, username, osName, homeDir, suffix)
}

// GenerateM365Seed generates the seed string for M365 compatibility mode
// Format: "{hostname}|{username}|{os_name}|{home_dir}|portunix-m365"
func GenerateM365Seed(hostname, username, osName, homeDir string) string {
	return GenerateSeed(hostname, username, osName, homeDir, "portunix-m365")
}

// GeneratePasswordSeed generates the seed string with password
// Format: "{hostname}|{username}|{os_name}|{home_dir}|portunix-credential|pw:{password}"
func GeneratePasswordSeed(hostname, username, osName, homeDir, password string) string {
	return fmt.Sprintf("%s|%s|%s|%s|portunix-credential|pw:%s", hostname, username, osName, homeDir, password)
}
