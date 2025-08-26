# 🐛 **Bug Report - Issue #012 PowerShell Linux Installation**

---

## **Bug Summary**
CLI argument parsing incorrectly handles `--help` flag when used with specific package argument `powershell`.

---

## **Bug Details**

| **Field** | **Value** |
|-----------|-----------|
| **Bug ID** | BUG-012-001 |
| **Reporter** | QA Tester |
| **Date** | 2025-08-23 |
| **Severity** | Medium |
| **Priority** | High |
| **Component** | CLI Command Parsing |
| **Affected Feature** | Issue #012 - PowerShell Linux Installation |
| **Branch** | feature/012-powershell-linux-installation |

---

## **Environment**
- **OS**: Linux
- **Build**: Current feature branch build
- **Test Phase**: Smoke Testing

---

## **Steps to Reproduce**
1. Compile portunix from feature/012-powershell-linux-installation branch
2. Execute command: `./portunix install powershell --help`
3. Observe output

---

## **Expected Result**
Command should display **specific help for PowerShell package**, similar to:
```
PowerShell Installation Help

Usage: 
  portunix install powershell [flags]

Description: 
  Install Microsoft PowerShell on Linux distributions

Available variants:
  ubuntu, debian, fedora, rocky, mint, elementary, snap

Examples:
  portunix install powershell
  portunix install powershell --variant ubuntu
  portunix install powershell --variant snap

Flags:
  --variant string   Specify installation variant
  -h, --help         help for powershell
```

---

## **Actual Result**
Command displays **generic install command help** showing all available packages, variants, and examples instead of PowerShell-specific help.

---

## **Impact Assessment**
- **User Experience**: Poor - users cannot get specific help for PowerShell installation
- **Documentation**: Missing - no way to discover PowerShell-specific options
- **Functionality**: Core install may work, but discoverability is broken
- **Testing**: Blocks comprehensive testing - testers need specific help to understand available options

---

## **Suggested Investigation Areas**
1. **CLI Command Structure**: Check if `install powershell --help` parsing is implemented
2. **Cobra CLI Framework**: Verify subcommand help handling for package-specific commands
3. **Help Function Routing**: Ensure help requests are routed to package-specific handlers
4. **cmd/install.go**: Review argument parsing logic for package + help combinations

---

## **Acceptance Criteria for Fix**
- [x] `./portunix install powershell --help` shows PowerShell-specific help
- [x] Help includes all available PowerShell variants (ubuntu, debian, fedora, rocky, mint, elementary, snap)
- [x] Help includes PowerShell-specific usage examples
- [x] Help includes PowerShell-specific description and flags
- [x] Other package help commands work consistently (e.g., `./portunix install java --help`)

---

## **Fix Implementation**
**Files Modified:**
- `cmd/install.go`: Enhanced argument parsing to detect package-specific help requests
- `app/install/install.go`: Added `ShowPackageHelp()` function with specific help for all packages

**Solution:**
1. Modified CLI parsing to detect pattern: `install <package> --help`
2. Implemented package-specific help system with fallback to general help
3. Added comprehensive PowerShell help showing all Linux distribution variants
4. Extended system to support help for all packages (java, python, vscode, docker, podman)

**Testing Results:**
- ✅ `./portunix install powershell --help` shows PowerShell-specific help
- ✅ All PowerShell variants displayed (ubuntu, debian, fedora, rocky, mint, elementary, snap)
- ✅ Auto-detection explanation included
- ✅ Platform-specific variants (Linux vs Windows)
- ✅ Consistent help for other packages (java, python, etc.)
- ✅ Graceful fallback for packages without specific help

---

## **Priority Justification**
**High Priority** because:
- ✅ ~~Blocks comprehensive QA testing of Issue #012~~ - **RESOLVED**
- ✅ ~~Poor user experience for PowerShell feature discovery~~ - **RESOLVED**  
- ✅ ~~Easy to fix but essential for feature usability~~ - **FIXED**
- ✅ ~~Required before integration testing can proceed~~ - **READY FOR TESTING**

---

**Status**: ✅ **RESOLVED - READY FOR QA TESTING**