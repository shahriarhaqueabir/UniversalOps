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

## Completed

| ID | Status | Description |
|----|--------|-------------|
| Sprint 18 | ✅ DONE | Release Hardening — tests, race detector, v1.1.1 tagged |
| T-11 | ✅ DONE | UI/UX Review & CSS Theming Fixes |

## T-11: UI/UX Review & CSS Theming Fixes

### Changes Made
1. **Fixed version strings**: Sidebar `v0.1.0` → `v1.1.1`, Settings `0.1.0` → `1.1.1`
2. **Fixed CSS theme references across all files**: Replaced undefined Tailwind color utilities (`bg-card`, `text-primary`, `text-muted`, `bg-primary`, `text-primary-foreground`, `bg-background`, `border-primary`, `placeholder-muted`) with project-defined CSS variables
3. **Dashboard loading state**: Replaced plain text "Synching Neural Bridge..." with proper skeleton loader
4. **Replaced hardcoded hex colors**: In AreaChart, Gauge, ConnectionLine with CSS variable references
5. **Added CSS animation fallbacks**: For `animate-in`, `fade-in`, `zoom-in-95`, `slide-in-from-*` classes (since `tailwindcss-animate` not in dependencies)
6. **Added global focus-visible styles**: Keyboard accessibility with accent-colored focus rings
7. **Fixed color mappings**: In DevOps StatusBadge, SecOps SecurityStatusBadge, AIOps StatusBadge, and many other components

### Files Changed (18 TSX files + 1 CSS file)
- `Sidebar.tsx`, `Settings.tsx` — version strings
- `ExportDialog.tsx`, `SettingsDialog.tsx` — full CSS theme refactor
- `AlertBadge.tsx`, `HealthCard.tsx` — CSS variable replacements
- `ConfirmDialog.tsx` — CSS variable replacements
- `DevOps.tsx`, `SecOps.tsx`, `NetOps.tsx` — StatusBadge + all broken classes
- `AIOps.tsx` — ChatBubble, StatusBadge, decorative elements
- `Logs.tsx` — bg-background → bg-[var(--color-bg)]
- `MainContent.tsx`, `SysOps.tsx` — bg-background → bg-[var(--color-bg)]
- `NetworkDesign.tsx` — full CSS theme refactor (toolbar, panels, canvas)
- `DeviceNode.tsx`, `ConnectionLine.tsx` — CSS vars
- `AreaChart.tsx`, `Gauge.tsx`, `MiniSparkline.tsx` — hardcoded hex → CSS vars
- `Dashboard.tsx` — skeleton loading state
- `globals.css` — animation keyframes + focus-visible styles

### Known Risks
- `tailwindcss-animate` package not added (avoided npm install without approval). Manual CSS animations used instead — behavior is equivalent
- Setting `appInfo.version` hardcoded to `1.1.1` — the `App.GetAppInfo()` Go binding will override this at runtime

## Known Issues
- TUI fully removed — no remnants ✅
- SQLite DB at `hawkward.db` in app working directory
- Build: `wails build` (NOT `go build`)

## Topics
- [[sprint-19-production-stabilization]] — Sprint 19 tickets and detail
- [[ui-ux-review]] — CSS theming fixes and UI audit results
