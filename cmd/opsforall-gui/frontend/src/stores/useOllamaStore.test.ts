import { describe, it, expect } from 'vitest'
import { useOllamaStore } from './useOllamaStore'
import { act } from 'react'

describe('useOllamaStore', () => {
  it('should have default state', () => {
    const state = useOllamaStore.getState()
    expect(state.status).toEqual({ available: false, model: '', version: '' })
  })

  it('should update status', () => {
    const newStatus = { available: true, model: 'llama3', version: '0.1.0' }
    act(() => {
      useOllamaStore.getState().setStatus(newStatus)
    })
    expect(useOllamaStore.getState().status).toEqual(newStatus)
  })
})
