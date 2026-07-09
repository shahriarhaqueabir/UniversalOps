# UI/UX Review & CSS Theming Audit

## Overview
Completed a comprehensive UI/UX audit across all Hawkward frontend pages and components. All fixes applied and verified with `tsc -b`.

## Findings

### 1. CSS Variable Theming Issues (CRITICAL)
The project uses custom CSS variables defined in `@theme` block (`--color-accent`, `--color-panel`, `--color-text`, etc.) but many components referenced Tailwind utility classes for colors that don't exist in the theme:

**Undefined colors that were in use:**
- `bg-card`, `text-primary`, `bg-primary`, `text-primary-foreground` — shadcn/ui naming, not in project
- `text-muted`, `bg-muted` — not in project theme
- `bg-background` — not defined (`--color-background` doesn't exist; project uses `--color-bg`)
- `border-primary`, `border-primary/*` 
- `placeholder-muted` — Tailwind v4 placeholder color utility

**All 18 TSX files fixed** by replacing with CSS variable equivalents:
- `bg-card` → `bg-panel`
- `text-primary` → `text-accent`
- `bg-primary` → `bg-accent`
- `text-primary-foreground` → `text-white`
- `text-muted` → `text-text-faint`
- `bg-muted` → `bg-text-faint` or `bg-panel-2`
- `bg-background` → `bg-[var(--color-bg)]`
- `border-primary` → `border-accent`
- `placeholder-muted` → `placeholder:text-text-faint`

### 2. Missing Animation Plugin
`tailwindcss-animate` (provides `animate-in`, `fade-in`, `zoom-in-95`, `slide-in-from-*` classes) is not in `package.json`. Manual CSS keyframe animations added as fallback.

### 3. Hardcoded Colors in Chart Components
Chart components used hardcoded Tailwind slate hex values:
- `#334155`, `#1e293b`, `#94a3b8`, `#f8fafc` — replaced with CSS variable references

### 4. Dashboard Loading State
Dashboard showed plain text "Synching Neural Bridge..." on loading. Replaced with proper skeleton matching the Squib-inspired design of other pages.

### 5. Version String Inconsistency
- Sidebar footer showed `v0.1.0` — actual version in `wails.json` is `1.1.1`
- Settings DEFAULT_APP_INFO showed `0.1.0` — fixed to `1.1.1`

### 6. Keyboard Accessibility
No global `:focus-visible` styles were defined. Added global CSS rule setting accent-colored focus ring on all interactive elements.

### 7. Dead Components (Identified but NOT removed)
These components exist but are never imported by any page:
- `components/dashboard/HealthCard.tsx` — never imported (Dashboard uses inline KpiCard)
- `components/dashboard/AlertBadge.tsx` — never imported
- `components/charts/AreaChart.tsx` — never imported (pages use Recharts directly)
- `components/charts/Gauge.tsx` — never imported
- `components/charts/MiniSparkline.tsx` — only imported by HealthCard (also dead)
- `components/dialogs/ExportDialog.tsx` — never imported
- `components/dialogs/SettingsDialog.tsx` — never imported

These were left in place as they don't affect build or runtime.

## Verification
- `tsc -b` passed with zero errors
- No new npm dependencies added
