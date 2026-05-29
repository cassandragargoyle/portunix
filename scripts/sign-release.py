#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = ["cryptography>=41.0"]
# ///
"""
Sign Release Script for Portunix - Issue #165

Manages Ed25519 keypair used to sign release checksums files. The detached
signature lets `portunix update` verify that downloaded checksums (and by
extension the archive) come from the holder of the private key, defending
the update channel against tampering.

Subcommands:
    generate-key   Generate an Ed25519 keypair, store private key in
                   <repo-root>/secrets/update-signing/private.key (mode 0600),
                   print the hex-encoded public key for embedding.

    sign FILE      Sign FILE with the private key. Writes FILE.sig containing
                   the hex-encoded 64-byte signature.

    verify FILE    Verify FILE against FILE.sig using either the embedded
                   pubkey from assets/update-signing-pubkey.txt or the
                   public key derived from the local private key.

Private key location (ADR-042):
    1. PORTUNIX_UPDATE_PRIVATE_KEY env var (override)
    2. <repo-root>/secrets/update-signing/private.key (canonical, gitignored)

The public key lives in assets/update-signing-pubkey.txt and is committed
to the repo (public keys carry no secret). The build pipeline reads it and
embeds it into the binary via ldflags.
"""

import argparse
import os
import sys
from pathlib import Path
from typing import Optional

try:
    from cryptography.hazmat.primitives import serialization
    from cryptography.hazmat.primitives.asymmetric.ed25519 import (
        Ed25519PrivateKey,
        Ed25519PublicKey,
    )
except ImportError:
    print("ERROR: cryptography library not available.", file=sys.stderr)
    print("Install with: uv pip install cryptography", file=sys.stderr)
    print("Or run via uv: uv run scripts/sign-release.py ...", file=sys.stderr)
    sys.exit(2)


PRIVATE_KEY_REPO_REL = Path("secrets") / "update-signing" / "private.key"
PUBKEY_FILE_REL = Path("assets") / "update-signing-pubkey.txt"


def repo_root() -> Path:
    return Path(__file__).resolve().parent.parent


def private_key_path() -> Path:
    """Resolve the private key location (ADR-042):
    env override (PORTUNIX_UPDATE_PRIVATE_KEY) or repo-local secrets/.
    """
    override = os.environ.get("PORTUNIX_UPDATE_PRIVATE_KEY")
    if override:
        return Path(override).expanduser()
    return repo_root() / PRIVATE_KEY_REPO_REL


def load_private_key() -> Ed25519PrivateKey:
    path = private_key_path()
    if not path.exists():
        raise FileNotFoundError(
            f"private key not found at {path}\n"
            f"Generate one with: scripts/sign-release.py generate-key"
        )

    raw = path.read_bytes()
    return serialization.load_pem_private_key(raw, password=None)


def cmd_generate_key(args: argparse.Namespace) -> int:
    path = private_key_path()
    if path.exists() and not args.force:
        print(f"ERROR: private key already exists at {path}", file=sys.stderr)
        print("Use --force to overwrite (will invalidate all prior signatures).", file=sys.stderr)
        return 1

    path.parent.mkdir(parents=True, exist_ok=True, mode=0o700)

    priv = Ed25519PrivateKey.generate()
    pem = priv.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    )
    path.write_bytes(pem)
    os.chmod(path, 0o600)

    pub_hex = priv.public_key().public_bytes(
        encoding=serialization.Encoding.Raw,
        format=serialization.PublicFormat.Raw,
    ).hex()

    print(f"✓ Private key written to {path} (mode 0600)")
    print()
    print(f"Public key (hex, 64 chars):")
    print(f"  {pub_hex}")
    print()

    pubkey_file = repo_root() / PUBKEY_FILE_REL
    if args.write_pubkey:
        pubkey_file.parent.mkdir(parents=True, exist_ok=True)
        pubkey_file.write_text(pub_hex + "\n")
        print(f"✓ Public key written to {pubkey_file}")
    else:
        print(f"To embed in builds, write it to {pubkey_file}:")
        print(f"  echo {pub_hex} > {pubkey_file}")

    return 0


