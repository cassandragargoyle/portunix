package shared

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
)

/*
TODO: Monitor Go clipboard library ecosystem and update this module accordingly

Current implementation uses:
- github.com/atotto/clipboard as primary clipboard library
- Custom D-Bus fallback for KDE Linux systems via dbus-send

Alternative libraries to consider:
- github.com/zyedidia/clipboard (requires xclip/xsel on Linux)
- github.com/d2r2/go-clipboard (more platform-specific options)
- Direct system API calls (Windows: user32.dll, macOS: NSPasteboard)

Last reviewed: 2025-09-26
Review schedule: Check quarterly for new Go clipboard libraries or improvements

Platform support status:
- ✅ Windows: Works via atotto/clipboard (Windows API)
- ✅ macOS: Works via atotto/clipboard (NSPasteboard)
- ✅ Linux X11: Works via atotto/clipboard (requires xclip/xsel) + D-Bus fallback for KDE
- ⚠️ Linux Wayland: D-Bus fallback may work, but not extensively tested
- ❌ Other Unix: Limited support

Known limitations:
- Linux requires either xclip/xsel utilities OR running KDE with Klipper
- Wayland clipboard support varies by compositor
- No support for rich text or binary clipboard content (text only)

Future improvements:
- Add Wayland wl-clipboard support
- Add GNOME clipboard D-Bus integration
- Consider binary/rich content support
- Add clipboard history integration where available
*/

// ClipboardManager provides cross-platform clipboard operations with multiple fallback mechanisms
type ClipboardManager struct {
	enabled          bool
	supportsStandard bool // Standard clipboard (atotto/clipboard)
	supportsDBus     bool // D-Bus clipboard (KDE Klipper)
}

// ClipboardInfo contains detailed information about clipboard support
type ClipboardInfo struct {
	Platform         string   `json:"platform"`
	Supported        bool     `json:"supported"`
	Available        bool     `json:"available"`
	Methods          []string `json:"methods"`           // Available clipboard methods
	StandardSupport  bool     `json:"standard_support"`  // Standard library support
	DBusSupport      bool     `json:"dbus_support"`      // D-Bus support (Linux)
	DetectedDesktop  string   `json:"detected_desktop"`  // Desktop environment (KDE, GNOME, etc.)
	RequiredPackages []string `json:"required_packages"` // Missing system packages
}

// NewClipboardManager creates a new cross-platform clipboard manager
func NewClipboardManager() *ClipboardManager {
	cm := &ClipboardManager{}
	cm.detectCapabilities()
	return cm
}

// detectCapabilities detects available clipboard mechanisms
func (cm *ClipboardManager) detectCapabilities() {
	cm.supportsStandard = cm.testStandardClipboard()
	cm.supportsDBus = false

	// On Linux, also check D-Bus clipboard support
	if runtime.GOOS == "linux" {
		cm.supportsDBus = cm.testDBusClipboard()
	}

	// Enable clipboard if any method is available
	cm.enabled = cm.supportsStandard || cm.supportsDBus
}

// IsSupported returns whether clipboard operations are supported
func (cm *ClipboardManager) IsSupported() bool {
	return cm.enabled
}

