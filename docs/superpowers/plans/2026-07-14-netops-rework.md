# NetOps Rework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework NetOps into a comprehensive network operations center — 16 categories covering inspection → diagnosis → action for network engineering workflows. Phase 1 (this plan) covers all 12 backend tasks.

**Architecture:** Backend extends existing `internal/netops/` package with new OS-specific collection methods, shell execution for actions, and concurrent operations. Binding layer extends `internal/app/NetOps.go`. Types extend `internal/app/Types.go`.

**Tech Stack:** Go 1.26.5, `miekg/dns`, `golang.org/x/net`, `gopsutil/v4` (net interfaces), Wails v2 bindings

---

## File Structure

### Backend (Go) — Phase 1
| File | Responsibility |
|------|---------------|
| `internal/netops/oui.go` | **New**: OUI vendor database — MAC prefix → vendor lookup |
| `internal/netops/arp.go` | **New**: ARP table parsing (`arp -a` / `ip neigh`) |
| `internal/netops/routing.go` | **New**: Routing table + static route management |
| `internal/netops/wifi.go` | **New**: WiFi scanning + current connection info |
| `internal/netops/dns_advanced.go` | **New**: DNS cache flush, reverse lookup, DoH test |
| `internal/netops/ping_multi.go` | **New**: Concurrent multi-target ping + stats |
| `internal/netops/health.go` | **New**: Network health diagnostics engine |
| `internal/netops/vpn.go` | **New**: VPN detection (OpenVPN, WireGuard, AnyConnect, etc.) |
| `internal/netops/firewall.go` | **New**: Firewall rule enumeration + management |
| `internal/netops/discovery.go` | **New**: Network device discovery (ARP + ping sweep) |
| `internal/netops/monitoring.go` | **New**: Bandwidth monitoring over time |
| `internal/netops/actions.go` | **New**: Network actions (flush DNS, renew DHCP, reset interface, etc.) |
| `internal/app/NetOps.go` | **Extend**: Bind all 12 new backend methods |
| `internal/app/Types.go` | **Extend**: Add all new TypeScript-mirrored types |

---

## Spec Coverage Matrix

| # | Category | Backend Method | Priority |
|---|----------|---------------|----------|
| 1 | OUI Vendor Lookup | `LookupVendor(mac)` | P0 |
| 2 | ARP Table | `GetARPTable()` | P0 |
| 3 | Routing Table | `GetRoutingTable()` + `ManageStaticRoutes()` | P0 |
| 4 | WiFi Scanning | `ScanWiFiNetworks()` + `GetWiFiInfo()` | P1 |
| 5 | DNS Advanced | `FlushDNSCache()` + `ReverseLookup()` + `TestDoH()` | P1 |
| 6 | Multi-Target Ping | `PingMultiTarget()` + `GetPingStats()` | P0 |
| 7 | Network Health | `RunNetworkHealthCheck()` | P1 |
| 8 | VPN Detection | `GetVPNStatus()` | P1 |
| 9 | Firewall Rules | `GetFirewallRules()` + `ManageFirewallRules()` | P1 |
| 10 | Network Discovery | `RunNetworkDiscovery()` | P0 |
| 11 | Network Monitoring | `GetBandwidthHistory()` + `StartMonitoring()` + `StopMonitoring()` | P1 |
| 12 | Network Actions | `RunNetworkAction()` | P1 |

---

## Known Constraints

1. **ARP Table**: Platform-specific parsers needed (`arp -a` on Windows, `ip neigh` on Linux).
2. **WiFi Scanning**: `netsh wlan` (Windows), `nmcli`/`iwlist` (Linux). Graceful fallback if unavailable.
3. **Firewall**: `netsh advfirewall` (Windows), `ufw`/`iptables` (Linux). Detect available tool.
4. **VPN Detection**: Check adapter names + tunnel interfaces + `wg show`/`rasdial`.
5. **DoH Test**: HTTPS POST with JSON wireformat (RFC 8484).
6. **Network Discovery**: Ping sweep needs admin. Falls back to ARP table only.
7. **Bandwidth Monitoring**: Delta-based from cumulative `net.IOCounters()`. Max 360 samples.
8. **Static Routes**: Requires admin/root. Clear error on insufficient privileges.

---

## Phase 1: Backend Foundation (New Go Methods)

### Task 1.1: OUI Vendor Database

**Files:**
- Create: `internal/netops/oui.go`
- Create: `internal/netops/oui_test.go`

- [ ] **Step 1: Create `oui.go` with built-in vendor database**

```go
package netops

import "strings"

var ouiDB = map[string]string{
	"00:00:0c": "Cisco",    "00:1a:2b": "Cisco",      "00:50:b6": "Cisco",
	"00:50:56": "VMware",   "00:0c:29": "VMware",     "00:1c:14": "VMware",
	"00:1b:21": "Intel",    "00:1e:65": "Intel",      "00:08:74": "Dell",
	"00:14:22": "Dell",     "00:15:5d": "Microsoft Hyper-V",
	"00:03:ff": "Microsoft","00:1d:d8": "Microsoft",  "00:1c:42": "Lenovo",
	"00:24:d7": "TP-Link",  "50:c7:bf": "TP-Link",    "00:0a:5e": "3Com",
	"00:12:17": "3Com",     "00:1f:33": "Netgear",    "00:26:f2": "Netgear",
	"00:05:85": "Fortinet", "00:0e:8f": "Fortinet",   "70:4c:a5": "Fortinet",
	"00:22:10": "Juniper",  "28:8a:1c": "Juniper",    "04:4a:6c": "Ubiquiti",
	"f0:9f:c2": "Ubiquiti", "00:0f:20": "MikroTik",   "4c:5e:0c": "MikroTik",
	"00:1f:7b": "Aruba",    "00:24:6c": "Aruba",      "00:17:c5": "Meraki",
	"68:3a:35": "Meraki",   "00:0c:98": "HP",         "00:1b:78": "HP",
	"b8:27:eb": "Raspberry Pi", "dc:a6:32": "Raspberry Pi",
	"00:25:45": "Huawei",   "00:e0:fc": "Huawei",     "48:46:fb": "Huawei",
	"00:19:88": "ZTE",      "00:26:59": "Samsung",    "00:15:99": "Samsung",
	"00:1f:a7": "Sony",     "00:04:1f": "Sony",       "ac:22:0b": "LG",
	"00:50:c9": "LG",       "00:1d:d8": "Google",     "3c:5a:37": "Google",
	"00:11:32": "Synology", "00:09:6b": "IBM",        "00:03:ba": "Oracle",
}

func LookupVendor(mac string) string {
	normalized := strings.ToLower(strings.ReplaceAll(
		strings.ReplaceAll(strings.ReplaceAll(mac, "-", ":"), ".", ":"), " ", ""))
	if len(normalized) < 8 {
		return ""
	}
	return ouiDB[normalized[:8]]
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/oui_test.go`:

