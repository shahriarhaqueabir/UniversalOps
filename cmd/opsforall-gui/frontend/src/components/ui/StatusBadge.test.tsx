import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatusBadge } from './StatusBadge'

describe('StatusBadge', () => {
  it('renders status text', () => {
    render(<StatusBadge status="running" />)
    expect(screen.getByText('running')).toBeInTheDocument()
  })

  it('applies success classes for "running"', () => {
    render(<StatusBadge status="running" />)
    const badge = screen.getByText('running')
    expect(badge.className).toContain('var(--color-success)')
  })

  it('applies danger classes for "stopped"', () => {
    render(<StatusBadge status="stopped" />)
    const badge = screen.getByText('stopped')
    expect(badge.className).toContain('var(--color-danger)')
  })

  it('applies warning classes for "manual"', () => {
    render(<StatusBadge status="manual" />)
    const badge = screen.getByText('manual')
    expect(badge.className).toContain('var(--color-warning)')
  })

  it('applies accent classes for "info"', () => {
    render(<StatusBadge status="info" />)
    const badge = screen.getByText('info')
    expect(badge.className).toContain('var(--color-accent)')
  })

  it('applies fallback classes for unknown status', () => {
    render(<StatusBadge status="unknown" />)
    const badge = screen.getByText('unknown')
    expect(badge.className).toContain('var(--color-text-faint)')
  })

  it('applies sm size classes', () => {
    render(<StatusBadge status="running" size="sm" />)
    const badge = screen.getByText('running')
    expect(badge.className).toContain('px-2')
  })

  it('applies md size classes by default', () => {
    render(<StatusBadge status="running" />)
    const badge = screen.getByText('running')
    expect(badge.className).toContain('px-3')
  })

  it('replaces underscores with spaces', () => {
    render(<StatusBadge status="not_running" />)
    expect(screen.getByText('not running')).toBeInTheDocument()
  })

  it('accepts custom className', () => {
    render(<StatusBadge status="running" className="extra-class" />)
    const badge = screen.getByText('running')
    expect(badge.className).toContain('extra-class')
  })
})
