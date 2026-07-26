package netops

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/shirou/gopsutil/v4/net"
)

// HealthCheck holds a single health check result.
type HealthCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "warn", "fail"
	Detail string `json:"detail"`
	Score  int    `json:"score"`
}

// HealthReport holds the full network health check report.
type HealthReport struct {
	Score    int           `json:"score"`
	Checks   []HealthCheck `json:"checks"`
	Summary  string        `json:"summary"`
	Duration string        `json:"duration"`
}

// NetworkTopology represents the high-level mapping of the local network segment.
type NetworkTopology struct {
	Gateway   GatewayInfo `json:"gateway"`
	DNS       []string    `json:"dns"`
	Neighbors []ARPEntry  `json:"neighbors"`
	IsVPN     bool        `json:"is_vpn"`
	PublicIP  string      `json:"public_ip"`
}

// GetTopologyContext gathers a high-fidelity map of the immediate network environment.
func GetTopologyContext() *NetworkTopology {
	topo := &NetworkTopology{
		Gateway: GetDefaultGateway(),
	}

	// 1. Get Neighbors
	if arp, err := GetARPTable(); err == nil {
		topo.Neighbors = arp
	}

	// 2. Get VPN status
	topo.IsVPN = IsVPNActive()

	// 3. Discover system DNS servers
	if sysDNS := GetSystemDNSServers(); len(sysDNS) > 0 {
		topo.DNS = sysDNS
	} else {
		// Fallback to well-known public DNS when discovery fails
		topo.DNS = []string{"8.8.8.8", "1.1.1.1"}
	}

	return topo
}

// NetworkForecast holds long-term bandwidth predictions.
type NetworkForecast struct {
	CurrentRX      float64 `json:"current_rx"`
	CurrentTX      float64 `json:"current_tx"`
	PredictedMax   float64 `json:"predicted_max_24h"`
	SaturationRisk string  `json:"saturation_risk"` // "low", "med", "high"
}

// GetLongTermForecast analyzes historical bandwidth to predict 24h trends.
func GetLongTermForecast() *NetworkForecast {
	s := common.GetStorage()
	if s == nil {
		return nil
	}

	// Analyze last 24h of throughput (approx 17k points at 5s intervals, but we take 1000 samples)
	rxHistory, _ := s.GetMetricHistory(common.MetricNetRX, 1000)
	txHistory, _ := s.GetMetricHistory(common.MetricNetTX, 1000)

	if len(rxHistory) < 100 {
		return nil
	}

	engine := common.NewForecastEngine(len(rxHistory))
	for _, v := range rxHistory {
		engine.Push(v)
	}

	lastRX := rxHistory[len(rxHistory)-1]
	lastTX := float64(0)
	if len(txHistory) > 0 {
		lastTX = txHistory[len(txHistory)-1]
	}

	// Predict 24h ahead (assuming 1 tick = 5s, 24h = 17280 ticks)
	prediction := engine.Predict(17280)

	risk := "low"
	if prediction > lastRX*2 && prediction > 1e7 { // 2x growth and > 10Mbps
		risk = "med"
	}
	if prediction > lastRX*5 {
		risk = "high"
	}

	return &NetworkForecast{
		CurrentRX:      lastRX,
		CurrentTX:      lastTX,
		PredictedMax:   prediction,
		SaturationRisk: risk,
	}
}

// RunNetworkHealthCheck runs a comprehensive set of network health checks.
func RunNetworkHealthCheck() HealthReport {
	start := time.Now()
	report := HealthReport{Score: 100}

	// 1. Internet reachability
	check := checkPingHealth("Internet Reachability", "8.8.8.8", 3)
	report.Checks = append(report.Checks, check)
	if check.Status == "fail" {
		report.Score -= 25
	} else if check.Status == "warn" {
		report.Score -= 10
	}

	// 2. DNS resolution
	dnsCheck := checkDNSHealth()
	report.Checks = append(report.Checks, dnsCheck)
	if dnsCheck.Status == "fail" {
		report.Score -= 25
	} else if dnsCheck.Status == "warn" {
		report.Score -= 10
	}

	// 3. Gateway
	gwCheck := checkGatewayHealth()
	report.Checks = append(report.Checks, gwCheck)
	if gwCheck.Status == "fail" {
		report.Score -= 30
	} else if gwCheck.Status == "warn" {
		report.Score -= 10
	}

	// 4. Latency
	latCheck := checkPingHealth("Internet Latency", "8.8.8.8", 5)
	report.Checks = append(report.Checks, latCheck)
	if latCheck.Status == "fail" {
		report.Score -= 20
	} else if latCheck.Status == "warn" {
		report.Score -= 5
	}

	// 5. Packet loss
	lossCheck := checkPacketLossHealth()
	report.Checks = append(report.Checks, lossCheck)
	if lossCheck.Status == "fail" {
		report.Score -= 25
	} else if lossCheck.Status == "warn" {
		report.Score -= 10
	}

	// 6. Interfaces
	ifaceCheck := checkInterfacesHealth()
	report.Checks = append(report.Checks, ifaceCheck)
	if ifaceCheck.Status == "fail" {
		report.Score -= 20
	}

	// 7. VPN
	active := IsVPNActive()
	vpnCheck := HealthCheck{Name: "VPN Status", Status: "pass", Score: 100}
	if active {
		vpnCheck.Detail = "VPN/Secure tunnel active"
	} else {
		vpnCheck.Detail = "No active VPN"
	}
	report.Checks = append(report.Checks, vpnCheck)

	if report.Score < 0 {
		report.Score = 0
	}
	report.Duration = time.Since(start).Round(time.Millisecond).String()
	switch {
	case report.Score >= 80:
		report.Summary = "Network health is good"
	case report.Score >= 60:
		report.Summary = "Network health is degraded"
	default:
		report.Summary = "Network health is poor"
	}
	return report
}