// Write writes content to the clipboard using the best available method
func (cm *ClipboardManager) Write(content string) error {
	if !cm.enabled {
		return fmt.Errorf("clipboard not supported on this platform")
	}

	// Try standard clipboard first (most reliable)
	if cm.supportsStandard {
		if err := clipboard.WriteAll(content); err == nil {
			return nil
		}
	}

	// Fallback to D-Bus for Linux systems
	if runtime.GOOS == "linux" && cm.supportsDBus {
		if err := cm.writeViaDBus(content); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to write to clipboard: all methods failed")
}

// Read reads content from the clipboard using the best available method
func (cm *ClipboardManager) Read() (string, error) {
	if !cm.enabled {
		return "", fmt.Errorf("clipboard not supported on this platform")
	}

	// Try standard clipboard first
	if cm.supportsStandard {
		if content, err := clipboard.ReadAll(); err == nil {
			return content, nil
		}
	}

	// Fallback to D-Bus for Linux systems
	if runtime.GOOS == "linux" && cm.supportsDBus {
		if content, err := cm.readViaDBus(); err == nil {
			return content, nil
		}
	}

	return "", fmt.Errorf("failed to read from clipboard: all methods failed")
}

// WriteWithFeedback writes to clipboard and provides user feedback
func (cm *ClipboardManager) WriteWithFeedback(content string) error {
	if !cm.enabled {
		fmt.Println("⚠️  Clipboard not supported on this platform")
		fmt.Println("📄 Content:")
		fmt.Println("----")
		fmt.Println(content)
		fmt.Println("----")
		return nil
	}

	err := cm.Write(content)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Content copied to clipboard (%d characters)\n", len(content))
	return nil
}

// GetSystemInfo returns detailed information about clipboard support
func (cm *ClipboardManager) GetSystemInfo() ClipboardInfo {
	info := ClipboardInfo{
		Platform:        runtime.GOOS,
		Supported:       cm.enabled,
		Available:       cm.enabled && cm.testCurrentClipboard(),
		Methods:         []string{},
		StandardSupport: cm.supportsStandard,
		DBusSupport:     cm.supportsDBus,
	}

	// Add available methods
	if cm.supportsStandard {
		info.Methods = append(info.Methods, "standard")
	}
	if cm.supportsDBus {
		info.Methods = append(info.Methods, "dbus")
		info.DetectedDesktop = "KDE" // We currently only support KDE via D-Bus
	}

	// Add required packages for Linux
	if runtime.GOOS == "linux" && !cm.supportsStandard {
		info.RequiredPackages = []string{"xclip", "xsel"}
	}

	return info
}

// testStandardClipboard tests if standard clipboard operations work
func (cm *ClipboardManager) testStandardClipboard() bool {
	testContent := "portunix-clipboard-test"

	if err := clipboard.WriteAll(testContent); err != nil {
		return false
	}

	content, err := clipboard.ReadAll()
	if err != nil {
		return false
	}

	return content == testContent
}

// testCurrentClipboard tests if clipboard is currently accessible
func (cm *ClipboardManager) testCurrentClipboard() bool {
	if !cm.enabled {
		return false
	}

	// Try to read current clipboard content (non-destructive test)
	_, err := cm.Read()
	return err == nil
}

// writeViaDBus writes content to clipboard via D-Bus (Linux KDE Klipper)
func (cm *ClipboardManager) writeViaDBus(content string) error {
	// Try KDE Klipper via D-Bus
	cmd := exec.Command("dbus-send",
		"--session",
		"--dest=org.kde.klipper",
		"--type=method_call",
		"/klipper",
		"org.kde.klipper.klipper.setClipboardContents",
		fmt.Sprintf("string:%s", content))

	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback: Try generic D-Bus clipboard portal
	cmd = exec.Command("dbus-send",
		"--session",
		"--dest=org.freedesktop.portal.Desktop",
		"--type=method_call",
		"/org/freedesktop/portal/desktop",
		"org.freedesktop.portal.Clipboard.SetSelection",
		fmt.Sprintf("string:%s", content))

	return cmd.Run()
}

// readViaDBus reads content from clipboard via D-Bus (Linux KDE Klipper)
func (cm *ClipboardManager) readViaDBus() (string, error) {
	// Try KDE Klipper via D-Bus
	cmd := exec.Command("dbus-send",
		"--session",
		"--dest=org.kde.klipper",
		"--type=method_call",
		"--print-reply",
		"/klipper",
		"org.kde.klipper.klipper.getClipboardContents")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read clipboard via D-Bus: %v", err)
	}

	// Parse D-Bus output format: method return, string "content"
	content := strings.TrimSpace(string(output))
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "string") && strings.Contains(line, "\"") {
			// Extract content between quotes
			parts := strings.SplitN(line, "\"", 3)
			if len(parts) >= 3 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("failed to parse clipboard content from D-Bus output")
}

// testDBusClipboard tests if D-Bus clipboard operations are available
func (cm *ClipboardManager) testDBusClipboard() bool {
	// Check if dbus-send command is available
	if _, err := exec.LookPath("dbus-send"); err != nil {
		return false
	}

	// Test if KDE Klipper service is available
	cmd := exec.Command("dbus-send",
		"--session",
		"--dest=org.kde.klipper",
		"--type=method_call",
		"/klipper",
		"org.kde.klipper.klipper.getClipboardContents")

	return cmd.Run() == nil
}

// ClearClipboard clears the clipboard content (sets to empty string)
func (cm *ClipboardManager) ClearClipboard() error {
	return cm.Write("")
}

// HasContent checks if clipboard has any content (non-empty)
func (cm *ClipboardManager) HasContent() (bool, error) {
	content, err := cm.Read()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(content) != "", nil
}