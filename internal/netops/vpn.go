package netops

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/net"
)

// VPNStatus holds the current VPN connection status.
type VPNStatus struct {
	Active    bool   `json:"active"`
	Type      string `json:"type"`
	Interface string `json:"interface"`
	RemoteIP  string `json:"remote_ip"`
	LocalIP   string `json:"local_ip"`
	Protocol  string `json:"protocol"`
}

// GetVPNStatus detects active VPN connections by checking interface names,
// tunnel adapters, and platform-specific VPN tools.
func GetVPNStatus() VPNStatus {
	status := VPNStatus{}
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			name := strings.ToLower(iface.Name)
			switch {
			case strings.Contains(name, "tun") || strings.Contains(name, "tap"):
				status.Active, status.Type, status.Interface = true, "OpenVPN/WireGuard", iface.Name
			case strings.Contains(name, "wg"):
				status.Active, status.Type, status.Interface = true, "WireGuard", iface.Name
			case strings.Contains(name, "ppp") || strings.Contains(name, "vpn"):
				status.Active, status.Type, status.Interface = true, "VPN", iface.Name
			}
			if status.Active {
				for _, addr := range iface.Addrs {
					ip := strings.Split(addr.Addr, "/")[0]
					if !strings.Contains(ip, ":") && status.LocalIP == "" {
						status.LocalIP = ip
					}
				}
				return status
			}
		}
	}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("rasdial").CombinedOutput()
		if err == nil && strings.Contains(string(out), "Connected") {
			status.Active, status.Type = true, "VPN (rasdial)"
		}
	} else if runtime.GOOS == "linux" {
		out, err := exec.Command("wg", "show").CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			status.Active, status.Type, status.Protocol = true, "WireGuard", "wg"
		}
	}
	return status
}
