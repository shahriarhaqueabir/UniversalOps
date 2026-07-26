import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { SecOps } from './SecOps'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  useMutation: vi.fn(() => ({
    mutate: vi.fn(),
    isPending: false,
    isError: false,
    error: null,
    data: null,
  })),
  useQueryClient: vi.fn(() => ({ invalidateQueries: vi.fn() })),
}))

vi.mock('@/hooks/useBackend', () => ({
  useBackend: vi.fn(),
}))

vi.mock('@/stores/useSettingsStore', () => ({
  useSettingsStore: () => ({ refreshInterval: 5000 }),
}))

const mockScore = { score: 85, grade: 'B', breakdown: { Defender: 30, Firewall: 18, Users: 8, Ports: 9, Events: 10 }, recommendations: ['Keep Defender signatures up to date'] }
const mockDefender = { enabled: true, real_time_protection: true, cloud_protection: true, up_to_date: true, threats_detected: 0, last_scan: '2026-07-13', signature_age: '1 day', full_scan_age: 3 }
const mockFwStatus = { enabled: true, profiles: [{ name: 'Domain', enabled: true }, { name: 'Private', enabled: true }, { name: 'Public', enabled: true }] }

describe('SecOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useBackend).mockReturnValue({ call: mockCall })
    vi.mocked(useQuery).mockImplementation((opts: any) => {
      const key = opts.queryKey[0]
      if (key === 'secops-score') return { data: mockScore, isLoading: false } as any
      if (key === 'secops-defender') return { data: mockDefender, isLoading: false } as any
      if (key === 'secops-firewall-status') return { data: mockFwStatus, isLoading: false } as any
      if (key === 'secops-users') return { data: [], isLoading: false } as any
      if (key === 'secops-listening') return { data: [], isLoading: false } as any
      if (key === 'secops-events') return { data: [], isLoading: false } as any
      if (key === 'secops-health') return { data: {}, isLoading: false } as any
      if (key === 'secops-password-policy') return { data: { max_age: 90, min_length: 8, complexity: true, lockout_threshold: 5, lockout_duration: 30 }, isLoading: false } as any
      if (key === 'secops-failed-logins') return { data: [], isLoading: false } as any
      if (key === 'secops-lockouts') return { data: [], isLoading: false } as any
      if (key === 'secops-tls-certs') return { data: [], isLoading: false } as any
      if (key === 'secops-public-exposure') return { data: [], isLoading: false } as any
      if (key === 'secops-disk-encryption') return { data: [], isLoading: false } as any
      if (key === 'secops-secure-boot') return { data: { enabled: true, state: 'OK' }, isLoading: false } as any
      if (key === 'secops-services') return { data: [], isLoading: false } as any
      if (key === 'secops-tasks') return { data: [], isLoading: false } as any
      if (key === 'secops-privilege-events') return { data: [], isLoading: false } as any
      if (key === 'secops-timeline') return { data: [], isLoading: false } as any
      if (key === 'secops-hardening') return { data: [], isLoading: false } as any
      if (key === 'secops-ssh-config') return { data: null, isLoading: false } as any
      if (key === 'secops-firewall-rules') return { data: [], isLoading: false } as any
      if (key === 'secops-listening-ports') return { data: [], isLoading: false } as any
      return { data: null, isLoading: false } as any
    })
    mockCall.mockResolvedValue(null)
  })

  it('renders page header', () => {
    render(<SecOps />)
    expect(screen.getByText('Security Operations')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /threat.*perimeter/i })).toBeInTheDocument()
  })

  it('renders sidebar category groups', () => {
    render(<SecOps />)
    expect(screen.getByText('ASSESSMENT')).toBeInTheDocument()
    expect(screen.getByText('DETECTION')).toBeInTheDocument()
    expect(screen.getByText('RESPONSE')).toBeInTheDocument()
  })

  it('renders sidebar category buttons', () => {
    render(<SecOps />)
    expect(screen.getByText('Overview')).toBeInTheDocument()
    expect(screen.getByText('Identity & Access')).toBeInTheDocument()
    expect(screen.getByText('Perimeter Security')).toBeInTheDocument()
    expect(screen.getByText('Endpoint Security')).toBeInTheDocument()
    expect(screen.getByText('Log & Events')).toBeInTheDocument()
    expect(screen.getByText('Security Hardening')).toBeInTheDocument()
    expect(screen.getByText('Security Audit')).toBeInTheDocument()
    expect(screen.getByText('Incident Response')).toBeInTheDocument()
  })

  it('shows overview by default with score', async () => {
    render(<SecOps />)
    await waitFor(() => {
      expect(screen.getByText('Security Operations Center')).toBeInTheDocument()
    })
    await waitFor(() => {
      expect(screen.getAllByText('85').length).toBeGreaterThan(0)
    })
    await waitFor(() => {
      expect(screen.getByText(/Grade B/i)).toBeInTheDocument()
    })
  })

  it('navigates to Identity & Access sidebar category', async () => {
    render(<SecOps />)
    fireEvent.click(screen.getByText('Identity & Access'))
    await waitFor(() => {
      expect(screen.getByText('Identity & Access', { selector: 'h3' })).toBeInTheDocument()
    })
  })

  it('navigates to Network Security sidebar category', async () => {
    render(<SecOps />)
    fireEvent.click(screen.getByText('Perimeter Security'))
    await waitFor(() => {
      expect(screen.getByText('Perimeter Security', { selector: 'h3' })).toBeInTheDocument()
    })
  })

  it('navigates to Endpoint Security sidebar category', async () => {
    render(<SecOps />)
    fireEvent.click(screen.getByText('Endpoint Security'))
    await waitFor(() => {
      expect(screen.getByText('Endpoint Security', { selector: 'h3' })).toBeInTheDocument()
    })
  })

  it('navigates to Log & Events sidebar category', async () => {
    render(<SecOps />)
    fireEvent.click(screen.getByText('Log & Events'))
    await waitFor(() => {
      expect(screen.getByText('Log & Event Analysis')).toBeInTheDocument()
    })
  })

  it('navigates to Security Hardening sidebar category', async () => {
    render(<SecOps />)
    fireEvent.click(screen.getByText('Security Hardening'))
    await waitFor(() => {
      expect(screen.getByText('Security Hardening', { selector: 'h3' })).toBeInTheDocument()
    })
  })

  it('navigates to Security Audit sidebar category', async () => {
    render(<SecOps />)
    fireEvent.click(screen.getByText('Security Audit'))
    await waitFor(() => {
      expect(screen.getByText('Run Security Audit')).toBeInTheDocument()
    })
  })

  it('navigates to Incident Response sidebar category', async () => {
    render(<SecOps />)
    fireEvent.click(screen.getByText('Incident Response'))
    await waitFor(() => {
      expect(screen.getByText('Isolate Host')).toBeInTheDocument()
    })
  })
})
