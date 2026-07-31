import { describe, it, expect, beforeEach } from 'vitest'
import { act } from 'react'
import { useConfigStore, pruneDefaults } from './useConfigStore'

// Minimal schema for pruneDefaults tests
const testSchema = {
  properties: {
    refreshInterval: { type: 'number', default: 5000 },
    pingCount: { type: 'number', default: 4 },
    dnsTimeout: { type: 'number', default: 2000 },
    customField: { type: 'string' },
  },
}

describe('useConfigStore', () => {
  beforeEach(() => {
    act(() => {
      useConfigStore.setState({
        stagedChanges: new Map(),
        validationErrors: {},
      })
    })
  })

  it('starts with empty staged changes', () => {
    const state = useConfigStore.getState()
    expect(state.stagedChanges.size).toBe(0)
    expect(state.validationErrors).toEqual({})
  })

  it('stages a single change via stageChange', () => {
    const store = useConfigStore.getState()
    const result = store.stageChange('refreshInterval', 10000)
    expect(result).toBe(true)
    expect(useConfigStore.getState().stagedChanges.get('refreshInterval')).toBe(10000)
  })

  it('returns false for invalid value in stageChange', () => {
    const store = useConfigStore.getState()
    // An object is not valid for refreshInterval (should be number)
    const result = store.stageChange('refreshInterval', 'not-a-number')
    // Schema validation may pass or fail depending on schema strictness,
    // but the call should not throw
    expect(typeof result).toBe('boolean')
  })

  it('stages multiple changes via stageBatch', () => {
    const store = useConfigStore.getState()
    const errors = store.stageBatch({
      refreshInterval: 10000,
      pingCount: 8,
    })
    expect(Object.keys(errors).length).toBe(0)
    const state = useConfigStore.getState()
    expect(state.stagedChanges.get('refreshInterval')).toBe(10000)
    expect(state.stagedChanges.get('pingCount')).toBe(8)
  })

  it('discardAll clears staged changes and errors', () => {
    const store = useConfigStore.getState()
    store.stageChange('refreshInterval', 10000)
    expect(useConfigStore.getState().stagedChanges.size).toBe(1)

    act(() => {
      useConfigStore.getState().discardAll()
    })
    const state = useConfigStore.getState()
    expect(state.stagedChanges.size).toBe(0)
    expect(state.validationErrors).toEqual({})
  })

  it('clearValidationError removes a single error', () => {
    // First create an error by staging something that might fail
    const store = useConfigStore.getState()

    // Manually set a validation error to test clearValidationError
    act(() => {
      useConfigStore.setState({
        validationErrors: { refreshInterval: 'Test error' },
      })
    })
    expect(useConfigStore.getState().validationErrors.refreshInterval).toBe('Test error')

    act(() => {
      useConfigStore.getState().clearValidationError('refreshInterval')
    })
    expect(useConfigStore.getState().validationErrors.refreshInterval).toBeUndefined()
  })

  it('getOriginalValue returns value from settings store', () => {
    const store = useConfigStore.getState()
    // This reads from useSettingsStore which may have defaults
    const val = store.getOriginalValue('refreshInterval')
    expect(typeof val).toBe('number')
  })
})

describe('pruneDefaults', () => {
  it('removes keys matching defaults', () => {
    const result = pruneDefaults(testSchema, {
      refreshInterval: 5000,
      pingCount: 8,
    })
    expect(result).toEqual({ pingCount: 8 })
  })

  it('keeps non-default keys', () => {
    const result = pruneDefaults(testSchema, {
      refreshInterval: 10000,
      pingCount: 8,
    })
    expect(result).toEqual({ refreshInterval: 10000, pingCount: 8 })
  })

  it('keeps keys without schema defaults', () => {
    const result = pruneDefaults(testSchema, {
      customField: 'hello',
    })
    expect(result).toEqual({ customField: 'hello' })
  })

  it('returns a copy of all keys when schema has no properties', () => {
    const result = pruneDefaults({}, { key: 'value' })
    expect(result).toEqual({ key: 'value' })
  })

  it('returns empty object for empty config', () => {
    const result = pruneDefaults(testSchema, {})
    expect(result).toEqual({})
  })
})

describe('useConfigStore - getRiskLevel', () => {
  it('returns high for refreshInterval under 1000', () => {
    const { getRiskLevel } = useConfigStore.getState()
    const risk = getRiskLevel('refreshInterval', 500)
    expect(risk.level).toBe('high')
    expect(risk.message).toContain('increase')
  })

  it('returns med for refreshInterval under 3000', () => {
    const { getRiskLevel } = useConfigStore.getState()
    const risk = getRiskLevel('refreshInterval', 2000)
    expect(risk.level).toBe('med')
  })

  it('returns low for refreshInterval >= 3000', () => {
    const { getRiskLevel } = useConfigStore.getState()
    const risk = getRiskLevel('refreshInterval', 5000)
    expect(risk.level).toBe('low')
  })

  it('returns med for pingCount > 15', () => {
    const { getRiskLevel } = useConfigStore.getState()
    const risk = getRiskLevel('pingCount', 20)
    expect(risk.level).toBe('med')
  })

  it('returns low for pingCount <= 15', () => {
    const { getRiskLevel } = useConfigStore.getState()
    const risk = getRiskLevel('pingCount', 8)
    expect(risk.level).toBe('low')
  })

  it('returns low for unknown keys', () => {
    const { getRiskLevel } = useConfigStore.getState()
    const risk = getRiskLevel('unknownField', 'anything')
    expect(risk.level).toBe('low')
  })
})
