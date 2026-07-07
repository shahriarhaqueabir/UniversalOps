# Repo Research Report: UI/UX & Feature Inspiration for Hawkward Ops

> **Date:** 2026-07-07
> **Status:** Complete — all three repositories fetched and analyzed
> **Context:** Research phase of the Hawkward TUI overhaul — a Go + Bubble Tea v2 terminal operations platform (SysOps, NetOps, SecOps, DevOps, AIOps).
> **Method:** Fetched each repo's raw README from GitHub, analyzed UI patterns, tech stacks, and adaptable features.

---

## 1. netscanner — Rust TUI Network Scanner

**URL:** https://github.com/Chleba/netscanner  
**Stack:** Rust, Ratatui (Rust TUI framework), libpnet  
**Status:** Active — packaged for Arch, Nix, Kali, Alpine. Demo GIF in README.

### Overview

A terminal-based network scanner and diagnostic tool. Uses Ratatui (the Rust equivalent of Bubble Tea) for its TUI rendering and libpnet for raw packet capture. It runs as a single binary with root privileges and presents a tabbed multi-panel interface with real-time packet dumping, WiFi scanning, port scanning, and CIDR ping sweep.

### Key UI Patterns & Design Elements

- **Tabbed multi-panel layout** — switches between views: HW interfaces, WiFi scan, packet dump, port scan.
- **Real-time packet dump** — live scrolling view with color-coded protocol entries (TCP = one color, UDP = another, ICMP, ARP). Start/pause/resume controls are always visible.
- **Signal strength charts** — inline bar rendering of WiFi signal levels per access point. Uses terminal block characters for visual bar charts.
- **Interface selector** — switch active network interface for scanning/dumping (loopback, wlan, eth, vpn).
- **Sidebar + detail split** — left pane lists devices/interfaces, right pane shows scan results or packet details.
- **Footer status bar** — shows active interface name and running packet count.
- **Export to CSV** — scanned IPs, ports, and packet logs can be exported.
- **Root-required guard** — explicit warnings and `sudo chown` instructions. Not a soft check — the app fails gracefully if not root.

### Features That Could Be Adapted

| Feature | How It Maps to Hawkward |
|---|---|
| Real-time packet dump with start/pause/resume | **NetOps** — live packet capture view with protocol filter |
| Protocol color-coding (TCP/UDP/ICMP/ARP) | **NetOps** — color-coded log entries in terminal |
| WiFi signal strength bar charts | **NetOps** — inline bar charts for interface metrics (WiFi, bandwidth, RSSI) |
| Port scanning with progress feedback | **SecOps** — TCP port scan with a progress bar and results table |
| CIDR ping sweep with hostname/MAC/OUI | **NetOps** — network discovery with device fingerprinting |
| Traffic counting + DNS records | **NetOps** — live bandwidth and DNS query monitor |
| Interface selector (loopback, wlan, eth, vpn) | **NetOps** — multi-interface support, useful in VPN/server contexts |
| Export scanned IPs/ports/packets to CSV | **SysOps** — export any scan results, logs, metrics to file |

### Visual Description (from README demo GIF)

The terminal is split into two main panes. The left pane lists network interfaces with signal strength bars (block characters). The right pane shows a table of scanned hosts with columns: IP address, hostname, MAC address, OUI (vendor). Colors distinguish reachable hosts (green) from unreachable (red). A bottom status bar shows the active interface and packet count. The overall aesthetic is utilitarian but information-dense — prioritized data density over whitespace.

### Relevance to Hawkward: **High for NetOps module**

The packet dump, port scan, and CIDR sweep patterns map directly to the Hawkward NetOps and SecOps modules. The multi-tab layout with a footer status bar is a proven pattern for operations TUIs.

---

## 2. hackingtool — Python All-in-One Hacking Tool Menu System

**URL:** https://github.com/Z4nzu/hackingtool (original, verified accessible via raw content)  
**Stack:** Python 3.10+, Bash/Python hybrid menu system, Go 1.21+ (for tool dependencies), external tool wrappers  
**Stars:** ~50k+ (popular repo)  
**Status:** Active — v2.0.0, 185+ tools across 20 categories, Docker support

### Overview

A massive aggregator menu system that wraps 185+ security tools into a categorized, searchable interface. It doesn't build its own TUI framework — it uses a numbered bash/Python hybrid menu — but its **navigation and information architecture** is the star. v2.0.0 added `/` search, tag filtering, and a "recommend" mode, making it far more usable than a typical menu-based tool.

