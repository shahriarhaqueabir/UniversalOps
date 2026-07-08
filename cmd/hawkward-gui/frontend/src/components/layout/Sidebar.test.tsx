import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { Sidebar } from './Sidebar'

describe('Sidebar', () => {
  it('renders all navigation items', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    expect(screen.getByText('Dashboard')).toBeDefined()
    expect(screen.getByText('System Ops')).toBeDefined()
    expect(screen.getByText('Network Ops')).toBeDefined()
    expect(screen.getByText('Security Ops')).toBeDefined()
    expect(screen.getByText('DevOps')).toBeDefined()
    expect(screen.getByText('AI Ops')).toBeDefined()
  })

  it('renders tools section items', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    expect(screen.getByText('Network Design')).toBeDefined()
    expect(screen.getByText('Logs')).toBeDefined()
    expect(screen.getByText('Settings')).toBeDefined()
  })

  it('calls onNavigate when an item is clicked', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    fireEvent.click(screen.getByText('System Ops'))
    expect(onNavigate).toHaveBeenCalledWith('sysops')
  })

  it('has a collapse toggle button', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    expect(screen.getByLabelText('Collapse sidebar')).toBeDefined()
  })

  it('collapses and expands when toggle is clicked', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    // Starts expanded — collapse label visible
    expect(screen.getByLabelText('Collapse sidebar')).toBeDefined()
    // Brand text visible when expanded
    expect(screen.getByText('HAWKWARD')).toBeDefined()

    // Click to collapse
    fireEvent.click(screen.getByLabelText('Collapse sidebar'))

    // Now shows expand label instead
    expect(screen.getByLabelText('Expand sidebar')).toBeDefined()
    // Brand text hidden when collapsed
    expect(screen.queryByText('HAWKWARD')).toBeNull()
  })
})
