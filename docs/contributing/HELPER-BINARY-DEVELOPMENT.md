# Helper Binary Development Guide

## Overview

Starting with Portunix v1.6.0, the project uses a Git-like dispatcher architecture where functionality is split between the main binary (`portunix`)
and helper binaries (`ptx-*`). This guide explains how to develop, build, and integrate helper binaries.

## Architecture

```text
portunix                  # Main dispatcher
├── ptx-ansible           # Infrastructure as Code (Ansible)
├── ptx-aiops             # AI/ML operations (Ollama, GPU)
├── ptx-container         # Container management (docker, podman)
├── ptx-credential        # Secure credential storage
├── ptx-installer         # Package installation engine
├── ptx-make              # Cross-platform Makefile utilities
├── ptx-mcp               # MCP server functionality
├── ptx-pft               # Product Feedback Tool
├── ptx-prompting         # Prompt templates management
├── ptx-python            # Python environment management
├── ptx-virt              # Virtualization (QEMU/KVM, VirtualBox)
└── ptx-plugin-*          # gRPC plugins (future)
```

## Development Guidelines

### Helper Binary Structure

Helper binaries are located in `src/helpers/` and follow this structure:

```text
src/helpers/
├── ptx-container/
│   ├── main.go          # Entry point
│   ├── go.mod           # Module definition
│   └── commands/        # Command implementations
└── ptx-mcp/
    ├── main.go
    ├── go.mod
    └── handlers/
```

### Creating a New Helper Binary

1. **Create Directory Structure**

```bash
mkdir -p src/helpers/ptx-newfeature
cd src/helpers/ptx-newfeature
```

2. **Initialize Go Module**

```bash
go mod init portunix.ai/portunix/src/helpers/ptx-newfeature
```

3. **Create main.go**

```go
   package main

   import (
       "fmt"
       "os"
       "github.com/spf13/cobra"
   )

   var version = "dev"

   var rootCmd = &cobra.Command{
       Use:   "ptx-newfeature",
       Short: "Portunix New Feature Helper",
       Long: `ptx-newfeature is a helper binary for Portunix that handles new feature operations.

   This binary is typically invoked by the main portunix dispatcher and should not be used directly.`,
       Version: version,
       Run: func(cmd *cobra.Command, args []string) {
           handleCommand(args)
       },
   }

   func handleCommand(args []string) {
       if len(args) == 0 {
           fmt.Println("No command specified")
           return
       }

       command := args[0]
       subArgs := args[1:]

       switch command {
       case "newfeature":
           if len(subArgs) == 0 {
               showHelp()
           } else {
               executeCommand(subArgs)
           }
       default:
           fmt.Printf("Unknown command: %s\n", command)
       }
   }

   func showHelp() {
       fmt.Println("Usage: portunix newfeature [subcommand]")
       fmt.Println("\nAvailable subcommands:")
       fmt.Println("  command1  - Description of command1")
       fmt.Println("  command2  - Description of command2")
       fmt.Println("  --help    - Show this help")
   }

   func executeCommand(args []string) {
       fmt.Printf("New feature command %s not yet implemented\n", args[0])
   }

   func init() {
       rootCmd.SetVersionTemplate("ptx-newfeature version {{.Version}}\n")
   }

   func main() {
       if err := rootCmd.Execute(); err != nil {
           fmt.Fprintf(os.Stderr, "Error: %v\n", err)
           os.Exit(1)
       }
   }
```

### Dispatcher Integration

4. **Register in Dispatcher**

   Edit `src/dispatcher/dispatcher.go` to add your helper:

```go
   func (d *Dispatcher) registerHelpers() {
       // Existing helpers...

       // Your new helper
       d.helpers["ptx-newfeature"] = &HelperConfig{
           Commands: []string{"newfeature", "alias1", "alias2"},
           Binary:   "ptx-newfeature",
           Required: false,
       }
   }
```

5. **Update Build System**

   Add to `Makefile`:

