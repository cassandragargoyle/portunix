# Issue #35: AI Assistant Installation Support

## Summary
Extend Portunix package management to support installation of AI assistants (Claude Code, Claude Desktop, Gemini CLI, etc.) using the standard `portunix install` command, similar to how other software packages are currently handled.

## Implementation Progress

**Status: 🔄 Partially implemented (bundles + detection + macOS install + recommend-ai/MCP-hook/version merged)**

- ✅ **AI assistant bundles** (merged to `main`, 2026-05-26): introduced a
  `Kind: "Bundle"` package type in `ptx-installer` registry; bundles install
  each listed member in order. Added bundle definitions `ai-assistant-basic`
  (claude-code + gemini-cli), `ai-assistant-full` (+ claude-desktop) and
  `mcp-ready` (python + claude-code + gemini-cli). Package definitions for
  `claude-code` / `claude-desktop` / `gemini-cli` already existed.
  Acceptance protocol: `docs/testing/acceptance-035.md` (PASS, 9/9).
- ✅ **Enhanced AI assistant detection** (merged to `main`, 2026-05-27): added
  `portunix package detect` (table + `--json`). It selects the installable AI
  assistants from the registry (`Kind: "Package"`, category
  `development/ai-tools`; bundles excluded) and runs each package's platform
  `verification.command` to report install state and parsed version. New
  `engine/detection.go`; no package JSON changes required. Acceptance protocol:
  `docs/testing/acceptance-035-detection.md` (PASS, 11/11).
- 🔄 **Cross-platform install (macOS)** (merged to `main`, 2026-05-27): added
  the `darwin` platform to `claude-desktop` (DMG download from
  `https://claude.ai/download/mac`, `hdiutil` mount, copy `Claude.app` to
  `/Applications`) and `gemini-cli` (`npm install -g @google/gemini-cli`). JSON
  only; no Go changes. Acceptance protocol:
  `docs/testing/acceptance-035-macos.md` (**CONDITIONAL** — structural checks
  PASS; functional macOS install tests T6–T8 still pending a macOS host; merged
  on owner override). The Windows/Linux variants for both packages already
  existed prior to this work.
