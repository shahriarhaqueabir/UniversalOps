package networkdesign

import (
	"fmt"
	"net"
	"strings"
)

// AnalyzeTopology takes nodes + edges and returns a health report.
func AnalyzeTopology(nodes []TopologyNode, edges []TopologyEdge) TopologyHealth {
	h := TopologyHealth{
		TotalNodes: len(nodes),
		TotalEdges: len(edges),
	}

	// Build lookup maps.
	nodeSet := make(map[string]TopologyNode, len(nodes))
	for _, n := range nodes {
		nodeSet[n.ID] = n
	}

	connected := make(map[string]bool)

	// 1. Broken links and missing labels on edges.
	for _, e := range edges {
		_, srcOK := nodeSet[e.Source]
		_, tgtOK := nodeSet[e.Target]
		if !srcOK || !tgtOK {
			h.BrokenLinks++
		} else {
			connected[e.Source] = true
			connected[e.Target] = true
		}
		if strings.TrimSpace(e.Label) == "" {
			h.MissingLabels++
		}
	}

	// 2. Missing labels on nodes.
	for _, n := range nodes {
		if strings.TrimSpace(n.Label) == "" {
			h.MissingLabels++
		}
	}

	// 3. Orphan nodes — nodes with no connections.
	for _, n := range nodes {
		if !connected[n.ID] {
			h.OrphanNodes = append(h.OrphanNodes, n.ID)
		}
	}

	// 4. Duplicate IPs — multiple nodes sharing a non-empty IP.
	ipMap := make(map[string][]string)
	for _, n := range nodes {
		ip := strings.TrimSpace(n.IP)
		if ip == "" {
			continue
		}
		ipMap[ip] = append(ipMap[ip], n.ID)
	}
	for ip, ids := range ipMap {
		if len(ids) > 1 {
			h.DuplicateIPs = append(h.DuplicateIPs, DuplicateIP{IP: ip, Nodes: ids})
		}
	}

	// 5. Subnet consistency — nodes directly connected should be in the same
	//    subnet (if both have subnet information).
	for _, e := range edges {
		src, srcOK := nodeSet[e.Source]
		tgt, tgtOK := nodeSet[e.Target]
		if !srcOK || !tgtOK {
			continue
		}
		if src.Subnet == "" || tgt.Subnet == "" {
			continue
		}
		if src.Subnet != tgt.Subnet {
			// Validate that it's actually different and not just cosmetic.
			if !subnetsOverlap(src.IP, src.Subnet, tgt.IP, tgt.Subnet) {
				h.SubnetErrors = append(h.SubnetErrors,
					fmt.Sprintf("Nodes %s (%s/%s) and %s (%s/%s) are on different subnets",
						src.ID, src.IP, src.Subnet, tgt.ID, tgt.IP, tgt.Subnet))
			}
		}
	}

	// 6. Generate suggestions based on findings.
	h.Suggestions = generateSuggestions(h, nodes, edges)

	return h
}

// subnetsOverlap checks whether two IP/subnet pairs are on the same network.
func subnetsOverlap(ip1, mask1, ip2, mask2 string) bool {
	a1 := net.ParseIP(ip1)
	a2 := net.ParseIP(ip2)
	if a1 == nil || a2 == nil {
		return true // can't determine, assume OK
	}

	// Build IPNet for each pair. Accept both CIDR notation ("/24") and dotted mask.
	net1 := parseIPNet(a1, mask1)
	net2 := parseIPNet(a2, mask2)
	if net1 == nil || net2 == nil {
		return true
	}

	// Two nodes are on the same subnet if their network prefixes match.
	prefix1 := net1.IP.Mask(net1.Mask)
	prefix2 := net2.IP.Mask(net2.Mask)
	return prefix1.Equal(prefix2)
}

// parseIPNet builds an IPNet from an IP and a mask.
// The mask parameter can be:
//   - CIDR prefix length ("24")
//   - Dotted mask ("255.255.255.0")
//   - Full CIDR notation ("192.168.1.0/24") — in this case the ip parameter is ignored.
func parseIPNet(ip net.IP, mask string) *net.IPNet {
	// If mask already contains "/", treat it as a complete CIDR.
	if strings.Contains(mask, "/") {
		if _, ipNet, err := net.ParseCIDR(mask); err == nil {
			return ipNet
		}
		return nil
	}

	// Try as CIDR notation (e.g. "192.168.1.10/24").
	cidr := ip.String() + "/" + mask
	if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
		return ipNet
	}

	// Try as dotted mask (e.g. "255.255.255.0").
	m := net.IPMask(net.ParseIP(mask).To4())
	if len(m) == 4 {
		return &net.IPNet{IP: ip.Mask(m), Mask: m}
	}
	return nil
}

// generateSuggestions produces actionable advice from the health data.
func generateSuggestions(h TopologyHealth, nodes []TopologyNode, edges []TopologyEdge) []string {
	var s []string

	if h.BrokenLinks > 0 {
		s = append(s, fmt.Sprintf("Fix %d broken link(s) — edges reference nodes that don't exist in the topology", h.BrokenLinks))
	}

	if h.MissingLabels > 0 {
		s = append(s, fmt.Sprintf("Add labels to %d node(s)/edge(s) for better documentation", h.MissingLabels))
	}

	if len(h.OrphanNodes) > 0 {
		s = append(s, fmt.Sprintf("%d node(s) have no connections: consider connecting or removing them", len(h.OrphanNodes)))
	}

	if len(h.DuplicateIPs) > 0 {
		s = append(s, fmt.Sprintf("%d IP address(es) are duplicated across nodes — check for conflicts", len(h.DuplicateIPs)))
	}

	if len(h.SubnetErrors) > 0 {
		s = append(s, fmt.Sprintf("%d subnet mismatch(es) detected between directly connected nodes", len(h.SubnetErrors)))
	}

	// Structural suggestions.
	routerCount := 0
	switchCount := 0
	for _, n := range nodes {
		switch strings.ToLower(n.Type) {
		case "router":
			routerCount++
		case "switch":
			switchCount++
		}
	}

	if len(nodes) > 3 && routerCount == 0 {
		s = append(s, "Consider adding a router for inter-network communication")
	}
	if len(nodes) > 5 && switchCount == 0 {
		s = append(s, "Consider adding a switch for local network segmentation")
	}

	// Check for nodes without IPs that should have them.
	noIPCount := 0
	for _, n := range nodes {
		if n.IP == "" && n.Type != "cloud" {
			noIPCount++
		}
	}
	if noIPCount > 0 {
		s = append(s, fmt.Sprintf("%d device(s) are missing IP addresses", noIPCount))
	}

	if len(s) == 0 {
		s = append(s, "Topology looks healthy — no issues detected")
	}

	return s
}
