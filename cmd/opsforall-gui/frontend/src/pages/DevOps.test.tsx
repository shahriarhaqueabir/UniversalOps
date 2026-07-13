import { render, screen, fireEvent, waitFor } from '@testing-library/react'
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

vi.mock('@tanstack/react-virtual', () => ({
  useVirtualizer: vi.fn(() => ({
    getVirtualItems: () => [],
    getTotalSize: () => 0,
    scrollToIndex: vi.fn(),
  })),
}))

const mockServices = [
  { name: 'nginx', display_name: 'Nginx', status: 'running' as const, start_type: 'auto' as const },
  { name: 'sshd', display_name: 'SSH Daemon', status: 'running' as const, start_type: 'auto' as const },
  { name: 'cups', display_name: 'CUPS', status: 'stopped' as const, start_type: 'manual' as const },
]

const mockContainers = { containers: [
  { id: 'abc123', name: 'web-app', image: 'node:20', status: 'running' as const, ports: '3000->3000', uptime: '2d' },
  { id: 'def456', name: 'redis-cache', image: 'redis:7', status: 'exited' as const, ports: '', uptime: '0s' },
]}

const mockGit = { repositories: [
  { path: '/repo/app', branch: 'main', clean: true, ahead: 0, behind: 0, lastCommit: 'fix: update config' },
]}

const mockServers = [
  { port: 3000, protocol: 'tcp', process: 'node', pid: 1234, framework: 'express', health: 'healthy' },
]

const mockTools = [
  { name: 'git', version: '2.45', status: 'available' as const },
  { name: 'node', version: '22', status: 'available' as const },
]

const mockEnv = { key_vars: [
  { name: 'NODE_ENV', value: 'production' },
  { name: 'DB_HOST', value: 'localhost' },
]}

describe('DevOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useBackend).mockReturnValue({ call: mockCall })
    vi.mocked(useQuery).mockImplementation((opts: { queryKey: string[] }) => {
      const key = opts.queryKey[0]
      if (key === 'devops-services') return { data: mockServices, isLoading: false }
      if (key === 'devops-docker-status') return { data: null, isLoading: false }
      if (key === 'devops-k8s-status') return { data: null, isLoading: false }
      if (key === 'devops-service-summary') return { data: null, isLoading: false }
      if (key === 'devops-workflows') return { data: null, isLoading: false }
      if (key === 'devops-containers') return { data: mockContainers, isLoading: false }
      if (key === 'devops-git') return { data: mockGit, isLoading: false }
      if (key === 'devops-servers') return { data: mockServers, isLoading: false }
      if (key === 'devops-env') return { data: mockEnv, isLoading: false }
      if (key === 'devops-tools') return { data: mockTools, isLoading: false }
      if (key === 'devops-ai-suggestions') return { data: [], isLoading: false }
      return { data: null, isLoading: false }
    })
    mockCall.mockResolvedValue(null)
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

  it('switches to Git tab', async () => {
    render(<DevOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Git/i }))
    await waitFor(() => {
      expect(screen.getByText(/main/i)).toBeInTheDocument()
      expect(screen.getByText(/Clean/i)).toBeInTheDocument()
    })
  })

  it('switches to Services tab and shows services', async () => {
    render(<DevOps />)
    fireEvent.click(screen.getByRole('tab', { name: /System Services/i }))
    await waitFor(() => {
      expect(screen.getByText(/Nginx/i)).toBeInTheDocument()
      expect(screen.getByText(/SSH Daemon/i)).toBeInTheDocument()
      expect(screen.getByText(/CUPS/i)).toBeInTheDocument()
    })
  })

  it('switches to Containers tab', async () => {
    render(<DevOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Containers/i }))
    await waitFor(() => {
      expect(screen.getByText(/web-app/i)).toBeInTheDocument()
      expect(screen.getByText(/redis-cache/i)).toBeInTheDocument()
    })
  })

  it('switches to Servers tab', async () => {
    render(<DevOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Servers/i }))
    await waitFor(() => {
      expect(screen.getByText(/:3000/i)).toBeInTheDocument()
      expect(screen.getByText(/node/i)).toBeInTheDocument()
    })
  })

  it('switches to Terminal tab', () => {
    render(<DevOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Interactive Terminal/i }))
    expect(screen.getByPlaceholderText(/Enter shell command/i)).toBeInTheDocument()
  })

  it('switches to Toolbox tab', async () => {
    render(<DevOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Toolbox/i }))
    await waitFor(() => {
      expect(screen.getByText(/git/i)).toBeInTheDocument()
      expect(screen.getByText(/node/i)).toBeInTheDocument()
    })
  })

  it('switches to Environment tab', async () => {
    render(<DevOps />)
    fireEvent.click(screen.getByRole('tab', { name: /Environment/i }))
    await waitFor(() => {
      expect(screen.getByText(/NODE_ENV/i)).toBeInTheDocument()
      expect(screen.getByText(/DB_HOST/i)).toBeInTheDocument()
    })
  })
})
