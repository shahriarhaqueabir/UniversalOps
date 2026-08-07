//go:build windows

package main

import (
	"log"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"

	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// checkWindowsPrereqs verifies WebView2 Runtime is installed before the app
// starts. If missing, it shows a Win32 MessageBox and exits with a clear error.
func checkWindowsPrereqs() {
	if os.Getenv("UNIVERSALOPS_SKIP_PREREQS") == "1" {
		return
	}
	if !webView2Installed() {
		showFatalError(
			"WebView2 Runtime Not Found",
			"Universal-Ops requires the Microsoft WebView2 Runtime.\n\n"+
				"Download it free from:\n"+
				"https://go.microsoft.com/fwlink/p/?LinkId=2124703\n\n"+
				"After installing, restart Universal-Ops.",
		)
	}
}

// webView2Installed checks the registry for the WebView2 Runtime client key.
// The Runtime registers itself under one of two paths depending on architecture.
func webView2Installed() bool {
	keys := []string{
		`SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
		`SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
	}
	for _, p := range keys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, p, registry.READ)
		if err == nil {
			k.Close()
			return true
		}
	}
	return false
}

// backdropType returns windows.Mica on Windows 11 (build 22000+) and
// windows.Auto on older builds (Windows 10) where Mica is unsupported.
func backdropType() windows.BackdropType {
	// Use RtlGetVersion from ntdll to get the true OS build number.
	// Unlike GetVersionEx, RtlGetVersion is not shimmed and returns the
	// real build even on Windows 10+.
	dll := syscall.NewLazyDLL("ntdll.dll")
	proc := dll.NewProc("RtlGetVersion")

	// RTL_OSVERSIONINFOW structure
	type osVersionInfoExW struct {
		dwOSVersionInfoSize uint32
		dwMajorVersion      uint32
		dwMinorVersion      uint32
		dwBuildNumber       uint32
		dwPlatformID        uint32
		szCSDVersion        [128]uint16
	}
	info := osVersionInfoExW{dwOSVersionInfoSize: uint32(unsafe.Sizeof(osVersionInfoExW{}))}
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 && info.dwMajorVersion >= 10 && info.dwBuildNumber >= 22000 {
		return windows.Mica
	}
	// Windows 10 or older — Mica would render as a black/blank window.
	return windows.Auto
}

// showFatalError displays a Win32 error dialog and terminates the process.
func showFatalError(title, text string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(text)
	proc.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), 0x00000010) // MB_ICONERROR
	log.Fatalf("%s: %s", title, text)
}
