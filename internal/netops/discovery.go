package netops

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

// DiscoveredDevice holds info about a discovered network device.
type DiscoveredDevice struct {
	IP             string `json:"ip"`
	MAC            string `json:"mac"`
	Vendor         string `json:"vendor"`
	Hostname       string `json:"hostname"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	IsGateway      bool   `json:"is_gateway"`
}

// DiscoveryResult holds the results of a network discovery scan.
type DiscoveryResult struct {
	Devices    []DiscoveredDevice `json:"devices"`
	Subnet     string             `json:"subnet"`
	ScanTimeMs int64              `json:"scan_time_ms"`
}

// GetLocalSubnet attempts to identify the primary local IPv4 subnet.
func GetLocalSubnet() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					// Return base (e.g. 192.168.1.0/24)
					return ipnet.String()
				}
			}
		}
	}
	return ""
}

// RunNetworkDiscovery discovers devices on a subnet using ARP + ping sweep.
func RunNetworkDiscovery(subnet string) DiscoveryResult {
	if subnet == "" {
		subnet = GetLocalSubnet()
	}

	start := time.Now()
	result := DiscoveryResult{Subnet: subnet}
	seen := make(map[string]bool)
	gateway := GetDefaultGateway().IP

	// 1. Initial Load from ARP table (High Reliability)
	for _, arp := range arpEntriesSafe() {
		if arp.MAC != "" && !seen[arp.IP] {
			seen[arp.IP] = true
			result.Devices = append(result.Devices, DiscoveredDevice{
				IP:        arp.IP,
				MAC:       arp.MAC,
				Vendor:    arp.Vendor,
				IsGateway: arp.IP == gateway,
			})
		}
	}

	// 2. Active Ping Sweep for discovery of new hosts
	if subnet != "" {
		hosts := generateSubnetHostsFromCIDR(subnet)
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := semaphore.NewWeighted(64) // Concurrent pings
		ctx := context.Background()

		for _, host := range hosts {
			if seen[host] {
				continue
			}
			wg.Add(1)
			_ = sem.Acquire(ctx, 1)

			go func(ip string) {
				defer wg.Done()
				defer sem.Release(1)

				pingStart := time.Now()
				pingResult, err := Ping(ip, 1)
				if err == nil && pingResult.Lost == 0 {
					// Host is up, try to get MAC from ARP again now that it's in cache
					mac := ""
					vendor := "Unknown"
					if newArp, _ := GetARPTable(); newArp != nil {
						for _, a := range newArp {
							if a.IP == ip {
								mac = a.MAC
								vendor = a.Vendor
								break
							}
						}
					}

					mu.Lock()
					result.Devices = append(result.Devices, DiscoveredDevice{
						IP:             ip,
						MAC:            mac,
						Vendor:         vendor,
						ResponseTimeMs: time.Since(pingStart).Milliseconds(),
					})
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

func generateSubnetHostsFromCIDR(cidr string) []string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}

	var hosts []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		hosts = append(hosts, ip.String())
	}

	// Filter out network and broadcast addresses
	if len(hosts) > 2 {
		return hosts[1 : len(hosts)-1]
	}
	return hosts
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
