import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { SecOps } from './SecOps'
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

describe('SecOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useBackend as any).mockReturnValue({ call: mockCall })

    ;(useQuery as any).mockImplementation(({ queryKey }: any) => {
      if (queryKey.includes('secops-health') || queryKey.includes('secops-security-score')) {
        return { data: { score: 85, grade: 'B', breakdown: {}, recommendations: [] }, isLoading: false }
      }
      if (queryKey.includes('secops-security-summary')) {
        return { data: { score: 85, summary: 'Good', risks: [], recommendations: [], analyzedAt: new Date().toISOString() }, isLoading: false }
      }
      return { data: [], isLoading: false }
    })
  })

  it('renders page header', () => {
    render(<SecOps />)
    expect(screen.getByText(/Security Operations/i)).toBeInTheDocument()
  })

  it('shows security score', () => {
    render(<SecOps />)
    expect(screen.getByText(/Security Score/i)).toBeInTheDocument()
    expect(screen.getByText(/85/i)).toBeInTheDocument()
  })

  it('navigates tabs', () => {
    render(<SecOps />)
    const usersTab = screen.getByRole('tab', { name: /Users tab/i })
    fireEvent.click(usersTab)
    expect(screen.getByText(/Identity & Access/i)).toBeInTheDocument()
  })
})