```makefile
   build-helpers: ## Build all helper binaries
       @echo "🔧 Building helper binaries..."
       @cd src/helpers/ptx-container && go build -o ../../../ptx-container .
       @cd src/helpers/ptx-mcp && go build -o ../../../ptx-mcp .
       @cd src/helpers/ptx-newfeature && go build -o ../../../ptx-newfeature .
       @echo "✅ Helper binaries built: ptx-container, ptx-mcp, ptx-newfeature"
```

   Add to `build-with-version.sh`:

```bash
   # Build ptx-newfeature
   echo "Building ptx-newfeature..."
   cd src/helpers/ptx-newfeature
   go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-newfeature .
   NEWFEATURE_BUILD=$?
   cd ../../..
```

6. **Update GoReleaser**

   Add to `.goreleaser.yml`:

```yaml
   builds:
     # ... existing builds ...

     # Helper binary: ptx-newfeature
     - id: ptx-newfeature
       binary: ptx-newfeature
       main: ./src/helpers/ptx-newfeature/main.go
       goos:
         - linux
         - windows
         - darwin
       goarch:
         - amd64
         - arm64
       ignore:
         - goos: darwin
           goarch: arm64
       ldflags:
         - -X main.version={{ .Version }}
         - -s -w
       env:
         - CGO_ENABLED=0

   archives:
     - id: default
       builds:
         - portunix
         - ptx-container
         - ptx-mcp
         - ptx-newfeature  # Add here
```

7. **Update Installation**

   Add to `src/app/selfinstall/install.go`:

```go
   helpers := []string{"ptx-container", "ptx-mcp", "ptx-newfeature"}
```

   Add to `src/app/update/github.go`:

```go
   expectedBinaries := []string{"portunix", "ptx-container", "ptx-mcp", "ptx-newfeature"}
```

## Best Practices

### Command Handling

- Always handle the case where no arguments are provided
- Show helpful usage information
- Support `--help` flag
- Return appropriate exit codes

### Error Handling

- Use clear, user-friendly error messages
- Log errors to stderr, not stdout
- Provide actionable error messages when possible

### Version Management

- Use the `version` variable that gets set at build time
- Implement `--version` flag
- Version template should follow: `ptx-{name} version {{.Version}}`

### Dependencies

- Keep dependencies minimal
- Use `github.com/spf13/cobra` for CLI framework (already used by main binary)
- Share common dependencies with main binary when possible

### Performance

- Helper binaries should start quickly (< 100ms)
- Avoid heavy initialization in main()
- Use lazy loading for expensive operations

## Testing

### Unit Testing

Create tests in the same directory as your helper:

```go
// main_test.go
package main

import "testing"

func TestHandleCommand(t *testing.T) {
    // Test your command handling logic
}
```

### Integration Testing

Create integration tests in `test/integration/`:

```go
// test/integration/ptx_newfeature_test.go
package integration

import (
    "testing"
    "portunix.ai/portunix/test/testframework"
)

func TestNewFeatureHelper(t *testing.T) {
    tf := testframework.NewTestFramework("NewFeature_Helper")
    tf.Start(t, "Test new feature helper binary integration")

    success := true
    defer tf.Finish(t, success)

    // Test helper binary execution
    tf.Step(t, "Test helper binary help")
    // ... test implementation
}
```

### Build Testing

```bash
# Test local build
make build-helpers

# Test with version
./build-with-version.sh v1.6.1

# Test installation
./portunix install-self --silent --path /tmp/test-install/portunix
```

## Deployment

### Release Process

1. Helper binaries are automatically built with GoReleaser
2. All binaries are packaged together in release archives
3. Installation scripts handle all binaries automatically
4. Update system downloads and installs all binaries

### Version Synchronization

- All binaries (main + helpers) use the same version number
- Version is embedded at build time using ldflags
- Version validation ensures compatibility between main and helpers

## Complete Helper Binary Checklist

This is the **definitive checklist** for creating a new helper binary. Follow each step in order.

