# Issue #180: Ed25519 Release Signing Pipeline (Phase 2 of #166)

**Type:** Enhancement
**Priority:** Medium
**Status:** ❌ Closed (Superseded)
**Labels:** enhancement, security, update, code-signing, supply-chain, release-pipeline
**Created:** 2026-04-22
**Closed:** 2026-05-03
**Related:** #166 (Phase 1 — Go verification code, ✅ Implemented)
**Superseded by:** #165 (which delivered the production keypair, signing
pipeline, and `.sig` upload before this issue was scheduled) plus
[ADR-042](../../adr/042-release-signing-key-management.md) and
[`docs/security/RELEASE-SIGNING-KEY-MANAGEMENT.md`](../../security/RELEASE-SIGNING-KEY-MANAGEMENT.md)
(operational key-management methodology that #180 had originally listed
under "release-signing.md")

## Closure Note (2026-05-03)

Every implementation goal of this issue was actually shipped under #165
(merged 2026-05-03). The remaining gap — operational methodology around
key holder, master build machine, USB backup, and compromise response —
is now tracked by ADR-042 and `docs/security/RELEASE-SIGNING-KEY-MANAGEMENT.md`.

What was originally planned vs. where it landed:

| #180 acceptance criterion | Where it shipped |
|---------------------------|------------------|
| Production Ed25519 keypair generated, private key offline | #165 — key in `secrets/update-signing/private.key` (ADR-042) |
| `PublicKeyHex` populated | #165 — `assets/update-signing-pubkey.txt` + ldflags embed |
| `make-release.py` produces `.sig` | #165 — `sign_checksums()` in pipeline |
| `.goreleaser.yml` uploads `.sig` | #165 — via `dist/` upload (existing path covers it) |
| `release-signing.md` documenting key lifecycle | `docs/security/RELEASE-SIGNING-KEY-MANAGEMENT.md` (ADR-042) |
| First signed release built end-to-end | Will land with the next normal release after #165 / ADR-042 — no separate tracking needed; failure would surface immediately as a hot issue |
| `RequireSignature` stays `false` (deferred enforcement) | Preserved by #165; flip is tracked as future work in `UPDATE-SIGNING.md` |

The original scope below remains as historical context for future
maintainers reading the archive.

---

## Description

Phase 2 of the Ed25519 update verification work tracked in #166. Phase 1
landed the in-binary verification logic but ships a **placeholder empty
public key**, so signature verification is currently a no-op at runtime.

This issue covers the **production rollout**: generate a real Ed25519
keypair, embed the public key into Portunix, sign release artifacts, and
publish the signature alongside the existing checksums file.

## Motivation

- Without Phase 2, the verification code from #166 is dead weight — it
  cannot detect a compromised GitHub release because no signature ever
  reaches `VerifyChecksumsSignature` and `HasPublicKey()` returns `false`.
- Closing the loop activates the supply-chain protection promised by
  #166 (defense against tampering with both binary and checksum file on
  GitHub Releases).
- Establishes the operational discipline (offline private key, key
  rotation procedure) before scope expands further.

## Goals

1. Generate a production Ed25519 keypair (private key kept offline,
   never in repo).
2. Embed the corresponding public key into the Portunix binary by
   populating `PublicKeyHex` in `src/app/update/signing.go`.
3. Sign the `checksums_<version>.txt` artifact during release as part of
   `scripts/make-release.py`.
4. Publish the resulting `checksums_<version>.txt.sig` as a release asset
   via `.goreleaser.yml`.
5. Validate end-to-end: a fresh release is downloaded by `portunix update`
   and the signature passes; a tampered checksum file is rejected.
6. Document the key management lifecycle (generation, storage, rotation,
   compromise response).

