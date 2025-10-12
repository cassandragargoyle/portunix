# Portunix Completion Command

## Quick Start

The `completion` command generates shell completion scripts for bash, zsh, fish, and PowerShell, enabling tab completion for Portunix commands and parameters.

### Simplest Usage
```bash
# Generate completion for bash
portunix completion bash

# Generate completion for current shell
portunix completion
```

### Basic Syntax
```bash
portunix completion [shell] [options]
```

### Supported Shells
- `bash` - Bash completion
- `zsh` - Zsh completion
- `fish` - Fish completion
- `powershell` - PowerShell completion
- Auto-detection if no shell specified

## Intermediate Usage

### Shell-Specific Installation

#### Bash Completion

```bash
# Generate and install bash completion
portunix completion bash > /etc/bash_completion.d/portunix

# User-specific installation
portunix completion bash > ~/.bash_completion.d/portunix

# Add to .bashrc
echo 'source <(portunix completion bash)' >> ~/.bashrc

# Temporary activation
source <(portunix completion bash)
```

#### Zsh Completion

```bash
# System-wide installation
portunix completion zsh > /usr/local/share/zsh/site-functions/_portunix

# User-specific installation (Oh My Zsh)
portunix completion zsh > ~/.oh-my-zsh/completions/_portunix

# Manual fpath setup
mkdir -p ~/.zsh/completions
portunix completion zsh > ~/.zsh/completions/_portunix
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc

# Add to .zshrc
echo 'autoload -U compinit; compinit' >> ~/.zshrc
```

#### Fish Completion

```bash
# Install fish completion
portunix completion fish > ~/.config/fish/completions/portunix.fish

# System-wide installation
sudo portunix completion fish > /usr/share/fish/completions/portunix.fish

# Temporary activation
portunix completion fish | source
```

#### PowerShell Completion

```powershell
# Add to PowerShell profile
portunix completion powershell >> $PROFILE

# Current session only
portunix completion powershell | Invoke-Expression

# Save to file
portunix completion powershell > portunix-completion.ps1
```

### Advanced Completion Features

The completion system provides intelligent suggestions for:

#### Commands and Subcommands
```bash
portunix <TAB>
# Suggests: install, update, plugin, mcp, docker, virt, system, completion

portunix plugin <TAB>
# Suggests: list, install, enable, disable, start, stop, create, validate

portunix docker <TAB>
# Suggests: run, list, ssh, exec, stop, remove, logs, install
```

#### Package Names
```bash
portunix install <TAB>
# Suggests: nodejs, python, java, go, vscode, docker, git, maven, claude-code

portunix install python --variant <TAB>
# Suggests: full, minimal, default, latest
```

#### File and Directory Paths
```bash
portunix virt create myvm --iso <TAB>
# Suggests available ISO files

portunix plugin install <TAB>
# Suggests local plugin directories
```

#### Available Options
```bash
portunix install nodejs --<TAB>
# Suggests: --variant, --dry-run, --force, --version, --timeout

portunix system info --output <TAB>
# Suggests: text, json, yaml, csv, markdown
```

## Advanced Usage

### Custom Completion Configuration

Configure completion behavior:

```bash
# Enable verbose completion
export PORTUNIX_COMPLETION_VERBOSE=1

# Cache completion data
export PORTUNIX_COMPLETION_CACHE=1

# Custom completion timeout
export PORTUNIX_COMPLETION_TIMEOUT=5s

# Disable network-based completions
export PORTUNIX_COMPLETION_OFFLINE=1
```

### Dynamic Completion

The completion system provides dynamic suggestions:

#### Plugin Names
```bash
# Lists actually installed plugins
portunix plugin enable <TAB>

# Lists available plugins from registry
portunix plugin install <TAB>
```

#### Container Names
```bash
# Lists running containers
portunix docker ssh <TAB>

# Lists all containers
portunix docker remove <TAB>
```

#### VM Names
```bash
# Lists available VMs
portunix virt start <TAB>

# Lists running VMs
portunix virt stop <TAB>
```

### Completion Debugging

Debug completion issues:

```bash
# Debug completion generation
portunix completion bash --debug

# Verbose completion output
PORTUNIX_COMPLETION_DEBUG=1 portunix completion zsh

# Test completion directly
portunix __complete install node

# Trace completion execution
PORTUNIX_COMPLETION_TRACE=1 portunix completion
```

