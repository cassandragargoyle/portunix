//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

const (
	// lockfileExclusiveLock is LOCKFILE_EXCLUSIVE_LOCK; without LOCKFILE_FAIL_IMMEDIATELY the call blocks
	lockfileExclusiveLock = 0x02
	// lockfileFailImmediately is LOCKFILE_FAIL_IMMEDIATELY
	lockfileFailImmediately = 0x01
	// errorLockViolation is ERROR_LOCK_VIOLATION, returned when another process holds the lock
	errorLockViolation = syscall.Errno(33)
)

// lockFile blocks until it holds an exclusive lock on the file
func lockFile(f *os.File) error {
	ol := new(syscall.Overlapped)
	r1, _, err := procLockFileEx.Call(f.Fd(), lockfileExclusiveLock, 0, 1, 0, uintptr(unsafe.Pointer(ol))) // #nosec G103 -- Win32 API call
	if r1 == 0 {
		return err
	}
	return nil
}

// tryLockFile takes an exclusive lock on the file without waiting; it
// reports false when another process holds the lock
func tryLockFile(f *os.File) (bool, error) {
	ol := new(syscall.Overlapped)
	r1, _, err := procLockFileEx.Call(f.Fd(), lockfileExclusiveLock|lockfileFailImmediately, 0, 1, 0, uintptr(unsafe.Pointer(ol))) // #nosec G103 -- Win32 API call
	if r1 != 0 {
		return true, nil
	}
	if err == errorLockViolation {
		return false, nil
	}
	return false, err
}

// unlockFile releases the lock taken by lockFile
func unlockFile(f *os.File) error {
	ol := new(syscall.Overlapped)
	r1, _, err := procUnlockFileEx.Call(f.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(ol))) // #nosec G103 -- Win32 API call
	if r1 == 0 {
		return err
	}
	return nil
}
