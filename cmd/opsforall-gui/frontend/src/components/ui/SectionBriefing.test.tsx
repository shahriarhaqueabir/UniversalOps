import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SectionBriefing } from './SectionBriefing'

describe('SectionBriefing', () => {
  const defaultProps = {
    title: 'Test Title',
    objective: 'Test objective text',
    checklist: ['Item one', 'Item two', 'Item three'],
  }

  it('renders title', () => {
    render(<SectionBriefing {...defaultProps} />)
    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('renders objective', () => {
    render(<SectionBriefing {...defaultProps} />)
    expect(screen.getByText('Test objective text')).toBeInTheDocument()
  })

  it('renders all checklist items', () => {
    render(<SectionBriefing {...defaultProps} />)
    expect(screen.getByText('Item one')).toBeInTheDocument()
    expect(screen.getByText('Item two')).toBeInTheDocument()
    expect(screen.getByText('Item three')).toBeInTheDocument()
  })

  it('renders with empty checklist', () => {
    render(<SectionBriefing {...defaultProps} checklist={[]} />)
    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('accepts custom className', () => {
    const { container } = render(<SectionBriefing {...defaultProps} className="extra" />)
    expect(container.firstChild).toHaveClass('extra')
  })
})
