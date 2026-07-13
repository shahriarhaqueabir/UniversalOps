// @ts-nocheck
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query'
import { Settings } from './Settings'

// Mock react-query
vi.mock('@tanstack/react-query', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-query')>()
  return {
    ...actual,
    useQuery: vi.fn(),
  }
})

// Mock the zustand stores
const mockToggle = vi.fn()
const mockSetRefreshInterval = vi.fn()
const mockSetPingCount = vi.fn()
const mockSetDnsTimeout = vi.fn()

vi.mock('../stores/useSettingsStore', () => ({
  useThemeStore: (selector: ((s: { theme: string; toggle: typeof mockToggle }) => unknown) | undefined) => {
    const state = { theme: 'dark' as const, toggle: mockToggle }
    return selector ? selector(state) : state
  },
  useSettingsStore: (selector: ((s: { refreshInterval: number; pingCount: number; dnsTimeout: number; setRefreshInterval: typeof mockSetRefreshInterval; setPingCount: typeof mockSetPingCount; setDnsTimeout: typeof mockSetDnsTimeout }) => unknown) | undefined) => {
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
  useBackend: () => ({ call: vi.fn().mockResolvedValue({ name: 'OpsForAll'
, version: '1.0.0', goVersion: '1.26', uptime: '2h' }) }),
}))

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

function renderWithProviders(ui: React.ReactElement) {
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>)
}

describe('Settings Page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useQuery).mockImplementation(({ queryKey }) => {
      if (queryKey.includes('app-info')) {
        return { data: { name: 'OpsForAll', version: '1.3.0', go_version: 'go1.26', uptime: '1h' }, isLoading: false }
      }
      if (queryKey.includes('alert-rules')) {
        return { data: [], isLoading: false }
      }
      return { data: null, isLoading: false }
    })
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
