// useBackend.ts — Real Wails runtime calls
import { useCallback } from 'react'

/**
 * useBackend provides a way to call Go methods bound by Wails.
 * It uses the window.go object which is populated by Wails at runtime.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails bridge requires dynamic types
type Args = any[]

// eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails bridge type
function getGo(): any {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- window.go is injected by Wails at runtime
  return (window as any).go?.app
}

export function useBackend() {
  const call = useCallback(async (method: string, ...args: Args) => {
    // Try Wails runtime
    try {
      const go = getGo()
      if (go) {
        // Navigate nested: "SysOps.GetCPUInfo" → go.SysOps.GetCPUInfo(...)
        const parts = method.split('.')
        // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails bridge navigation
        let target: any = go
        for (const part of parts) {
          if (!target) break
          target = target[part]
        }
        if (typeof target === 'function') {
          return await target(...args)
        }
      }

      console.warn(`[useBackend] Wails method not found: ${method}`)
    } catch (err) {
      console.error(`[useBackend] Wails call failed for ${method}:`, err)
      throw err
    }

    return null
  }, [])

  return { call }
}
