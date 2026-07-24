import { cn } from '@/lib/utils'

/** Semantic color mapping for sidebar category groups */
const groupColors: Record<string, { dot: string; activeBorder: string; activeShadow: string; iconDefault: string }> = {
  // NetOps / SysOps groups
  inspection: {
    dot: 'bg-[var(--color-accent)]',
    activeBorder: 'border-l-[var(--color-accent)]',
    activeShadow: 'shadow-[var(--color-accent)]/20',
    iconDefault: 'text-[var(--color-accent)]',
  },
  diagnosis: {
    dot: 'bg-[var(--color-success)]',
    activeBorder: 'border-l-[var(--color-success)]',
    activeShadow: 'shadow-[var(--color-success)]/20',
    iconDefault: 'text-[var(--color-success)]',
  },
  action: {
    dot: 'bg-[var(--color-warning)]',
    activeBorder: 'border-l-[var(--color-warning)]',
    activeShadow: 'shadow-[var(--color-warning)]/20',
    iconDefault: 'text-[var(--color-warning)]',
  },
  // SecOps groups
  assessment: {
    dot: 'bg-[var(--color-success)]',
    activeBorder: 'border-l-[var(--color-success)]',
    activeShadow: 'shadow-[var(--color-success)]/20',
    iconDefault: 'text-[var(--color-success)]',
  },
  detection: {
    dot: 'bg-[var(--color-warning)]',
    activeBorder: 'border-l-[var(--color-warning)]',
    activeShadow: 'shadow-[var(--color-warning)]/20',
    iconDefault: 'text-[var(--color-warning)]',
  },
  response: {
    dot: 'bg-[var(--color-danger)]',
    activeBorder: 'border-l-[var(--color-danger)]',
    activeShadow: 'shadow-[var(--color-danger)]/20',
    iconDefault: 'text-[var(--color-danger)]',
  },
}

/** Override map: per-page accent colors used for active item background */
const pageAccent: Record<string, string> = {
  // NetOps pages use accent
  'netops-inspection': 'bg-[var(--color-accent)]',
  'netops-diagnosis': 'bg-[var(--color-accent)]',
  'netops-action': 'bg-[var(--color-accent)]',
  // SysOps pages use accent
  'sysops-inspection': 'bg-[var(--color-accent)]',
  'sysops-diagnosis': 'bg-[var(--color-accent)]',
  'sysops-action': 'bg-[var(--color-accent)]',
  // SecOps pages use danger
  'secops-assessment': 'bg-[var(--color-danger)]',
  'secops-detection': 'bg-[var(--color-danger)]',
  'secops-response': 'bg-[var(--color-danger)]',
}

export interface CategoryGroupProps<T extends string> {
  label: string
  group: string
  page: string
  categories: Array<{ id: T; label: string; icon: React.ReactNode }>
  active: T
  onSelect: (id: T) => void
}

/**
 * Shared sidebar category group with colored dot indicator,
 * consistent spacing, and active-state accent border.
 *
 * Replaces the 3 separate CategoryGroup implementations in NetOps, SysOps, and SecOps.
 */
export function CategoryGroup<T extends string>({
  label,
  group,
  page,
  categories,
  active,
  onSelect,
}: CategoryGroupProps<T>) {
  const colors = groupColors[group] || groupColors.inspection
  const activeBg = pageAccent[`${page}-${group}`] || 'bg-[var(--color-accent)]'

  return (
    <div className="mb-5">
      {/* Group header with colored dot */}
      <div className="flex items-center gap-2 px-2.5 mb-2.5">
        <div className={cn('w-1.5 h-1.5 rounded-full shrink-0', colors.dot)} />
        <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-[0.2em]">{label}</p>
      </div>

      {/* Items */}
      <div className="space-y-1">
        {categories.map(cat => {
          const isActive = active === cat.id
          return (
            <button
              key={cat.id}
              onClick={() => onSelect(cat.id)}
              data-automation-id={`${page}-tab-${cat.id}`}
              className={cn(
                'w-full flex items-center gap-3 px-2.5 py-2 rounded-xl text-sm font-bold transition-all active:scale-[0.97]',
                isActive
                  ? cn(activeBg, 'text-white', cn('shadow-lg', colors.activeShadow))
                  : 'text-[var(--color-text-dim)] hover:text-[var(--color-text)] hover:bg-[var(--color-sidebar-hover)]'
              )}
            >
              <div className={cn(
                'transition-colors',
                isActive ? 'text-white' : colors.iconDefault,
              )}>
                {cat.icon}
              </div>
              {cat.label}
            </button>
          )
        })}
      </div>
    </div>
  )
}
