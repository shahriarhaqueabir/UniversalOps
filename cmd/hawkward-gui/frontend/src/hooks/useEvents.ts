import { useEffect, useRef } from 'react'

/**
 * useEvents subscribes to Wails events emitted from the Go backend.
 * eslint-disable-next-line @typescript-eslint/no-explicit-any -- event payloads are dynamic
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails event payload type
type EventPayload = any

// eslint-disable-next-line @typescript-eslint/no-explicit-any -- Wails runtime type
type Runtime = any

export function useEvents(eventName: string, handler: (data: EventPayload) => void) {
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    // Wails v2 runtime events
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- window.runtime is injected by Wails at runtime
    const runtime: Runtime = (window as any).runtime
    if (runtime?.EventsOn) {
      const wrappedHandler = (data: EventPayload) => {
        handlerRef.current(data)
      }
      runtime.EventsOn(eventName, wrappedHandler)
      return () => {
        runtime.EventsOff(eventName, wrappedHandler)
      }
    }
  }, [eventName])
}
