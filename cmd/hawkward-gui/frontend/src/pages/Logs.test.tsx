import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { Logs } from './Logs'

const mockLogs = [
  { timestamp: '2026/07/08 12:00:00', level: 'INFO', module: 'system', message: 'System started' },
  { timestamp: '2026/07/08 12:00:05', level: 'WARN', module: 'network', message: 'High latency detected' },
  { timestamp: '2026/07/08 12:00:10', level: 'ERROR', module: 'disk', message: 'Disk space low' },
]

// Mock react-query
vi.mock('@tanstack/react-query', () => ({
  useQuery: ({ queryFn }: { queryFn: () => Promise<any> }) => {
    const data = queryFn ? queryFn() : []
    return { data: Promise.resolve(data) as any, isLoading: false, refetch: vi.fn() }
  },
  useQueryClient: () => ({ getQueryData: vi.fn(), setQueryData: vi.fn() }),
}))

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
      expect(screen.getByText('Live Event Aggregator')).toBeInTheDocument()
    })
  })

  it('shows level filter badges', async () => {
    render(<Logs />)
    await waitFor(() => {
      expect(screen.getByText('INFO')).toBeInTheDocument()
      expect(screen.getByText('WARN')).toBeInTheDocument()
      expect(screen.getByText('ERROR')).toBeInTheDocument()
      expect(screen.getByText('DEBUG')).toBeInTheDocument()
    })
  })
})
