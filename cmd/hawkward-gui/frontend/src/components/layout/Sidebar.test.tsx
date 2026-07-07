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

  it('calls onNavigate when an item is clicked', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    fireEvent.click(screen.getByText('System Ops'))
    expect(onNavigate).toHaveBeenCalledWith('sysops')
  })

  it('collapses when toggle button is clicked', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    const toggle = screen.getByLabelText('Collapse sidebar')
    fireEvent.click(toggle)

    // In collapsed mode, text labels are hidden or truncated
    // Check if the expand button is now visible
    expect(screen.getByLabelText('Expand sidebar')).toBeDefined()
  })
})
