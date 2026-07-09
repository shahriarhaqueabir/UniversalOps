import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Settings } from './Settings'

// Mock the zustand stores
const mockToggle = vi.fn()
const mockSetRefreshInterval = vi.fn()
const mockSetPingCount = vi.fn()
const mockSetDnsTimeout = vi.fn()

vi.mock('../stores/useSettingsStore', () => ({
  useThemeStore: (selector: any) => {
    const state = { theme: 'dark', toggle: mockToggle }
    return selector ? selector(state) : state
  },
  useSettingsStore: (selector: any) => {
    const state = {
      refreshInterval: 5000,
      pingCount: 4,
      dnsTimeout: 2000,
      setRefreshInterval: mockSetRefreshInterval,
      setPingCount: mockSetPingCount,
      setDnsTimeout: mockSetDnsTimeout,
    }
    return selector ? selector(state) : state
  },
  useAlertStore: () => ({ alerts: [], alertCount: 0 }),
}))

vi.mock('../hooks/useBackend', () => ({
  useBackend: () => ({ call: vi.fn().mockResolvedValue({ name: 'Hawkward', version: '1.0.0', goVersion: '1.26', uptime: '2h' }) }),
}))

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function renderWithProviders(ui: React.ReactElement) {
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>)
}

describe('Settings Page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders settings heading', async () => {
    renderWithProviders(<Settings />)
    await waitFor(() => {
      expect(screen.getByText('Settings')).toBeInTheDocument()
    })
  })

  it('renders appearance section', async () => {
    renderWithProviders(<Settings />)
    await waitFor(() => {
      expect(screen.getByText('Appearance')).toBeInTheDocument()
    })
  })

  it('theme toggle triggers toggle', async () => {
    renderWithProviders(<Settings />)
    await waitFor(() => {
      expect(screen.getByText('Light')).toBeInTheDocument()
    })
    const lightBtn = screen.getByText('Light')
    fireEvent.click(lightBtn)
    expect(mockToggle).toHaveBeenCalledTimes(1)
  })
})
