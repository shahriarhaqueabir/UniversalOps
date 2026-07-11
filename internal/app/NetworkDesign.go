package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/networkdesign"
)

// NetworkDesignAPI exposes topology analysis and discovery to the Wails frontend.
type NetworkDesignAPI struct {
	app   *App
	mu    sync.Mutex
	nodes []networkdesign.TopologyNode
	edges []networkdesign.TopologyEdge
}

// topologyFile is the filename for persisted topology data.
const topologyFile = "topology.json"

// topologyPersist is the on-disk structure for topology save/load.
type topologyPersist struct {
	Nodes []networkdesign.TopologyNode `json:"nodes"`
	Edges []networkdesign.TopologyEdge `json:"edges"`
}

// NewNetworkDesignAPI creates a new NetworkDesignAPI.
func NewNetworkDesignAPI(app *App) *NetworkDesignAPI {
	nd := &NetworkDesignAPI{app: app}
	nd.loadFromDisk()
	return nd
}

// DiscoverLocalNetwork scans the local machine for network devices via ARP
// table and interface enumeration. No admin/root required.
func (nd *NetworkDesignAPI) DiscoverLocalNetwork() []networkdesign.TopologyNode {
	return networkdesign.DiscoverLocalNetwork()
}

// AnalyzeTopology runs health analysis on the currently loaded topology.
func (nd *NetworkDesignAPI) AnalyzeTopology() networkdesign.TopologyHealth {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	return networkdesign.AnalyzeTopology(nd.nodes, nd.edges)
}

// SetTopology receives the full topology (nodes + edges) from the frontend
// and stores it in memory for analysis.
func (nd *NetworkDesignAPI) SetTopology(nodesJSON string, edgesJSON string) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	var nodes []networkdesign.TopologyNode
	if err := json.Unmarshal([]byte(nodesJSON), &nodes); err != nil {
		return fmt.Errorf("invalid nodes JSON: %w", err)
	}

	var edges []networkdesign.TopologyEdge
	if err := json.Unmarshal([]byte(edgesJSON), &edges); err != nil {
		return fmt.Errorf("invalid edges JSON: %w", err)
	}

	nd.nodes = nodes
	nd.edges = edges
	return nil
}

// SaveTopology persists the current topology to disk and returns the health.
func (nd *NetworkDesignAPI) SaveTopology() (networkdesign.TopologyHealth, error) {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	persist := topologyPersist{
		Nodes: nd.nodes,
		Edges: nd.edges,
	}

	data, err := json.MarshalIndent(persist, "", "  ")
	if err != nil {
		return networkdesign.TopologyHealth{}, fmt.Errorf("marshal topology: %w", err)
	}

	path := nd.filePath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return networkdesign.TopologyHealth{}, fmt.Errorf("write topology file: %w", err)
	}

	common.LogInfo("Topology saved to %s (%d nodes, %d edges)", path, len(nd.nodes), len(nd.edges))
	return networkdesign.AnalyzeTopology(nd.nodes, nd.edges), nil
}

// LoadTopology reads a topology from disk and returns it as JSON strings
// for nodes and edges, plus the health analysis.
func (nd *NetworkDesignAPI) LoadTopology() (string, string, networkdesign.TopologyHealth, error) {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	path := nd.filePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "[]", "[]", networkdesign.TopologyHealth{}, nil
		}
		return "", "", networkdesign.TopologyHealth{}, fmt.Errorf("read topology file: %w", err)
	}

	var persist topologyPersist
	if err := json.Unmarshal(data, &persist); err != nil {
		return "", "", networkdesign.TopologyHealth{}, fmt.Errorf("parse topology file: %w", err)
	}

	nd.nodes = persist.Nodes
	nd.edges = persist.Edges

	nodesJSON, _ := json.Marshal(nd.nodes)
	edgesJSON, _ := json.Marshal(nd.edges)

	h := networkdesign.AnalyzeTopology(nd.nodes, nd.edges)
	common.LogInfo("Topology loaded from %s (%d nodes, %d edges)", path, len(nd.nodes), len(nd.edges))
	return string(nodesJSON), string(edgesJSON), h, nil
}

// GetInventory returns an inventory summary of all devices in the topology.
func (nd *NetworkDesignAPI) GetInventory() map[string]interface{} {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	vendors := make(map[string]int)
	types := make(map[string]int)
	noIP := 0
	noMAC := 0

	for _, n := range nd.nodes {
		if n.Vendor != "" {
			vendors[n.Vendor]++
		}
		types[n.Type]++
		if n.IP == "" {
			noIP++
		}
		if n.MAC == "" {
			noMAC++
		}
	}

	return map[string]interface{}{
		"totalDevices": len(nd.nodes),
		"types":        types,
		"vendors":      vendors,
		"missingIPs":   noIP,
		"missingMACs":  noMAC,
	}
}

// loadFromDisk attempts to load a saved topology at startup.
func (nd *NetworkDesignAPI) loadFromDisk() {
	path := nd.filePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var persist topologyPersist
	if err := json.Unmarshal(data, &persist); err != nil {
		common.LogWarn("Failed to parse saved topology: %v", err)
		return
	}
	nd.nodes = persist.Nodes
	nd.edges = persist.Edges
	common.LogInfo("Loaded persisted topology (%d nodes, %d edges)", len(nd.nodes), len(nd.edges))
}

// filePath returns the absolute path to the topology JSON file.
func (nd *NetworkDesignAPI) filePath() string {
	// Store topology next to the database file.
	dir := "."
	return filepath.Join(dir, topologyFile)
}
