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

vi.mock('../stores', () => ({
  useThemeStore: (selector: ((s: { theme: string; toggle: typeof mockToggle }) => unknown) | undefined) => {
    const state = { theme: 'dark' as const, toggle: mockToggle }
    return selector ? selector(state) : state
  },
  useSettingsStore: (selector: ((s: { refreshInterval: number; pingCount: number; dnsTimeout: number; companionName: string; setRefreshInterval: typeof mockSetRefreshInterval; setPingCount: typeof mockSetPingCount; setDnsTimeout: typeof mockSetDnsTimeout; setCompanionName: any }) => unknown) | undefined) => {
    const state = {
      refreshInterval: 5000,
      pingCount: 4,
      dnsTimeout: 2000,
      companionName: 'Hawk',
      setRefreshInterval: mockSetRefreshInterval,
      setPingCount: mockSetPingCount,
      setDnsTimeout: mockSetDnsTimeout,
      setCompanionName: vi.fn(),
    }
    return selector ? selector(state) : state
  },
  useAlertStore: () => ({ alerts: [], alertCount: 0 }),
  useConfigStore: (selector: any) => {
    const state = { stagedChanges: new Map(), stageChange: vi.fn(), discardAll: vi.fn(), getOriginalValue: vi.fn() }
    return selector ? selector(state) : state
  }
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
    vi.mocked(useQuery).mockImplementation(({ queryKey }: any) => {
      if (queryKey.includes('app-info')) {
        return { data: { name: 'OpsForAll', version: '1.3.0', go_version: 'go1.26', uptime: '1h' }, isLoading: false } as any
      }
      if (queryKey.includes('alert-rules')) {
        return { data: [], isLoading: false } as any
      }
      return { data: null, isLoading: false } as any
    })
  })

  it('renders control plane heading', async () => {
    renderWithProviders(<Settings />)
    await waitFor(() => {
      // Check for the main H1 heading specifically
      expect(screen.getByRole('heading', { level: 1, name: /Control Plane/i })).toBeInTheDocument()
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
