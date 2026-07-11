package networkdesign

import "testing"

func TestAnalyzeTopology_Empty(t *testing.T) {
	h := AnalyzeTopology(nil, nil)
	if h.TotalNodes != 0 || h.TotalEdges != 0 {
		t.Errorf("expected 0 nodes/edges, got %d/%d", h.TotalNodes, h.TotalEdges)
	}
	if len(h.Suggestions) == 0 {
		t.Error("expected at least one suggestion for empty topology")
	}
}

func TestAnalyzeTopology_BrokenLinks(t *testing.T) {
	nodes := []TopologyNode{
		{ID: "a", Label: "A"},
		{ID: "b", Label: "B"},
	}
	edges := []TopologyEdge{
		{ID: "e1", Source: "a", Target: "b"},
		{ID: "e2", Source: "a", Target: "c"}, // "c" doesn't exist
	}
	h := AnalyzeTopology(nodes, edges)
	if h.BrokenLinks != 1 {
		t.Errorf("expected 1 broken link, got %d", h.BrokenLinks)
	}
}

func TestAnalyzeTopology_OrphanNodes(t *testing.T) {
	nodes := []TopologyNode{
		{ID: "a", Label: "A"},
		{ID: "b", Label: "B"},
		{ID: "c", Label: "C"},
	}
	edges := []TopologyEdge{
		{ID: "e1", Source: "a", Target: "b"},
	}
	h := AnalyzeTopology(nodes, edges)
	if len(h.OrphanNodes) != 1 || h.OrphanNodes[0] != "c" {
		t.Errorf("expected orphan [c], got %v", h.OrphanNodes)
	}
}

func TestAnalyzeTopology_DuplicateIPs(t *testing.T) {
	nodes := []TopologyNode{
		{ID: "a", Label: "A", IP: "10.0.0.1"},
		{ID: "b", Label: "B", IP: "10.0.0.1"},
		{ID: "c", Label: "C", IP: "10.0.0.2"},
	}
	h := AnalyzeTopology(nodes, nil)
	if len(h.DuplicateIPs) != 1 {
		t.Fatalf("expected 1 duplicate IP group, got %d", len(h.DuplicateIPs))
	}
	if h.DuplicateIPs[0].IP != "10.0.0.1" {
		t.Errorf("expected duplicate IP 10.0.0.1, got %s", h.DuplicateIPs[0].IP)
	}
}

func TestAnalyzeTopology_MissingLabels(t *testing.T) {
	nodes := []TopologyNode{
		{ID: "a", Label: ""}, // missing label
		{ID: "b", Label: "B"},
	}
	edges := []TopologyEdge{
		{ID: "e1", Source: "a", Target: "b", Label: ""}, // missing label
	}
	h := AnalyzeTopology(nodes, edges)
	if h.MissingLabels != 2 {
		t.Errorf("expected 2 missing labels (1 node + 1 edge), got %d", h.MissingLabels)
	}
}

func TestAnalyzeTopology_SubnetMismatch(t *testing.T) {
	nodes := []TopologyNode{
		{ID: "a", Label: "A", IP: "192.168.1.10", Subnet: "192.168.1.0/24"},
		{ID: "b", Label: "B", IP: "10.0.0.5", Subnet: "10.0.0.0/24"},
	}
	edges := []TopologyEdge{
		{ID: "e1", Source: "a", Target: "b", Label: "link"},
	}
	h := AnalyzeTopology(nodes, edges)
	if len(h.SubnetErrors) != 1 {
		t.Errorf("expected 1 subnet error, got %d", len(h.SubnetErrors))
	}
}

func TestAnalyzeTopology_HealthyTopology(t *testing.T) {
	nodes := []TopologyNode{
		{ID: "r", Label: "Router", IP: "192.168.1.1", Subnet: "192.168.1.0/24", Type: "router"},
		{ID: "s", Label: "Switch", IP: "192.168.1.2", Subnet: "192.168.1.0/24", Type: "switch"},
		{ID: "w", Label: "Workstation", IP: "192.168.1.10", Subnet: "192.168.1.0/24", Type: "workstation"},
	}
	edges := []TopologyEdge{
		{ID: "e1", Source: "r", Target: "s", Label: "uplink"},
		{ID: "e2", Source: "s", Target: "w", Label: "access"},
	}
	h := AnalyzeTopology(nodes, edges)
	if h.BrokenLinks != 0 {
		t.Errorf("expected 0 broken links, got %d", h.BrokenLinks)
	}
	if len(h.OrphanNodes) != 0 {
		t.Errorf("expected 0 orphan nodes, got %d", len(h.OrphanNodes))
	}
	if len(h.DuplicateIPs) != 0 {
		t.Errorf("expected 0 duplicate IPs, got %d", len(h.DuplicateIPs))
	}
	if len(h.SubnetErrors) != 0 {
		t.Errorf("expected 0 subnet errors, got %d", len(h.SubnetErrors))
	}
	for _, s := range h.Suggestions {
		if s == "Topology looks healthy — no issues detected" {
			return // good
		}
	}
	t.Error("expected 'healthy' suggestion")
}

func TestParseARPOutput_Windows(t *testing.T) {
	output := `Interface: 192.168.1.5 --- 0x3
  Internet Address      Physical Address      Type
  192.168.1.1           aa-bb-cc-dd-ee-ff     dynamic
  192.168.1.100         11-22-33-44-55-66     dynamic
  192.168.1.255         ff-ff-ff-ff-ff-ff     static
`
	entries := parseARPOutput(output)
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 ARP entries, got %d", len(entries))
	}
	if entries[0].IP != "192.168.1.1" {
		t.Errorf("expected first IP 192.168.1.1, got %s", entries[0].IP)
	}
	if entries[0].MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("expected MAC aa:bb:cc:dd:ee:ff, got %s", entries[0].MAC)
	}
}

func TestParseARPOutput_Linux(t *testing.T) {
	output := `? (192.168.1.1) at aa:bb:cc:dd:ee:ff [ether] on eth0
? (192.168.1.100) at 11:22:33:44:55:66 [ether] on eth0
`
	entries := parseARPOutput(output)
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 ARP entries, got %d", len(entries))
	}
	if entries[0].IP != "192.168.1.1" {
		t.Errorf("expected first IP 192.168.1.1, got %s", entries[0].IP)
	}
}

func TestLookupOUI(t *testing.T) {
	tests := []struct {
		mac      string
		expected string
	}{
		{"aa:bb:cc:dd:ee:ff", ""},
		{"00:50:56:12:34:56", "VMware"},
		{"b8:27:eb:12:34:56", "Raspberry Pi Foundation"},
		{"00:1b:21:12:34:56", "Intel"},
		{"", ""},
	}
	for _, tt := range tests {
		got := lookupOUI(tt.mac)
		if got != tt.expected {
			t.Errorf("lookupOUI(%q) = %q, want %q", tt.mac, got, tt.expected)
		}
	}
}
