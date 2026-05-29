//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import "errors"

func mkFIFO(path string, mode uint32) error {
	return errors.New("FIFOs are not supported on Windows; use --write-fd instead")
}
