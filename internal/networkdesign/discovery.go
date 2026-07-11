package networkdesign

import (
	"net"
	"os/exec"
	"runtime"
	"strings"

	gopsnet "github.com/shirou/gopsutil/v4/net"
)

// DiscoverLocalNetwork scans the local machine for network information and
// ARP table entries to build a list of topology nodes. This does NOT perform
// active scanning — it only reads data already available to any user.
func DiscoverLocalNetwork() []TopologyNode {
	var nodes []TopologyNode
	seenIDs := make(map[string]bool)

	// 1. Gather local interfaces via gopsutil.
	ifaces, err := gopsnet.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			for _, addr := range iface.Addrs {
				ip := parseAddr(addr.String())
				if ip == "" {
					continue
				}
				// Skip loopback.
				if strings.HasPrefix(ip, "127.") {
					continue
				}

				node := TopologyNode{
					ID:     "local-" + iface.Name,
					Type:   classifyLocalDevice(iface),
					Label:  iface.Name,
					IP:     ip,
					MAC:    iface.HardwareAddr,
					Online: true,
				}
				if !seenIDs[node.ID] {
					seenIDs[node.ID] = true
					nodes = append(nodes, node)
				}
			}
		}
	}

	// 2. Parse ARP table for neighboring devices.
	arpEntries := parseARPTable()
	for _, entry := range arpEntries {
		if entry.IP == "" {
			continue
		}
		id := "arp-" + entry.IP
		if seenIDs[id] {
			continue
		}
		seenIDs[id] = true

		vendor := lookupOUI(entry.MAC)
		node := TopologyNode{
			ID:     id,
			Type:   guessDeviceType(entry, vendor),
			Label:  entry.IP,
			IP:     entry.IP,
			MAC:    entry.MAC,
			Vendor: vendor,
			Online: entry.State == "reachable" || entry.State == "stale" || entry.State == "",
		}
		nodes = append(nodes, node)
	}

	return nodes
}

// arpEntry holds a single row from the ARP table.
type arpEntry struct {
	IP    string
	MAC   string
	State string
}

// parseARPTable runs the platform-appropriate command and parses the output.
func parseARPTable() []arpEntry {
	var out []byte
	var err error

	switch runtime.GOOS {
	case "windows":
		out, err = exec.Command("arp", "-a").Output()
	default:
		out, err = exec.Command("arp", "-an").Output()
	}

	if err != nil {
		return nil
	}
	return parseARPOutput(string(out))
}

// parseARPOutput parses the text output of `arp -a` (Windows) or `arp -an` (Linux/macOS).
func parseARPOutput(output string) []arpEntry {
	var entries []arpEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Auto-detect format: Windows lines start with an IP or interface name,
		// Linux/macOS lines typically start with "?" or an IP followed by "(".
		if strings.Contains(line, "(") && (strings.HasPrefix(line, "?") || isValidIP(fields[0])) {
			entries = append(entries, parseLinuxARPLine(line, fields)...)
		} else if isValidIP(fields[0]) {
			// Windows format: "192.168.1.1  aa-bb-cc-dd-ee-ff  dynamic"
			state := ""
			if len(fields) >= 3 {
				state = fields[2]
			}
			entries = append(entries, arpEntry{
				IP:    fields[0],
				MAC:   normalizeMAC(fields[1]),
				State: state,
			})
		}
	}
	return entries
}

// parseLinuxARPLine parses a single Linux/macOS ARP line.
func parseLinuxARPLine(line string, fields []string) []arpEntry {
	ip := ""
	mac := ""
	state := ""

	// Extract IP in parentheses.
	if idx := strings.Index(line, "("); idx >= 0 {
		end := strings.Index(line[idx:], ")")
		if end > 0 {
			ip = line[idx+1 : idx+end]
		}
	}
	// If no parentheses, try the first field as IP.
	if ip == "" && len(fields) > 0 {
		candidate := strings.TrimRight(fields[0], ":")
		if isValidIP(candidate) {
			ip = candidate
		}
	}

	// Extract MAC after "at".
	for i, f := range fields {
		if f == "at" && i+1 < len(fields) {
			mac = fields[i+1]
			break
		}
	}

	// Extract state.
	for _, f := range fields {
		switch f {
		case "lladdr", "permanent", "[ether]":
			state = f
		}
	}

	if ip != "" {
		return []arpEntry{{
			IP:    ip,
			MAC:   normalizeMAC(mac),
			State: state,
		}}
	}
	return nil
}

