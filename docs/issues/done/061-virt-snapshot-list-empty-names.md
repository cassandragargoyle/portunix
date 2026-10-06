# Issue #61: Virtual Machine Snapshot List Shows Empty Names

**Created**: 2025-09-24
**Status**: ✅ Implemented
**Priority**: High
**Type**: Bug Fix

## Summary
The command `portunix virt snapshot list Win11-ZH` displays snapshot entries with empty names, showing only timestamps but missing snapshot names, sizes, and descriptions.

## Current Behavior
```
Snapshots for VM 'Win11-ZH':

NAME                 CREATED              SIZE       DESCRIPTION
----                 -------              ----       -----------
            2025-09-24 11:21     -          -
            2025-09-24 11:21     -          -
       2025-09-24 11:21     -          -
```

## Expected Behavior
```
Snapshots for VM 'Win11-ZH':

NAME                 CREATED              SIZE       DESCRIPTION
----                 -------              ----       -----------
BaseInstall          2025-09-24 11:21     2.1GB      Fresh Windows 11 installation
PostUpdates          2025-09-24 11:22     2.3GB      After Windows updates
DevEnvironment       2025-09-24 11:23     2.5GB      Development tools installed
```

## Problem Analysis
The issue appears to be in the snapshot parsing logic where:
- Snapshot names are not being extracted correctly from the backend
- Size information is not being retrieved or parsed
- Description fields are not being populated
- Only timestamp data is being successfully parsed

## Affected Components
- `portunix virt snapshot list` command
- Virtual machine snapshot parsing logic
- Backend integration (likely VirtualBox/QEMU snapshot queries)

## Technical Investigation Needed
1. **Backend Query Analysis**
   - Verify the underlying VBoxManage or QEMU command being executed
   - Check if the backend returns complete snapshot information
   - Analyze the raw output from the virtualization backend

2. **Parsing Logic Review**
   - Review snapshot data parsing functions
   - Check for field extraction errors
   - Verify column alignment and formatting logic

3. **Error Handling**
   - Check if parsing errors are being silently ignored
   - Verify that all snapshot fields are being processed
   - Ensure proper error reporting for malformed data

## Reproduction Steps
1. Create VM with snapshots using VirtualBox or QEMU
2. Run `portunix virt snapshot list <VM-NAME>`
3. Observe empty name, size, and description fields

## Expected Implementation Areas
- `cmd/virt_snapshot.go` - Command implementation
- `app/virtualization/` - Backend integration
- Snapshot parsing and formatting functions

## Error Scenarios to Test
- VMs with no snapshots (should show empty list)
- VMs with single snapshot
- VMs with multiple snapshots
- Snapshots with special characters in names
- Snapshots without descriptions
- Corrupted or partially available snapshot data

## Acceptance Criteria
- [ ] Snapshot names are displayed correctly
- [ ] Snapshot sizes are shown in human-readable format (GB/MB)
- [ ] Snapshot descriptions are displayed when available
- [ ] Empty descriptions show as "-" or "(no description)"
- [ ] Timestamps are formatted consistently
- [ ] Command works for both VirtualBox and QEMU backends
- [ ] Error messages are clear when snapshot data is unavailable
- [ ] Table formatting is aligned correctly

## Testing Requirements
- Test with VirtualBox VMs containing multiple snapshots
- Test with QEMU VMs containing snapshots
- Test edge cases (no snapshots, corrupted snapshot data)
- Cross-platform testing (Windows/Linux)
- Test with VMs having special characters in snapshot names

## Priority
**High** - This significantly impacts the usability of VM snapshot management functionality

## Labels
- bug
- virtualization
- snapshot-management
- virtualbox
- qemu
- data-parsing

## Related Commands
- `portunix virt snapshot create`
- `portunix virt snapshot restore`
- `portunix virt snapshot delete`
- `portunix virt list`

## Expected Fix Areas
1. **Snapshot Data Extraction**
   - Fix parsing of VBoxManage snapshot output
   - Fix parsing of QEMU snapshot information
   - Ensure all fields are extracted correctly

2. **Data Formatting**
   - Proper column alignment
   - Size formatting (bytes to GB/MB)
   - Date/time formatting consistency

3. **Error Handling**
   - Better error messages for parsing failures
   - Graceful handling of missing snapshot data
   - Clear indication when snapshots cannot be retrieved

## Debugging Information Needed
- Raw output from `VBoxManage snapshot <VM> list --machinereadable`
- Raw output from QEMU snapshot queries
- Portunix debug logs during snapshot list execution
- VM configuration and snapshot structure