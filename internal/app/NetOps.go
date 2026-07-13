package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/netops"
)

// NetOps exposes network operations bindings to the frontend.
type NetOps struct {
	app   *App
	model *netOpsModel
}

type netOpsModel struct {
	lastCounters  map[string]netops.BandwidthCounter
	lastCapture   time.Time
	lastIfaces    []InterfaceInfo
	recentChanges []NetworkChange
}

// BandwidthCounter captures a snapshot of interface byte counters.
// Re-exported from netops for JSON serialization.
type BandwidthCounter struct {
	Name    string
	RXBytes uint64
	TXBytes uint64
}

// NewNetOps creates a new NetOps facade.
func NewNetOps(app *App) *NetOps {
	return &NetOps{
		app:   app,
		model: &netOpsModel{},
	}
}

// Ping sends ICMP echo requests to a target host.
func (n *NetOps) Ping(host string, count int) PingResult {
	if count <= 0 {
		count = 4
	}
	result, err := netops.Ping(host, count)
	if err != nil {
		common.LogWarn("Ping failed: %v", err)

		n.app.eventBus.Emit(common.NewEvent(
			common.CatNetwork,
			common.EventWarning,
			"netops",
			"Ping failed",
			fmt.Sprintf("Ping to %s failed: %s", host, err.Error()),
		))

		return PingResult{Target: host, Error: err.Error()}
	}

	// Emit timeline event for significant packet loss
	if result.Lost > 0 {
		level := common.EventInfo
		if float64(result.Lost)/float64(result.Sent) > 0.5 {
			level = common.EventWarning
		}
		n.app.eventBus.Emit(common.NewEventWithMeta(
			common.CatNetwork,
			level,
			"netops",
			"Ping packet loss",
			fmt.Sprintf("Ping to %s: %d/%d packets lost (avg RTT %dms)", host, result.Lost, result.Sent, result.Avg.Milliseconds()),
			map[string]string{"host": host, "lost": fmt.Sprintf("%d", result.Lost), "sent": fmt.Sprintf("%d", result.Sent)},
		))
	}

	return PingResult{
		Target:   result.Target,
		IP:       result.IP,
		Sent:     result.Sent,
		Received: result.Received,
		Lost:     result.Lost,
		MinMs:    result.Min.Milliseconds(),
		MaxMs:    result.Max.Milliseconds(),
		AvgMs:    result.Avg.Milliseconds(),
		JitterMs: float64(result.Jitter.Microseconds()) / 1000.0,
		TTL:      result.TTL,
	}
}

// DNSLookup performs DNS lookups for a given hostname with the specified timeout (ms).
func (n *NetOps) DNSLookup(hostname string, server string, timeoutMs int) DNSResult {
	if timeoutMs <= 0 {
		timeoutMs = 2000
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	var servers []string
	if server != "" {
		servers = []string{server}
	}

	result, err := netops.LookupDNSWithContext(ctx, hostname, servers...)
	if err != nil {
		common.LogWarn("DNSLookup failed: %v", err)

		n.app.eventBus.Emit(common.NewEvent(
			common.CatNetwork,
			common.EventWarning,
			"netops",
			"DNS resolution failed",
			fmt.Sprintf("DNS lookup for %s failed: %s", hostname, err.Error()),
		))

		return DNSResult{Hostname: hostname, Error: err.Error()}
	}
	return DNSResult{
		Hostname: result.Hostname,
		A:        result.A,
		AAAA:     result.AAAA,
		MX:       result.MX,
		NS:       result.NS,
		CNAME:    result.CNAME,
		TXT:      result.TXT,
	}
}

// PortScan scans specific ports on a host.
func (n *NetOps) PortScan(host string, ports []int) []PortResult {
	if len(ports) == 0 {
		ports = netops.DefaultScanPorts()
	}
	results, err := netops.ScanPorts(host, ports)
	if err != nil {
		common.LogWarn("PortScan failed: %v", err)
		return []PortResult{}
	}
	out := make([]PortResult, 0, len(results))
	for _, r := range results {
		out = append(out, PortResult{
			Port:    r.Port,
			Open:    r.Open,
			Service: r.Service,
		})
	}
	return out
}

// Traceroute runs traceroute to a target host with a 30-second timeout.
func (n *NetOps) Traceroute(host string) TraceResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := netops.TraceRouteWithContext(ctx, host)
	if err != nil {
		common.LogWarn("Traceroute failed: %v", err)
		return TraceResult{Target: host, Error: err.Error()}
	}
	hops := make([]TraceHop, 0, len(result.Hops))
	for _, h := range result.Hops {
		rttsMs := make([]int64, 0, len(h.RTTs))
		for _, rtt := range h.RTTs {
			rttsMs = append(rttsMs, rtt.Milliseconds())
		}
		hops = append(hops, TraceHop{
			Number: h.Number,
			Host:   h.Host,
			IP:     h.IP,
			RTTsMs: rttsMs,
			Timed:  h.Timed,
		})
	}
	return TraceResult{
		Target: result.Target,
		Hops:   hops,
	}
}

