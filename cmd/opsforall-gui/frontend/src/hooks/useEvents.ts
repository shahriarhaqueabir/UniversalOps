import { useEffect, useRef } from 'react'

interface WailsRuntime {
  EventsOn(event: string, handler: (data: unknown) => void): void
  EventsOff(event: string, handler: (data: unknown) => void): void
}

export function useEvents(eventName: string, handler: (data: unknown) => void) {
  const handlerRef = useRef(handler)

  useEffect(() => {
    handlerRef.current = handler
  }, [handler])

  useEffect(() => {
    const runtime = (window as { runtime?: WailsRuntime }).runtime
    if (runtime?.EventsOn) {
      const wrappedHandler = (data: unknown) => {
        handlerRef.current(data)
      }
      runtime.EventsOn(eventName, wrappedHandler)
      return () => {
        runtime.EventsOff(eventName, wrappedHandler)
      }
    }
  }, [eventName])
}
