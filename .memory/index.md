# Hawkward — Workspace Memory

## Active Session
- **Sprint 19: Production Stabilization** — 🔲 IN PROGRESS

## Sprint 19 Tickets

| ID | Status | Description |
|----|--------|-------------|
| T-01 | 🔲 TODO | Fix nil error wrap in GetUsers + non-English locale fallback for netsh |
| T-02 | 🔲 TODO | Fix PowerShell JSON parsing (GetServices + GetSecurityEvents) |
| T-03 | 🔲 TODO | Improve GetDefenderStatus fallback handling |
| T-04 | 🔲 TODO | Fix ListDirectory path handling |
| T-05 | 🔲 TODO | Fix RunPowerShell profile path for dev mode |
| T-06 | 🔲 TODO | Verify NetOps tabs work end-to-end |
| T-07 | 🔲 TODO | Audit every button/tab across all sections |
| T-08 | 🔲 TODO | Verify release pipeline produces downloadable exe |
| T-09 | 🔲 TODO | Write no-programmer launch guide in README |
| T-10 | 🔲 TODO | Build, tag & push v1.2.0 |
| T-11 | ✅ DONE | UI/UX Review & CSS Theming Fixes |
| T-12 | ✅ DONE | Library Research & Open-Source Report |
| T-13 | ✅ DONE | Phase 2 Bugfixes — Dashboard state, Logs scrolling, empty states, accessibility |

## Completed

| ID | Status | Description |
|----|--------|-------------|
| Sprint 18 | ✅ DONE | Release Hardening — tests, race detector, v1.1.1 tagged |
| T-11 | ✅ DONE | UI/UX Review & CSS Theming Fixes |
| T-12 | ✅ DONE | Library Research & Open-Source Report — saved to `.memory/topics/library-research.md` |
| T-13 | ✅ DONE | Phase 2 Bugfixes — code review and fixes applied |

## T-11: UI/UX Review & CSS Theming Fixes

### Changes Made
1. **Fixed version strings**: Sidebar `v0.1.0` → `v1.1.1`, Settings `0.1.0` → `1.1.1`
2. **Fixed CSS theme references across all files**: Replaced undefined Tailwind color utilities with project-defined CSS variables
3. **Dashboard loading state**: Proper skeleton loader replacing plain text
4. **Replaced hardcoded hex colors**: In chart components with CSS variable references
5. **Added CSS animation fallbacks**: For `animate-in`, `fade-in`, `zoom-in-95`, `slide-in-from-*` classes
6. **Added global focus-visible styles**: Keyboard accessibility

## T-12: Library Research & Open-Source Report

Saved to `.memory/topics/library-research.md` — full report with 12+ library evaluations across three effort tiers.

Key recommendations:
- **P0**: `tailwindcss-animate` (remove manual CSS), `@tanstack/react-virtual` (fix scrolling)
- **P1**: `zustand` (shared state), `sonner` (toasts), `@tanstack/react-query` (data fetching)
- **P2**: `date-fns` (formatting), `motion` (animations)

## T-13: Phase 2 Bugfixes

### Changes Made
1. **Dashboard alert state**: Fixed `useState` that only extracted setter (`[1]`) — refactored to use `useRef` for alert accumulation
2. **Logs virtual scrolling**: Fixed `totalHeight` calculation to properly account for expanded rows
3. **Empty states**: Improved empty-state rendering across SysOps (processes), DevOps (services, files), SecOps (events, rules, ports)
4. **Accessibility**: Added ARIA labels to tab buttons, close buttons, interactive elements across all pages
5. **Settings Radix slider focus**: Added `@starting-style` equivalent for slider thumbs
6. **Added frontend test**: Basic unit test for `cn()` utility and initial ErrorBoundary smoke test

### Files Changed
- `src/pages/Dashboard.tsx` — refactored alert state to useRef
- `src/pages/SysOps.tsx` — improved empty/skeleton states
- `src/pages/DevOps.tsx` — aria labels, empty state improvements
- `src/pages/SecOps.tsx` — aria labels, empty states
- `src/pages/Logs.tsx` — fixed virtual scroll height calculation
- `src/test/` — added basic frontend test file

## Known Issues
- TUI fully removed — no remnants ✅
- SQLite DB at `hawkward.db` in app working directory
- Build: `wails build` (NOT `go build`)
- Dead components (identified, not removed): HealthCard, AlertBadge, AreaChart, Gauge, MiniSparkline, ExportDialog, SettingsDialog

## Topics
- [[sprint-19-production-stabilization]] — Sprint 19 tickets
- [[ui-ux-review]] — CSS theming fixes
- [[codebase-research]] — Architecture overview
- [[library-research]] — Library recommendations