### Key UI Patterns & Design Elements

- **Category grid menu** — 20 numbered categories displayed in a two-column table on the main screen. Each entry has an emoji, category name, and tool count. Categories: Anonymity, Info Gathering, Wordlist, Wireless, SQLi, Phishing, Web Attack, Post-Exploitation, Forensics, Payload Creation, Exploit Framework, Reverse Engineering, DDoS, RAT, XSS, Steganography, Active Directory, Cloud Security, Mobile Security, Other.
- **`/` search** — type `/query` to search all 185+ tools by name, description, or keyword. This is the standout UX feature. Works from the main menu.
- **`t` tag filter** — type `t` to filter by 19 tags: osint, web, c2, cloud, mobile, scanner, etc.
- **`r` recommend** — natural language input — "I want to scan a network" → shows relevant tools.
- **Install status indicators** — ✔ (green checkmark) or ✘ (red cross) shown next to every tool. Know instantly what's installed.
- **Batch install** — option `97` installs all tools in a category in one command.
- **Smart update** — per-tool update command auto-detects `git pull` / `pip upgrade` / `go install` based on the tool's install type.
- **Open folder** — option to jump into any tool's directory for manual inspection.
- **`?` help** — quick reference card available at any screen.
- **`q` quit / `99` back** — consistent navigation at any depth. `q` exits the app, `99` goes back one level.
- **OS-aware menus** — Linux-only tools hidden on macOS automatically.
- **Docker support** — builds locally (no unverified external images), with dev mode for live source mounting.

### Navigation Architecture (Key Takeaway)

```
Main Menu (20 categories + search/tags/recommend)
  │
  ├── /query  →  Instant search across all tools
  ├── t       →  Tag filter (osint, web, scanner...)
  ├── r       →  Natural language recommendation
  ├── ?       →  Help card
  │
  ├── Category 1 → Tools list (with install status ✔/✘)
  │                  ├── 97  → Install all in category
  │                  └── Tool N → Info, Install/Update, Open Folder
  ├── Category 2 → ...
  └── ...
```

This is a **flat hierarchy with smart cross-cutting access** — you never need to drill down to find something because `/` search cuts through everything.

### Features That Could Be Adapted

| Feature | How It Maps to Hawkward |
|---|---|
| **`/` fuzzy search** | **All modules** — instant search across all ops tools, commands, views, hosts |
| **Tag filtering** | **All modules** — filter tools/views by tag: `sysops`, `netops`, `secops`, `devops`, `aiops` |
| **Recommend mode** | **AI/AIOps** — "I want to scan the network" → launches correct tool/view chain |
| **Install status indicators** | **SysOps** — show installed/available/not-installed status per tool with icon/color |
| **Category grid menu** | **Main navigation** — ops module selector as the default screen on launch |
| **`?` context help** | **All modules** — inline help overlay showing keybindings |
| **`99` back / `q` quit** | **All modules** — consistent back navigation (Esc or q) |
| **Batch/bulk operations** | **SysOps** — batch run commands across multiple hosts or services |
| **OS-aware filtering** | **SysOps** — hide tools not applicable to the current OS |
| **Per-tool update tracking** | **SysOps** — track version/update status of each bundled tool |

### Relevance to Hawkward: **Critical for Navigation**

The navigation patterns in hackingtool v2.0.0 are the single most important research finding for Hawkward. The `/` search, `t` tag filter, `r` recommend, and `?` help form a cohesive information-access system that transforms a potentially overwhelming tool collection into something discoverable and fast.

---

## 3. squib — Node/TS Card Dashboard (PRIMARY)

**URL:** https://github.com/teknetai/squib  
**Stack:** Node.js 22+, TypeScript, SQLite (`node:sqlite`), WebSocket, system `ping`, pure CSS dashboard  
**Status:** Active — production-grade, MIT license, CI pipeline, Docker support

### Overview

Squib is a network monitoring dashboard that tracks ICMP up/down status, latency, and packet loss with real-time WebSocket updates, flap-detection alerting, and a clean dark-themed web UI. It is the **primary inspiration** because its design sensibility, architecture, and feature set align most closely with the Hawkward vision — except Hawkward runs in the terminal rather than a browser.

Squib explicitly targets the gap between "powerful but painful" monitoring (Zabbix, Nagios) and "polished but heavy/pricey" solutions, aiming for something that can be stood up in minutes and produces a dashboard you'd "put on a NOC wall."

### Key UI Patterns & Design Elements