func checkPingHealth(name, target string, count int) HealthCheck {
	check := HealthCheck{Name: name, Score: 100}
	result, err := Ping(target, count)
	if err != nil {
		check.Status = "fail"
		check.Detail = err.Error()
		return check
	}
	packetLoss := float64(result.Lost) / float64(result.Sent) * 100
	if packetLoss > 50 {
		check.Status = "fail"
		check.Detail = fmt.Sprintf("%.0f%% loss to %s", packetLoss, target)
		check.Score = 0
	} else if packetLoss > 10 {
		check.Status = "warn"
		check.Detail = fmt.Sprintf("%.0f%% loss to %s", packetLoss, target)
		check.Score = 50
	} else {
		check.Status = "pass"
		check.Detail = fmt.Sprintf("Avg %dms to %s", result.Avg.Milliseconds(), target)
	}
	return check
}

func checkDNSHealth() HealthCheck {
	check := HealthCheck{Name: "DNS Resolution", Score: 100}
	client := &dns.Client{Timeout: 5 * time.Second}
	msg := new(dns.Msg)
	msg.SetQuestion("google.com.", dns.TypeA)
	_, rtt, err := client.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		check.Status, check.Detail = "fail", err.Error()
		return check
	}
	ms := float64(rtt.Microseconds()) / 1000.0
	if ms > 500 {
		check.Status, check.Detail, check.Score = "warn", fmt.Sprintf("Slow DNS: %.0fms", ms), 50
	} else {
		check.Status, check.Detail = "pass", fmt.Sprintf("DNS: %.0fms", ms)
	}
	return check
}

func checkGatewayHealth() HealthCheck {
	check := HealthCheck{Name: "Gateway Reachability", Score: 100}
	routes, err := GetRoutingTable()
	if err != nil {
		check.Status, check.Detail, check.Score = "warn", "Could not get routing table", 50
		return check
	}
	for _, r := range routes {
		if r.IsDefault && r.Gateway != "" && r.Gateway != "0.0.0.0" {
			result, err := Ping(r.Gateway, 2)
			if err != nil {
				check.Status, check.Detail = "fail", fmt.Sprintf("Cannot reach gateway %s", r.Gateway)
				return check
			}
			packetLoss := float64(result.Lost) / float64(result.Sent) * 100
			if packetLoss > 0 {
				check.Status, check.Detail, check.Score = "warn", fmt.Sprintf("Gateway %s: %.0f%% loss", r.Gateway, packetLoss), 60
			} else {
				check.Status, check.Detail = "pass", fmt.Sprintf("Gateway %s: %dms", r.Gateway, result.Avg.Milliseconds())
			}
			return check
		}
	}
	check.Status, check.Detail, check.Score = "warn", "No default gateway found", 50
	return check
}

func checkPacketLossHealth() HealthCheck {
	check := HealthCheck{Name: "Packet Loss", Score: 100}
	result, err := Ping("8.8.8.8", 10)
	if err != nil {
		check.Status, check.Detail, check.Score = "warn", "Could not measure", 50
		return check
	}
	packetLoss := float64(result.Lost) / float64(result.Sent) * 100
	if packetLoss > 25 {
		check.Status, check.Detail, check.Score = "fail", fmt.Sprintf("%.0f%% loss", packetLoss), 20
	} else if packetLoss > 5 {
		check.Status, check.Detail, check.Score = "warn", fmt.Sprintf("%.0f%% loss", packetLoss), 50
	} else {
		check.Status, check.Detail = "pass", fmt.Sprintf("%.0f%% loss", packetLoss)
	}
	return check
}

func checkInterfacesHealth() HealthCheck {
	check := HealthCheck{Name: "Interface Status", Score: 100}
	ifaces, err := net.Interfaces()
	if err != nil {
		check.Status, check.Detail, check.Score = "warn", "Could not enumerate interfaces", 50
		return check
	}
	upCount := 0
	for _, iface := range ifaces {
		hasUp := false
		hasLoopback := false
		for _, f := range iface.Flags {
			if f == "up" {
				hasUp = true
			}
			if f == "loopback" {
				hasLoopback = true
			}
		}
		if hasUp && !hasLoopback {
			upCount++
		}
	}
	if upCount == 0 {
		check.Status, check.Detail, check.Score = "fail", "No active interfaces", 0
	} else {
		check.Status, check.Detail = "pass", fmt.Sprintf("%d active interface(s)", upCount)
	}
	return check
}
