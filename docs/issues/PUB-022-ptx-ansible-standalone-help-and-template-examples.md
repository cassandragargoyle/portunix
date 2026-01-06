# Issue #119: PTX-Ansible Standalone Help and Template Examples System

**Status**: 📋 Open
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, helper-binary, ptx-ansible, templates, user-experience, documentation
**Created**: 2026-01-02
**Assigned to**: Development Team

## Summary

Implement standalone `--help` support for `ptx-ansible` helper and add template-based `.ptxbook` example generation system with static documentation site as first template.

## Problem Statement

### 1. Standalone Help Not Working

Currently, running `./ptx-ansible --help` directly does not provide help output. The helper binary is designed to be called from the main `portunix` dispatcher but should also work standalone for debugging and development purposes.

**Current behavior:**

```bash
$ ./ptx-ansible --help
# No output or unexpected behavior
```

**Expected behavior:**

```bash
$ ./ptx-ansible --help
ptx-ansible - Portunix Ansible Infrastructure as Code Helper

Usage:
  ptx-ansible [command] [flags]

Commands:
  playbook    Execute, validate, or generate .ptxbook files
  mcp         MCP integration tools
  secrets     Secret management
  ...

Use "ptx-ansible [command] --help" for more information about a command.
```

### 2. Missing Template/Example Generation

ADR-016 defines `portunix playbook init <name> --template development|production|minimal` but currently no template system exists. Users need practical examples to understand `.ptxbook` format.

## Requirements

### Part 1: Standalone Help Support

**Must not break dispatcher pattern:**

- When called from `portunix playbook`, continue to work as helper
- When called directly (`./ptx-ansible --help`), show standalone help
- Detect invocation context (direct vs dispatcher) if needed

**Main portunix help must include playbook:**

- `./portunix --help` must list `playbook` command in common commands
- Add `playbook` to `CommandRegistry` in `src/cmd/help_registry.go`
- Add `playbook` to `essentials` list in `GetBasicCommands()`

**Implementation approach:**

```go
func main() {
    // Check if called with --help/-h as first argument
    if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
        showStandaloneHelp()
        return
    }

    // Normal dispatcher-compatible execution
    handleCommand(os.Args[1:])
}
```

### Part 2: Template Examples System

**Architecture:**

- Embedded templates using `//go:embed` directive
- OS-detection for platform-specific adjustments
- Parameter-based template selection
- Variable substitution in templates

**New commands:**

```bash
# Template management (consistent with virt template, plugin list pattern)
portunix playbook template list              # List available templates
portunix playbook template show <name>       # Show template details and parameters

# Generate from template
portunix playbook init [name] --template <template-name> [--engine <engine>] [--target <target>]
#                      ↑ optional - if omitted, uses current directory name

# Examples:
portunix playbook init my-docs --template static-docs --engine hugo
portunix playbook init my-docs --template static-docs --engine hugo --target container  # default
portunix playbook init my-docs --template static-docs --engine hugo --target local
portunix playbook init --template static-docs --engine docusaurus   # uses dir name, container target
```

### Part 3: First Template - Static Documentation Site

Create template for generating static documentation sites using different engines.

### Part 4: RBAC Disabled by Default for Standalone Usage

For standalone/development usage, RBAC (Role-Based Access Control) should be **disabled by default**.

**Problem:**
When running `playbook run` without RBAC setup, execution fails with "access denied: User not found" because the current system user is not registered in the RBAC database.

**Solution:**
Change default RBAC configuration in `src/helpers/ptx-ansible/rbac.go`:

```go
func GetDefaultRBACConfig() *RBACConfig {
    return &RBACConfig{
        Enabled: false,  // Changed from true - disabled by default for standalone usage
        // ...
    }
}
```

**Rationale:**

- Standalone/development usage should work "out of the box" without configuration
- Enterprise features (RBAC, audit, secrets) can be explicitly enabled when needed
- Users who need RBAC can enable it via configuration

**Template name:** `static-docs`

**Parameters:**

| Parameter | Required | Default | Values | Description |
|-----------|----------|---------|--------|-------------|
| `--engine` | No | `docusaurus` | `hugo`, `docusaurus`, `docsify` | Documentation engine to use |
| `--target` | No | `container` | `container`, `local` | Where to run the installation |

**Engine options:**

- `hugo` - Hugo static site generator
- `docusaurus` - Facebook Docusaurus
- `docsify` - Docsify documentation generator

**Target options:**

- `container` (default) - Run installation in isolated container (recommended, safe)
- `local` - Run installation directly on host machine