// parseAddr strips the CIDR suffix from an address string like "192.168.1.5/24".
func parseAddr(addr string) string {
	if idx := strings.Index(addr, "/"); idx >= 0 {
		return addr[:idx]
	}
	return addr
}

// isValidIP checks that a string is a valid IPv4 address.
func isValidIP(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

// normalizeMAC converts various MAC formats (dash-separated, dot-separated)
// to colon-separated lowercase.
func normalizeMAC(mac string) string {
	mac = strings.ToLower(mac)
	mac = strings.ReplaceAll(mac, "-", ":")
	mac = strings.ReplaceAll(mac, ".", ":")
	return mac
}

// classifyLocalDevice determines the type for a local interface.
func classifyLocalDevice(iface gopsnet.InterfaceStat) string {
	// Check if this is the default gateway interface (has gateway flag).
	name := strings.ToLower(iface.Name)
	if strings.Contains(name, "loopback") || strings.HasPrefix(name, "lo") || strings.HasPrefix(name, "lo0") {
		return "client"
	}
	// On Windows, common names: "Ethernet", "Wi-Fi", "vEthernet"
	// On Linux: "eth0", "wlan0", "enp0s3"
	// Heuristic: virtual adapters (vEthernet, docker, br-) are typically server/infrastructure.
	if strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") {
		return "server"
	}
	if strings.Contains(name, "wi-fi") || strings.Contains(name, "wlan") || strings.Contains(name, "wl") {
		return "client"
	}
	return "client" // default
}

// guessDeviceType infers device type from ARP entry and vendor info.
func guessDeviceType(entry arpEntry, vendor string) string {
	vendorLower := strings.ToLower(vendor)

	// Known router/gateway vendor patterns.
	routerVendors := []string{"cisco", "netgear", "ubiquiti", "tp-link", "linksys", "dlink", "d-link", "mikrotik", "fortinet", "juniper", "arris", "asus", "router"}
	for _, rv := range routerVendors {
		if strings.Contains(vendorLower, rv) {
			return "router"
		}
	}

	// Known switch vendor patterns.
	switchVendors := []string{"aruba", "hp procurve", "dell networking", "mellanox"}
	for _, sv := range switchVendors {
		if strings.Contains(vendorLower, sv) {
			return "switch"
		}
	}

	// Default to client (workstation) if we can't determine.
	return "workstation"
}

