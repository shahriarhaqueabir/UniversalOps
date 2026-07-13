// @ts-nocheck
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import type { ReactNode } from 'react'
import { NetOps } from './NetOps'
import { useQuery } from '@tanstack/react-query'

// Mock useQuery
vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  QueryClient: class { clear() {} },
  QueryClientProvider: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}))

// Mock Recharts
vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  AreaChart: ({ children }: { children: ReactNode }) => <div>{children}</div>,
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
    vi.mocked(useQuery).mockImplementation(() => {
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
    expect(screen.getByText(/NETWORK OPERATIONS/i)).toBeTruthy()
    expect(screen.getByRole('tab', { name: /Probes/i })).toBeTruthy()
  })

  it('switches tabs correctly', () => {
    render(<NetOps />)

    const resolutionTab = screen.getByRole('tab', { name: /Resolution/i })
    fireEvent.click(resolutionTab)

    expect(screen.getByText('Domain Resolution')).toBeTruthy()
  })

  it('handles ping probe toggle', async () => {
    vi.useFakeTimers()
    render(<NetOps />)

    // Select the Probes tab first
    const probesTab = screen.getByRole('tab', { name: /Probes/i })
    fireEvent.click(probesTab)

    const startButton = screen.getByRole('button', { name: /START PROBE/i })
    fireEvent.click(startButton)

    expect(screen.getByText('STOP PROBE')).toBeTruthy()

    await act(async () => {
      vi.advanceTimersByTime(1100)
    })

    expect(mockCall).toHaveBeenCalledWith('NetOps.Ping', expect.any(String), expect.any(Number))
    vi.useRealTimers()
  })

  it('handles DNS resolution', async () => {
    render(<NetOps />)

    fireEvent.click(screen.getByRole('tab', { name: /Resolution/i }))

    const resolveButton = screen.getByRole('button', { name: /RESOLVE/i })
    fireEvent.click(resolveButton)

    await waitFor(() => {
      // Look for the result header specifically if possible, or just the text
      expect(screen.getAllByText('google.com').length).toBeGreaterThan(0)
      expect(screen.getByText('1.2.3.4')).toBeTruthy()
    })
  })

  it('displays interfaces data', async () => {
    const mockInterfaces = [
      { name: 'eth0', is_up: true, mac: 'AA:BB:CC', speed: '1 Gbps', ips: ['192.168.1.1'], rx_rate_bps: 0, tx_rate_bps: 0 }
    ]
    vi.mocked(useQuery).mockImplementation(({ queryKey }) => {
      if (queryKey[0] === 'netops-interfaces') return { data: mockInterfaces, isLoading: false }
      if (queryKey[0] === 'connectivity-interfaces') return { data: mockInterfaces, isLoading: false }
      return { data: [], isLoading: false }
    })

    render(<NetOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Hardware/i }))

    await waitFor(() => {
      expect(screen.getByText('eth0')).toBeTruthy()
      expect(screen.getByText('AA:BB:CC')).toBeTruthy()
    })
  })

  it('displays connection table', async () => {
    const mockConnections = [
      { protocol: 'TCP', local_addr: '127.0.0.1', local_port: 80, remote_addr: '0.0.0.0', remote_port: 0, state: 'LISTENING', pid: 1234, process_name: 'test' }
    ]
    vi.mocked(useQuery).mockImplementation(({ queryKey }) => {
      if (queryKey[0] === 'netops-connections') return { data: mockConnections, isLoading: false }
      return { data: [], isLoading: false }
    })

    render(<NetOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Endpoints/i }))

    await waitFor(() => {
      // Use a more flexible matcher for the address
      expect(screen.getByText(/127\.0\.0\.1:80/)).toBeTruthy()
      expect(screen.getByText('LISTENING')).toBeTruthy()
    })
  })
})
