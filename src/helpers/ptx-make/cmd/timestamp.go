/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"
	"time"
)

// RunTimestamp executes the timestamp command
func RunTimestamp(args []string) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	fmt.Println(timestamp)
}
