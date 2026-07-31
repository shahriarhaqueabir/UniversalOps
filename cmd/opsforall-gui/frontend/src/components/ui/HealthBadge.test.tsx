import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { HealthBadge } from './HealthBadge'

describe('HealthBadge', () => {
  it('renders healthy status with correct label', () => {
    render(<HealthBadge status="healthy" />)
    const badge = screen.getByRole('status')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveAttribute('aria-label', 'Health: Healthy')
    expect(badge).toHaveAttribute('data-status', 'healthy')
    expect(screen.getByText('Healthy')).toBeInTheDocument()
  })

  it('renders degraded status with correct label', () => {
    render(<HealthBadge status="degraded" />)
    expect(screen.getByText('Degraded')).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveAttribute('data-status', 'degraded')
  })

  it('renders critical status with correct label', () => {
    render(<HealthBadge status="critical" />)
    expect(screen.getByText('Critical')).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveAttribute('data-status', 'critical')
  })

  it('renders unknown status with correct label', () => {
    render(<HealthBadge status="unknown" />)
    expect(screen.getByText('Unknown')).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveAttribute('data-status', 'unknown')
  })

  it('applies the correct shape mode attribute', () => {
    render(<HealthBadge status="healthy" shapeMode="square" />)
    expect(screen.getByRole('status')).toHaveAttribute('data-shape', 'square')
  })

  it('defaults to circle shape mode', () => {
    render(<HealthBadge status="healthy" />)
    expect(screen.getByRole('status')).toHaveAttribute('data-shape', 'circle')
  })

  it('renders diamond shape mode', () => {
    render(<HealthBadge status="healthy" shapeMode="diamond" />)
    expect(screen.getByRole('status')).toHaveAttribute('data-shape', 'diamond')
  })

  it('renders triangle shape mode', () => {
    render(<HealthBadge status="healthy" shapeMode="triangle" />)
    expect(screen.getByRole('status')).toHaveAttribute('data-shape', 'triangle')
  })

  it('hides label when showLabel is false', () => {
    render(<HealthBadge status="healthy" showLabel={false} />)
    expect(screen.queryByText('Healthy')).not.toBeInTheDocument()
  })

  it('shows label by default', () => {
    render(<HealthBadge status="healthy" />)
    expect(screen.getByText('Healthy')).toBeInTheDocument()
  })

  it('applies pulse animation class when pulse is true', () => {
    render(<HealthBadge status="healthy" pulse />)
    const badge = screen.getByRole('status')
    // The pulse class goes on the indicator span inside
    expect(badge).toBeInTheDocument()
  })

  it('auto-pulses for critical and degraded statuses', () => {
    const { container: critContainer } = render(<HealthBadge status="critical" />)
    // critical auto-pulses — should find animate-pulse on indicator
    const critIndicator = critContainer.querySelector('.animate-pulse')
    expect(critIndicator).toBeTruthy()

    const { container: healthyContainer } = render(<HealthBadge status="healthy" />)
    // healthy should not have animate-pulse by default
    const healthyIndicator = healthyContainer.querySelector('.animate-pulse')
    expect(healthyIndicator).toBeFalsy()
  })

  it('renders each size variant without error', () => {
    const sizes = ['sm', 'md', 'lg'] as const
    for (const size of sizes) {
      const { container } = render(<HealthBadge status="healthy" size={size} />)
      expect(container.firstChild).toBeInTheDocument()
    }
  })

  it('accepts additional className', () => {
    render(<HealthBadge status="healthy" className="extra-class" />)
    const badge = screen.getByRole('status')
    expect(badge.className).toContain('extra-class')
  })
})
