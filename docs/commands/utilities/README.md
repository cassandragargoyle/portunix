# Utility Commands

Essential utility commands that enhance the Portunix user experience through shell integration, help systems, and productivity tools.

## Commands in this Category

### [`completion`](completion.md) - Shell Completion System
Advanced shell completion system providing intelligent tab completion for all Portunix commands across multiple shells.

**Quick Examples:**
```bash
# Install bash completion
portunix completion bash > ~/.bash_completion.d/portunix

# Install zsh completion
portunix completion zsh > ~/.zsh/completions/_portunix

# Install fish completion
portunix completion fish > ~/.config/fish/completions/portunix.fish
```

**Key Features:**
- Multi-shell support (bash, zsh, fish, PowerShell)
- Dynamic completion suggestions
- Context-aware parameter completion
- Performance optimization with caching
- Intelligent file path completion
- Package name and plugin completion

**Common Use Cases:**
- Enhanced command-line productivity
- Reduced typing and syntax errors
- Discovery of available commands and options
- Faster navigation through Portunix features

---

### `help` - Advanced Help System *(Coming Soon)*
Comprehensive help system with context-aware assistance and interactive guides.

**Planned Features:**
- Interactive help with examples
- Context-sensitive assistance
- Progressive disclosure for beginners
- Expert-level detailed documentation
- AI-powered help suggestions

---

### `version` - Version Information *(Coming Soon)*
Detailed version information and compatibility checking.

**Planned Features:**
- Semantic version display
- Compatibility matrix
- Update information
- Component version details

---

### `config` - Configuration Management *(Coming Soon)*
Global configuration management for Portunix settings and preferences.

**Planned Features:**
- Global settings management
- Profile-based configurations
- Import/export configurations
- Validation and schema checking

## Category Overview

The **Utilities** category contains commands that enhance the overall user experience and productivity when working with Portunix. These tools focus on making the command-line interface more efficient, discoverable, and user-friendly.

### Shell Integration Excellence

#### Multi-Shell Completion Support

Portunix provides sophisticated completion across all major shells:

```bash
# Bash (most common)
portunix completion bash > /etc/bash_completion.d/portunix

# Zsh (macOS default, popular among developers)
portunix completion zsh > /usr/local/share/zsh/site-functions/_portunix

# Fish (modern shell with great features)
portunix completion fish > ~/.config/fish/completions/portunix.fish

# PowerShell (Windows and cross-platform)
portunix completion powershell >> $PROFILE
```

#### Intelligent Completion Features

The completion system goes beyond basic command completion:

##### Dynamic Package Suggestions
```bash
portunix install <TAB>
# Suggests: nodejs, python, java, go, docker, vscode, git...

portunix install python --variant <TAB>
# Suggests: full, minimal, default, latest
```

##### Context-Aware Container Completion
```bash
portunix docker ssh <TAB>
# Shows only running containers

portunix docker logs <TAB>
# Shows all containers (running and stopped)
```

##### Plugin Discovery
```bash
portunix plugin install <TAB>
# Shows available plugins from registry and local directories

portunix plugin enable <TAB>
# Shows only installed but disabled plugins
```

#### File and Path Completion
```bash
portunix virt create myvm --iso <TAB>
# Suggests available ISO files in common locations

portunix docker run ubuntu --volumes <TAB>
# Intelligent path completion for volume mounting
```

## Advanced Completion Features

### Performance Optimization

```bash
# Enable completion caching for better performance
export PORTUNIX_COMPLETION_CACHE=1
export PORTUNIX_COMPLETION_CACHE_TTL=3600  # 1 hour

# Parallel completion loading
export PORTUNIX_COMPLETION_PARALLEL=1

# Limit suggestions for performance
export PORTUNIX_COMPLETION_MAX_SUGGESTIONS=20
```

### Custom Completion Extensions

```bash
# Add custom completion functions
_portunix_custom_dev_setup() {
    local dev_stacks="frontend backend fullstack mobile devops"
    COMPREPLY=($(compgen -W "$dev_stacks" -- "${COMP_WORDS[COMP_CWORD]}"))
}

# Register custom completion
complete -F _portunix_custom_dev_setup portunix-dev-setup
```

### Cross-Platform Installation

