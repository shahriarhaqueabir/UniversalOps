import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, act } from '@testing-library/react'
import { TopBar } from './TopBar'
import { useAlertStore, useThemeStore } from '../../stores/useSettingsStore'

describe('TopBar Component', () => {
  beforeEach(() => {
    act(() => {
      useAlertStore.getState().clearAlerts()
      useThemeStore.getState().setTheme('dark')
    })
  })

  it('renders correctly for a given page', () => {
    render(<TopBar currentPage="dashboard" />)
    expect(screen.getByText('Dashboard')).toBeDefined()
    expect(screen.getByText('Hawkward')).toBeDefined()
  })

  it('displays "All Systems Nominal" when no alerts', () => {
    render(<TopBar currentPage="dashboard" />)
    expect(screen.getByText('All Systems Nominal')).toBeDefined()
  })

  it('displays alert count when alerts exist', () => {
    const alert = { id: '1', metric: 'CPU', message: 'High load' } as any
    act(() => {
      useAlertStore.getState().addAlert(alert)
    })
    render(<TopBar currentPage="dashboard" />)
    expect(screen.getByText('1 Alert Active')).toBeDefined()
    expect(screen.getByText('1')).toBeDefined() // Badge
  })

  it('toggles theme when button is clicked', () => {
    render(<TopBar currentPage="dashboard" />)
    const toggleButton = screen.getByLabelText(/Switch to light mode/i)
    fireEvent.click(toggleButton)
    expect(useThemeStore.getState().theme).toBe('light')
  })

  it('opens alert panel and shows alerts', () => {
    const alert = { id: '1', metric: 'Memory', message: 'Usage > 90%' } as any
    act(() => {
      useAlertStore.getState().addAlert(alert)
    })
    render(<TopBar currentPage="dashboard" />)

    const bellButton = screen.getByLabelText(/Notifications/i)
    fireEvent.click(bellButton)

    expect(screen.getByText('Alerts')).toBeDefined()
    expect(screen.getByText('Memory')).toBeDefined()
    expect(screen.getByText('Usage > 90%')).toBeDefined()
  })

  it('clears alerts when Clear All is clicked', () => {
    const alert = { id: '1', metric: 'Disk', message: 'Full' } as any
    act(() => {
      useAlertStore.getState().addAlert(alert)
    })
    render(<TopBar currentPage="dashboard" />)

    // Open panel
    fireEvent.click(screen.getByLabelText(/Notifications/i))

    // Click clear
    const clearButton = screen.getByText(/Clear All/i)
    fireEvent.click(clearButton)

    expect(useAlertStore.getState().alertCount).toBe(0)
    expect(screen.queryByText('Disk')).toBeNull()
  })
})
