# VM SSH Deployment Guide

This guide describes how to deploy and test Portunix on remote Windows/Linux VMs via SSH.

## Prerequisites

- SSH client installed on development machine
- Target VM with SSH server running
- Network connectivity to VM

## Setting Up SSH on Windows VM

### Using Portunix (Recommended)

Portunix provides three installation variants for OpenSSH:

| Variant | Description | Requires Admin |
|---------|-------------|----------------|
| `client` | SSH client only (default) | No |
| `server` | SSH client + server | Yes |
| `portable` | Portable installation, no system changes | No |

#### Install SSH Server (for remote access)

```powershell
# Run PowerShell as Administrator
portunix.exe install openssh --variant server
```

This will:

- Download Win32-OpenSSH from GitHub
- Extract to `C:\Portunix\OpenSSH\`
- Install sshd and ssh-agent Windows services
- Add firewall rule for port 22
- Start sshd service and set to auto-start

#### Install SSH Client Only

```powershell
# No admin required
portunix.exe install openssh --variant client
```

Installs to `%LOCALAPPDATA%\Programs\OpenSSH\`

#### Portable Installation (no system changes)

```powershell
# No admin required, no services installed
portunix.exe install openssh --variant portable
```

Installs to `C:\Portunix\bin\openssh-portable\`

#### Preview Installation (Dry Run)

```powershell
# See what would be installed without making changes
portunix.exe install openssh --variant server --dry-run
```

#### Verify Installation

```powershell
# Check SSH version
ssh -V

# Check sshd service status
Get-Service sshd

# Test local connection
ssh localhost
```

### Manual Setup (without Portunix)

```powershell
# Install OpenSSH Server via Windows Optional Features
Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0

# Start and enable service
Start-Service sshd
Set-Service -Name sshd -StartupType Automatic

# Add firewall rule
New-NetFirewallRule -Name sshd -DisplayName 'OpenSSH Server' -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22
```

## Deployment Commands

### Test Connection

```bash
# Test SSH connection
ssh user@VM_IP_ADDRESS

# Test with specific command
ssh user@VM_IP_ADDRESS "echo Hello from VM"
```

### Deploy Single File

```bash
# Upload single file
scp /path/to/file user@VM_IP_ADDRESS:"C:/Destination/Path/"

# Example: Upload portunix.exe
scp dist/platforms/windows-amd64/portunix.exe demo@192.168.53.66:"C:/Portunix/"
```

### Deploy All Binaries

```bash
# Upload all Windows binaries
scp dist/platforms/windows-amd64/*.exe user@VM_IP_ADDRESS:"C:/Portunix/"
```

### Deploy from ZIP

```bash
# Upload ZIP and extract on VM
scp dist/platforms/windows-amd64.zip user@VM_IP_ADDRESS:"C:/Temp/"
ssh user@VM_IP_ADDRESS "powershell Expand-Archive -Path C:\Temp\windows-amd64.zip -DestinationPath C:\Portunix -Force"
```

## Remote Command Execution

### Run Portunix Commands

```bash
# Check version
ssh user@VM_IP_ADDRESS "C:\Portunix\portunix.exe version"

# List packages
ssh user@VM_IP_ADDRESS "C:\Portunix\portunix.exe package list"

# Install package
ssh user@VM_IP_ADDRESS "C:\Portunix\portunix.exe install nodejs"
```

### Run as Administrator

For commands requiring admin privileges, use PowerShell remoting or RDP.

SSH alone cannot elevate privileges on Windows.

## Useful SSH Options

```bash
# Skip host key verification (for testing only)
ssh -o StrictHostKeyChecking=no user@VM_IP

# Set connection timeout
ssh -o ConnectTimeout=10 user@VM_IP

# Use specific port
ssh -p 2222 user@VM_IP

# Verbose output for debugging
ssh -v user@VM_IP
```

## Troubleshooting

### Connection Refused

```bash
# Check if SSH service is running on VM
ssh user@VM_IP "Get-Service sshd"

# Check firewall
ssh user@VM_IP "netsh advfirewall firewall show rule name=sshd"
```

### Permission Denied

- Verify username and password
- Check if user has SSH access rights
- On Windows, ensure user is in "Allowed Users" for SSH

### Path Issues with SCP

Windows paths in SCP:
```bash
# Use forward slashes
scp file.txt user@VM:"C:/Path/To/Dest/"

# Or escape backslashes
scp file.txt user@VM:"C:\\Path\\To\\Dest\\"
```

## Automation Script Example

```bash
#!/bin/bash
# deploy-to-vm.sh

VM_USER="demo"
VM_HOST="192.168.53.66"
VM_PATH="C:/Portunix"

# Build
make build-all-platforms

# Deploy
echo "Deploying to $VM_USER@$VM_HOST..."
scp dist/platforms/windows-amd64/*.exe "$VM_USER@$VM_HOST:\"$VM_PATH/\""

# Verify
ssh "$VM_USER@$VM_HOST" "$VM_PATH/portunix.exe version"

echo "Deployment complete!"
```

## Security Notes

- Use SSH keys instead of passwords for automated deployments
- Disable password authentication in production
- Keep SSH server updated
- Limit SSH access to specific users/IPs

---

**Created**: 2026-01-04
**Related Issues**: #127 (OpenSSH Installation)