### Phase 1: Source Code

- [ ] **1.1** Create directory: `src/helpers/ptx-{name}/`
- [ ] **1.2** Initialize Go module: `go mod init portunix.ai/portunix/src/helpers/ptx-{name}`
- [ ] **1.3** Create `main.go` with:
  - `var version = "dev"` for version injection
  - Cobra root command with `Use: "ptx-{name}"`
  - `--version` flag support
  - Proper command handling
- [ ] **1.4** Add unit tests: `*_test.go`
- [ ] **1.5** Run `go mod tidy` to resolve dependencies

### Phase 2: Dispatcher Integration

- [ ] **2.1** Edit `src/dispatcher/dispatcher.go`:

```go
  d.helpers["ptx-{name}"] = &HelperConfig{
      Commands: []string{"{command1}", "{command2}"},
      Binary:   "ptx-{name}",
      Required: false,
  }
```

### Phase 3: Build System (Makefile)

- [ ] **3.1** Add to `build-helpers:` target:

```makefile
  @cd src/helpers/ptx-{name} && go build -o ../../../ptx-{name}$(EXE_EXT) .
```

- [ ] **3.2** Update `build-helpers:` echo message with new binary name
- [ ] **3.3** Add to `clean:` target:

```makefile
  -$(RM) ... ptx-{name}$(EXE_EXT) ...
```

- [ ] **3.4** Add to `build-all-platforms:` target (for cross-platform distribution):

```makefile
  cd src/helpers/ptx-{name} && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o $$abs_dist/ptx-{name}$$ext . && cd ../../..;
```

### Phase 4: Version Build Script

- [ ] **4.1** Edit `build-with-version.sh`:

```bash
  # Build ptx-{name}
  echo "Building ptx-{name}..."
  cd src/helpers/ptx-{name}
  go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-{name} .
  NAME_BUILD=$?
  cd ../../..
```

- [ ] **4.2** Add exit code check at the end of script

### Phase 5: GoReleaser Configuration

- [ ] **5.1** Add build configuration to `.goreleaser.yml`:

```yaml
  # Helper binary: ptx-{name} ({description})
  - id: ptx-{name}
    binary: ptx-{name}
    main: ./
    dir: ./src/helpers/ptx-{name}/
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: darwin
        goarch: arm64
    ldflags:
      - -X main.version={{ .Version }}
      - -s -w
    env:
      - CGO_ENABLED=0
```

- [ ] **5.2** Add to `archives.builds` list:

```yaml
  archives:
    - id: default
      builds:
        - ...
        - ptx-{name}
```

- [ ] **5.3** Update `release.header` with new helper description
- [ ] **5.4** Update `release.footer` installation notes

### Phase 6: Installation System

- [ ] **6.1** Edit `src/app/selfinstall/install.go`:
  - Add to `helpers` slice
- [ ] **6.2** Edit `src/app/update/github.go`:
  - Add to `expectedBinaries` slice

### Phase 7: Deploy Scripts

- [ ] **7.1** Edit `scripts/deploy-local.py`:
  - Add `"ptx-{name}"` to `HELPER_BINARIES` list
- [ ] **7.2** Edit `scripts/undeploy-local.py`:
  - Add `"ptx-{name}"` to binaries list

### Phase 8: Platform Archives Script

- [ ] **8.1** Edit `scripts/create-platform-archives.py`:
  - Add `"ptx-{name}"` to `BINARIES` list

### Phase 9: Documentation

- [ ] **9.1** Create/update issue in `docs/issues/internal/{number}-ptx-{name}-*.md`
- [ ] **9.2** Update `docs/issues/README.md` table
- [ ] **9.3** Create ADR if architecturally significant: `docs/adr/{number}-ptx-{name}-*.md`
- [ ] **9.4** Update `docs/FEATURES_OVERVIEW.md` if user-facing feature

### Phase 10: Testing

