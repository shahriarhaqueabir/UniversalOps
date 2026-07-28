import { cn } from '@/lib/utils'
import type { VariantProps } from 'class-variance-authority'
import { cva } from 'class-variance-authority'

/**
 * HealthBadge — a shape-aware health indicator.
 *
 * Supports four shape modes for accessibility (colour is never the
 * sole differentiator) and three severity levels plus "unknown".
 *
 * @example
 * ```tsx
 * <HealthBadge status="healthy" />
 * <HealthBadge status="critical" shapeMode="diamond" showLabel pulse />
 * ```
 */

// ── Health status type ──

export type HealthStatus = 'healthy' | 'degraded' | 'critical' | 'unknown'

// ── Shape definitions ──

type ShapeMode = 'circle' | 'square' | 'triangle' | 'diamond'

const shapeStyles: Record<ShapeMode, string> = {
  circle: 'rounded-full',
  square: 'rounded-sm',
  triangle: '',
  diamond: 'rotate-45',
}

// ── CVA variant definitions ──

const badgeVariants = cva(
  'inline-flex items-center gap-1.5 shrink-0',
  {
    variants: {
      size: {
        sm: 'text-[10px]',
        md: 'text-xs',
        lg: 'text-sm',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  },
)

const indicatorVariants = cva(
  'shrink-0',
  {
    variants: {
      size: {
        sm: 'h-2 w-2',
        md: 'h-2.5 w-2.5',
        lg: 'h-3 w-3',
      },
      status: {
        healthy: 'bg-emerald-500',
        degraded: 'bg-amber-400',
        critical: 'bg-red-500',
        unknown: 'bg-neutral-400',
      },
      pulse: {
        true: 'animate-pulse',
        false: '',
      },
    },
    defaultVariants: {
      size: 'md',
      status: 'unknown',
      pulse: false,
    },
  },
)

const labelVariants = cva(
  'font-medium select-none',
  {
    variants: {
      size: {
        sm: 'text-[10px]',
        md: 'text-xs',
        lg: 'text-sm',
      },
      status: {
        healthy: 'text-emerald-600 dark:text-emerald-400',
        degraded: 'text-amber-600 dark:text-amber-400',
        critical: 'text-red-600 dark:text-red-400',
        unknown: 'text-neutral-500 dark:text-neutral-400',
      },
    },
    defaultVariants: {
      size: 'md',
      status: 'unknown',
    },
  },
)

// ── Triangle SVG (can't be done with CSS alone) ──

function TriangleShape({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 10 10"
      className={cn('shrink-0', className)}
      aria-hidden="true"
    >
      <polygon points="5,1 9,9 1,9" className="fill-current" />
    </svg>
  )
}

// ── Props ──

export interface HealthBadgeProps extends VariantProps<typeof badgeVariants> {
  status: HealthStatus
  shapeMode?: ShapeMode
  showLabel?: boolean
  pulse?: boolean
  className?: string
}

// ── Component ──

export function HealthBadge({
  status,
  shapeMode = 'circle',
  showLabel = true,
  pulse: shouldPulse,
  size,
  className,
}: HealthBadgeProps) {
  const pulseClass = shouldPulse ?? (status === 'critical' || status === 'degraded')

  const statusLabel: Record<HealthStatus, string> = {
    healthy: 'Healthy',
    degraded: 'Degraded',
    critical: 'Critical',
    unknown: 'Unknown',
  }

  return (
    <span
      className={cn(badgeVariants({ size }), className)}
      role="status"
      aria-label={`Health: ${statusLabel[status]}`}
      data-status={status}
      data-shape={shapeMode}
    >
      {shapeMode === 'triangle' ? (
        <TriangleShape className={indicatorVariants({ size, status, pulse: pulseClass })} />
      ) : (
        <span
          className={cn(
            indicatorVariants({ size, status, pulse: pulseClass }),
            shapeStyles[shapeMode],
            shapeMode === 'diamond' && 'mb-0.5',
          )}
          aria-hidden="true"
        />
      )}
      {showLabel && (
        <span className={labelVariants({ size, status })}>
          {statusLabel[status]}
        </span>
      )}
    </span>
  )
}
