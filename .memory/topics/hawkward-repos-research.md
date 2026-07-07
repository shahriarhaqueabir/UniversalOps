# Repo Research — Design & Feature Inspirations

## netscanner (Chleba/netscanner)
**Tech**: Rust + ratatui | **Stars**: 1.8k | **License**: MIT

### Key Features to Adapt
| Feature | Hawkward Priority | Notes |
|---------|------------------|-------|
| ASCII art header banner | High | Rust-based TUI network scanner; ratatui equivalent of Bubble Tea |
| Real-time packet dump (TCP/UDP/ICMP) with live filtering | Medium | Uses libpnet for raw sockets — Windows needs Npcap |
| WiFi network scanning + signal strength charts | Medium | Shows SSID, signal dBm, channel, encryption type |
| CIDR ping sweep with hostname/MAC/OUI resolution | High | Bulk ping with vendor lookup |
| Port scanning with real-time progress | High | Progress bar during scan |
| Traffic counting + DNS records | Medium | Per-connection bandwidth tracking |
| Export scan results to CSV | High | CSV export for all scan types |

### UI Patterns
- Banner at top of screen (ASCII art)
- Sidebar for interface selection
- Tabbed main content area
- Real-time updating charts (signal strength)
- Color-coded status indicators

---

## hackingtool (Z4nzu/hackingtool)
**Tech**: Python | **Stars**: 16k+ | **License**: MIT

### Key Features to Adapt
| Feature | Hawkward Priority | Notes |
|---------|------------------|-------|
| `/` search across all tools | **Critical** — Phase 2.1 | Search by name, description, keyword |
| `t` tag filtering | **Critical** — Phase 2.1 | Filter by tag (osint, web, network, etc.) |
| `r` recommend mode | High | Natural language → matching operations |
| Install status ✔/✘ | High — Phase 2.3 | Check tool availability at startup |
| Batch install per category | Medium | "Install all" operation per ops layer |
| Smart update detection | Medium | git pull / go install detection |
| OS-aware menus | High | Hide Linux-only ops on macOS/Windows |

### UI Patterns
- Clean category cards with emoji icons
- Tool count badges on categories
- Search bar overlay
- Help card (`?`) accessible from anywhere
- Consistent 2-column layout

---

## squib (teknetai/squib) — PRIMARY DESIGN INSPIRATION
**Tech**: Node/TS + SQLite | **License**: MIT

### Core Design Language
The squib aesthetic is what we're targeting: **clean, card-based, information-dense but readable, with a dark/light toggle and real-time data visualization.**

### Key Design Elements to Adopt
| Element | Implementation in Hawkward |
|---------|--------------------------|
| Card-based dashboard | Replace tab-based views with card grids |
| Status cards (per-host → per-operation) | Each operation becomes a card with status indicator |
| RTT/latency charts | Already have sparklines; enhance to squib-level clarity |
| Incidents/alerting | Anomaly detection already exists; add incident log |
| Clean dark/light themes | Extend from 4 to 8+ themes (squib-inspired) |
| Minimal, "nice to look at" aesthetic | Better padding, consistent borders, color harmony |
| Real-time updates | Already have TickMsg; enhance streaming |
| Subnet discovery | CIDR scan with ICMP sweep |
| Per-host detail drawer | Squib's host detail panel concept |

### Squib Theme Palette Reference
```
Dark mode:
  Background: #0f172a (slate-900)
  Card bg:    #1e293b (slate-800)
  Primary:    #38bdf8 (sky-400)
  Success:    #4ade80 (green-400)
  Warning:    #fbbf24 (amber-400)
  Danger:     #f87171 (red-400)
  Text:       #f8fafc (slate-50)
  Muted:      #94a3b8 (slate-400)
  Border:     #334155 (slate-700)

Light mode:
  Background: #f8fafc (slate-50)
  Card bg:    #ffffff (white)
  Primary:    #0ea5e9 (sky-500)
  Success:    #22c55e (green-500)
  Warning:    #eab308 (yellow-500)
  Danger:     #ef4444 (red-500)
  Text:       #0f172a (slate-900)
  Muted:      #64748b (slate-500)
  Border:     #e2e8f0 (slate-200)
```

### Additional Squib-Inspired Features
- **Dashboard landing page** with live health gauges for all 5 layers (Phase 1.1)
- **Per-operation cards** instead of text tables (Phase 1.2)
- **Flap-detection** for alerts (don't fire on single blips)
- **Webhook-style alerting** via notification system
- **Clean help overlay** inspired by squib's simplicity

---

## Synthesis: The Squib-First Design Direction

### Visual Hierarchy
```
1. ASCII Banner (netscanner) → App identity
2. Dashboard Cards (squib) → Live health gauges + operation cards
3. Layer Views (squib cards) → Card-grid data presentation
4. Status Bar (existing ++) → Enhanced with anomaly count + theme indicator
5. Command Palette (hackingtool) → Global search with `/`
```

### Theme Expansion Plan
Base 4 themes (existing) → 8+ themes:
1. `default` — Purple/green dark (current)
2. `dark` — Purple/green dark variant (current)
3. `light` — Light mode (current)
4. `high-contrast` — Accessibility (current)
5. `squib-dark` — Slate/sky dark (inspired by squib)
6. `squib-light` — Slate/sky light (inspired by squib)
7. `amber` — Amber-on-black retro terminal
8. `green` — Green-on-black classic terminal
9. `dracula` — Dracula-inspired (user requested)
10. `nord` — Nord-inspired (user requested)
