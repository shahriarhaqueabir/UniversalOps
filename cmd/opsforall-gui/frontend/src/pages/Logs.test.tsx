// @ts-nocheck
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
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

const mockLogs = [
  { timestamp: '2026/07/08 12:00:00', level: 'INFO', module: 'system', message: 'System started' },
  { timestamp: '2026/07/08 12:00:05', level: 'WARN', module: 'network', message: 'High latency detected' },
  { timestamp: '2026/07/08 12:00:10', level: 'ERROR', module: 'disk', message: 'Disk space low' },
]

// Mock react-query — return data based on query key
vi.mock('@tanstack/react-query', () => ({
  useQuery: (opts: { queryKey?: string[] }) => {
    const key = opts?.queryKey?.[0]
    if (key === 'logStats') {
      return {
        data: {
          totalToday: 0, totalThisHour: 0, totalLastMin: 0,
          errorCount: 0, warningCount: 0, infoCount: 0, debugCount: 0,
          topSources: [], trendingErrors: [],
        },
        isLoading: false,
        refetch: vi.fn(),
      } as const
    }
    return { data: [] as LogEntry[], isLoading: false, refetch: vi.fn() } as const
  },
  useQueryClient: () => ({ getQueryData: vi.fn(), setQueryData: vi.fn() }),
}))

import type { LogEntry } from '@/types'

// Mock react-virtual
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
  })

  it('renders without crashing', async () => {
    render(<Logs />)
    await waitFor(() => {
      expect(screen.getByText('Logs & Event Stream')).toBeInTheDocument()
    })
  })

  it('shows Overview and Live Stream tabs', async () => {
    render(<Logs />)
    await waitFor(() => {
      expect(screen.getByText('Overview')).toBeInTheDocument()
      expect(screen.getByText('Live Stream')).toBeInTheDocument()
    })
  })
})
