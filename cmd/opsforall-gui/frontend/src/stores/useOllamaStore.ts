import { create } from 'zustand'
import { getRuntime } from '@/hooks/useBackend'
import type { OllamaProgress, OllamaStatus } from '@/types'

interface OllamaState {
  status: OllamaStatus
  progress: OllamaProgress | null
  setStatus: (status: OllamaStatus) => void
  setProgress: (progress: OllamaProgress | null) => void
}

export const useOllamaStore = create<OllamaState>((set) => ({
  status: { available: false, binary_exists: false, model: '', version: '' },
  progress: null,
  setStatus: (status) => set({ status }),
  setProgress: (progress) => set({ progress }),
}))

// ── Singleton progress subscription ─────────────────────────────────────────
// Wails v2's EventsOff(eventName) removes ALL listeners for that event (it
// accepts event names, not handlers), so multiple components subscribing to
// 'ollama:progress' can clobber each other's listeners on unmount. To avoid
// that, we maintain exactly ONE subscription here and let every component read
// progress from the store. The subscription lives for the app's lifetime.
let progressSubscribed = false

/**
 * Ensures the single 'ollama:progress' listener is registered. Safe to call
 * from multiple components — only the first call actually subscribes. Returns
 * a no-op cleanup (we never unsubscribe, since Wails cannot remove a single
 * handler and this is the only subscriber).
 */
export function subscribeToOllamaProgress(): () => void {
  if (progressSubscribed) return () => {}
  progressSubscribed = true
  const runtime = getRuntime()
  if (!runtime?.EventsOn) return () => {}
  const handler = (p: unknown) => useOllamaStore.getState().setProgress(p as OllamaProgress)
  runtime.EventsOn('ollama:progress', handler)
  return () => {}
}

// Test-only hook to reset the singleton flag between test cases.
export function _resetProgressSubscriptionForTests() {
  progressSubscribed = false
}