#### Template Source File (embedded in binary)

This Go template file is stored as `static-docs/hugo.ptxbook.tmpl` and embedded in the `ptx-ansible` binary. It contains template syntax with `{{ }}` placeholders:

```gotemplate
# File: templates/examples/static-docs/hugo.ptxbook.tmpl
apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "{{ .Name }}"
  description: "Static documentation site using Hugo"

spec:
  variables:
    site_title: "{{ .Name }} Documentation"
    output_dir: "./public"

  portunix:
    packages:
      - name: "hugo"
        variant: "extended"

  scripts:
    {{- if eq .OS "windows" }}
    build: "hugo.exe build"
    {{- else }}
    build: "hugo build"
    {{- end }}
```

#### Usage Examples

**With explicit name:**

```bash
portunix playbook init my-docs --template static-docs --engine hugo
# → Creates: my-docs.ptxbook
```

**Without name (uses current directory name):**

```bash
$ pwd
/home/user/company-docs

$ portunix playbook init --template static-docs --engine hugo
# → Creates: company-docs.ptxbook (name derived from current directory)
```

The system:

1. Determines project name (from argument or current directory name)
2. Loads template `static-docs/hugo.ptxbook.tmpl` from embedded FS
3. Detects current OS (`linux`)
4. Substitutes variables: `.Name` → `my-docs`, `.OS` → `linux`
5. Writes output to `<name>.ptxbook` in current directory

#### Generated Output File

The resulting `my-docs.ptxbook` file (clean YAML, no template syntax):

```yaml
# File: my-docs.ptxbook (generated)
apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "my-docs"
  description: "Static documentation site using Hugo"

spec:
  variables:
    site_title: "my-docs Documentation"
    output_dir: "./public"

  portunix:
    packages:
      - name: "hugo"
        variant: "extended"

  scripts:
    build: "hugo build"
```

#### Running the Generated Playbook

**IMPORTANT:** This issue includes not only playbook generation but also ensuring the generated playbook is executable.

After generation, user can execute the playbook:

```bash
portunix playbook run my-docs.ptxbook
```

**Expected execution flow (container target):**

```
$ portunix playbook run my-docs.ptxbook

Executing playbook: my-docs.ptxbook
Target: container (docker)
Image: ubuntu:22.04

[1/3] Creating container...
[2/3] Installing packages via Portunix...
      - hugo (extended)
[3/3] Running build script...
      - hugo build

Success! Documentation site generated in ./public/
```

**Expected execution flow (local target):**

```
$ portunix playbook run my-docs.ptxbook

Executing playbook: my-docs.ptxbook
Target: local

[1/2] Installing packages via Portunix...
      - hugo (extended)
[2/2] Running build script...
      - hugo build

Success! Documentation site generated in ./public/
```

The implementation must ensure end-to-end workflow works:
1. `portunix playbook init` → generates valid .ptxbook
2. `portunix playbook run` → executes the generated .ptxbook successfully
3. Final result → working documentation site in `./public/`

## Technical Implementation

### File Structure

```
src/helpers/ptx-ansible/
├── main.go              # Add standalone help detection
├── templates/           # New directory
│   ├── embed.go         # //go:embed directive
│   ├── engine.go        # Template engine
│   └── examples/        # Embedded template files
│       ├── static-docs/
│       │   ├── hugo.ptxbook.tmpl
│       │   ├── docusaurus.ptxbook.tmpl
│       │   ├── docsify.ptxbook.tmpl
│       │   └── metadata.json
│       ├── development/
│       │   └── ...
│       └── production/
│           └── ...
└── ...
```

### Template Engine

```go
package templates

import (
    "embed"
    "text/template"
)

//go:embed examples/*
var templateFS embed.FS

type TemplateContext struct {
    Name    string
    Engine  string
    OS      string
    Arch    string
    OSFamily string
    // Add more context variables as needed
}

func GenerateFromTemplate(templateName, outputPath string, ctx TemplateContext) error {
    // Auto-populate OS context using Portunix system detection
    // See "OS Detection Requirements" section below
    ctx.OS, ctx.Arch, ctx.OSFamily = detectSystemInfo()

    // Load and execute template
    // ...
}
```

### OS Detection Requirements

**IMPORTANT:** For OS detection, the implementation MUST use Portunix functionality, NOT `runtime.GOOS`.

**Priority order:**

1. **Use Portunix code directly** - Import and use the system detection code from `app/system/` package if possible (shared library approach)

