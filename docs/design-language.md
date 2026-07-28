# UniversalOps — Design Language

> Visual design principles and component conventions for UniversalOps.

---

## Core Mood

**Technical · Dense · Professional**

UniversalOps is a tool for SREs, security engineers, and developers. The UI should feel like an instrument panel — every pixel serves a purpose. No "delightful" animations that slow down information density. No padding for padding's sake.

---

## Color Palette

| Token | Hex | Usage |
|-------|-----|-------|
| `--bg-primary` | `#0f0f13` | Main background (dark) |
| `--bg-secondary` | `#1a1a24` | Card backgrounds, sidebars |
| `--bg-tertiary` | `#252533` | Hover states, elevated surfaces |
| `--border` | `#2e2e3e` | Borders, dividers |
| `--text-primary` | `#e8e8f0` | Primary text |
| `--text-secondary` | `#9090a8` | Secondary text, labels |
| `--accent` | `#7c6cff` | Primary accent (links, active states) |
| `--success` | `#34d399` | Healthy / OK states |
| `--warning` | `#fbbf24` | Warning states |
| `--danger` | `#ef4444` | Critical / error states |
| `--info` | `#60a5fa` | Informational states |

---

## Typography

- **UI**: System font stack (`-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`)
- **Monospace**: `"JetBrains Mono", "Cascadia Code", "Fira Code", monospace` — used for all telemetry values, logs, and terminal output
- **Scale**: 12px (label) / 14px (body) / 16px (h3) / 20px (h2) / 28px (h1)
- **Weights**: 400 (regular) for body, 500 (medium) for labels, 600 (semibold) for headings, 700 (bold) for metric values

---

## Spacing

- **Base unit**: 4px
- **Card padding**: 16px (`p-4`)
- **Section gap**: 24px (`gap-6`)
- **Grid gap**: 12px (`gap-3`)
- **Content max-width**: 1440px (dashboard), 1200px (detail views)

---

## Component Conventions

### Cards
- Background: `--bg-secondary`
- Border: `1px solid --border`
- Border-radius: 8px (`rounded-lg`)
- Padding: 16px
- No shadow (flat design — shadows add visual noise)

### Health Indicators
- Circular ring with color-coded arc
- Colors: success (green) / warning (yellow) / danger (red)
- Value displayed in center in monospace bold
- Label below in secondary text

### Metric Displays
- Value: monospace, semibold, larger than label
- Unit: secondary text, smaller, after value
- Trend arrow (↑↓) when historical data available
- Color-coded by severity

### Buttons
- Primary: filled accent (`--accent`)
- Secondary: outlined with border
- Danger: filled danger (`--danger`)
- Icon buttons: 32×32px, transparent bg, icon in --text-secondary
- All buttons: 6px border-radius, 8px horizontal padding

### Tables
- Header: `--bg-tertiary`, semibold, uppercase label
- Rows: alternating `--bg-secondary` / transparent
- Hover: `--bg-tertiary`
- Border: `1px solid --border` between rows
- Monospace for data columns

---

## Motion Principles

- **Duration**: 150ms for micro-interactions (hover, focus), 200ms for transitions (panel open/close)
- **Easing**: `cubic-bezier(0.4, 0, 0.2, 1)` (Material Design standard)
- **What moves**: Panel transitions, tooltip fades, progress bars, loading skeletons
- **What doesn't**: Metric values (they update in place), health indicators (instant state change)
- **No**: parallax, confetti, celebratory animations, or any "delight" that delays information delivery

---

## Layout

### Dashboard
- Full-width grid layout
- Health score + resource summary at top
- Per-ops-layer cards below in 2-3 column responsive grid
- No horizontal scroll at 1440px

### Detail Views (SysOps, NetOps, etc.)
- Left sidebar with tab navigation
- Main content area with tab-specific panels
- Sub-tabs for granular views (CPU / Memory / Disk / etc.)

### Responsive Breakpoints
- 1440px+: 3-column grid
- 1024px-1439px: 2-column grid
- 768px-1023px: single column, sidebar collapses to top nav
- Below 768px: not targeted (desktop-only application)

---

## Iconography

- **Library**: Lucide React (consistent 16px/20px/24px sizes)
- **Style**: Line icons, 1.5px stroke, no filled variants
- **Color**: Inherits from parent text color
- **Usage**: Section headers, button labels, status indicators, navigation items