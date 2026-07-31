import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EmptyState } from './EmptyState'

describe('EmptyState', () => {
  const defaultProps = {
    icon: <span data-testid="test-icon">🔍</span>,
    title: 'No Results Found',
    description: 'Try adjusting your search or filter criteria.',
  }

  it('renders the title', () => {
    render(<EmptyState {...defaultProps} />)
    expect(screen.getByText('No Results Found')).toBeInTheDocument()
  })

  it('renders the description', () => {
    render(<EmptyState {...defaultProps} />)
    expect(screen.getByText('Try adjusting your search or filter criteria.')).toBeInTheDocument()
  })

  it('renders the icon', () => {
    render(<EmptyState {...defaultProps} />)
    expect(screen.getByTestId('test-icon')).toBeInTheDocument()
  })

  it('does not render action button when no action prop', () => {
    render(<EmptyState {...defaultProps} />)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('renders action button with label when action prop is provided', () => {
    render(<EmptyState {...defaultProps} action={{ label: 'Retry', onClick: () => {} }} />)
    const button = screen.getByRole('button')
    expect(button).toBeInTheDocument()
    expect(button).toHaveTextContent('Retry')
  })

  it('calls action onClick when button is clicked', async () => {
    const onClick = vi.fn()
    render(<EmptyState {...defaultProps} action={{ label: 'Retry', onClick }} />)
    const user = userEvent.setup()
    await user.click(screen.getByRole('button'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })
})
