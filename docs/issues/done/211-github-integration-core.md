# Issue #211: GitHub Integration for Portunix Core

> **Renumbered:** formerly internal issue #025. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

## Overview
**Title**: GitHub Integration and Repository Management  
**Status**: ✅ Implemented  
**Closed**: 2026-05-09
**Priority**: High  
**Type**: Enhancement  
**Labels**: github, git, api-integration, core-enhancement

## Implementation Summary

All three phases delivered:

- **Phase 1 (MVP)** — landed 2025-09-07 in commit `2677ebb` as part of #024
  (Plugin Registration System). The `app/github` package provides
  `GitHubClient`, `Service`, `DownloadManager` and is consumed by the plugin
  installer.
- **Phase 2 + 3** — landed under this issue as the **`ptx-github`** helper
  binary (`src/helpers/ptx-github/`), wired into the Git-like dispatcher.

### What ships

CLI: `portunix github clone | checkout | tags | releases | download |
info | status | auth {login,logout,status,list,set-default}`.

Highlights:

- Pure-Go clone, checkout, tags, status via `go-git/v5` (no system git needed).
- AES-256-GCM encrypted token store at `~/.portunix/github/auth.json`
  (mode 0600) with PBKDF2 machine-bound key, optional `--password`
  protection, and multiple named accounts with a default selector.
- Token resolution order: `--account` → `PORTUNIX_GITHUB_TOKEN` /
  `GITHUB_TOKEN` / `GH_TOKEN` → store default. `auth login` validates
  against `/user` before persisting (skip with `--skip-verify`).
- `download` supports `--asset`, `--platform [auto]`, `--all`, `--resume`,
  `--sha256`, and `--overwrite`.
- JSON output (`--json`) on `releases`, `info`, `status`, `tags`,
  `auth status`, `auth list`.
- Unit tests: AES round-trip + tamper detection, store/manager precedence,
  resume + checksum mismatch, repo-spec parsing, asset selectors.

## Problem Description

