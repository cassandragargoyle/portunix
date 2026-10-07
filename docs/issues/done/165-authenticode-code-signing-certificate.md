# Issue #165: Self-Signed Update Channel Verification

**Type**: Enhancement
**Priority**: Medium
**Labels**: enhancement, security, update, code-signing
**Status**: ✅ Implemented
**Created**: 2026-04-03
**Implemented**: 2026-05-03
**Supersedes scope**: Originally targeted Authenticode CA-cert distribution
signing — that scope is deferred (depends on project monetization) and tracked
as future work below. This issue now covers self-signed Ed25519 verification
of the `portunix update` channel, which is independent of CA certs and free
to ship.

## Problem Description

`portunix update` downloads a release archive from GitHub, validates its
sha256 against a checksums file, and applies the binary. Both the archive and
the checksums file are downloaded over HTTPS but are otherwise unauthenticated
— anything that can write to the GitHub release (compromised maintainer
account, GitHub mirror, MITM with a forged cert) can replace both files
consistently and the user installs an attacker-controlled binary.

A CA-issued code signing certificate would solve this end-to-end (and also
remove SmartScreen friction on Windows installs), but it costs USD 100–400
per year and is gated on project monetization.

A **self-signed Ed25519 signature over the checksums file** closes the
update-channel hole at zero cost. Trust anchor is the public key embedded
into the previously installed binary — the user's first install is not
defended (that's what CA certs are for), but every subsequent update is
verified against a key the maintainer holds offline.

## Solution

Embed an Ed25519 public key into the binary at build time. Sign the
release `checksums_X.Y.Z.txt` file with the matching private key during
release. `portunix update` downloads `checksums_X.Y.Z.txt.sig`, verifies
the signature **before** parsing checksums, then verifies the archive's
sha256 against the checksums file.

### What was already in place (preexisting)

The Go-side verification was implemented earlier and unused:

- `src/app/update/signing.go` — `PublicKeyHex` (ldflag-injected),
  `RequireSignature` flag, `VerifySignature`, `VerifyChecksumsSignature`,
  `decodeSignatureBytes` (accepts hex or raw)
- `src/app/update/verify.go` — `VerifyArchiveChecksumSigned` checks signature
  before trusting checksums
- `src/app/update/types.go` — `SignatureURL` field, `GetSignatureName()`
- `src/app/update/github.go` — `buildReleaseInfo` resolves `.sig` asset,
  `DownloadUpdate` invokes signed verification
- `src/app/update/signing_test.go` — 11 tests covering valid/tampered/
  wrong-key/missing-sig paths

What was missing was every link of the chain that produces, transports,
and embeds the key.

### What this issue adds

1. **`scripts/sign-release.py`** — CLI tool with three subcommands:
   - `generate-key [--write-pubkey]` — produces Ed25519 keypair, stores
     private key at `~/.config/portunix/update-signing/private.key`
     (mode 0600), prints/writes hex public key
   - `sign FILE` — produces `FILE.sig` (hex-encoded 64-byte signature)
   - `verify FILE [--signature P] [--pubkey-file P]` — sanity-checks a
     `.sig` against a public key (repo file, explicit file, or derived
     from local private key)

2. **`assets/update-signing-pubkey.txt`** — single-line hex public key,
   committed to repo (public keys carry no secret).

3. **Build-time embedding** in two paths:
   - `build-with-version.sh` — reads `assets/update-signing-pubkey.txt`,
     adds `-X portunix.ai/app/update.PublicKeyHex=<hex>` to `LDFLAGS_COMMON`,
     overridable via `PORTUNIX_UPDATE_PUBKEY` env var
   - `.goreleaser.yml` — `{{ .Env.PORTUNIX_UPDATE_PUBKEY }}` template in
     ldflags for the main `portunix` binary

4. **Release pipeline integration** in `scripts/make-release.py`:
   - `load_update_pubkey()` reads the assets file
   - exports `PORTUNIX_UPDATE_PUBKEY` before invoking goreleaser so the
     ldflag template resolves
   - `sign_checksums()` calls `sign-release.py sign` on
     `dist/checksums_X.Y.Z.txt` after `copy_quickstart_scripts()` (which
     appends quickstart hashes), producing `checksums_X.Y.Z.txt.sig`
   - resulting `.sig` ships automatically (existing upload script
     uploads everything in `dist/`)

5. **`.gitignore`** — `*.key`, `*.pem`, `private.key` exclusions to prevent
   accidental commit of the private key.

6. **`docs/contributing/UPDATE-SIGNING.md`** — threat model, architecture,
   key generation, signing workflow, runtime verification, build-time
   embedding, key rotation strategy, `RequireSignature` enforcement plan.

### Initial keypair generated

A keypair was generated locally during implementation. Public key:

```text
c70876c02d19ff4a70b4bb3f30a23b6a55c5a60b3bc527c1918c800ddcb7cd3b
```

stored in `assets/update-signing-pubkey.txt`. Private key is at
`~/.config/portunix/update-signing/private.key` on the maintainer machine
(mode 0600, never committed).

The first signed release will be the next normal release after this issue
merges. `RequireSignature` stays `false` (warn-only) for at least two
release cycles to give the field time to roll forward.

## Acceptance Criteria

- [x] Ed25519 verification logic in update flow (preexisting)
- [x] CLI tool to generate keypair (`sign-release.py generate-key`)
- [x] CLI tool to sign checksums file (`sign-release.py sign`)
- [x] CLI tool to verify signature locally (`sign-release.py verify`)
- [x] Public key stored in repo at `assets/update-signing-pubkey.txt`
- [x] Private key path documented and gitignored
- [x] `build-with-version.sh` embeds `PublicKeyHex` via ldflags
- [x] `.goreleaser.yml` embeds `PublicKeyHex` via env-var template
- [x] `make-release.py` exports pubkey before goreleaser runs
- [x] `make-release.py` signs `checksums_X.Y.Z.txt` after generation
- [x] Initial keypair generated, pubkey committed
- [x] Documentation in `docs/contributing/UPDATE-SIGNING.md`
- [x] End-to-end test: build with embedded pubkey, sign sample checksums
      with Python, verify with Go via `crypto/ed25519` — passes positive
      and negative (tampered) cases
- [ ] **Future**: Flip `RequireSignature` to `true` after two signed
      release cycles (separate small PR, document in changelog)

## Future Work — Authenticode CA Certificate (deferred)

The original scope of this issue (paid Authenticode certificate to remove
SmartScreen warnings on Windows install) is **not** implemented and remains
deferred. It is a separate concern from the update channel:

- **Update channel** (this issue): protects users who already have Portunix
  installed from receiving a malicious update. Self-signed Ed25519 is
  sufficient and free.
- **Distribution channel** (deferred): protects first-install users from
  malicious binaries on the distribution path. Requires CA-signed
  Authenticode certificate (USD 100–400/year). Depends on project
  monetization.

When monetization makes this viable, open a new issue for Authenticode
distribution signing. The two systems can coexist (Authenticode for
Windows binary trust, Ed25519 for update integrity).

## References

- Related: #164 (Windows install.ps1 Usability Overhaul)
- Related: #125 (Cross-Platform Binary Distribution)
- ADR-031: Cross-Platform Binary Distribution
- `docs/contributing/UPDATE-SIGNING.md`