// lookupOUI attempts to identify a vendor from the MAC OUI prefix.
// Returns empty string if unknown.
func lookupOUI(mac string) string {
	if mac == "" {
		return ""
	}
	normalized := normalizeMAC(mac)
	prefix := strings.ReplaceAll(normalized[:8], ":", "")
	if len(prefix) < 6 {
		return ""
	}
	prefix = prefix[:6]

	// Small OUI database of common networking vendors.
	ouiMap := map[string]string{
		"005056": "VMware",
		"000c29": "VMware",
		"001c42": "Parallels",
		"080027": "Oracle VirtualBox",
		"525400": "QEMU/KVM",
		"00163e": "Xen",
		"00155d": "Hyper-V",
		"001122": "Dell",
		"f8bc12": "Dell",
		"d89ef3": "Cloud Network Technology",
		"001a2b": "Avalan Networks",
		"b42e99": "Apple",
		"a483e7": "Apple",
		"dc86d8": "Apple",
		"f0b479": "Apple",
		"001b63": "Apple",
		"002500": "Parallels",
		"3c0630": "Apple",
		"f8ff0c": "Apple",
		"0019e3": "Apple",
		"000393": "Apple",
		"7c6d62": "Apple",
		"0017f2": "Apple",
		"e0b9ba": "Apple",
		"acde48": "Private",
		"0242ac": "Docker",
		"e45f01": "Raspberry Pi Foundation",
		"28cdc1": "Raspberry Pi Foundation",
		"b827eb": "Raspberry Pi Foundation",
		"dca632": "Raspberry Pi Foundation",
		"d43d2e": "Xiaomi Communications",
		"64cc2e": "Xiaomi Communications",
		"50642b": "Samsung Electronics",
		"18227e": "Samsung Electronics",
		"c09727": "Samsung Electronics",
		"94350a": "Samsung Electronics",
		"e0cb1d": "Samsung Electronics",
		"002637": "Samsung Electronics",
		"ec1f72": "Samsung Electronics",
		"3096fb": "Samsung Electronics",
		"001e58": "Dell",
		"180373": "Dell",
		"001ec2": "Dell",
		"30d042": "HP",
		"3c4a92": "HP",
		"00237d": "HP",
		"001b78": "HP",
		"b499ba": "HP",
		"101f74": "HP",
		"0025b3": "HP",
		"001635": "HP",
		"000bcd": "Cisco",
		"001b0d": "Cisco",
		"002304": "Cisco",
		"001a2f": "Cisco",
		"e86549": "Cisco",
		"001794": "Cisco",
		"5c5015": "Cisco",
		"c46413": "Cisco",
		"0024d7": "Cisco",
		"b0aa77": "Cisco",
		"54781a": "Cisco",
		"64f69d": "Cisco",
		"f02929": "Cisco",
		"788a20": "Ubiquiti Networks",
		"fcecda": "Ubiquiti Networks",
		"0418d6": "Ubiquiti Networks",
		"802aa8": "Ubiquiti Networks",
		"245a4c": "Ubiquiti Networks",
		"b4fbe4": "Ubiquiti Networks",
		"74acac": "Ubiquiti Networks",
		"783ebf": "Netgear",
		"200cc8": "Netgear",
		"c03f0e": "Netgear",
		"4494fc": "Netgear",
		"6ca404": "Netgear",
		"b07fb9": "Netgear",
		"a42b8c": "Netgear",
		"18e829": "Netgear",
		"001f33": "Netgear",
		"cc40d0": "TP-Link Technologies",
		"50c7bf": "TP-Link Technologies",
		"14cc20": "TP-Link Technologies",
		"54c80f": "TP-Link Technologies",
		"60e327": "TP-Link Technologies",
		"18a6f7": "TP-Link Technologies",
		"e8de27": "TP-Link Technologies",
		"ec086b": "TP-Link Technologies",
		"20dce6": "TP-Link Technologies",
		"002401": "D-Link",
		"000d88": "D-Link",
		"00265a": "D-Link",
		"1c5f2f": "D-Link",
		"b8a386": "D-Link",
		"c8be19": "D-Link",
		"fc7516": "D-Link",
		"28107b": "D-Link",
		"1cb72c": "ASUSTeK",
		"04d4c4": "ASUSTeK",
		"1c872c": "ASUSTeK",
		"086266": "ASUSTeK",
		"2c56dc": "ASUSTeK",
		"f832e4": "ASUSTeK",
		"002522": "ASUSTeK",
		"c0ee40": "Lenovo",
		"507b9d": "Lenovo",
		"28d244": "Lenovo",
		"90e2ba": "Lenovo",
		"482ae3": "Lenovo",
		"e82a44": "Intel",
		"001b21": "Intel",
		"6805ca": "Intel",
		"3c970e": "Intel",

		"40ec99": "Intel",
		"a4bf01": "Intel",
		"f8db88": "Intel",
		"7c5cf8": "Intel",
		"001517": "Intel",
		"5cd2e4": "Intel",
	}

	if v, ok := ouiMap[prefix]; ok {
		return v
	}
	return ""
}
