package netops

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// ── Topology Types ────────────────────────────────────────────────────────────

// DeviceType categorises a discovered or manual device.
type DeviceType string

const (
	DeviceRouter      DeviceType = "router"
	DeviceSwitch      DeviceType = "switch"
	DeviceServer      DeviceType = "server"
	DeviceWorkstation DeviceType = "workstation"
	DeviceFirewall    DeviceType = "firewall"
	DeviceCloud       DeviceType = "cloud"
	DeviceGateway     DeviceType = "gateway"
	DevicePrinter     DeviceType = "printer"
	DeviceIoT         DeviceType = "iot"
	DeviceUnknown     DeviceType = "unknown"
)

// TopologyStatus represents device health.
type TopologyStatus string

const (
	StatusHealthy  TopologyStatus = "healthy"
	StatusWarning  TopologyStatus = "warning"
	StatusCritical TopologyStatus = "critical"
)

// ConnectionType categorises a link between devices.
type ConnectionType string

const (
	ConnEthernet ConnectionType = "ethernet"
	ConnFiber    ConnectionType = "fiber"
	ConnWireless ConnectionType = "wireless"
	ConnVPN      ConnectionType = "vpn"
	ConnDirect   ConnectionType = "direct"
)

// TopologyDevice represents a node in the network topology graph.
type TopologyDevice struct {
	ID       string         `json:"id"`
	Type     DeviceType     `json:"type"`
	Label    string         `json:"label"`
	IP       string         `json:"ip,omitempty"`
	MAC      string         `json:"mac,omitempty"`
	Subnet   string         `json:"subnet,omitempty"`
	Vendor   string         `json:"vendor,omitempty"`
	Hostname string         `json:"hostname,omitempty"`
	Status   TopologyStatus `json:"status"`
	X        float64        `json:"x"`
	Y        float64        `json:"y"`
	Online   bool           `json:"online"`
	Notes    string         `json:"notes,omitempty"`
}

// TopologyConnection represents an edge between two devices.
type TopologyConnection struct {
	ID       string         `json:"id"`
	SourceID string         `json:"source_id"`
	TargetID string         `json:"target_id"`
	Type     ConnectionType `json:"type"`
	Label    string         `json:"label,omitempty"`
	Metric   int            `json:"metric,omitempty"`
}

// GraphTopology is the full graph of discovered devices and connections.
// Named GraphTopology to avoid conflict with health.go's NetworkTopology.
type GraphTopology struct {
	Devices     []TopologyDevice     `json:"devices"`
	Connections []TopologyConnection `json:"connections"`
	GeneratedAt string               `json:"generated_at"`
	Subnet      string               `json:"subnet"`
}

// DiscoveryTemplate defines a named discovery strategy.
type DiscoveryTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Commands to run during discovery
	RunPing      bool `json:"run_ping"`
	RunDNS       bool `json:"run_dns"`
	RunTrace     bool `json:"run_trace"`
	RunARP       bool `json:"run_arp"`
	RunRouting   bool `json:"run_routing"`
	RunPortScan  bool `json:"run_port_scan"`
	PingCount    int  `json:"ping_count"`
	TraceTargets int  `json:"trace_targets"`
}

// ── Topology Engine ───────────────────────────────────────────────────────────

// TopologyEngine builds and manages network topology graphs.
type TopologyEngine struct {
	mu       sync.RWMutex
	topology *GraphTopology
}

// NewTopologyEngine creates a new topology engine.
func NewTopologyEngine() *TopologyEngine {
	return &TopologyEngine{
		topology: &GraphTopology{
			Devices:     []TopologyDevice{},
			Connections: []TopologyConnection{},
		},
	}
}

// GetTopology returns the current topology graph.
func (te *TopologyEngine) GetTopology() GraphTopology {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return *te.topology
}

// SetTopology replaces the current topology.
func (te *TopologyEngine) SetTopology(t GraphTopology) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.topology = &t
}

