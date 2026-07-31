import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { DevOps } from './DevOps'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { mockQueryReturn } from '@/test/mockQuery'

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
  { id: 'abc123', name: 'web-app', image: 'node:20', status: 'running' as const, ports: '3000->3000', uptime: '2d', state: 'running' },
  { id: 'def456', name: 'redis-cache', image: 'redis:7', status: 'exited' as const, ports: '', uptime: '0s', state: 'exited' },
]}

const mockGit = { repositories: [
  { path: '/repo/app', branch: 'main', clean: true, ahead: 0, behind: 0, lastCommit: 'fix: update config' },
]}

const mockGitBranches = [
  { name: 'main', current: true, last_commit: 'fix: update config' },
  { name: 'develop', current: false, last_commit: 'feat: add feature' },
]

const mockGitTags = [
  { name: 'v1.3.0', date: '2025-01-15' },
]

const mockGitStash = [
  { index: 0, message: 'WIP: temp changes' },
]

const mockGitRemotes = [
  { name: 'origin', url: 'https://github.com/test/repo.git' },
]

const mockServers = [
  { port: 3000, protocol: 'tcp', process: 'node', pid: 1234, framework: 'express', health: 'healthy' },
]

const mockTools = [
  { name: 'git', version: '2.45', status: 'available' as const },
  { name: 'node', version: '22', status: 'available' as const },
]

const mockEnv = {
  key_vars: [
    { name: 'NODE_ENV', value: 'production' },
    { name: 'DB_HOST', value: 'localhost' },
  ],
  sdks: [
    { name: 'Node.js', version: '22.0' },
  ],
  package_managers: [
    { name: 'npm', version: '10.8' },
  ],
}

describe('DevOps Page', () => {
  const mockCall = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useBackend).mockReturnValue({ call: mockCall })
    vi.mocked(useQuery).mockImplementation((opts: any) => {
      const key = opts.queryKey[0]
      if (key === 'devops-services') return mockQueryReturn({ data: mockServices })
      if (key === 'devops-docker-status') return mockQueryReturn({ data: { installed: true, running: true, version: '24.0', containers: { running: 1, stopped: 1, failed: 0, total: 2, containers: mockContainers.containers } } })
      if (key === 'devops-k8s-status') return mockQueryReturn({ data: null })
      if (key === 'devops-service-summary') return mockQueryReturn({ data: null })
      if (key === 'devops-dora') return mockQueryReturn({ data: null })
      if (key === 'devops-workflows') return mockQueryReturn({ data: null })
      if (key === 'devops-containers') return mockQueryReturn({ data: mockContainers })
      if (key === 'devops-git') return mockQueryReturn({ data: mockGit })
      if (key === 'devops-git-branches') return mockQueryReturn({ data: mockGitBranches })
      if (key === 'devops-git-tags') return mockQueryReturn({ data: mockGitTags })
      if (key === 'devops-git-stash') return mockQueryReturn({ data: mockGitStash })
      if (key === 'devops-git-remotes') return mockQueryReturn({ data: mockGitRemotes })
      if (key === 'devops-servers') return mockQueryReturn({ data: mockServers })
      if (key === 'devops-env') return mockQueryReturn({ data: mockEnv })
      if (key === 'devops-tools') return mockQueryReturn({ data: mockTools })
      if (key === 'devops-ai-suggestions') return mockQueryReturn({ data: [] })
      return mockQueryReturn({ data: undefined })
    })
    mockCall.mockResolvedValue(null)
  })

  it('renders page header', () => {
    render(<DevOps />)
    expect(screen.getByText(/DevOps Console/i)).toBeInTheDocument()
  })

  it('shows tabs', () => {
    render(<DevOps />)
    expect(screen.getByRole('tab', { name: /Overview/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Services/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Docker/i })).toBeInTheDocument()
  })

  it('switches to Services tab and shows services', async () => {
    const user = userEvent.setup()
    render(<DevOps />)
    await user.click(screen.getByRole('tab', { name: /Services/i }))
    await waitFor(() => {
      // Use exact text to avoid matching both "Nginx" (display_name) and "nginx" (name)
      expect(screen.getByText('Nginx')).toBeInTheDocument()
      expect(screen.getByText('SSH Daemon')).toBeInTheDocument()
      expect(screen.getByText('CUPS')).toBeInTheDocument()
    })
  })

  it('switches to Docker tab and shows containers', async () => {
    const user = userEvent.setup()
    render(<DevOps />)
    await user.click(screen.getByRole('tab', { name: /Docker/i }))
    await waitFor(() => {
      expect(screen.getByText(/web-app/i)).toBeInTheDocument()
      expect(screen.getByText(/redis-cache/i)).toBeInTheDocument()
    })
  })

  it('switches to Servers tab', async () => {
    const user = userEvent.setup()
    render(<DevOps />)
    await user.click(screen.getByRole('tab', { name: /Servers/i }))
    await waitFor(() => {
      expect(screen.getByText(/:3000/i)).toBeInTheDocument()
      expect(screen.getByText(/node/i)).toBeInTheDocument()
    })
  })

  it('switches to PS tab', async () => {
    const user = userEvent.setup()
    render(<DevOps />)
    await user.click(screen.getByRole('tab', { name: /^PS$/i }))
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/Enter PowerShell command/i)).toBeInTheDocument()
    })
  })

  it('switches to Environment tab', async () => {
    const user = userEvent.setup()
    render(<DevOps />)
    await user.click(screen.getByRole('tab', { name: /Env/i }))
    await waitFor(() => {
      expect(screen.getByText('NODE_ENV')).toBeInTheDocument()
      expect(screen.getByText('DB_HOST')).toBeInTheDocument()
    })
  })
})
