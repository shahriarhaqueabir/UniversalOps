import { describe, it, expect, beforeEach } from 'vitest'
import { useSettingsStore, useAlertStore, useThemeStore } from './useSettingsStore'
import { act } from 'react'

describe('useSettingsStore', () => {
  beforeEach(() => {
    localStorage.clear()
    // Reset store state to defaults
    act(() => {
      useSettingsStore.setState({
        refreshInterval: 5000,
        pingCount: 4,
        dnsTimeout: 2000,
      })
    })
  })

  it('should have default values', () => {
    const state = useSettingsStore.getState()
    expect(state.refreshInterval).toBe(5000)
    expect(state.pingCount).toBe(4)
    expect(state.dnsTimeout).toBe(2000)
  })

  it('should update refreshInterval and persist to localStorage', () => {
    act(() => {
      useSettingsStore.getState().setRefreshInterval(10000)
    })
    expect(useSettingsStore.getState().refreshInterval).toBe(10000)
    expect(localStorage.getItem('hawkward_refreshInterval')).toBe('10000')
  })

  it('should update pingCount and persist to localStorage', () => {
    act(() => {
      useSettingsStore.getState().setPingCount(8)
    })
    expect(useSettingsStore.getState().pingCount).toBe(8)
    expect(localStorage.getItem('hawkward_pingCount')).toBe('8')
  })

  it('should update dnsTimeout and persist to localStorage', () => {
    act(() => {
      useSettingsStore.getState().setDnsTimeout(5000)
    })
    expect(useSettingsStore.getState().dnsTimeout).toBe(5000)
    expect(localStorage.getItem('hawkward_dnsTimeout')).toBe('5000')
  })
})

describe('useAlertStore', () => {
  beforeEach(() => {
    act(() => {
      useAlertStore.getState().clearAlerts()
    })
  })

  it('should start with empty alerts', () => {
    const state = useAlertStore.getState()
    expect(state.alerts).toEqual([])
    expect(state.alertCount).toBe(0)
  })

  it('should add an alert', () => {
    const alert = { id: '1', timestamp: 'now', category: 'SYSTEM', level: 'info', title: 'Test', detail: '' } as any
    act(() => {
      useAlertStore.getState().addAlert(alert)
    })
    const state = useAlertStore.getState()
    expect(state.alerts).toHaveLength(1)
    expect(state.alerts[0]).toEqual(alert)
    expect(state.alertCount).toBe(1)
  })

  it('should set multiple alerts', () => {
    const alerts = [
      { id: '1', title: 'A' },
      { id: '2', title: 'B' }
    ] as any[]
    act(() => {
      useAlertStore.getState().setAlerts(alerts)
    })
    const state = useAlertStore.getState()
    expect(state.alerts).toHaveLength(2)
    expect(state.alertCount).toBe(2)
  })

  it('should limit alerts to 100', () => {
    act(() => {
      for (let i = 0; i < 110; i++) {
        useAlertStore.getState().addAlert({ id: `${i}`, title: `Test ${i}` } as any)
      }
    })
    const state = useAlertStore.getState()
    expect(state.alerts).toHaveLength(100)
    expect(state.alertCount).toBe(110) // alertCount is a running total in current implementation
  })
})

describe('useThemeStore', () => {
  beforeEach(() => {
    localStorage.clear()
    act(() => {
      useThemeStore.getState().setTheme('dark')
    })
  })

  it('should toggle theme', () => {
    act(() => {
      useThemeStore.getState().toggle()
    })
    expect(useThemeStore.getState().theme).toBe('light')
    expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    expect(localStorage.getItem('hawkward-theme')).toBe('light')

    act(() => {
      useThemeStore.getState().toggle()
    })
    expect(useThemeStore.getState().theme).toBe('dark')
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
  })

  it('should set specific theme', () => {
    act(() => {
      useThemeStore.getState().setTheme('light')
    })
    expect(useThemeStore.getState().theme).toBe('light')
  })
})
