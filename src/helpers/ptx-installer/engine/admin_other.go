//go:build !windows

package engine

import (
	"os"
)

// IsAdmin checks if the current process is running with root privileges
func IsAdmin() bool {
	return os.Geteuid() == 0
}