def cmd_sign(args: argparse.Namespace) -> int:
    target = Path(args.file)
    if not target.is_file():
        print(f"ERROR: file not found: {target}", file=sys.stderr)
        return 1

    priv = load_private_key()
    data = target.read_bytes()
    sig = priv.sign(data)

    sig_path = Path(str(target) + ".sig")
    sig_path.write_text(sig.hex() + "\n")
    print(f"✓ Signed {target.name}")
    print(f"  Signature: {sig_path}")
    print(f"  Length:    {len(sig)} bytes ({len(sig.hex())} hex chars)")

    return 0


def load_public_key_from_file(path: Path) -> Optional[Ed25519PublicKey]:
    if not path.is_file():
        return None
    raw_hex = path.read_text().strip()
    if not raw_hex:
        return None
    raw = bytes.fromhex(raw_hex)
    if len(raw) != 32:
        raise ValueError(f"public key must be 32 bytes, got {len(raw)}")
    return Ed25519PublicKey.from_public_bytes(raw)


def cmd_verify(args: argparse.Namespace) -> int:
    target = Path(args.file)
    sig_path = Path(args.signature) if args.signature else Path(str(target) + ".sig")

    if not target.is_file():
        print(f"ERROR: file not found: {target}", file=sys.stderr)
        return 1
    if not sig_path.is_file():
        print(f"ERROR: signature not found: {sig_path}", file=sys.stderr)
        return 1

    pub: Optional[Ed25519PublicKey] = None
    source = ""

    if args.pubkey_file:
        pub = load_public_key_from_file(Path(args.pubkey_file))
        source = str(args.pubkey_file)
    else:
        repo_pubkey = repo_root() / PUBKEY_FILE_REL
        pub = load_public_key_from_file(repo_pubkey)
        if pub is not None:
            source = str(repo_pubkey)

    if pub is None:
        try:
            priv = load_private_key()
            pub = priv.public_key()
            source = f"derived from {private_key_path()}"
        except FileNotFoundError as e:
            print(f"ERROR: no public key available.", file=sys.stderr)
            print(f"  {e}", file=sys.stderr)
            return 1

    sig_text = sig_path.read_text().strip()
    sig = bytes.fromhex(sig_text)
    if len(sig) != 64:
        print(f"ERROR: signature must be 64 bytes, got {len(sig)}", file=sys.stderr)
        return 1

    data = target.read_bytes()
    try:
        pub.verify(sig, data)
    except Exception as e:
        print(f"✗ Signature INVALID for {target.name}", file=sys.stderr)
        print(f"  Public key source: {source}", file=sys.stderr)
        print(f"  Reason: {e}", file=sys.stderr)
        return 1

    print(f"✓ Signature valid for {target.name}")
    print(f"  Public key source: {source}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(
        prog="sign-release",
        description="Manage Ed25519 signing for Portunix release checksums (issue #165)",
    )
    sub = parser.add_subparsers(dest="cmd", required=True)

    gen = sub.add_parser("generate-key", help="Generate Ed25519 keypair")
    gen.add_argument("--force", action="store_true",
                     help="overwrite existing private key (DESTRUCTIVE)")
    gen.add_argument("--write-pubkey", action="store_true",
                     help=f"also write public key to {PUBKEY_FILE_REL}")
    gen.set_defaults(func=cmd_generate_key)

    sgn = sub.add_parser("sign", help="Sign a file with the private key")
    sgn.add_argument("file", help="path to file to sign (e.g., dist/checksums_1.10.9.txt)")
    sgn.set_defaults(func=cmd_sign)

    vfy = sub.add_parser("verify", help="Verify a file against its .sig")
    vfy.add_argument("file", help="path to signed file")
    vfy.add_argument("--signature", help="explicit signature path (default: FILE.sig)")
    vfy.add_argument("--pubkey-file", help="explicit pubkey hex file")
    vfy.set_defaults(func=cmd_verify)

    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
