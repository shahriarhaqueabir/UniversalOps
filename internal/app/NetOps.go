package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/shahriarhaqueabir/UniversalOps/internal/netops"
	"github.com/shahriarhaqueabir/UniversalOps/internal/secops"
)

// NetOps exposes network operations bindings to the frontend.
type NetOps struct {
	eventBus *common.EventBus
	model    *netOpsModel
}

type netOpsModel struct {
	mu            sync.RWMutex
	lastCounters  map[string]netops.BandwidthCounter
	lastCapture   time.Time
	lastIfaces    []InterfaceInfo
	speedCache    map[string]int64 // caches link speeds to avoid PowerShell on every call
	speedCacheAge time.Time
	recentChanges []NetworkChange
}

// BandwidthCounter captures a snapshot of interface byte counters.
// Re-exported from netops for JSON serialization.
type BandwidthCounter struct {
	Name    string
	RXBytes uint64
	TXBytes uint64
}

// NewNetOps creates a NetOps facade and seeds initial bandwidth counters
// so the first frontend call already has a baseline for rate calculation.
func NewNetOps(eventBus *common.EventBus) *NetOps {
	n := &NetOps{
		eventBus: eventBus,
		model:    &netOpsModel{},
	}
	// Seed initial counters so the first GetInterfaces() call can compute rates
	counters, err := netops.GetBandwidthCounters()
	if err == nil {
		n.model.mu.Lock()
		n.model.lastCounters = counters
		n.model.lastCapture = time.Now()
		n.model.mu.Unlock()
	}
	return n
}