- [ ] **10.1** Unit tests pass: `go test ./src/helpers/ptx-{name}/...`
- [ ] **10.2** Build works: `make build`
- [ ] **10.3** Helper responds to `--version`
- [ ] **10.4** Dispatcher routes commands correctly
- [ ] **10.5** Integration test in `test/integration/ptx_{name}_test.go`
- [ ] **10.6** Test on Linux
- [ ] **10.7** Test on Windows (if possible)

### Phase 11: Release Verification

- [ ] **11.1** Snapshot build works: `python3 scripts/make-release.py vX.Y.Z-SNAPSHOT`
- [ ] **11.2** All binaries present in `dist/` archives
- [ ] **11.3** `make deploy-local` installs all binaries

---

## Quick Reference: Files to Modify

| File | Section/Change |
| ---- | -------------- |
| `src/helpers/ptx-{name}/` | New directory with main.go, go.mod |
| `src/dispatcher/dispatcher.go` | Add to `registerHelpers()` |
| `Makefile` | `build-helpers`, `clean`, `build-all-platforms` |
| `build-with-version.sh` | Add build block |
| `.goreleaser.yml` | `builds`, `archives.builds`, `release.header/footer` |
| `src/app/selfinstall/install.go` | Add to helpers slice |
| `src/app/update/github.go` | Add to expectedBinaries |
| `scripts/deploy-local.py` | Add to HELPER_BINARIES |
| `scripts/undeploy-local.py` | Add to binaries list |
| `scripts/create-platform-archives.py` | Add to BINARIES |
| `docs/issues/README.md` | Add issue entry |
| `docs/FEATURES_OVERVIEW.md` | Add feature description |

---

## Example: Adding ptx-trace Helper

```bash
# 1. Create structure
mkdir -p src/helpers/ptx-trace
cd src/helpers/ptx-trace
go mod init portunix.ai/portunix/src/helpers/ptx-trace

# 2. Create main.go (see template above)

# 3. Update all files listed in Quick Reference

# 4. Build and test
make build
./portunix trace --help

# 5. Test release build
python3 scripts/make-release.py v1.9.3-SNAPSHOT
```

---

## Troubleshooting

### Common Issues

**Helper not found by dispatcher:**

- Check if binary is in the same directory as main binary
- Verify binary has correct name and executable permissions
- Check dispatcher registration in `src/dispatcher/dispatcher.go`

**Version mismatch:**

- Ensure all binaries use the same version variable
- Check ldflags in build scripts
- Verify version validation logic

**Build failures:**

- Check Go module path is correct
- Verify all dependencies are available
- Ensure build scripts are updated

### Debug Mode

Enable debug output in dispatcher:

```bash
export PORTUNIX_DEBUG=1
./portunix your-command
```

## MCP Tool Integration

When a helper binary needs to expose functionality to AI assistants (Claude Code, etc.), you can add MCP tools to the existing `ptx-mcp` helper.

### MCP Tools Overview

MCP (Model Context Protocol) tools are defined in `src/app/mcp/handlers.go`. The `ptx-mcp` helper binary uses this package to serve tools via JSON-RPC.

### Adding New MCP Tools

#### Step 1: Add Tool Definition

In `src/app/mcp/handlers.go`, find `handleToolsList()` and add your tool definition:

```go
// In handleToolsList() - add to the tools slice
{
    "name":        "mytool_action",
    "description": "Description of what the tool does",
    "inputSchema": map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "Description of param1",
            },
            "param2": map[string]interface{}{
                "type":        "integer",
                "description": "Optional parameter",
                "default":     10,
            },
        },
        "required": []string{"param1"},
    },
},
```

#### Step 2: Add Switch Case

In `handleToolsCall()`, add a case for your tool:

```go
case "mytool_action":
    result, err = s.handleMyToolAction(request.Arguments)
```

#### Step 3: Implement Handler

Add handler function that calls the helper binary:

```go
// executePtxMyTool executes a ptx-mytool command
func (s *Server) executePtxMyTool(args ...string) (string, error) {
    // Get directory containing current executable (ptx-mcp)
    execPath, err := os.Executable()
    if err != nil {
        return "", fmt.Errorf("failed to get executable path: %w", err)
    }
    execDir := filepath.Dir(execPath)

    // Find portunix binary in the same directory
    binaryPath := filepath.Join(execDir, "portunix")
    if runtime.GOOS == "windows" {
        binaryPath += ".exe"
    }

    // Check if portunix exists
    if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
        return "", fmt.Errorf("portunix binary not found at %s", binaryPath)
    }

    // Build command with subcommand
    cmdArgs := append([]string{"mytool"}, args...)
    cmd := exec.Command(binaryPath, cmdArgs...)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return string(output), fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
    }

    return string(output), nil
}

// handleMyToolAction handles the mytool_action MCP tool
func (s *Server) handleMyToolAction(args map[string]interface{}) (interface{}, error) {
    // Get required parameter
    param1, ok := args["param1"].(string)
    if !ok || param1 == "" {
        return nil, fmt.Errorf("param1 is required")
    }

    cmdArgs := []string{"action", param1}

    // Get optional parameter
    if param2, ok := args["param2"].(float64); ok {
        cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", int(param2)))
    }

    // Request JSON output for structured data
    cmdArgs = append(cmdArgs, "--format", "json")

    output, err := s.executePtxMyTool(cmdArgs...)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "status": "success",
        "result": output,
    }, nil
}
```

### Important Notes

1. **Binary Path Resolution**: The MCP server runs as `ptx-mcp`, so `os.Executable()` returns the helper
   path, not `portunix`. Use `filepath.Dir()` to find the directory and locate `portunix` there.

2. **JSON Output**: Request `--format json` from CLI commands for structured data that AI can parse.

3. **Error Handling**: Return clear error messages that help AI assistants understand what went wrong.

4. **Parameter Types**: JSON numbers come as `float64`, convert to `int` when needed.

5. **Optional Parameters**: Check with type assertion and use sensible defaults.

### Rebuild After Changes

After modifying `src/app/mcp/handlers.go`, rebuild the `ptx-mcp` helper:

```bash
cd src/helpers/ptx-mcp
go build -o ../../../ptx-mcp .
cd ../../..

# Copy to same directory as portunix for testing
cp ptx-mcp bin/
```

### Testing MCP Tools

Test your new tool with JSON-RPC:

```bash
# List all tools (verify your tool appears)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./portunix mcp serve 2>/dev/null | grep mytool

# Call your tool
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"mytool_action","arguments":{"param1":"test"}}}' | ./portunix mcp serve 2>&1 | tail -1
```

### MCP Tool Naming Conventions

- Use snake_case for tool names: `trace_start_session`, `vm_create`
- Group related tools with common prefix: `trace_*`, `vm_*`, `pft_*`
- Use descriptive action verbs: `start`, `stop`, `list`, `get`, `create`, `delete`

### Example: PTX-TRACE MCP Integration

See Issue #141 for a complete example of MCP integration:

- **8 tools added**: `trace_start_session`, `trace_end_session`, `trace_list_sessions`, `trace_view_events`, `trace_get_statistics`,
`trace_show_errors`, `trace_export_ai`, `trace_query`
- **Handler pattern**: `executePtxTrace()` calls `portunix trace` commands
- **JSON output**: All tools request `--format json` for structured responses

---

## Examples

Refer to existing helper binaries for implementation examples:

- `src/helpers/ptx-container/` - Container management
- `src/helpers/ptx-mcp/` - MCP server functionality
- `src/helpers/ptx-trace/` - Tracing system with MCP integration (Issue #141)

## Support

For questions about helper binary development:

1. Check existing helper implementations
2. Review dispatcher code in `src/dispatcher/`
3. Create an issue with `enhancement` label
4. Follow the contribution guidelines in `docs/contributing/`

---

This guide is part of Issue #051: Git-like Dispatcher with Python Distribution Architecture.
