# Portunix System Command

## Quick Start

The `system` command provides comprehensive information about your system environment, hardware, software, and Portunix configuration.

### Simplest Usage
```bash
portunix system info
```

This displays a complete system overview including OS, hardware, and installed software.

### Basic Syntax
```bash
portunix system [subcommand] [options]
```

### Common Subcommands
- `info` - Complete system information
- `os` - Operating system details
- `hardware` - Hardware specifications
- `network` - Network configuration
- `software` - Installed software
- `env` - Environment variables

## Intermediate Usage

### System Information Overview

Get comprehensive system information:

```bash
portunix system info
```

Example output:
```
System Information
==================
Operating System:
  Name: Ubuntu
  Version: 22.04.3 LTS
  Kernel: 6.14.0-29-generic
  Architecture: x86_64
  Hostname: dev-machine
  Uptime: 5 days, 3:45:22

Hardware:
  CPU: Intel Core i7-9750H @ 2.60GHz (12 cores)
  Memory: 16.0 GB (8.5 GB used)
  Disk: 512 GB SSD (234 GB free)
  GPU: NVIDIA GeForce GTX 1650

Network:
  IP Address: 192.168.1.100
  Gateway: 192.168.1.1
  DNS: 8.8.8.8, 8.8.4.4

Portunix:
  Version: v1.5.14
  Installation: /usr/local/bin/portunix
  Config: ~/.portunix/config.json
  Plugins: 3 installed, 2 active
```

### Operating System Details

Detailed OS information:

```bash
portunix system os

# Specific OS information
portunix system os --field version
portunix system os --field kernel
portunix system os --field distribution
```

Output includes:
- Distribution name and version
- Kernel version
- System architecture
- Boot time
- System locale
- Time zone
- SELinux/AppArmor status

### Hardware Specifications

Detailed hardware information:

```bash
portunix system hardware

# Specific components
portunix system hardware --cpu
portunix system hardware --memory
portunix system hardware --disk
portunix system hardware --gpu
```

### Network Configuration

Network information and diagnostics:

```bash
portunix system network

# Network interfaces
portunix system network interfaces

# Routing table
portunix system network routes

# DNS configuration
portunix system network dns

# Port usage
portunix system network ports
```

### Software Inventory

List installed software:

```bash
# All detected software
portunix system software

# Specific categories
portunix system software --category development
portunix system software --category runtime
portunix system software --category tools

# Check specific software
portunix system software --check nodejs
portunix system software --check python
```

## Advanced Usage

### System Diagnostics

Comprehensive system diagnostics:

```bash
# Full diagnostic report
portunix system diagnose

# Performance diagnostics
portunix system diagnose --performance

# Security diagnostics
portunix system diagnose --security

# Compatibility check
portunix system diagnose --compatibility
```

### Resource Monitoring

Real-time resource monitoring:

```bash
# Live system metrics
portunix system monitor

# CPU monitoring
portunix system monitor --cpu --interval 1s

# Memory monitoring
portunix system monitor --memory --duration 60s

# Disk I/O monitoring
portunix system monitor --disk-io

# Network traffic
portunix system monitor --network
```

### System Benchmarks

Performance benchmarking:

```bash
# Run all benchmarks
portunix system benchmark

# CPU benchmark
portunix system benchmark --cpu

# Memory benchmark
portunix system benchmark --memory

# Disk benchmark
portunix system benchmark --disk

# Network benchmark
portunix system benchmark --network
```

### Environment Analysis

Environment variable management:

```bash
# Show all environment variables
portunix system env

# Show Portunix-specific variables
portunix system env --portunix

# Show PATH analysis
portunix system env --path

# Validate environment
portunix system env --validate

# Export environment
portunix system env --export > environment.env
```

### System Requirements Check

Verify system requirements for packages:

```bash
# Check requirements for specific package
portunix system requirements nodejs

# Check requirements for all packages
portunix system requirements --all

# Check requirements for profile
portunix system requirements --profile full
```

### Process Management

View Portunix-related processes:

```bash
# List Portunix processes
portunix system processes

# Show process tree
portunix system processes --tree

# Show resource usage
portunix system processes --resources

# Kill Portunix process
portunix system processes --kill <pid>
```

