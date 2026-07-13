# Hawkward — Sandbox

## Current State
- `internal/common/sandbox.go` defines `SandboxedCommand()` and `SandboxedCommandWithConfig()` — unified interface
- **Not yet wired** into DevOps shell or any other caller

## Platform Details
- **Windows** (`sandbox_windows.go`): Basic `NoInheritHandles` + `HideWindow` only. True restricted-token sandboxing needs `golang.org/x/sys/windows` for `CreateRestrictedToken`.
- **Linux** (`sandbox_linux.go`): Uses `CLONE_NEWNET`, `CLONE_NEWNS`, `CLONE_NEWPID`, `CLONE_NEWUSER` unshare flags.

## Next
- Wire sandbox into DevOps `RunCommand()` / `RunCommandWithLiveOutput()`
- Add graceful degradation: if sandbox init fails, fall back to `exec.Command`
