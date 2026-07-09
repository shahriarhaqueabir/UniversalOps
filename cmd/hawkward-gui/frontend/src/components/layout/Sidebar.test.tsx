import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Sidebar } from './Sidebar'

describe('Sidebar', () => {
  it('renders all navigation items', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    expect(screen.getByText('HAWKWARD')).toBeInTheDocument()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('System Ops')).toBeInTheDocument()
    expect(screen.getByText('Network Ops')).toBeInTheDocument()
    expect(screen.getByText('Security Ops')).toBeInTheDocument()
    expect(screen.getByText('DevOps')).toBeInTheDocument()
    expect(screen.getByText('AI Ops')).toBeInTheDocument()
    expect(screen.getByText('Network Design')).toBeInTheDocument()
    expect(screen.getByText('Logs')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('highlights the current page', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="sysops" onNavigate={onNavigate} />)

    const sysopsBtn = screen.getByText('System Ops').closest('button')
    expect(sysopsBtn).toHaveClass('bg-accent')
  })

  it('calls onNavigate when clicking a nav item', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    fireEvent.click(screen.getByText('Logs'))
    expect(onNavigate).toHaveBeenCalledWith('logs')
  })

  it('shows version footer', () => {
    render(<Sidebar currentPage="dashboard" onNavigate={vi.fn()} />)

    expect(screen.getByText('Hawkward v1.3.0')).toBeInTheDocument()
  })

  it('collapses and expands when toggle is clicked', () => {
    const onNavigate = vi.fn()
    render(<Sidebar currentPage="dashboard" onNavigate={onNavigate} />)

    const collapseBtn = screen.getByLabelText('Collapse sidebar')
    fireEvent.click(collapseBtn)

    // After collapse, labels should be hidden
    expect(screen.queryByText('Dashboard')).not.toBeInTheDocument()

    // Toggle button should now say expand
    expect(screen.getByLabelText('Expand sidebar')).toBeInTheDocument()
  })
})
