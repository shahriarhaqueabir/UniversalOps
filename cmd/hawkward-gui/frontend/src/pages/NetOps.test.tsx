import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { NetOps } from './NetOps'
import { useQuery } from '@tanstack/react-query'

// Mock useQuery
vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  QueryClient: class { clear() {} },
  QueryClientProvider: ({ children }: any) => <div>{children}</div>,
}))

// Mock Recharts
vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: any) => <div>{children}</div>,
  AreaChart: ({ children }: any) => <div>{children}</div>,
  Area: () => <div />,
  XAxis: () => <div />,
  YAxis: () => <div />,
  Tooltip: () => <div />,
  CartesianGrid: () => <div />,
}))

// Mock useBackend hook
const mockCall = vi.fn()
vi.mock('@/hooks/useBackend', () => ({
  useBackend: () => ({
    call: mockCall,
  }),
}))

describe('NetOps Page', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    // Default mock for useQuery
    ;(useQuery as any).mockImplementation(() => {
      return { data: [], isLoading: false }
    })

    mockCall.mockImplementation(async (method: string) => {
      if (method === 'NetOps.Ping') return { ip: '8.8.8.8', avg_ms: 10, ttl: 64, lost: 0 }
      if (method === 'NetOps.DNSLookup') return {
        hostname: 'google.com',
        a: ['1.2.3.4'],
        aaaa: [],
        mx: [],
        ns: [],
        txt: [],
        cname: ''
      }
      if (method === 'NetOps.Traceroute') return { target: '8.8.8.8', hops: [] }
      return []
    })
  })

  it('renders correctly', () => {
    render(<NetOps />)
    expect(screen.getByText(/NETWORK OPERATIONS/i)).toBeDefined()
    expect(screen.getByText('Probes')).toBeDefined()
  })

  it('switches tabs correctly', () => {
    render(<NetOps />)

    const resolutionTab = screen.getByText('Resolution')
    fireEvent.click(resolutionTab)

    expect(screen.getByText('Domain Resolution')).toBeDefined()
  })

  it('handles ping probe toggle', async () => {
    vi.useFakeTimers()
    render(<NetOps />)

    const startButton = screen.getByText('START PROBE')
    fireEvent.click(startButton)

    expect(screen.getByText('STOP PROBE')).toBeDefined()

    await act(async () => {
      vi.advanceTimersByTime(1100)
    })

    expect(mockCall).toHaveBeenCalledWith('NetOps.Ping', expect.any(String), expect.any(Number))
    vi.useRealTimers()
  })

  it('handles DNS resolution', async () => {
    render(<NetOps />)

    fireEvent.click(screen.getByText('Resolution'))

    const resolveButton = screen.getByText('RESOLVE')
    fireEvent.click(resolveButton)

    await waitFor(() => {
      expect(screen.getByText('google.com')).toBeDefined()
      expect(screen.getByText('1.2.3.4')).toBeDefined()
    })
  })

  it('displays interfaces data', async () => {
    const mockInterfaces = [
      { name: 'eth0', is_up: true, mac: 'AA:BB:CC', speed: '1 Gbps', ips: ['192.168.1.1'], rx_rate_bps: 0, tx_rate_bps: 0 }
    ]
    ;(useQuery as any).mockImplementation(({ queryKey }: any) => {
      if (queryKey[0] === 'netops-interfaces') return { data: mockInterfaces, isLoading: false }
      return { data: [], isLoading: false }
    })

    render(<NetOps />)
    fireEvent.click(screen.getByText('Hardware'))

    await waitFor(() => {
      expect(screen.getByText('eth0')).toBeDefined()
      expect(screen.getByText('AA:BB:CC')).toBeDefined()
    })
  })

  it('displays connection table', async () => {
    const mockConnections = [
      { protocol: 'TCP', local_addr: '127.0.0.1', local_port: 80, remote_addr: '0.0.0.0', remote_port: 0, state: 'LISTENING', pid: 1234, process_name: 'test' }
    ]
    ;(useQuery as any).mockImplementation(({ queryKey }: any) => {
      if (queryKey[0] === 'netops-connections') return { data: mockConnections, isLoading: false }
      return { data: [], isLoading: false }
    })

    render(<NetOps />)
    fireEvent.click(screen.getByText('Endpoints'))

    await waitFor(() => {
      expect(screen.getByText(/127\.0\.0\.1:80/)).toBeDefined()
      expect(screen.getByText('LISTENING')).toBeDefined()
    })
  })
})
