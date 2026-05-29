# Update Signing

How Portunix verifies that release downloads installed by `portunix update`
come from the project maintainer and have not been tampered with on the
distribution channel (GitHub release assets, mirrors, MITM, etc.).

This document covers issue #165 — self-signed update channel verification.
It is **not** Authenticode code signing for Windows (which requires a paid
CA certificate and is tracked separately).

For the **operational** side (who holds the key, USB backup strategy,
master build machine, compromise response), see
[`docs/security/RELEASE-SIGNING-KEY-MANAGEMENT.md`](../security/RELEASE-SIGNING-KEY-MANAGEMENT.md)
and [ADR-042](../adr/042-release-signing-key-management.md).

## Threat Model

What we defend against:

- **Tampered checksums file** — attacker uploads/replaces `checksums_X.Y.Z.txt`
  to point at a malicious archive's hash.
- **Tampered archive with matching checksum** — attacker replaces both files
  consistently.
- **MITM during download** — attacker rewrites bytes in flight.

What we do **not** defend against:

- **Initial install bootstrap** — the first install gets the public key by
  trusting whatever binary the user downloads. A self-signed cert cannot
  bootstrap trust; that is what CA certificates / Authenticode are for.
  Use TLS + checksum verification + manual reading of `install.sh` for the
  first install.
- **Compromise of the maintainer's signing key** — see [Key Rotation](#key-rotation).

## Architecture

```text
┌──────────────────┐     ┌──────────────────────────┐     ┌──────────────┐
│ assets/update-   │     │ secrets/update-signing/  │     │  GitHub      │
│ signing-         │     │ private.key (mode 0600)  │     │  release     │
│ pubkey.txt       │     │ (gitignored, ADR-042)    │     │  assets      │
│                  │     │                          │     │              │
│  PUBLIC KEY      │     │  PRIVATE KEY             │     │  checksums   │
│  hex, in repo    │     │  PEM, OFF git, on        │     │  + .sig file │
│                  │     │  master build machine    │     │              │
└────────┬─────────┘     └────────────┬─────────────┘     └──────┬───────┘
         │                            │                          │
         │ embedded via               │ used by                  │ verified by
         │ ldflags at                 │ scripts/sign-release.py  │ portunix update
         │ build time                 │ during release build     │ at install time
         ▼                            ▼                          ▼
   binary embeds               .sig file uploaded          ed25519 verifies
   PublicKeyHex                with release                signature against
                                                          checksums file
```

The signature is **detached** and signs the whole `checksums_X.Y.Z.txt` file.
This means a single signature covers all archive variants for that release.
The signed checksums file is then trusted to provide per-archive sha256 sums.

## Scheme

- **Algorithm**: Ed25519 (32-byte public key, 64-byte signature, deterministic)
- **Signature format**: hex-encoded 64 bytes, with optional trailing newline
- **File naming**: signature for `checksums_X.Y.Z.txt` is `checksums_X.Y.Z.txt.sig`
- **Public key transport**: hex string in `assets/update-signing-pubkey.txt`,
  embedded into the binary via `-X portunix.ai/app/update.PublicKeyHex=…`
- **Verification flow** (in `src/app/update/`):
  1. Download the archive
  2. Download `checksums_X.Y.Z.txt`
  3. Download `checksums_X.Y.Z.txt.sig`
  4. Verify Ed25519 signature over checksums file (FAIL → abort)
  5. Find expected sha256 in checksums file
  6. Compute archive sha256, compare (FAIL → abort)
  7. Apply update

## Key Generation

One-time, on the maintainer machine:

```bash
uv run scripts/sign-release.py generate-key --write-pubkey
```

This writes:

- `<repo-root>/secrets/update-signing/private.key` — PKCS#8 PEM, mode 0600
  (canonical location per ADR-042; override with
  `PORTUNIX_UPDATE_PRIVATE_KEY=…` if needed for testing)
- `assets/update-signing-pubkey.txt` — hex-encoded 32-byte public key

Commit `assets/update-signing-pubkey.txt` to the repo. The private key
**must never** be committed — `.gitignore` excludes both the whole
`secrets/` directory and `*.key` / `*.pem` / `private.key` patterns.

## Signing a Release

Automated by `scripts/make-release.py` — after `goreleaser` produces the
`dist/checksums_X.Y.Z.txt` and `copy_quickstart_scripts()` appends quickstart
hashes to it, `sign_checksums()` invokes:

