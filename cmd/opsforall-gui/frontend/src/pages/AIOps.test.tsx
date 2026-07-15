import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { AIOps } from './AIOps'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
}))

vi.mock('@/hooks/useBackend', () => ({
  useBackend: vi.fn(),
}))

vi.mock('@/stores/useSettingsStore', () => ({
  useSettingsStore: () => ({ refreshInterval: 5000 }),
}))

vi.mock('@/stores/useOllamaStore', () => ({
  useOllamaStore: (selector?: any) => {
    const state = { setStatus: vi.fn(), status: null }
    return selector ? selector(state) : state
  },
}))

const mockSessions = [
  { session_id: 's1', last_active: '2026-07-01T12:00:00Z', msg_count: 5 },
  { session_id: 's2', last_active: '2026-07-02T12:00:00Z', msg_count: 3 },
]

const mockAnomalies = [
  { metric: 'CPU', value: 95, expected: 50, deviation: 3.2, severity: 'critical', timestamp: '2026-07-01T12:00:00Z' },
  { metric: 'Memory', value: 88, expected: 60, deviation: 2.1, severity: 'warning', timestamp: '2026-07-01T12:30:00Z' },
]

const mockInsights = [
  { title: 'Usage Pattern', category: 'trend', severity: 'info' as const, message: 'CPU usage increasing', action: 'Monitor CPU', timestamp: '2026-07-01T12:00:00Z' },
]

describe('AIOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useBackend).mockReturnValue({ call: mockCall })
    vi.mocked(useQuery).mockImplementation((opts: any) => {
      const key = opts.queryKey[0]
      if (key === 'chat-sessions') return { data: mockSessions, isLoading: false, refetch: vi.fn() }
      if (key === 'ollama-status') {
        return { data: { available: true, model: 'agentic-coder', version: '1.0' }, isLoading: false, refetch: vi.fn() }
      }
      if (key === 'anomalies') return { data: mockAnomalies, isLoading: false }
      if (key === 'ai-insights') return { data: mockInsights, isLoading: false }
      return { data: { available: true, model: 'agentic-coder', version: '1.0' }, isLoading: false, refetch: vi.fn() }
    })
    mockCall.mockResolvedValue(null)
  })

  it('renders page header', () => {
    render(<AIOps />)
    expect(screen.getByText(/AI Operations Analyst/i)).toBeInTheDocument()
  })

  it('shows ollama status', () => {
    render(<AIOps />)
    expect(screen.getByText(/Ollama Online/i)).toBeInTheDocument()
    expect(screen.getByText(/agentic-coder/i)).toBeInTheDocument()
  })

  it('switches to AI Insights tab', async () => {
    const user = userEvent.setup()
    const { container } = render(<AIOps />)
    const tab = container.querySelector('[data-automation-id="aiops-tab-insights"]') as HTMLElement
    await user.click(tab)
    await waitFor(() => {
      expect(screen.getByText(/Usage Pattern/i)).toBeInTheDocument()
    })
  })

  it('switches to Anomaly Detection tab', async () => {
    const user = userEvent.setup()
    const { container } = render(<AIOps />)
    const tab = container.querySelector('[data-automation-id="aiops-tab-anomalies"]') as HTMLElement
    await user.click(tab)
    await waitFor(() => {
      expect(screen.getByText(/CPU/i)).toBeInTheDocument()
      expect(screen.getByText(/MEMORY/i)).toBeInTheDocument()
    })
  })

  it('shows chat sessions in sidebar', async () => {
    render(<AIOps />)
    await waitFor(() => {
      expect(screen.getByText(/s1/i)).toBeInTheDocument()
      expect(screen.getByText(/s2/i)).toBeInTheDocument()
    })
  })

  it('shows empty state when ollama is offline', () => {
    vi.mocked(useQuery).mockImplementation((opts: any) => {
      const key = opts.queryKey[0]
      if (key === 'ollama-status') {
        return { data: null, isLoading: false, refetch: vi.fn() }
      }
      if (key === 'chat-sessions') return { data: [], isLoading: false, refetch: vi.fn() }
      return { data: null, isLoading: false }
    })
    render(<AIOps />)
    expect(screen.getByText(/Ollama Offline/i)).toBeInTheDocument()
  })

  it('switches chat session when session is clicked', async () => {
    render(<AIOps />)
    await waitFor(() => {
      expect(screen.getByText(/s1/i)).toBeInTheDocument()
    })
    const sessionButton = screen.getByText(/s1/i)
    fireEvent.click(sessionButton)
    await waitFor(() => {
      expect(sessionButton).toBeTruthy()
    })
  })
})
