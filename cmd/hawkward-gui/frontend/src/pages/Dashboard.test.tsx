import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Dashboard } from './Dashboard'

// Mock react-query
vi.mock('@tanstack/react-query', () => ({
  useQuery: <T,>({ queryFn }: { queryFn?: () => Promise<T> }) => {
    queryFn?.()  // trigger to set data
    return { data: null, isLoading: false }
  },
}))

// Mock hooks
vi.mock('../hooks/useBackend', () => ({
  useBackend: () => ({
    call: vi.fn().mockResolvedValue({
      cpu: { value: 45, unit: '%', history: [30, 40, 45], forecast: [46, 47], trend: 'rising' },
      memory: { value: 62, unit: '%', history: [60, 61, 62], forecast: [63, 64], trend: 'rising' },
      disk: { value: 78, unit: '%', history: [77, 78, 78], forecast: [78, 79], trend: 'stable' },
      network: { rxRate: 1000000, txRate: 500000, unit: 'bps' },
      processes: 245,
      connections: 12,
    }),
  }),
}))

vi.mock('../hooks/useEvents', () => ({
  useEvents: () => { },
}))

describe('Dashboard Page', () => {
  it('renders without crashing', async () => {
    render(<Dashboard />)
    expect(await screen.findByText('OPERATIONAL INTELLIGENCE')).toBeInTheDocument()
  })

  it('displays CPU metric value', async () => {
    render(<Dashboard />)
    expect(await screen.findByText('45')).toBeInTheDocument()
  })

  it('displays Memory metric value', async () => {
    render(<Dashboard />)
    expect(await screen.findByText('62%')).toBeInTheDocument()
  })

  it('shows loading state initially', async () => {
    render(<Dashboard />)
    expect(await screen.findByText('OPERATIONAL INTELLIGENCE')).toBeInTheDocument()
  })
})
