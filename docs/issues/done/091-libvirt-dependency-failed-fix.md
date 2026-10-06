# Issue #91: Libvirt Dependency Failed - Root Cause Analysis and Fix

## 🎯 Priority
**HIGH** - Critical for virt-manager functionality

## 📋 Status
- **Created**: 2025-10-04
- **Status**: 🔴 Open
- **Assignee**: TBD
- **Branch**: `feature/issue-091-libvirt-dependency-failed`

## 📝 Problem Description

### Current Situation
After implementing socket failure detection (#090), we can now detect that libvirt socket is in failed state, but the underlying root cause is still unresolved.

**Symptoms:**
```
📊 Libvirt Status:
   Version: 11.0.0
   Daemon type: monolithic
   Daemon name: libvirtd
   Running: false
   Enabled: true
   Masked: false
   Socket: libvirtd.socket
   Socket activated: false
   Socket state: ❌ FAILED (Result=trigger-limit-hit
ActiveState=failed
SubState=failed)
```

**System logs show:**
```
říj 04 17:59:50 black-deamon systemd[1]: Dependency failed for libvirtd.service - libvirt legacy monolithic daemon.
říj 04 17:59:50 black-deamon systemd[1]: libvirtd.service: Job libvirtd.service/start failed with result 'dependency'.
```

This error repeats in crash loop, triggering systemd's trigger-limit-hit protection.

### Impact
- Virtual Machine Manager (virt-manager) cannot connect to libvirt
- QEMU/KVM VMs cannot be managed
- Socket activation fails due to dependency issues
- Users cannot use virtualization despite having libvirt installed

### Related Issues
- **#090**: Libvirt Daemon Detection and Auto-Fix ✅ (implemented socket failure detection)
- **#088**: VirtualBox/KVM Conflict Detection ✅ (implemented)
- **#089**: QEMU/KVM Adapter Implementation ✅ (implemented)

## 🔍 Root Cause Analysis Required

### What We Know
1. ✅ Socket detection works - we can detect the failed state
2. ✅ Basic libvirt is installed (virsh --version returns 11.0.0)
3. ❌ Some systemd dependency is missing or failing
4. ❌ Daemon crashes immediately when socket tries to start it

### What We Need to Find
1. **Which specific dependency is failing?**
   - Check: `systemctl show libvirtd.service -p Requires -p Wants -p After`
   - Likely candidates: virtlogd, virtlockd, dbus services

2. **Why is the dependency failing?**
   - Missing package?
   - Configuration error?
   - Permission issue?

3. **What is the correct fix?**
   - Install missing packages?
   - Enable missing services?
   - Fix configuration?

## 🎯 Acceptance Criteria

### Detection
- [ ] Detect specific failed dependency (not just "dependency failed")
- [ ] Identify which systemd unit is causing the failure
- [ ] Report missing or broken dependencies clearly

### Fix Implementation
- [ ] Implement automatic detection of missing libvirt dependencies
- [ ] Add fix function to enable/start required dependencies
- [ ] Update `--fix-libvirt` to handle dependency failures
- [ ] Provide clear user messages about what's being fixed

### User Experience
```bash
$ ./portunix virt check --fix-libvirt

📊 Libvirt Status:
   Version: 11.0.0
   Daemon type: monolithic
   Daemon name: libvirtd
   Missing dependencies:
   • virtlogd.socket - not enabled
   • virtlockd.socket - not enabled

⚠️  Libvirt Issues Detected:
   • Libvirt daemon dependencies not running

🔧 Fix available: Enable missing dependencies
   Would execute:
   • sudo systemctl enable virtlogd.socket
   • sudo systemctl start virtlogd.socket
   • sudo systemctl enable virtlockd.socket
   • sudo systemctl start virtlockd.socket

Proceed? [y/N]: y

✅ Dependencies fixed
✅ Libvirt daemon started successfully
```

## 📊 Technical Details

### Investigation Steps
1. **Identify Dependencies:**
   ```bash
   systemctl show libvirtd.service -p Requires -p Wants -p After -p Before --no-pager
   systemctl list-dependencies libvirtd.service --all
   ```

2. **Check Dependency Status:**
   ```bash
   systemctl status virtlogd.socket
   systemctl status virtlogd.service
   systemctl status virtlockd.socket
   systemctl status virtlockd.service
   ```

3. **Check Detailed Logs:**
   ```bash
   journalctl -u libvirtd -u virtlogd -u virtlockd -n 100 --no-pager
   ```

### Expected Dependencies (Typical Libvirt Setup)
- `virtlogd.socket` - Logging daemon socket
- `virtlockd.socket` - Lock manager socket
- `dbus.service` - D-Bus system bus
- Network dependencies for bridge networking

### Implementation Location
**Files to Modify:**
1. `src/app/virt/libvirt_detector.go`:
   - Add `MissingDependencies []string` to LibvirtStatus
   - Add function `DetectLibvirtDependencies() []string`
   - Update `DetectLibvirtStatus()` to check dependencies

2. `src/cmd/virt_exec.go`:
   - Add `executeFixDependencies()` function
   - Update `handleLibvirtFix()` to handle dependency failures
   - Add dependency info to status display

### Code Structure
```go
// In libvirt_detector.go
type LibvirtStatus struct {
    // ... existing fields ...
    MissingDependencies []string
    FailedDependencies  []string
}

func DetectLibvirtDependencies(daemonName string) (missing, failed []string) {
    // Check common dependencies
    deps := []string{
        "virtlogd.socket",
        "virtlockd.socket",
        "dbus.service",
    }

    for _, dep := range deps {
        if !checkService(dep) {
            missing = append(missing, dep)
        } else if isServiceFailed(dep) {
            failed = append(failed, dep)
        } else if !isServiceActive(dep) && !isServiceEnabled(dep) {
            missing = append(missing, dep)
        }
    }
    return
}

func EnableDependency(depName string) error {
    // Enable and start dependency
}
```

## 🧪 Testing Strategy

### Test Cases
1. **TC001: Detect Missing Dependencies**
   - Setup: Fresh libvirt install without dependencies
   - Execute: `portunix virt check`
   - Expected: Reports missing virtlogd/virtlockd

2. **TC002: Fix Dependencies**
   - Setup: Disabled virtlogd/virtlockd
   - Execute: `portunix virt check --fix-libvirt`
   - Expected: Enables and starts dependencies

3. **TC003: Verify Working State**
   - Setup: After dependency fix
   - Execute: `virt-manager` connection test
   - Expected: virt-manager connects successfully

### Manual Testing
```bash
# Disable dependencies to reproduce issue
sudo systemctl stop virtlogd.socket
sudo systemctl disable virtlogd.socket
sudo systemctl stop virtlockd.socket
sudo systemctl disable virtlockd.socket

# Test detection
./portunix virt check

# Test fix
sudo ./portunix virt check --fix-libvirt

# Verify virt-manager works
virt-manager
```

## 📚 References

### Systemd Documentation
- [systemd.unit - Unit Dependencies](https://www.freedesktop.org/software/systemd/man/systemd.unit.html#Requires=)
- [Socket Activation](https://www.freedesktop.org/software/systemd/man/systemd.socket.html)

### Libvirt Documentation
- [Libvirt Daemon Architecture](https://libvirt.org/daemons.html)
- [Modular vs Monolithic Daemon](https://libvirt.org/manpages/libvirtd.html)

### Related Code
- `src/app/virt/libvirt_detector.go:77-96` - Current issue detection logic
- `src/cmd/virt_exec.go:327-379` - handleLibvirtFix() function
- Issue #090 implementation (commit: 414c214)

## 💡 Implementation Notes

### Common Libvirt Dependencies
```
libvirtd.service depends on:
├── virtlogd.socket (logging)
├── virtlockd.socket (locking)
├── dbus.service (communication)
├── systemd-machined.service (container/VM tracking)
└── network.target (networking)

Socket activation requires:
├── libvirtd.socket (main socket)
├── libvirtd-ro.socket (read-only socket)
├── libvirtd-admin.socket (admin socket)
└── libvirtd-tcp.socket (TCP socket, optional)
```

### Error Patterns to Detect
```
"Dependency failed" → Check dependencies
"trigger-limit-hit" → Too many start attempts
"start-limit-hit" → Service crashing repeatedly
```

## ⚠️ Known Issues
- Current implementation detects socket failure but not the root cause
- Reset socket fails because dependency is still broken
- Switch to direct daemon also fails due to same dependency issue

## 🎯 Success Metrics
- [ ] Portunix detects specific missing/failed dependencies
- [ ] `--fix-libvirt` successfully enables dependencies
- [ ] virt-manager connects after fix
- [ ] No more "dependency failed" errors in logs
- [ ] Socket activation works correctly

## 📅 Estimated Effort
**Complexity**: Medium
**Time Estimate**: 2-4 hours
- Investigation: 1 hour
- Implementation: 1-2 hours
- Testing: 1 hour

## 🔗 Related Files
```
src/app/virt/libvirt_detector.go
src/cmd/virt_exec.go
src/helpers/ptx-virt/qemu_adapter.go
docs/issues/done/090-libvirt-daemon-detection-and-fix.md
```

---

**Note for AI Assistant:**
This issue requires manual system investigation first. You MUST run the systemd commands to identify the actual failing dependency before implementing the fix. Do not guess - use `systemctl show`, `systemctl list-dependencies`, and `journalctl` to find the real problem.

Start with:
```bash
systemctl show libvirtd.service -p Requires -p Wants --no-pager
systemctl list-dependencies libvirtd.service --all | grep -E "●|×"
```

Look for any service marked with "●" (failed) or "×" (missing).
