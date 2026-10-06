# Issue #136: PTX-Credential Helper Implementation

**Status**: ✅ Implemented
**Priority**: High
**Type**: Feature / Architecture
**Created**: 2026-01-17
**Labels**: enhancement, helper-binary, security, credential-storage, encryption, cross-platform

---

## Summary

Implement a dedicated `ptx-credential` helper binary for secure storage and retrieval of credentials (API keys, passwords, tokens).

**The implementation is based on the proven Java implementation `TokenStorage.java` from the `m365-extractor` plugin**, which already handles OAuth token encryption. The new Go helper will:

1. **Adopt the cryptographic design** from m365-extractor (AES-256-GCM, PBKDF2, seed format)
2. **Ensure full compatibility** - Go can read Java-encrypted files and vice versa
3. **Provide unified CLI interface** for all Portunix tools and plugins

After implementation, the **m365-extractor plugin** has two options:

- **Option A (Recommended)**: Migrate to use `portunix credential` commands instead of internal TokenStorage
- **Option B**: Keep internal TokenStorage.java but ensure bidirectional compatibility with ptx-credential

## Problem Statement

### Current State

1. **No unified credential storage** - Portunix lacks central functionality for secure credential storage
2. **Isolated plugin implementations** - m365-extractor has its own TokenStorage.java (508 lines, proven solution)
3. **Duplication risk** - Other plugins/tools would need to reimplement the same cryptographic logic
4. **Security risk** - Without standard solution, credentials may end up in:
   - Plain text configuration files
   - Environment variables (visible in process list)
   - Source code (accidentally committed)

### Reference Implementation

The **m365-extractor plugin** (`portunix-plugins`) already contains a production-ready Java implementation:

```text
portunix-plugins/plugins/m365-extractor/src/main/java/
└── org/cassandragargoyle/portunix/m365/auth/
    └── TokenStorage.java  (508 lines - encryption, key derivation, file storage)
```

This implementation is the **source of truth** for cryptographic parameters and will serve as the reference for the Go implementation.

### Triggering Use Cases

1. **pft-harvester** (external project) - needs API keys for Hlídač státu and other services
2. **m365-extractor plugin** - already has working Java implementation, candidate for migration
3. **GitHub integration** - requires Personal Access Tokens
4. **AI integrations** - API keys for various AI services
5. **Future plugins** - unified credential access without reimplementation

### Requirements from External Projects

- External project request: CLI and Python API for credential management
- **m365-extractor**: Source of cryptographic specification + future migration candidate

## Goals

1. **Secure storage** - credentials encrypted on disk using industry-standard algorithms
2. **Cross-platform** - works identically on Windows, Linux, macOS
3. **Full compatibility with m365-extractor** - bidirectional read/write of encrypted files
4. **Easy to use** - simple CLI for storing and retrieving credentials
5. **Integration ready** - usable from other tools (pft-harvester, scripts, plugins)
6. **Headless support** - works on servers without GUI or keyring
7. **Migration path** - enable m365-extractor to transition to unified solution

## Technical Architecture

### Helper Binary Pattern

Following ADR-014 (Git-like Dispatcher) and established helper pattern:

```text
User Command: portunix credential set github-token "ghp_xxx"
                    ↓
         ┌──────────────────┐
         │ Main Dispatcher  │  (lightweight, fast startup)
         │   (portunix)     │
         └────────┬─────────┘
                  │ delegates to
                  ↓
         ┌──────────────────┐
         │  PTX-Credential  │  (credential management subsystem)
         │   Helper Binary  │
         └──────────────────┘
                  │
          ┌───────┴────────┐
          ↓                ↓
    [Key Derivation]  [Encrypted Storage]
    [PBKDF2/AES-GCM]  [~/.portunix/credentials/]
```

### Storage Architecture

```text
~/.portunix/
├── credentials/
│   ├── default.enc           # Default credential store (machine-bound key)
│   ├── secure.enc            # Password-protected store (optional)
│   └── .password-protected   # Marker file for password-required stores
└── .portunix-m365-tokens.enc # Legacy m365 compatibility (existing location)
```

### Credential Data Structure

```json
{
  "version": 1,
  "credentials": [
    {
      "name": "github-token",
      "label": "GitHub Personal Access Token",
      "value": "ghp_xxxxxxxxxxxx",
      "created": "2026-01-17T10:30:00Z",
      "updated": "2026-01-17T10:30:00Z",
      "metadata": {
        "service": "github.com",
        "scope": "repo,workflow"
      }
    }
  ]
}
```

## Cryptographic Specification

### CRITICAL: Must Match Java Implementation

The implementation MUST be byte-for-byte compatible with `TokenStorage.java` from m365-extractor.

