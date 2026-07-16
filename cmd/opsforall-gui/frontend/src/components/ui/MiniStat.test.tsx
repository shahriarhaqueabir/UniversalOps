import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MiniStat } from './MiniStat'
import { Activity } from 'lucide-react'

describe('MiniStat', () => {
  it('renders label and value', () => {
    render(<MiniStat label="Uptime" value="99.9%" />)
    expect(screen.getByText('Uptime')).toBeInTheDocument()
    expect(screen.getByText('99.9%')).toBeInTheDocument()
  })

  it('renders unit when provided', () => {
    render(<MiniStat label="Speed" value={42} unit="ms" />)
    expect(screen.getByText('ms')).toBeInTheDocument()
  })

  it('does not render unit when omitted', () => {
    render(<MiniStat label="Speed" value={42} />)
    expect(screen.queryByText('ms')).not.toBeInTheDocument()
  })

  it('renders icon when provided', () => {
    render(<MiniStat label="Test" value={0} icon={<Activity data-testid="icon" />} />)
    expect(screen.getByTestId('icon')).toBeInTheDocument()
  })

  it('applies variant classes', () => {
    const { container } = render(<MiniStat label="Test" value={0} variant="danger" />)
    const iconContainer = container.querySelector('.w-14.h-14')
    expect(iconContainer?.className).toContain('var(--color-danger)')
  })

  it('accepts custom className', () => {
    const { container } = render(<MiniStat label="Test" value={0} className="extra" />)
    expect(container.firstChild).toHaveClass('extra')
  })
})
