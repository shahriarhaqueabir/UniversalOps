# Sprint 9: Theme + Polish + Test Coverage (HARDENING)

> **Created**: 2026-07-08
> **Goal**: Close remaining known issues — light theme, log viewer performance, test coverage gaps, Settings page polish, and sidebar test fix.

---

## Recon Summary

After completing Sprints 7 (GUI finalization), 8 (Multi-Perspective Audit), and 8.2 (Residual Hardening), the following gaps remain:

### Verified Remaining Issues

| # | Issue | Severity | Root Cause |
|---|-------|----------|------------|
| P-01 | **Light theme CSS missing** | High | `globals.css` has no `[data-theme="light"]` block; `useTheme.ts` sets attribute but no light values defined |
| P-02 | **Settings theme toggle disconnected** | High | `Settings.tsx` has its own `darkMode` state that doesn't call `useTheme()` — toggling does nothing |
| P-03 | **Log viewer renders all rows** | Medium | `Logs.tsx` renders unlimited `<tr>` with no virtual scrolling; thousands of entries will cause jank |
| P-04 | **Settings page uses stale CSS classes** | Medium | Uses `bg-card`, `text-primary`, `text-muted` etc. not defined in Squib design system |
| P-05 | **Sidebar test collapse assertion broken** | Low | `Sidebar.test.tsx` calls `getByLabelText('Collapse sidebar')` but Sidebar has no such aria-label; collapsed is hardcoded to `false` |
| P-06 | **Limited frontend test coverage** | Medium | Only `Sidebar.test.tsx` exists; Dashboard, Logs, Settings have no tests |
| P-07 | **No offline/loading states on non-Dashboard pages** | Low | Dashboard has loading skeleton; other pages don't handle missing backend gracefully |

---

## Kanban Board

| ID | Ticket | Status | Priority | File Scope | DOD |
|----|--------|--------|----------|------------|-----|
| P-01 | **Light Theme CSS Variables** | 🔲 TODO | High | `frontend/src/styles/globals.css` | - `[data-theme="light"]` block exists with all required vars<br>- Light theme renders with light bg, dark text<br>- All components readable |
| P-02 | **Fix Settings Theme Toggle** | 🔲 TODO | High | `frontend/src/pages/Settings.tsx` | - Settings theme toggle uses `useTheme()` hook<br>- Dark/light switching works end-to-end from Settings page<br>- All settings component CSS vars match design system |
| P-03 | **Log Viewer Virtual Scrolling** | 🔲 TODO | Medium | `frontend/src/pages/Logs.tsx` | - Virtual scroll using windowing (react-window or IntersectionObserver)<br>- Max ~50 visible rows rendered at any time<br>- Auto-scroll and search still work<br>- No new npm deps if avoidable |
| P-04 | **Settings Page Design System Alignment** | 🔲 TODO | Medium | `frontend/src/pages/Settings.tsx` | - All class names use Squib vars: `bg-panel`, `text-text`, `text-text-dim`, `border-border` etc.<br>- Section cards match Dashboard card style<br>- Consistent spacing and typography |
| P-05 | **Fix Sidebar Collapse Test** | 🔲 TODO | Low | `frontend/src/components/layout/Sidebar.tsx`<br>`frontend/src/components/layout/Sidebar.test.tsx` | - Sidebar has working aria-labels for collapse/expand<br>- Collapse test passes<br>- `npm test` runs clean |
| P-06 | **Frontend Test Coverage Expansion** | 🔲 TODO | Medium | `*.test.tsx` files for pages | - Tests for Dashboard, Logs, Settings pages<br>- Tests verify render and key interactions<br>- Coverage threshold 50%+ |
| P-07 | **Loading & Offline States** | 🔲 TODO | Low | `SysOps.tsx`, `NetOps.tsx`, `SecOps.tsx`, `DevOps.tsx`, `AIOps.tsx`, `Settings.tsx` | - Each page shows loading skeleton while data fetches<br>- Graceful fallback when backend call returns error/null<br>- No blank/hanging states |

---

## Lane Matrix

| Lane | Write Surface | Risk | Parallelizable |
|------|---------------|------|----------------|
| P-01 Light Theme CSS | `globals.css` only | Low | Yes (with P-02) |
| P-02 Settings Theme Fix | `Settings.tsx` only | Low | Yes (with P-01) |
| P-03 Log Virtual Scrolling | `Logs.tsx` only | Medium | No — needs careful testing |
| P-04 Settings CSS Align | `Settings.tsx` only | Low | Yes (with P-02) |
| P-05 Sidebar Test Fix | `Sidebar.tsx` + `.test.tsx` | Low | Yes |
| P-06 Frontend Tests | `*.test.tsx` files | Low | Yes |
| P-07 Offline States | Multiple pages | Low | Yes |

---

## Baseline Verification

Before implementation, establish:
- [ ] `go test ./internal/...` passes
- [ ] `go vet ./internal/...` clean
- [ ] Frontend builds (`npm run build` or `tsc -b`)
- [ ] Current test failures documented