Portunix currently lacks built-in support for working with GitHub repositories. For the future plugin system (#024) and general developer experience improvement, we need basic GitHub integration in Portunix core.

**Required functionalities:**
1. **GitHub repository operations** - cloning, checkout, download releases
2. **GitHub API integration** - metadata, releases, authentication
3. **Git operations** - basic Git commands integrated into Portunix
4. **Authentication management** - GitHub tokens, SSH keys
5. **Release management** - download binary assets from GitHub releases

## Current State

From analysis of existing codebase:
- Portunix has no dedicated GitHub integration
- No centralized Git functionality
- Missing API wrapper for GitHub operations
- No authentication management for external services

## Requirements

### 1. Basic Git Operations
- [ ] Git repository cloning
- [ ] Branch checkout and switching
- [ ] Tag listing and checkout
- [ ] Repository status checking
- [ ] Basic Git configuration

### 2. GitHub API Integration
- [ ] Repository metadata (description, stars, forks)
- [ ] Release listing and download
- [ ] Asset download with progress reporting
- [ ] Rate limiting handling
- [ ] Error handling and retry logic

### 3. Authentication Management
- [ ] GitHub Personal Access Token support
- [ ] SSH key management
- [ ] Token storage and encryption
- [ ] Authentication validation
- [ ] Multi-account support

### 4. CLI Commands
- [ ] `portunix github clone <repo>`
- [ ] `portunix github releases <repo>`
- [ ] `portunix github download <repo> <version>`
- [ ] `portunix github auth login`
- [ ] `portunix github auth status`

### 5. Core Integration
- [ ] GitHub service as part of Portunix core
- [ ] Reusable GitHub client for other components
- [ ] Configuration management
- [ ] Caching mechanism
- [ ] Progress reporting interface

## Technical Approach

### Architecture

```
Portunix Core
├── GitHub Service
│   ├── API Client
│   ├── Authentication Manager
│   ├── Repository Manager
│   └── Release Manager
├── Git Operations
│   ├── Clone Manager
│   ├── Checkout Operations
│   └── Repository Info
└── CLI Integration
    ├── GitHub Commands
    ├── Progress Reporting
    └── Error Handling
```

### Implementation Components

**1. GitHub API Client:**
```go
type GitHubClient struct {
    client       *github.Client
    auth         AuthManager
    rateLimiter  *RateLimiter
    cache        *Cache
}

func (g *GitHubClient) GetRepository(owner, repo string) (*Repository, error)
func (g *GitHubClient) ListReleases(owner, repo string) ([]*Release, error)
func (g *GitHubClient) DownloadAsset(asset *Asset, dest string) error
```

**2. Git Operations Manager:**
```go
type GitManager struct {
    workDir string
    config  *GitConfig
}

func (g *GitManager) Clone(url, dest string) error
func (g *GitManager) Checkout(repo, ref string) error
func (g *GitManager) GetTags(repo string) ([]string, error)
```

**3. Authentication Manager:**
```go
type AuthManager struct {
    tokenStore TokenStore
    keyManager SSHKeyManager
}

func (a *AuthManager) Login(token string) error
func (a *AuthManager) GetToken() (string, error)
func (a *AuthManager) ValidateAuth() error
```

## Implementation Plan

### Development Strategy
This issue will be implemented together with Issue #024 (Plugin Registration System) using a **monorepo approach**:

```
main
└── feature/plugin-system-with-github
    ├── commits: Issue #025 Phase 1 implementation (GitHub integration)
    ├── commits: Issue #024 basic plugin system using #025 API
    ├── commits: integration testing and refinement
    └── merge → main (both issues delivered together)
```

**Benefits:**
- GitHub integration and plugin system developed together
- Easy integration testing throughout development
- Single comprehensive PR with full functionality
- No dependency coordination issues between branches

### Phase 1: Essential Plugin Support (MVP for Issue #024)
**Focus: Minimal functionality needed for plugin listing and binary download**

1. **GitHub API Client (Basic)**
   ```go
   type GitHubClient struct {
       client *github.Client
   }
   
   // Essential methods for plugin system
   func (g *GitHubClient) GetRepository(owner, repo string) (*Repository, error)
   func (g *GitHubClient) ListReleases(owner, repo string) ([]*Release, error)
   func (g *GitHubClient) DownloadAsset(owner, repo, tag, assetName, dest string) error
   ```

2. **Basic Authentication (Optional Token)**
   - Simple token support (no encryption yet)
   - Environment variable or config file
   - Optional - works without auth but with rate limits

3. **Essential Download Manager**
   ```go
   type DownloadManager struct {
       client *http.Client
   }
   
   func (d *DownloadManager) DownloadFile(url, dest string) error
   func (d *DownloadManager) DownloadWithProgress(url, dest string, progress chan<- int) error
   ```

4. **Core Integration (Minimal)**
   - Basic service registration
   - Simple error handling
   - No CLI commands yet (only programmatic API)

**Deliverables for Plugin System:**
- Plugin installer can list available plugins from GitHub
- Plugin installer can download binary releases
- Basic progress reporting during downloads
- Simple error handling for network issues

### Phase 2: Git Operations & Enhanced Features
1. **Git command wrapper**
   - Clone operations (for source plugins)
   - Checkout functionality
   - Repository management

2. **Enhanced Authentication**
   - Token storage and encryption
   - SSH key management
   - Multi-account support

3. **Advanced Download Features**
   - Resume capability
   - Checksum verification
   - Parallel downloads

### Phase 3: CLI Commands & User Experience
1. **GitHub commands**
   - `portunix github` subcommands
   - Interactive authentication
   - Progress bars and feedback

2. **Help and completion**
   - Command documentation
   - Shell completion
   - Usage examples

## CLI Command Examples

### Repository Operations
```bash
# Clone repository
portunix github clone microsoft/vscode

# List releases
portunix github releases microsoft/vscode
NAME                TAG       PUBLISHED                SIZE
VS Code 1.85.0     1.85.0    2024-01-04T10:30:00Z     150MB
VS Code 1.84.2     1.84.2    2023-12-14T09:15:00Z     148MB

# Download specific release
portunix github download microsoft/vscode 1.85.0 --asset="linux-x64"
```

### Authentication
```bash
# Login with token
portunix github auth login
? GitHub Personal Access Token: [hidden input]
✅ Successfully authenticated as username

# Check auth status
portunix github auth status
✅ Authenticated as: username
🔑 Token expires: 2025-12-31
📊 Rate limit: 4950/5000 remaining
```

### Repository Information
```bash
# Repository details
portunix github info microsoft/vscode
📦 Repository: microsoft/vscode
📄 Description: Visual Studio Code
⭐ Stars: 162,847
🍴 Forks: 28,543
📅 Updated: 2025-09-07T14:30:00Z
🏷️  Latest Release: 1.85.0
```

## Integration with Plugin System

This GitHub integration will serve as foundation for Plugin System (#024):

```go
// Plugin system will leverage GitHub service
type PluginInstaller struct {
    github *GitHubClient  // Reuse GitHub functionality
    git    *GitManager    // Reuse Git operations
}

func (p *PluginInstaller) InstallFromGitHub(repo, version string) error {
    // Use existing GitHub integration
    release, err := p.github.GetRelease(repo, version)
    if err != nil {
        return err
    }
    
    return p.github.DownloadAsset(release.Asset, pluginDir)
}
```

## Security

### Authentication Security
- Token encryption in storage
- Secure token transmission
- Token rotation support
- Permission validation

### Download Security
- Checksum verification
- Signature validation (future)
- Safe download directories
- Path traversal protection

## External Dependencies

### Phase 1 (MVP)
- `github.com/google/go-github/v56/github` - GitHub API client
- `net/http` - Standard HTTP client for downloads (built-in)

### Phase 2 (Enhanced)
- `github.com/go-git/go-git/v5` - Git operations
- `golang.org/x/oauth2` - OAuth2 for GitHub authentication
- `golang.org/x/crypto` - Token encryption

### Phase 3 (CLI)
- `github.com/spf13/cobra` - CLI commands

## Success Criteria

### Phase 1 Requirements (MVP for Plugin System)
- [ ] **Plugin system can list plugin releases programmatically**
- [ ] **Plugin system can download binary assets from GitHub releases**
- [ ] **Basic progress reporting during downloads**
- [ ] **Simple error handling for network failures**
- [ ] **Optional authentication (token from env/config)**
- [ ] **GitHub API client integrated into Portunix core**

### Phase 2 Requirements (Enhanced Features)
- [ ] Git repositories can be cloned programmatically
- [ ] Enhanced authentication with secure token storage
- [ ] Advanced download features (resume, checksum verification)
- [ ] Repository operations for source plugins

### Phase 3 Requirements (CLI & UX)
- [ ] CLI commands: `portunix github clone/releases/download/auth`
- [ ] Interactive authentication workflows
- [ ] Shell completion for all commands
- [ ] User-friendly progress bars and messages

### Performance
- [ ] Repository clone < 30s for typical repo
- [ ] API response time < 2s
- [ ] Asset download with progress reporting
- [ ] Memory usage < 50MB during operations

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| GitHub API rate limits | Medium | Rate limiting, caching, authentication |
| Network failures | Medium | Retry logic, offline mode, error handling |
| Authentication issues | High | Multiple auth methods, clear error messages |
| Large file downloads | Medium | Progress reporting, resume capability |

## Future Enhancements

- GitHub Enterprise support
- GraphQL API integration
- Repository templates
- Issue/PR management
- GitHub Actions integration
- Webhook support

---

**Created**: 2025-09-07  
**Author**: Zdenek  
**Target Release**: v1.6.0
**Related Issues**: Implemented together with #024 Plugin Registration System