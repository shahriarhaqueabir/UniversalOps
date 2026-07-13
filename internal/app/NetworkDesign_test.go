package app

import (
	"os"
	"testing"
)

func TestNetworkDesignAPI_New(t *testing.T) {
	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	if nd == nil {
		t.Fatal("NewNetworkDesignAPI returned nil")
	}
}

func TestNetworkDesignAPI_DiscoverLocalNetwork(t *testing.T) {
	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	nodes := nd.DiscoverLocalNetwork()
	if nodes == nil {
		t.Fatal("DiscoverLocalNetwork returned nil, expected non-nil slice")
	}
}

func TestNetworkDesignAPI_AnalyzeTopology_Empty(t *testing.T) {
	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	health := nd.AnalyzeTopology()
	if health.TotalNodes != 0 {
		t.Errorf("AnalyzeTopology.TotalNodes = %d, want 0", health.TotalNodes)
	}
	if health.TotalEdges != 0 {
		t.Errorf("AnalyzeTopology.TotalEdges = %d, want 0", health.TotalEdges)
	}
}

func TestNetworkDesignAPI_SetTopology_Valid(t *testing.T) {
	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	err := nd.SetTopology("[]", "[]")
	if err != nil {
		t.Fatalf("SetTopology failed: %v", err)
	}
	health := nd.AnalyzeTopology()
	if health.TotalNodes != 0 {
		t.Errorf("TotalNodes = %d, want 0", health.TotalNodes)
	}
}

func TestNetworkDesignAPI_SetTopology_InvalidJSON(t *testing.T) {
	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	err := nd.SetTopology("{invalid}", "[]")
	if err == nil {
		t.Error("SetTopology with invalid JSON should return error")
	}
}

func TestNetworkDesignAPI_SaveTopology_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	health, err := nd.SaveTopology()
	if err != nil {
		t.Fatalf("SaveTopology failed: %v", err)
	}
	if health.TotalNodes != 0 {
		t.Errorf("SaveTopology.TotalNodes = %d, want 0", health.TotalNodes)
	}
}

func TestNetworkDesignAPI_LoadTopology_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	nodesJSON, edgesJSON, health, err := nd.LoadTopology()
	if err != nil {
		t.Fatalf("LoadTopology failed: %v", err)
	}
	if nodesJSON != "[]" {
		t.Errorf("nodesJSON = %q, want %q", nodesJSON, "[]")
	}
	if edgesJSON != "[]" {
		t.Errorf("edgesJSON = %q, want %q", edgesJSON, "[]")
	}
	if health.TotalNodes != 0 {
		t.Errorf("LoadTopology.TotalNodes = %d, want 0", health.TotalNodes)
	}
}

func TestNetworkDesignAPI_SaveAndLoadTopology(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	a := NewApp()
	nd := NewNetworkDesignAPI(a)

	nd.SetTopology("[]", "[]")
	_, err := nd.SaveTopology()
	if err != nil {
		t.Fatalf("SaveTopology failed: %v", err)
	}

	a2 := NewApp()
	nd2 := NewNetworkDesignAPI(a2)
	_, _, health, err := nd2.LoadTopology()
	if err != nil {
		t.Fatalf("LoadTopology failed: %v", err)
	}
	if health.TotalNodes != 0 {
		t.Errorf("TotalNodes = %d, want 0", health.TotalNodes)
	}
}

func TestNetworkDesignAPI_GetInventory(t *testing.T) {
	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	inventory := nd.GetInventory()
	if inventory == nil {
		t.Fatal("GetInventory returned nil, expected non-nil map")
	}
}

func TestNetworkDesignAPI_GetInventory_WithData(t *testing.T) {
	a := NewApp()
	nd := NewNetworkDesignAPI(a)
	err := nd.SetTopology(`[{"id":"r1","label":"Router","type":"router","vendor":"Cisco"}]`, `[]`)
	if err != nil {
		t.Fatalf("SetTopology failed: %v", err)
	}
	inventory := nd.GetInventory()
	totalDevices, _ := inventory["totalDevices"].(int)
	if totalDevices != 1 {
		t.Errorf("totalDevices = %d, want 1", totalDevices)
	}
}

func TestNetworkDesignAPI_TopologyFile(t *testing.T) {
	if topologyFile != "topology.json" {
		t.Errorf("topologyFile = %q, want %q", topologyFile, "topology.json")
	}
}

func TestNetworkDesignAPI_healthCheckProbe_InvalidPort(t *testing.T) {
	status := healthCheckProbe(0)
	if status == "" {
		t.Error("healthCheckProbe returned empty string")
	}
}

func TestNetworkDesignAPI_healthCheckProbe_Unreachable(t *testing.T) {
	status := healthCheckProbe(9999)
	if status == "" {
		t.Error("healthCheckProbe returned empty string")
	}
}

func TestNetworkDesignAPI_detectFramework(t *testing.T) {
	tests := []struct {
		process string
		want    string
	}{
		{"node.exe", "Node.js"},
		{"python3", "Python"},
		{"dockerd", "Docker"},
		{"mysqld", ""},
		{"java.exe", "Java"},
		{"", ""},
	}
	for _, tt := range tests {
		got := detectFramework(tt.process)
		if got != tt.want {
			t.Errorf("detectFramework(%q) = %q, want %q", tt.process, got, tt.want)
		}
	}
}

func TestNetworkDesignAPI_formatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{2048, "2.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestNetworkDesignAPI_parseIntOr(t *testing.T) {
	tests := []struct {
		s        string
		fallback int
		want     int
	}{
		{"42", 0, 42},
		{"abc", 10, 10},
		{"", 5, 5},
	}
	for _, tt := range tests {
		got := parseIntOr(tt.s, tt.fallback)
		if got != tt.want {
			t.Errorf("parseIntOr(%q, %d) = %d, want %d", tt.s, tt.fallback, got, tt.want)
		}
	}
}

func TestNetworkDesignAPI_findGitRepos(t *testing.T) {
	repos := findGitRepos(5)
	if repos == nil {
		t.Log("findGitRepos returned nil (no git repos found)")
	}
}

func TestNetworkDesignAPI_gitRun(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	output := gitRun(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if output == "" {
		t.Log("gitRun returned empty (not a git repo or git not installed)")
	}
}
