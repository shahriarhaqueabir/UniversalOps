# Hawkward — Cross-Platform

## Windows
- Go at `E:\Program Files\Go` — NOT in `$PATH`, must use full path
- SecOps Defender/tasks use `powershell.exe` — blocked in locked-down environments (no WMI/CIM fallback)
- Sandbox: basic `NoInheritHandles` + `HideWindow` only
- First-run marker: `%APPDATA%/hawkward/.onboarded`

## Linux
- Sandbox: unshare flags (`CLONE_NEWNET`, `NS`, `PID`, `USER`) — needs manual compile check
- File browser uses `os.ReadDir()` — `/proc` paths may behave unexpectedly but won't crash
- First-run marker: `~/.config/hawkward/.onboarded`

## General
- No cross-compilation CI configured
- All `exec.Command` calls use OS-native binaries
