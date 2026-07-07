// useBackend.ts — Real Wails runtime calls
import { useCallback } from 'react'

/**
 * useBackend provides a way to call Go methods bound by Wails.
 * It uses the window.go object which is populated by Wails at runtime.
 */
export function useBackend() {
  const call = useCallback(async (method: string, ...args: any[]) => {
    // Try Wails runtime
    try {
      const go = (window as any).go?.app
      if (go) {
        // Navigate nested: "SysOps.GetCPUInfo" → go.SysOps.GetCPUInfo(...)
        const parts = method.split('.')
        let target: any = go
        for (const part of parts) {
          if (!target) break
          target = target[part]
        }
        if (typeof target === 'function') {
          return await target(...args)
        }
      }

      // If we are here, Wails is not available or the method wasn't found
      console.warn(`[useBackend] Wails method not found: ${method}`)
    } catch (err) {
      console.error(`[useBackend] Wails call failed for ${method}:`, err)
      throw err
    }

    return null
  }, [])

  return { call }
}