- **Dark theme by default** — modern dark background (#1a1a2e-ish), light text, accent colors for status: green (up), red (down), yellow (degraded), gray (unreachable/unknown).
- **Card-based host dashboard** — hosts rendered as individual cards in a responsive grid. Each card shows:
  - Hostname (bold, prominent)
  - IP address (secondary text)
  - Status badge (colored dot + label: up/down/degraded)
  - Latency (ms, with inline sparkline)
  - Packet loss (%)
  - Last-seen timestamp
- **Detail drawer** — clicking a host card opens a slide-in panel (from right) with:
  - RTT chart (sparkline over time)
  - State-change timeline/incident history
  - Inline edit/remove controls
  - Quick actions
- **Light/Dark mode toggle** — theme switch persisted across sessions.
- **Sortable host table** — toggle between card grid and table view, with sortable columns.
- **Live WebSocket updates** — no page refresh needed; latency, status, and packet-loss values push in real time.
- **Subnet discovery flow** — CIDR or range scan → review alive hosts → bulk-add them. Uses nmap if present, else built-in ICMP sweep.
- **Stealth access gate** — minimal, unbranded login page (`gate.html`) when `GATE_PASSWORD` is set. Guards the dashboard, every API route, and WebSocket.

### Architecture

```
 system `ping` ──▶ poller (flap-detecting state machine) ──▶ SQLite (time-series)
                          │                                       │
                     webhook alerts                  REST API + WebSocket ──▶ dashboard
```

Key architectural points:
- **No agents** — the poller runs in the server process, pinging hosts directly.
- **Flap detection state machine** — a host goes *down* only after N consecutive misses (configurable `FAIL_THRESHOLD`), *recovers* after M replies (`OK_THRESHOLD`). Prevents alert fatigue.
- **Time-series storage** — SQLite (`node:sqlite`, built-in) stores samples, incidents, host config. Schema is simple: hosts table, samples table, incidents table.
- **Webhook alerts** — one alert channel compatible with Slack, Teams, Discord, PagerDuty. Configured via env var or UI.

### Screenshot Descriptions

| Image | Content |
|---|---|
| `docs/overview.png` | Full dashboard with host card grid. Dark background. Each card shows hostname, IP, status (colored dot + text), latency with sparkline, packet loss %, last-seen. Cards arranged in a responsive grid. Green/gray/red status differentiation. Clean typography. |
| `docs/hosts.png` | Host detail drawer open from the right side. Shows RTT chart (area chart), state-change timeline (vertical timeline with status dots), incident list, and edit/delete controls for the selected host. |
| `docs/gate.png` | Minimal, unbranded access gate login page — just a centered password input and a "sign in" button on a dark background. No logo, no branding. |

### Configuration via Environment Variables

| Var | Default | Meaning |
|---|---|---|
| `PORT` / `HOST` | `4180` / `0.0.0.0` | Bind address |
| `INTERVAL` | `5000` | Poll interval (ms) |
| `PING_COUNT` | `3` | Packets per probe |
| `PING_TIMEOUT` | `4` | Seconds to wait per probe |
| `FAIL_THRESHOLD` | `3` | Misses before marking **down** |
| `OK_THRESHOLD` | `2` | Replies before marking **recovered** |
| `LAT_THRESHOLD` / `LOSS_THRESHOLD` | `120` / `5` | Thresholds for **degraded** status |
| `WEBHOOK_URL` | — | Alert webhook URL |
| `GATE_PASSWORD` | — | Access gate password (stealth login) |
| `DATA_DIR` | `data` | Where `squib.db` lives |
| `DEMO_SEED` | — | Load demo fleet when set |

### REST API Routes

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/hosts` | Fleet summary (status, latency, history) |
| GET | `/api/host/:id` | One host + state-change timeline |
| POST | `/api/hosts` | Add a host |
| PATCH | `/api/host/:id` | Edit a host |
| DELETE | `/api/host/:id` | Remove a host |
| POST | `/api/hosts/import` | CSV import |
| POST | `/api/hosts/bulk` | Bulk add (discovery) |
| POST | `/api/discover` | Subnet scan |
| GET/POST | `/api/settings` | Webhook URL config |
| POST | `/api/login` / `/api/logout` | Access gate |

### Features That Could Be Adapted

| Feature | How It Maps to Hawkward |
|---|---|
| **Card-based dashboard** | **SysOps/All** — host/service/process cards with live status, metrics, sparklines |
| **Dark theme with status accent colors** | **UI** — default dark palette; green/red/yellow/gray for status |
| **Detail drawer (slide-in panel)** | **UI** — side or bottom pane for host/service/process details |
| **Real-time WebSocket updates** | **All** — live push via Bubble Tea `tea.Cmd` + `tea.Msg` pattern (already in codebase) |
| **Flap detection state machine** | **All** — debounce alerts; N-fail threshold before status change |
| **Incident/state-change timeline** | **SecOps/SysOps** — per-resource state change history |
| **Subnet discovery** | **NetOps** — CIDR sweep → review → bulk add |
| **Webhook alerts** | **All** — Slack/Discord/PagerDuty integration |
| **Access gate (password guard)** | **Security** — optional auth on TUI launch |
| **Inline sparklines** | **NetOps** — latency, CPU, memory sparklines rendered in terminal |
| **Light/Dark toggle** | **UI** — theme switching (Lip Gloss supports this natively) |
| **Env-var-based configuration** | **SysOps** — env-driven config for poll interval, thresholds, etc. |
| **Host CSV import/export** | **SysOps** — import/export host lists |
| **Sortable table view toggle** | **UI** — switch between card grid and table view |

### Relevance to Hawkward: **Primary — Architecture & Design**

Squib's architecture (poller → state machine → storage → real-time UI → alerts) is nearly a direct blueprint for Hawkward's operations platform. The flap detection, card-based dashboard, detail drawer, and dark theme are all directly adaptable. The key difference is output: Squib renders in a browser (HTML/CSS/WebSocket), while Hawkward renders in the terminal (Bubble Tea/Lip Gloss). This means the **design patterns transfer, the rendering medium changes**.

---

## Synthesis: Actionable Recommendations for Hawkward Ops

### A. Navigation & Information Architecture (source: hackingtool)

**Priority: Critical**

1. **Adopt a category grid as the default launch screen.** Mirror the 20-category 2-column table. For Hawkward:
   - **SysOps** — Processes, Disks, Memory, Users, Logs, Services
   - **NetOps** — Ping, Traceroute, DNS, Packet Capture, Port Scan, Bandwidth, WiFi
   - **SecOps** — Port Audit, Vulnerability Scan, Firewall Rules, TLS Check, Log Monitor
   - **DevOps** — Deployments, Containers, CI Status, Config Audit
   - **AIOps** — Models, Prompts, Agents, Vector Search, Log Analysis

2. **Implement `/` fuzzy search from the main screen.** This is the single most impactful UX improvement. When the user presses `/`, show a search input that filters all tools/commands/views/hosts across all categories in real time. Use Go `strings.Contains` or a simple prefix trie.

3. **Implement `?` context-sensitive help.** Every screen responds to `?` with a brief overlay showing available keybindings for that screen. A footer line can hint: `? help | q quit`.

4. **Implement `t` tag filtering.** Each tool/view gets tags. Pressing `t` opens a tag selector sidebar. Tags for Hawkward:
   - `sysops`: process, disk, memory, user, log, service
   - `netops`: ping, traceroute, dns, packet, scan, wifi, bandwidth
   - `secops`: port, vuln, audit, firewall, tls, monitor
   - `devops`: deploy, container, ci, config, docker
   - `aiops`: model, prompt, agent, vector, chat

5. **Consistent navigation everywhere.** `q` quits the current view (or the app at the root). `Esc` goes back one level. `99` returns to root. Every screen follows the same rule.

6. **Install status / availability indicators.** Show icons next to each tool/view: ✔ available, ✘ not available (dependency missing), ⏳ loading.

### B. Dashboard & Real-Time Monitoring (source: squib)

**Priority: High**

1. **Build a card-based host/service dashboard** as the default landing screen for SysOps and NetOps. Each card renders with Lip Gloss borders and contains:
   - Name/IP (bold title)
   - Status badge (● up / ● down / ● degraded / ● unknown) with color
   - Primary metric (latency ms, CPU%, disk space, packet loss %)
   - Inline sparkline for time-series data
   - Last-updated timestamp

2. **Implement a detail drawer (split-pane).** When a card is selected (Enter), the right portion of the screen slides open to show:
   - Expanded metrics (more sparklines, recent values)
   - State-change / incident timeline (vertical timeline list)
   - Quick-action buttons mapped to keys: `s` SSH, `p` ping, `r` restart, `e` edit
   - The main card grid shifts left or compresses, keeping both visible.

3. **Use Bubble Tea `tea.Cmd` + `tea.Msg` for real-time updates.** Model after squib's WebSocket push. The existing `ResultMsg` pattern in Hawkward already supports this — extend it for metric samples. A background goroutine polls on interval and sends `MetricResultMsg` to the update loop.

4. **Implement flap detection.** Don't fire alerts on a single miss. Track consecutive failures (`FAIL_THRESHOLD`) and recoveries (`OK_THRESHOLD`) per host/service. Configurable via env vars or config file.

5. **Incident/alert timeline.** Record every state transition with a timestamp. Display as a scrollable timeline in the detail drawer. This becomes the alert/incident log.

6. **Webhook alert channel.** When a host crosses the flap threshold, fire a webhook. Compatible with Slack, Discord, Teams, PagerDuty. Configurable per-host or globally.

### C. Packet Capture & Network Tooling (source: netscanner)

**Priority: High**

1. **Build the NetOps packet capture view** with:
   - Start/pause/resume controls (mapped to keys: `s` start, `p` pause, `r` resume)
   - Protocol color-coding: TCP (cyan), UDP (yellow), ICMP (green), ARP (magenta), Other (gray)
   - Filter bar activated by `/` (type `tcp`, `udp`, `port 80`, `host 10.0.0.1`)
   - Live traffic counter in the footer (packets/sec, bytes/sec)
   - Scrolling log of captured packets with columns: Time, Protocol, Source → Dest, Length, Info

2. **Build the NetOps port scanner** with:
   - Target input (IP or hostname)
   - Port range input (or common ports preset)
   - Progress bar showing scan completion percentage
   - Results table: Port, State (open/closed/filtered), Service
   - Export to CSV/file

3. **Build the NetOps CIDR ping sweep** with:
   - CIDR input (e.g., `10.0.0.0/24`)
   - Progress bar showing completion
   - Results table: IP, Hostname, MAC, OUI, Status (alive/dead)
   - Bulk-add alive hosts to the dashboard (squib flow)

4. **Build WiFi scanning** with signal-strength bar indicators (inline terminal block bars).

### D. Theme & Visual Design (source: squib + Lip Gloss)

**Priority: Medium**

1. **Default to dark theme.** Use a dark background (`#1a1b2e` or similar), light text, and these status colors:
   - Green (`#00ff88`): up/healthy/online/success
   - Red (`#ff3355`): down/critical/offline/error
   - Yellow (`#ffaa00`): degraded/warning
   - Cyan (`#00ccff`): informational/running/in-progress
   - Gray (`#555555`): idle/unreachable/unknown
   - Magenta (`#ff00ff`): AIOps/agent activity (accent)

2. **Light/dark theme toggle.** Store preference in config file. Lip Gloss `lipgloss.AdaptiveColor` supports this natively.

3. **Card-based layout for all list views.** Hosts, services, processes, tools — all rendered as bordered cards with:
   - Title bar (name, icon, status dot)
   - Body (metrics, key-value pairs)
   - Footer (actions, timestamp)
   - Use Lip Gloss `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())`

4. **Footer status bar** on every screen showing:
   - Current screen/module name
   - Keybinding hints
   - Polling status (if applicable)
   - Packet/traffic counters (if applicable)

5. **Sparkline charts.** Render time-series data using block characters (`▁▂▃▄▅▆▇█` or half-blocks). Simple `<5 lines of Go` implementation. Apply to latency, CPU, memory, bandwidth.

### E. Search & Discovery System (source: hackingtool)

**Priority: High**

1. **`/` search** — input appears at the top/bottom on `/` keypress. Filters the current list in real time as the user types. Match against tool name, description, tags. Display results with category path (e.g., `NetOps > Ping Sweep`).

2. **Tag filtering** — `t` key opens a tag selector. Tags are displayed as a horizontal row of selectable chips. Multi-select supported. Filtered view updates in real time.

3. **Recommend mode** — `r` key opens a text input. User types a natural language request: "scan network for open ports" or "show me all slow hosts." The system matches keywords against tool/view names, tags, and descriptions, then presents relevant options. This is a simple keyword matcher in v1, could evolve to use AI/AIOps in the future.

### F. Architecture Patterns (all three sources)

**Priority: Medium**

| Pattern | Source | How to Apply |
|---|---|---|
| Real-time poller + state machine | squib | Background `tea.Cmd` goroutine polls hosts/services, sends `MetricResultMsg` to update loop |
| Flap detection state machine | squib | `FAIL_THRESHOLD` / `OK_THRESHOLD` per host. Track consecutive failures in a struct field |
| Category + tool hierarchy | hackingtool | Map to `internal/{sysops,netops,secops,devops,aiops}` package structure. Each package registers its tools in a central registry |
| Root-required guard | netscanner | `os.Geteuid()` check at startup; show warning banner if not root; graceful degradation (some tools work without root) |
| CSV/JSON export | netscanner, squib | Export scan results, host lists, logs, metrics. Implement as a utility in `internal/common/export.go` |
| Env-var + config file | squib | Precedence: CLI flags > env vars > config file > defaults. All thresholds, intervals, ports configurable |
| Detail drawer pattern | squib | Split-pane or overlay pane that opens on card selection. Implementation: store focused item ID in model, render detail view in right panel |
| Access gate | squib | Optional password prompt on startup. If `HAWKWARD_SECRET` is set, show a password input before allowing access |
| OS-aware features | hackingtool | Check platform at startup; disable Linux-only features on Windows/macOS |
| Docker support | hackingtool, squib | Dockerfile + compose.yml for containerized deployment with `CAP_NET_RAW` for ICMP |

### G. Priority Order for Implementation

```
Phase 1 — Foundation (Critical)
├── Navigation system (category grid + consistent back/quit)
├── `/` fuzzy search
├── `?` context help
└── Dark theme with status colors

Phase 2 — Core Monitoring (High)
├── Card-based host/service dashboard
├── Detail drawer (split-pane)
├── Background poller with real-time updates (tea.Cmd + tea.Msg)
├── Flap detection state machine
└── Incident/state-change timeline

Phase 3 — NetOps Tools (High)
├── Packet capture view (start/pause/resume, protocol color-coding, filter)
├── Port scanner with progress bar
├── CIDR ping sweep
└── Export results to CSV/JSON

Phase 4 — Search & Discovery (High)
├── Tag filtering (`t` key + chip selector)
├── Recommend mode (`r` key + keyword matching)
└── Tool registry with status indicators

Phase 5 — Alerts & Integrations (Medium)
├── Webhook alert channel (Slack/Discord/PagerDuty)
├── Sparklinecharts for time-series data (latency, CPU, memory)
├── Light/Dark theme toggle
└── CSV/JSON import/export utilities

Phase 6 — Polish & Security (Medium)
├── Access gate (optional password on launch)
├── OS-aware feature filtering
├── Docker support
└── Config file + env var system
```

---

## Quick Reference: Key Files & Patterns in Existing Codebase

| File | Purpose | How It Maps to Research |
|---|---|---|
| `internal/common/types.go` | `Screen` enum and shared types | Add screens for: `ScreenCategoryGrid`, `ScreenDashboard`, `ScreenSearch`, `ScreenPacketCapture`, `ScreenPortScan`, `ScreenDetailDrawer` |
| `internal/common/styles.go` | Shared Lip Gloss styles | Add: `StatusGreen`, `StatusRed`, `StatusYellow`, `StatusCyan`, `StatusGray`, `CardStyle`, `DetailDrawerStyle`, `FooterStyle`, `SearchInputStyle` |
| `update.go` | `handleKeyPress` keyboard dispatch | Add: `/` search trigger, `t` tag filter, `r` recommend, `?` help, `99`/`Esc` back, consistent `q` quit |
| `internal/common/` | Shared utilities | Add: `export.go` (CSV/JSON), `sparkline.go` (inline charts), `config.go` (env + config file loading) |
| `cmd/hawkward/` | Entry point | Add: root check (`os.Geteuid`), optional access gate, config initialization |
| `internal/ui/` | TUI components | Add: `dashboard.go` (card grid), `search.go` (search input + filtered list), `detail.go` (drawer panel), `category_grid.go` (launch screen), `packet_capture.go`, `portscan.go`, `ping_sweep.go` |
| `internal/{sysops,netops,secops,devops,aiops}/` | Domain modules | Each package registers its tools/views in a central registry. Add `Register()` function per package |

---

## References

| Repo | URL | Primary Insight |
|---|---|---|
| netscanner | https://github.com/Chleba/netscanner | Packet dump UI, port scan, CIDR sweep, inline bar charts |
| hackingtool | https://github.com/Z4nzu/hackingtool | `/` search, tag filter, category menu, `?` help, consistent nav |
| squib | https://github.com/teknetai/squib | Card dashboard, dark theme, detail drawer, flap detection, real-time polling, webhook alerts, access gate |
| Bubble Tea v2 | `charm.land/bubbletea/v2` | TUI framework for Hawkward |
| Lip Gloss v2 | `charm.land/lipgloss/v2` | Styling and theming for Hawkward |
