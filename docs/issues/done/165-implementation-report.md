# Issue #165 — Implementation Report

**Branch**: `feature/165-update-signing`
**Implemented**: 2026-05-03
**Scope**: Self-Signed Update Channel Verification (Ed25519)

> The original issue scope (paid Authenticode CA certificate) was deferred —
> see issue file for details. This report covers the self-signed Ed25519
> update verification that was actually implemented.

## Modified Files

| File | Change |
|------|--------|
| `build-with-version.sh` | Reads `assets/update-signing-pubkey.txt` (or `PORTUNIX_UPDATE_PUBKEY` env), adds `-X portunix.ai/app/update.PublicKeyHex=…` to shared `LDFLAGS_COMMON` |
| `.goreleaser.yml` | For `portunix` builder added template ldflag `{{ with index .Env "PORTUNIX_UPDATE_PUBKEY" }}…{{ end }}` |
| `scripts/make-release.py` | `load_update_pubkey()`, exports env before goreleaser, `sign_checksums()` after quickstart copy |
| `.gitignore` | `*.key`, `*.pem`, `private.key` exclusions |
| `docs/issues/done/165-authenticode-code-signing-certificate.md` | Rewritten: Authenticode → Self-Signed Update Verification, status `✅ Implemented`, date 2026-05-03 |

## New Files

| File | Purpose |
|------|---------|
| `scripts/sign-release.py` | CLI: `generate-key` / `sign` / `verify` (Ed25519 via `cryptography`) |
| `assets/update-signing-pubkey.txt` | Hex pubkey (32B / 64 chars), source of truth for build pipeline |
| `docs/contributing/UPDATE-SIGNING.md` | Threat model, architecture, key rotation, RequireSignature plan |

## Key Decisions

1. **Reuse existing Go code** — `signing.go` / `verify.go` were already
   implemented but unused. No runtime code was written; only the missing
   tooling that surrounds them.
2. **Pubkey in repo, privkey off repo** — `assets/update-signing-pubkey.txt`
   in git (public, no secret), `~/.config/portunix/update-signing/private.key`
   off repo (mode 0600).
3. **Empty pubkey = backward compat** — when the file is empty / env unset,
   `PublicKeyHex=""`, runtime verification silently disabled (same as before
   this issue). Dev builds are not broken.
4. **Signing does not block release** — if private key is missing (CI without
   secret), `sign_checksums()` warns and release continues without `.sig`.
   `RequireSignature=false` in Go code, so old clients are not rejected.
5. **Hex as wire format** — `.sig` contains hex (128 chars + newline). Go
   code accepts both hex and raw bytes via `decodeSignatureBytes`.

## Test Results

| Test | Result |
|------|--------|
| `go test portunix.ai/app/update/...` | ✅ ok (existing 11 tests) |
| `bash build-with-version.sh v1.10.9-SNAPSHOT` | ✅ Embedded 64 hex chars into binary |
| `strings ./portunix \| grep <pubkey>` | ✅ Pubkey found in binary |
| `sign-release.py sign` → `.sig` (64B) | ✅ |
| `sign-release.py verify` (clean) | ✅ Valid |
| `sign-release.py verify` (tampered) | ✅ Detected, exit 1 |
| Cross-language: Python sign → Go `crypto/ed25519` verify | ✅ VALID |
| Cross-language: tampered → Go verify | ✅ INVALID, exit 1 |
| `make build` (regular dev build) | ✅ Works, no regression |

## Generated Keypair

A keypair was generated locally during implementation. Public key:

```text
c70876c02d19ff4a70b4bb3f30a23b6a55c5a60b3bc527c1918c800ddcb7cd3b
```

stored in `assets/update-signing-pubkey.txt`. Private key at
`~/.config/portunix/update-signing/private.key` on the maintainer machine
(mode 0600, never committed).

The first signed release will be the next normal release after this issue
merges. `RequireSignature` stays `false` (warn-only) for at least two
release cycles to give the field time to roll forward.

## Rollback

If this needs to be reverted:

```bash
git checkout main -- build-with-version.sh .goreleaser.yml scripts/make-release.py .gitignore
git rm assets/update-signing-pubkey.txt scripts/sign-release.py docs/contributing/UPDATE-SIGNING.md
git checkout main -- docs/issues/done/165-authenticode-code-signing-certificate.md
# Private key on disk remains — remove manually if desired:
# rm ~/.config/portunix/update-signing/private.key
```

No runtime Go code was changed, so binaries without an embedded key
(existing in the field) behave identically to pre-issue builds.
