import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Logs } from './Logs'

// Ensure localStorage is available for zustand stores
const localStorageMock = {
  getItem: vi.fn(() => null),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
  get length() { return 0 },
  key: vi.fn(() => null),
}
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
  configurable: true,
})

import type { LogEntry } from '@/types'

const mockLogs = [
  { timestamp: '2026/07/08 12:00:00', level: 'INFO', module: 'system', message: 'System started' },
  { timestamp: '2026/07/08 12:00:05', level: 'WARN', module: 'network', message: 'High latency detected' },
  { timestamp: '2026/07/08 12:00:10', level: 'ERROR', module: 'disk', message: 'Disk space low' },
]

// Hoisted mutable ref so individual tests can override timeline data.
// Data is defined inside hoisted to avoid hoisting TDZ issues.
const timelineDataRef = vi.hoisted(() => ({
  current: [
    { timestamp: '2026-07-25T10:00:00Z', total: 5, errors: 1, warnings: 1, info: 3 },
    { timestamp: '2026-07-25T11:00:00Z', total: 8, errors: 2, warnings: 3, info: 3 },
    { timestamp: '2026-07-25T12:00:00Z', total: 4, errors: 0, warnings: 1, info: 3 },
  ],
}))

// Mock react-query — return data based on query key
vi.mock('@tanstack/react-query', () => ({
  useQuery: (opts: { queryKey?: string[] }) => {
    const key = opts?.queryKey?.[0]
    const subKey = opts?.queryKey?.[1]
    if (key === 'logStats') {
      return {
        data: {
          totalToday: 42, totalThisHour: 5, totalLastMin: 1,
          errorCount: 3, warningCount: 7, infoCount: 32, debugCount: 0,
          topSources: [{ source: 'system', count: 20 }], trendingErrors: [],
        },
        isLoading: false,
        refetch: vi.fn(),
      } as const
    }
    if (key === 'logs' && subKey === 'timeline') {
      return { data: timelineDataRef.current, isLoading: false, refetch: vi.fn() } as const
    }
    return { data: [] as LogEntry[], isLoading: false, refetch: vi.fn() } as const
  },
  useQueryClient: () => ({ getQueryData: vi.fn(), setQueryData: vi.fn() }),
}))

// Mock react-virtual — return empty virtual items
vi.mock('@tanstack/react-virtual', () => ({
  useVirtualizer: () => ({
    getVirtualItems: () => [],
    getTotalSize: () => 0,
  }),
}))

vi.mock('../hooks/useBackend', () => ({
  useBackend: () => ({
    call: vi.fn().mockResolvedValue(mockLogs),
  }),
}))

describe('Logs Page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset timeline data to default for each test
    timelineDataRef.current = [
      { timestamp: '2026-07-25T10:00:00Z', total: 5, errors: 1, warnings: 1, info: 3 },
      { timestamp: '2026-07-25T11:00:00Z', total: 8, errors: 2, warnings: 3, info: 3 },
      { timestamp: '2026-07-25T12:00:00Z', total: 4, errors: 0, warnings: 1, info: 3 },
    ]
  })

  it('renders without crashing', async () => {
    render(<Logs />)
    await waitFor(() => {
      expect(screen.getByText('Logs & Event Stream')).toBeInTheDocument()
    })
  })

  it('shows Overview, Live Stream, and Audit tabs', async () => {
    render(<Logs />)
    await waitFor(() => {
      expect(screen.getByText('Overview')).toBeInTheDocument()
      expect(screen.getByText('Live Stream')).toBeInTheDocument()
      expect(screen.getByText('Audit')).toBeInTheDocument()
    })
  })

  it('renders timeline chart with data instead of empty state', async () => {
    render(<Logs />)
    // The chart should render instead of the "No timeline data available" empty state
    await waitFor(() => {
      expect(screen.queryByText('No timeline data available.')).not.toBeInTheDocument()
    })
  })

  it('shows empty timeline state when data is empty', async () => {
    timelineDataRef.current = []
    render(<Logs />)
    await waitFor(() => {
      expect(screen.getByText('No timeline data available.')).toBeInTheDocument()
    })
  })

  it('shows no AI summary available when no summary data', async () => {
    render(<Logs />)
    await waitFor(() => {
      expect(screen.getByText('No summary available.')).toBeInTheDocument()
    })
  })

  it('switches tabs between Overview, Live Stream, and Audit', async () => {
    const user = userEvent.setup()
    render(<Logs />)

    // Start on Overview tab
    await waitFor(() => {
      expect(screen.getByText('Log Volume Timeline (24h)')).toBeInTheDocument()
    })

    // Click Live Stream tab
    await user.click(screen.getByText('Live Stream'))
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Search live stream...')).toBeInTheDocument()
    })

    // Click Audit tab
    await user.click(screen.getByText('Audit'))
    await waitFor(() => {
      expect(screen.getByText('Total Logs')).toBeInTheDocument()
    })

    // Click back to Overview
    await user.click(screen.getByText('Overview'))
    await waitFor(() => {
      expect(screen.getByText('Log Volume Timeline (24h)')).toBeInTheDocument()
    })
  })
})
