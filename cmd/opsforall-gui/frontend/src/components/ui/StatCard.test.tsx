import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatCard } from './StatCard'
import { Cpu } from 'lucide-react'

describe('StatCard', () => {
  it('renders label and value', () => {
    render(<StatCard label="Cores" value={8} />)
    expect(screen.getByText('Cores')).toBeInTheDocument()
    expect(screen.getByText('8')).toBeInTheDocument()
  })

  it('renders string values', () => {
    render(<StatCard label="Status" value="Healthy" />)
    expect(screen.getByText('Healthy')).toBeInTheDocument()
  })

  it('renders icon when provided', () => {
    render(<StatCard label="Test" value={0} icon={<Cpu data-testid="icon" />} />)
    expect(screen.getByTestId('icon')).toBeInTheDocument()
  })

  it('does not render icon slot when omitted', () => {
    const { container } = render(<StatCard label="Test" value={0} />)
    const iconSlot = container.querySelector('span')
    expect(iconSlot).not.toBeInTheDocument()
  })

  it('accepts custom className', () => {
    const { container } = render(<StatCard label="Test" value={0} className="extra" />)
    expect(container.firstChild).toHaveClass('extra')
  })

  it('applies valueClassName to value text', () => {
    render(<StatCard label="Used" value="4.2 GB" valueClassName="text-[var(--color-success)]" />)
    const valueEl = screen.getByText('4.2 GB')
    expect(valueEl.className).toContain('color-success')
  })
})
