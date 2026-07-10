# Handoff: Sprint 20 — CI Fix + UI/UX Deep Audit + Library Enhancement

## Goal

The user wants a production-quality, polished desktop operations platform (Hawkward GUI — Wails v2 + Go + React/TypeScript). This sprint's objectives were:
1. Fix CI `npm ci` eslint peer dep conflict (BLOCKER)
2. Deep UI/UX audit — fix all hardcoded theme colors, incomplete features, and accessibility gaps
3. Research & add missing recommended libraries
4. Fix the README badges showing "repo not found"

## State

### What's Done ✓

#### T-16: CI ESLint Fix
- Upgraded eslint from `^9.25.1` → `^10.6.0`, `@eslint/js` → `^10.0.0`, `eslint-plugin-react-hooks` → `^7.1.1`
- Added `jiti` devDep (eslint 10 requirement)
- Disabled 2 overly-strict react-hooks v7 rules (`set-state-in-effect`, `incompatible-library`) — they conflict with Wails desktop patterns
- Fixed 12 pre-existing lint errors across 7 files (typed mocks, ref handling, etc.)
- **Result**: `npm ci` ✅, `npm run lint` ✅ (0 errors, 0 warnings), `npm run build` ✅, `npm test` ✅ (27/27)

#### T-17: UI/UX Fixes
- **Sidebar.tsx**: Replaced `hover:bg-white/5` (breaks light mode) with `hover:bg-[var(--color-sidebar-hover)]` (3 instances)
- **DevOps.tsx**: Same fix in breadcrumb (2 instances)
- **TopBar.tsx**: Moved inline `style={{background}}` to `bg-[var(--color-bg)]` class
- **globals.css**: Light theme border changed from `rgba(0,0,0,0.07)` → `0.12` (was nearly invisible)
- **NetOps.tsx**: Hardcoded hex `#14b8a6`/`#3b82f6` → CSS vars `var(--color-success)`/`var(--color-accent)` (6 instances)
- **NetOps.tsx**: Ping entries now clear on START (`setPingEntries([])` when toggling on)
- **Dashboard.tsx**: `import type { Page }` moved from bottom to top of file
- **NetworkDesign.tsx**: Added localStorage auto-save (500ms debounce) with auto-restore on mount
- **App.tsx**: `any` types → typed `WailsRuntime` interface for Wails events
- **useEvents.ts**: Moved `handlerRef.current = handler` to dedicated `useEffect` (was writing to ref during render)

#### T-18: Library Installation
Verified every recommendation against actual `package.json`. Missing libraries found and installed:
- `@tanstack/react-table@^8.21.3` ✅
- `@radix-ui/react-collapsible@^1.1.16` ✅
- `@radix-ui/react-progress@^1.1.12` ✅
- `@radix-ui/react-toggle@^1.1.14` ✅
Not installed (P4): `prometheus/client_golang` (needs HTTP endpoint), `google/gopacket` (needs Npcap runtime)

#### README Fix
- Badges were dynamically pulling from `shields.io` → **repo is private** so shields showed "repo not found" / "no releases"
- Replaced with static badges showing actual values: `version-v1.2.0`, `Go-1.26.5`, `license-MIT`

### Current Blockers
- **No GitHub Releases exist** — tags are pushed (`v1.0.0` → `v1.2.0`) but `release.yml` pipeline either hasn't run successfully or `GITHUB_TOKEN` can't create releases on a private repo. The README download links point to an empty releases page.
- **TestPing** (`internal/netops/netops_test.go`) fails with `exit status 255` — requires admin/elevated privileges, expected on non-admin runs

### Git State
- **Branch**: `main`
- **Working tree**: Dirty — 16 source files modified, plus `package-lock.json` and `node_modules/`
- **Not yet committed or pushed**

### Files Changed (source only)
```
.memory/index.md
README.md
cmd/hawkward-gui/frontend/package.json
cmd/hawkward-gui/frontend/package-lock.json
cmd/hawkward-gui/frontend/eslint.config.js
cmd/hawkward-gui/frontend/src/App.tsx
cmd/hawkward-gui/frontend/src/hooks/useEvents.ts
cmd/hawkward-gui/frontend/src/components/layout/Sidebar.tsx
cmd/hawkward-gui/frontend/src/components/layout/TopBar.tsx
cmd/hawkward-gui/frontend/src/pages/Dashboard.tsx
cmd/hawkward-gui/frontend/src/pages/Dashboard.test.tsx
cmd/hawkward-gui/frontend/src/pages/DevOps.tsx
cmd/hawkward-gui/frontend/src/pages/Logs.test.tsx
cmd/hawkward-gui/frontend/src/pages/NetOps.tsx
cmd/hawkward-gui/frontend/src/pages/NetworkDesign.tsx
cmd/hawkward-gui/frontend/src/pages/Settings.test.tsx
cmd/hawkward-gui/frontend/src/test/utils.test.tsx
cmd/hawkward-gui/frontend/src/styles/globals.css
```

