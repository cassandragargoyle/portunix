//go:build !windows

package service

import "syscall"

// platformSysProcAttr returns SysProcAttr for detaching child process
func platformSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// platformFileLock acquires an exclusive file lock
func platformFileLock(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_EX)
}

// platformFileUnlock releases a file lock
func platformFileUnlock(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_UN)
}

// platformIsProcessAlive checks if a process is alive via signal 0
func platformIsProcessAlive(pid int) bool {
	proc, err := findProcessByPID(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
