//go:build windows

package common

import (
	"fmt"
	"testing"
)

func TestPingIntegrity(t *testing.T) {
	// Try Ping with Low Integrity (ReadOnlyFS = true)
	cfgLow := SystemQuerySandbox()
	cfgLow.DenyNetworkAccess = false
	cfgLow.ReadOnlyFS = true

	cmdLow := SandboxedCommandWithConfig(cfgLow, "ping", "-n", "1", "127.0.0.1")
	outLow, errLow := cmdLow.CombinedOutput()
	fmt.Printf("Ping Low Integrity: err=%v, out=%s\n", errLow, string(outLow))

	// Try Ping with Medium Integrity (ReadOnlyFS = false)
	cfgMed := SystemQuerySandbox()
	cfgMed.DenyNetworkAccess = false
	cfgMed.ReadOnlyFS = false

	cmdMed := SandboxedCommandWithConfig(cfgMed, "ping", "-n", "1", "127.0.0.1")
	outMed, errMed := cmdMed.CombinedOutput()
	fmt.Printf("Ping Medium Integrity: err=%v, out=%s\n", errMed, string(outMed))
}

func TestNetshIntegrity(t *testing.T) {
	// Try netsh with Low Integrity
	cfgLow := SystemQuerySandbox()
	cfgLow.ReadOnlyFS = true

	cmdLow := SandboxedCommandWithConfig(cfgLow, "netsh", "advfirewall", "firewall", "show", "rule", "name=all")
	outLow, errLow := cmdLow.CombinedOutput()
	fmt.Printf("Netsh Low Integrity: err=%v, outLen=%d\n", errLow, len(outLow))

	// Try netsh with Medium Integrity
	cfgMed := SystemQuerySandbox()
	cfgMed.ReadOnlyFS = false

	cmdMed := SandboxedCommandWithConfig(cfgMed, "netsh", "advfirewall", "firewall", "show", "rule", "name=all")
	outMed, errMed := cmdMed.CombinedOutput()
	fmt.Printf("Netsh Medium Integrity: err=%v, outLen=%d\n", errMed, len(outMed))
}