2. **Fallback: Use `portunix system info` command** - If direct code import is not possible (e.g., due to module separation), execute `portunix system info --format json` and parse the output

**Rationale:**

- Portunix has sophisticated OS detection (distro, version, WSL, container detection)
- Ensures consistency across all Portunix tools
- Avoids duplicating detection logic
- Follows project guidelines from CLAUDE.md

**Example fallback implementation:**

```go
func detectSystemInfo() (os, arch, osFamily string) {
    // Try to execute portunix system info
    cmd := exec.Command("portunix", "system", "info", "--format", "json")
    output, err := cmd.Output()
    if err != nil {
        // Ultimate fallback if portunix not available
        return runtime.GOOS, runtime.GOARCH, getOSFamily(runtime.GOOS)
    }

    var info SystemInfo
    json.Unmarshal(output, &info)
    return info.OS, info.Arch, info.OSFamily
}
```

### Container Runtime Detection and Auto-Installation

When `--target container` is used (default), the system must ensure container runtime is available.

**Detection flow:**

1. **Check availability via `portunix system info`** - The command already reports container runtime status (Docker/Podman availability)

2. **If container runtime is NOT available:**
   - Automatically install using `portunix install docker` or `portunix install podman`
   - Portunix install logic already handles OS-specific selection (e.g., Podman preferred on Fedora, Docker on Ubuntu)
   - User should be informed about the installation

3. **If container runtime IS available:**
   - Proceed with playbook generation for container target

**Implementation:**

```go
func ensureContainerRuntime() (runtime string, err error) {
    // Step 1: Check if container runtime is available via portunix system info
    cmd := exec.Command("portunix", "system", "info", "--format", "json")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get system info: %w", err)
    }

    var info SystemInfo
    json.Unmarshal(output, &info)

    // Step 2: Check container availability from system info
    if info.ContainerRuntime != "" {
        // Docker or Podman already available
        return info.ContainerRuntime, nil
    }

    // Step 3: No container runtime - install via Portunix
    fmt.Println("Container runtime not found. Installing...")

    // Portunix will choose the appropriate runtime for the OS
    installCmd := exec.Command("portunix", "install", "docker")
    installCmd.Stdout = os.Stdout
    installCmd.Stderr = os.Stderr

    if err := installCmd.Run(); err != nil {
        return "", fmt.Errorf("failed to install container runtime: %w", err)
    }

    // Verify installation
    cmd = exec.Command("portunix", "system", "info", "--format", "json")
    output, _ = cmd.Output()
    json.Unmarshal(output, &info)

    if info.ContainerRuntime == "" {
        return "", fmt.Errorf("container runtime installation failed")
    }

    return info.ContainerRuntime, nil
}
```

**User feedback during auto-installation:**

```
$ portunix playbook init --template static-docs --engine hugo

Checking container runtime availability...
Container runtime not found.
Installing Docker via Portunix...
[portunix install docker output]
Docker installed successfully.

Generating my-project.ptxbook...
Done! Run 'portunix playbook run my-project.ptxbook' to execute.
```

**Generated .ptxbook with container target:**

```yaml
spec:
  environment:
    target: container
    runtime: docker  # or podman, detected/installed automatically
    image: "ubuntu:22.04"

  portunix:
    packages:
      - name: "hugo"
        variant: "extended"
```

### Template Metadata

Each template directory contains `metadata.json`:

```json
{
  "name": "static-docs",
  "description": "Static documentation site generator",
  "version": "1.0.0",
  "parameters": [
    {
      "name": "engine",
      "required": true,
      "type": "choice",
      "choices": ["hugo", "docusaurus", "docsify"],
      "description": "Documentation engine to use"
    },
    {
      "name": "target",
      "required": false,
      "type": "choice",
      "choices": ["container", "local"],
      "default": "container",
      "description": "Where to run the installation (container is recommended)"
    }
  ],
  "os_support": ["linux", "windows", "darwin"],
  "tags": ["documentation", "static-site", "web"]
}
```

## Acceptance Criteria

### Standalone Help

- [ ] `./ptx-ansible --help` displays comprehensive help
- [ ] `./ptx-ansible -h` works identically
- [ ] Help shows all available commands and their descriptions
- [ ] Version information included in help output
- [ ] Dispatcher-based invocation continues to work unchanged
- [ ] No breaking changes to existing `portunix playbook` commands
- [ ] `./portunix --help` lists `playbook` in common commands

### Template System

