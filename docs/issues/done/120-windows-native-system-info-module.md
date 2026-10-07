# Issue #120: Windows Native System Info Module

## Summary

Replace PowerShell `Get-ComputerInfo` and `wmic` commands with native Windows API calls in Go for significantly improved performance and reliability.

## Current Implementation

The current `portunix system info` command uses external processes:

**File**: `src/app/system/system.go`

| Function | Command Used | Line |
| -------- | ------------ | ---- |
| `getWindowsInfo()` | `wmic os get Caption,Version,BuildNumber /format:csv` | 155 |
| `getWindowsInfo()` | `cmd /c ver` (fallback) | 180 |
| `detectEnvironment()` | `tasklist /FI "IMAGENAME eq CExecSvc.exe"` | 273 |
| `detectEnvironment()` | `wmic computersystem get manufacturer,model` | 308 |
| `checkCapabilities()` | `net session` | 392 |
| `checkHardwareVirtualization()` | `powershell Get-ComputerInfo \| HyperVisorPresent` | 495 |
| `queryWindowsRegistry()` | `reg query` | 634 |

### Problems with Current Approach

1. **Performance**: Each external process spawn adds 100-500ms latency
2. **Reliability**: Depends on external tools being available
3. **Resource Usage**: Multiple process spawns consume significant resources
4. **Localization Issues**: `wmic` output varies by system locale

## Proposed Solution

Create a new `src/app/system/windows/` package with native Go implementations using Windows API calls via `golang.org/x/sys/windows`.

### Reference Implementation: fastfetch

