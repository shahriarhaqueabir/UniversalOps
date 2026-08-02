import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AlertsDashboard } from './AlertsDashboard'

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

const mockAlerts = [
  {
    id: 'alert-1',
    level: 'CRITICAL',
    metric: 'cpu.percent',
    message: 'cpu.percent > 90.0 (current 95.0)',
    value: 95,
    threshold: 90,
    timestamp: '2026-08-02T10:00:00Z',
    resolved: false,
  },
]

// Hoisted refs so tests can override the call mock and rules data
const callRef = vi.hoisted(() => ({ fn: vi.fn() }))
const rulesRef = vi.hoisted(() => ({
  current: [
    {
      metric: 'cpu.percent',
      condition: '>',
      threshold: 90,
      flap_count: 2,
      severity: 'CRITICAL',
      message: '',
    },
  ],
}))

vi.mock('@tanstack/react-query', () => ({
  useQuery: (opts: { queryKey?: string[] }) => {
    const key = opts?.queryKey?.[0]
    if (key === 'alertHistory') {
      return { data: mockAlerts, isLoading: false, refetch: vi.fn() } as const
    }
    if (key === 'alertRules') {
      return { data: rulesRef.current, isLoading: false, refetch: vi.fn() } as const
    }
    return { data: [], isLoading: false, refetch: vi.fn() } as const
  },
  useMutation: (opts: { mutationFn?: (args: unknown) => Promise<unknown> }) => ({
    mutate: (args: unknown) => {
      if (opts?.mutationFn) {
        opts.mutationFn(args)
      }
    },
    isPending: false,
  }),
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}))

vi.mock('../hooks/useBackend', () => ({
  useBackend: () => ({
    call: callRef.fn,
  }),
}))

describe('AlertsDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    callRef.fn.mockResolvedValue([])
  })

  it('renders without crashing', async () => {
    render(<AlertsDashboard />)
    await waitFor(() => {
      expect(screen.getByText('Alerts')).toBeInTheDocument()
    })
  })

  it('shows the Rules tab with existing rules', async () => {
    const user = userEvent.setup()
    render(<AlertsDashboard />)
    await waitFor(() => {
      expect(screen.getByText('Rules')).toBeInTheDocument()
    })
    await user.click(screen.getByText('Rules'))
    await waitFor(() => {
      expect(screen.getByText('cpu.percent')).toBeInTheDocument()
    })
  })

  it('opens the Add Rule form and submits a new rule', async () => {
    const user = userEvent.setup()
    render(<AlertsDashboard />)
    await waitFor(() => {
      expect(screen.getByText('Rules')).toBeInTheDocument()
    })
    await user.click(screen.getByText('Rules'))

    // Open the form
    await user.click(screen.getByText('+ Add Rule'))
    await waitFor(() => {
      expect(screen.getByText('New Alert Rule')).toBeInTheDocument()
    })

    // Fill the form
    await user.type(screen.getByPlaceholderText('cpu.percent'), 'mem.used')
    await user.type(screen.getByPlaceholderText('90'), '85')
    await user.click(screen.getByText('Add Rule'))

    // Verify the backend call was made with the right args
    await waitFor(() => {
      expect(callRef.fn).toHaveBeenCalledWith(
        'AlertAPI.AddRule',
        'mem.used',
        85,
        'warning',
        'gt',
        2,
        '',
      )
    })
  })

  it('validates that metric is required', async () => {
    const user = userEvent.setup()
    render(<AlertsDashboard />)
    await waitFor(() => {
      expect(screen.getByText('Rules')).toBeInTheDocument()
    })
    await user.click(screen.getByText('Rules'))
    await user.click(screen.getByText('+ Add Rule'))
    await waitFor(() => {
      expect(screen.getByText('New Alert Rule')).toBeInTheDocument()
    })

    // Submit without metric
    await user.click(screen.getByText('Add Rule'))

    // Should NOT call the backend
    expect(callRef.fn).not.toHaveBeenCalledWith(
      'AlertAPI.AddRule',
      expect.anything(),
      expect.anything(),
      expect.anything(),
      expect.anything(),
      expect.anything(),
      expect.anything(),
    )
  })
})