package netops

import (
	"net"
	"strings"
)

// IsVPNActive checks if a VPN interface is active on the system.
func IsVPNActive() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}

	vpnKeywords := []string{"tun", "tap", "ppp", "wireguard", "vpn", "ipsec"}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		name := strings.ToLower(iface.Name)
		for _, kw := range vpnKeywords {
			if strings.Contains(name, kw) {
				return true
			}
		}
	}

	return false
}
