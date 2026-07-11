package netops

import (
	"bufio"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// GatewayInfo holds information about the default gateway.
type GatewayInfo struct {
	IP        string
	Interface string
	Reachable bool
}

// GetDefaultGateway detects the system default gateway.
func GetDefaultGateway() GatewayInfo {
	if runtime.GOOS == "windows" {
		return getGatewayWindows()
	}
	return getGatewayLinux()
}

// getGatewayWindows parses the default gateway from netstat -rn output.
func getGatewayWindows() GatewayInfo {
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false
	cmd := common.SandboxedCommandWithConfig(cfg, "netstat", "-rn")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return GatewayInfo{}
	}

	return parseWindowsRouteTable(string(output))
}

// parseWindowsRouteTable extracts the default gateway from netstat -rn output.
// The default route has destination 0.0.0.0 (or blank/default line after "Active Routes:").
func parseWindowsRouteTable(output string) GatewayInfo {
	scanner := bufio.NewScanner(strings.NewReader(output))
	inActiveRoutes := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "Active Routes") {
			inActiveRoutes = true
			// Skip the column header lines (Network Destination, Netmask, Gateway, Interface, Metric)
			continue
		}

		if !inActiveRoutes {
			continue
		}

		// Skip empty lines and header lines
		if line == "" || strings.HasPrefix(line, "Network Dest") || strings.HasPrefix(line, "---") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Default route: Network Destination 0.0.0.0
		if fields[0] == "0.0.0.0" {
			gatewayIP := fields[2]
			iface := ""
			if len(fields) >= 4 {
				iface = fields[3]
			}

			return GatewayInfo{
				IP:        gatewayIP,
				Interface: iface,
			}
		}
	}

	return GatewayInfo{}
}

// getGatewayLinux parses the default gateway from /proc/net/route.
func getGatewayLinux() GatewayInfo {
	// /proc/net/route format:
	// Iface   Destination     Gateway         Flags   RefCnt  Use     Metric  Mask            MTU     Window  IRTT
	// eth0    00000000        0101A8C0        0003    0       0       100     00000000        0       0       0
	//
	// Destination and Gateway are in hex, little-endian byte order for each 4-byte group.
	cfg := common.SystemQuerySandbox()
	cfg.DenyNetworkAccess = false
	cmd := common.SandboxedCommandWithConfig(cfg, "cat", "/proc/net/route")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: use ip route
		cmd2 := common.SandboxedCommandWithConfig(cfg, "ip", "route")
		output2, err2 := cmd2.CombinedOutput()
		if err2 != nil {
			return GatewayInfo{}
		}
		return parseIPRoute(string(output2))
	}

	return parseProcNetRoute(string(output))
}

// parseProcNetRoute extracts the default gateway from /proc/net/route.
func parseProcNetRoute(output string) GatewayInfo {
	scanner := bufio.NewScanner(strings.NewReader(output))
	firstLine := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if firstLine {
			firstLine = false
			continue // skip header
		}
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		iface := fields[0]
		destHex := fields[1]
		gwHex := fields[2]

		// Default route has destination 00000000 (0.0.0.0)
		if destHex == "00000000" {
			gatewayIP := hexToIPv4(gwHex)
			if gatewayIP == "" {
				continue
			}
			return GatewayInfo{
				IP:        gatewayIP,
				Interface: iface,
			}
		}
	}

	return GatewayInfo{}
}

// parseIPRoute parses the output of `ip route` for the default route.
// Example line: "default via 192.168.1.1 dev eth0 proto dhcp metric 100"
func parseIPRoute(output string) GatewayInfo {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "default") {
			continue
		}

		fields := strings.Fields(line)
		gw := ""
		iface := ""
		for i, f := range fields {
			if f == "via" && i+1 < len(fields) {
				gw = fields[i+1]
			}
			if f == "dev" && i+1 < len(fields) {
				iface = fields[i+1]
			}
		}
		if gw != "" {
			return GatewayInfo{IP: gw, Interface: iface}
		}
	}

	return GatewayInfo{}
}

// hexToIPv4 converts a hex-encoded gateway address from /proc/net/route
// to a dotted-quad string. The hex is in little-endian byte order per 4-byte group.
// Example: "0101A8C0" → "192.168.1.1"
func hexToIPv4(hex string) string {
	if len(hex) != 8 {
		return ""
	}
	a := hexByte(hex[6:8])
	b := hexByte(hex[4:6])
	c := hexByte(hex[2:4])
	d := hexByte(hex[0:2])
	if a == 0 && b == 0 && c == 0 && d == 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

// hexByte converts two hex characters to a uint8.
func hexByte(s string) uint8 {
	var v uint8
	for i := 0; i < 2 && i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			v = v*16 + uint8(c-'0')
		case c >= 'a' && c <= 'f':
			v = v*16 + uint8(c-'a'+10)
		case c >= 'A' && c <= 'F':
			v = v*16 + uint8(c-'A'+10)
		}
	}
	return v
}

// CheckReachable does a quick TCP dial to check if an IP is reachable.
// It tries common ports (80, 443) with a short timeout.
func CheckReachable(ip string) bool {
	// Try a quick TCP connect to a common port
	ports := []int{80, 443, 53}
	for _, port := range ports {
		addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}
