//go:build !windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import "syscall"

func mkFIFO(path string, mode uint32) error {
	return syscall.Mkfifo(path, mode)
}
