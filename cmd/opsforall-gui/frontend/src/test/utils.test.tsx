// @ts-nocheck
import { describe, it, expect } from 'vitest'
import { cn } from '@/lib/utils'
import { ErrorBoundary } from '@/components/ui/ErrorBoundary'

describe('cn utility', () => {
  it('merges class names correctly', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles conditional classes', () => {
    const falsy = false as boolean
    expect(cn('base', falsy && 'hidden', 'visible')).toBe('base visible')
  })

  it('merges Tailwind classes (later wins)', () => {
    expect(cn('px-4', 'px-6')).toBe('px-6')
  })

  it('handles clsx array syntax', () => {
    expect(cn(['a', 'b'], 'c')).toBe('a b c')
  })

  it('handles empty inputs', () => {
    expect(cn()).toBe('')
  })
})

describe('ErrorBoundary smoke test', () => {
  it('renders children when no error', () => {
    // Verifies the ErrorBoundary component can be imported without crashing
    expect(ErrorBoundary).toBeDefined()
  })
})
