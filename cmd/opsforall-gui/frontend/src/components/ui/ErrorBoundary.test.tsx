// @ts-nocheck
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ErrorBoundary } from './ErrorBoundary'

const ThrowError = () => {
  throw new Error('Test error')
}

describe('ErrorBoundary', () => {
  it('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <div>Safe Content</div>
      </ErrorBoundary>,
    )
    expect(screen.getByText('Safe Content')).toBeInTheDocument()
  })

  it('catches errors and shows fallback UI', () => {
    // Suppress console.error from the caught error
    vi.spyOn(console, 'error').mockImplementation(() => { })

    render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>,
    )

    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText('Error: Test error')).toBeInTheDocument()
    expect(screen.getByText('Reload Application')).toBeInTheDocument()
  })
})
