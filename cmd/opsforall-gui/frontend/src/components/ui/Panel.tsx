import { cn } from '@/lib/utils'

export interface PanelProps {
  children: React.ReactNode
  /** Visual variant */
  variant?: 'default' | 'elevated' | 'inset' | 'flat'
  /** Category tint — adds a subtle colored left border and icon glow */
  category?: 'network' | 'security' | 'system' | 'devops' | 'ai' | 'none'
  /** Padding size */
  padding?: 'none' | 'sm' | 'md' | 'lg'
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
}

/**
 * Panel — consistent card/frame wrapper used across the application.
 *
 * Standardizes radius (`var(--radius-lg)`), border, shadow, and provides
 * optional category tinting for colored left-border accents.
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
        'rounded-[var(--radius-lg)] transition-all duration-200',
        variantStyles[variant],
        paddingStyles[padding],
        category !== 'none' && ['border-l-2', cat.border, cat.bg],
        onClick && 'cursor-pointer hover:scale-[1.005] active:scale-[0.995]',
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
    <div className={cn('flex items-center justify-between', className)}>
      <div className="flex items-center gap-3">
        {icon && (
          <div className={cn(
            'w-9 h-9 rounded-lg flex items-center justify-center border border-[var(--color-border)] bg-[var(--color-panel-3)] shrink-0',
            category !== 'none' && cat.icon,
          )}>
            {icon}
          </div>
        )}
        <div>
          <h3 className="text-sm font-bold text-[var(--color-text)] uppercase tracking-widest">{title}</h3>
          {subtitle && <p className="text-xs text-[var(--color-text-dim)] mt-0.5">{subtitle}</p>}
        </div>
      </div>
      {action}
    </div>
  )
}