#### Universal Installation Script
```bash
#!/bin/bash
# install-completions.sh - Universal completion installer

detect_shell() {
    case "$(basename "$SHELL")" in
        bash) echo "bash" ;;
        zsh) echo "zsh" ;;
        fish) echo "fish" ;;
        *) echo "unknown" ;;
    esac
}

install_completion() {
    local shell_type=$(detect_shell)

    case "$shell_type" in
        bash)
            portunix completion bash > ~/.bash_completion.d/portunix
            echo "source ~/.bash_completion.d/portunix" >> ~/.bashrc
            ;;
        zsh)
            mkdir -p ~/.zsh/completions
            portunix completion zsh > ~/.zsh/completions/_portunix
            echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
            echo 'autoload -U compinit; compinit' >> ~/.zshrc
            ;;
        fish)
            mkdir -p ~/.config/fish/completions
            portunix completion fish > ~/.config/fish/completions/portunix.fish
            ;;
        *)
            echo "Unsupported shell: $SHELL"
            exit 1
            ;;
    esac

    echo "Completion installed for $shell_type"
    echo "Please restart your shell or source your configuration file"
}

install_completion
```

## Integration with Development Workflows

### IDE Integration

#### VS Code Integration
```json
{
    "terminal.integrated.shellArgs.linux": [
        "-c", "source ~/.bash_completion.d/portunix && exec bash"
    ],
    "terminal.integrated.shellArgs.osx": [
        "-c", "source ~/.zsh/completions/_portunix && exec zsh"
    ]
}
```

#### IntelliJ/PyCharm Integration
```bash
# Add to shell integration in IDE settings
source ~/.bash_completion.d/portunix
```

### CI/CD Integration

#### Automated Completion Setup
```yaml
# .github/workflows/setup-dev-env.yml
name: Setup Development Environment
on: [push]

jobs:
  setup:
    runs-on: ubuntu-latest
    steps:
      - name: Install Portunix
        run: curl -sSL https://install.portunix.ai | bash

      - name: Setup Shell Completion
        run: |
          portunix completion bash > ~/.bash_completion.d/portunix
          echo "source ~/.bash_completion.d/portunix" >> ~/.bashrc

      - name: Test Completion
        run: |
          source ~/.bash_completion.d/portunix
          # Test completion functionality
```

### Team Productivity Features

#### Shared Completion Configurations
```bash
# Team-wide completion setup
# team-setup.sh
#!/bin/bash

# Install common completions for team
portunix completion bash > /etc/bash_completion.d/portunix

# Add team-specific custom completions
cat >> /etc/bash_completion.d/portunix-team << 'EOF'
# Team-specific project shortcuts
_portunix_team_projects() {
    local projects="webapp mobile-app api-gateway data-pipeline"
    COMPREPLY=($(compgen -W "$projects" -- "${COMP_WORDS[COMP_CWORD]}"))
}

complete -F _portunix_team_projects project-setup
EOF
```

## Advanced Utility Features

### Completion Debugging and Testing

```bash
# Debug completion generation
portunix completion bash --debug

# Test specific completions
compgen -W "install update plugin" -- "ins"

# Validate completion syntax
bash -n <(portunix completion bash)

# Performance testing
time portunix completion bash > /dev/null
```

### Custom Utility Development

```bash
# Create custom utility script
cat > ~/.local/bin/portunix-dev-shortcuts << 'EOF'
#!/bin/bash
# Development shortcuts using Portunix

case "$1" in
    "setup-node")
        portunix install nodejs
        portunix docker run-in-container nodejs --name node-dev
        ;;
    "setup-python")
        portunix install python --variant full
        portunix virt create python-dev --template python-dev
        ;;
    "team-env")
        portunix install default
        portunix plugin install agile-software-development
        portunix mcp configure
        ;;
    *)
        echo "Usage: $0 {setup-node|setup-python|team-env}"
        exit 1
        ;;
esac
EOF

chmod +x ~/.local/bin/portunix-dev-shortcuts
```

## Platform-Specific Features

### Windows PowerShell Integration

```powershell
# Enhanced PowerShell integration
# Microsoft.PowerShell_profile.ps1

# Load Portunix completion
& portunix completion powershell | Invoke-Expression

# Custom PowerShell functions
function PortunixDevSetup {
    param(
        [Parameter(Mandatory=$true)]
        [ValidateSet("node", "python", "java", "full")]
        [string]$Stack
    )

    switch ($Stack) {
        "node" { portunix install nodejs }
        "python" { portunix install python --variant full }
        "java" { portunix install java }
        "full" { portunix install default }
    }
}

# Register argument completer
Register-ArgumentCompleter -CommandName PortunixDevSetup -ParameterName Stack -ScriptBlock {
    param($commandName, $parameterName, $wordToComplete, $commandAst, $fakeBoundParameters)
    @("node", "python", "java", "full") | Where-Object { $_ -like "$wordToComplete*" }
}
```

### Linux Distribution-Specific Setup

#### Ubuntu/Debian
```bash
# System-wide installation
sudo portunix completion bash > /usr/share/bash-completion/completions/portunix

# Package manager integration
echo 'if command -v portunix >/dev/null 2>&1; then
  source /usr/share/bash-completion/completions/portunix
fi' | sudo tee /etc/bash_completion.d/portunix-load
```

