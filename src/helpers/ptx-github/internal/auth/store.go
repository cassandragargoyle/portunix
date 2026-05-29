/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Account represents a single stored GitHub account.
type Account struct {
	Name           string `json:"name"`
	EncryptedToken string `json:"encrypted_token"`
	Username       string `json:"username,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	PasswordSet    bool   `json:"password_protected,omitempty"`
}

// File is the JSON layout written to disk.
type File struct {
	Version  int                 `json:"version"`
	Default  string              `json:"default,omitempty"`
	Accounts map[string]*Account `json:"accounts"`
}

// Store manages encrypted GitHub authentication entries on disk.
type Store struct {
	path   string
	crypto *CryptoService
}

// DefaultStorePath returns the path used by the helper unless overridden.
func DefaultStorePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".portunix", "github", "auth.json"), nil
}

// NewStore opens (or initializes) a store at the given path.
// If password is non-empty the encryption key is also derived from it.
func NewStore(path, password string) (*Store, error) {
	if path == "" {
		var err error
		path, err = DefaultStorePath()
		if err != nil {
			return nil, fmt.Errorf("resolve default store path: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}

	seed, err := MachineSeed(password)
	if err != nil {
		return nil, fmt.Errorf("generate seed: %w", err)
	}

	return &Store{
		path:   path,
		crypto: NewCryptoService(seed),
	}, nil
}

// Path returns the underlying file path.
func (s *Store) Path() string {
	return s.path
}

func (s *Store) load() (*File, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{Version: 1, Accounts: map[string]*Account{}}, nil
		}
		return nil, fmt.Errorf("read store: %w", err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse store: %w", err)
	}
	if f.Accounts == nil {
		f.Accounts = map[string]*Account{}
	}
	if f.Version == 0 {
		f.Version = 1
	}
	return &f, nil
}

func (s *Store) save(f *File) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal store: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace store: %w", err)
	}
	return nil
}

// SaveToken stores (or replaces) the named account, encrypting the token.
// If makeDefault is true, the account becomes the default for lookups.
// passwordProtected is recorded for display only — actual key derivation
// is controlled by the password passed to NewStore.
func (s *Store) SaveToken(name, token, username string, makeDefault, passwordProtected bool) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	enc, err := s.crypto.Encrypt(token)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	acc, exists := f.Accounts[name]
	if !exists {
		acc = &Account{Name: name, CreatedAt: now}
	}
	acc.EncryptedToken = enc
	acc.Username = username
	acc.UpdatedAt = now
	acc.PasswordSet = passwordProtected
	f.Accounts[name] = acc
	if makeDefault || f.Default == "" {
		f.Default = name
	}
	return s.save(f)
}

// GetToken returns the decrypted token for the named account.
// Empty name resolves to the default account.
func (s *Store) GetToken(name string) (token string, account *Account, err error) {
	f, err := s.load()
	if err != nil {
		return "", nil, err
	}
	if name == "" {
		name = f.Default
	}
	if name == "" {
		return "", nil, fmt.Errorf("no GitHub account configured")
	}
	acc, ok := f.Accounts[name]
	if !ok {
		return "", nil, fmt.Errorf("account %q not found", name)
	}
	plain, err := s.crypto.Decrypt(acc.EncryptedToken)
	if err != nil {
		return "", nil, fmt.Errorf("decrypt token (wrong password?): %w", err)
	}
	return plain, acc, nil
}

// Delete removes an account. If the deleted account was the default,
// the default is cleared (or moved to the first remaining account).
func (s *Store) Delete(name string) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := f.Accounts[name]; !ok {
		return fmt.Errorf("account %q not found", name)
	}
	delete(f.Accounts, name)
	if f.Default == name {
		f.Default = ""
		for k := range f.Accounts {
			f.Default = k
			break
		}
	}
	return s.save(f)
}

// SetDefault changes the default account.
func (s *Store) SetDefault(name string) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := f.Accounts[name]; !ok {
		return fmt.Errorf("account %q not found", name)
	}
	f.Default = name
	return s.save(f)
}

// List returns all accounts and the default account name.
func (s *Store) List() (accounts []*Account, defaultName string, err error) {
	f, err := s.load()
	if err != nil {
		return nil, "", err
	}
	for _, a := range f.Accounts {
		accounts = append(accounts, a)
	}
	return accounts, f.Default, nil
}