| Parameter | Value | Notes |
| --------- | ----- | ----- |
| Algorithm | `AES/GCM/NoPadding` | Authenticated encryption |
| Key length | 256 bits (32 bytes) | AES-256 |
| IV length | 12 bytes | GCM standard |
| GCM tag length | 128 bits | Authentication tag |
| Key derivation | `PBKDF2-HMAC-SHA256` | Industry standard |
| PBKDF2 iterations | 65536 | Balance security/performance |

### Key Derivation

Machine-bound key derivation (no password):

```text
seed = "{hostname}|{username}|{os_name}|{home_dir}|portunix-credential"
salt = seed.bytes()
key = PBKDF2(seed, salt, 65536, 32, SHA256)
```

Password-protected key derivation:

```text
seed = "{hostname}|{username}|{os_name}|{home_dir}|portunix-credential|pw:{password}"
salt = seed.bytes()
key = PBKDF2(seed, salt, 65536, 32, SHA256)
```

### OS Name Compatibility

**WARNING**: Go's `runtime.GOOS` returns "windows" but Java returns "Windows 10".

Must detect full OS name to match Java:

- Windows: "Windows 10", "Windows 11"
- Linux: "Linux"
- macOS: "Mac OS X"

> **Implementation Note**: Portunix already has OS detection functionality in `portunix system info` command.
> The existing code from `app/system/` package should be reused/shared to ensure consistency
> and avoid code duplication. See `app/system/info.go` for current implementation.

### Encrypted Format

```text
Base64( IV[12 bytes] || Ciphertext || GCM_AuthTag[16 bytes] )
```

## CLI Interface

### Commands

```bash
# Store credential
portunix credential set <name> <value> [flags]
  --label "Description"      # Human-readable label
  --store <name>             # Use specific store (default: "default")
  --password                 # Use password-protected store

# Retrieve credential (prints value to stdout)
portunix credential get <name> [flags]
  --store <name>             # Use specific store
  --quiet                    # No newline at end (for scripting)

# List credentials (names and labels only, never values)
portunix credential list [flags]
  --store <name>             # Use specific store
  --json                     # Output as JSON

# Delete credential
portunix credential delete <name> [flags]
  --store <name>             # Use specific store

# Store management
portunix credential store create <name> [--password]
portunix credential store list
portunix credential store delete <name>

# M365 compatibility mode
portunix credential m365 get           # Get m365 tokens (legacy format)
portunix credential m365 set <json>    # Set m365 tokens (legacy format)
portunix credential m365 delete        # Delete m365 tokens
```

### Usage Examples

```bash
# Basic usage
portunix credential set github-token "ghp_xxxxxxxxxxxx" --label "GitHub PAT"
portunix credential get github-token

# List all credentials
portunix credential list
# Output:
# NAME             LABEL                          UPDATED
# github-token     GitHub PAT                     2026-01-17
# hlidac-api       Hlídač státu API key           2026-01-15

# Password-protected store
portunix credential set company-secret "xxx" --store secure --password
# Prompts for password

# Use in scripts
API_KEY=$(portunix credential get hlidac-api --quiet)
curl -H "Authorization: Bearer $API_KEY" https://api.example.com

# M365 compatibility
portunix credential m365 get
# Returns JSON with accessToken, refreshToken, etc.
```

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `PORTUNIX_CREDENTIAL_PASSWORD` | Password for encrypted stores (avoids interactive prompt) |
| `PORTUNIX_CREDENTIAL_STORE` | Default store name |

## Security Requirements

1. **No plain-text storage** - credentials always encrypted at rest
2. **Secure memory handling** - clear sensitive data from memory after use
3. **File permissions** - credential files readable only by owner (600)
4. **No logging of values** - never log credential values, only names
5. **Audit trail** (optional) - log credential access events (without values)

## Implementation Phases

### Phase 1: Core Infrastructure

- [ ] Create `helpers/ptx-credential/` directory structure
- [ ] Implement key derivation (PBKDF2-HMAC-SHA256)
- [ ] Implement AES-256-GCM encryption/decryption
- [ ] Implement credential store file format
- [ ] Add `set`, `get`, `delete`, `list` commands
- [ ] Basic tests with test vectors

### Phase 2: M365 Compatibility

- [ ] Implement legacy m365 token format support
- [ ] Cross-validation tests with Java TokenStorage
- [ ] OS name detection for exact seed matching
- [ ] Password-protected mode

### Phase 3: Main Binary Integration

- [ ] Add `credential` command to main dispatcher
- [ ] Helper binary discovery and execution
- [ ] Help text integration
- [ ] Shell completion for credential names

### Phase 4: Advanced Features

