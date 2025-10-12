# Portunix Update Command

## Quick Start

The `update` command keeps Portunix itself up-to-date by downloading and installing the latest version from GitHub releases.

### Simplest Usage
```bash
portunix update
```

This checks for updates and installs the latest version if available.

### Basic Syntax
```bash
portunix update [options]
```

### Common Operations
- Check for updates without installing: `portunix update --check`
- Force reinstall current version: `portunix update --force`
- Update to specific version: `portunix update --version v1.5.14`

## Intermediate Usage

### Check for Updates Only

To see if updates are available without installing:

```bash
portunix update --check
```

Output example:
```
Current version: v1.5.13
Latest version: v1.5.14
Update available: Yes
Release date: 2024-01-15
Download size: 15.2 MB
```

### Force Update

Force reinstallation of the current version (useful for fixing corrupted installations):

```bash
portunix update --force
```

### Update Process Overview

1. **Version Check** - Queries GitHub API for latest release
2. **Download** - Downloads new binary and checksums
3. **Verification** - Verifies SHA256 checksum
4. **Backup** - Creates backup of current binary
5. **Replace** - Replaces old binary with new one
6. **Validation** - Runs self-test on new binary
7. **Cleanup** - Removes temporary files

### Rollback Mechanism

If update fails, Portunix automatically:
- Restores from backup
- Verifies restored binary works
- Reports rollback status

Manual rollback:
```bash
# List available backups
portunix update --list-backups

# Restore specific backup
portunix update --rollback v1.5.13
```

## Advanced Usage

### Update Architecture

The update system implements a secure, atomic update process:

```
┌─────────────────┐
│ GitHub Releases │
└────────┬────────┘
         │ API Query
         ▼
┌─────────────────┐
│ Version Compare │
└────────┬────────┘
         │ If newer
         ▼
┌─────────────────┐
│ Download Binary │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Verify Checksum │
└────────┬────────┘
         │ If valid
         ▼
┌─────────────────┐
│  Create Backup  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Atomic Replace │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Self-Test     │
└────────┬────────┘
         │ Pass/Fail
         ▼
┌─────────────────┐
│ Commit/Rollback │
└─────────────────┘
```

### Update Channels

```bash
# Stable channel (default)
portunix update --channel stable

# Beta channel
portunix update --channel beta

# Nightly builds
portunix update --channel nightly

# Release candidates
portunix update --channel rc
```

### Custom Update Sources

```bash
# Use custom GitHub repository
portunix update --repo mycompany/portunix-fork

# Use private registry
portunix update --registry https://registry.company.com

# Use local file
portunix update --local /path/to/portunix-v1.5.14.tar.gz
```

### Update Configuration

Configure update behavior in `~/.portunix/config.json`:

```json
{
  "update": {
    "channel": "stable",
    "autoCheck": true,
    "checkInterval": "24h",
    "autoInstall": false,
    "keepBackups": 5,
    "verifySignature": true,
    "proxy": "http://proxy.company.com:8080"
  }
}
```

### Automated Updates

#### Cron Job (Linux/macOS)
```bash
# Add to crontab
0 3 * * * /usr/local/bin/portunix update --silent --auto-confirm
```

#### Task Scheduler (Windows)
```powershell
# Create scheduled task
$action = New-ScheduledTaskAction -Execute "portunix.exe" -Argument "update --silent --auto-confirm"
$trigger = New-ScheduledTaskTrigger -Daily -At 3am
Register-ScheduledTask -TaskName "PortunixUpdate" -Action $action -Trigger $trigger
```

### Update Verification

The update process includes multiple verification steps:

1. **Checksum Verification**
   - SHA256 hash comparison
   - Prevents corrupted downloads
   - Ensures file integrity

2. **Signature Verification**
   ```bash
   # Enable GPG signature verification
   portunix update --verify-signature

   # Import Portunix public key
   portunix update --import-key https://github.com/portunix.gpg
   ```

3. **Self-Test Validation**
   - Runs `--version` on new binary
   - Validates core functionality
   - Checks plugin compatibility

