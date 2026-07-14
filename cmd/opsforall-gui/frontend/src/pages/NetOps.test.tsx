// @ts-nocheck
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import type { ReactNode } from 'react'
import { NetOps } from './NetOps'
import { useQuery } from '@tanstack/react-query'

// Mock useQuery
vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
  QueryClient: class { clear() {} },
  QueryClientProvider: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}))

// Mock useBackend hook
const mockCall = vi.fn()
vi.mock('@/hooks/useBackend', () => ({
  useBackend: () => ({
    call: mockCall,
  }),
}))

vi.mock('@/stores/useSettingsStore', () => ({
  useSettingsStore: () => ({
    refreshInterval: 5000,
    pingCount: 4,
    dnsTimeout: 3000,
  }),
}))

vi.mock('@/components/ui/DataFreshnessIndicator', () => ({
  DataFreshnessIndicator: () => null,
}))

describe('NetOps Page', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    vi.mocked(useQuery).mockImplementation(() => {
      return { data: [], isLoading: false, dataUpdatedAt: Date.now() }
    })

    mockCall.mockImplementation(async (method: string) => {
      if (method === 'NetOps.GetInterfaces') return []
      if (method === 'NetOps.GetConnections') return []
      if (method === 'NetOps.GetRecentChanges') return []
      return []
    })
  })

  it('renders correctly with sidebar', () => {
    render(<NetOps />)
    expect(screen.getByText(/NETWORK OPERATIONS/i)).toBeTruthy()
    expect(screen.getByText('Overview')).toBeTruthy()
    expect(screen.getByText('Ping')).toBeTruthy()
    expect(screen.getAllByText('DNS').length).toBeGreaterThanOrEqual(1)
  })

  it('renders all sidebar category groups', () => {
    render(<NetOps />)
    expect(screen.getByText('INSPECTION')).toBeTruthy()
    expect(screen.getByText('DIAGNOSIS')).toBeTruthy()
    expect(screen.getByText('ACTION')).toBeTruthy()
  })

  it('renders all sidebar categories', () => {
    render(<NetOps />)
    const expectedCategories = [
      'Overview', 'Connections', 'Interfaces', 'ARP Table', 'Routing', 'WiFi',
      'Ping', 'Traceroute', 'Port Scan', 'Bandwidth',
      'DNS Advanced', 'Multi-Ping', 'Health Check',
      'Firewall', 'Discovery', 'Actions',
    ]
    for (const cat of expectedCategories) {
      expect(screen.getByText(cat)).toBeTruthy()
    }
    // DNS and VPN appear in both sidebar and OverviewTab connectivity cards
    expect(screen.getAllByText('DNS').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('VPN').length).toBeGreaterThanOrEqual(1)
  })

  it('switches categories via sidebar', () => {
    render(<NetOps />)

    fireEvent.click(screen.getByText('Ping'))
    expect(screen.getByText(/ICMP Probe/i)).toBeTruthy()

    fireEvent.click(screen.getByText('DNS'))
    expect(screen.getByText(/Domain Resolution/i)).toBeTruthy()
  })
})
