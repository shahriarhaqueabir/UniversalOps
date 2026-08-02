import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useOllamaStore, subscribeToOllamaProgress, _resetProgressSubscriptionForTests } from './useOllamaStore'
import { act } from 'react'

describe('useOllamaStore', () => {
  beforeEach(() => {
    useOllamaStore.setState({ status: { available: false, binary_exists: false, model: '', version: '' }, progress: null })
    _resetProgressSubscriptionForTests()
  })

  it('should have default state', () => {
    const state = useOllamaStore.getState()
    expect(state.status).toEqual({ available: false, binary_exists: false, model: '', version: '' })
    expect(state.progress).toBeNull()
  })

  it('should update status', () => {
    const newStatus = { available: true, binary_exists: true, model: 'llama3', version: '0.1.0' }
    act(() => {
      useOllamaStore.getState().setStatus(newStatus)
    })
    expect(useOllamaStore.getState().status).toEqual(newStatus)
  })

  it('should update progress', () => {
    const progress = { status: 'pulling manifest', percent: 42, total: 100, completed: 42 }
    act(() => {
      useOllamaStore.getState().setProgress(progress)
    })
    expect(useOllamaStore.getState().progress).toEqual(progress)
  })

  it('should register a single ollama:progress listener', () => {
    const eventsOn = vi.fn()
    ;(window as any).runtime = { EventsOn: eventsOn, EventsOff: vi.fn() }

    subscribeToOllamaProgress()
    subscribeToOllamaProgress()

    expect(eventsOn).toHaveBeenCalledTimes(1)
    expect(eventsOn).toHaveBeenCalledWith('ollama:progress', expect.any(Function))
  })

  it('should feed progress events into the store', () => {
    const eventsOn = vi.fn()
    ;(window as any).runtime = { EventsOn: eventsOn, EventsOff: vi.fn() }

    subscribeToOllamaProgress()
    const handler = eventsOn.mock.calls[0][1]
    const progress = { status: 'pulling', percent: 10, total: 1000, completed: 100 }
    act(() => handler(progress))

    expect(useOllamaStore.getState().progress).toEqual(progress)
  })
})
