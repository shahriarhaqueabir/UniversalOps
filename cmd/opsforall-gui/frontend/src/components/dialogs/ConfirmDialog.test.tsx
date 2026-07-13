// @ts-nocheck
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ConfirmDialog } from './ConfirmDialog'

describe('ConfirmDialog', () => {
  it('renders nothing when closed', () => {
    const { container } = render(
      <ConfirmDialog
        open={false}
        title="Test"
        description="Test description"
        onConfirm={vi.fn()}
        onClose={vi.fn()}
      />,
    )
    expect(container.innerHTML).toBe('')
  })

  it('renders title and description when open', () => {
    render(
      <ConfirmDialog
        open={true}
        title="Delete Item"
        description="Are you sure you want to delete this item?"
        onConfirm={vi.fn()}
        onClose={vi.fn()}
      />,
    )
    expect(screen.getByText('Delete Item')).toBeInTheDocument()
    expect(screen.getByText('Are you sure you want to delete this item?')).toBeInTheDocument()
  })

  it('calls onConfirm when confirm button is clicked', () => {
    const onConfirm = vi.fn()
    const onClose = vi.fn()
    render(
      <ConfirmDialog
        open={true}
        title="Confirm"
        description="Proceed?"
        confirmText="Yes"
        onConfirm={onConfirm}
        onClose={onClose}
      />,
    )
    fireEvent.click(screen.getByText('Yes'))
    expect(onConfirm).toHaveBeenCalledTimes(1)
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('calls onClose when cancel button is clicked', () => {
    const onClose = vi.fn()
    render(
      <ConfirmDialog
        open={true}
        title="Confirm"
        description="Cancel? "
        cancelText="Cancel"
        onConfirm={vi.fn()}
        onClose={onClose}
      />,
    )
    fireEvent.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('renders with danger type styling', () => {
    render(
      <ConfirmDialog
        open={true}
        title="Danger"
        description="This is dangerous"
        type="danger"
        onConfirm={vi.fn()}
        onClose={vi.fn()}
      />,
    )
    expect(screen.getByText('Danger')).toBeInTheDocument()
  })
})
