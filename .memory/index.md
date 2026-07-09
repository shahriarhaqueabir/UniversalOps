# Hawkward — Workspace Memory

## Active Session
- **Sprint 19: Production Stabilization** — ✅ COMPLETE

## Completed Tickets

| ID | Status | Description |
|----|--------|-------------|
| T-01 | ✅ DONE | Fix nil error wrap in GetUsers + non-English locale fallback for netsh |
| T-02 | ✅ DONE | Fix PowerShell JSON parsing (GetServices + GetSecurityEvents) |
| T-03 | ✅ DONE | Improve GetDefenderStatus fallback handling |
| T-04 | ✅ DONE | Fix ListDirectory path handling |
| T-05 | ✅ DONE | Fix RunPowerShell profile path for dev mode |
| T-06 | ✅ DONE | Verify NetOps tabs work end-to-end |
| T-07 | ✅ DONE | Audit every button/tab across all sections |
| T-08 | ✅ DONE | Verify release pipeline produces downloadable exe |
| T-09 | ✅ DONE | Write no-programmer launch guide in README |
| T-10 | ✅ DONE | Build, tag & push v1.2.0 |
| T-11 | ✅ DONE | UI/UX Review & CSS Theming Fixes — version strings, CSS vars, skeleton loader, animations |
| T-12 | ✅ DONE | Library Research & Open-Source Report |
| T-13 | ✅ DONE | Phase 2 Bugfixes — Dashboard alert state (useRef), Logs virtual scroll, empty states, ARIA labels |
| T-14 | ✅ DONE | Fresh UI/UX Audit — hardcoded colors, error handling, empty dir cleanup, new tests |

## Changes Made (2026-07-10 — Fresh UI/UX Audit)

### CSS Variable Fixes
1. **TopBar.tsx**: Replaced hardcoded `rgba(8,10,15,0.6)` → `var(--color-bg)` to fix light theme
2. **DevOps.tsx** (6x): Replaced `bg-[#0b1120]` → `bg-[var(--color-bg)]` on Terminal input, Terminal output, PowerShellPro output, Services table, FileBrowser table, FileBrowser preview
3. **AIOps.tsx** (3x): Replaced `bg-[#0b1120]` → `bg-[var(--color-bg)]` on ChatTab container, Reports output, Anomalies table container

### Error Handling
4. **Settings.tsx**: Added `.catch()` to AppInfo fetch call to prevent silent failure when backend is unavailable

### Empty Directory Cleanup
5. Removed empty `components/charts/`, `components/dashboard/`, `components/logs/`

### New Frontend Tests
6. **Sidebar.test.tsx**: 5 tests — renders all items, highlights current page, calls onNavigate, shows version, collapse/expand
7. **ConfirmDialog.test.tsx**: 5 tests — closed state, open state, onConfirm, onClose, danger type
8. **ErrorBoundary.test.tsx**: 2 tests — renders children, catches errors with fallback UI

### TypeScript Fixes
9. **NetOps.tsx**: Fixed implicit `any` types on 2 tickFormatter callbacks → `(v: number)`
10. **Settings.tsx**: Fixed implicit `any` types on 2 Radix Slider `onValueChange` callbacks → `([v]: number[])`

## Test Results
- **28 frontend tests pass** across 7 test files
- Test files: utils.test.tsx, ErrorBoundary.test.tsx, ConfirmDialog.test.tsx, Sidebar.test.tsx, Dashboard.test.tsx, Logs.test.tsx, Settings.test.tsx
- Missing test coverage: TopBar, NetOps, SecOps, DevOps, AIOps, NetworkDesign pages

## Library Presence Check (Verified)

All recommended libraries from T-12 library research are **NOT currently installed**:

| Library | Priority | Present? | Notes |
|---------|----------|----------|-------|
| `tailwindcss-animate` | P0 | ❌ | Manual CSS keyframes in globals.css (50+ lines) |
| `@tanstack/react-virtual` | P0 | ❌ | Custom virtual scroll in Logs.tsx |
| `zustand` | P1 | ❌ | Raw useState/useEffect patterns |
| `sonner` | P1 | ❌ | No toast notification system |
| `@tanstack/react-query` | P1 | ❌ | Manual useEffect + setInterval polling |
| `date-fns` | P2 | ❌ | Ad-hoc toLocaleTimeString() calls |
| `motion` | P2 | ❌ | Only CSS transitions |
| `rs/zerolog` | P3 | ❌ | Custom common.LogInfo/LogWarn logging |
| `prometheus/client_golang` | P4 | ❌ | No metrics export |
| `gopacket` | P4 | ❌ | Uses golang.org/x/net only |
| Ollama Go SDK | P3 | ❌ | Raw http.Post REST calls |
| `nanoid` | P3 | ❌ | Sequential genId() in NetworkDesign |

## UI/UX Issues Identified & Fixed (2026-07-10)

| Issue | Severity | Status |
|-------|----------|--------|
| TopBar hardcoded bg breaks light theme | High | ✅ Fixed |
| DevOps.tsx 6x hardcoded dark colors | Medium | ✅ Fixed |
| AIOps.tsx 3x hardcoded dark colors | Medium | ✅ Fixed |
| Settings AppInfo fetch no error handling | Medium | ✅ Fixed |
| NetOps.tsx implicit any types | Medium | ✅ Fixed |
| Empty component directories | Low | ✅ Fixed |
| Missing frontend test coverage | Medium | Partially Fixed |

## Known Issues (Remaining)
- 7 dead components removed (HealthCard, AlertBadge, AreaChart, Gauge, MiniSparkline, ExportDialog, SettingsDialog)
- SQLite DB at `hawkward.db` in app working directory
- Build: `wails build` (NOT `go build`)
- No toast notification system — event alerts only visible on Dashboard
- No page transition animations
- Some inline styles remain in chart components (Tooltip contentStyle in NetOps)
- All "Cannot find module" diagnostics are stale LSP issues — packages exist on disk
- NetworkDesign topology canvas uses hardcoded initial devices (BY DESIGN)