- [ ] `portunix playbook template list` lists all available templates
- [ ] `portunix playbook template show <name>` displays template details
- [ ] `portunix playbook init <name> --template <template>` generates .ptxbook file
- [ ] Templates are embedded in binary (no external files needed)
- [ ] Template metadata properly validated
- [ ] Missing required parameters show helpful error message

### Static-Docs Template

- [ ] Hugo template generates valid .ptxbook for Hugo site
- [ ] Docusaurus template generates valid .ptxbook for Docusaurus
- [ ] Docsify template generates valid .ptxbook for Docsify
- [ ] OS-specific paths and commands handled correctly
- [ ] Generated .ptxbook files pass validation
- [ ] Generated .ptxbook files can be executed successfully
- [ ] `--target container` (default) generates container-based playbook
- [ ] `--target local` generates local installation playbook

### Container Runtime Detection

- [ ] Uses `portunix system info` to detect container runtime availability
- [ ] Auto-installs container runtime via `portunix install` if not available
- [ ] Portunix chooses appropriate runtime for OS (Docker/Podman)
- [ ] User is informed about container runtime installation
- [ ] Detected/installed runtime is recorded in generated .ptxbook
- [ ] Graceful fallback if container installation fails (suggest `--target local`)

### End-to-End Playbook Execution

- [ ] Generated .ptxbook can be executed via `portunix playbook run`
- [ ] Container target: creates container, installs packages, runs scripts
- [ ] Local target: installs packages locally, runs scripts
- [ ] Hugo template produces working documentation site in `./public/`
- [ ] Docusaurus template produces working documentation site
- [ ] Docsify template produces working documentation site
- [ ] Error handling with clear messages if execution fails

### RBAC Default Configuration

- [ ] RBAC is disabled by default (`Enabled: false` in `GetDefaultRBACConfig()`)
- [ ] `playbook run` works without RBAC setup for standalone usage
- [ ] `playbook run --dry-run` works without RBAC setup
- [ ] RBAC can be explicitly enabled via configuration when needed

## Test Cases

### Unit Tests

- [ ] Template loading from embedded FS
- [ ] Template variable substitution
- [ ] OS detection via portunix system info
- [ ] Metadata JSON parsing
- [ ] Parameter validation
- [ ] Container runtime detection logic

### Integration Tests

- [ ] `./ptx-ansible --help` output verification
- [ ] `portunix playbook template list` listing
- [ ] `portunix playbook template show` details display
- [ ] Template generation for each engine
- [ ] Template generation with `--target container`
- [ ] Template generation with `--target local`
- [ ] Container runtime auto-installation (if not present)
- [ ] Cross-platform template generation (Linux/Windows)

### End-to-End Tests

- [ ] Full workflow: `init` → `run` → verify output (Hugo + container)
- [ ] Full workflow: `init` → `run` → verify output (Hugo + local)
- [ ] Full workflow: `init` → `run` → verify output (Docusaurus + container)
- [ ] Full workflow: `init` → `run` → verify output (Docsify + container)
- [ ] Verify generated documentation site is accessible/valid

### Edge Cases

- [ ] Missing required parameter (--engine)
- [ ] Invalid engine name
- [ ] Output file already exists
- [ ] Read-only output directory
- [ ] Template not found

## E2E Testing Environment

### Testing Architecture

End-to-end tests require isolated VM environment to avoid host contamination. Testing workflow:

```
┌─────────────────┐     HTTP      ┌─────────────────┐
│   Host Machine  │──────────────▶│   Target VM     │
│                 │               │                 │
│ 1. Python HTTP  │               │ 3. Download     │
│    file server  │               │    portunix     │
│                 │               │                 │
│ 2. SSH to VM    │◀─────────────▶│ 4. Install SSH  │
│    via portunix │    SSH        │    via portunix │
│                 │               │                 │
│ 5. Execute test │               │ 6. Run playbook │
│    script       │               │    in VM        │
└─────────────────┘               └─────────────────┘
```

### Existing Testing Infrastructure

**Note:** The following components already exist in the project:

- **`scripts/file-server.py`** - Python HTTP file server for VM/container downloads
- **`test/integration/windows11_ea_vm_playbook_test.py`** - VM playbook test framework

### Test Setup Steps

1. **Start Python file server on host** (existing tool)

   ```bash
   # Use existing file-server.py - auto-generates install scripts
   python3 scripts/file-server.py --port 8080

   # Server provides:
   # - Auto-detected local IP address
   # - Linux install script: curl -fsSL http://<ip>:8080/install-from-server.sh | sudo bash
   # - Windows install script: install-from-server.ps1
   ```

