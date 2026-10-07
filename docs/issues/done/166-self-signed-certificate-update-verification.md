# Issue #166: Self-Signed Certificate for Update Verification

**Type**: Enhancement
**Priority**: Medium
**Labels**: enhancement, security, update, code-signing, supply-chain
**Status**: ✅ Implemented (Phase 1 — Go verification code)
**Created**: 2026-04-03
**Closed**: 2026-04-22

## Implementation Summary

Phase 1 (Go signing/verification code) implemented and merged on 2026-04-22.
Branch: `feature/166-ed25519-update-signature-verification`.
Acceptance protocol: [`docs/testing/internal/acceptance-166.md`](../../testing/internal/acceptance-166.md) — PASS.

**Delivered in this issue:**

- `src/app/update/signing.go` — `crypto/ed25519` verification (`PublicKeyHex`,
  `RequireSignature`, `VerifySignature`, `VerifyChecksumsSignature`)
- `ReleaseInfo.SignatureURL` + asset lookup for `checksums_*.txt.sig`
- `VerifyArchiveChecksumSigned` — verifies signature before parsing checksum
- CLI surfaces "Signature verified" / warn for unsigned releases
- 13 unit tests covering positive/negative scenarios
- Backward compatible: empty `PublicKeyHex` disables verification entirely
  (no-op runtime until Phase 2 lands real key)

**Follow-up work — separate issues required:**

- Phase 2: Generate production Ed25519 keypair, populate `PublicKeyHex`, sign
  `checksums_*.txt` in `scripts/make-release.py`, upload `.sig` via `.goreleaser.yml`
- Phase 3: Flip `RequireSignature = true` after ≥1 signed release cycle
- Phase 4: Key rotation mechanism (multi-key support)

## Problem Description

Currently `portunix update` downloads new binaries from GitHub Releases and verifies
integrity using SHA256 checksums. However, checksums alone do not provide **authenticity
verification** — if an attacker compromises the release assets on GitHub, they can
replace both the binary and the checksum file.

## Proposed Solution

Use a self-signed Ed25519 (or RSA) keypair to sign release artifacts. The public key
is embedded in the Portunix binary at build time. During updates, the downloaded
artifacts are verified against this embedded public key.

This approach is **free** (no CA certificate needed) and provides meaningful protection
against supply chain attacks on the update channel.

### How It Works

1. **Build time**: Generate Ed25519 keypair (one-time), embed public key in Go source
2. **Release time**: Sign checksums file (or individual binaries) with private key
3. **Update time**: `portunix update` downloads signature file, verifies against
   embedded public key before applying the update

### Trust Model

The trust anchor is the initially installed binary itself. If the user trusts
the binary they installed (via install.ps1, install.bat, or manual download),
they implicitly trust the public key embedded in it. All future updates are
verified against that key.

## Implementation Steps

### Phase 1: Key Generation and Embedding

1. Generate Ed25519 keypair (one-time, store private key securely)
2. Embed public key as constant in Go source (e.g. `app/update/signing.go`)
3. Implement signature verification function using `crypto/ed25519`

### Phase 2: Release Pipeline Integration

1. Update `scripts/make-release.py` to sign `checksums_{version}.txt` with private key
2. Output signature file: `checksums_{version}.txt.sig`
3. Include `.sig` file in GitHub Release assets via `.goreleaser.yml`

### Phase 3: Update Verification

1. Modify update download logic to also fetch `.sig` file
2. Verify signature before applying update
3. If verification fails — abort update with clear error message
4. If `.sig` file is missing (older releases) — warn but allow update (backward compatibility)

### Phase 4: Key Rotation (future)

1. Support multiple trusted public keys for key rotation
2. New binary can contain both old and new public key
3. Transition period where both keys are accepted

## Technical Notes

```go
// Example: embedded public key
package update

import "crypto/ed25519"

// PublicKey is the Ed25519 public key used to verify update signatures
// Generated: 2026-XX-XX
var PublicKey = ed25519.PublicKey{...}

func VerifySignature(data, signature []byte) bool {
    return ed25519.Verify(PublicKey, data, signature)
}
```

```python
# Example: signing in make-release.py
import nacl.signing  # or subprocess call to signify/minisign

signing_key = load_private_key("release-signing.key")
checksums = open(f"checksums_{version}.txt", "rb").read()
signature = signing_key.sign(checksums)
write_file(f"checksums_{version}.txt.sig", signature)
```

### Alternative: Use minisign/signify

Instead of custom Go implementation, consider using `minisign` (by Frank Denis):

- Battle-tested, simple, Ed25519-based
- `minisign -S -m checksums.txt` (sign)
- Go library available: `github.com/jedisct1/go-minisign`
- Used by WinGet, Zig, and other projects

## Acceptance Criteria

- [ ] Ed25519 keypair generated and private key stored securely
- [ ] Public key embedded in Portunix binary
- [ ] Release pipeline signs checksums file
- [ ] Signature file included in GitHub Release assets
- [ ] `portunix update` verifies signature before applying update
- [ ] Clear error message on signature verification failure
- [ ] Backward compatible with unsigned older releases
- [ ] Key rotation mechanism documented

## References

- Related: #164 (Windows install.ps1 Usability Overhaul)
- Related: #165 (Authenticode Code Signing Certificate)
- Go stdlib: `crypto/ed25519`
- minisign: https://jedisct1.github.io/minisign/
- Go minisign library: `github.com/jedisct1/go-minisign`
