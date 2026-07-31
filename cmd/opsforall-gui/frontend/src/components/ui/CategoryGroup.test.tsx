import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CategoryGroup } from './CategoryGroup'

describe('CategoryGroup', () => {
  const categories = [
    { id: 'ping' as const, label: 'Ping', icon: <span data-testid="icon-ping">📍</span> },
    { id: 'traceroute' as const, label: 'Traceroute', icon: <span data-testid="icon-trace">🔄</span> },
  ]

  const defaultProps = {
    label: 'Diagnostics',
    group: 'diagnosis',
    page: 'netops',
    categories,
    active: 'ping' as const,
    onSelect: vi.fn(),
  }

  it('renders the group label', () => {
    render(<CategoryGroup {...defaultProps} />)
    expect(screen.getByText('Diagnostics')).toBeInTheDocument()
  })

  it('renders all category items', () => {
    render(<CategoryGroup {...defaultProps} />)
    expect(screen.getByText('Ping')).toBeInTheDocument()
    expect(screen.getByText('Traceroute')).toBeInTheDocument()
  })

  it('renders category icons', () => {
    render(<CategoryGroup {...defaultProps} />)
    expect(screen.getByTestId('icon-ping')).toBeInTheDocument()
    expect(screen.getByTestId('icon-trace')).toBeInTheDocument()
  })

  it('marks the active item', () => {
    render(<CategoryGroup {...defaultProps} active="traceroute" />)
    const activeButton = screen.getByText('Traceroute').closest('button')
    expect(activeButton).toBeTruthy()
    // Active item gets a shadown class
    expect(activeButton!.className).toContain('shadow')
  })

  it('sets data-automation-id on each button', () => {
    render(<CategoryGroup {...defaultProps} />)
    const pingBtn = screen.getByText('Ping').closest('button')
    expect(pingBtn).toHaveAttribute('data-automation-id', 'netops-tab-ping')

    const traceBtn = screen.getByText('Traceroute').closest('button')
    expect(traceBtn).toHaveAttribute('data-automation-id', 'netops-tab-traceroute')
  })

  it('calls onSelect when a category is clicked', async () => {
    const onSelect = vi.fn()
    render(<CategoryGroup {...defaultProps} onSelect={onSelect} />)
    const user = userEvent.setup()
    await user.click(screen.getByText('Traceroute'))
    expect(onSelect).toHaveBeenCalledWith('traceroute')
  })

  it('works with different page/group combinations', () => {
    render(<CategoryGroup {...defaultProps} page="secops" group="response" />)
    const activeBtn = screen.getByText('Ping').closest('button')
    expect(activeBtn).toBeTruthy()
    // secops-response uses danger color
    expect(activeBtn!.className).toContain('shadow')
  })

  it('renders with empty categories without error', () => {
    render(<CategoryGroup {...defaultProps} categories={[]} />)
    expect(screen.getByText('Diagnostics')).toBeInTheDocument()
  })
})
