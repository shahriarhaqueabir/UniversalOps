import { useCallback } from 'react'

export interface WailsRuntime {
  EventsOn: (event: string, handler: (...args: unknown[]) => void) => void
  EventsOff: (event: string, handler: (...args: unknown[]) => void) => void
}

export interface WailsApp {
  [key: string]: Record<string, (...args: unknown[]) => Promise<unknown>> | undefined
}

/** Access the Wails Go backend bindings (window.go.app). */
export function getGo(): WailsApp | undefined {
  const w = window as { go?: { app?: WailsApp } }
  return w.go?.app
}

/** Access the Wails runtime API (window.runtime) for event subscriptions. */
export function getRuntime(): WailsRuntime | null {
  const w = window as { runtime?: WailsRuntime }
  return w.runtime ?? null
}

export function useBackend() {
  const call = useCallback(async (method: string, ...args: unknown[]) => {
    try {
      const go = getGo()
      if (go) {
        const parts = method.split('.')
        let target: unknown = go
        for (const part of parts) {
          if (!target) break
          target = (target as Record<string, unknown>)[part]
        }
        if (typeof target === 'function') {
          return await (target as (...args: unknown[]) => Promise<unknown>)(...args)
        }
      }

      const error = new Error(`Wails method not found: ${method}`)
      console.error(`[useBackend] ${error.message}`)
      throw error
    } catch (err: unknown) {
      console.error(`[useBackend] Wails call failed for ${method}:`, err)
      throw err
    }
  }, [])

  return { call }
}
