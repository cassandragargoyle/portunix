# Issue #41: Node.js/npm Installation Support

## Summary
Add Node.js and npm installation support to Portunix to enable AI assistant integrations and complete development environments.

## Problem Description

During E2E testing of Claude Code container integration (`TestClaudeCodeContainerInstallAndSetup`), we discovered that **Portunix cannot install Node.js/npm**, which are essential dependencies for modern development environments and AI assistant integrations.

**Current Error:**
```
bash: line 1: npm: command not found
```

**Test Context:**
- Test tries to install Claude Code CLI via npm
- Container lacks Node.js/npm runtime  
- Current workaround uses manual `apt-get install nodejs npm` 
- This violates Portunix design principle of unified package management

## Expected Behavior

Portunix should provide Node.js/npm installation through its unified install command:

```bash
# Should work:
portunix install nodejs
portunix install npm

# Or combined:
portunix install nodejs-npm
```

## Current Available Packages

Portunix currently supports these packages:
- ✅ `java` - Java Development Kit (OpenJDK)  
- ✅ `python` - Python programming language
- ✅ `go` - Go programming language
- ✅ `claude-code` - Anthropic's official CLI (requires npm!)
- ❌ `nodejs` - **MISSING**
- ❌ `npm` - **MISSING**

## Impact

**Blocking Issues:**
1. **Claude Code Installation Fails** - `claude-code` package requires npm but Portunix can't install it
2. **E2E Testing Broken** - Container tests fail due to missing Node.js runtime
3. **Development Environment Incomplete** - Modern development requires Node.js ecosystem
4. **AI Assistant Integration** - Many AI tools (Claude Code, GitHub Copilot CLI) depend on Node.js

**Affected Components:**
- 🔴 `test/integration/claude_code_container_install_test.go` - E2E test fails
- 🔴 `claude-code` package installation - dependency missing  
- 🔴 AI assistant integrations - incomplete toolchain

## Proposed Solution

### 1. Add Node.js Package Definition
Add to `assets/install-packages.json`:

```json
{
  "nodejs": {
    "name": "Node.js JavaScript Runtime",
    "description": "JavaScript runtime built on Chrome's V8 engine",
    "windows": {
      "chocolatey": "nodejs",
      "winget": "OpenJS.NodeJS"
    },
    "linux": {
      "ubuntu": {
        "apt": ["nodejs", "npm"],
        "snap": "node --classic"
      },
      "fedora": {
        "dnf": ["nodejs", "npm"]
      },
      "direct": {
        "url": "https://nodejs.org/dist/latest/node-{version}-linux-x64.tar.xz",
        "install_script": "install-nodejs.sh"
      }
    }
  }
}
```

### 2. Package Variants
Support different Node.js versions:
- `nodejs --variant 18` - Node.js LTS 18.x  
- `nodejs --variant 20` - Node.js LTS 20.x (default)
- `nodejs --variant latest` - Latest stable version

### 3. Integration with Existing Packages
Update `claude-code` package to depend on `nodejs`:
- **Prerequisite Check**: Before installing Claude Code, check if Node.js/npm is available
- **Auto-installation**: If Node.js/npm is missing, automatically install it first
- **Version Validation**: Ensure compatible Node.js version for Claude Code
- **Clear Messaging**: Inform user about prerequisite installation
- **Error Handling**: Provide clear error messages if prerequisite installation fails

## Acceptance Criteria

- [ ] `portunix install nodejs` works on Windows/Linux
- [ ] `portunix install npm` works (or included with nodejs)  
- [ ] Multiple Node.js versions supported via `--variant`
- [ ] `claude-code` package auto-installs nodejs dependency when missing
- [ ] `portunix install claude-code` displays prerequisite installation messages
- [ ] Prerequisites are validated before main package installation
- [ ] E2E test `TestClaudeCodeContainerInstallAndSetup` passes
- [ ] Installation verified with `node --version` and `npm --version`
- [ ] Cross-platform compatibility (Windows/Ubuntu/Fedora)
- [ ] Graceful handling of prerequisite installation failures

## Files to Modify

1. **`assets/install-packages.json`** - Add nodejs package definition
2. **`app/install/nodejs.go`** - Node.js installation logic
3. **`app/install/claude.go`** - Add nodejs dependency check
4. **`test/integration/claude_code_container_install_test.go`** - Use Portunix for nodejs
5. **`docs/FEATURES_OVERVIEW.md`** - Document nodejs support

## Priority

**HIGH** - Blocks AI assistant integration and E2E testing

## Test Case

After implementation, this should work:

```bash
# Test standalone Node.js installation
portunix install nodejs
node --version  # Should output: v20.x.x
npm --version   # Should output: 10.x.x

# Test Claude Code installation with auto-prerequisite
portunix install claude-code
# Expected output:
# "Checking prerequisites for claude-code..."
# "Node.js not found, installing nodejs..."
# "Installing nodejs..."
# "nodejs installed successfully"
# "Installing claude-code..."
# "claude-code installed successfully"

claude --version  # Should work without errors

# E2E test should pass
go test ./test/integration/claude_code_container_install_test.go -v
```

---

**Reporter:** Claude Code Assistant  
**Date:** 2025-09-12  
**Category:** Feature Request / Bug Fix  
**Priority:** High  
**Labels:** installation, nodejs, ai-assistant, e2e-testing  
**Status:** 📋 Open