package main

import (
	"encoding/base64"
	"testing"
)

func TestCryptoServiceEncryptDecrypt(t *testing.T) {
	seed := "TESTHOST|testuser|Windows 10|C:\\Users\\testuser|portunix-credential"
	crypto := NewCryptoService(seed)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple", "hello world"},
		{"api key", "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
		{"json", `{"accessToken":"abc123","refreshToken":"xyz789"}`},
		{"unicode", "Héllo Wörld 你好世界"},
		{"empty", ""},
		{"long", string(make([]byte, 10000))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := crypto.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Verify it's base64 encoded
			_, err = base64.StdEncoding.DecodeString(encrypted)
			if err != nil {
				t.Fatalf("Encrypted output is not valid base64: %v", err)
			}

			// Decrypt
			decrypted, err := crypto.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Verify roundtrip
			if decrypted != tt.plaintext {
				t.Errorf("Roundtrip failed: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestCryptoServiceEncryptWithIV(t *testing.T) {
	seed := "TESTHOST|testuser|Windows 10|C:\\Users\\testuser|portunix-credential"
	crypto := NewCryptoService(seed)

	plaintext := "test-api-key-12345"
	iv := make([]byte, 12)
	// Use fixed IV for reproducibility
	for i := range iv {
		iv[i] = byte(i)
	}

	// Encrypt with fixed IV
	encrypted1, err := crypto.EncryptWithIV(plaintext, iv)
	if err != nil {
		t.Fatalf("EncryptWithIV failed: %v", err)
	}

	// Encrypt again with same IV should produce same result
	encrypted2, err := crypto.EncryptWithIV(plaintext, iv)
	if err != nil {
		t.Fatalf("EncryptWithIV failed: %v", err)
	}

	if encrypted1 != encrypted2 {
		t.Errorf("EncryptWithIV should be deterministic with same IV")
	}

	// Decrypt should work
	decrypted, err := crypto.Decrypt(encrypted1)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestCryptoServiceInvalidIV(t *testing.T) {
	seed := "TESTHOST|testuser|Windows 10|C:\\Users\\testuser|portunix-credential"
	crypto := NewCryptoService(seed)

	// Test invalid IV length
	invalidIV := make([]byte, 10) // Should be 12
	_, err := crypto.EncryptWithIV("test", invalidIV)
	if err == nil {
		t.Error("Expected error for invalid IV length")
	}
}

func TestCryptoServiceDecryptInvalid(t *testing.T) {
	seed := "TESTHOST|testuser|Windows 10|C:\\Users\\testuser|portunix-credential"
	crypto := NewCryptoService(seed)

	tests := []struct {
		name  string
		input string
	}{
		{"not base64", "not valid base64!!!"},
		{"too short", base64.StdEncoding.EncodeToString([]byte("short"))},
		{"corrupted", base64.StdEncoding.EncodeToString(make([]byte, 100))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := crypto.Decrypt(tt.input)
			if err == nil {
				t.Error("Expected error for invalid input")
			}
		})
	}
}

func TestCryptoServiceDifferentSeeds(t *testing.T) {
	seed1 := "HOST1|user1|Windows 10|C:\\Users\\user1|portunix-credential"
	seed2 := "HOST2|user2|Windows 10|C:\\Users\\user2|portunix-credential"

	crypto1 := NewCryptoService(seed1)
	crypto2 := NewCryptoService(seed2)

	plaintext := "secret-api-key"

	// Encrypt with first seed
	encrypted, err := crypto1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Decrypting with different seed should fail
	_, err = crypto2.Decrypt(encrypted)
	if err == nil {
		t.Error("Expected error when decrypting with different seed")
	}

	// Decrypting with same seed should work
	decrypted, err := crypto1.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestGenerateSeed(t *testing.T) {
	seed := GenerateSeed("myhost", "myuser", "Windows 10", "C:\\Users\\myuser", "portunix-credential")
	expected := "myhost|myuser|Windows 10|C:\\Users\\myuser|portunix-credential"
	if seed != expected {
		t.Errorf("GenerateSeed: got %q, want %q", seed, expected)
	}
}

func TestGenerateM365Seed(t *testing.T) {
	seed := GenerateM365Seed("myhost", "myuser", "Windows 10", "C:\\Users\\myuser")
	expected := "myhost|myuser|Windows 10|C:\\Users\\myuser|portunix-m365"
	if seed != expected {
		t.Errorf("GenerateM365Seed: got %q, want %q", seed, expected)
	}
}

func TestGeneratePasswordSeed(t *testing.T) {
	seed := GeneratePasswordSeed("myhost", "myuser", "Windows 10", "C:\\Users\\myuser", "mypassword")
	expected := "myhost|myuser|Windows 10|C:\\Users\\myuser|portunix-credential|pw:mypassword"
	if seed != expected {
		t.Errorf("GeneratePasswordSeed: got %q, want %q", seed, expected)
	}
}

func TestSecureWipe(t *testing.T) {
	data := []byte("sensitive data here")
	original := make([]byte, len(data))
	copy(original, data)

	SecureWipe(data)

	// Verify all bytes are zeroed
	for i, b := range data {
		if b != 0 {
			t.Errorf("SecureWipe: byte %d not zeroed, got %d", i, b)
		}
	}
}

// BenchmarkEncrypt benchmarks encryption performance
func BenchmarkEncrypt(b *testing.B) {
	seed := "TESTHOST|testuser|Windows 10|C:\\Users\\testuser|portunix-credential"
	crypto := NewCryptoService(seed)
	plaintext := "benchmark-test-api-key-1234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = crypto.Encrypt(plaintext)
	}
}

// BenchmarkDecrypt benchmarks decryption performance
func BenchmarkDecrypt(b *testing.B) {
	seed := "TESTHOST|testuser|Windows 10|C:\\Users\\testuser|portunix-credential"
	crypto := NewCryptoService(seed)
	plaintext := "benchmark-test-api-key-1234567890"
	encrypted, _ := crypto.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = crypto.Decrypt(encrypted)
	}
}
