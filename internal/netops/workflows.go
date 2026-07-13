package netops

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// NetworkReport is a combined network diagnostic report.
type NetworkReport struct {
	Ping        *PingResult
	DNS         *DNSResult
	PortScan    []PortResult
	Trace       *TraceRouteResult
	Connections []ConnectionInfo
	Interfaces  []InterfaceInfo
}

// RunNetworkDiagnostics runs all network checks and returns a combined report.
func RunNetworkDiagnostics() (*NetworkReport, error) {
	report := &NetworkReport{}
	var errs []string
	var errsMu sync.Mutex
	var wg sync.WaitGroup

	// Ping default targets
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer common.RecoverPanic()
		ping, err := Ping("8.8.8.8", 3)
		if err != nil {
			errsMu.Lock()
			errs = append(errs, fmt.Sprintf("Ping: %v", err))
			errsMu.Unlock()
		} else {
			report.Ping = ping
		}
	}()

	// DNS lookup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer common.RecoverPanic()
		dns, err := LookupDNS("google.com")
		if err != nil {
			errsMu.Lock()
			errs = append(errs, fmt.Sprintf("DNS: %v", err))
			errsMu.Unlock()
		} else {
			report.DNS = dns
		}
	}()

	// Port scan (common ports on localhost)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer common.RecoverPanic()
		ports, err := ScanPorts("localhost", DefaultScanPorts())
		if err != nil {
			errsMu.Lock()
			errs = append(errs, fmt.Sprintf("PortScan: %v", err))
			errsMu.Unlock()
		} else {
			report.PortScan = ports
		}
	}()

	// Traceroute
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer common.RecoverPanic()
		trace, err := TraceRoute("8.8.8.8")
		if err != nil {
			errsMu.Lock()
			errs = append(errs, fmt.Sprintf("Traceroute: %v", err))
			errsMu.Unlock()
		} else {
			report.Trace = trace
		}
	}()

	// Network connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer common.RecoverPanic()
		conns, err := GetConnections()
		if err != nil {
			errsMu.Lock()
			errs = append(errs, fmt.Sprintf("Connections: %v", err))
			errsMu.Unlock()
		} else {
			report.Connections = conns
		}
	}()

	// Interfaces
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer common.RecoverPanic()
		ifacesRes, err := GetInterfaces(nil, 0)
		if err != nil {
			errsMu.Lock()
			errs = append(errs, fmt.Sprintf("Interfaces: %v", err))
			errsMu.Unlock()
		} else {
			report.Interfaces = ifacesRes.Interfaces
		}
	}()

	wg.Wait()

	if len(errs) > 0 && report.Ping == nil && report.DNS == nil && len(report.PortScan) == 0 {
		return nil, fmt.Errorf("all network checks failed: %s", strings.Join(errs, "; "))
	}

	return report, nil
}

// String returns a plain-text summary of the network report.
func (r *NetworkReport) String() string {
	var b strings.Builder

	b.WriteString("=== Network Diagnostic Report ===\n\n")

	if r.Ping != nil {
		b.WriteString(fmt.Sprintf("PING %s (%s): sent=%d recv=%d lost=%d avg=%s\n",
			r.Ping.Target, r.Ping.IP, r.Ping.Sent, r.Ping.Received, r.Ping.Lost, r.Ping.Avg))
	}

	if r.DNS != nil {
		b.WriteString(fmt.Sprintf("DNS %s: %d A records, %d AAAA, %d MX\n",
			r.DNS.Hostname, len(r.DNS.A), len(r.DNS.AAAA), len(r.DNS.MX)))
	}

	if len(r.PortScan) > 0 {
		open := 0
		for _, p := range r.PortScan {
			if p.Open {
				open++
			}
		}
		b.WriteString(fmt.Sprintf("PORT SCAN: %d/%d ports open\n", open, len(r.PortScan)))
	}

	if r.Trace != nil {
		b.WriteString(fmt.Sprintf("TRACEROUTE %s: %d hops\n", r.Trace.Target, len(r.Trace.Hops)))
	}

	if len(r.Connections) > 0 {
		listening := 0
		estab := 0
		for _, c := range r.Connections {
			switch c.State {
			case "LISTENING", "LISTEN":
				listening++
			case "ESTABLISHED":
				estab++
			}
		}
		b.WriteString(fmt.Sprintf("CONNECTIONS: %d total (%d listening, %d established)\n",
			len(r.Connections), listening, estab))
	}

	if len(r.Interfaces) > 0 {
		up := 0
		for _, iface := range r.Interfaces {
			if iface.IsUp {
				up++
			}
		}
		b.WriteString(fmt.Sprintf("INTERFACES: %d total (%d up)\n", len(r.Interfaces), up))
	}

	return b.String()
}

