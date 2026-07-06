//go:build !windows && !linux

package common

import (
	"os/exec"
)

// applyPlatformSandbox is a no-op on platforms where sandboxing is not yet implemented.
func applyPlatformSandbox(cmd *exec.Cmd, cfg SandboxConfig) {
	// Not implemented for this platform yet.
}