```bash
uv run scripts/sign-release.py sign dist/checksums_X.Y.Z.txt
```

Resulting `dist/checksums_X.Y.Z.txt.sig` is uploaded as a GitHub release asset
alongside the checksums file. `upload-release-to-github.py` already uploads
everything in `dist/`, so no additional changes are needed.

When the maintainer is on a CI runner without the private key, signing is
**skipped with a warning** and the release ships without a `.sig` file.
Older Portunix versions in the field will warn but still install
(`RequireSignature = false` until the grace period elapses).

## Verification — what `portunix update` does

Implemented in `src/app/update/`:

- `signing.go::PublicKeyHex` — set at build time via ldflags
- `signing.go::RequireSignature` — global flag; default `false` (warn-only),
  toggle to `true` once we are confident every release in the field carries
  a `.sig`. Two-release grace period recommended.
- `verify.go::VerifyArchiveChecksumSigned` — verifies signature **before**
  parsing the checksums file, so a tampered file fails fast.

When `PublicKeyHex` is empty (development builds, or releases predating
this issue), verification is skipped silently — backward compatible.

## Build-time Embedding

Two paths embed the pubkey into the binary:

1. **`build-with-version.sh`** (local/dev releases) reads
   `assets/update-signing-pubkey.txt` and adds
   `-X portunix.ai/app/update.PublicKeyHex=<hex>` to `LDFLAGS_COMMON`. Override
   via `PORTUNIX_UPDATE_PUBKEY` env var.

2. **`.goreleaser.yml`** (proper release builds) reads
   `PORTUNIX_UPDATE_PUBKEY` from the environment. `make-release.py` exports
   this from the assets file before invoking goreleaser, so the variable is
   always set.

If the pubkey is missing/empty, `PublicKeyHex` ends up as `""` in the binary
and verification is silently disabled — same effect as a pre-#165 build.

## Key Rotation

Self-signed schemes have no external trust anchor: the trust is "this binary
already runs, and it was installed by a user who trusted us once." Rotating
the key requires those users to either:

- **Re-install from scratch** (download a fresh binary that embeds the new
  pubkey), or
- **Run an in-place update signed by the OLD key** that ships a new binary
  carrying the NEW pubkey. From that point, the user's installed binary
  trusts the new key.

Operationally, key rotation should:

1. Cut a release X.Y.Z signed with the **OLD** key, but whose embedded
   `PublicKeyHex` is the **NEW** key. Users who update to X.Y.Z transition
   trust to the new key seamlessly.
2. Cut release X.Y.Z+1 signed with the **NEW** key. Users who skip X.Y.Z
   cannot update directly to X.Y.Z+1 — they must re-download.
3. Document the rotation prominently in the changelog and release notes.

Treat key rotation as a major maintenance event. Don't rotate without a
disclosed reason (suspected compromise, key-format upgrade, etc.).

## Enforcement Toggle (`RequireSignature`)

Currently `false` in `signing.go` — missing `.sig` triggers a warning, not
an abort. After at least **two consecutive signed releases** are out and the
field has a chance to rotate, flip to `true` so future releases without a
`.sig` are rejected. Document the flip in the release changelog as a
breaking-update note.

## Local Verification

To sanity-check a freshly built `dist/checksums_X.Y.Z.txt.sig` without going
through `portunix update`:

```bash
uv run scripts/sign-release.py verify dist/checksums_X.Y.Z.txt
```

Reports the public key source (repo file or derived from local private key)
and a pass/fail.

## Files Touched

| Path | Purpose |
|------|---------|
| `assets/update-signing-pubkey.txt` | hex pubkey, source of truth |
| `secrets/update-signing/private.key` | PEM private key, canonical location (gitignored, ADR-042) |
| `scripts/sign-release.py` | generate-key / sign / verify |
| `scripts/make-release.py` | exports pubkey, calls sign on checksums |
| `build-with-version.sh` | embeds `PublicKeyHex` via ldflags |
| `.goreleaser.yml` | embeds `PublicKeyHex` for cross-platform builds |
| `src/app/update/signing.go` | runtime verification (preexisting) |
| `src/app/update/verify.go` | `VerifyArchiveChecksumSigned` (preexisting) |
| `.gitignore` | `secrets/` + `*.key` exclusions |
| `docs/adr/042-release-signing-key-management.md` | architectural decision |
| `docs/security/RELEASE-SIGNING-KEY-MANAGEMENT.md` | operational methodology |
