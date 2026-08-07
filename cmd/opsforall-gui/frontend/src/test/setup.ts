import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Wails runtime
const mockWails = {
  go: {
    app: {
      App: {
        GetAppInfo: vi.fn().mockResolvedValue({ name: 'UniversalOps', version: '1.6.3' }),
      },
      SysOps: {
        GetCPUInfo: vi.fn(),
        GetMemoryInfo: vi.fn(),
        GetDiskInfo: vi.fn(),
      },
      Dashboard: {
        GetDashboardData: vi.fn(),
        RunQuickDiag: vi.fn(),
        GenerateDashboardBriefing: vi.fn(),
      }
    },
  },
  runtime: {
    EventsOn: vi.fn(),
    EventsOff: vi.fn(),
  },
}

Object.defineProperty(window, 'go', { value: mockWails.go, writable: true, configurable: true })
Object.defineProperty(window, 'runtime', { value: mockWails.runtime, writable: true, configurable: true })

// Force a consistent localStorage mock for vitest/jsdom
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value.toString()
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
    key: vi.fn((index: number) => Object.keys(store)[index] || null),
    get length() {
      return Object.keys(store).length
    },
  }
})()

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
})

// Polyfill ResizeObserver for Radix UI components
class ResizeObserverMock {
  observe() { }
  unobserve() { }
  disconnect() { }
}
window.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver
