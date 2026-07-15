import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useBackend } from './useBackend'

describe('useBackend', () => {
  let originalGo: unknown

  beforeEach(() => {
    originalGo = (window as any).go
    ;(window as any).go = { app: {} }
  })

  afterEach(() => {
    ;(window as any).go = originalGo
  })

  it('returns a call function', () => {
    const { result } = renderHook(() => useBackend())
    expect(result.current.call).toBeDefined()
    expect(typeof result.current.call).toBe('function')
  })

  it('calls a Wails method successfully', async () => {
    const mockFn = vi.fn().mockResolvedValue({ name: 'test', version: '1.0' })
    ;(window as any).go.app.App = { GetAppInfo: mockFn }

    const { result } = renderHook(() => useBackend())
    const response = await result.current.call('App.GetAppInfo')

    expect(mockFn).toHaveBeenCalledOnce()
    expect(response).toEqual({ name: 'test', version: '1.0' })
  })

  it('passes arguments to the Wails method', async () => {
    const mockFn = vi.fn().mockResolvedValue('ok')
    ;(window as any).go.app.SysOps = { GetCPUInfo: mockFn }

    const { result } = renderHook(() => useBackend())
    await result.current.call('SysOps.GetCPUInfo', 1, 'arg')

    expect(mockFn).toHaveBeenCalledWith(1, 'arg')
  })

  it('throws when method is not found', async () => {
    const { result } = renderHook(() => useBackend())
    await expect(
      result.current.call('Nonexistent.DoSomething')
    ).rejects.toThrow('Wails method not found')
  })

  it('throws on Wails error', async () => {
    const mockFn = vi.fn().mockRejectedValue(new Error('backend failure'))
    ;(window as any).go.app.SysOps = { GetCPUInfo: mockFn }

    const { result } = renderHook(() => useBackend())
    await expect(result.current.call('SysOps.GetCPUInfo')).rejects.toThrow('backend failure')
  })
})
