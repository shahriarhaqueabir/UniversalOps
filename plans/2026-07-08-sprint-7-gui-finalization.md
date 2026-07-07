# Sprint 7: GUI Finalization & Production Readiness

> **Active Sprint Plan** | Created: 2026-07-08
>
> **Goal**: Complete the transition from a mock-powered GUI to a fully functional Wails desktop app
> with real data pipelines, comprehensive testing, and production-ready architecture.

---

## Kanban Board

| ID | Ticket | Status | Priority | DOD |
|----|--------|--------|----------|-----|
| T-01 | **Add Test Coverage for `internal/app`** | ✅ DONE | High | - [x] `App_test.go` exists<br>- [x] Subsystem facades tested<br>- [x] 80%+ coverage for `internal/app` |
| T-02 | **Wails Bindings & Dev Mode Fix** | ✅ DONE | High | - [x] Bindings generated in `wailsjs/`<br>- [x] Dev mode uses bindings instead of `window.go` directly |
| T-03 | **End-to-End Event System** | ✅ DONE | High | - [x] `useEvents.ts` subscribes to real Wails events<br>- [x] Events correctly update frontend state |
| T-04 | **Remove Mock Data & Real Pipelines** | ✅ DONE | High | - [x] `useBackend` fallback to mock removed/disabled<br>- [x] All 9 pages use real Go data (Dashboard updated as example) |
| T-05 | **Frontend Testing Setup** | ✅ DONE | Med | - [x] Vitest + RTL configured<br>- [x] Baseline tests for core components |
| T-06 | **Code Splitting & Optimization** | ✅ DONE | Med | - [x] `React.lazy()` for all pages<br>- [x] Bundle size warning resolved |
| T-07 | **Error Boundaries & Logging** | ✅ DONE | Med | - [x] `ErrorBoundary` component added<br>- [x] Proper logging in Go and JS |
| T-08 | **Network Designer Save/Load** | ✅ DONE | Med | - [x] Wails File Dialog used for Save/Load<br>- [x] Persistence to JSON working |
| T-09 | **Settings Persistence** | ✅ DONE | Med | - [x] Settings changes update Go `common.Config`<br>- [x] Persistence across restarts |
| T-10 | **Remove Old TUI Legacy** | ✅ DONE | Low | - [x] `cmd/hawkward/` deleted<br>- [x] References to TUI removed from build scripts |

---

## Active Work

### T-01: Add Test Coverage for `internal/app`
We need to ensure the Go bridge is robust.
- Create `internal/app/App_test.go`
- Create `internal/app/SysOps_test.go`
- Mock the `runtime` if possible or test the logic that prepares the events.

### T-02: Wails Bindings
- Run `wails generate module` or just `wails build` without `-skipbindings`.
- Update `frontend/src/hooks/useBackend.ts` to use the generated bindings.

### T-03: Event System
- Verify `runtime.EventsOn` in `useEvents.ts`.
- Ensure `collectAndEmit` in `App.go` is actually firing.

### T-10: Remove TUI Legacy
- Delete `cmd/hawkward/`
- Update `CLAUDE.md` and `README.md` to reflect the new GUI focus.
