# Portunix Shared Clipboard Module

## Overview

This module provides cross-platform clipboard operations for Portunix and its helpers. It combines multiple clipboard access methods to ensure maximum compatibility across different operating systems and desktop environments.

## Features

- **Cross-platform support**: Windows, macOS, Linux
- **Multiple fallback mechanisms**: Standard clipboard libraries + D-Bus integration
- **Desktop environment detection**: KDE, GNOME support
- **Detailed system information**: Comprehensive capability detection
- **User-friendly feedback**: Progress indicators and error messages

## Usage

```go
import "portunix.ai/portunix/src/shared"

// Create clipboard manager
cm := shared.NewClipboardManager()

// Check if clipboard is supported
if cm.IsSupported() {
    // Write to clipboard
    err := cm.Write("Hello, clipboard!")

    // Write with user feedback
    err := cm.WriteWithFeedback("Content with feedback")

    // Read from clipboard
    content, err := cm.Read()

    // Check if clipboard has content
    hasContent, err := cm.HasContent()

    // Clear clipboard
    err := cm.ClearClipboard()
}

// Get detailed system information
info := cm.GetSystemInfo()
fmt.Printf("Platform: %s, Methods: %v\n", info.Platform, info.Methods)
```

## Platform Support

### Windows
- ✅ **Full support** via standard Windows clipboard API
- Uses `github.com/atotto/clipboard` library
- No additional dependencies required

### macOS
- ✅ **Full support** via NSPasteboard
- Uses `github.com/atotto/clipboard` library
- No additional dependencies required

### Linux
- ✅ **Multi-method support**:
  - **Standard**: xclip/xsel utilities (via `github.com/atotto/clipboard`)
  - **D-Bus fallback**: KDE Klipper integration
  - **Future**: Wayland wl-clipboard support (planned)

#### Linux Dependencies
- **Option 1**: Install `xclip` or `xsel` packages
- **Option 2**: Use KDE desktop environment with Klipper
- **Automatic fallback**: D-Bus when standard methods fail

## Architecture

### ClipboardManager
Main interface providing clipboard operations with automatic method selection.

### Detection Logic
1. Test standard clipboard support (`github.com/atotto/clipboard`)
2. On Linux: Test D-Bus clipboard support (KDE Klipper)
3. Enable clipboard if any method is available
4. Use best available method for each operation

### Fallback Chain (Linux)
```
Standard Clipboard (xclip/xsel)
         ↓ (if fails)
D-Bus KDE Klipper
         ↓ (if fails)
Error with helpful message
```

## Error Handling

The module provides graceful error handling with informative messages:

- **Not supported**: Clear indication of missing system requirements
- **Method failures**: Automatic fallback to alternative methods
- **User guidance**: Suggestions for installing missing packages

## System Information

```go
info := cm.GetSystemInfo()
// Returns ClipboardInfo struct with:
// - Platform: OS name (windows/darwin/linux)
// - Supported: Whether clipboard is available
// - Available: Whether clipboard is currently accessible
// - Methods: List of available clipboard methods
// - StandardSupport: Standard library support status
// - DBusSupport: D-Bus integration support (Linux)
// - DetectedDesktop: Desktop environment (KDE/GNOME/etc.)
// - RequiredPackages: Missing system packages
```

## Integration Examples

### PTX-Prompting Helper
```go
// In cmd/build.go
clipboardMgr := shared.NewClipboardManager()

// Default behavior: copy to clipboard + show on stdout
if !noCopy {
    err := clipboardMgr.WriteWithFeedback(result.Content)
    // Handles fallback automatically
}
```

### Future Helpers
```go
// Any Portunix helper can use shared clipboard
import "portunix.ai/portunix/src/shared"

func saveToClipboard(content string) error {
    cm := shared.NewClipboardManager()
    return cm.Write(content)
}
```

## TODO: Library Maintenance

**Last reviewed**: 2025-09-26
**Review schedule**: Quarterly

### Monitoring Checklist
- [ ] Check for updates to `github.com/atotto/clipboard`
- [ ] Evaluate new Go clipboard libraries
- [ ] Test compatibility with new Linux distributions
- [ ] Monitor Wayland clipboard standards
- [ ] Review D-Bus clipboard specifications

### Alternative Libraries to Consider
- `github.com/zyedidia/clipboard` - Better Linux support but requires packages
- `github.com/d2r2/go-clipboard` - More platform-specific options
- Direct system API integration (Windows user32.dll, macOS NSPasteboard)

### Platform Improvements Needed
- **Wayland**: Add `wl-clipboard` support
- **GNOME**: Add GNOME clipboard D-Bus integration
- **Android**: Termux clipboard support (future mobile support)
- **Web**: WASM clipboard API integration (future web support)

### Known Issues
1. **Linux Wayland**: Limited testing, compositor-dependent
2. **Rich content**: Only text clipboard supported (no images/files)
3. **Clipboard history**: No integration with system clipboard managers
4. **Security**: No encrypted clipboard support

## Testing

Run the included test:
```bash
go run test_shared_clipboard.go
```

Expected output shows:
- Platform detection
- Available methods
- Capability information
- Write/read test results

## Migration Guide

### From Helper-Specific Clipboard
```go
// Old (helper-specific)
import "portunix.ai/portunix/src/helpers/ptx-prompting/internal/clipboard"
clipboardMgr := clipboard.NewClipboardManager()

// New (shared)
import "portunix.ai/portunix/src/shared"
clipboardMgr := shared.NewClipboardManager()
```

### API Compatibility
The shared module maintains the same API as the original helper clipboard modules for easy migration.

## Security Considerations

- **No credential storage**: Module only handles text content
- **No network access**: All operations are local system calls
- **User consent**: Operations are explicit, no background clipboard monitoring
- **Sandboxing**: Compatible with containerized environments

---

**Maintainer**: Portunix Development Team
**License**: Same as Portunix project
**Issues**: Report via main Portunix issue tracker