- 🔄 **`--recommend-ai`, MCP auto-dependency hook, version selection**
  (merged to `main`, 2026-05-29):
  - `portunix install --recommend-ai` lists the installable AI assistants for
    the current platform with their install state and offers to install the
    missing ones (`--dry-run` / `-y` supported). Reuses `engine.DetectAIAssistants`.
  - `portunix mcp init` now offers to install any missing known assistant when
    at least one is already present (interactive prompt), and a new
    `--install-missing` flag auto-installs the requested assistant via
    ptx-installer in the non-interactive path. (Note: the wizard is `mcp init`,
    not `mcp serve init` as drafted above.)
  - `portunix install <pkg> --version <v>` resolves a version selector to a
    variant (by variant name or by the variant's `version` field) and
    `portunix package update <pkg>` reinstalls the latest available version
    (Force). New unit tests cover `Installer.ResolveVersion`.
  - Acceptance protocol: `docs/testing/acceptance-035-recommend-mcp-version.md`
    (**CONDITIONAL** — Windows CLI/logic 12/13 PASS; real install/update
    execution + Linux/macOS coverage pending container validation; merged on
    owner override).
- ⬜ **Open**: full version pinning/downgrade with installed-state tracking and
  uninstall — requires the version-management strategy decision (architect; see
  Discussion Points §3).

> Note: the original technical design below references the legacy
> `assets/install-packages.json` / `app/install/` layout. The project has since
> migrated to the registry-based `ptx-installer` (individual package JSON files);
> the bundle work follows that current architecture.

## Motivation
- **Consistency**: All AI assistants mentioned in MCP setup should be installable via Portunix
- **User Experience**: One-stop solution for development environment setup including AI tools
- **Dependency Management**: Automatic installation of required AI assistants before MCP configuration
- **Cross-Platform**: Unified installation experience across Windows, Linux, macOS
- **Integration**: Seamless integration with existing package management system

## Current State Analysis
Current package support (from `assets/install-packages.json`):
- ✅ Development tools: Python, Java, Go, VS Code, Maven
- ✅ Package managers: Chocolatey, WinGet
- ✅ Web browsers: Google Chrome
- ✅ Other tools: PowerShell, Claude Code (already supported!)
- 🔄 **Missing**: Claude Desktop installation support
- 🔄 **Missing**: Gemini CLI installation support
- 🔄 **Missing**: AI assistant detection and version management

## Requirements

### Core Enhancement: AI Assistant Package Definitions

#### Target AI Assistants for Installation Support
```
Priority 1 (Phase 1):
├─ Claude Code ✅ (already supported)
├─ Claude Desktop
└─ Gemini CLI

Priority 2 (Future):
├─ ChatGPT Desktop (if available)
├─ Anthropic Claude API CLI
└─ OpenAI CLI
```

#### Package Installation Commands
```bash
# Individual installations
portunix install claude-code          # Already supported
portunix install claude-desktop       # New
portunix install gemini-cli           # New

# Batch installations
portunix install ai-assistant-basic   # Claude Code + Gemini CLI
portunix install ai-assistant-full    # Claude Code + Claude Desktop + Gemini CLI

# Integration with existing profiles
portunix install default,claude-desktop  # Add Claude Desktop to default profile
```

### Package Definitions Structure

#### Claude Desktop Package Definition
```json
{
  "claude-desktop": {
    "description": "Anthropic Claude Desktop application",
    "category": "AI Assistant",
    "platforms": {
      "windows": {
        "installer_type": "exe",
        "download_url": "https://storage.googleapis.com/anthropic-claude/claude-desktop-windows.exe",
        "install_command": "claude-desktop-windows.exe /S",
        "detection_command": "where claude",
        "detection_paths": [
          "%LOCALAPPDATA%/Programs/Claude",
          "%PROGRAMFILES%/Claude"
        ]
      },
      "macos": {
        "installer_type": "dmg", 
        "download_url": "https://storage.googleapis.com/anthropic-claude/claude-desktop-macos.dmg",
        "install_command": "hdiutil attach claude-desktop-macos.dmg && cp -R '/Volumes/Claude/Claude.app' /Applications/",
        "detection_command": "ls /Applications/Claude.app",
        "detection_paths": ["/Applications/Claude.app"]
      },
      "linux": {
        "installer_type": "appimage",
        "download_url": "https://storage.googleapis.com/anthropic-claude/claude-desktop-linux.AppImage",
        "install_command": "chmod +x claude-desktop-linux.AppImage && mv claude-desktop-linux.AppImage ~/.local/bin/claude-desktop",
        "detection_command": "which claude-desktop",
        "detection_paths": [
          "~/.local/bin/claude-desktop",
          "/usr/local/bin/claude-desktop"
        ]
      }
    },
    "version": "latest",
    "dependencies": [],
    "post_install": {
      "create_desktop_entry": true,
      "add_to_path": false
    }
  }
}
```

#### Gemini CLI Package Definition  
```json
{
  "gemini-cli": {
    "description": "Google Gemini Command Line Interface",
    "category": "AI Assistant",
    "platforms": {
      "windows": {
        "installer_type": "zip",
        "download_url": "https://github.com/google/gemini-cli/releases/latest/download/gemini-cli-windows.zip",
        "install_command": "unzip gemini-cli-windows.zip && move gemini.exe %LOCALAPPDATA%\\bin\\",
        "detection_command": "where gemini",
        "detection_paths": ["%LOCALAPPDATA%\\bin\\gemini.exe"]
      },
      "macos": {
        "installer_type": "tar.gz",
        "download_url": "https://github.com/google/gemini-cli/releases/latest/download/gemini-cli-macos.tar.gz",
        "install_command": "tar -xzf gemini-cli-macos.tar.gz && cp gemini /usr/local/bin/",
        "detection_command": "which gemini",
        "detection_paths": ["/usr/local/bin/gemini"]
      },
      "linux": {
        "installer_type": "tar.gz",
        "download_url": "https://github.com/google/gemini-cli/releases/latest/download/gemini-cli-linux.tar.gz", 
        "install_command": "tar -xzf gemini-cli-linux.tar.gz && sudo cp gemini /usr/local/bin/",
        "detection_command": "which gemini",
        "detection_paths": ["/usr/local/bin/gemini"]
      }
    },
    "version": "latest",
    "dependencies": [],
    "post_install": {
      "create_desktop_entry": false,
      "add_to_path": true
    }
  }
}
```

### Installation Profiles Extension

#### New AI Assistant Profiles
```json
{
  "ai-assistant-basic": {
    "description": "Basic AI assistant setup for CLI development",
    "packages": ["claude-code", "gemini-cli"],
    "category": "AI Development"
  },
  "ai-assistant-full": {
    "description": "Complete AI assistant suite",
    "packages": ["claude-code", "claude-desktop", "gemini-cli"],
    "category": "AI Development"  
  },
  "mcp-ready": {
    "description": "MCP-ready development environment",
    "packages": ["python", "claude-code", "gemini-cli"],
    "category": "AI Development"
  }
}
```

### Integration with MCP Server Setup

#### Automatic Dependency Detection
```bash
# When running MCP server init, detect missing assistants
portunix mcp serve init

> "Detecting AI assistants..."
> "✅ Claude Code found"
> "❌ Claude Desktop not found"
> "❌ Gemini CLI not found"
> ""
> "Install missing assistants? [Y/n]: y"
> "Installing Claude Desktop..."
> "Installing Gemini CLI..."
> "✅ All AI assistants ready for MCP configuration"
```

#### Smart Installation Recommendations
```bash
# Based on user's platform and preferences
portunix install --recommend-ai

> "Recommended AI assistants for your system:"
> "  claude-code     - ✅ Already installed"  
> "  claude-desktop  - Desktop GUI interface"
> "  gemini-cli      - Google's CLI assistant"
> ""
> "Install recommended assistants? [Y/n]: "
```

## Technical Implementation

### Package Management Extensions

#### New Components
```
app/install/
├── ai_assistants.go        # AI assistant specific installation logic
├── detection.go           # Enhanced detection for AI tools
├── profiles_ai.go         # AI assistant installation profiles
└── integration.go         # MCP integration hooks

assets/
├── ai-assistant-packages.json  # AI assistant package definitions
└── install-packages.json       # Extended with AI assistants
```

#### Enhanced Detection System
```go
type AIAssistantDetector struct {
    Name             string
    DetectionCommand string
    DetectionPaths   []string
    VersionCommand   string
    ConfigPaths      []string
}

func DetectAIAssistants() []AIAssistantStatus {
    // Enhanced detection logic for AI assistants
    // Integration with MCP configuration detection
}
```

### Installation Flow Integration

#### Pre-MCP Installation Hook
```go
func PreMCPSetupHook(selectedAssistants []string) error {
    missing := DetectMissingAssistants(selectedAssistants)
    if len(missing) > 0 {
        return PromptInstallMissing(missing)
    }
    return nil
}
```

#### Version Management
```bash
# AI assistant version management
portunix install claude-desktop --version 1.2.3
portunix update claude-desktop
portunix list --category "AI Assistant"
```

## Implementation Plan

### Phase 1: Core AI Assistant Installation
1. Research and verify download URLs for Claude Desktop and Gemini CLI
2. Create package definitions in `ai-assistant-packages.json`
3. Implement AI assistant specific installation logic
4. Add AI assistant profiles to existing system

### Phase 2: Enhanced Detection & Integration
1. Implement advanced detection for AI assistants
2. Create integration hooks with MCP server setup
3. Add version management for AI assistants
4. Implement smart recommendation system

### Phase 3: User Experience & Automation
1. Automatic dependency resolution for MCP setup
2. Installation progress tracking and error handling
3. Post-installation verification and testing
4. Documentation and usage examples

### Phase 4: Extended Support & Maintenance
1. Add support for additional AI assistants (ChatGPT CLI, etc.)
2. Implement update notifications for AI assistants
3. Create uninstallation support
4. Performance optimization and caching

## Success Criteria
- [x] `portunix install claude-desktop` - Cross-platform Claude Desktop installation
      (Windows/Linux/macOS package definitions present; macOS functional test pending)
- [x] `portunix install gemini-cli` - Cross-platform Gemini CLI installation
      (Windows/Linux/macOS package definitions present; macOS functional test pending)
- [x] `portunix install ai-assistant-full` - Batch installation of AI assistants
      (also `ai-assistant-basic`, `mcp-ready` — implemented as bundles, see Implementation Progress)
- [x] Enhanced detection of installed AI assistants
      (`portunix package detect`, table + `--json`)
- [x] Integration with MCP server setup (automatic dependency installation)
      (`portunix mcp init` missing-assistant offer + `--install-missing`;
      merged to `main` 2026-05-29, CONDITIONAL acceptance)
- [x] Version management for AI assistants
      (`install <pkg> --version <v>` resolution + `portunix package update <pkg>`;
      merged to `main` 2026-05-29, CONDITIONAL — full pinning/downgrade tracked under Open above)
- [x] Smart installation recommendations
      (`portunix install --recommend-ai`; merged to `main` 2026-05-29, CONDITIONAL acceptance)
- [ ] Cross-platform compatibility (Windows, Linux, macOS)
- [ ] Proper error handling and rollback capabilities
- [ ] Documentation and usage examples

## Technical Notes
- Maintain single `portunix` binary
- Integration with existing package management system
- Cross-platform compatibility required
- Secure download and verification of packages
- Proper cleanup and error handling

## Dependencies
- Issue #034: MCP Server Command Restructuring + Interactive Wizard (integration point)
- Issue #010: Self-Update Command (existing package management system)
- Existing package management infrastructure

## Priority
**High** - Essential for seamless AI development environment setup and MCP integration.

## Labels
- enhancement
- package-management
- ai-integration
- installation
- cross-platform
- mcp

## Discussion Points

1. **Package Source Verification:**
   - Official download URLs for Claude Desktop and Gemini CLI
   - Package signature verification and security
   - Fallback sources and mirrors

2. **Platform-Specific Considerations:**
   - Different installation methods per platform
   - Platform-specific detection paths
   - Permission requirements (sudo, admin)

3. **Version Management Strategy:**
   - Automatic updates vs manual control
   - Version pinning for enterprise environments
   - Compatibility matrix with MCP server

4. **Integration with MCP Setup:**
   - When to prompt for installation during MCP setup
   - Silent installation options
   - Rollback on installation failure

5. **User Experience:**
   - Progress indicators for downloads
   - Offline installation support
   - Bandwidth optimization for large packages

---
*Created: 2025-09-11*
*Last updated: 2026-05-29*