// GetConnections returns current network connections.
func (n *NetOps) GetConnections() []ConnectionInfo {
	conns, err := netops.GetConnections()
	if err != nil {
		common.LogWarn("GetConnections failed: %v", err)
		return []ConnectionInfo{}
	}
	out := make([]ConnectionInfo, 0, len(conns))
	for _, c := range conns {
		out = append(out, ConnectionInfo{
			LocalAddr:   c.LocalAddr,
			RemoteAddr:  c.RemoteAddr,
			LocalPort:   c.LocalPort,
			RemotePort:  c.RemotePort,
			Protocol:    c.Protocol,
			State:       c.State,
			ProcessName: c.ProcessName,
			PID:         c.PID,
		})
	}
	return out
}

// GetInterfaces returns network interface information.
func (n *NetOps) GetInterfaces() []InterfaceInfo {
	ifaces, err := n.collectInterfaces()
	if err != nil {
		common.LogWarn("GetInterfaces failed: %v", err)
		return []InterfaceInfo{}
	}
	return ifaces
}

// maxRecentChanges is the ring-buffer capacity for network state changes.
const maxRecentChanges = 20

// collectInterfaces gathers interface data with bandwidth rate calculation
// and diffs against the previous snapshot to detect state changes.
func (n *NetOps) collectInterfaces() ([]InterfaceInfo, error) {
	elapsed := time.Since(n.model.lastCapture)

	// Use gopsutil counters directly via our netops wrapper
	result, err := netops.GetInterfaces(n.model.lastCounters, elapsed)
	if err != nil {
		return nil, err
	}

	// Store counters for next rate calculation
	n.model.lastCounters = result.Counters
	n.model.lastCapture = time.Now()

	out := make([]InterfaceInfo, 0, len(result.Interfaces))
	for _, iface := range result.Interfaces {
		out = append(out, InterfaceInfo{
			Name:      iface.Name,
			MAC:       iface.MAC,
			IPs:       iface.IPs,
			IsUp:      iface.IsUp,
			Speed:     iface.Speed,
			MTU:       iface.MTU,
			Flags:     iface.Flags,
			RXBytes:   iface.RXBytes,
			TXBytes:   iface.TXBytes,
			RXRateBps: iface.RXRateBps,
			TXRateBps: iface.TXRateBps,
			RXHistory: iface.RXHistory,
			TXHistory: iface.TXHistory,
		})
	}

	// Diff against previous snapshot to detect state changes.
	if prev := n.model.lastIfaces; len(prev) > 0 {
		n.diffInterfaces(prev, out)
	}

	n.model.lastIfaces = out
	return out, nil
}

