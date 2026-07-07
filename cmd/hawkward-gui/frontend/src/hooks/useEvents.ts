import { useEffect, useRef } from 'react'

/**
 * useEvents subscribes to Wails events emitted from the Go backend.
 */
export function useEvents(eventName: string, handler: (data: any) => void) {
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    // Wails v2 runtime events
    const runtime = (window as any).runtime
    if (runtime?.EventsOn) {
      const wrappedHandler = (data: any) => {
        handlerRef.current(data)
      }
      runtime.EventsOn(eventName, wrappedHandler)
      return () => {
        runtime.EventsOff(eventName, wrappedHandler)
      }
    }
  }, [eventName])
}