Out of scope (deferred to Phase 3 / #166 Phase 4):

- Flipping `RequireSignature = true` (enforce mode). Stays `false` for
  ≥1 signed release cycle to avoid bricking updates if pipeline misfires.
- Multi-key support / key rotation tooling (Phase 4 of #166).

## Design Options

### Option A — Sign at release time inside `make-release.py` (recommended)

Generate signature locally during `make-release.py vX.Y.Z`, using a
private key file passed via env var or `--signing-key` flag. Upload `.sig`
alongside `checksums_*.txt`.

- **Pros:** One script owns the entire release flow; signing key never
  needs to be exposed to CI; trivial to skip for SNAPSHOT/dry-run builds.
- **Cons:** Release author must have the private key locally — a
  deliberate choice that keeps the secret off shared infrastructure.

### Option B — Sign in CI/CD (e.g., GitHub Actions secret)

Store private key as a CI secret, sign during workflow.

- **Pros:** Anyone authorized to trigger a release can produce a signed
  build; release process is fully automated.
- **Cons:** Private key lives on a third party's infrastructure;
  expanding the trust boundary contradicts the "offline trust anchor"
  premise of #166. Compromise of CI = compromise of signing key.

### Option C — Hybrid: build in CI, sign locally

CI produces unsigned artifacts; release author downloads, signs locally,
re-uploads `.sig` separately.

- **Pros:** Keeps key offline like Option A.
- **Cons:** Extra manual step, error-prone; window between unsigned
  upload and signed upload is exploitable.

**Recommendation: Option A.** Aligns with the issue's threat model (keep
the key under the release author's control), minimal pipeline changes,
no new infrastructure.

## Technical Notes

### Key generation (one-time, off the project machine if possible)

```bash
# 32-byte Ed25519 private key + matching public key
openssl genpkey -algorithm Ed25519 -out portunix-release-signing.key
openssl pkey -in portunix-release-signing.key -pubout -outform DER \
  | tail -c 32 | xxd -p -c 64    # 64 hex chars → PublicKeyHex value
```

Store private key encrypted (e.g. age, gpg, or a password manager
attachment); never commit it; back it up to two independent locations.

### Signing flow (in `make-release.py`)

```python
from nacl.signing import SigningKey  # or call `openssl pkeyutl -sign`
key = SigningKey(open(args.signing_key, "rb").read())
data = open(f"dist/checksums_{version}.txt", "rb").read()
sig = key.sign(data).signature  # 64 bytes
open(f"dist/checksums_{version}.txt.sig", "wb").write(sig)
```

(If avoiding a Python crypto dependency, shell out to `openssl pkeyutl
-sign -inkey ... -rawin -in checksums.txt -out checksums.txt.sig`.)

### Sequence — release pipeline with signing

```
release-author    make-release.py    .goreleaser.yml    GitHub Releases
      |                  |                   |                |
      |--vX.Y.Z--------->|                   |                |
      |                  |--build artifacts->|                |
      |                  |<------dist/-------|                |
      |                  | sign checksums    |                |
      |                  |   (private key)   |                |
      |                  |--upload assets-------------------->|
      |                  |   incl. *.sig                      |
```

### Sequence — update with signature verification (already in place)

```
portunix-client                          GitHub Releases
      |                                       |
      |--GET /releases/latest---------------->|
      |<------release JSON (incl. .sig URL)---|
      |--GET checksums_X.Y.Z.txt------------->|
      |--GET checksums_X.Y.Z.txt.sig--------->|
      | VerifyChecksumsSignature(             |
      |   checksumsData, sigBytes,            |
      |   embedded PublicKey)                 |
      | -> ok → proceed; fail → abort         |
```

## Implementation Steps

1. **Key generation & embedding**
   - Generate Ed25519 keypair (off-repo machine ideally)
   - Securely store private key, distribute backup to release authors
   - Update `PublicKeyHex` in `src/app/update/signing.go` with the
     64-hex-char public key
   - Verify with a one-shot Go test signing/verifying with the real key

2. **Release script**
   - Extend `scripts/make-release.py`:
     - Add `--signing-key PATH` argument (or env var
       `PORTUNIX_SIGNING_KEY`)
     - After `dist/checksums_*.txt` is produced, sign it and write
       `dist/checksums_*.txt.sig`
     - Skip silently for SNAPSHOT versions if no key provided (keep
       existing dev-friendly behavior)
   - Update upload logic in `scripts/upload-release-to-github.py` and
     `upload-release-to-gitea.py` to include `.sig`

3. **GoReleaser**
   - Update `.goreleaser.yml` `release.extra_files` (or equivalent
     section) to include `checksums_*.txt.sig`

4. **Documentation**
   - New `docs/security/release-signing.md` covering: key generation,
     storage, distribution to backup authors, rotation cadence,
     compromise response procedure
   - Update `docs/commands/core/update.md` to mention signature
     verification
   - Cross-link from #166 archive entry

5. **End-to-end validation** (in container per testing methodology)
   - Build a SNAPSHOT release with the new pipeline, verify `.sig` is
     present in `dist/`
   - Run `portunix update --check` against a fixture release; observe
     "Signature verified" output
   - Negative test: tamper `checksums_*.txt` after signing, confirm
     update aborts with signature error

## Acceptance Criteria

- [ ] Production Ed25519 keypair generated; private key stored offline
- [ ] `PublicKeyHex` in `signing.go` populated with the production key
- [ ] `scripts/make-release.py` produces `checksums_*.txt.sig` when
      signing key is provided
- [ ] `.goreleaser.yml` uploads `.sig` as a release asset
- [ ] First signed release built and published end-to-end
- [ ] `portunix update` against the signed release prints
      "Signature verified"
- [ ] Tampered checksum file is rejected with clear error
- [ ] `RequireSignature` remains `false` (enforcement deferred per scope
      decision above)
- [ ] `docs/security/release-signing.md` exists and covers key lifecycle

## References

- Parent: #166 — Self-Signed Certificate for Update Verification
  (Phase 1 ✅ Implemented 2026-04-22)
- Phase 1 acceptance protocol:
  `docs/testing/internal/acceptance-166.md`
- Related: #165 (Authenticode Code Signing Certificate for Windows) —
  complementary, OS-level signing for the installer itself
- Go stdlib: `crypto/ed25519`
- Existing release pipeline: `scripts/make-release.py`,
  `scripts/upload-release-to-github.py`, `.goreleaser.yml`
