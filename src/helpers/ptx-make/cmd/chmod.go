package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"portunix.ai/portunix/src/pkg/platform"
)

// RunChmod executes the chmod command
func RunChmod(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: chmod <mode> <file>")
	}

	mode := args[0]
	file := args[1]

	// No-op on Windows
	if platform.IsWindows() {
		return nil
	}

	// Parse mode
	var perm os.FileMode

	if strings.HasPrefix(mode, "+") || strings.HasPrefix(mode, "-") {
		// Symbolic mode (+x, -w, etc.)
		info, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("file not found: %s", file)
		}
		perm = info.Mode()

		switch mode {
		case "+x":
			perm |= 0111 // Add execute for all
		case "-x":
			perm &^= 0111 // Remove execute for all
		case "+w":
			perm |= 0200 // Add write for owner
		case "-w":
			perm &^= 0222 // Remove write for all
		case "+r":
			perm |= 0444 // Add read for all
		case "-r":
			perm &^= 0444 // Remove read for all
		default:
			return fmt.Errorf("unsupported symbolic mode: %s", mode)
		}
	} else {
		// Octal mode (755, 644, etc.)
		parsed, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return fmt.Errorf("invalid mode: %s", mode)
		}
		perm = os.FileMode(parsed)
	}

	return os.Chmod(file, perm)
}
