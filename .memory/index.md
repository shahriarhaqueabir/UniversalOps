# Hawkward — Workspace Memory

## Active Sprint
- [[plans/2026-07-08-sprint-7-gui-finalization]] — **Sprint 7: GUI Finalization & Production Readiness** (ACTIVE PLAN)

## Current Session Context
- **Status**: Full project audit complete, research gathered from netscanner, hackingtool, squib
- **Goal**: Transform TUI-based Bubble Tea app into a Wails v2 native GUI with squib-inspired design
- **Architecture**: Go backend (Wails bindings) + React/TypeScript/Vite frontend (Tailwind v4 + Recharts)
- **Key Insight**: Frontend has mock data stubs — all 9 pages, components, hooks exist but use mockData.ts not real Wails runtime calls
- **Design Target**: Squib-inspired dark theme with purple accent (`#7c6cff`), Inter + JetBrains Mono typography, 252px sidebar, KPI cards with sparklines, glassmorphism topbar, status dots with glow

## Topics
- [[plans/gui-overhaul-v2]] — Comprehensive overhaul plan: design system, architecture, sprints
- [[hawkward-card-system]] — Card types, layout engine, states, interaction
- [[hawkward-command-palette]] — Search algorithm, key bindings, operation registry
- [[hawkward-architecture]] — Go + Wails v2 architecture, package layout
- [[hawkward-charts]] — Chart library design (Recharts in GUI)
- [[hawkward-cross-platform]] — Windows/Linux notes

## Known Issues
- Wails bindings need real frontend calls (currently mock data)
- No `wailsjs/go/` bindings generated (uses `-skipbindings`)
- Light theme not fully implemented
- No frontend tests
- No network topology persistence
- Log viewer lacks virtual scrolling
- Missing error boundaries and offline states