### Update Hooks

Execute custom scripts during update:

```bash
# Pre-update hook
export PORTUNIX_PRE_UPDATE="echo 'Starting update...'"

# Post-update hook
export PORTUNIX_POST_UPDATE="echo 'Update completed!'"

portunix update
```

### Differential Updates

For large deployments, use differential updates:

```bash
# Download only changed files
portunix update --differential

# Binary patching (smaller downloads)
portunix update --patch

# Resume interrupted download
portunix update --resume
```

### Network Configuration

```bash
# Use proxy
export HTTP_PROXY=http://proxy:8080
export HTTPS_PROXY=http://proxy:8080
portunix update

# Bypass proxy for GitHub
export NO_PROXY=github.com,api.github.com
portunix update

# Custom timeout
portunix update --timeout 60s

# Retry configuration
portunix update --retries 5 --retry-delay 10s
```

## Expert Tips & Tricks

### 1. Update in Isolated Environment

```bash
# Test update in container first
portunix docker run ubuntu
portunix docker exec container portunix update --dry-run
```

### 2. Batch Updates Across Systems

```bash
#!/bin/bash
# Update multiple systems
for host in server1 server2 server3; do
    ssh $host "portunix update --silent"
done
```

### 3. Update with Validation Script

```bash
# Custom validation after update
portunix update --validate-script "portunix --version && portunix plugin list"
```

### 4. Progressive Rollout

```bash
# Update percentage of fleet
portunix update --rollout-percentage 10

# Canary deployment
portunix update --canary --canary-duration 24h
```

### 5. Update Metrics and Monitoring

```bash
# Export update metrics
portunix update --metrics-export prometheus

# Send update events
portunix update --notify-webhook https://hooks.slack.com/xxx
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Permission Denied
```bash
# Linux/macOS
sudo portunix update

# Windows - run as Administrator
# Or install to user directory
portunix update --install-path ~/bin/portunix
```

#### 2. Network Connection Failed
```bash
# Check connectivity
curl -I https://api.github.com

# Use alternative download method
portunix update --download-method wget

# Use mirror
portunix update --mirror https://mirror.company.com
```

#### 3. Checksum Mismatch
```bash
# Re-download with cache clear
portunix update --clear-cache

# Skip checksum (NOT recommended)
portunix update --skip-checksum

# Verify manually
sha256sum portunix-linux-amd64.tar.gz
```

#### 4. Update Loop
```bash
# Clear update state
rm ~/.portunix/update.state

# Force specific version
portunix update --version v1.5.13 --force
```

### Debug Mode

```bash
# Verbose output
portunix update -v

# Debug logging
portunix update --debug

# Dry run (simulate update)
portunix update --dry-run
```

### Recovery Options

```bash
# List all backups
ls -la ~/.portunix/backups/

# Manual restore
cp ~/.portunix/backups/portunix-v1.5.13 /usr/local/bin/portunix

# Download specific version manually
wget https://github.com/cassandragargoyle/portunix/releases/download/v1.5.13/portunix-linux-amd64.tar.gz
```

## Platform-Specific Considerations

### Windows

```bash
# Update with Windows Defender exclusion
portunix update --windows-defender-exclude

# Update Windows service
portunix update --update-service
```

### Linux

```bash
# Update with systemd service restart
portunix update --restart-service portunix.service

# Update SELinux contexts
portunix update --selinux-restore
```

### macOS

```bash
# Update with Gatekeeper bypass
portunix update --gatekeeper-bypass

# Update Launch Agent
portunix update --update-launch-agent
```

## Security Best Practices

### 1. Always Verify Signatures
```bash
# Configure mandatory signature verification
portunix config set update.requireSignature true
```

### 2. Use Secure Channels
```bash
# Force HTTPS only
portunix config set update.httpsOnly true
```

### 3. Audit Update History
```bash
# View update log
portunix update --show-history

