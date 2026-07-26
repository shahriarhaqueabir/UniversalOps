# UniversalOps — Documentation Hub

> A high-performance native desktop operations platform built with Go (Wails v2) and React 19/TypeScript.

---

## 📋 Quick Navigation

| Section | Description |
|---------|-------------|
| [Getting Started](#rocket-getting-started) | Install, build, and first-run guide |
| [Architecture](#building-architecture--internals) | How UniversalOps works under the hood |
| [Operations Guide](#wrench-operations-guide) | Using each ops module |
| [Development](#computer-development) | Contributing, coding standards, tools |
| [Reference](#book-reference) | Tools, data flow, troubleshooting |

---

## 🚀 Getting Started

| Doc | What it covers |
|-----|----------------|
| [INSTALL.md](INSTALL.md) | One-click install, prerequisites, building from source, optional extras (Ollama, LibreHardwareMonitor) |
| [USER_GUIDE.md](USER_GUIDE.md) | Full walkthrough — installation, AI setup, feature deep-dive, customization, troubleshooting |
| [ONBOARDING.md](ONBOARDING.md) | Developer onboarding — tech stack, architecture map, directory layout, conventions, common tasks |

---

## 🏗️ Architecture & Internals

| Doc | What it covers |
|-----|----------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | High-level architecture — engine loop, capability registry, module design, security sandbox |
| [collector-architecture.md](collector-architecture.md) | Collector system design — goroutine-per-collector, sharded data stores, exponential backoff, ring buffer |
| [DATA_FLOW.md](DATA_FLOW.md) | Telemetry lifecycle and command flow — sequence diagrams for inbound pipeline and outbound commands |
| [STANDARDS.md](STANDARDS.md) | Development standards — Go conventions, naming patterns, error handling, frontend component guidelines |

---

## 🔧 Operations Guide

| Doc | What it covers |
|-----|----------------|
| [MustHaves.md](MustHaves.md) | Product requirements — dashboard health score, resource summaries, module overview pages |
| [ToolsCommands.md](ToolsCommands.md) | Tools & commands reference — data source hierarchy, system information tables, network/security tool mappings |

---

## 💻 Development

| Doc | What it covers |
|-----|----------------|
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contribution process, testing standards, commit conventions, project structure, coding standards |
| [developing.md](developing.md) | Deep development guide — git strategy, environment setup, debugging, PR workflow, project tree |
| [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) | Community standards and enforcement guidelines |

---

## 📖 Reference

| Doc | What it covers |
|-----|----------------|
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Common issues — GPU detection, WMI/permissions, network tools, database migration, process access |
| [DATA_FLOW.md](DATA_FLOW.md) | Sequence diagrams for telemetry pipeline and command execution |
| [SEO_STRATEGY.md](SEO_STRATEGY.md) | SEO and discoverability strategy for the project |

---

## 📁 Complete File Tree

```
docs/
├── readme.md              ← You are here
├── ARCHITECTURE.md        — System architecture overview
├── collector-architecture.md  — Collector design deep-dive
├── DATA_FLOW.md           — Telemetry & command flow diagrams
├── INSTALL.md             — Installation & quick start
├── MustHaves.md           — Product requirements & UI baseline
├── ONBOARDING.md          — Developer onboarding
├── SEO_STRATEGY.md        — SEO & discoverability
├── STANDARDS.md           — Coding standards & conventions
├── ToolsCommands.md       — Tools & commands reference
├── TROUBLESHOOTING.md     — Common issues & solutions
├── USER_GUIDE.md          — Full user guide
├── developing.md          — Deep development guide
├── screenshots/           — Application screenshots
│   ├── final_execution_1_dashboard.png
│   ├── final_execution_2_reports.png
│   ├── final_execution_3_sysops.png
│   ├── final_execution_4_netops.png
│   ├── final_execution_5_secops.png
│   ├── final_execution_6_devops.png
│   ├── final_execution_7_aiops.png
│   ├── final_execution_8_logs.png
│   └── final_execution_9_workflowlibrary.png
├── archives/              — Historical archives
├── audit/                 — Security & wiring audit reports
├── plans/                 — Implementation plans & roadmaps
└── superpowers/           — Agent skill specifications & blueprints
```

---

> **Tip**: Use your editor's file search (Ctrl+P / Cmd+P) to quickly jump to any document by name.
