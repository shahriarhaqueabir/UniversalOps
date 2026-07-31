import type { UseQueryResult } from '@tanstack/react-query'
import { vi } from 'vitest'

/**
 * Lightweight helper for TanStack Query mock return values.
 * Eliminates `as any` casts in test files while keeping boilerplate minimal.
 *
 * @example
 * ```ts
 * import { mockQueryReturn } from '@/test/mockQuery'
 * vi.mocked(useQuery).mockReturnValue(mockQueryReturn({ data: ['a', 'b'] }))
 * ```
 */
export function mockQueryReturn<T>(
  overrides: { data: T } & Record<string, unknown>,
): UseQueryResult<T, Error> {
  const base: Omit<UseQueryResult<T, Error>, 'data'> & { data: T } = {
    data: overrides.data,
    isLoading: false,
    error: null,
    isError: false,
    isSuccess: true,
    isFetching: false,
    status: 'success',
    dataUpdatedAt: Date.now(),
    errorUpdatedAt: 0,
    failureCount: 0,
    failureReason: null,
    errorUpdateCount: 0,
    isFetched: true,
    isFetchedAfterMount: true,
    isPlaceholderData: false,
    isPreviousData: false,
    isRefetching: false,
    isStale: false,
    isInitialLoading: false,
    isLoadingError: false,
    isRefetchError: false,
    isPaused: false,
    refetch: vi.fn(),
  }
  return { ...base, ...overrides } as UseQueryResult<T, Error>
}
