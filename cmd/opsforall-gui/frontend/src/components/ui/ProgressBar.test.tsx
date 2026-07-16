import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ProgressBar } from './ProgressBar'

describe('ProgressBar', () => {
  it('renders label when provided', () => {
    render(<ProgressBar value={50} label="CPU" />)
    expect(screen.getByText('CPU')).toBeInTheDocument()
  })

  it('renders value when showValue is true', () => {
    render(<ProgressBar value={75.3} />)
    expect(screen.getByText('75.3%')).toBeInTheDocument()
  })

  it('hides value when showValue is false', () => {
    render(<ProgressBar value={50} showValue={false} />)
    expect(screen.queryByText('50.0%')).not.toBeInTheDocument()
  })

  it('renders custom unit', () => {
    render(<ProgressBar value={42} unit="ms" />)
    expect(screen.getByText('42.0ms')).toBeInTheDocument()
  })

  it('auto-selects danger variant for high values', () => {
    render(<ProgressBar value={80} />)
    const bar = document.querySelector('.rounded-full.transition-all')
    expect(bar?.getAttribute('style')).toContain('var(--color-danger)')
  })

  it('auto-selects success variant for low values', () => {
    render(<ProgressBar value={10} />)
    const bar = document.querySelector('.rounded-full.transition-all')
    expect(bar?.getAttribute('style')).toContain('var(--color-success)')
  })

  it('respects explicit variant over auto', () => {
    render(<ProgressBar value={10} variant="danger" />)
    const bar = document.querySelector('.rounded-full.transition-all')
    expect(bar?.getAttribute('style')).toContain('var(--color-danger)')
  })

  it('clamps value to max', () => {
    render(<ProgressBar value={150} max={100} />)
    const bar = document.querySelector('.rounded-full.transition-all')
    expect(bar?.getAttribute('style')).toContain('width: 100%')
  })

  it('accepts custom className', () => {
    const { container } = render(<ProgressBar value={50} className="extra" />)
    expect(container.firstChild).toHaveClass('extra')
  })
})
