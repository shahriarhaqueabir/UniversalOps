# Design Doc: NET-001 — Bandwidth Sparklines

Add real-time bandwidth monitoring with visual sparklines to the NetOps interface view.

## 1. Problem Statement
The current `NetOps` interface view lists network interfaces but lacks real-time traffic visualization. Furthermore, the existing bandwidth calculation logic in `interfaces.go` uses `time.Sleep(500ms)`, which blocks the main TUI update loop and causes significant UI lag.

## 2. Proposed Solution
We will implement a non-blocking, delta-based bandwidth monitoring system that integrates seamlessly into the Bubble Tea "The Elm Architecture" (TEA).

### 2.1 Architecture Changes
- **Refactor `GetInterfaces`**: Move from a self-contained "measure-sleep-measure" model to a "stateless" model.
    - `GetInterfaces` will now accept the `lastCounters` map and the `elapsed` time since the last capture.
    - It will return the new `InterfaceInfo` array and the latest `bandwidthCounter` map.
- **Model Updates**: 
    - Store `lastCounters map[string]bandwidthCounter` and `lastCapture time.Time` in `netops.Model`.
    - Track `selectedIndex int` to enable a "Detail Panel" view for the focused interface.
- **Update Loop**: 
    - The `root.go` already sends a periodic `TickMsg`. `netops.Update` will use this trigger to fire off a `tea.Cmd` that fetches fresh counters.
    - On `InterfacesResultMsg`, the model calculates rates based on the actual time elapsed between samples.

### 2.2 UI Design (Detail Panel)
The `NetOps` view for the "Interfaces" tab (Index 5) will be split into two sections:
- **Interface List (Top)**: Existing table showing interface names, status, and primary IP.
- **Detail Panel (Bottom)**: Triggered by selection.
    - **Stats (Left)**: MAC address, all IP addresses (IPv4/v6), estimated speed, total Bytes received/sent, and interface flags (Up, Loopback, Multicast, etc.).
    - **Trends (Right)**: Two sparklines rendered using `common.RenderSparkline` for RX and TX rates. Sparklines will be color-coded (Info for RX, Secondary for TX).

## 3. Data Flow
1. `TickMsg` received in `Update`.
2. `Model` checks if `Interfaces` tab is active.
3. If active, it sends a `tea.Cmd` to fetch `IOCounters`.
4. Command returns `InterfacesResultMsg` with fresh counters.
5. `Update` calculates: `rate = (current_bytes - last_bytes) / elapsed_seconds`.
6. `Update` appends rate to `InterfaceInfo.RXHistory/TXHistory`.
7. `Update` stores current counters and current time for the next cycle.
8. `View` renders the list and the detail panel for the `selectedIndex`.

## 4. Error Handling & Edge Cases
- **Initial Sample**: On the first capture (no `lastCounters`), the rates will be reported as 0.
- **Counter Reset**: `counterDelta` handles cases where a system restart or interface reset causes the new byte count to be lower than the previous one by returning 0.
- **Interface Disappearance**: If an interface exists in history but not in the fresh capture, its history is preserved but its status is marked down/missing.

## 5. Testing Plan
- **Unit Tests**:
    - `TestCalculateBandwidthRates`: Verify rate calculation given two sets of counters and a specific duration.
    - `TestCounterDelta`: Verify wrap-around/reset logic.
    - `TestAppendRateHistory`: Ensure history length is capped at `bandwidthHistoryLimit`.
- **Manual Verification**:
    - Run `hawkward.exe`, navigate to NetOps -> Interfaces.
    - Verify that the UI remains responsive (no 500ms lag).
    - Perform network activity (e.g., download a file) and verify sparkline movement.
