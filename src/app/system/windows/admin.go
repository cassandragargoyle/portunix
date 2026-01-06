//go:build windows

package windows

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// IsAdmin checks if the current process is running with administrator privileges
// using Windows token-based check instead of external commands like "net session"
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

// IsElevated checks if the current process token is elevated
func IsElevated() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	var elevation uint32
	var outLen uint32

	err = windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&outLen,
	)
	if err != nil {
		return false
	}

	return elevation != 0
}

// GetCurrentUsername returns the current user's name using Windows API
func GetCurrentUsername() string {
	token := windows.Token(0)

	user, err := token.GetTokenUser()
	if err != nil {
		return ""
	}

	account, _, _, err := user.User.Sid.LookupAccount("")
	if err != nil {
		return ""
	}

	return account
}

// IsLocalSystem checks if running as SYSTEM account
func IsLocalSystem() bool {
	var sid *windows.SID

	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		1,
		windows.SECURITY_LOCAL_SYSTEM_RID,
		0, 0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}
