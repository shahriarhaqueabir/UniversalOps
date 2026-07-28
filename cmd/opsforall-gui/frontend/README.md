# UniversalOps Frontend

> React 19 + TypeScript + Vite + Tailwind v4 frontend for UniversalOps.

---

## Tech Stack

| Layer | Library | Purpose |
|-------|---------|---------|
| **Framework** | React 19 | UI components and rendering |
| **Language** | TypeScript (strict) | Type safety |
| **Build** | Vite | Dev server and production bundling |
| **Styling** | Tailwind v4 | Utility-first CSS |
| **State** | Zustand | Lightweight global state |
| **Data Fetching** | TanStack Query (React Query) | Server state, caching, refetching |
| **Charts** | Recharts | Telemetry visualizations |
| **Icons** | Lucide React | Consistent icon set |
| **Components** | Radix UI | Accessible, unstyled primitives |
| **Animation** | motion/react | Micro-interactions and transitions |
| **Testing** | Vitest + React Testing Library | Component and hook tests |

---

## Project Structure

```
frontend/
├── src/
│   ├── components/       ← Shared UI components (HealthRing, MiniStat, etc.)
│   ├── hooks/            ← Custom React hooks
│   ├── lib/              ← Utility functions, API client
│   ├── pages/            ← Page-level components
│   │   ├── SysOps/       ← System Operations tabs
│   │   ├── NetOps/       ← Network Operations
│   │   ├── SecOps/       ← Security Operations
│   │   ├── DevOps/       ← DevOps Automation
│   │   └── AIOps/        ← AI Operations
│   ├── store/            ← Zustand stores
│   ├── types/            ← TypeScript type definitions
│   └── App.tsx           ← Root component with routing
├── public/               ← Static assets
├── index.html            ← Entry HTML
├── vite.config.ts        ← Vite configuration
├── tailwind.config.ts    ← Tailwind configuration
└── package.json          ← Dependencies and scripts
```

---

## Available Scripts

```bash
# Development (run via `wails dev` from project root)
npm run dev

# Testing
npm test            # Run tests once
npm run test:watch  # Watch mode
npm run test:ui     # Vitest UI mode

# Linting & Type Checking
npm run lint        # ESLint
npx tsc --noEmit    # TypeScript check

# Build (run via `wails build` from project root)
npm run build
```

---

## Key Patterns

### Data Fetching
All backend calls go through TanStack Query:
```typescript
const { data, isLoading } = useQuery({
  queryKey: ['sysops', 'cpu'],
  queryFn: () => window.go.app.SysOps.GetCPUUsage(),
  refetchInterval: 3000, // Poll every 3 seconds
});
```

### State Management
Use Zustand for UI-only state (selected tabs, preferences):
```typescript
const useSysOpsStore = create<Store>((set) => ({
  activeTab: 'overview',
  setActiveTab: (tab) => set({ activeTab: tab }),
}));
```

### Component Conventions
- Functional components with TypeScript interfaces for props
- `React.memo` for components that render frequently (metric cards, health rings)
- CSS variables for theming (defined in `index.css`)
- Lucide icons for all iconography

### Percentage Display
All percentage values use `Math.round()` for integer display. No decimal places in telemetry values.

---

## Testing

- **Framework**: Vitest + React Testing Library
- **Coverage target**: 80%+ for new components
- **What to test**: Data fetching hooks, component rendering, state transitions, error states
- **What not to test**: Wails bridge calls (tested in Go backend), third-party library behavior

```bash
# Run with coverage
npx vitest run --coverage
```