// GetDiscoveryTemplates returns the built-in discovery templates.
func (te *TopologyEngine) GetDiscoveryTemplates() []DiscoveryTemplate {
	return []DiscoveryTemplate{
		{
			ID:          "quick-scan",
			Name:        "Quick Scan",
			Description: "ARP table + gateway — fast, no network noise. Best for initial baseline.",
			RunPing:     false,
			RunDNS:      false,
			RunTrace:    false,
			RunARP:      true,
			RunRouting:  true,
			RunPortScan: false,
		},
		{
			ID:          "ping-sweep",
			Name:        "Ping Sweep",
			Description: "ARP + ping sweep across subnet. Discovers responsive hosts with latency.",
			RunPing:     true,
			RunDNS:      true,
			RunTrace:    false,
			RunARP:      true,
			RunRouting:  true,
			RunPortScan: false,
			PingCount:   1,
		},
		{
			ID:           "full-discovery",
			Name:         "Full Discovery",
			Description:  "ARP + ping + DNS + traceroute to gateway + routing table. Maximum detail.",
			RunPing:      true,
			RunDNS:       true,
			RunTrace:     true,
			RunARP:       true,
			RunRouting:   true,
			RunPortScan:  false,
			PingCount:    2,
			TraceTargets: 1,
		},
		{
			ID:           "traceroute-scan",
			Name:         "Traceroute Scan",
			Description:  "Traceroute to common targets (gateway, 8.8.8.8, 1.1.1.1) to map network path.",
			RunPing:      false,
			RunDNS:       true,
			RunTrace:     true,
			RunARP:       true,
			RunRouting:   true,
			RunPortScan:  false,
			TraceTargets: 3,
		},
		{
			ID:           "deep-inspect",
			Name:         "Deep Inspect",
			Description:  "Full ARP + ping + DNS + traceroute + port scan on gateway. Most thorough.",
			RunPing:      true,
			RunDNS:       true,
			RunTrace:     true,
			RunARP:       true,
			RunRouting:   true,
			RunPortScan:  true,
			PingCount:    3,
			TraceTargets: 2,
		},
	}
}