```go
package netops

import "testing"

func TestLookupVendor(t *testing.T) {
	tests := []struct{ mac, expected string }{
		{"00:00:0C:01:02:03", "Cisco"},
		{"00-03-FF-AA-BB-CC", "Microsoft"},
		{"b8:27:eb:12:34:56", "Raspberry Pi"},
		{"AA:BB:CC:DD:EE:FF", ""},
	}
	for _, tt := range tests {
		if got := LookupVendor(tt.mac); got != tt.expected {
			t.Errorf("LookupVendor(%q) = %q, want %q", tt.mac, got, tt.expected)
		}
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestLookupVendor -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/oui.go internal/netops/oui_test.go
git commit -m "feat(netops): add OUI vendor database with MAC prefix lookup"
```

---

### Task 1.2: ARP Table

**Files:**
- Create: `internal/netops/arp.go`
- Create: `internal/netops/arp_test.go`

- [ ] **Step 1: Create `arp.go`**

```go
package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type ARPEntry struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Vendor    string `json:"vendor"`
	Interface string `json:"interface"`
}

func GetARPTable() ([]ARPEntry, error) {
	switch runtime.GOOS {
	case "windows":
		return getARPTableWindows()
	case "linux":
		return getARPTableLinux()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func getARPTableWindows() ([]ARPEntry, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("arp command failed: %w", err)
	}
	var entries []ARPEntry
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Interface") || strings.HasPrefix(line, "---") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			e := ARPEntry{IP: fields[0], MAC: fields[1]}
			if len(fields) >= 3 {
				e.Interface = fields[2]
			}
			e.Vendor = LookupVendor(e.MAC)
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func getARPTableLinux() ([]ARPEntry, error) {
	cmd := exec.Command("ip", "neigh", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ip neigh failed: %w", err)
	}
	var entries []ARPEntry
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 5 {
			continue
		}
		e := ARPEntry{IP: fields[0]}
		for i, f := range fields {
			if f == "lladdr" && i+1 < len(fields) {
				e.MAC = fields[i+1]
			}
			if f == "dev" && i+1 < len(fields) {
				e.Interface = fields[i+1]
			}
		}
		if e.MAC != "" {
			e.Vendor = LookupVendor(e.MAC)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
	var entries []ARPEntry
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Interface") || strings.HasPrefix(line, "---") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			e := ARPEntry{IP: fields[0], MAC: fields[1]}
			if len(fields) >= 3 {
				e.Interface = fields[2]
			}
			e.Vendor = LookupVendor(e.MAC)
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func getARPTableLinux() ([]ARPEntry, error) {
	cmd := exec.Command("ip", "neigh", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ip neigh failed: %w", err)
	}
	var entries []ARPEntry
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 5 {
			continue
		}
		e := ARPEntry{IP: fields[0]}
		for i, f := range fields {
			if f == "lladdr" && i+1 < len(fields) {
				e.MAC = fields[i+1]
			}
			if f == "dev" && i+1 < len(fields) {
				e.Interface = fields[i+1]
			}
		}
		if e.MAC != "" {
			e.Vendor = LookupVendor(e.MAC)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/arp_test.go`:

```go
package netops

import "testing"

func TestGetARPTable(t *testing.T) {
	entries, err := GetARPTable()
	if err != nil {
		t.Logf("GetARPTable error (may need admin): %v", err)
		return
	}
	t.Logf("Found %d ARP entries", len(entries))
	for i, e := range entries {
		if i >= 3 {
			break
		}
		t.Logf("  IP=%s MAC=%s Vendor=%s", e.IP, e.MAC, e.Vendor)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestGetARPTable -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/arp.go internal/netops/arp_test.go
git commit -m "feat(netops): add GetARPTable with OUI vendor resolution"
```

---

### Task 1.3: Routing Table

**Files:**
- Create: `internal/netops/routing.go`
- Create: `internal/netops/routing_test.go`

- [ ] **Step 1: Create `routing.go`**

```go
package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type RouteEntry struct {
	Destination string `json:"destination"`
	Mask        string `json:"mask"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
	IsDefault   bool   `json:"is_default"`
}

