# Release Signing — Key Management Methodology

Operational manual for the Ed25519 private key that signs Portunix release
checksums. Cryptographic and build-time mechanics are covered in
[`docs/contributing/UPDATE-SIGNING.md`](../contributing/UPDATE-SIGNING.md);
the architectural decision is recorded in
[`docs/adr/042-release-signing-key-management.md`](../adr/042-release-signing-key-management.md).
This document covers **the human side**: who holds the key, where it lives,
how it is backed up, how it is transferred.

## 1. Who holds the key

The private signing key is held by the **Project Lead — CassandraGargoyle**.
No active operational copies exist outside the master build machine and
the two offline backups described in §3.

If you are reading this and you are not the project lead, you should not
have a copy of this key. If you do, contact the project lead to coordinate
destruction.

Co-maintainers and contributors **never need access to the private key** to
do their work — they build, test, and ship dev/SNAPSHOT releases just
fine without it. Only signed production releases require the key, and
only the project lead cuts those.

## 2. Master location

The master copy lives in the project tree:

```text
<repo-root>/secrets/update-signing/private.key
```

Properties:

- File permissions: **0600** (owner read/write only)
- Format: PKCS#8 PEM
- Excluded from git via `secrets/` line in `.gitignore`
- The directory `secrets/` is intentionally visible (not hidden), so it
  appears on `ls` and reminds the operator it exists during routine
  project work

The master machine is whichever development machine currently holds this
file. As of ADR-042, that is the project lead's primary build PC.

## 3. Backup strategy

Two **independent offline USB backups** are maintained:

| Slot | Location | Purpose |
| ---- | -------- | ------- |
| Primary backup | Lead's primary residence/office, in safe storage | Recover from master-machine failure |
| Secondary backup | Geographically separate site (e.g. relative's home, safety-deposit box) | Recover from fire/theft of primary site |

### What goes on each USB

```text
USB-portunix-signing-YYYY-MM-DD/
├── private.key                  ← copy of secrets/update-signing/private.key
├── pubkey.txt                   ← copy of assets/update-signing-pubkey.txt
└── README.txt                   ← key fingerprint, date, Portunix version
```

`README.txt` template:

```text
Portunix release signing key — backup
=====================================
Backup date:        2026-05-03
Master machine:     <hostname / serial>
Key holder:         CassandraGargoyle (project-lead)
First-used version: <portunix vX.Y.Z>
Public key (hex):   c70876c02d19ff4a70b4bb3f30a23b6a55c5a60b3bc527c1918c800ddcb7cd3b
Key fingerprint:    <sha256 of private.key>     # see §5 for command

Restore: copy private.key to <repo-root>/secrets/update-signing/private.key
         on the new master machine, chmod 600.
```

### Refresh procedure

Refresh either backup whenever:

- A new key is generated (rotation event)
- The master machine changes
- A backup USB has been physically inspected and shown signs of bit-rot
  (replace flash media every ≈ 3 years regardless)

Refresh command, run on the master machine:

```bash
cp -a secrets/update-signing/ /media/usb/portunix-signing-$(date +%F)/
cp assets/update-signing-pubkey.txt /media/usb/portunix-signing-$(date +%F)/pubkey.txt
# generate fingerprint
sha256sum secrets/update-signing/private.key | tee \
  /media/usb/portunix-signing-$(date +%F)/private.key.sha256
# write README.txt by hand or copy from previous backup, updating dates
```

### What does NOT go on the USB

- The repo source tree (it's on Gitea + GitHub; no need to duplicate).
- Any other secret. The USB is single-purpose so its loss has a single
  blast radius.

## 4. Routine signing flow

This is the day-to-day procedure on the master machine.

```bash
# from the repo root
python3 scripts/make-release.py vX.Y.Z
```

`make-release.py` automatically locates the key via `sign-release.py`,
which looks at:

1. `PORTUNIX_UPDATE_PRIVATE_KEY` env var (override — for testing)
2. `<repo-root>/secrets/update-signing/private.key` (canonical, ADR-042)

If the key is missing, `make-release.py` warns and ships the release
**unsigned**. This is the correct behavior for a fresh machine that does
not have the key yet — but it means: **before cutting a production
release on a new machine, restore the key from a USB backup into
`secrets/update-signing/`.**

## 5. Verifying you have the right key

To confirm a private key file matches the public key embedded in the
binaries:

```bash
# fingerprint of the master key
sha256sum secrets/update-signing/private.key

# derive the public key from the private key file and compare
uv run scripts/sign-release.py verify <some-signed-file>
```

`verify` reports whether the local private key produces a public key
matching the one in `assets/update-signing-pubkey.txt`. Mismatch means
either the master key has been replaced (silent rotation — investigate)
or the public key in the repo is stale (commit the corrected pubkey).

## 6. Transfer to a new master machine

When the project lead replaces or migrates the build PC:

1. **On the new machine**: clone the repo from Gitea. Confirm
   `secrets/` is gitignored and untracked (`git status` shows nothing
   new about `secrets/`).
2. **Restore from USB**: `cp /media/usb/portunix-signing-YYYY-MM-DD/private.key \
   secrets/update-signing/private.key && chmod 600 secrets/update-signing/private.key`
3. **Verify fingerprint**: `sha256sum secrets/update-signing/private.key`
   matches the value in the USB's `README.txt`.
4. **Verify against public key**: `uv run scripts/sign-release.py verify
   --pubkey-file assets/update-signing-pubkey.txt <test-file>`
5. **Cut a SNAPSHOT release** as a smoke test
   (`python3 scripts/make-release.py vX.Y.Z-SNAPSHOT`); confirm a `.sig`
   is produced in `dist/`.
6. **Old master**: securely wipe `secrets/update-signing/private.key`
   from the old machine before disposal/repurpose:
   `shred -uvz -n 3 secrets/update-signing/private.key` (Linux), or
   destroy the disk if reuse is not planned.

## 7. Compromise response

If the private key is suspected compromised (USB lost without recovery,
master machine seized, accidental commit/push of the key, etc.), follow
this sequence:

### 7.1 Immediate

- Stop cutting new releases until the new key is in place.
- File a security advisory in
  [`docs/security/`](.) describing the suspected compromise window
  (start time → discovery time).

### 7.2 Generate replacement key

On the master machine:

```bash
mv secrets/update-signing/private.key \
   secrets/update-signing/private.key.compromised-$(date +%F)
uv run scripts/sign-release.py generate-key --write-pubkey
```

This regenerates `secrets/update-signing/private.key` and overwrites
`assets/update-signing-pubkey.txt`.

### 7.3 Transitional release

Per the rotation pattern in
[`UPDATE-SIGNING.md` §Key Rotation](../contributing/UPDATE-SIGNING.md#key-rotation):

1. Cut release X.Y.Z signed with the **OLD** key (still valid for
   already-installed binaries) but embedding the **NEW** pubkey.
   This requires temporarily moving the new private.key aside, restoring
   the old one (from the `.compromised-DATE` rename) for signing only, and
   running `build-with-version.sh` with `PORTUNIX_UPDATE_PUBKEY=<new-hex>`.
2. Cut release X.Y.Z+1 signed with the **NEW** key normally.
3. Document the rotation prominently in the release changelog and in
   `docs/security/`.
4. Destroy the old key (the `.compromised-DATE` file): `shred -uvz`,
   then remove from USB backups too.

### 7.4 Refresh backups

Two new USB backups (§3) once the new key is in place. Old USBs holding
the compromised key are physically destroyed.

### 7.5 If the key was committed/pushed to a public repo

Treat the key as instantly burned:

- Skip §7.3 step 1's transitional release if the public exposure window
  is wide enough that an attacker could plausibly have already cut a
  signed-by-old-key forged release. In that case go straight to a new
  key with no rotation bridge — affected users will need to re-download.
- Force a release X.Y.Z+1 quickly so the new pubkey reaches the field.
- Communicate the breach in release notes and any out-of-band channels
  (project README, social, etc.).

## 8. Operational checklist (project lead)

Use this when a new project lead takes over, or when running an internal
audit on signing posture.

- [ ] `secrets/update-signing/private.key` exists on master machine
      with mode 0600
- [ ] `secrets/` is in `.gitignore` and `git status` shows it untracked
- [ ] `assets/update-signing-pubkey.txt` matches the private key
      (verified via `sign-release.py verify`)
- [ ] At least two USB backups exist, each with a `README.txt`
      and matching fingerprint
- [ ] USB backups are stored in geographically separate locations
- [ ] A test SNAPSHOT release was cut from the master machine within
      the last 90 days (signing pipeline is exercised)
- [ ] No co-maintainer or CI system holds a copy of the private key

## References

- ADR-042 — Release Signing Key Management & Master Build Machine
- `docs/contributing/UPDATE-SIGNING.md` — crypto & build mechanics
- `scripts/sign-release.py` — generate-key / sign / verify CLI
- `scripts/make-release.py` — release pipeline that drives signing
- Issue #165 — original implementation
- Issue #180 — superseded by #165 + this document + ADR-042
