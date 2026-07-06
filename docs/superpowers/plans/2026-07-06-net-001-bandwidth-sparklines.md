# NET-001: Bandwidth Sparklines Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add real-time bandwidth monitoring with visual sparklines to the NetOps interface view using a non-blocking delta-based approach.

**Architecture:** Refactor `GetInterfaces` to be stateless, store counters in the `Model`, and use periodic `TickMsg` triggers to calculate rates and update sparkline history.

**Tech Stack:** Go, Bubble Tea v2, Lip Gloss v2, gopsutil.

---

### Task 1: Refactor `GetInterfaces` for Non-Blocking Operation

**Files:**
- Modify: `internal/netops/interfaces.go`

- [ ] **Step 1: Update `InterfaceInfo` and add `BandwidthResult`**

Update `InterfaceInfo` to include more metadata and a `BandwidthResult` struct for the return type of the refactored function.

```go
// InterfaceInfo holds information about a network interface.
type InterfaceInfo struct {
	Name      string
	MAC       string
	IPs       []string
	IsUp      bool
	Speed     string
	MTU       int
	Flags     string
	RXBytes   uint64
	TXBytes   uint64
	RXRateBps float64
	TXRateBps float64
	RXHistory []float64
	TXHistory []float64
}

type BandwidthResult struct {
	Interfaces []InterfaceInfo
	Counters   map[string]bandwidthCounter
}
```

- [ ] **Step 2: Refactor `GetInterfaces` to remove `time.Sleep`**

Modify `GetInterfaces` to accept `lastCounters` and `elapsed` time.

```go
func GetInterfaces(lastCounters map[string]bandwidthCounter, elapsed time.Duration) (BandwidthResult, error) {
	current, err := getBandwidthCounters()
	if err != nil {
		return BandwidthResult{}, err
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return BandwidthResult{}, fmt.Errorf("get interfaces: %w", err)
	}

	var rates map[string]bandwidthRate
	if lastCounters != nil && elapsed > 0 {
		rates = calculateBandwidthRates(lastCounters, current, elapsed)
	}

	var result []InterfaceInfo
	for _, iface := range interfaces {
		info := InterfaceInfo{
			Name:  iface.Name,
			IsUp:  iface.Flags&net.FlagUp != 0,
			MTU:   iface.MTU,
			Flags: iface.Flags.String(),
		}
		if iface.HardwareAddr != nil {
			info.MAC = iface.HardwareAddr.String()
		}
		if rate, ok := rates[iface.Name]; ok {
			info.RXBytes = rate.RXBytes
			info.TXBytes = rate.TXBytes
			info.RXRateBps = rate.RXRateBps
			info.TXRateBps = rate.TXRateBps
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			info.IPs = append(info.IPs, addr.String())
		}
		result = append(result, info)
	}

	return BandwidthResult{Interfaces: result, Counters: current}, nil
}
```

- [ ] **Step 3: Update `calculateBandwidthRates` parameter names for clarity**

Ensure it uses the same terminology.

- [ ] **Step 4: Commit**

```bash
git add internal/netops/interfaces.go
git commit -m "refactor: make GetInterfaces stateless and non-blocking"
```

### Task 2: Update NetOps Model and Messages

**Files:**
- Modify: `internal/netops/model.go`
- Modify: `internal/netops/update.go`

- [ ] **Step 1: Update `Model` struct in `model.go`**

Add state tracking for counters and selection.

```go
type Model struct {
    // ... existing fields
    InterfaceData []InterfaceInfo
    lastCounters  map[string]bandwidthCounter
    lastCapture   time.Time
    selectedIndex int
}
```

- [ ] **Step 2: Update `InterfacesResultMsg` in `update.go`**

```go
type InterfacesResultMsg struct {
	Interfaces []InterfaceInfo
	Counters   map[string]bandwidthCounter
	Err        error
}
```

- [ ] **Step 3: Update `Update` handler for `InterfacesResultMsg`**

```go
case InterfacesResultMsg:
    if msg.Err != nil {
        m.err = msg.Err
    } else {
        m.InterfaceData = mergeInterfaceBandwidthHistory(m.InterfaceData, msg.Interfaces)
        m.lastCounters = msg.Counters
        m.lastCapture = time.Now()
        m.err = nil
        m.ready = true
    }
```

- [ ] **Step 4: Update `refreshCurrentTab` and `actOnCurrentTab` in `update.go`**

Update the call to `GetInterfaces` to pass the stored state.

```go
case 5: // Interfaces
    last := m.lastCounters
    elapsed := time.Since(m.lastCapture)
    return func() tea.Msg {
        res, err := GetInterfaces(last, elapsed)
        return InterfacesResultMsg{Interfaces: res.Interfaces, Counters: res.Counters, Err: err}
    }
```

- [ ] **Step 5: Commit**

```bash
git add internal/netops/model.go internal/netops/update.go
git commit -m "feat: wire bandwidth counters and history into NetOps model"
```

### Task 3: Implement Detail Panel and Sparklines in View

**Files:**
- Modify: `internal/netops/view.go`

- [ ] **Step 1: Add `handleKeyPress` navigation for selection**

Update `handleKeyPress` in `update.go` to handle `up`/`down` arrows for interface selection.

- [ ] **Step 2: Implement `renderInterfacesDetail` helper in `view.go`**

```go
func (m *Model) renderInterfacesDetail(width int) string {
    if m.selectedIndex >= len(m.InterfaceData) {
        return "Select an interface to see details"
    }
    iface := m.InterfaceData[m.selectedIndex]
    
    // Left: Stats
    stats := fmt.Sprintf("MAC: %s\nIPs: %s\nMTU: %d\nFlags: %s", 
        iface.MAC, strings.Join(iface.IPs, ", "), iface.MTU, iface.Flags)
    
    // Right: Sparklines
    rxSpark := common.RenderSparkline(iface.RXHistory, width/3, common.Styles.Info)
    txSpark := common.RenderSparkline(iface.TXHistory, width/3, common.Styles.Secondary)
    
    return lipgloss.JoinHorizontal(lipgloss.Top, 
        common.Styles.Border.Render(stats),
        lipgloss.JoinVertical(lipgloss.Left, 
            "RX: "+common.FormatBytes(uint64(iface.RXRateBps))+"/s\n"+rxSpark,
            "TX: "+common.FormatBytes(uint64(iface.TXRateBps))+"/s\n"+txSpark,
        ),
    )
}
```

- [ ] **Step 3: Update `View` to split the screen**

If `tabIndex == 5`, render the list on top and the detail panel on the bottom.

- [ ] **Step 4: Commit**

```bash
git add internal/netops/view.go internal/netops/update.go
git commit -m "feat: add interface detail panel with RX/TX sparklines"
```

### Task 4: Verification and Testing

**Files:**
- Create: `internal/netops/bandwidth_test.go`

- [ ] **Step 1: Write unit tests for rate calculation**

Test `calculateBandwidthRates` with mock counters and known elapsed time.

- [ ] **Step 2: Run all tests**

Run: `go test ./internal/netops/...`

- [ ] **Step 3: Manual TUI Test**

Build and run `hawkward.exe`, go to NetOps -> Interfaces, select an interface, and verify sparklines update on `r` (refresh) or auto-tick.

- [ ] **Step 4: Commit**

```bash
git add internal/netops/bandwidth_test.go
git commit -m "test: verify bandwidth rate calculation and sparkline logic"
```
