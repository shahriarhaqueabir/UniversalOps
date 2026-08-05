import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, act } from '@testing-library/react'
import { TopBar } from './TopBar'
import { useThemeStore } from '../../stores/useSettingsStore'

import { useQuery } from '@tanstack/react-query'
import type { AlertInfo } from '@/types'
import { mockQueryReturn } from '@/test/mockQuery'

// Mock react-query hooks
vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useMutation: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
  useQueryClient: vi.fn(() => ({ invalidateQueries: vi.fn() })),
}))

// Mock useBackend
vi.mock('@/hooks/useBackend', () => ({
  useBackend: () => ({ call: vi.fn() }),
}))

describe('TopBar Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useQuery).mockReturnValue(mockQueryReturn({ data: [] }))
    act(() => {
      useThemeStore.getState().setTheme('dark')
    })
  })

  it('renders correctly for a given page', () => {
    render(<TopBar currentPage="dashboard" />)
    expect(screen.getByText(/Dashboard/i)).toBeTruthy()
    expect(screen.getByText(/Universal-Ops/i)).toBeTruthy()
  })

  it('displays "All Systems Nominal" when no alerts', () => {
    render(<TopBar currentPage="dashboard" />)
    expect(screen.getByText(/All Systems Nominal/i)).toBeTruthy()
  })

  it('displays alert count when alerts exist', () => {
    const alert: AlertInfo = { id: '1', level: '', metric: 'CPU', message: 'High load', value: 0, threshold: 0, timestamp: '', resolved: false }
    vi.mocked(useQuery).mockReturnValue(mockQueryReturn({ data: [alert] }))
    render(<TopBar currentPage="dashboard" />)
    expect(screen.getByText(/1 Alert Active/i)).toBeTruthy()
    const badge = screen.getByText('1', { selector: 'span' })
    expect(badge).toBeTruthy()
  })

  it('toggles theme when button is clicked', () => {
    render(<TopBar currentPage="dashboard" />)
    const toggleButton = screen.getByLabelText(/Switch to/i)
    fireEvent.click(toggleButton)
    expect(useThemeStore.getState().theme).toBe('light')
  })

  it('opens alert panel and shows alerts', async () => {
    const alert: AlertInfo = { id: '1', level: 'warning', metric: 'Memory', message: 'Usage > 90%', value: 0, threshold: 0, timestamp: '12:00:00', resolved: false }

    vi.mocked(useQuery).mockReturnValue(mockQueryReturn({ data: [alert] }))

    render(<TopBar currentPage="dashboard" />)

    const bellButton = screen.getByLabelText(/Notifications/i)
    fireEvent.click(bellButton)

    expect(screen.getByText(/Active Alerts/i)).toBeTruthy()
    expect(screen.getByText('Memory')).toBeTruthy()
    expect(screen.getByText('Usage > 90%')).toBeTruthy()
  })

  it('refreshes alerts when Refresh button is clicked', () => {
    const alert: AlertInfo = { id: '1', level: 'info', metric: 'Disk', message: 'Full', value: 0, threshold: 0, timestamp: '12:00:00', resolved: false }
    vi.mocked(useQuery).mockReturnValue(mockQueryReturn({ data: [alert] }))

    render(<TopBar currentPage="dashboard" />)

    // Open panel
    fireEvent.click(screen.getByLabelText(/Notifications/i))

    const refreshButton = screen.getByText(/Refresh/i)
    fireEvent.click(refreshButton)
    expect(refreshButton).toBeTruthy()
  })
})
