# Helper Binary Development Guide

## Overview

Starting with Portunix v1.6.0, the project uses a Git-like dispatcher architecture where functionality is split between the main binary (`portunix`) and helper binaries (`ptx-*`). This guide explains how to develop, build, and integrate helper binaries.

## Architecture

```
portunix                  # Main dispatcher
├── ptx-container         # Container management (docker, podman)
├── ptx-mcp               # MCP server functionality
└── ptx-plugin-*          # gRPC plugins (future)
```

## Development Guidelines

### Helper Binary Structure

Helper binaries are located in `src/helpers/` and follow this structure:

```
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

## Helper Binary Checklist

Before submitting a new helper binary:

- [ ] Helper binary follows naming convention `ptx-{feature}`
- [ ] Located in `src/helpers/ptx-{feature}/`
- [ ] Has proper Go module setup
- [ ] Implements version flag with correct format
- [ ] Registered in dispatcher
- [ ] Added to build system (Makefile, build-with-version.sh)
- [ ] Added to GoReleaser configuration
- [ ] Added to installation system
- [ ] Added to update system
- [ ] Has unit tests
- [ ] Has integration tests
- [ ] Documentation updated
- [ ] Tested on Linux and Windows

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

## Examples

Refer to existing helper binaries for implementation examples:
- `src/helpers/ptx-container/` - Container management
- `src/helpers/ptx-mcp/` - MCP server functionality

## Support

For questions about helper binary development:
1. Check existing helper implementations
2. Review dispatcher code in `src/dispatcher/`
3. Create an issue with `enhancement` label
4. Follow the contribution guidelines in `docs/contributing/`

---

*This guide is part of Issue #051: Git-like Dispatcher with Python Distribution Architecture*