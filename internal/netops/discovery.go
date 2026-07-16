package netops

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

// DiscoveredDevice holds info about a discovered network device.
type DiscoveredDevice struct {
	IP              string `json:"ip"`
	MAC             string `json:"mac"`
	Vendor          string `json:"vendor"`
	Hostname        string `json:"hostname"`
	OpenPorts       []int  `json:"open_ports"`
	ResponseTimeMs  int64  `json:"response_time_ms"`
}

// DiscoveryResult holds the results of a network discovery scan.
type DiscoveryResult struct {
	Devices    []DiscoveredDevice `json:"devices"`
	Subnet     string             `json:"subnet"`
	ScanTimeMs int64              `json:"scan_time_ms"`
}

// RunNetworkDiscovery discovers devices on a subnet using ARP + ping sweep.
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
		sem := semaphore.NewWeighted(32) // Limit concurrency to 32 pings
		ctx := context.Background()

		for _, host := range generateSubnetHosts(subnet) {
			if seen[host] {
				continue
			}
			wg.Add(1)
			if err := sem.Acquire(ctx, 1); err != nil {
				wg.Done()
				continue
			}

			go func(ip string) {
				defer wg.Done()
				defer sem.Release(1)

				pingStart := time.Now()
				pingResult, err := Ping(ip, 1)
				if err == nil && pingResult.Lost < pingResult.Sent {
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