### Multiple Shell Support

Set up completion for multiple shells:

```bash
#!/bin/bash
# install-completions.sh

# Detect available shells and install completions
for shell in bash zsh fish; do
    if command -v $shell >/dev/null 2>&1; then
        echo "Installing $shell completion..."
        portunix completion $shell > ~/.portunix-completion-$shell

        case $shell in
            bash)
                echo "source ~/.portunix-completion-bash" >> ~/.bashrc
                ;;
            zsh)
                mkdir -p ~/.zsh/completions
                cp ~/.portunix-completion-zsh ~/.zsh/completions/_portunix
                ;;
            fish)
                mkdir -p ~/.config/fish/completions
                cp ~/.portunix-completion-fish ~/.config/fish/completions/portunix.fish
                ;;
        esac
    fi
done
```

### IDE Integration

#### VS Code Integration

```json
{
    "terminal.integrated.shellArgs.linux": [
        "-c", "source <(portunix completion bash) && exec bash"
    ],
    "terminal.integrated.shellArgs.osx": [
        "-c", "source <(portunix completion zsh) && exec zsh"
    ]
}
```

#### IntelliJ/PyCharm Integration

```bash
# Add to shell integration
echo 'source <(portunix completion bash)' >> ~/.bashrc
```

## Expert Tips & Tricks

### 1. Conditional Completion Loading

```bash
# Load completion only if portunix is available
if command -v portunix >/dev/null 2>&1; then
    source <(portunix completion bash)
fi
```

### 2. Custom Completion Functions

Extend completion with custom functions:

```bash
# Custom completion for frequently used commands
_portunix_dev_setup() {
    local commands="nodejs python java docker vscode git"
    COMPREPLY=($(compgen -W "$commands" -- "${COMP_WORDS[COMP_CWORD]}"))
}

# Register custom completion
complete -F _portunix_dev_setup portunix-dev-setup
```

### 3. Completion Performance Optimization

```bash
# Cache completion data
export PORTUNIX_COMPLETION_CACHE_DIR=~/.cache/portunix
export PORTUNIX_COMPLETION_CACHE_TTL=3600  # 1 hour

# Parallel completion loading
export PORTUNIX_COMPLETION_PARALLEL=1

# Limit completion suggestions
export PORTUNIX_COMPLETION_MAX_SUGGESTIONS=20
```

### 4. Cross-Platform Setup

```bash
# Universal completion installer
install_completion() {
    local shell_name=$(basename "$SHELL")

    case "$shell_name" in
        bash)
            portunix completion bash > ~/.bash_completion
            echo "source ~/.bash_completion" >> ~/.bashrc
            ;;
        zsh)
            portunix completion zsh > ~/.zsh_completion
            echo "source ~/.zsh_completion" >> ~/.zshrc
            ;;
        fish)
            mkdir -p ~/.config/fish/completions
            portunix completion fish > ~/.config/fish/completions/portunix.fish
            ;;
        *)
            echo "Unsupported shell: $shell_name"
            return 1
            ;;
    esac
}
```

### 5. Completion Testing

Test completion functionality:

```bash
# Test completion generation
portunix completion bash > /tmp/test-completion.bash
bash /tmp/test-completion.bash

# Test specific completions
compgen -W "install update plugin" -- "ins"

# Validate completion syntax
bash -n <(portunix completion bash)
zsh -n <(portunix completion zsh)
```

## Troubleshooting

### Common Issues

#### 1. Completion Not Working
```bash
# Check if completion is loaded
complete -p portunix

# Reload shell configuration
source ~/.bashrc  # or ~/.zshrc

# Verify completion file exists
ls -la ~/.bash_completion.d/portunix

# Test completion generation
portunix completion bash | head -10
```

#### 2. Slow Completion
```bash
# Enable completion caching
export PORTUNIX_COMPLETION_CACHE=1

# Reduce completion timeout
export PORTUNIX_COMPLETION_TIMEOUT=1s

# Disable network completions
export PORTUNIX_COMPLETION_OFFLINE=1
```

#### 3. Incomplete Suggestions
```bash
# Update completion cache
rm -rf ~/.cache/portunix/completion

# Reinstall completion
portunix completion bash > ~/.bash_completion.d/portunix
source ~/.bash_completion.d/portunix

# Check completion version
portunix completion --version
```

