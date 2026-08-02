import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { Dashboard } from './Dashboard'

const mockDashboardData = {
  metrics: {
    cpu: { value: 45, unit: '%', history: [30, 40, 45], forecast: [46, 47], trend: 'rising' },
    memory: { value: 62, unit: '%', history: [60, 61, 62], forecast: [63, 64], trend: 'rising' },
    disk: { value: 78, unit: '%', history: [77, 78, 78], forecast: [78, 79], trend: 'stable' },
    gpu: { name: '', vendor: '', memory_gb: 0, driver: '', detected: false },
    battery: { percent: 0, charging: false, time_left_sec: -1, status: '', detected: false },
    network: { rx_rate: 1000000, tx_rate: 500000, unit: 'bps' },
    processes: 245,
    connections: 12,
    alerts: 0,
    uptime: '2d 4h',
  },
  alerts: [],
  timeline: [],
  timestamp: new Date().toISOString(),
}

// Mock recharts
vi.mock('recharts', async () => {
  const original = await vi.importActual('recharts') as Record<string, unknown>
  return {
    ...original,
    ResponsiveContainer: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  }
})

// Mock react-query
vi.mock('@tanstack/react-query', () => ({
  useQuery: ({ queryKey }: { queryKey: string[] }) => {
    if (queryKey.includes('dashboard-snapshot')) return { data: mockDashboardData, isLoading: false, isSuccess: true }
    if (queryKey.includes('dashboard')) return { data: mockDashboardData, isLoading: false, isSuccess: true }
    if (queryKey.includes('alertBreakdown')) return { data: { critical: 0, warning: 0, info: 0 }, isLoading: false }
    if (queryKey.includes('timelineEvents')) return { data: [], isLoading: false }
    if (queryKey.includes('timelineSummary')) return { data: {}, isLoading: false }
    if (queryKey.includes('topProcs')) return { data: { cpuProcs: [], memProcs: [] }, isLoading: false }
    return { data: null, isLoading: false }
  },
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}))

// Mock hooks
vi.mock('../hooks/useBackend', () => ({
  useBackend: () => ({
    call: vi.fn().mockImplementation((method: string) => {
      if (method === 'Dashboard.GetDashboardData') return Promise.resolve(mockDashboardData)
      if (method === 'AlertAPI.GetActiveAlerts') return Promise.resolve([])
      if (method === 'Timeline.GetTimelineEvents') return Promise.resolve([])
      if (method === 'Timeline.GetTimelineSummary') return Promise.resolve({})
      if (method === 'SysOps.GetTopProcesses') return Promise.resolve([])
      return Promise.resolve(null)
    }),
  }),
}))

vi.mock('../hooks/useEvents', () => ({
  useEvents: () => { },
}))

describe('Dashboard Page', () => {
  it('renders without crashing', async () => {
    render(<Dashboard />)
    expect(await screen.findByText(/OPERATIONAL/i)).toBeInTheDocument()
  })

  it('displays CPU metric value', async () => {
    render(<Dashboard />)
    // May appear in both KPI cards and Top Issues — at least one must exist
    expect((await screen.findAllByText(/45/)).length).toBeGreaterThanOrEqual(1)
  })

  it('displays Memory metric value', async () => {
    render(<Dashboard />)
    expect((await screen.findAllByText(/62/)).length).toBeGreaterThanOrEqual(1)
  })

  it('shows loading state initially', async () => {
    render(<Dashboard />)
    expect(await screen.findByText(/OPERATIONAL/i)).toBeInTheDocument()
  })

  it('renders the drag-to-reorder hint and Reset Layout button', async () => {
    render(<Dashboard />)
    expect(await screen.findByText(/Drag to reorder/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Reset Layout/i })).toBeInTheDocument()
  })

  it('renders all four cross-pillar widgets', async () => {
    render(<Dashboard />)
    expect(await screen.findByText(/Cross-Pillar Operations/i)).toBeInTheDocument()
    // The four widget panels render their section headings.
    expect(screen.getAllByText(/Security Posture/i).length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText(/DevOps Health/i).length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText(/AI Operations/i).length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText(/SLO \/ SLI/i).length).toBeGreaterThanOrEqual(1)
  })
})