### System Profiles

Export and import system profiles:

```bash
# Export current system profile
portunix system profile export > system-profile.json

# Import system profile
portunix system profile import system-profile.json

# Compare profiles
portunix system profile compare profile1.json profile2.json

# Validate profile
portunix system profile validate system-profile.json
```

### Container Environment Detection

Detect if running in container:

```bash
# Check container environment
portunix system container-check

# Get container details
portunix system container-info
```

Output:
```
Container Detection:
  Running in Container: Yes
  Container Type: Docker
  Container ID: abc123def456
  Container Name: dev-container
  Host System: Ubuntu 22.04
```

### Virtualization Detection

Detect virtualization environment:

```bash
# Check virtualization
portunix system virt-check

# Get hypervisor details
portunix system virt-info
```

Output:
```
Virtualization Detection:
  Virtualized: Yes
  Hypervisor: VMware
  VM Type: Full virtualization
  CPU Virtualization: VT-x enabled
  Nested Virtualization: Supported
```

### System Capabilities

Check system capabilities:

```bash
# All capabilities
portunix system capabilities

# Docker support
portunix system capabilities --docker

# Kubernetes support
portunix system capabilities --kubernetes

# Virtualization support
portunix system capabilities --virtualization
```

## Expert Tips & Tricks

### 1. System Health Monitoring

```bash
# Create health report
portunix system health > health-report.txt

# Continuous health monitoring
portunix system health --watch --interval 5m

# Alert on issues
portunix system health --alert-webhook https://hooks.slack.com/xxx
```

### 2. Performance Tuning

```bash
# Analyze performance bottlenecks
portunix system analyze --performance

# Suggest optimizations
portunix system optimize --suggest

# Apply optimizations (with confirmation)
portunix system optimize --apply
```

### 3. Security Audit

```bash
# Security scan
portunix system security-scan

# Check for vulnerabilities
portunix system security-scan --cve

# Generate security report
portunix system security-scan --report security-audit.pdf
```

### 4. System Backup

```bash
# Backup system configuration
portunix system backup --config

# Backup Portunix data
portunix system backup --data

# Full system backup
portunix system backup --full --output backup.tar.gz
```

### 5. Remote System Information

```bash
# Get remote system info via SSH
portunix system info --remote user@host

# Compare local and remote systems
portunix system compare --remote user@host
```

## Output Formats

### JSON Output

```bash
# JSON format for parsing
portunix system info --output json

# Pretty JSON
portunix system info --output json --pretty

# Specific field in JSON
portunix system info --output json --field hardware.cpu
```

### YAML Output

```bash
# YAML format
portunix system info --output yaml

# Compact YAML
portunix system info --output yaml --compact
```

### CSV Output

```bash
# CSV for spreadsheets
portunix system software --output csv > software.csv

# Custom delimiter
portunix system software --output csv --delimiter ";"
```

### Markdown Output

```bash
# Markdown report
portunix system info --output markdown > system-report.md

# With tables
portunix system info --output markdown --tables
```

## Filtering and Queries

### Query Language

```bash
# Query specific fields
portunix system query "hardware.memory.total > 8GB"

# Complex queries
portunix system query "os.name == 'Ubuntu' AND hardware.cpu.cores >= 4"

# JSONPath queries
portunix system info --jsonpath "$.hardware.cpu.model"
```

### Filtering

```bash
# Filter output
portunix system info --filter hardware

# Multiple filters
portunix system info --filter hardware,network

# Exclude fields
portunix system info --exclude environment
```

## Integration with Other Commands

### Pre-installation Checks

```bash
# Check before installing
portunix system requirements docker && portunix install docker

# Verify compatibility
if portunix system compatible nodejs; then
    portunix install nodejs
fi
```

### Conditional Execution

```bash
# Install based on OS
OS=$(portunix system os --field name)
if [ "$OS" = "Ubuntu" ]; then
    portunix install apt-package
fi
```

## Troubleshooting

### Common Issues

#### 1. Permission Denied
```bash
# Some information requires elevated privileges
sudo portunix system hardware --detailed

# Or run specific commands
portunix system info --no-privileged
```

