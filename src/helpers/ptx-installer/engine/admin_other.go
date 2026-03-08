//go:build !windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os"
)

// IsAdmin checks if the current process is running with root privileges
func IsAdmin() bool {
	return os.Geteuid() == 0
}