#### 4. Permission Issues
```bash
# Use user-specific installation
mkdir -p ~/.bash_completion.d
portunix completion bash > ~/.bash_completion.d/portunix

# Fix permissions
chmod +r ~/.bash_completion.d/portunix
```

### Debug Mode

```bash
# Enable debug output
export PORTUNIX_COMPLETION_DEBUG=1
portunix completion bash

# Trace completion execution
set -x
source <(portunix completion bash)
set +x

# Check completion variables
echo $BASH_COMPLETION_VERSINFO
echo $ZSH_VERSION
```

## Platform-Specific Installation

### Linux

```bash
# System-wide installation (requires sudo)
sudo portunix completion bash > /etc/bash_completion.d/portunix

# Distribution-specific paths
# Debian/Ubuntu
sudo portunix completion bash > /usr/share/bash-completion/completions/portunix

# RHEL/Fedora
sudo portunix completion bash > /usr/share/bash-completion/completions/portunix

# Arch Linux
sudo portunix completion bash > /usr/share/bash-completion/completions/portunix
```

### macOS

```bash
# Homebrew bash-completion
brew install bash-completion
portunix completion bash > $(brew --prefix)/etc/bash_completion.d/portunix

# Zsh (default in macOS Catalina+)
portunix completion zsh > /usr/local/share/zsh/site-functions/_portunix
```

### Windows

```powershell
# PowerShell (Windows PowerShell and PowerShell Core)
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\.portunix
portunix completion powershell > $env:USERPROFILE\.portunix\completion.ps1

# Add to profile
Add-Content $PROFILE ". $env:USERPROFILE\.portunix\completion.ps1"

# Windows Subsystem for Linux (WSL)
portunix completion bash > ~/.bash_completion.d/portunix
```

## API Integration

### Programmatic Completion

```go
// Go API for completion
package main

import (
    "fmt"
    "os/exec"
    "strings"
)

func getCompletions(args []string) []string {
    cmd := exec.Command("portunix", "__complete")
    cmd.Args = append(cmd.Args, args...)

    output, err := cmd.Output()
    if err != nil {
        return nil
    }

    return strings.Split(strings.TrimSpace(string(output)), "\n")
}
```

### Custom Completion Server

```bash
# Start completion server
portunix completion server --port 8080

# Query completions via HTTP
curl "http://localhost:8080/complete?args=install,node"
```

## Configuration Files

### Completion Configuration

```yaml
# ~/.portunix/completion.yaml
completion:
  cache:
    enabled: true
    ttl: 3600
    directory: ~/.cache/portunix/completion

  performance:
    timeout: 5s
    max_suggestions: 50
    parallel: true

  features:
    dynamic_suggestions: true
    file_completion: true
    network_completion: false

  shells:
    bash:
      version: 4.0
      features: ["programmable_completion"]
    zsh:
      version: 5.0
      features: ["completion_system"]
    fish:
      version: 3.0
      features: ["completions"]
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORTUNIX_COMPLETION_CACHE` | Enable completion caching | false |
| `PORTUNIX_COMPLETION_TIMEOUT` | Completion timeout | 5s |
| `PORTUNIX_COMPLETION_DEBUG` | Debug mode | false |
| `PORTUNIX_COMPLETION_OFFLINE` | Offline mode | false |
| `PORTUNIX_COMPLETION_MAX_SUGGESTIONS` | Max suggestions | 50 |

## Related Commands

- [`help`](help.md) - Command help system
- [`version`](version.md) - Version information
- [`config`](config.md) - Configuration management

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--debug` | boolean | `false` | Debug completion generation |
| `--no-cache` | boolean | `false` | Disable completion caching |
| `--timeout` | duration | `5s` | Completion timeout |
| `--format` | string | `shell` | Output format |
| `--version` | boolean | `false` | Show completion version |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Unsupported shell |
| 3 | Generation failed |
| 4 | Permission denied |

## Version History

- **v1.5.0** - Added PowerShell completion
- **v1.4.0** - Implemented dynamic completions
- **v1.3.0** - Added completion caching
- **v1.2.0** - Enhanced fish shell support
- **v1.1.0** - Added zsh completion
- **v1.0.0** - Initial bash completion