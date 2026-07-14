package netops

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// RunNetworkAction executes a named network action with optional parameters.
func RunNetworkAction(action string, params map[string]string) error {
	iface := params["interface"]
	if iface != "" && !common.ValidInterfaceName(iface) {
		return fmt.Errorf("invalid interface name: %q", iface)
	}
	switch action {
	case "flush_dns":
		return FlushDNSCache()
	case "renew_dhcp":
		return renewDHCP(iface)
	case "reset_interface":
		return resetInterface(iface)
	case "disable_interface":
		return setInterfaceState(iface, false)
	case "enable_interface":
		return setInterfaceState(iface, true)
	case "clear_arp_cache":
		return clearARPCache()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func renewDHCP(iface string) error {
	switch runtime.GOOS {
	case "windows":
		if iface == "" {
			iface = "*"
		}
		_, err := exec.Command("ipconfig", "/renew", iface).CombinedOutput()
		return err
	case "linux":
		if iface == "" {
			return fmt.Errorf("interface name required")
		}
		_, err := exec.Command("sudo", "dhclient", "-r", iface).CombinedOutput()
		return err
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func resetInterface(iface string) error {
	if iface == "" {
		return fmt.Errorf("interface name required")
	}
	if err := setInterfaceState(iface, false); err != nil {
		return err
	}
	cmd := exec.Command("sleep", "2")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("timeout", "2", "cmd", "/c", "echo.")
	}
	cmd.Run()
	return setInterfaceState(iface, true)
}

func setInterfaceState(iface string, enable bool) error {
	if iface == "" {
		return fmt.Errorf("interface name required")
	}
	state := "enable"
	if !enable {
		state = "disable"
	}
	switch runtime.GOOS {
	case "windows":
		_, err := exec.Command("netsh", "interface", "set", "interface", "name="+iface, state).CombinedOutput()
		return err
	case "linux":
		action := "up"
		if !enable {
			action = "down"
		}
		_, err := exec.Command("sudo", "ip", "link", "set", iface, action).CombinedOutput()
		return err
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func clearARPCache() error {
	switch runtime.GOOS {
	case "windows":
		_, err := exec.Command("netsh", "interface", "ipv4", "delete", "arpcache").CombinedOutput()
		return err
	case "linux":
		_, err := exec.Command("sudo", "ip", "neigh", "flush", "all").CombinedOutput()
		return err
	default:
		return fmt.Errorf("unsupported platform")
	}
}
