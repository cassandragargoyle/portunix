//go:build windows

package engine

import (
	"golang.org/x/sys/windows"
)

// IsAdmin checks if the current process is running with administrator privileges
func IsAdmin() bool {
	var sid *windows.SID

	// Create a SID for the Administrators group
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// Get the current process token
	token := windows.Token(0)

	// Check if the token is a member of the Administrators group
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}