#### 2. Slow Information Gathering
```bash
# Skip slow checks
portunix system info --fast

# Timeout for checks
portunix system info --timeout 10s

# Parallel execution
portunix system info --parallel
```

#### 3. Missing Information
```bash
# Install required tools
portunix system deps --install

# Use alternative methods
portunix system info --fallback

# Verbose output for debugging
portunix system info -v
```

### Debug Mode

```bash
# Debug output
portunix system info --debug

# Trace system calls
portunix system info --trace

# Show command execution
portunix system info --show-commands
```

## Performance Considerations

### Caching

```bash
# Use cached information
portunix system info --cached

# Update cache
portunix system info --update-cache

# Clear cache
portunix system cache clear
```

### Selective Information

```bash
# Only basic info (fast)
portunix system info --basic

# Exclude slow operations
portunix system info --no-network --no-software

# Specific components only
portunix system hardware --cpu --memory
```

## API Integration

### REST API

```bash
# Get system info via API
curl http://localhost:8080/api/system/info

# Specific component
curl http://localhost:8080/api/system/hardware/cpu
```

### gRPC API

```go
// Go client example
client := portunix.NewClient()
info, err := client.System.GetInfo(context.Background(), &SystemInfoRequest{
    Components: []string{"hardware", "os"},
})
```

### WebSocket Monitoring

```javascript
// Real-time monitoring
const ws = new WebSocket('ws://localhost:8080/system/monitor');
ws.on('message', (data) => {
  const metrics = JSON.parse(data);
  console.log(`CPU: ${metrics.cpu}%, Memory: ${metrics.memory}%`);
});
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORTUNIX_SYSTEM_CACHE` | Cache duration | 5m |
| `PORTUNIX_SYSTEM_TIMEOUT` | Command timeout | 30s |
| `PORTUNIX_SYSTEM_PARALLEL` | Parallel execution | true |
| `PORTUNIX_SYSTEM_VERBOSE` | Verbose output | false |
| `PORTUNIX_SYSTEM_FALLBACK` | Use fallback methods | true |

## Platform-Specific Features

### Windows

```bash
# Windows-specific information
portunix system windows

# Windows features
portunix system windows features

# Windows updates
portunix system windows updates

# PowerShell version
portunix system windows powershell
```

### Linux

```bash
# Linux distribution details
portunix system linux distro

# Package managers
portunix system linux packages

# Systemd services
portunix system linux services

# Kernel modules
portunix system linux modules
```

### macOS

```bash
# macOS version details
portunix system macos version

# Homebrew packages
portunix system macos brew

# System extensions
portunix system macos extensions
```

## Related Commands

- [`install`](install.md) - Install packages
- [`update`](update.md) - Update Portunix
- [`config`](config.md) - Configuration management
- [`diagnose`](diagnose.md) - System diagnostics

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--output` | string | `text` | Output format (text/json/yaml/csv/markdown) |
| `--filter` | string | - | Filter output sections |
| `--exclude` | string | - | Exclude sections |
| `--cached` | boolean | `false` | Use cached information |
| `--timeout` | duration | `30s` | Operation timeout |
| `--parallel` | boolean | `true` | Parallel information gathering |
| `--verbose` | boolean | `false` | Verbose output |
| `--debug` | boolean | `false` | Debug mode |
| `--fast` | boolean | `false` | Skip slow operations |
| `--detailed` | boolean | `false` | Detailed information |
| `--jsonpath` | string | - | JSONPath query |
| `--field` | string | - | Specific field |
| `--pretty` | boolean | `false` | Pretty print output |
| `--no-color` | boolean | `false` | Disable colored output |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Permission denied |
| 3 | Timeout |
| 4 | Unsupported platform |
| 5 | Missing dependencies |
| 6 | Invalid query |
| 7 | Cache error |
| 8 | Network error |

## Version History

- **v1.5.0** - Added real-time monitoring
- **v1.4.0** - Implemented system profiles
- **v1.3.0** - Added container detection
- **v1.2.0** - Performance improvements
- **v1.1.0** - Added JSON/YAML output
- **v1.0.0** - Initial system information