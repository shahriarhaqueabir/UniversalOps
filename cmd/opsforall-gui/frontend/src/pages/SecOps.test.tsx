import { render, screen, fireEvent, waitFor } from '@testing-library/react'
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

const mockFirewallRules = [
  { name: 'Block SSH', enabled: true, action: 'BLOCK', protocol: 'TCP', port: 22, direction: 'inbound' as const, remote_ip: '0.0.0.0/0' },
  { name: 'Allow HTTP', enabled: false, action: 'ALLOW', protocol: 'TCP', port: 80, direction: 'inbound' as const, remote_ip: '0.0.0.0/0' },
]

const mockUsers = [
  { name: 'admin', uid: 500, gid: 500, groups: ['wheel'], shell: '/bin/bash', home: '/home/admin', status: 'active' as const, last_login: '2026-07-01' },
  { name: 'guest', uid: 1000, gid: 1000, groups: ['users'], shell: '/sbin/nologin', home: '/home/guest', status: 'locked' as const, last_login: '' },
]

const mockListeningPorts = [
  { port: 443, protocol: 'tcp', process_name: 'nginx', pid: 1234, state: 'listening', is_external: false },
  { port: 3306, protocol: 'tcp', process_name: 'mysqld', pid: 5678, state: 'listening', is_external: false },
]

describe('SecOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useBackend).mockReturnValue({ call: mockCall })
    vi.mocked(useQuery).mockImplementation((opts: { queryKey: string[] }) => {
      const key = opts.queryKey[0]
      if (key === 'secops-firewall') return { data: mockFirewallRules, isLoading: false }
      if (key === 'secops-firewall-status') return { data: { enabled: true, profiles: [] }, isLoading: false }
      if (key === 'secops-users') return { data: mockUsers, isLoading: false }
      if (key === 'secops-listening') return { data: mockListeningPorts, isLoading: false }
      if (key === 'secops-risks') return { data: [], isLoading: false }
      if (key === 'secops-events') return { data: [], isLoading: false }
      if (key === 'secops-tasks') return { data: [], isLoading: false }
      if (key === 'secops-security-score') {
        return { data: { score: 85, grade: 'B', breakdown: {}, recommendations: [] }, isLoading: false }
      }
      if (key === 'secops-security-summary') {
        return { data: { score: 85, summary: 'Good', risks: [], recommendations: [], analyzedAt: new Date().toISOString() }, isLoading: false }
      }
      if (key === 'secops-defender') return { data: { enabled: true, status: 'active', definitions: [] }, isLoading: false }
      if (key === 'secops-health') return { data: {}, isLoading: false }
      return { data: null, isLoading: false }
    })
    mockCall.mockResolvedValue(null)
  })

  it('renders page header', () => {
    render(<SecOps />)
    expect(screen.getByText(/Security Operations/i)).toBeInTheDocument()
  })

  it('shows security score', () => {
    render(<SecOps />)
    expect(screen.getByText(/Security Score/i)).toBeInTheDocument()
    expect(screen.getAllByText(/85/i).length).toBeGreaterThan(0)
  })

  it('navigates to Users tab', () => {
    render(<SecOps />)
    const usersTab = screen.getByRole('tab', { name: /Users tab/i })
    fireEvent.click(usersTab)
    expect(screen.getByText(/Identity & Access/i)).toBeInTheDocument()
  })

  it('navigates to Firewall tab and shows rules', async () => {
    render(<SecOps />)
    const firewallTab = screen.getByRole('tab', { name: /Firewall tab/i })
    fireEvent.click(firewallTab)
    await waitFor(() => {
      expect(screen.getByText(/Block SSH/i)).toBeInTheDocument()
      expect(screen.getByText(/Allow HTTP/i)).toBeInTheDocument()
    })
  })

  it('navigates to Listening Ports tab', () => {
    render(<SecOps />)
    const portsTab = screen.getByRole('tab', { name: /Listening tab/i })
    fireEvent.click(portsTab)
    expect(screen.getByText(/Total Listening/i)).toBeInTheDocument()
  })

  it('shows security summary panel', () => {
    render(<SecOps />)
    expect(screen.getByText(/Security Summary/i)).toBeInTheDocument()
  })

  it('displays fallback score when summary is missing', () => {
    vi.mocked(useQuery).mockImplementation((opts: { queryKey: string[] }) => {
      const key = opts.queryKey[0]
      if (key === 'secops-security-summary') return { data: null, isLoading: false }
      if (key === 'secops-security-score') return { data: { score: 85, grade: 'B', breakdown: {}, recommendations: [] }, isLoading: false }
      const emptyKeys = ['secops-firewall', 'secops-listening', 'secops-risks', 'secops-events', 'secops-tasks', 'secops-users']
      return { data: emptyKeys.includes(key) ? [] : null, isLoading: false, refetch: vi.fn() }
    })
    render(<SecOps />)
    expect(screen.getByText(/Security Score/i)).toBeInTheDocument()
  })

  it('renders risk assessment panel', () => {
    render(<SecOps />)
    expect(screen.getByText(/Risk Assessment/i)).toBeInTheDocument()
  })

  it('renders DataFreshnessIndicator in summary', () => {
    render(<SecOps />)
    expect(screen.getByText(/Last Analyzed/i)).toBeInTheDocument()
  })
})
