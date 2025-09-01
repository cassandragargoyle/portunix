# Issues Documentation & Tracking

This directory contains detailed documentation for all issues, feature requests, and development planning.

## Dual Numbering System

We use a dual numbering system to separate internal development tracking from public GitHub issues:
- **Internal**: All issues (bugs, security, features) tracked in `` with sequential numbering (#001, #002, etc.)
- **Public**: Selected features and enhancements published to GitHub with PUB- prefix (PUB-001, PUB-002, etc.)

## Issues List

| Internal | Public | Title | Status | Priority | Type | Labels |
|----------|--------|-------|--------|----------|------|--------|
| [#001](001-cross-platform-os-detection.md) | PUB-001 | Cross-Platform Intelligent OS Detection System | ✅ Implemented | High | Feature | enhancement, cross-platform, powershell |
| [#002](002-docker-management-command.md) | - | Docker Management Command | 📋 Open | High | Feature | enhancement, docker, cross-platform |
| [#003](003-podman-management-command.md) | - | Podman Management Command | 📋 Open | High | Feature | enhancement, podman, cross-platform |
| [#004](004-mcp-server-ai-integration.md) | PUB-002 | MCP Server for AI Assistant Integration | ✅ Implemented | High | Feature | enhancement, mcp, ai-integration |
| [#007](007-plugin-system-grpc.md) | - | Plugin System with gRPC Architecture | 📋 Open | High | Feature | enhancement, plugin-system, grpc |
| [#008](008-virtual-development-disk.md) | - | Virtual Development Disk Management | 📋 Open | High | Feature | enhancement, virtual-disk, cross-platform |
| [#009](009-configurable-datastore.md) | - | Configurable Datastore System | 📋 Open | High | Feature | enhancement, datastore, enterprise |
| [#010](010-update.md) | PUB-003 | Self-Update Command | ✅ Implemented | High | Feature | enhancement, self-update, cross-platform |
| [#012](012-powershell-linux-installation.md) | - | PowerShell Installation Support for Linux | ✅ Implemented | High | Bug Fix | enhancement, powershell, linux |
| [#013](013-database-management-plugin.md) | - | Database Management Plugin | 📋 Open | High | Plugin | plugin, database, mcp |
| [#014](014-wizard-framework.md) | PUB-004 | Wizard Framework for Interactive CLI | 📋 Open | High | Enhancement | enhancement, cli, wizard, ux |
| [#015](015-vps-edge-bastion-infrastructure.md) | PUB-005 | VPS Edge/Bastion Infrastructure | ✅ Implemented | High | Feature | infrastructure, edge-computing |
| [#016](016-protoc-plugin-development-dependency.md) | - | Protocol Buffers Compiler (protoc) | ✅ Implemented | Critical | Bug Fix | critical, plugin-system, build |
| [#017](017-qemu-kvm-windows-virtualization.md) | - | QEMU/KVM Windows 11 Virtualization with Snapshots | 📋 Open | High | Feature | virtualization, qemu, kvm, windows, snapshot |
| [#019](019-docker-windows-install-issues.md) | - | Docker Installation Issues on Windows | 🔄 In Progress | High | Bug Fix | bug, docker, windows |
| [#020](020-qemu-windows-clipboard-integration.md) | - | QEMU Windows VM Clipboard Integration | 📋 Open | Medium | Enhancement | enhancement, qemu, windows, clipboard, spice |
| [#021](021-github-actions-local-testing.md) | - | GitHub Actions Local Testing Support with Act | 📋 Open | Medium | Feature | feature, github-actions, act, ci-cd, testing |
| [#022](022-google-chrome-installation.md) | - | Google Chrome Installation Implementation | ✅ Implemented | Medium | Feature | enhancement, package-management, cross-platform |

## Directory Structure

```
docs/issues/
├── README.md           # This file - main tracking table
├──            # All internal issues (not published to GitHub)
│   ├── 001-*.md
│   ├── 002-*.md
│   └── ...
└── public/            
    └── mapping.json   # Mapping between internal and public issue numbers
```

## Usage

### Creating New Issues

1. **Internal Issue (all types):**
   - Create file: `{next-number}-{short-title}.md`
   - Update this README with issue entry
   - Set Public column to `-` initially

2. **Publishing to GitHub (features/enhancements only):**
   - Assign next PUB- number in mapping.json
   - Update Public column in this README
   - Create GitHub issue with PUB- number
   - Never publish: bugs, security issues, internal tasks

### Issue Types

- **Feature**: New functionality (can be public)
- **Enhancement**: Improvement to existing features (can be public)  
- **Bug Fix**: Fixing broken functionality (internal only)
- **Security**: Security-related issues (internal only)
- **Plugin**: Plugin-specific features (selective public)

### Status Legend

- 📋 Open - Issue is open and needs work
- 🔄 In Progress - Issue is being actively worked on  
- ✅ Implemented - Issue has been completed and implemented
- ❌ Closed - Issue has been closed without implementation
- ⏸️ On Hold - Issue is temporarily paused

### Priority Legend

- **Critical** - Must be fixed immediately
- **High** - Important feature or significant bug
- **Medium** - Nice to have feature or minor bug
- **Low** - Enhancement or cosmetic issue

## Publishing Guidelines

✅ **Can be published to GitHub:**
- New features
- Enhancements
- Feature requests
- Roadmap items
- Success stories

❌ **Keep internal only:**
- Bug reports and fixes
- Security vulnerabilities
- Performance issues
- Critical errors
- Internal refactoring
- Technical debt