# Export audit log
portunix update --export-audit /path/to/audit.log
```

### 4. Restricted Update Window
```bash
# Only allow updates during maintenance window
portunix update --window "02:00-04:00"
```

## API Integration

### REST API
```bash
# Trigger update via API
curl -X POST http://localhost:8080/api/update \
  -H "Content-Type: application/json" \
  -d '{"version": "latest", "force": false}'
```

### gRPC API
```go
// Go client example
client := portunix.NewClient()
result, err := client.Update(context.Background(), &UpdateRequest{
    Version: "latest",
    Force:   false,
    DryRun:  false,
})
```

### Webhook Notifications
```json
{
  "event": "update.completed",
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "previousVersion": "v1.5.13",
    "newVersion": "v1.5.14",
    "duration": "45s",
    "status": "success"
  }
}
```

## Performance Optimization

### Download Optimization
- **Parallel chunk download**: Splits file into chunks
- **Resume support**: Continues interrupted downloads
- **Compression**: Uses gzip compression
- **CDN support**: Automatically selects fastest mirror

### Update Speed Metrics
- Average download speed: 10-50 MB/s
- Checksum verification: <1 second
- Binary replacement: <100ms
- Total update time: typically <1 minute

## Integration with CI/CD

### GitHub Actions
```yaml
- name: Update Portunix
  run: |
    portunix update --check
    if [ $? -eq 0 ]; then
      portunix update --auto-confirm
    fi
```

### Jenkins Pipeline
```groovy
stage('Update Portunix') {
    steps {
        sh 'portunix update --silent --auto-confirm'
    }
}
```

### GitLab CI
```yaml
update-portunix:
  script:
    - portunix update --check || true
    - portunix update --auto-confirm
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORTUNIX_UPDATE_URL` | Custom update server | GitHub API |
| `PORTUNIX_UPDATE_CHANNEL` | Update channel | stable |
| `PORTUNIX_UPDATE_PROXY` | Proxy server | - |
| `PORTUNIX_NO_UPDATE` | Disable updates | false |
| `PORTUNIX_UPDATE_TIMEOUT` | Update timeout | 5m |
| `PORTUNIX_BACKUP_DIR` | Backup directory | ~/.portunix/backups |

## Related Commands

- [`install`](install.md) - Install packages
- [`plugin update`](plugin.md#update) - Update plugins
- [`system`](system.md) - System information
- [`version`](version.md) - Version information

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--check` | boolean | `false` | Check for updates only |
| `--force` | boolean | `false` | Force reinstall current version |
| `--version` | string | `latest` | Specific version to install |
| `--channel` | string | `stable` | Update channel |
| `--dry-run` | boolean | `false` | Simulate update |
| `--auto-confirm` | boolean | `false` | Skip confirmation prompts |
| `--silent` | boolean | `false` | Silent mode |
| `--rollback` | string | - | Rollback to specific version |
| `--list-backups` | boolean | `false` | List available backups |
| `--verify-signature` | boolean | `false` | Verify GPG signature |
| `--skip-checksum` | boolean | `false` | Skip checksum verification |
| `--timeout` | duration | `5m` | Update timeout |
| `--retries` | int | `3` | Number of retries |
| `--proxy` | string | - | Proxy server |
| `--mirror` | string | - | Mirror URL |
| `--differential` | boolean | `false` | Use differential updates |
| `--output` | string | `text` | Output format (text/json/yaml) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successful update or no update needed |
| 1 | General error |
| 2 | Network error |
| 3 | Checksum verification failed |
| 4 | Permission denied |
| 5 | Rollback performed |
| 6 | Update already in progress |
| 7 | Incompatible version |
| 8 | Signature verification failed |
| 9 | Backup creation failed |
| 10 | Self-test failed |

## Version History

- **v1.5.14** - Added differential updates
- **v1.5.0** - Implemented signature verification
- **v1.4.0** - Added automatic rollback
- **v1.3.0** - Introduced update channels
- **v1.2.0** - Added checksum verification
- **v1.1.0** - Improved error handling
- **v1.0.0** - Initial update functionality