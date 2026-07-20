import { cn } from '@/lib/utils'

export interface PanelProps {
  children: React.ReactNode
  /** Visual variant */
  variant?: 'default' | 'elevated' | 'inset' | 'flat'
  /** Category tint — adds a subtle colored left border and icon glow */
  category?: 'network' | 'security' | 'system' | 'devops' | 'ai' | 'none'
  /** Padding size */
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl'
  /** Additional class names */
  className?: string
  /** Optional click handler */
  onClick?: () => void
}

const categoryColors = {
  network: {
    border: 'border-l-[var(--color-accent)]',
    bg: 'bg-[var(--color-accent)]/5',
    icon: 'text-[var(--color-accent)]',
  },
  security: {
    border: 'border-l-[var(--color-danger)]',
    bg: 'bg-[var(--color-danger)]/5',
    icon: 'text-[var(--color-danger)]',
  },
  system: {
    border: 'border-l-[var(--color-success)]',
    bg: 'bg-[var(--color-success)]/5',
    icon: 'text-[var(--color-success)]',
  },
  devops: {
    border: 'border-l-[var(--color-warning)]',
    bg: 'bg-[var(--color-warning)]/5',
    icon: 'text-[var(--color-warning)]',
  },
  ai: {
    border: 'border-l-[var(--color-accent-2)]',
    bg: 'bg-[var(--color-accent-2)]/5',
    icon: 'text-[var(--color-accent-2)]',
  },
  none: {
    border: '',
    bg: '',
    icon: '',
  },
}

const variantStyles = {
  default: 'bg-[var(--color-panel)] border border-[var(--color-border)] shadow-xl',
  elevated: 'bg-[var(--color-panel)] border border-[var(--color-border)] shadow-2xl',
  inset: 'bg-[var(--color-panel-2)] border border-[var(--color-border)] shadow-inner',
  flat: 'bg-[var(--color-panel)] border border-[var(--color-border)]',
}

const paddingStyles = {
  none: '',
  sm: 'p-4',
  md: 'p-6',
  lg: 'p-8',
  xl: 'p-12', // Generous padding for high-radius corners
}

/**
 * Panel — consistent card/frame wrapper used across the application.
 *
 * Re-aligned to new high-radius industrial design standards.
 */
export function Panel({
  children,
  variant = 'default',
  category = 'none',
  padding = 'md',
  className,
  onClick,
}: PanelProps) {
  const cat = categoryColors[category]

  return (
    <div
      onClick={onClick}
      className={cn(
        'rounded-[2rem] transition-all duration-300', // Standardized to 2rem
        variantStyles[variant],
        paddingStyles[padding],
        category !== 'none' && ['border-l-4', cat.border, cat.bg],
        onClick && 'cursor-pointer hover:scale-[1.01] active:scale-[0.99] hover:border-accent/30',
        className,
      )}
    >
      {children}
    </div>
  )
}

/**
 * PanelHeader — consistent panel header with icon, title, and optional action.
 */
export function PanelHeader({
  icon,
  title,
  subtitle,
  action,
  category = 'none',
  className,
}: {
  icon?: React.ReactNode
  title: string
  subtitle?: string
  action?: React.ReactNode
  category?: PanelProps['category']
  className?: string
}) {
  const cat = categoryColors[category]

  return (
    <div className={cn('flex items-center justify-between mb-6', className)}>
      <div className="flex items-center gap-4">
        {icon && (
          <div className={cn(
            'w-11 h-11 rounded-xl flex items-center justify-center border border-[var(--color-border)] bg-[var(--color-panel-3)] shrink-0 shadow-inner',
            category !== 'none' && cat.icon,
          )}>
            {icon}
          </div>
        )}
        <div>
          <h3 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">{title}</h3>
          {subtitle && <p className="text-[10px] text-[var(--color-text-dim)] mt-1 font-bold uppercase tracking-widest">{subtitle}</p>}
        </div>
      </div>
      {action}
    </div>
  )
}
