# UniversalOps Roadmap

> **Status**: Active Development — v1.6.0 (In Progress)
> **Latest Release**: v1.5.0 (2026-07-31, Windows, hash-chain verified)
>
> This roadmap is the single source of truth for what we build and why. Priorities
> are set by the project manager based on the product vision: **100% local,
> mechanics-first, zero telemetry** operations intelligence.

---

## 🎯 Product Vision

UniversalOps is a high-performance native desktop operations platform (Go/Wails v2 +
React 19). It turns a machine's raw telemetry into **actionable operations
intelligence** — answering three questions in under 10 seconds on every screen:
*What is the current state? What changed recently? What needs my attention?*

**Strategic insight (from codebase audit):** *Most backend infrastructure is built.
The gap is the frontend observation layer.* The backend already collects, stores,
and analyzes rich telemetry. The highest-value work is surfacing that data into
actionable, configurable, exportable views.

---

## ✅ Recently Completed

- **CI/CD Pipeline**: Full test matrix (Linux/Windows/macOS), release automation, Codecov integration
- **Security Hardening**: `SECURITY.md`, panic recovery across all facade methods, input validation in MCP server
- **Storage Fixes**: Thread-safe `Close()` with `sync.Once`, anomaly detection with proper standard deviation
- **Documentation**: Architecture guide, data flow diagrams, troubleshooting guide, onboarding docs

---

## 🗓️ v1.6.0 — Observation Layer (CURRENT SPRINT)

The theme: **turn collected telemetry into configurable, exportable, alertable intelligence.**

### P0 — Alert Rules Engine (highest value)
- [ ] **ALR-01** User-configurable threshold-based alerts (CPU/mem/disk/network) with
      desktop notifications. Backend: alert rule CRUD + evaluation loop; frontend:
      rule editor + notification wiring. *Builds on existing Alert engine in `internal/common/`.*
- [ ] **ALR-02** Alert history + acknowledgement workflow (persist, dismiss, snooze).
- [ ] **UPD-01** In-app update checker with one-click upgrade (GitHub Releases API, local-first).

### P1 — Export & Reporting
- [ ] **EXP-01** PDF/CSV export for reports and metrics (ReportsCenter + per-discipline overviews).
- [ ] **EXP-02** Scheduled report generation (daily digest) with local file output.

### P1 — Custom Dashboard Layouts
- [ ] **DASH-01** Drag-and-drop widget arrangement (persisted via existing ConfigSchema + AJV).
- [ ] **DASH-02** Widget presets + per-widget refresh (builds on `useWidgetRefresh`).

### P2 — Performance Profiling
- [ ] **PERF-01** pprof integration for long-running session analysis.

---

## 🗓️ v1.7.0 — AI & Intelligence

- [ ] **AI-01** Multi-model AI support (OpenAI/Anthropic alongside Ollama) — *local-first caveat: keep optional, off by default.*
- [ ] **AI-02** Predictive analytics — trend forecasting for disk/memory/CPU.
- [ ] **AI-03** Natural language query — ask about system state in plain English.
- [ ] **AI-04** Automated RCA reports — scheduled root-cause digests.

### Network
- [ ] **NET-01** Packet capture viewer (pcap analysis + protocol dissection).
- [ ] **NET-02** Per-process bandwidth monitoring.
- [ ] **NET-03** VPN detection & audit.

### Security
- [ ] **SEC-01** File integrity monitoring (hash-based change detection).
- [ ] **SEC-02** Local CVE matching for installed software.
- [ ] **SEC-03** CIS benchmark compliance reporting.

---

## 🗓️ v2.0 — Platform & Ecosystem

### Platform
- [ ] **PLT-01** Plugin system — community collectors/analyzers.
- [ ] **PLT-02** Remote agent — lightweight collector for headless servers.
- [ ] **PLT-03** Mobile companion — read-only dashboard.
- [ ] **PLT-04** Web UI mode — optional browser access.

### Ecosystem
- [ ] **ECO-01** Community marketplace — share dashboards/alerts/workflows.
- [ ] **ECO-02** REST API gateway for third-party integration.
- [ ] **ECO-03** Team collaboration — shared workspaces with RBAC.

---

## 🧭 Guiding Principles

1. **100% local-first** — no data leaves the machine; cloud AI is always opt-in.
2. **Mechanics-first** — high-density technical telemetry over marketing fluff.
3. **Zero telemetry** — no cloud dependencies or tracking in the product.
4. **Observation over collection** — the backend already collects; surface it.
5. **Tests required** — every feature ships with tests; full suite must pass.

---

## 💡 Feature Requests

Have an idea? [Open an issue](https://github.com/shahriarhaqueabir/UniversalOps/issues/new?template=feature_request.md) with the `enhancement` label.

---

*Roadmap updated: 2026-08-02*