package update

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

// withTestKey installs a freshly generated keypair for the duration of a test
// and restores the previous public-key state on cleanup.
func withTestKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	prevHex := PublicKeyHex
	prevReq := RequireSignature
	PublicKeyHex = hex.EncodeToString(pub)
	t.Cleanup(func() {
		PublicKeyHex = prevHex
		RequireSignature = prevReq
	})

	return priv
}

func TestHasPublicKey_EmptyByDefault(t *testing.T) {
	prev := PublicKeyHex
	PublicKeyHex = ""
	defer func() { PublicKeyHex = prev }()

	if HasPublicKey() {
		t.Fatal("HasPublicKey should be false when PublicKeyHex is empty")
	}
}

func TestGetPublicKey_InvalidHex(t *testing.T) {
	prev := PublicKeyHex
	PublicKeyHex = "not-hex-at-all"
	defer func() { PublicKeyHex = prev }()

	if _, err := GetPublicKey(); err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestGetPublicKey_WrongSize(t *testing.T) {
	prev := PublicKeyHex
	PublicKeyHex = "deadbeef" // 4 bytes, not 32
	defer func() { PublicKeyHex = prev }()

	if _, err := GetPublicKey(); err == nil {
		t.Fatal("expected error for wrong key size")
	}
}

func TestVerifySignature_Valid(t *testing.T) {
	priv := withTestKey(t)

	data := []byte("a1b2c3  portunix_1.0.0_linux_amd64.tar.gz\n")
	sig := ed25519.Sign(priv, data)

	if err := VerifySignature(data, sig); err != nil {
		t.Fatalf("expected valid signature to verify, got: %v", err)
	}
}

func TestVerifySignature_TamperedData(t *testing.T) {
	priv := withTestKey(t)

	data := []byte("original")
	sig := ed25519.Sign(priv, data)

	if err := VerifySignature([]byte("tampered"), sig); err == nil {
		t.Fatal("expected verification failure for tampered data")
	}
}

func TestVerifySignature_WrongKey(t *testing.T) {
	priv1 := withTestKey(t)

	// Replace public key with a different one without changing private key.
	pub2, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate second key: %v", err)
	}
	PublicKeyHex = hex.EncodeToString(pub2)

	data := []byte("payload")
	sig := ed25519.Sign(priv1, data)

	if err := VerifySignature(data, sig); err == nil {
		t.Fatal("expected verification failure when signed with different key")
	}
}

func TestVerifySignature_MalformedSignatureSize(t *testing.T) {
	withTestKey(t)

	if err := VerifySignature([]byte("data"), []byte{0x01, 0x02}); err == nil {
		t.Fatal("expected error for too-short signature")
	}
}

func TestVerifyChecksumsSignature_NoKey_Skips(t *testing.T) {
	prev := PublicKeyHex
	PublicKeyHex = ""
	defer func() { PublicKeyHex = prev }()

	// Should return nil even with empty signatureURL when no key is configured.
	if err := VerifyChecksumsSignature([]byte("checksums"), ""); err != nil {
		t.Fatalf("expected nil when no key configured, got: %v", err)
	}
}

func TestVerifyChecksumsSignature_MissingSig_BackwardCompat(t *testing.T) {
	withTestKey(t)
	RequireSignature = false

	if err := VerifyChecksumsSignature([]byte("checksums"), ""); err != nil {
		t.Fatalf("expected nil when sig missing and not required, got: %v", err)
	}
}

func TestVerifyChecksumsSignature_MissingSig_Required(t *testing.T) {
	withTestKey(t)
	RequireSignature = true

	if err := VerifyChecksumsSignature([]byte("checksums"), ""); err == nil {
		t.Fatal("expected error when sig missing and RequireSignature is true")
	}
}

func TestDecodeSignatureBytes_Hex(t *testing.T) {
	// 64-byte signature, hex-encoded with trailing newline.
	raw := make([]byte, ed25519.SignatureSize)
	for i := range raw {
		raw[i] = byte(i)
	}
	hexed := append([]byte(hex.EncodeToString(raw)), '\n')

	got := decodeSignatureBytes(hexed)
	if len(got) != ed25519.SignatureSize {
		t.Fatalf("expected %d bytes after decode, got %d", ed25519.SignatureSize, len(got))
	}
	for i := range got {
		if got[i] != raw[i] {
			t.Fatalf("byte %d: expected %d, got %d", i, raw[i], got[i])
		}
	}
}

func TestDecodeSignatureBytes_Raw(t *testing.T) {
	raw := make([]byte, ed25519.SignatureSize)
	raw[0] = 0xFF
	got := decodeSignatureBytes(raw)
	if len(got) != ed25519.SignatureSize || got[0] != 0xFF {
		t.Fatalf("raw bytes should pass through unchanged")
	}
}

func TestGetSignatureName(t *testing.T) {
	if got, want := GetSignatureName("v1.10.8"), "checksums_1.10.8.txt.sig"; got != want {
		t.Fatalf("GetSignatureName: got %q, want %q", got, want)
	}
}
