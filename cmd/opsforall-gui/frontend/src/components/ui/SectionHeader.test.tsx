import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SectionHeader } from './SectionHeader'
import { Wifi, Globe } from 'lucide-react'

describe('SectionHeader', () => {
  it('renders title', () => {
    render(<SectionHeader icon={Wifi} title="Interfaces" />)
    expect(screen.getByText('Interfaces')).toBeInTheDocument()
  })

  it('renders count badge when provided', () => {
    render(<SectionHeader icon={Wifi} title="Interfaces" count={5} />)
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('does not render count badge when omitted', () => {
    render(<SectionHeader icon={Wifi} title="Interfaces" />)
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('renders zero count when explicitly set', () => {
    render(<SectionHeader icon={Wifi} title="Interfaces" count={0} />)
    expect(screen.getByText('0')).toBeInTheDocument()
  })

  it('accepts custom className', () => {
    const { container } = render(<SectionHeader icon={Globe} title="Test" className="extra" />)
    expect(container.firstChild).toHaveClass('extra')
  })
})
