package update

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"strings"
)

// PublicKeyHex is the hex-encoded Ed25519 public key used to verify update
// signatures. Empty value disables signature verification entirely (development
// builds, no key configured yet). Real key is injected in release builds.
//
// Generated: TBD (placeholder until first signed release)
// To populate: generate keypair with `openssl genpkey -algorithm Ed25519` or
// equivalent, then embed the 32-byte public key here as 64 hex chars.
var PublicKeyHex = ""

// RequireSignature controls whether a missing signature aborts the update.
// false = warn-only (backward compatibility for releases without .sig file).
// true  = enforce — abort if signature is missing or invalid.
var RequireSignature = false

// HasPublicKey reports whether a public key is configured for verification.
func HasPublicKey() bool {
	return strings.TrimSpace(PublicKeyHex) != ""
}

// GetPublicKey decodes and returns the embedded Ed25519 public key.
func GetPublicKey() (ed25519.PublicKey, error) {
	if !HasPublicKey() {
		return nil, fmt.Errorf("no public key configured")
	}

	raw, err := hex.DecodeString(strings.TrimSpace(PublicKeyHex))
	if err != nil {
		return nil, fmt.Errorf("invalid public key hex: %w", err)
	}

	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d bytes, want %d",
			len(raw), ed25519.PublicKeySize)
	}

	return ed25519.PublicKey(raw), nil
}

// VerifySignature verifies a detached Ed25519 signature over data using the
// embedded public key. Returns nil on success, error on failure.
func VerifySignature(data, signature []byte) error {
	pubKey, err := GetPublicKey()
	if err != nil {
		return err
	}

	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature size: got %d bytes, want %d",
			len(signature), ed25519.SignatureSize)
	}

	if !ed25519.Verify(pubKey, data, signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// VerifyChecksumsSignature downloads the signature file from signatureURL and
// verifies it against checksumsData using the embedded public key.
//
// Behavior when signatureURL is empty (release without .sig asset):
//   - RequireSignature == false → returns nil (warn handled by caller)
//   - RequireSignature == true  → returns error
//
// Behavior when no public key is configured:
//   - returns nil (verification disabled, warn handled by caller)
func VerifyChecksumsSignature(checksumsData []byte, signatureURL string) error {
	if !HasPublicKey() {
		// Verification disabled — caller decides whether to warn.
		return nil
	}

	if signatureURL == "" {
		if RequireSignature {
			return fmt.Errorf("no signature available and signature is required")
		}
		return nil
	}

	signature, err := DownloadFile(signatureURL)
	if err != nil {
		return fmt.Errorf("failed to download signature: %w", err)
	}

	// Signature files may be hex-encoded with trailing whitespace; tolerate it.
	signature = decodeSignatureBytes(signature)

	return VerifySignature(checksumsData, signature)
}

// decodeSignatureBytes accepts either raw 64-byte signature or hex-encoded
// (128 chars + optional whitespace). Returns raw bytes for ed25519.Verify.
func decodeSignatureBytes(raw []byte) []byte {
	trimmed := strings.TrimSpace(string(raw))

	// Try hex decode first (common format for sig files).
	if decoded, err := hex.DecodeString(trimmed); err == nil &&
		len(decoded) == ed25519.SignatureSize {
		return decoded
	}

	return raw
}