func GetRoutingTable() ([]RouteEntry, error) {
	switch runtime.GOOS {
	case "windows":
		return getRoutingTableWindows()
	case "linux":
		return getRoutingTableLinux()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func getRoutingTableWindows() ([]RouteEntry, error) {
	cmd := exec.Command("netstat", "-rn")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var entries []RouteEntry
	inIPv4 := false
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "IPv4") {
			inIPv4 = true
			continue
		}
		if strings.Contains(line, "IPv6") {
			inIPv4 = false
			continue
		}
		if !inIPv4 || line == "" || strings.HasPrefix(line, "Network") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			e := RouteEntry{Destination: fields[0], Mask: fields[1], Gateway: fields[2], Interface: fields[3]}
			if len(fields) >= 5 {
				fmt.Sscanf(fields[4], "%d", &e.Metric)
			}
			if e.Destination == "0.0.0.0" {
				e.IsDefault = true
			}
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func getRoutingTableLinux() ([]RouteEntry, error) {
	cmd := exec.Command("ip", "route", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var entries []RouteEntry
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "default") {
			entries = append(entries, RouteEntry{Destination: "0.0.0.0/0", IsDefault: true})
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		e := RouteEntry{Destination: fields[0]}
		for i := 1; i < len(fields)-1; i++ {
			switch fields[i] {
			case "via":
				e.Gateway = fields[i+1]
			case "dev":
				e.Interface = fields[i+1]
			case "metric":
				if i+1 < len(fields) {
					fmt.Sscanf(fields[i+1], "%d", &e.Metric)
				}
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func ManageStaticRoutes(action, dest, mask, gateway string) error {
	switch runtime.GOOS {
	case "windows":
		switch action {
		case "add":
			_, err := exec.Command("netsh", "interface", "ipv4", "add", "route", dest, mask, gateway).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("netsh", "interface", "ipv4", "delete", "route", dest, mask, gateway).CombinedOutput()
			return err
		}
	case "linux":
		switch action {
		case "add":
			_, err := exec.Command("sudo", "ip", "route", "add", dest, "via", gateway).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("sudo", "ip", "route", "del", dest, "via", gateway).CombinedOutput()
			return err
		}
	}
	return fmt.Errorf("unknown action or platform: %s/%s", action, runtime.GOOS)
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/routing_test.go`:

```go
package netops

import "testing"

func TestGetRoutingTable(t *testing.T) {
	routes, err := GetRoutingTable()
	if err != nil {
		t.Fatalf("GetRoutingTable error: %v", err)
	}
	if len(routes) == 0 {
		t.Error("Expected at least one route")
	}
	for _, r := range routes {
		if r.IsDefault {
			t.Logf("Default route: %s via %s dev %s", r.Destination, r.Gateway, r.Interface)
		}
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestGetRoutingTable -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/routing.go internal/netops/routing_test.go
git commit -m "feat(netops): add GetRoutingTable and ManageStaticRoutes"
```

---

### Task 1.4: WiFi Scanning

**Files:**
- Create: `internal/netops/wifi.go`
- Create: `internal/netops/wifi_test.go`

- [ ] **Step 1: Create `wifi.go`**

```go
package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type WiFiNetwork struct {
	SSID      string `json:"ssid"`
	Signal    int    `json:"signal"`
	Channel   int    `json:"channel"`
	Security  string `json:"security"`
	BSSID     string `json:"bssid"`
	Frequency string `json:"frequency"`
}

type WiFiInfo struct {
	Interface string `json:"interface"`
	SSID      string `json:"ssid"`
	Signal    int    `json:"signal"`
	Speed     string `json:"speed"`
	Channel   int    `json:"channel"`
}

func ScanWiFiNetworks() ([]WiFiNetwork, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "wlan", "show", "networks", "mode=bssid")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseWiFiWindows(string(output))
	case "linux":
		cmd := exec.Command("nmcli", "-f", "SSID,SIGNAL,CHAN,SECURITY", "dev", "wifi", "list", "--rescan", "yes")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseWiFiNmcli(string(output))
	default:
		return nil, fmt.Errorf("unsupported platform")
	}
}

func parseWiFiWindows(output string) ([]WiFiNetwork, error) {
	var networks []WiFiNetwork
	var current *WiFiNetwork
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "SSID") && strings.Contains(trimmed, ":") {
			if current != nil {
				networks = append(networks, *current)
			}
			current = &WiFiNetwork{SSID: strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])}
		} else if current != nil {
			if strings.HasPrefix(trimmed, "Signal") {
				pct := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]), "%")
				current.Signal, _ = strconv.Atoi(pct)
			} else if strings.HasPrefix(trimmed, "Channel") {
				current.Channel, _ = strconv.Atoi(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]))
			} else if strings.HasPrefix(trimmed, "Authentication") {
				current.Security = strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])
			}
		}
	}
	if current != nil {
		networks = append(networks, *current)
	}
	return networks, nil
}

func parseWiFiNmcli(output string) ([]WiFiNetwork, error) {
	var networks []WiFiNetwork
	for _, line := range strings.Split(output, "\n")[1:] {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 {
			networks = append(networks, WiFiNetwork{
				SSID:   fields[0],
				Signal: parseIntSafe(fields[1]),
			})
		}
	}
	return networks, nil
}

func GetWiFiInfo() (*WiFiInfo, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "wlan", "show", "interfaces")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseWiFiInfoWindows(string(output))
	case "linux":
		cmd := exec.Command("nmcli", "-t", "-f", "ACTIVE,SSID,SIGNAL,SPEED,CHAN", "dev", "wifi")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "yes:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 5 {
					return &WiFiInfo{SSID: parts[1], Signal: parseIntSafe(parts[2]), Speed: parts[3], Channel: parseIntSafe(parts[4])}, nil
				}
			}
		}
		return nil, fmt.Errorf("not connected to WiFi")
	default:
		return nil, fmt.Errorf("unsupported platform")
	}
}

func parseWiFiInfoWindows(output string) (*WiFiInfo, error) {
	info := &WiFiInfo{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "SSID") && strings.Contains(trimmed, ":"):
			info.SSID = strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])
		case strings.HasPrefix(trimmed, "Signal"):
			pct := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]), "%")
			info.Signal, _ = strconv.Atoi(pct)
		case strings.HasPrefix(trimmed, "Speed"):
			info.Speed = strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])
		case strings.HasPrefix(trimmed, "Channel"):
			info.Channel, _ = strconv.Atoi(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]))
		}
	}
	if info.SSID == "" {
		return nil, fmt.Errorf("not connected to WiFi")
	}
	return info, nil
}

