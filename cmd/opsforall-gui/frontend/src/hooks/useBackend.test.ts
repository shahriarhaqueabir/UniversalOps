import { describe, it, expect, vi } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useBackend } from './useBackend'

describe('useBackend', () => {
  it('returns a call function', () => {
    const { result } = renderHook(() => useBackend())
    expect(result.current.call).toBeDefined()
    expect(typeof result.current.call).toBe('function')
  })

  it('calls a Wails method successfully', async () => {
    const mockFn = vi.fn().mockResolvedValue({ name: 'test', version: '1.0' })
    window.go.app.App.GetAppInfo = mockFn

    const { result } = renderHook(() => useBackend())
    const response = await result.current.call('App.GetAppInfo')

    expect(mockFn).toHaveBeenCalledOnce()
    expect(response).toEqual({ name: 'test', version: '1.0' })
  })

  it('passes arguments to the Wails method', async () => {
    const mockFn = vi.fn().mockResolvedValue('ok')
    window.go.app.SysOps.GetCPUInfo = mockFn

    const { result } = renderHook(() => useBackend())
    await result.current.call('SysOps.GetCPUInfo', 1, 'arg')

    expect(mockFn).toHaveBeenCalledWith(1, 'arg')
  })

  it('returns null when method is not found', async () => {
    const { result } = renderHook(() => useBackend())
    const response = await result.current.call('Nonexistent.DoSomething')
    expect(response).toBeNull()
  })

  it('throws on Wails error', async () => {
    const mockFn = vi.fn().mockRejectedValue(new Error('backend failure'))
    window.go.app.SysOps.GetCPUInfo = mockFn

    const { result } = renderHook(() => useBackend())
    await expect(result.current.call('SysOps.GetCPUInfo')).rejects.toThrow('backend failure')
  })
})