// diffInterfaces compares two interface snapshots and appends changes to the ring buffer.
func (n *NetOps) diffInterfaces(prev, curr []InterfaceInfo) {
	prevByName := make(map[string]InterfaceInfo, len(prev))
	for _, iface := range prev {
		prevByName[iface.Name] = iface
	}

	currByName := make(map[string]InterfaceInfo, len(curr))
	for _, iface := range curr {
		currByName[iface.Name] = iface
	}

	now := time.Now().Format(time.RFC3339)

	// Check for changed or disappeared interfaces.
	for name, old := range prevByName {
		cur, exists := currByName[name]
		if !exists {
			n.appendChange(NetworkChange{
				Type:      ChangeDisappeared,
				Interface: name,
				Detail:    "Interface removed",
				Timestamp: now,
			})
			continue
		}

		// Up/down transition.
		if old.IsUp != cur.IsUp {
			ct := ChangeDown
			if cur.IsUp {
				ct = ChangeUp
			}
			n.appendChange(NetworkChange{
				Type:      ct,
				Interface: name,
				Detail:    fmt.Sprintf("State changed to %s", map[bool]string{true: "up", false: "down"}[cur.IsUp]),
				Timestamp: now,
			})
		}

		// IP address changes.
		oldIPs := make(map[string]bool, len(old.IPs))
		for _, ip := range old.IPs {
			oldIPs[ip] = true
		}
		curIPs := make(map[string]bool, len(cur.IPs))
		for _, ip := range cur.IPs {
			curIPs[ip] = true
		}
		for _, ip := range cur.IPs {
			if !oldIPs[ip] {
				n.appendChange(NetworkChange{
					Type:      ChangeIPAdded,
					Interface: name,
					Detail:    fmt.Sprintf("New IP: %s", ip),
					Timestamp: now,
				})
			}
		}
		for _, ip := range old.IPs {
			if !curIPs[ip] {
				n.appendChange(NetworkChange{
					Type:      ChangeIPRemoved,
					Interface: name,
					Detail:    fmt.Sprintf("Removed IP: %s", ip),
					Timestamp: now,
				})
			}
		}
	}

	// Check for newly appeared interfaces.
	for name := range currByName {
		if _, exists := prevByName[name]; !exists {
			n.appendChange(NetworkChange{
				Type:      ChangeAppeared,
				Interface: name,
				Detail:    "Interface appeared",
				Timestamp: now,
			})
		}
	}
}

// appendChange adds a change to the ring buffer, evicting the oldest when full.
func (n *NetOps) appendChange(c NetworkChange) {
	if len(n.model.recentChanges) >= maxRecentChanges {
		n.model.recentChanges = n.model.recentChanges[1:]
	}
	n.model.recentChanges = append(n.model.recentChanges, c)
}

// GetRecentChanges returns the last N network state changes, most recent first.
func (n *NetOps) GetRecentChanges() []NetworkChange {
	changes := n.model.recentChanges
	out := make([]NetworkChange, len(changes))
	copy(out, changes)
	// Reverse so most recent is first.
	sort.Slice(out, func(i, j int) bool { return i > j })
	return out
}

// GetNetworkSummary returns a deterministic summary of the current network state.
func (n *NetOps) GetNetworkSummary() NetworkSummary {
	ifaces := n.GetInterfaces()
	conns := n.GetConnections()
	issues := []string{}

	// Find top interface by traffic (highest TX+RX bytes)
	topIface := ""
	var topTraffic uint64
	upCount := 0
	downCount := 0
	var totalRX, totalTX uint64

	for _, iface := range ifaces {
		traffic := iface.RXBytes + iface.TXBytes
		totalRX += iface.RXBytes
		totalTX += iface.TXBytes
		if traffic > topTraffic {
			topTraffic = traffic
			topIface = iface.Name
		}
		if iface.IsUp {
			upCount++
		} else {
			downCount++
			issues = append(issues, fmt.Sprintf("%s is down", iface.Name))
		}
	}

	// Count connection states
	established := 0
	listening := 0
	for _, c := range conns {
		switch c.State {
		case "ESTABLISHED":
			established++
		case "LISTEN", "LISTENING":
			listening++
		}
	}

	// Check for high connection count as a potential issue
	if established > 200 {
		issues = append(issues, fmt.Sprintf("high established connection count: %d", established))
	}

	// Build summary text
	parts := []string{}
	parts = append(parts, fmt.Sprintf("%d/%d interfaces up.", upCount, upCount+downCount))
	if topIface != "" {
		parts = append(parts, fmt.Sprintf("Primary interface: %s.", topIface))
	}
	parts = append(parts, fmt.Sprintf("%d active connections (%d listening).", established, listening))
	parts = append(parts, fmt.Sprintf("Traffic: %s RX / %s TX.", formatBytes(totalRX), formatBytes(totalTX)))

	summaryText := strings.Join(parts, " ")

	return NetworkSummary{
		SummaryText:  summaryText,
		TopInterface: topIface,
		Issues:       issues,
	}
}

// formatBytes formats byte counts into human-readable strings.
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetDefaultGateway returns information about the system default gateway.
func (n *NetOps) GetDefaultGateway() GatewayInfo {
	gw := netops.GetDefaultGateway()
	reachable := false
	if gw.IP != "" {
		reachable = netops.CheckReachable(gw.IP)
	}
	return GatewayInfo{
		IP:        gw.IP,
		Interface: gw.Interface,
		Reachable: reachable,
	}
}