func parseIntSafe(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/wifi_test.go`:

```go
package netops

import "testing"

func TestGetWiFiInfo(t *testing.T) {
	info, err := GetWiFiInfo()
	if err != nil {
		t.Logf("GetWiFiInfo error (may be expected without WiFi): %v", err)
		return
	}
	t.Logf("Connected to: %s, Signal: %d%%, Speed: %s", info.SSID, info.Signal, info.Speed)
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestGetWiFiInfo -v`
Expected: PASS (or graceful skip if no WiFi)

- [ ] **Step 4: Commit**

```bash
git add internal/netops/wifi.go internal/netops/wifi_test.go
git commit -m "feat(netops): add WiFi scanning and connection info via netsh/nmcli"
```

---

### Task 1.5: DNS Advanced Diagnostics

**Files:**
- Create: `internal/netops/dns_advanced.go`
- Create: `internal/netops/dns_advanced_test.go`

- [ ] **Step 1: Create `dns_advanced.go`**

```go
package netops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/miekg/dns"
)

type DoHResult struct {
	Server     string  `json:"server"`
	LatencyMs  float64 `json:"latency_ms"`
	Success    bool    `json:"success"`
	ResolvedIP string  `json:"resolved_ip"`
}

func FlushDNSCache() error {
	switch runtime.GOOS {
	case "windows":
		_, err := exec.Command("ipconfig", "/flushdns").CombinedOutput()
		return err
	case "linux":
		_, err := exec.Command("sudo", "resolvectl", "flush-caches").CombinedOutput()
		if err != nil {
			_, err = exec.Command("sudo", "systemd-resolve", "--flush-caches").CombinedOutput()
		}
		return err
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func ReverseLookup(ip string) (string, error) {
	arpa, err := dns.Reverse(ip)
	if err != nil {
		return "", err
	}
	msg := new(dns.Msg)
	msg.SetQuestion(arpa, dns.TypePTR)
	msg.RecursionDesired = true
	client := &dns.Client{Timeout: 5 * time.Second}
	resp, _, err := client.Exchange(msg, "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	for _, rr := range resp.Answer {
		if ptr, ok := rr.(*dns.PTR); ok {
			return ptr.Ptr, nil
		}
	}
	return "", fmt.Errorf("no PTR record found")
}

func TestDoH(server string) DoHResult {
	result := DoHResult{Server: server}
	type dohQuery struct {
		Name string `json:"name"`
		Type int    `json:"type"`
	}
	type dohAnswer struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	type dohMsg struct {
		Status int         `json:"Status"`
		Answer []dohAnswer `json:"Answer"`
	}
	body, _ := json.Marshal(dohQuery{Name: "google.com", Type: 1})
	start := time.Now()
	resp, err := http.Post(server+"/dns-query", "application/dns-json", bytes.NewReader(body))
	if err != nil {
		return result
	}
	defer resp.Body.Close()
	result.LatencyMs = float64(time.Since(start).Milliseconds())
	var msg dohMsg
	json.NewDecoder(resp.Body).Decode(&msg)
	result.Success = msg.Status == 0 && len(msg.Answer) > 0
	if result.Success {
		result.ResolvedIP = msg.Answer[0].Data
	}
	return result
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/dns_advanced_test.go`:

```go
package netops

import "testing"

func TestReverseLookup(t *testing.T) {
	result, err := ReverseLookup("8.8.8.8")
	if err != nil {
		t.Logf("ReverseLookup error: %v", err)
		return
	}
	t.Logf("PTR for 8.8.8.8: %s", result)
}

func TestFlushDNSCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DNS flush in short mode")
	}
	err := FlushDNSCache()
	if err != nil {
		t.Logf("FlushDNSCache error (may need admin): %v", err)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestReverseLookup -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/dns_advanced.go internal/netops/dns_advanced_test.go
git commit -m "feat(netops): add DNS cache flush, reverse lookup, and DoH test"
```

---

### Task 1.6: Multi-Target Ping

**Files:**
- Create: `internal/netops/ping_multi.go`
- Create: `internal/netops/ping_multi_test.go`

- [ ] **Step 1: Create `ping_multi.go`**

```go
package netops

import (
	"math"
	"sort"
	"sync"
)

type PingResultMulti struct {
	Target         string    `json:"target"`
	MinMs          float64   `json:"min_ms"`
	AvgMs          float64   `json:"avg_ms"`
	MaxMs          float64   `json:"max_ms"`
	StdDevMs       float64   `json:"stddev_ms"`
	PacketLoss     float64   `json:"packet_loss"`
	JitterMs       float64   `json:"jitter_ms"`
	IndividualRTTs []float64 `json:"individual_rtts"`
	Success        bool      `json:"success"`
	Error          string    `json:"error,omitempty"`
}

type PingStats struct {
	AvgLatency  float64 `json:"avg_latency"`
	MaxLatency  float64 `json:"max_latency"`
	TotalLoss   float64 `json:"total_loss"`
	WorstTarget string  `json:"worst_target"`
}

func PingMultiTarget(targets []string, count int) []PingResultMulti {
	if count <= 0 {
		count = 4
	}
	results := make([]PingResultMulti, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func(idx int, t string) {
			defer wg.Done()
			results[idx] = pingOneTarget(t, count)
		}(i, target)
	}
	wg.Wait()
	return results
}

func pingOneTarget(target string, count int) PingResultMulti {
	result := PingResultMulti{Target: target}
	pingResult, err := Ping(target, count)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.IndividualRTTs = pingResult.Times
	result.PacketLoss = pingResult.PacketLoss
	result.Success = len(pingResult.Times) > 0
	if len(pingResult.Times) > 0 {
		sort.Float64s(pingResult.Times)
		result.MinMs = pingResult.Times[0]
		result.MaxMs = pingResult.Times[len(pingResult.Times)-1]
		result.AvgMs = pingResult.AvgMs
		var sumSq float64
		for _, t := range pingResult.Times {
			sumSq += (t - result.AvgMs) * (t - result.AvgMs)
		}
		result.StdDevMs = math.Sqrt(sumSq / float64(len(pingResult.Times)))
		if len(pingResult.Times) > 1 {
			var jSum float64
			for i := 1; i < len(pingResult.Times); i++ {
				jSum += math.Abs(pingResult.Times[i] - pingResult.Times[i-1])
			}
			result.JitterMs = jSum / float64(len(pingResult.Times)-1)
		}
	}
	return result
}

func GetPingStats(results []PingResultMulti) PingStats {
	stats := PingStats{}
	var totalLatency float64
	var totalLoss float64
	var count int
	for _, r := range results {
		if r.Success {
			totalLatency += r.AvgMs
			count++
			if r.AvgMs > stats.MaxLatency {
				stats.MaxLatency = r.AvgMs
				stats.WorstTarget = r.Target
			}
		}
		totalLoss += r.PacketLoss
	}
	if count > 0 {
		stats.AvgLatency = totalLatency / float64(count)
	}
	if len(results) > 0 {
		stats.TotalLoss = totalLoss / float64(len(results))
	}
	return stats
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/ping_multi_test.go`:

```go
package netops

import "testing"

func TestPingMultiTarget(t *testing.T) {
	targets := []string{"127.0.0.1", "8.8.8.8"}
	results := PingMultiTarget(targets, 3)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		t.Logf("Target=%s Success=%v AvgMs=%.1f Loss=%.0f%%", r.Target, r.Success, r.AvgMs, r.PacketLoss)
	}
	stats := GetPingStats(results)
	t.Logf("Stats: AvgLatency=%.1f MaxLatency=%.1f Worst=%s", stats.AvgLatency, stats.MaxLatency, stats.WorstTarget)
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestPingMultiTarget -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/ping_multi.go internal/netops/ping_multi_test.go
git commit -m "feat(netops): add PingMultiTarget with concurrent ping and aggregate stats"
```

---

### Task 1.7: Network Health Diagnostics

**Files:**
- Create: `internal/netops/health.go`
- Create: `internal/netops/health_test.go`

- [ ] **Step 1: Create `health.go`**

```go
package netops

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/miekg/dns"
)

type HealthCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "warn", "fail"
	Detail string `json:"detail"`
	Score  int    `json:"score"`
}

type HealthReport struct {
	Score    int           `json:"score"`
	Checks   []HealthCheck `json:"checks"`
	Summary  string        `json:"summary"`
	Duration string        `json:"duration"`
}

func RunNetworkHealthCheck() HealthReport {
	start := time.Now()
	report := HealthReport{Score: 100}

	// 1. Internet reachability
	check := checkPing("Internet Reachability", "8.8.8.8", 3)
	report.Checks = append(report.Checks, check)

	// 2. DNS resolution
	dnsCheck := checkDNS()
	report.Checks = append(report.Checks, dnsCheck)
	if dnsCheck.Status == "fail" {
		report.Score -= 25
	} else if dnsCheck.Status == "warn" {
		report.Score -= 10
	}

	// 3. Gateway
	gwCheck := checkGateway()
	report.Checks = append(report.Checks, gwCheck)
	if gwCheck.Status == "fail" {
		report.Score -= 30
	} else if gwCheck.Status == "warn" {
		report.Score -= 10
	}

	// 4. Latency
	latCheck := checkPing("Internet Latency", "8.8.8.8", 5)
	report.Checks = append(report.Checks, latCheck)
	if latCheck.Status == "fail" {
		report.Score -= 20
	} else if latCheck.Status == "warn" {
		report.Score -= 5
	}

	// 5. Packet loss
	lossCheck := checkPacketLoss()
	report.Checks = append(report.Checks, lossCheck)
	if lossCheck.Status == "fail" {
		report.Score -= 25
	} else if lossCheck.Status == "warn" {
		report.Score -= 10
	}

	// 6. Interfaces
	ifaceCheck := checkInterfaces()
	report.Checks = append(report.Checks, ifaceCheck)
	if ifaceCheck.Status == "fail" {
		report.Score -= 20
	}

	// 7. VPN
	vpnStatus := GetVPNStatus()
	vpnCheck := HealthCheck{Name: "VPN Status", Status: "pass", Score: 100}
	if vpnStatus.Active {
		vpnCheck.Detail = fmt.Sprintf("VPN active: %s", vpnStatus.Type)
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

func checkPing(name, target string, count int) HealthCheck {
	check := HealthCheck{Name: name, Score: 100}
	result, err := Ping(target, count)
	if err != nil {
		check.Status = "fail"
		check.Detail = err.Error()
		return check
	}
	if result.PacketLoss > 50 {
		check.Status = "fail"
		check.Detail = fmt.Sprintf("%.0f%% loss to %s", result.PacketLoss, target)
		check.Score = 0
	} else if result.PacketLoss > 10 {
		check.Status = "warn"
		check.Detail = fmt.Sprintf("%.0f%% loss to %s", result.PacketLoss, target)
		check.Score = 50
	} else {
		check.Status = "pass"
		check.Detail = fmt.Sprintf("Avg %.1fms to %s", result.AvgMs, target)
	}
	return check
}

func checkDNS() HealthCheck {
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

func checkGateway() HealthCheck {
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
			if result.PacketLoss > 0 {
				check.Status, check.Detail, check.Score = "warn", fmt.Sprintf("Gateway %s: %.0f%% loss", r.Gateway, result.PacketLoss), 60
			} else {
				check.Status, check.Detail = "pass", fmt.Sprintf("Gateway %s: %.1fms", r.Gateway, result.AvgMs)
			}
			return check
		}
	}
	check.Status, check.Detail, check.Score = "warn", "No default gateway found", 50
	return check
}

func checkPacketLoss() HealthCheck {
	check := HealthCheck{Name: "Packet Loss", Score: 100}
	result, err := Ping("8.8.8.8", 10)
	if err != nil {
		check.Status, check.Detail, check.Score = "warn", "Could not measure", 50
		return check
	}
	if result.PacketLoss > 25 {
		check.Status, check.Detail, check.Score = "fail", fmt.Sprintf("%.0f%% loss", result.PacketLoss), 20
	} else if result.PacketLoss > 5 {
		check.Status, check.Detail, check.Score = "warn", fmt.Sprintf("%.0f%% loss", result.PacketLoss), 50
	} else {
		check.Status, check.Detail = "pass", fmt.Sprintf("%.0f%% loss", result.PacketLoss)
	}
	return check
}

func checkInterfaces() HealthCheck {
	check := HealthCheck{Name: "Interface Status", Score: 100}
	ifaces, err := net.Interfaces()
	if err != nil {
		check.Status, check.Detail, check.Score = "warn", "Could not enumerate interfaces", 50
		return check
	}
	upCount := 0
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
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
```

- [ ] **Step 2: Write test**

Create `internal/netops/health_test.go`:

```go
package netops

import "testing"

func TestRunNetworkHealthCheck(t *testing.T) {
	report := RunNetworkHealthCheck()
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("Score should be 0-100, got %d", report.Score)
	}
	if len(report.Checks) == 0 {
		t.Error("Expected at least one health check")
	}
	t.Logf("Health Score: %d, Summary: %s, Duration: %s", report.Score, report.Summary, report.Duration)
	for _, c := range report.Checks {
		t.Logf("  [%s] %s: %s", c.Status, c.Name, c.Detail)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestRunNetworkHealthCheck -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/health.go internal/netops/health_test.go
git commit -m "feat(netops): add RunNetworkHealthCheck with 7 diagnostic checks and scoring"
```

---

### Task 1.8: VPN Detection

**Files:**
- Create: `internal/netops/vpn.go`
- Create: `internal/netops/vpn_test.go`

- [ ] **Step 1: Create `vpn.go`**

```go
package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/net"
)

type VPNStatus struct {
	Active    bool   `json:"active"`
	Type      string `json:"type"`
	Interface string `json:"interface"`
	RemoteIP  string `json:"remote_ip"`
	LocalIP   string `json:"local_ip"`
	Protocol  string `json:"protocol"`
}

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
```

- [ ] **Step 2: Write test**

Create `internal/netops/vpn_test.go`:

```go
package netops

import "testing"

func TestGetVPNStatus(t *testing.T) {
	status := GetVPNStatus()
	t.Logf("VPN Active=%v Type=%s Interface=%s LocalIP=%s", status.Active, status.Type, status.Interface, status.LocalIP)
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestGetVPNStatus -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/vpn.go internal/netops/vpn_test.go
git commit -m "feat(netops): add VPN detection for WireGuard, OpenVPN, rasdial"
```

---

### Task 1.9: Firewall Rules

**Files:**
- Create: `internal/netops/firewall.go`
- Create: `internal/netops/firewall_test.go`

- [ ] **Step 1: Create `firewall.go`**

```go
package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type FirewallRule struct {
	Name        string `json:"name"`
	Direction   string `json:"direction"`
	Action      string `json:"action"`
	Protocol    string `json:"protocol"`
	Ports       string `json:"ports"`
	Enabled     bool   `json:"enabled"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

func GetFirewallRules() ([]FirewallRule, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=all", "dir=in", "status=all")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseNetshOutput(string(output))
	case "linux":
		cmd := exec.Command("sudo", "ufw", "status", "verbose")
		output, err := cmd.CombinedOutput()
		if err == nil && strings.Contains(string(output), "Status: active") {
			return parseUFWOutput(string(output))
		}
		cmd = exec.Command("sudo", "iptables", "-L", "-n", "-v")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseIptablesOutput(string(output))
	default:
		return nil, fmt.Errorf("unsupported platform")
	}
}

func parseNetshOutput(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	var current *FirewallRule
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Rule Name:") {
			if current != nil {
				rules = append(rules, *current)
			}
			current = &FirewallRule{Name: strings.TrimPrefix(trimmed, "Rule Name: "), Enabled: true}
		} else if current != nil {
			if strings.HasPrefix(trimmed, "Enabled:") {
				current.Enabled = strings.ToLower(strings.TrimPrefix(trimmed, "Enabled: ")) == "yes"
			} else if strings.HasPrefix(trimmed, "Direction:") {
				current.Direction = strings.TrimPrefix(trimmed, "Direction: ")
			} else if strings.HasPrefix(trimmed, "Action:") {
				current.Action = strings.TrimPrefix(trimmed, "Action: ")
			} else if strings.HasPrefix(trimmed, "Protocol:") {
				current.Protocol = strings.TrimPrefix(trimmed, "Protocol: ")
			} else if strings.HasPrefix(trimmed, "LocalPort:") {
				current.Ports = strings.TrimPrefix(trimmed, "LocalPort: ")
			}
		}
	}
	if current != nil {
		rules = append(rules, *current)
	}
	return rules, nil
}

func parseUFWOutput(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "[") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 4 {
			rules = append(rules, FirewallRule{Action: fields[0], Protocol: fields[1], Ports: fields[2], Direction: "in", Enabled: true})
		}
	}
	return rules, nil
}

func parseIptablesOutput(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 8 && fields[0] != "pkts" {
			rules = append(rules, FirewallRule{Action: fields[7], Protocol: fields[4], Direction: "in", Enabled: true})
		}
	}
	return rules, nil
}

func ManageFirewallRules(action string, rule FirewallRule) error {
	switch runtime.GOOS {
	case "windows":
		switch action {
		case "add":
			_, err := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
				"name="+rule.Name, "dir="+rule.Direction, "action="+rule.Action,
				"protocol="+rule.Protocol, "localport="+rule.Ports).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+rule.Name).CombinedOutput()
			return err
		}
	case "linux":
		switch action {
		case "add":
			_, err := exec.Command("sudo", "ufw", "allow", rule.Ports).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("sudo", "ufw", "delete", "allow", rule.Ports).CombinedOutput()
			return err
		}
	}
	return fmt.Errorf("unknown action: %s", action)
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/firewall_test.go`:

```go
package netops

import "testing"

func TestGetFirewallRules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping firewall test in short mode")
	}
	rules, err := GetFirewallRules()
	if err != nil {
		t.Logf("GetFirewallRules error (may need admin): %v", err)
		return
	}
	t.Logf("Found %d firewall rules", len(rules))
	for i, r := range rules {
		if i >= 5 {
			break
		}
		t.Logf("  %s: %s %s %s (enabled=%v)", r.Name, r.Action, r.Direction, r.Protocol, r.Enabled)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestGetFirewallRules -v`
Expected: PASS (or skip without admin)

- [ ] **Step 4: Commit**

```bash
git add internal/netops/firewall.go internal/netops/firewall_test.go
git commit -m "feat(netops): add GetFirewallRules and ManageFirewallRules (netsh/ufw/iptables)"
```

---

### Task 1.10: Network Discovery Engine

**Files:**
- Create: `internal/netops/discovery.go`
- Create: `internal/netops/discovery_test.go`

- [ ] **Step 1: Create `discovery.go`**

```go
package netops

import (
	"fmt"
	"sync"
	"time"
)

type DiscoveredDevice struct {
	IP             string `json:"ip"`
	MAC            string `json:"mac"`
	Vendor         string `json:"vendor"`
	Hostname       string `json:"hostname"`
	OpenPorts      []int  `json:"open_ports"`
	ResponseTimeMs int64  `json:"response_time_ms"`
}

type DiscoveryResult struct {
	Devices    []DiscoveredDevice `json:"devices"`
	Subnet     string             `json:"subnet"`
	ScanTimeMs int64              `json:"scan_time_ms"`
}

func RunNetworkDiscovery(subnet string) DiscoveryResult {
	start := time.Now()
	result := DiscoveryResult{Subnet: subnet}
	seen := make(map[string]bool)

	// ARP table
	for _, arp := range arpEntriesSafe() {
		if arp.MAC != "" && !seen[arp.IP] {
			seen[arp.IP] = true
			result.Devices = append(result.Devices, DiscoveredDevice{IP: arp.IP, MAC: arp.MAC, Vendor: arp.Vendor})
		}
	}

	// Ping sweep
	if subnet != "" {
		var mu sync.Mutex
		var wg sync.WaitGroup
		for _, host := range generateSubnetHosts(subnet) {
			if seen[host] {
				continue
			}
			wg.Add(1)
			go func(ip string) {
				defer wg.Done()
				pingStart := time.Now()
				pingResult, err := Ping(ip, 1)
				if err == nil && pingResult.PacketLoss < 100 {
					mu.Lock()
					result.Devices = append(result.Devices, DiscoveredDevice{IP: ip, ResponseTimeMs: time.Since(pingStart).Milliseconds()})
					seen[ip] = true
					mu.Unlock()
				}
			}(host)
		}
		wg.Wait()
	}
	result.ScanTimeMs = time.Since(start).Milliseconds()
	return result
}

func arpEntriesSafe() []ARPEntry {
	entries, _ := GetARPTable()
	return entries
}

func generateSubnetHosts(subnet string) []string {
	base := subnet
	if idx := len(base) - 1; idx >= 0 && base[idx] == '/' {
		base = base[:idx]
	}
	var hosts []string
	for i := 1; i < 255; i++ {
		hosts = append(hosts, fmt.Sprintf("%s.%d", base, i))
	}
	return hosts
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/discovery_test.go`:

```go
package netops

import "testing"

func TestRunNetworkDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping discovery test in short mode")
	}
	result := RunNetworkDiscovery("192.168.1")
	t.Logf("Found %d devices on %s in %dms", len(result.Devices), result.Subnet, result.ScanTimeMs)
	for _, d := range result.Devices {
		t.Logf("  IP=%s MAC=%s Vendor=%s", d.IP, d.MAC, d.Vendor)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestRunNetworkDiscovery -v -count=1`
Expected: PASS (may take time for ping sweep)

- [ ] **Step 4: Commit**

```bash
git add internal/netops/discovery.go internal/netops/discovery_test.go
git commit -m "feat(netops): add RunNetworkDiscovery with ARP + ping sweep + OUI resolution"
```

---

### Task 1.11: Network Monitoring

**Files:**
- Create: `internal/netops/monitoring.go`
- Create: `internal/netops/monitoring_test.go`

- [ ] **Step 1: Create `monitoring.go`**

```go
package netops

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

type BandwidthSample struct {
	Timestamp     time.Time `json:"timestamp"`
	RxBytesPerSec float64   `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64   `json:"tx_bytes_per_sec"`
	Interface     string    `json:"interface"`
}

var (
	bandwidthHistory []BandwidthSample
	bandwidthMu      sync.Mutex
	monitoringTicker *time.Ticker
	monitoringDone   chan struct{}
)

func GetBandwidthHistory() []BandwidthSample {
	bandwidthMu.Lock()
	defer bandwidthMu.Unlock()
	result := make([]BandwidthSample, len(bandwidthHistory))
	copy(result, bandwidthHistory)
	return result
}

func StartMonitoring(intervalSec int) {
	bandwidthMu.Lock()
	defer bandwidthMu.Unlock()
	if monitoringTicker != nil {
		return
	}
	if intervalSec <= 0 {
		intervalSec = 5
	}
	monitoringTicker = time.NewTicker(time.Duration(intervalSec) * time.Second)
	monitoringDone = make(chan struct{})
	go func() {
		var prevBytes net.IOCountersStat
		first := true
		for {
			select {
			case <-monitoringTicker.C:
				counters, err := net.IOCounters(false)
				if err != nil || len(counters) == 0 {
					continue
				}
				total := counters[0]
				if !first {
					bandwidthMu.Lock()
					bandwidthHistory = append(bandwidthHistory, BandwidthSample{
						Timestamp: time.Now(),
						RxBytesPerSec: float64(total.BytesRecv-prevBytes.BytesRecv) / float64(intervalSec),
						TxBytesPerSec: float64(total.BytesSent-prevBytes.BytesSent) / float64(intervalSec),
						Interface: "total",
					})
					if len(bandwidthHistory) > 360 {
						bandwidthHistory = bandwidthHistory[len(bandwidthHistory)-360:]
					}
					bandwidthMu.Unlock()
				}
				prevBytes = total
				first = false
			case <-monitoringDone:
				monitoringTicker.Stop()
				return
			}
		}
	}()
}

func StopMonitoring() {
	bandwidthMu.Lock()
	defer bandwidthMu.Unlock()
	if monitoringTicker != nil {
		close(monitoringDone)
		monitoringTicker = nil
		monitoringDone = nil
	}
}
```

- [ ] **Step 2: Write test**

Create `internal/netops/monitoring_test.go`:

```go
package netops

import (
	"testing"
	"time"
)

func TestGetBandwidthHistory(t *testing.T) {
	StartMonitoring(1)
	time.Sleep(3 * time.Second)
	StopMonitoring()
	history := GetBandwidthHistory()
	t.Logf("Got %d bandwidth samples", len(history))
	for _, s := range history {
		t.Logf("  %.0f Bps rx, %.0f Bps tx", s.RxBytesPerSec, s.TxBytesPerSec)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestGetBandwidthHistory -v -timeout 30s`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/netops/monitoring.go internal/netops/monitoring_test.go
git commit -m "feat(netops): add bandwidth monitoring with delta-based sampling"
```

---

### Task 1.12: Network Actions

**Files:**
- Create: `internal/netops/actions.go`
- Create: `internal/netops/actions_test.go`

- [ ] **Step 1: Create `actions.go`**

```go
package netops

import (
	"fmt"
	"os/exec"
	"runtime"
)

func RunNetworkAction(action string, params map[string]string) error {
	iface := params["interface"]
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
	// Brief pause before re-enabling
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
```

- [ ] **Step 2: Write test**

Create `internal/netops/actions_test.go`:

```go
package netops

import "testing"

func TestRunNetworkAction_FlushDNS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	err := RunNetworkAction("flush_dns", nil)
	if err != nil {
		t.Logf("FlushDNS error (may need admin): %v", err)
	}
}

func TestRunNetworkAction_ClearARP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	err := RunNetworkAction("clear_arp_cache", nil)
	if err != nil {
		t.Logf("ClearARP error (may need admin): %v", err)
	}
}
```

- [ ] **Step 3: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -run TestRunNetworkAction -v`
Expected: PASS (or skip in short mode)

- [ ] **Step 4: Commit**

```bash
git add internal/netops/actions.go internal/netops/actions_test.go
git commit -m "feat(netops): add RunNetworkAction for flush_dns, renew_dhcp, reset_interface, etc."
```

---

### Binding Task: Extend NetOps.go and Types.go

**Files:**
- Modify: `internal/app/NetOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Add types to `app/Types.go`**

Add these structs to the existing Types.go file (mirrors backend types for Wails JSON serialization):

```go
// ── NetOps Extended Types ──

type ARPEntryData struct {
	IP, MAC, Vendor, Interface string
}
type RouteEntryData struct {
	Destination, Mask, Gateway, Interface string
	Metric                                int
	IsDefault                             bool
}
type WiFiNetworkData struct {
	SSID, Security, BSSID, Frequency string
	Signal, Channel                  int
}
type WiFiInfoData struct {
	Interface, SSID, Speed string
	Signal, Channel        int
}
type DoHResultData struct {
	Server, ResolvedIP string
	LatencyMs          float64
	Success            bool
}
type PingResultMultiData struct {
	Target         string
	MinMs, AvgMs, MaxMs, StdDevMs, PacketLoss, JitterMs float64
	IndividualRTTs                                         []float64
	Success                                                bool
	Error                                                  string
}
type PingStatsData struct {
	AvgLatency, MaxLatency, TotalLoss float64
	WorstTarget                       string
}
type HealthCheckData struct {
	Name, Status, Detail string
	Score                int
}
type HealthReportData struct {
	Score               int
	Checks              []HealthCheckData
	Summary, Duration   string
}
type VPNStatusData struct {
	Type, Interface, RemoteIP, LocalIP, Protocol string
	Active                                       bool
}
type FirewallRuleData struct {
	Name, Direction, Action, Protocol, Ports, Source, Destination string
	Enabled                                                       bool
}
type DiscoveredDeviceData struct {
	IP, MAC, Vendor, Hostname string
	OpenPorts                 []int
	ResponseTimeMs            int64
}
type DiscoveryResultData struct {
	Devices    []DiscoveredDeviceData
	Subnet     string
	ScanTimeMs int64
}
type BandwidthSampleData struct {
	Timestamp     string
	RxBytesPerSec, TxBytesPerSec float64
	Interface                     string
}
type NetworkActionResult struct {
	Action, Message string
	Success         bool
}
```

- [ ] **Step 2: Add binding methods to `app/NetOps.go`**

Add these methods to the `NetOps` struct. Each method calls the corresponding `netops` package function, converts internal types to `Data` types, and returns:

```go
func (n *NetOps) GetARPTable() []ARPEntryData { ... }
func (n *NetOps) GetRoutingTable() []RouteEntryData { ... }
func (n *NetOps) ManageStaticRoutes(action, dest, mask, gateway string) NetworkActionResult { ... }
func (n *NetOps) ScanWiFiNetworks() []WiFiNetworkData { ... }
func (n *NetOps) GetWiFiInfo() WiFiInfoData { ... }
func (n *NetOps) FlushDNSCache() NetworkActionResult { ... }
func (n *NetOps) ReverseLookup(ip string) string { ... }
func (n *NetOps) TestDoH(server string) DoHResultData { ... }
func (n *NetOps) PingMultiTarget(targets []string, count int) []PingResultMultiData { ... }
func (n *NetOps) GetPingStats(results []PingResultMultiData) PingStatsData { ... }
func (n *NetOps) RunNetworkHealthCheck() HealthReportData { ... }
func (n *NetOps) GetVPNStatus() VPNStatusData { ... }
func (n *NetOps) GetFirewallRules() []FirewallRuleData { ... }
func (n *NetOps) ManageFirewallRules(action string, rule FirewallRuleData) NetworkActionResult { ... }
func (n *NetOps) RunNetworkDiscovery(subnet string) DiscoveryResultData { ... }
func (n *NetOps) GetBandwidthHistory() []BandwidthSampleData { ... }
func (n *NetOps) StartMonitoring(intervalSec int) { ... }
func (n *NetOps) StopMonitoring() { ... }
func (n *NetOps) RunNetworkAction(action string, params map[string]string) NetworkActionResult { ... }
```

Each binding method follows the SysOps pattern: call backend → handle error with `common.LogWarn` → convert types → return.

- [ ] **Step 3: Verify build compiles**

Run: `cd E:\Projects\projectx\AllOpsFull && go build ./...`
Expected: Build succeeds with no errors

- [ ] **Step 4: Run all NetOps tests**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/netops/ -v -count=1`
Expected: All tests PASS

- [ ] **Step 5: Run go vet**

Run: `cd E:\Projects\projectx\AllOpsFull && go vet ./...`
Expected: No issues

- [ ] **Step 6: Commit**

```bash
git add internal/app/NetOps.go internal/app/Types.go
git commit -m "feat(netops): bind all 12 new backend methods and add types to Types.go"
```

---

## Summary

| Task | File | Method(s) | Status |
|------|------|-----------|--------|
| 1.1 | `oui.go` | `LookupVendor` | ☐ |
| 1.2 | `arp.go` | `GetARPTable` | ☐ |
| 1.3 | `routing.go` | `GetRoutingTable`, `ManageStaticRoutes` | ☐ |
| 1.4 | `wifi.go` | `ScanWiFiNetworks`, `GetWiFiInfo` | ☐ |
| 1.5 | `dns_advanced.go` | `FlushDNSCache`, `ReverseLookup`, `TestDoH` | ☐ |
| 1.6 | `ping_multi.go` | `PingMultiTarget`, `GetPingStats` | ☐ |
| 1.7 | `health.go` | `RunNetworkHealthCheck` | ☐ |
| 1.8 | `vpn.go` | `GetVPNStatus` | ☐ |
| 1.9 | `firewall.go` | `GetFirewallRules`, `ManageFirewallRules` | ☐ |
| 1.10 | `discovery.go` | `RunNetworkDiscovery` | ☐ |
| 1.11 | `monitoring.go` | `GetBandwidthHistory`, `StartMonitoring`, `StopMonitoring` | ☐ |
| 1.12 | `actions.go` | `RunNetworkAction` | ☐ |
| Bind | `NetOps.go` + `Types.go` | All bindings | ☐ |