#### RHEL/CentOS/Fedora
```bash
# RHEL family installation
sudo portunix completion bash > /usr/share/bash-completion/completions/portunix

# SELinux context setup
sudo restorecon -R /usr/share/bash-completion/completions/
```

#### Arch Linux
```bash
# Arch Linux installation
sudo portunix completion bash > /usr/share/bash-completion/completions/portunix

# AUR package integration
mkdir -p ~/.config/pacman
echo "Include = /etc/pacman.d/portunix-completion.conf" >> ~/.config/pacman/makepkg.conf
```

### macOS-Specific Features

```bash
# Homebrew integration
brew install bash-completion
portunix completion bash > $(brew --prefix)/etc/bash_completion.d/portunix

# Zsh on macOS (default since Catalina)
portunix completion zsh > /usr/local/share/zsh/site-functions/_portunix

# iTerm2 integration
echo 'source <(portunix completion zsh)' >> ~/.zshrc
```

## Performance Monitoring and Optimization

### Completion Performance Analysis

```bash
# Measure completion performance
time_completion() {
    local cmd="$1"
    local iterations=${2:-100}

    echo "Testing completion performance for: $cmd"

    for i in $(seq 1 $iterations); do
        time portunix __complete $cmd 2>/dev/null >/dev/null
    done | grep real | awk '{sum+=$2} END {print "Average:", sum/NR, "seconds"}'
}

# Test various completion scenarios
time_completion "install node"
time_completion "docker ssh"
time_completion "plugin install"
```

### Memory Usage Optimization

```bash
# Monitor completion memory usage
measure_completion_memory() {
    local pid
    portunix __complete install & pid=$!

    while kill -0 $pid 2>/dev/null; do
        ps -o pid,vsz,rss,comm -p $pid
        sleep 0.1
    done
}
```

## Troubleshooting Utilities

### Completion Diagnostics

```bash
# Comprehensive completion diagnostics
portunix completion diagnose

# Test shell compatibility
portunix completion test --shell bash
portunix completion test --shell zsh
portunix completion test --shell fish

# Verify installation
portunix completion verify --installation

# Clean and reinstall
portunix completion clean
portunix completion install --shell auto
```

### Common Issues and Solutions

#### Completion Not Working
```bash
# Check if completion is loaded
complete -p portunix

# Reload completion
source ~/.bash_completion.d/portunix

# Verify completion file exists and is readable
ls -la ~/.bash_completion.d/portunix
```

#### Slow Completion Performance
```bash
# Enable caching
export PORTUNIX_COMPLETION_CACHE=1

# Reduce completion timeout
export PORTUNIX_COMPLETION_TIMEOUT=1s

# Disable network-based completions
export PORTUNIX_COMPLETION_OFFLINE=1
```

## Future Roadmap

### Planned Utility Enhancements

#### Interactive Help System
- Context-aware help with examples
- Progressive help levels (beginner to expert)
- Interactive tutorials and walkthroughs
- AI-powered help suggestions

#### Configuration Management
- Global configuration system
- Profile-based settings
- Team configuration sharing
- Configuration validation and migration

#### Advanced Version Management
- Detailed version information
- Compatibility checking
- Update recommendations
- Component version tracking

### Integration Improvements

#### Enhanced Shell Integration
- Real-time command suggestions
- Syntax highlighting for Portunix commands
- Command history analysis and optimization
- Predictive command completion

#### IDE and Editor Plugins
- VS Code extension with IntelliSense
- Vim/Neovim plugin integration
- Emacs mode for Portunix
- JetBrains IDE integration

#### Mobile and Web Interfaces
- Web-based command interface
- Mobile app for system monitoring
- Progressive Web App (PWA) support
- Voice command integration

## Best Practices

### Completion Setup
- Install completion for all team members
- Use system-wide installation where appropriate
- Regular updates and maintenance
- Test completion after shell updates

### Performance Optimization
- Enable caching for better performance
- Monitor and optimize completion times
- Use appropriate timeout settings
- Regular cleanup of completion cache

### Team Productivity
- Standardize completion setup across team
- Create team-specific custom completions
- Document completion features and shortcuts
- Regular training on productivity features

## Related Categories

- **[Core](../core/)** - Commands enhanced by completion
- **[Integration](../integration/)** - MCP tools with completion
- **[Plugins](../plugins/)** - Plugin commands with completion
- **[Containers](../containers/)** - Container management with completion

---

*Utility commands transform the command-line experience from functional to delightful, making Portunix accessible to users of all skill levels.*