## Next Steps

### 1. Commit & Push
Stage all source changes and the lockfile, then commit and push. Suggested message:
```
feat: eslint 10 migration, UI/UX deep audit, new libraries

- Upgrade eslint to ^10.6.0, fix peer dep conflict blocking CI
- Fix 12 lint errors across 7 files (typed mocks, ref handling)
- Replace hardcoded white/hex colors with CSS vars in Sidebar, DevOps, NetOps
- Fix light theme border visibility, TopBar inline style
- Clear ping entries on new probe, auto-save NetworkDesign canvas
- Add @tanstack/react-table, @radix-ui/react-collapsible/progress/toggle
- Fix README badges for private repo (static badges)
```

### 2. Create GitHub Release
The `release.yml` pipeline creates releases on tag push. Options:
- **Easiest**: Push a new tag `v1.3.0` to trigger `release.yml` — but verify `GITHUB_TOKEN` has release permissions on the private repo
- **Manual**: Create a GitHub Release manually through the UI and upload `wails build` artifacts
- **Fix pipeline**: If `GITHUB_TOKEN` can't create releases on private repos, switch to a Personal Access Token (PAT) stored as a repo secret

### 3. Fix TestPing (Optional)
`internal/netops/netops_test.go:14` — `TestPing` fails because ping requires admin. Fix: skip test when not admin:
```go
func TestPing(t *testing.T) {
    if os.Getuid() != 0 {
        t.Skip("Skipping ping test: requires admin privileges")
    }
    // ...
}
```

### 4. Complete Missing Test Coverage (P3)
These pages have no tests:
- `TopBar`, `NetOps`, `SecOps`, `DevOps`, `AIOps`, `NetworkDesign`, `useBackend`, `useEvents`, `useSettingsStore`, `useOllamaStore`
Current test count: 27 tests across 7 files. Target: add tests for at least the core stores and hooks.

## Context & Constraints

### Build System
- **Always use `wails build`** (not `go build`) — Wails embeds frontend dist via `//go:embed`
- Frontend dev: `cd cmd/hawkward-gui/frontend && npm run dev`
- Go tests: `go test ./internal/... -count=1`
- Frontend tests: `cd cmd/hawkward-gui/frontend && npm test`

### Tech Stack
- **Go 1.26.5** with `gopsutil/v4`, `miekg/dns`, `zerolog`, `ollama SDK`, `modernc.org/sqlite`
- **React 19 + TypeScript + Vite 6 + Tailwind v4** (Wails v2 embed)
- **State**: zustand v5 (stores in `src/stores/useSettingsStore.ts`)
- **Data fetching**: @tanstack/react-query v5
- **Theme**: CSS variables via `src/styles/globals.css` with `data-theme` attribute on `<html>`
- **All CSS vars** are defined in `@theme` block in globals.css — never use hardcoded colors

### Key Lint Learning
The `eslint-plugin-react-hooks@7.x` has strict new rules. For Wails desktop patterns:
- `react-hooks/set-state-in-effect` — data fetching in `useEffect` is standard for Wails (no RSC/`use()` available)
- `react-hooks/incompatible-library` — `@tanstack/react-virtual`'s `useVirtualizer` triggers this
Both are disabled in `eslint.config.js`.

### Private Repo Impact
- shields.io badges can't read private repo metadata → use static badges
- `GITHUB_TOKEN` in Actions may not have permission to create releases on private repos → may need PAT
- Download links in README are non-functional until a release is created

## Pitfalls / Tried & Didn't Work

1. **Disabling eslint rule with inline comment with description**: `// eslint-disable-next-line rule-name -- description` needs `--` separator. Without `--`, the description text is parsed as additional rule names.
2. **`useRef` + lazy `useState` init**: Using `useRef(loadSavedTopology())` then accessing `saved.current` inside `useState(() => saved.current)` triggers `react-hooks/refs` rule. Fix: call the sync function directly in the lazy init (it's synchronous localStorage, so calling twice is harmless).
3. **`npm test` intermittent failure**: "Worker exited unexpectedly" — transient tinypool worker crash. Always retry before debugging.
4. **`wails build` vs `go build`**: The memory explicitly warns about this. `go build` fails because it can't resolve `//go:embed` for the frontend dist. Always use `wails build`.
