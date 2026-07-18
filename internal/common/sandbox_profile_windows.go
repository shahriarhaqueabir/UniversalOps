//go:build windows

package common

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	moduserenv    = windows.NewLazySystemDLL("userenv.dll")
	procDeriveSid = moduserenv.NewProc("DeriveAppContainerSidFromAppContainerName")
	procCreate    = moduserenv.NewProc("CreateAppContainerProfile")
)

// DeriveAppContainerSID generates a security identifier for a given sandbox name.
func DeriveAppContainerSID(name string) (*windows.SID, error) {
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	var sid *windows.SID
	r1, _, err := procDeriveSid.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(&sid)),
	)
	if r1 != 0 { // HRESULT S_OK is 0
		return nil, fmt.Errorf("DeriveAppContainerSid failed: 0x%x", r1)
	}
	return sid, nil
}

// EnsureAppContainerProfile ensures that an AppContainer profile exists for the app.
func EnsureAppContainerProfile(name, displayName string) (*windows.SID, error) {
	namePtr, _ := windows.UTF16PtrFromString(name)
	displayPtr, _ := windows.UTF16PtrFromString(displayName)
	descPtr, _ := windows.UTF16PtrFromString("Isolated execution environment for OpsForAll remediation tools.")

	var sid *windows.SID
	r1, _, _ := procCreate.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(displayPtr)),
		uintptr(unsafe.Pointer(descPtr)),
		0, 0,
		uintptr(unsafe.Pointer(&sid)),
	)

	const HRESULT_ALREADY_EXISTS = 0x800700B7
	if r1 != 0 && uint32(r1) != HRESULT_ALREADY_EXISTS {
		return nil, fmt.Errorf("CreateAppContainerProfile failed: 0x%x", r1)
	}

	if sid == nil {
		return DeriveAppContainerSID(name)
	}
	return sid, nil
}