// Ping sends ICMP echo requests to a target host.
func (n *NetOps) Ping(host string, count int) PingResult {
	if count <= 0 {
		count = 4
	}
	result, err := netops.Ping(host, count)
	if err != nil {
		common.LogDebug("NetOps: Ping(%q) failed: %v", host, err)
		if n.eventBus != nil {
			n.eventBus.Emit(common.NewEvent(
				common.CatNetwork,
				common.EventWarning,
				"netops",
				"Ping failed",
				fmt.Sprintf("Ping to %s failed: %s", host, err.Error()),
			))
		}

		return PingResult{Target: host, Error: err.Error()}
	}

	// Emit timeline event for significant packet loss
	if result.Lost > 0 {
		level := common.EventInfo
		if float64(result.Lost)/float64(result.Sent) > 0.5 {
			level = common.EventWarning
		}
		if n.eventBus != nil {
			n.eventBus.Emit(common.NewEventWithMeta(
				common.CatNetwork,
				level,
				"netops",
				"Ping packet loss",
				fmt.Sprintf("Ping to %s: %d/%d packets lost (avg RTT %dms)", host, result.Lost, result.Sent, result.Avg.Milliseconds()),
				map[string]string{"host": host, "lost": fmt.Sprintf("%d", result.Lost), "sent": fmt.Sprintf("%d", result.Sent)},
			))
		}
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
		common.LogDebug("NetOps: DNSLookup(%q) failed: %v", hostname, err)
		if n.eventBus != nil {
			n.eventBus.Emit(common.NewEvent(
				common.CatNetwork,
				common.EventWarning,
				"netops",
				"DNS resolution failed",
				fmt.Sprintf("DNS lookup for %s failed: %s", hostname, err.Error()),
			))
		}

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
		common.LogDebug("NetOps: PortScan(%q) failed: %v", host, err)
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

	result, err := netops.TraceRouteWithContext(ctx, host, 0)
	if err != nil {
		common.LogDebug("NetOps: Traceroute(%q) failed: %v", host, err)
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
		common.LogDebug("NetOps: GetConnections failed: %v", err)
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
// Maintains an internal speed cache (refreshed every 5 min) to avoid
// spawning PowerShell for link speeds on every frontend poll.
func (n *NetOps) GetInterfaces() []InterfaceInfo {
	n.model.mu.Lock()
	defer n.model.mu.Unlock()

	// Refresh speed cache every 5 minutes
	if n.model.speedCache == nil || time.Since(n.model.speedCacheAge) > 5*time.Minute {
		fresh := netops.GetLinkSpeeds()
		n.model.speedCache = fresh
		n.model.speedCacheAge = time.Now()
	}
	ifaces, err := n.collectInterfaces(n.model.speedCache)
	if err != nil {
		common.LogDebug("NetOps: collectInterfaces failed: %v", err)
		return []InterfaceInfo{}
	}
	return ifaces
}

// collectInterfaces gathers interface data with bandwidth rate calculation
// and diffs against the previous snapshot to detect state changes.
func (n *NetOps) collectInterfaces(cachedSpeeds map[string]int64) ([]InterfaceInfo, error) {
	elapsed := time.Since(n.model.lastCapture)

	// Use gopsutil counters directly via our netops wrapper
	// Pass cached link speeds to avoid PowerShell on every tick
	result, err := netops.GetInterfaces(n.model.lastCounters, elapsed, cachedSpeeds)
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

	n.model.lastIfaces = out
	return out, nil
}

// RunNetworkHealthCheck runs a comprehensive network health check and returns the report.
// This exposes the domain-level netops.RunNetworkHealthCheck via Wails IPC.
func (n *NetOps) RunNetworkHealthCheck() NetworkHealthReport {
	report := netops.RunNetworkHealthCheck()

	checks := make([]NetworkHealthCheck, 0, len(report.Checks))
	for _, c := range report.Checks {
		checks = append(checks, NetworkHealthCheck{
			Name:   c.Name,
			Status: c.Status,
			Detail: c.Detail,
			Score:  c.Score,
		})
	}

	return NetworkHealthReport{
		Score:    report.Score,
		Checks:   checks,
		Summary:  report.Summary,
		Duration: report.Duration,
	}
}

// PingMultiTarget pings multiple targets concurrently and returns results.
func (n *NetOps) PingMultiTarget(targets []string, count int) []PingResultMultiData {
	if len(targets) == 0 {
		return []PingResultMultiData{}
	}
	if count <= 0 {
		count = 4
	}
	results := netops.PingMultiTarget(targets, count)
	out := make([]PingResultMultiData, 0, len(results))
	for _, r := range results {
		out = append(out, PingResultMultiData{
			Target:         r.Target,
			MinMs:          r.MinMs,
			AvgMs:          r.AvgMs,
			MaxMs:          r.MaxMs,
			StdDevMs:       r.StdDevMs,
			PacketLoss:     r.PacketLoss,
			JitterMs:       r.JitterMs,
			IndividualRTTs: r.IndividualRTTs,
			Success:        r.Success,
			Error:          r.Error,
		})
	}
	return out
}

// GetPingStats computes aggregate stats across multiple ping results.
func (n *NetOps) GetPingStats(results []PingResultMultiData) PingStatsData {
	if len(results) == 0 {
		return PingStatsData{}
	}
	// Convert to netops.PingResultMulti for the domain function
	domain := make([]netops.PingResultMulti, 0, len(results))
	for _, r := range results {
		domain = append(domain, netops.PingResultMulti{
			Target:         r.Target,
			MinMs:          r.MinMs,
			AvgMs:          r.AvgMs,
			MaxMs:          r.MaxMs,
			StdDevMs:       r.StdDevMs,
			PacketLoss:     r.PacketLoss,
			JitterMs:       r.JitterMs,
			IndividualRTTs: r.IndividualRTTs,
			Success:        r.Success,
			Error:          r.Error,
		})
	}
	stats := netops.GetPingStats(domain)
	return PingStatsData{
		AvgLatency:  stats.AvgLatency,
		MaxLatency:  stats.MaxLatency,
		TotalLoss:   stats.TotalLoss,
		WorstTarget: stats.WorstTarget,
	}
}

// ── NetOps Extended Methods ──────────────────────────────────────────────────

// GetARPTable returns the system ARP table.
func (n *NetOps) GetARPTable() []netops.ARPEntry {
	entries, err := netops.GetARPTable()
	if err != nil {
		return []netops.ARPEntry{}
	}
	if entries == nil {
		return []netops.ARPEntry{}
	}
	return entries
}

// GetRoutingTable returns the system routing table.
func (n *NetOps) GetRoutingTable() []netops.RouteEntry {
	entries, err := netops.GetRoutingTable()
	if err != nil {
		return []netops.RouteEntry{}
	}
	if entries == nil {
		return []netops.RouteEntry{}
	}
	return entries
}

// GetFirewallRules returns firewall rules for the NetOps frontend.
// Converts from secops.FirewallRule to NetOpsFirewallRuleData.
func (n *NetOps) GetFirewallRules() []NetOpsFirewallRuleData {
	rules, err := secops.GetFirewallRules()
	if err != nil {
		return []NetOpsFirewallRuleData{}
	}
	out := make([]NetOpsFirewallRuleData, 0, len(rules))
	for _, r := range rules {
		ports := r.LocalPort
		if r.RemotePort != "" && r.RemotePort != r.LocalPort {
			ports = r.LocalPort + "->" + r.RemotePort
		}
		out = append(out, NetOpsFirewallRuleData{
			Name:        r.Name,
			Direction:   r.Direction,
			Action:      r.Action,
			Protocol:    r.Protocol,
			Ports:       ports,
			Enabled:     r.Enabled,
			Source:      r.RemoteIP,
			Destination: "",
		})
	}
	return out
}

// GetVPNStatus returns the current VPN connection status.
func (n *NetOps) GetVPNStatus() VPNStatusData {
	active := netops.IsVPNActive()
	return VPNStatusData{
		Active:    active,
		Type:      "",
		Interface: "",
		RemoteIP:  "",
		LocalIP:   "",
		Protocol:  "",
	}
}

// FlushDNSCache flushes the system DNS resolver cache.
func (n *NetOps) FlushDNSCache() map[string]any {
	err := netops.FlushDNSCache()
	if err != nil {
		return map[string]any{"success": false, "message": "", "error": err.Error()}
	}
	return map[string]any{"success": true, "message": "DNS cache flushed successfully", "error": ""}
}

// ReverseLookup performs a reverse DNS PTR lookup for the given IP address.
func (n *NetOps) ReverseLookup(ip string) map[string]any {
	hostname, err := netops.ReverseLookup(ip)
	if err != nil {
		return map[string]any{"hostname": "", "error": err.Error()}
	}
	return map[string]any{"hostname": hostname, "error": ""}
}

// TestDoH tests DNS-over-HTTPS connectivity to the specified server.
func (n *NetOps) TestDoH(server string) DoHResultData {
	result := netops.TestDoH(server)
	return DoHResultData{
		Server:     result.Server,
		LatencyMs:  result.LatencyMs,
		Success:    result.Success,
		ResolvedIP: result.ResolvedIP,
	}
}

// GetRecentChanges returns recent network interface state changes.
func (n *NetOps) GetRecentChanges() []NetworkChange {
	n.model.mu.RLock()
	defer n.model.mu.RUnlock()
	if len(n.model.recentChanges) == 0 {
		return []NetworkChange{}
	}
	out := make([]NetworkChange, len(n.model.recentChanges))
	copy(out, n.model.recentChanges)
	return out
}

// ScanWiFiNetworks scans for visible WiFi access points.
func (n *NetOps) ScanWiFiNetworks() []netops.WiFiNetwork {
	networks, err := netops.ScanWiFiNetworks()
	if err != nil {
		return []netops.WiFiNetwork{}
	}
	if networks == nil {
		return []netops.WiFiNetwork{}
	}
	return networks
}

// GetWiFiInfo returns the current WiFi adapter connection state.
func (n *NetOps) GetWiFiInfo() netops.WiFiInfo {
	info, err := netops.GetWiFiInfo()
	if err != nil {
		return netops.WiFiInfo{}
	}
	return info
}

// RunNetworkDiscovery performs a network discovery scan on the given subnet.
func (n *NetOps) RunNetworkDiscovery(subnet string) netops.DiscoveryResult {
	return netops.RunNetworkDiscovery(subnet)
}

// RunNetworkAction executes a named network action with the given parameters.
func (n *NetOps) RunNetworkAction(action string, params map[string]string) NetworkActionResult {
	err := netops.RunNetworkAction(action, params)
	if err != nil {
		return NetworkActionResult{
			Action:  action,
			Message: err.Error(),
			Success: false,
		}
	}
	return NetworkActionResult{
		Action:  action,
		Message: "Action completed successfully",
		Success: true,
	}
}

// GetDefaultGateway returns the system default gateway information.
func (n *NetOps) GetDefaultGateway() GatewayInfo {
	gw := netops.GetDefaultGateway()
	return GatewayInfo{
		IP:        gw.IP,
		Interface: gw.Interface,
		Reachable: gw.Reachable,
	}
}

// GetNetworkSummary returns an aggregated summary of the current network state.
func (n *NetOps) GetNetworkSummary() NetworkSummary {
	ifaces := n.GetInterfaces()
	changes := n.GetRecentChanges()
	summary := NetworkSummary{
		SummaryText:  "Network is operational",
		TopInterface: "",
		Issues:       []string{},
	}
	if len(ifaces) > 0 {
		summary.TopInterface = ifaces[0].Name
	}
	var issueStrs []string
	for _, ch := range changes {
		if ch.Type == ChangeDown || ch.Type == ChangeDisappeared {
			issueStrs = append(issueStrs, ch.Interface+": "+string(ch.Type))
		}
	}
	if len(issueStrs) > 0 {
		summary.Issues = issueStrs
		summary.SummaryText = "Network has active issues"
	}
	return summary
}
