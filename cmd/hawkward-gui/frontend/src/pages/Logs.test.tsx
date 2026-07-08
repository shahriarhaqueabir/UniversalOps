import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Logs } from './Logs'

vi.mock('../hooks/useBackend', () => ({
  useBackend: () => ({
    call: vi.fn().mockResolvedValue([
      { timestamp: '2026/07/08 12:00:00', level: 'INFO', module: 'system', message: 'System started' },
      { timestamp: '2026/07/08 12:00:05', level: 'WARN', module: 'network', message: 'High latency detected' },
      { timestamp: '2026/07/08 12:00:10', level: 'ERROR', module: 'disk', message: 'Disk space low' },
    ]),
  }),
}))

describe('Logs Page', () => {
  it('renders without crashing', () => {
    render(<Logs />)
    expect(screen.getByText('Live Event Aggregator')).toBeInTheDocument()
  })

  it('displays log entries after loading', async () => {
    render(<Logs />)
    expect(await screen.findByText('System started')).toBeInTheDocument()
    expect(await screen.findByText('High latency detected')).toBeInTheDocument()
    expect(await screen.findByText('Disk space low')).toBeInTheDocument()
  })

  it('shows level filter badges', () => {
    render(<Logs />)
    expect(screen.getByText('INFO')).toBeInTheDocument()
    expect(screen.getByText('WARN')).toBeInTheDocument()
    expect(screen.getByText('ERROR')).toBeInTheDocument()
    expect(screen.getByText('DEBUG')).toBeInTheDocument()
  })
})
