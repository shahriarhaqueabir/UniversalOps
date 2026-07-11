import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { AIOps } from './AIOps'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'

// Mock dependencies
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
  useOllamaStore: () => ({ setStatus: vi.fn() }),
}))

describe('AIOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useBackend as any).mockReturnValue({ call: mockCall })

    // Default mock implementation for useQuery
    ;(useQuery as any).mockImplementation(() => ({
      data: { available: true, model: 'agentic-coder', version: '1.0' },
      isLoading: false,
      refetch: vi.fn(),
    }))
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

  it('switches between tabs', async () => {
    render(<AIOps />)

    const reportsTab = screen.getByRole('tab', { name: /Intelligence Reports/i })
    fireEvent.click(reportsTab)

    expect(screen.getByText(/Intelligence Templates/i)).toBeInTheDocument()
  })
})
