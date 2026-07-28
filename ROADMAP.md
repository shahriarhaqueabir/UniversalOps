# UniversalOps Roadmap

> **Status**: Active Development — v1.3.1 (Unreleased)
>
> This roadmap outlines planned features and improvements. Timelines are approximate and subject to change.

---

## ✅ Recently Completed

- **CI/CD Pipeline**: Full test matrix (Linux/Windows/macOS), release automation, Codecov integration
- **Security Hardening**: `SECURITY.md`, panic recovery across all facade methods, input validation in MCP server
- **Storage Fixes**: Thread-safe `Close()` with `sync.Once`, anomaly detection with proper standard deviation
- **Documentation**: Architecture guide, data flow diagrams, troubleshooting guide, onboarding docs

---

## 🗓️ Short-Term (v1.4.0)

### Platform & Stability
- [ ] **macOS code signing automation** — full notarization pipeline in CI
- [ ] **Linux AppImage packaging** — distribution via AppImage for broader Linux support
- [ ] **Auto-update mechanism** — in-app update checker with one-click upgrade
- [ ] **Performance profiling** — pprof integration for long-running session analysis

### Features
- [ ] **Custom dashboard layouts** — drag-and-drop widget arrangement
- [ ] **Alert rules engine** — user-configurable threshold-based alerts with desktop notifications
- [ ] **Export & reporting** — PDF/CSV export for reports and metrics
- [ ] **Multi-workspace support** — switch between multiple system profiles

---

## 🗓️ Medium-Term (v1.5.0)

### AI & Intelligence
- [ ] **Multi-model AI support** — OpenAI/Anthropic API integration alongside Ollama
- [ ] **Predictive analytics** — trend forecasting for disk, memory, and CPU utilization
- [ ] **Natural language query** — ask questions about system state in plain English
- [ ] **Automated RCA reports** — scheduled root cause analysis digests

### Network
- [ ] **Packet capture viewer** — integrated pcap analysis with protocol dissection
- [ ] **Bandwidth monitoring** — per-process network usage tracking
- [ ] **VPN detection & audit** — tunnel interface identification and routing table analysis

### Security
- [ ] **File integrity monitoring** — hash-based change detection on critical system files
- [ ] **Vulnerability scanning** — local CVE matching for installed software
- [ ] **Compliance reporting** — CIS benchmark checks with pass/fail scoring

---

## 🗓️ Long-Term (v2.0)

### Platform
- [ ] **Plugin system** — community-contributed collectors and analyzers
- [ ] **Remote agent** — lightweight collector for monitoring headless servers
- [ ] **Mobile companion** — read-only dashboard for iOS/Android
- [ ] **Web UI mode** — optional web-based interface for browser access

### Ecosystem
- [ ] **Community marketplace** — share custom dashboards, alerts, and workflows
- [ ] **API gateway** — REST API for third-party integration
- [ ] **Team collaboration** — shared workspaces with role-based access

---

## 💡 Feature Requests

Have an idea? [Open an issue](https://github.com/shahriarhaqueabir/UniversalOps/issues/new?template=feature_request.md) with the `enhancement` label.

---

*Roadmap updated: 2026-07-28*