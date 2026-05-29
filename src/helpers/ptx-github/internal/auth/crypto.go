/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/user"
	"runtime"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	keyLength        = 32
	ivLength         = 12
	gcmTagLength     = 16
	pbkdf2Iterations = 65536
)

// CryptoService handles AES-256-GCM encryption.
type CryptoService struct {
	key []byte
}

// NewCryptoService derives an encryption key from the given seed.
func NewCryptoService(seed string) *CryptoService {
	seedBytes := []byte(seed)
	key := pbkdf2.Key(seedBytes, seedBytes, pbkdf2Iterations, keyLength, sha256.New)
	return &CryptoService{key: key}
}

// Encrypt returns base64(IV || ciphertext || authTag).
func (cs *CryptoService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(cs.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("generate IV: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)
	out := make([]byte, ivLength+len(ciphertext))
	copy(out[:ivLength], iv)
	copy(out[ivLength:], ciphertext)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt expects base64(IV || ciphertext || authTag).
func (cs *CryptoService) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}
	if len(data) < ivLength+gcmTagLength {
		return "", fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(cs.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	iv := data[:ivLength]
	ciphertext := data[ivLength:]
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}

// MachineSeed builds a machine-bound seed from hostname, user, OS and home dir.
// If password is non-empty, it is appended to make the key password-derived.
func MachineSeed(password string) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	username := ""
	if u, err := user.Current(); err == nil {
		username = u.Username
		if runtime.GOOS == "windows" {
			parts := strings.Split(username, `\`)
			username = parts[len(parts)-1]
		}
	} else {
		username = os.Getenv("USER")
		if username == "" {
			username = os.Getenv("USERNAME")
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
	}

	osName := osNameForSeed()
	seed := fmt.Sprintf("%s|%s|%s|%s|portunix-github", hostname, username, osName, homeDir)
	if password != "" {
		seed += "|pw:" + password
	}
	return seed, nil
}

func osNameForSeed() string {
	switch runtime.GOOS {
	case "linux":
		return "Linux"
	case "darwin":
		return "Mac OS X"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}
