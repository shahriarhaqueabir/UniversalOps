import { useState, useEffect } from 'react'
import { RefreshCw } from 'lucide-react'

interface DataFreshnessIndicatorProps {
  /** Timestamp (Date or ISO string) of when data was last fetched */
  lastUpdated: Date | string | null
  /** Optional label shown before the timestamp */
  label?: string
  /** Show a spinning icon when stale (older than staleAfterMs) */
  showSpinner?: boolean
  /** Milliseconds after which data is considered stale — default 30s */
  staleAfterMs?: number
  /** Additional CSS classes */
  className?: string
}

function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 5) return 'just now'
  if (seconds < 60) return `${seconds}s ago`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

/**
 * Shows when data was last refreshed, with auto-updating "X ago" text.
 * Turns stale-colored after `staleAfterMs` milliseconds.
 */
export function DataFreshnessIndicator({
  lastUpdated,
  label = 'Last updated',
  showSpinner = true,
  staleAfterMs = 30_000,
  className = '',
}: DataFreshnessIndicatorProps) {
  const [, setTick] = useState(0)

  // Re-render every second to update the "X ago" text
  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 1000)
    return () => clearInterval(id)
  }, [])

  if (!lastUpdated) return null

  const date = lastUpdated instanceof Date ? lastUpdated : new Date(lastUpdated)
  const isStale = Date.now() - date.getTime() > staleAfterMs

  return (
    <span
      className={`inline-flex items-center gap-1.5 text-xs tabular-nums ${isStale
          ? 'text-[var(--color-warning, #f59e0b)]'
          : 'text-[var(--color-text-muted, #6b7280)]'
        } ${className}`}
    >
      {showSpinner && isStale && (
        <RefreshCw className="h-3 w-3 animate-spin" />
      )}
      {!isStale && (
        <RefreshCw className="h-3 w-3" />
      )}
      {label}: {timeAgo(date)}
    </span>
  )
}
