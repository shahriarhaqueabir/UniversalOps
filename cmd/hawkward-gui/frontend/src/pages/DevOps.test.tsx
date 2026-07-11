import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { DevOps } from './DevOps'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useQueryClient: vi.fn(() => ({ invalidateQueries: vi.fn() })),
}))

vi.mock('@/hooks/useBackend', () => ({
  useBackend: vi.fn(),
}))

vi.mock('@/stores/useSettingsStore', () => ({
  useSettingsStore: () => ({ refreshInterval: 5000 }),
}))

// Mock scroll utilities
vi.mock('@tanstack/react-virtual', () => ({
  useVirtualizer: vi.fn(() => ({
    getVirtualItems: () => [],
    getTotalSize: () => 0,
    scrollToIndex: vi.fn(),
  })),
}))

describe('DevOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useBackend as any).mockReturnValue({ call: mockCall })

    ;(useQuery as any).mockImplementation(() => ({
      data: null,
      isLoading: false,
    }))
  })

  it('renders page header', () => {
    render(<DevOps />)
    expect(screen.getByText(/DevOps Console/i)).toBeInTheDocument()
  })

  it('shows tabs', () => {
    render(<DevOps />)
    expect(screen.getByText(/Interactive Terminal/i)).toBeInTheDocument()
    expect(screen.getByText(/System Services/i)).toBeInTheDocument()
    expect(screen.getByText(/Git/i)).toBeInTheDocument()
  })

  it('switches to Git tab', () => {
    render(<DevOps />)
    const gitTab = screen.getByText(/Git/i)
    fireEvent.click(gitTab)
    // React Query would fetch Git data here
  })
})
