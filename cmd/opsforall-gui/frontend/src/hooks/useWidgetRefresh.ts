import { useEffect, useRef, useCallback, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'

interface WidgetRefreshOptions<T> {
  /** Unique query key for TanStack Query dedup/caching */
  queryKey: string | readonly unknown[]
  /** Async fetcher returning widget data */
  fetcher: () => Promise<T>
  /** Polling interval in ms (default: 5000) */
  intervalMs?: number
  /** Enable polling (default: true) */
  enabled?: boolean
  /** Optional Wails runtime event name that triggers a refresh (avoids polling) */
  eventTrigger?: string
  /** Stale time in ms (default: intervalMs) */
  staleTimeMs?: number
}

interface WidgetRefreshResult<T> {
  data: T | undefined
  isLoading: boolean
  error: Error | null
  isStale: boolean
  lastUpdated: Date | null
  refresh: () => Promise<void>
}

interface WailsRuntime {
  EventsOn: (event: string, handler: (...args: unknown[]) => void) => void
  EventsOff: (event: string, handler: (...args: unknown[]) => void) => void
}

function getRuntime(): WailsRuntime | null {
  const w = window as { runtime?: WailsRuntime }
  return w.runtime ?? null
}

/**
 * useWidgetRefresh — drives a widget or dashboard panel's data lifecycle.
 *
 * Combines TanStack Query polling with optional Wails runtime event
 * triggers so widgets stay fresh without redundant intervals when the
 * backend pushes an update event.
 *
 * @example
 * ```tsx
 * const { data, isLoading, refresh } = useWidgetRefresh({
 *   queryKey: ['system-snapshot'],
 *   fetcher: () => call('Dashboard.GetSystemSnapshot'),
 *   intervalMs: 3000,
 *   eventTrigger: 'metrics:update',
 * })
 * ```
 */
export function useWidgetRefresh<T>(options: WidgetRefreshOptions<T>): WidgetRefreshResult<T> {
  const {
    queryKey,
    fetcher,
    intervalMs = 5000,
    enabled = true,
    eventTrigger,
    staleTimeMs = intervalMs,
  } = options

  const queryClient = useQueryClient()
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)
  const fetcherRef = useRef(fetcher)
  const handlerRef = useRef<((...args: unknown[]) => void) | null>(null)

  // Keep fetcher ref fresh
  useEffect(() => {
    fetcherRef.current = fetcher
  }, [fetcher])

  const query = useQuery<T, Error>({
    queryKey: Array.isArray(queryKey) ? queryKey : [queryKey],
    queryFn: async () => {
      const result = await fetcherRef.current()
      setLastUpdated(new Date())
      return result
    },
    refetchInterval: enabled ? intervalMs : false,
    staleTime: staleTimeMs,
    enabled,
  })

  // Manual refresh helper
  const refresh = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: Array.isArray(queryKey) ? queryKey : [queryKey] })
  }, [queryClient, queryKey])

  // Wails runtime event trigger — when the backend pushes an event,
  // invalidate the query so it re-fetches on next tick.
  useEffect(() => {
    if (!eventTrigger || !enabled) return

    const runtime = getRuntime()
    if (!runtime?.EventsOn) return

    const handler = () => {
      queryClient.invalidateQueries({ queryKey: Array.isArray(queryKey) ? queryKey : [queryKey] })
    }
    handlerRef.current = handler
    runtime.EventsOn(eventTrigger, handler)

    return () => {
      if (handlerRef.current) {
        runtime.EventsOff(eventTrigger, handlerRef.current)
        handlerRef.current = null
      }
    }
  }, [eventTrigger, enabled, queryClient, queryKey])

  return {
    data: query.data,
    isLoading: query.isLoading,
    error: query.error,
    isStale: query.isStale,
    lastUpdated,
    refresh,
  }
}
