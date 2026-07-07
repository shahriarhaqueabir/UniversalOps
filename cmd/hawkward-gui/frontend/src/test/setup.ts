import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Wails runtime
const mockWails = {
  go: {
    app: {
      App: {
        GetAppInfo: vi.fn().mockResolvedValue({ name: 'Hawkward', version: '1.0.0' }),
      },
      SysOps: {
        GetCPUInfo: vi.fn(),
        GetMemoryInfo: vi.fn(),
        GetDiskInfo: vi.fn(),
      },
    },
  },
  runtime: {
    EventsOn: vi.fn(),
    EventsOff: vi.fn(),
  },
}

Object.defineProperty(window, 'go', { value: mockWails.go })
Object.defineProperty(window, 'runtime', { value: mockWails.runtime })
