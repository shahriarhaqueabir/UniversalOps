import { useCallback } from 'react'

interface WailsApp {
  [key: string]: Record<string, (...args: unknown[]) => Promise<unknown>> | undefined
}

function getGo(): WailsApp | undefined {
  const w = window as { go?: { app?: WailsApp } }
  return w.go?.app
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

      throw new Error(`[useBackend] Wails method not found: ${method}`)
    } catch (err: unknown) {
      console.error(`[useBackend] Wails call failed for ${method}:`, err)
      throw err
    }
  }, [])

  return { call }
}