The [fastfetch](https://github.com/fastfetch-cli/fastfetch) project demonstrates how to get system information using native Windows APIs. Key files:

| File | Information Type |
| ---- | ---------------- |
| `src/util/platform/FFPlatform_windows.c` | OS version, architecture, hostname |
| `src/detection/host/host_windows.c` | Host/manufacturer via SMBIOS |
| `src/detection/cpu/cpu_windows.c` | CPU info via registry, SMBIOS |
| `src/util/windows/registry.h` | Registry access utilities |
| `src/util/windows/nt.h` | NT API definitions |

### Windows APIs to Use (Convert C to Go)

#### 1. OS Version (`RtlGetVersion`)

```c
// fastfetch: FFPlatform_windows.c:186-225
RTL_OSVERSIONINFOW osVersion = { .dwOSVersionInfoSize = sizeof(osVersion) };
RtlGetVersion(&osVersion);
// osVersion.dwMajorVersion, dwMinorVersion, dwBuildNumber
```

**Go equivalent** (to implement):

```go
import "golang.org/x/sys/windows"

func getWindowsVersion() (major, minor, build uint32, err error) {
    // Use RtlGetVersion from ntdll.dll
}
```

#### 2. Architecture (`GetNativeSystemInfo`)

```c
// fastfetch: FFPlatform_windows.c:227-282
SYSTEM_INFO sysInfo;
GetNativeSystemInfo(&sysInfo);
// sysInfo.wProcessorArchitecture
```

#### 3. Hostname (`GetComputerNameExW`)

```c
// fastfetch: FFPlatform_windows.c:156-162
wchar_t buffer[128];
DWORD len = ARRAY_SIZE(buffer);
GetComputerNameExW(ComputerNameDnsHostname, buffer, &len);
```

#### 4. Registry Access (Direct)

```c
// fastfetch: FFPlatform_windows.c:192-205
ffRegOpenKeyForRead(HKEY_LOCAL_MACHINE,
    L"SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", &hKey, NULL);
ffRegReadUint(hKey, L"UBR", &ubr, NULL);
ffRegReadStrbuf(hKey, L"BuildLabEx", &info->version, NULL);
```

**Key registry paths**:

- `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion` - OS info
- `HKLM\HARDWARE\DESCRIPTION\System\CentralProcessor\0` - CPU info
- `HKLM\SOFTWARE\Oracle\VirtualBox` - VirtualBox detection

#### 5. Hardware Virtualization (CPU Features)

```c
// fastfetch uses CPUID instruction or IsProcessorFeaturePresent()
// For Hyper-V detection, check CPUID leaf 0x1 bit 31 (hypervisor present)
```

#### 6. Admin Check (Token-based)

```go
// Instead of `net session`, use token checking:
import "golang.org/x/sys/windows"

func isAdmin() bool {
    var sid *windows.SID
    err := windows.AllocateAndInitializeSid(
        &windows.SECURITY_NT_AUTHORITY,
        2,
        windows.SECURITY_BUILTIN_DOMAIN_RID,
        windows.DOMAIN_ALIAS_RID_ADMINS,
        0, 0, 0, 0, 0, 0,
        &sid)
    if err != nil {
        return false
    }
    defer windows.FreeSid(sid)

    member, err := windows.Token(0).IsMember(sid)
    return err == nil && member
}
```

#### 7. VM Detection (SMBIOS)

```c
// fastfetch: host_windows.c:32-80
// Uses SMBIOS data via GetSystemFirmwareTable()
const FFSmbiosSystemInfo* data = ffGetSmbiosHeaderTable()[FF_SMBIOS_TYPE_SYSTEM_INFO];
// data->Manufacturer contains "VMware", "VirtualBox", etc.
```

## Implementation Plan

### Phase 1: Core Module Structure

1. Create `src/app/system/windows/` package
2. Implement base Windows API bindings
3. Add build tags for Windows-only compilation

### Phase 2: Replace Individual Functions

1. `getWindowsVersion()` - Replace wmic with RtlGetVersion
2. `getArchitecture()` - Replace with GetNativeSystemInfo
3. `getHostname()` - Replace with GetComputerNameExW
4. `isAdmin()` - Replace `net session` with token check
5. `readRegistry()` - Replace `reg query` with direct API

### Phase 3: Advanced Detection

1. `getHardwareVirtualization()` - Replace PowerShell with CPUID/registry
2. `detectVM()` - Replace wmic with SMBIOS via GetSystemFirmwareTable
3. `detectSandbox()` - Optimize process detection

### Phase 4: Integration & Testing

1. Update `src/app/system/system.go` to use new module
2. Add fallback to external commands for edge cases
3. Performance benchmarking
4. Cross-platform build verification

## Expected Benefits

| Metric | Current | Expected |
| ------ | ------- | -------- |
| `system info` execution time | 2-5 seconds | <100ms |
| External process spawns | 5-7 | 0 |
| Memory usage | High (multiple processes) | Low |
| Reliability | Locale-dependent | Consistent |

## Files to Create/Modify

### New Files

- `src/app/system/windows/version.go` - OS version detection
- `src/app/system/windows/arch.go` - Architecture detection
- `src/app/system/windows/registry.go` - Registry utilities
- `src/app/system/windows/admin.go` - Admin privilege check
- `src/app/system/windows/smbios.go` - SMBIOS/firmware access
- `src/app/system/windows/virtualization.go` - VM/hypervisor detection

### Modified Files

- `src/app/system/system.go` - Use new Windows module (build tag)
- `src/app/system/system_windows.go` - Windows-specific implementation

## Dependencies

- `golang.org/x/sys/windows` - Windows API bindings (already available)
- No external dependencies required

## Reference Links

- [fastfetch GitHub](https://github.com/fastfetch-cli/fastfetch)
- [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows)
- [Windows SDK - System Information](https://docs.microsoft.com/en-us/windows/win32/sysinfo/system-information-functions)
- [SMBIOS Specification](https://www.dmtf.org/standards/smbios)

## Acceptance Criteria

- [ ] `portunix system info` executes in <100ms on Windows
- [ ] No external process spawns for standard system info
- [ ] All current information fields preserved
- [ ] Fallback mechanism for edge cases
- [ ] Unit tests for Windows API wrappers
- [ ] Integration tests passing on Windows 10/11

## Priority

**High** - This is part of Issue #099 (System Info Performance Optimization)

## Labels

`enhancement`, `performance`, `windows`, `system-info`, `native-api`

---

**Created**: 2025-12-31
**Status**: ✅ Implemented
**Implemented**: 2026-05-12
**Related Issues**: #099 (System Info Performance Optimization)
**Implementation Commits**: `783906d` (feat: implement native Windows system info module), `48d52f7` (fix: lint), `1b1aedc` (perf comparison metrics)