2. **Download portunix in VM**

   ```bash
   # From inside VM - use auto-generated install script
   curl -fsSL http://<host-ip>:8080/install-from-server.sh | sudo bash

   # Or manual download
   curl -O http://<host-ip>:8080/portunix
   chmod +x portunix
   ```

3. **Install SSH in VM via portunix**

   ```bash
   ./portunix install openssh-server
   ```

4. **Connect from host via SSH**

   ```bash
   # Using portunix SSH integration
   portunix virt ssh <vm-name>
   ```

5. **Execute test script remotely**

   ```bash
   # Test script runs on VM via SSH
   ./test-playbook-e2e.sh
   ```

### Test Script Requirements

The test script (`test/integration/playbook_e2e_test.sh`) should:

- [ ] Generate playbook using `portunix playbook init`
- [ ] Run playbook using `portunix playbook run`
- [ ] Verify packages were installed correctly
- [ ] Verify scripts executed successfully
- [ ] Report success/failure back to host

### Benefits of VM-based Testing

- **Isolation**: No host contamination from package installations
- **Reproducibility**: Clean VM state for each test run
- **Real-world**: Tests actual installation and execution flow
- **Cross-platform**: Can test on different Linux distributions

## Future Templates (Roadmap)

After `static-docs`, plan for additional templates:

1. **development** - Local development environment setup
2. **production** - Production deployment playbook
3. **container-app** - Containerized application setup
4. **minimal** - Minimal starting point
5. **microservices** - Multi-service deployment
6. **database** - Database setup and configuration

## Dependencies

- Issue #056: Ansible Infrastructure as Code Integration (parent)
- Issue #059: Playbook Help Command (related fix, implemented)
- Issue #074: Post-Release Documentation (static site context)
- Issue #075: Hugo Installation Support (hugo engine dependency)

## External Requests (GitHub)

- **[GitHub #22](../public/github-022-docusaurus-multiplatform.md)**: Implementácia podpory pre Docusaurus na multiplatformnom prostredí
  - Reporter: @Roman-Kazicka
  - Relates to: Docusaurus engine support in static-docs template (Part 3)

## Related Documentation

- [ADR-016](../../adr/016-ansible-infrastructure-as-code-integration.md) - Ansible IaC Integration
- [README.md](../../../src/helpers/ptx-ansible/README.md) - Helper documentation

## Implementation Notes

### Dispatcher Compatibility

The main `portunix` binary dispatches to `ptx-ansible` with specific command structure:

```
portunix playbook run foo.ptxbook
  → ptx-ansible playbook run foo.ptxbook
```

Standalone help must work without breaking this pattern. Key is detecting whether `--help` is the first argument (standalone) or comes after a command (dispatcher context).

### Template Selection Logic

```
User runs: portunix playbook init my-docs --template static-docs --engine hugo

1. Parse arguments
2. Load template metadata from embedded FS
3. Validate required parameters (engine is required for static-docs)
4. Detect current OS
5. Select appropriate .ptxbook.tmpl file (hugo.ptxbook.tmpl)
6. Execute template with context
7. Write output to my-docs.ptxbook
8. Display success message with next steps
```

## Success Metrics

1. **Usability**: Users can discover help without main binary
2. **Productivity**: New users can start with working examples in < 1 minute
3. **Discoverability**: `template list` and `template show` help users understand capabilities
4. **Correctness**: All generated .ptxbook files pass validation

---

## Part 5: Fix `portunix virt start` Command

### Problem

Running `./portunix virt start win11` fails with exit status 1 without providing useful error information.

**Current behavior:**

```bash
$ ./portunix virt start win11
Starting VM 'win11'...
Error starting VM: exit status 1
```

**Expected behavior:**

```bash
$ ./portunix virt start win11
Starting VM 'win11'...
VM 'win11' started successfully.
```

Or if there's an error, provide detailed information:

```bash
$ ./portunix virt start win11
Starting VM 'win11'...
Error starting VM: [specific error message from libvirt/virsh]
```

### Requirements

- [ ] `portunix virt start <vm-name>` should start the VM successfully
- [ ] Error messages should include the underlying error from virsh/libvirt
- [ ] If VM is already running, provide appropriate message
- [ ] Handle common error cases (VM not found, insufficient permissions, etc.)

### Acceptance Criteria

- [ ] `portunix virt start win11` successfully starts the VM
- [ ] Error output includes detailed error message from virsh
- [ ] Command returns proper exit codes (0 for success, non-zero for failure)

---

**Created**: 2026-01-02
**Last Updated**: 2026-01-02 (Part 5: Fix virt start command added)