// Markdown returns a markdown-formatted network report.
func (r *NetworkReport) Markdown() string {
	var b strings.Builder

	b.WriteString("# 🌐 Network Diagnostic Report\n\n")

	// Summary table
	b.WriteString("## Summary\n\n")
	b.WriteString("| Check | Status |\n|-------|--------|\n")

	if r.Ping != nil {
		pct := float64(r.Ping.Lost) / float64(r.Ping.Sent) * 100
		b.WriteString(fmt.Sprintf("| Ping (%s) | sent=%d recv=%d loss=%.0f%% avg=%s |\n",
			r.Ping.Target, r.Ping.Sent, r.Ping.Received, pct, r.Ping.Avg))
	}
	if r.DNS != nil {
		b.WriteString(fmt.Sprintf("| DNS (%s) | %d records |\n", r.DNS.Hostname, len(r.DNS.A)+len(r.DNS.AAAA)))
	}
	if len(r.PortScan) > 0 {
		open := 0
		for _, p := range r.PortScan {
			if p.Open {
				open++
			}
		}
		b.WriteString(fmt.Sprintf("| Port Scan (localhost) | %d/%d open |\n", open, len(r.PortScan)))
	}
	if r.Trace != nil {
		b.WriteString(fmt.Sprintf("| Traceroute (%s) | %d hops |\n", r.Trace.Target, len(r.Trace.Hops)))
	}
	if len(r.Connections) > 0 {
		b.WriteString(fmt.Sprintf("| Connections | %d total |\n", len(r.Connections)))
	}
	if len(r.Interfaces) > 0 {
		b.WriteString(fmt.Sprintf("| Interfaces | %d total |\n", len(r.Interfaces)))
	}

	// Connections detail
	if len(r.Connections) > 0 {
		b.WriteString("\n## Active Connections\n\n")
		b.WriteString("| Local | Remote | State | PID |\n|-------|--------|-------|-----|\n")
		for i, c := range r.Connections {
			if i >= common.MaxConnections {
				b.WriteString(fmt.Sprintf("| ... and %d more | | | |\n", len(r.Connections)-common.MaxConnections))
				break
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n",
				c.LocalAddr, c.RemoteAddr, c.State, c.PID))
		}
	}

	// Interfaces detail
	if len(r.Interfaces) > 0 {
		b.WriteString("\n## Network Interfaces\n\n")
		for _, iface := range r.Interfaces {
			status := "🟢 UP"
			if !iface.IsUp {
				status = "🔴 DOWN"
			}
			b.WriteString(fmt.Sprintf("### %s (%s)\n\n", iface.Name, status))
			if iface.MAC != "" {
				b.WriteString(fmt.Sprintf("- MAC: `%s`\n", iface.MAC))
			}
			if iface.Speed != "" {
				b.WriteString(fmt.Sprintf("- Speed: %s\n", iface.Speed))
			}
			if len(iface.IPs) > 0 {
				b.WriteString("- IPs:\n")
				for _, ip := range iface.IPs {
					b.WriteString(fmt.Sprintf("  - `%s`\n", ip))
				}
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}
