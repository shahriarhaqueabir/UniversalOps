package networkdesign

// TopologyNode represents a device in the topology.
type TopologyNode struct {
	ID     string            `json:"id"`
	Type   string            `json:"type"` // "router", "switch", "server", "workstation", "firewall", "cloud"
	Label  string            `json:"label"`
	IP     string            `json:"ip"`
	Subnet string            `json:"subnet"`
	MAC    string            `json:"mac"`
	Notes  string            `json:"notes"`
	Vendor string            `json:"vendor"`
	VLAN   string            `json:"vlan"`
	Online bool              `json:"online"`
	Props  map[string]string `json:"props"`
}

// TopologyEdge represents a connection between two nodes.
type TopologyEdge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Label     string `json:"label"`
	Type      string `json:"type"` // "ethernet", "fiber", "wireless", "vpn", "wan"
	Bandwidth string `json:"bandwidth"`
	Status    string `json:"status"` // "up", "down", "unknown"
}

// Topology represents the full network topology.
type Topology struct {
	Nodes  []TopologyNode `json:"nodes"`
	Edges  []TopologyEdge `json:"edges"`
	Health TopologyHealth `json:"health"`
}

// TopologyHealth holds the results of topology analysis.
type TopologyHealth struct {
	TotalNodes    int           `json:"totalNodes"`
	TotalEdges    int           `json:"totalEdges"`
	BrokenLinks   int           `json:"brokenLinks"`
	MissingLabels int           `json:"missingLabels"`
	OrphanNodes   []string      `json:"orphanNodes"`
	DuplicateIPs  []DuplicateIP `json:"duplicateIPs"`
	SubnetErrors  []string      `json:"subnetErrors"`
	Suggestions   []string      `json:"suggestions"`
}

// DuplicateIP tracks multiple nodes sharing the same IP.
type DuplicateIP struct {
	IP    string   `json:"ip"`
	Nodes []string `json:"nodes"`
}
