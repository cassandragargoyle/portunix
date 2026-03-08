---
title: "credential"
description: "Secure credential storage"
---

# credential

Secure credential storage

## Usage

```bash
portunix credential [options] [arguments]
```

## Full Help

```
PTX-Credential - Secure Credential Management

Securely store and retrieve credentials (API keys, passwords, tokens) with
AES-256-GCM encryption and PBKDF2 key derivation.

Features:
  - AES-256-GCM encryption with PBKDF2-HMAC-SHA256 key derivation
  - Machine-bound encryption (no password needed for default usage)
  - Optional password protection for additional security
  - M365 token compatibility with Java TokenStorage
  - Multiple named credential stores

Storage location: ~/.portunix/credentials/

Usage:
  ptx-credential [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  delete      Delete a credential
  get         Retrieve a credential
  help        Help about any command
  info        Show system information used for key derivation
  list        List all credentials
  m365        M365 token compatibility mode
  set         Store a credential
  store       Manage credential stores

Flags:
  -h, --help           help for ptx-credential
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")
  -v, --version        version for ptx-credential

Use "ptx-credential [command] --help" for more information about a command.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [completion](#completion) | Generate the autocompletion script for the specified shell |
| [delete](#delete) | Delete a credential |
| [get](#get) | Retrieve a credential |
| [help](#help) | Help about any command |
| [info](#info) | Show system information used for key derivation |
| [list](#list) | List all credentials |
| [m365](#m365) | M365 token compatibility mode |
| [set](#set) | Store a credential |
| [store](#store) | Manage credential stores |

### completion

Generate the autocompletion script for the specified shell

```
Generate the autocompletion script for ptx-credential for the specified shell.
See each sub-command's help for details on how to use the generated script.

Usage:
  ptx-credential completion [command]

Available Commands:
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh

Flags:
  -h, --help   help for completion

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")

Use "ptx-credential completion [command] --help" for more information about a command.
```

### delete

Delete a credential

```
Delete a credential from the store.

Examples:
  portunix credential delete github-token
  portunix credential delete api-key --store secure

Usage:
  ptx-credential delete <name> [flags]

Flags:
  -h, --help   help for delete

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")
```

### get

Retrieve a credential

```
Retrieve a credential value.

Examples:
  portunix credential get github-token
  portunix credential get api-key --quiet
  portunix credential get company-secret --store secure --password

Usage:
  ptx-credential get <name> [flags]

Flags:
  -h, --help   help for get

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")
```

### help

Help about any command

```
Help provides help for any command in the application.
Simply type ptx-credential help [path to command] for full details.

Usage:
  ptx-credential help [command] [flags]

Flags:
  -h, --help   help for help

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")
```

### info

Show system information used for key derivation

```
Show system information used for cryptographic key derivation.

This is useful for debugging compatibility issues with Java TokenStorage.

Usage:
  ptx-credential info [flags]

Flags:
  -h, --help   help for info

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")
```

### list

List all credentials

```
List all credentials in a store (names and labels only, never values).

Examples:
  portunix credential list
  portunix credential list --store secure
  portunix credential list --json

Usage:
  ptx-credential list [flags]

Flags:
  -h, --help   help for list

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")
```

### m365

M365 token compatibility mode

```
M365 token compatibility mode for Java TokenStorage compatibility.

These commands allow reading and writing M365 tokens in a format compatible
with the Java TokenStorage implementation used by m365-extractor plugin.

Usage:
  ptx-credential m365 [command]

Available Commands:
  delete      Delete M365 tokens
  get         Get M365 tokens
  set         Set M365 tokens

Flags:
  -h, --help   help for m365

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")

Use "ptx-credential m365 [command] --help" for more information about a command.
```

### set

Store a credential

```
Store a credential securely.

Examples:
  portunix credential set github-token "ghp_xxxxxxxxxxxx"
  portunix credential set api-key "secret123" --label "Production API Key"
  portunix credential set company-secret "xxx" --store secure --password

Usage:
  ptx-credential set <name> <value> [flags]

Flags:
  -h, --help           help for set
      --label string   Human-readable label for the credential

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")
```

### store

Manage credential stores

```
Manage credential stores (create, list, delete).

Usage:
  ptx-credential store [command]

Available Commands:
  create      Create a new credential store
  delete      Delete a credential store
  list        List all credential stores

Flags:
  -h, --help   help for store

Global Flags:
      --json           Output in JSON format
      --password       Use password-protected store
      --quiet          Suppress non-essential output
      --store string   Credential store name (default: "default")

Use "ptx-credential store [command] --help" for more information about a command.
```

