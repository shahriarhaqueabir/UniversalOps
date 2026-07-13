// @ts-nocheck
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useEvents } from './useEvents'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('useEvents', () => {
  it('subscribes to event on mount', () => {
    const handler = vi.fn()
    renderHook(() => useEvents('test-event', handler))

    expect((window as any).runtime.EventsOn).toHaveBeenCalledWith(
      'test-event',
      expect.any(Function),
    )
  })

  it('unsubscribes on unmount', () => {
    const handler = vi.fn()
    const { unmount } = renderHook(() => useEvents('test-event', handler))

    unmount()

    expect((window as any).runtime.EventsOff).toHaveBeenCalledWith(
      'test-event',
      expect.any(Function),
    )
  })

  it('calls handler when event fires', () => {
    const handler = vi.fn()
    renderHook(() => useEvents('data-update', handler))

    const registeredHandler = vi.mocked((window as any).runtime.EventsOn).mock.lastCall?.[1]

    act(() => { registeredHandler?.({ value: 42 }) })

    expect(handler).toHaveBeenCalledWith({ value: 42 })
  })

  it('re-subscribes when eventName changes', () => {
    const handler = vi.fn()
    const { rerender } = renderHook(
      ({ name }) => useEvents(name, handler),
      { initialProps: { name: 'event-a' } },
    )

    rerender({ name: 'event-b' })

    expect((window as any).runtime.EventsOff).toHaveBeenCalledWith('event-a', expect.any(Function))
    expect((window as any).runtime.EventsOn).toHaveBeenCalledWith('event-b', expect.any(Function))
  })

  it('updates handler ref without re-subscribing', () => {
    const handler1 = vi.fn()
    const { rerender } = renderHook(
      ({ h }) => useEvents('test', h),
      { initialProps: { h: handler1 } },
    )

    const handler2 = vi.fn()
    rerender({ h: handler2 })

    const registeredHandler = vi.mocked((window as any).runtime.EventsOn).mock.lastCall?.[1]

    act(() => { registeredHandler?.('payload') })

    expect(handler2).toHaveBeenCalledWith('payload')
    expect(handler1).not.toHaveBeenCalled()
  })
})