// AutoDiscover runs the selected discovery template and builds a topology graph.
func (te *TopologyEngine) AutoDiscover(template DiscoveryTemplate) (*GraphTopology, error) {
	common.LogInfo("[Topology] Starting auto-discovery with template: %s", template.Name)

	subnet := GetLocalSubnet()
	gateway := GetDefaultGateway()

	topo := &GraphTopology{
		Devices:     []TopologyDevice{},
		Connections: []TopologyConnection{},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Subnet:      subnet,
	}

	deviceMap := make(map[string]*TopologyDevice) // keyed by IP
	nextID := 1
	deviceID := func() string {
		id := fmt.Sprintf("dev-%d", nextID)
		nextID++
		return id
	}
	connID := func() string {
		id := fmt.Sprintf("conn-%d", nextID)
		nextID++
		return id
	}

	addDevice := func(d TopologyDevice) {
		if d.IP != "" {
			if existing, ok := deviceMap[d.IP]; ok {
				// Merge: prefer richer data
				if d.Hostname != "" && existing.Hostname == "" {
					existing.Hostname = d.Hostname
				}
				if d.Vendor != "" && existing.Vendor == "" {
					existing.Vendor = d.Vendor
				}
				if d.MAC != "" && existing.MAC == "" {
					existing.MAC = d.MAC
				}
				return
			}
		}
		d.ID = deviceID()
		deviceMap[d.IP] = &d
		topo.Devices = append(topo.Devices, d)
	}

	// ── 1. ARP Table ──
	if template.RunARP {
		common.LogInfo("[Topology] Fetching ARP table...")
		arpEntries, err := GetARPTable()
		if err == nil {
			for _, arp := range arpEntries {
				dt := classifyDeviceType(arp.IP, arp.MAC, arp.Vendor, gateway.IP)
				status := StatusHealthy
				if arp.IP == gateway.IP {
					status = StatusHealthy
				}
				addDevice(TopologyDevice{
					IP:     arp.IP,
					MAC:    arp.MAC,
					Vendor: arp.Vendor,
					Label:  formatDeviceLabel(arp.IP, arp.Vendor, dt),
					Type:   dt,
					Status: status,
					Online: true,
					Subnet: subnet,
				})
			}
			common.LogInfo("[Topology] ARP found %d devices", len(arpEntries))
		} else {
			common.LogWarn("[Topology] ARP table error: %v", err)
		}
	}

	// ── 2. Routing Table ──
	if template.RunRouting {
		common.LogInfo("[Topology] Fetching routing table...")
		routes, err := GetRoutingTable()
		if err == nil {
			for _, r := range routes {
				if r.IsDefault && r.Gateway != "" && r.Gateway != "0.0.0.0" {
					addDevice(TopologyDevice{
						IP:     r.Gateway,
						Label:  fmt.Sprintf("Gateway (%s)", r.Gateway),
						Type:   DeviceGateway,
						Status: StatusHealthy,
						Online: true,
						Subnet: subnet,
					})
				}
			}
		}
	}

	// ── 3. Ping Sweep ──
	if template.RunPing && subnet != "" {
		common.LogInfo("[Topology] Running ping sweep on %s...", subnet)
		discoveryResult := RunNetworkDiscovery(subnet)
		for _, d := range discoveryResult.Devices {
			dt := classifyDeviceType(d.IP, d.MAC, d.Vendor, gateway.IP)
			addDevice(TopologyDevice{
				IP:       d.IP,
				MAC:      d.MAC,
				Vendor:   d.Vendor,
				Hostname: d.Hostname,
				Label:    formatDeviceLabel(d.IP, d.Vendor, dt),
				Type:     dt,
				Status:   StatusHealthy,
				Online:   true,
				Subnet:   subnet,
			})
		}
		common.LogInfo("[Topology] Ping sweep found %d devices", len(discoveryResult.Devices))
	}

	// ── 4. DNS Lookup on gateway ──
	if template.RunDNS && gateway.IP != "" {
		common.LogInfo("[Topology] DNS lookup on gateway %s...", gateway.IP)
		names, err := net.LookupAddr(gateway.IP)
		if err == nil && len(names) > 0 {
			if dev, ok := deviceMap[gateway.IP]; ok {
				dev.Hostname = strings.TrimSuffix(names[0], ".")
				dev.Label = formatDeviceLabel(gateway.IP, dev.Vendor, dev.Type)
			}
		}
	}

	// ── 5. Traceroute ──
	if template.RunTrace {
		targets := []string{}
		if gateway.IP != "" {
			targets = append(targets, gateway.IP)
		}
		if template.TraceTargets >= 2 {
			targets = append(targets, "8.8.8.8")
		}
		if template.TraceTargets >= 3 {
			targets = append(targets, "1.1.1.1")
		}

		for _, target := range targets {
			common.LogInfo("[Topology] Traceroute to %s...", target)
			traceResult, err := TraceRoute(target)
			if err != nil {
				common.LogWarn("[Topology] Traceroute to %s failed: %v", target, err)
				continue
			}
			for _, hop := range traceResult.Hops {
				if hop.IP == "" || hop.IP == "*" || hop.Timed {
					continue
				}
				dt := DeviceRouter
				if hop.IP == gateway.IP {
					dt = DeviceGateway
				}
				addDevice(TopologyDevice{
					IP:     hop.IP,
					Label:  formatDeviceLabel(hop.IP, "", dt),
					Type:   dt,
					Status: StatusHealthy,
					Online: true,
					Subnet: subnet,
				})
			}
		}
	}

	// ── 6. Port Scan on gateway ──
	if template.RunPortScan && gateway.IP != "" {
		common.LogInfo("[Topology] Port scan on gateway %s...", gateway.IP)
		ports, err := ScanPorts(gateway.IP, []int{22, 80, 443, 3389, 8080, 8443})
		if err != nil {
			common.LogWarn("[Topology] Port scan error: %v", err)
		} else {
			openPorts := []string{}
			for _, p := range ports {
				if p.Open {
					openPorts = append(openPorts, fmt.Sprintf("%d/%s", p.Port, p.Service))
				}
			}
			if len(openPorts) > 0 {
				if dev, ok := deviceMap[gateway.IP]; ok {
					dev.Notes = fmt.Sprintf("Open ports: %s", strings.Join(openPorts, ", "))
				}
			}
		}
	}

	// ── Build Connections ──
	// Connect gateway to all directly reachable devices
	if gw, ok := deviceMap[gateway.IP]; ok {
		for _, dev := range deviceMap {
			if dev.IP == gateway.IP {
				continue
			}
			topo.Connections = append(topo.Connections, TopologyConnection{
				ID:       connID(),
				SourceID: gw.ID,
				TargetID: dev.ID,
				Type:     ConnEthernet,
				Label:    "LAN",
			})
		}
	}

	// Add connections from routing table (non-default routes)
	if template.RunRouting {
		routes, err := GetRoutingTable()
		if err == nil {
			for _, r := range routes {
				if r.IsDefault || r.Gateway == "" || r.Gateway == "0.0.0.0" {
					continue
				}
				if src, ok := deviceMap[r.Gateway]; ok {
					// Find a device on that destination subnet
					for _, dev := range deviceMap {
						if dev.IP == r.Gateway {
							continue
						}
						if isInSubnet(dev.IP, r.Destination, r.Mask) {
							topo.Connections = append(topo.Connections, TopologyConnection{
								ID:       connID(),
								SourceID: src.ID,
								TargetID: dev.ID,
								Type:     ConnEthernet,
								Label:    fmt.Sprintf("route/%s", r.Destination),
								Metric:   r.Metric,
							})
						}
					}
				}
			}
		}
	}

	// Layout: auto-position devices in a radial layout around the gateway
	autoLayout(topo, gateway.IP)

	te.mu.Lock()
	te.topology = topo
	te.mu.Unlock()

	common.LogInfo("[Topology] Auto-discovery complete: %d devices, %d connections",
		len(topo.Devices), len(topo.Connections))

	return topo, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func classifyDeviceType(ip, mac, vendor, gatewayIP string) DeviceType {
	if ip == gatewayIP {
		return DeviceGateway
	}
	vendorLower := strings.ToLower(vendor)
	if strings.Contains(vendorLower, "cisco") ||
		strings.Contains(vendorLower, "juniper") ||
		strings.Contains(vendorLower, "mikrotik") ||
		strings.Contains(vendorLower, "ubiquiti") ||
		strings.Contains(vendorLower, "fortinet") ||
		strings.Contains(vendorLower, "aruba") ||
		strings.Contains(vendorLower, "meraki") {
		return DeviceRouter
	}
	if strings.Contains(vendorLower, "hp") ||
		strings.Contains(vendorLower, "netgear") ||
		strings.Contains(vendorLower, "tp-link") ||
		strings.Contains(vendorLower, "d-link") ||
		strings.Contains(vendorLower, "dell") {
		return DeviceSwitch
	}
	if strings.Contains(vendorLower, "vmware") ||
		strings.Contains(vendorLower, "microsoft") ||
		strings.Contains(vendorLower, "synology") ||
		strings.Contains(vendorLower, "qnap") {
		return DeviceServer
	}
	if strings.Contains(vendorLower, "raspberry") ||
		strings.Contains(vendorLower, "intel") ||
		strings.Contains(vendorLower, "samsung") {
		return DeviceWorkstation
	}
	// Default: workstation for unknown
	return DeviceWorkstation
}

func formatDeviceLabel(ip, vendor string, dt DeviceType) string {
	label := ip
	if vendor != "" && vendor != "Unknown" {
		label = fmt.Sprintf("%s (%s)", vendor, ip)
	}
	return label
}

func isInSubnet(ipStr, dest, mask string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	destIP := net.ParseIP(dest)
	if destIP == nil {
		return false
	}
	maskIP := net.IPMask(net.ParseIP(mask).To4())
	if maskIP == nil {
		// Try CIDR notation
		if strings.Contains(dest, "/") {
			_, cidr, err := net.ParseCIDR(dest)
			if err != nil {
				return false
			}
			return cidr.Contains(ip)
		}
		return false
	}
	return destIP.Mask(maskIP).Equal(ip.Mask(maskIP))
}

func autoLayout(topo *GraphTopology, gatewayIP string) {
	const (
		centerX = 400.0
		centerY = 300.0
		ring1   = 120.0
		ring2   = 250.0
		ring3   = 380.0
	)

	// Place gateway at center
	var gatewayDev *TopologyDevice
	var others []*TopologyDevice
	for i := range topo.Devices {
		dev := &topo.Devices[i]
		if dev.IP == gatewayIP {
			gatewayDev = dev
		} else {
			others = append(others, dev)
		}
	}

	if gatewayDev != nil {
		gatewayDev.X = centerX
		gatewayDev.Y = centerY
	}

	// Place routers/switches on ring 1
	var ring1Devs, ring2Devs, ring3Devs []*TopologyDevice
	for _, dev := range others {
		switch dev.Type {
		case DeviceRouter, DeviceGateway, DeviceFirewall, DeviceSwitch:
			ring1Devs = append(ring1Devs, dev)
		case DeviceServer:
			ring2Devs = append(ring2Devs, dev)
		default:
			ring3Devs = append(ring3Devs, dev)
		}
	}

	placeOnRing(ring1Devs, centerX, centerY, ring1)
	placeOnRing(ring2Devs, centerX, centerY, ring2)
	placeOnRing(ring3Devs, centerX, centerY, ring3)
}

func placeOnRing(devs []*TopologyDevice, cx, cy, radius float64) {
	n := len(devs)
	if n == 0 {
		return
	}
	for i, dev := range devs {
		angle := (float64(i) / float64(n)) * 2 * 3.14159
		dev.X = cx + radius*cos(angle)
		dev.Y = cy + radius*sin(angle)
	}
}

func cos(a float64) float64 {
	return 1 - a*a/2 + a*a*a*a/24 - a*a*a*a*a*a/720
}

func sin(a float64) float64 {
	return a - a*a*a/6 + a*a*a*a*a/120
}

// SortDevices sorts devices by type then IP for consistent display.
func SortDevices(devices []TopologyDevice) {
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Type != devices[j].Type {
			return devices[i].Type < devices[j].Type
		}
		return devices[i].IP < devices[j].IP
	})
}