- [ ] Multiple named stores
- [ ] Credential metadata support
- [ ] Export/import functionality
- [ ] Audit logging (optional)

### Phase 5: Documentation & Testing

- [ ] User documentation
- [ ] Integration tests
- [ ] Security review
- [ ] Cross-platform validation (Windows, Linux, macOS)

## M365-Extractor Migration Strategy

> **Note**: Migration decision is responsibility of m365-extractor team. This section documents the options available after ptx-credential is implemented.

After ptx-credential is implemented and validated, the **m365-extractor plugin** should transition:

### Current State (Before)

```text
m365-extractor (Java)
    │
    └── TokenStorage.java (508 lines)
            │
            └── ~/.portunix/.portunix-m365-tokens.enc
```

### Target State (After Migration)

```text
m365-extractor (Java)
    │
    └── calls: portunix credential m365 get/set
                    │
                    └── ptx-credential (Go helper)
                            │
                            └── ~/.portunix/.portunix-m365-tokens.enc
```

### Migration Options for m365-extractor

#### Option A: Full Migration (Recommended)

1. Remove internal TokenStorage.java
2. Call `portunix credential m365 get` via ProcessBuilder
3. Benefits:
   - Single source of truth for credential management
   - Automatic updates when ptx-credential improves
   - Reduced code maintenance in plugin

```java
// Before (internal implementation)
TokenStorage storage = new TokenStorage();
String token = storage.getAccessToken();

// After (delegated to portunix)
ProcessBuilder pb = new ProcessBuilder("portunix", "credential", "m365", "get");
String tokenJson = readOutput(pb.start());
```

#### Option B: Compatibility Mode

1. Keep TokenStorage.java as-is
2. Ensure it remains byte-compatible with ptx-credential
3. Use when:
   - Plugin needs to work without portunix installed
   - Transition period while validating ptx-credential

### Compatibility Contract

Both implementations MUST:

- Use identical seed format: `{hostname}|{username}|{os_name}|{home_dir}|portunix-m365[|pw:{password}]`
- Use identical PBKDF2 parameters (65536 iterations, SHA-256)
- Use identical file path: `~/.portunix/.portunix-m365-tokens.enc`
- Produce byte-identical ciphertext for same plaintext + IV

### Validation Test

Before m365-extractor migration:

1. Encrypt token with Java TokenStorage
2. Decrypt with `portunix credential m365 get`
3. Encrypt with `portunix credential m365 set`
4. Decrypt with Java TokenStorage
5. All operations must succeed with identical data

## Scope

### In Scope

- Encrypted credential storage (API keys, passwords, tokens)
- CLI commands for credential management
- Compatibility with Java m365-extractor TokenStorage
- Cross-platform support (Windows, Linux, macOS)
- Password-protected stores (optional)
- Multiple named stores

### Out of Scope

- OAuth flow implementation (handled by consumers like pft-harvester)
- Integration with external secret managers (Vault, AWS Secrets Manager)
- GUI for credential management
- Python API (may be added later as separate issue)
- Automatic credential rotation

## Acceptance Criteria

1. [ ] `portunix credential set/get/delete/list` commands work correctly
2. [ ] Credentials encrypted with AES-256-GCM
3. [ ] Go implementation can decrypt Java-encrypted m365 tokens
4. [ ] Java implementation can decrypt Go-encrypted credentials
5. [ ] Works on Windows, Linux, macOS
6. [ ] Credential files have correct permissions (600)
7. [ ] No credential values in logs
8. [ ] Password-protected stores work correctly
9. [ ] Help text and shell completion available

## Related Issues & Documents

- External project request: Add Secure Credential Storage to portunix
- **secure-token-storage-spec.md**: Cryptographic specification from portunix-plugins
- **m365-extractor TokenStorage.java**: Reference Java implementation
- **ADR-014**: Git-like Dispatcher Pattern (foundation for helpers)
- **Issue #051**: Dispatcher Architecture Implementation

## Test Vectors

For cross-implementation validation:

```text
# Test vector 1: Simple credential
Seed: "TESTHOST|testuser|Windows 10|C:\Users\testuser|portunix-credential"
Plaintext: "test-api-key-12345"
# Expected ciphertext will be generated during implementation
```

## References

- [Go crypto/aes package](https://pkg.go.dev/crypto/aes)
- [Go crypto/cipher GCM](https://pkg.go.dev/crypto/cipher#NewGCM)
- [Go x/crypto/pbkdf2](https://pkg.go.dev/golang.org/x/crypto/pbkdf2)
- [OWASP Credential Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [NIST SP 800-132: Password-Based Key Derivation](https://csrc.nist.gov/publications/detail/sp/800-132/final)

---

**Last Updated**: 2026-01-17
