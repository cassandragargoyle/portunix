//go:build windows

package service

import (
	"syscall"
	"unsafe"
)

// CREATE_NEW_PROCESS_GROUP detaches child from parent console
const createNewProcessGroup = 0x00000200

// platformSysProcAttr returns SysProcAttr for detaching child process on Windows
func platformSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: createNewProcessGroup,
	}
}

// platformFileLock acquires an exclusive file lock using LockFileEx
func platformFileLock(fd uintptr) error {
	// LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY = 0x02 | 0x01
	// Use blocking lock (just LOCKFILE_EXCLUSIVE_LOCK = 0x02)
	ol := new(syscall.Overlapped)
	return lockFileEx(syscall.Handle(fd), 0x02, 0, 1, 0, ol)
}

// platformFileUnlock releases a file lock using UnlockFileEx
func platformFileUnlock(fd uintptr) error {
	ol := new(syscall.Overlapped)
	return unlockFileEx(syscall.Handle(fd), 0, 1, 0, ol)
}

// platformIsProcessAlive checks if a process is alive on Windows
func platformIsProcessAlive(pid int) bool {
	proc, err := findProcessByPID(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds; try Signal(nil) which calls
	// OpenProcess - if it fails the process is gone
	return proc.Signal(syscall.Signal(0)) == nil
}

var (
	modkernel32    = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

func lockFileEx(handle syscall.Handle, flags uint32, reserved uint32, bytesLow uint32, bytesHigh uint32, ol *syscall.Overlapped) error {
	r1, _, err := procLockFileEx.Call(
		uintptr(handle),
		uintptr(flags),
		uintptr(reserved),
		uintptr(bytesLow),
		uintptr(bytesHigh),
		uintptr(unsafe.Pointer(ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func unlockFileEx(handle syscall.Handle, reserved uint32, bytesLow uint32, bytesHigh uint32, ol *syscall.Overlapped) error {
	r1, _, err := procUnlockFileEx.Call(
		uintptr(handle),
		uintptr(reserved),
		uintptr(bytesLow),
		uintptr(bytesHigh),
		uintptr(unsafe.Pointer(ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}
