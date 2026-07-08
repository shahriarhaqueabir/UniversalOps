import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Settings } from './Settings'

// Mock hooks
const mockToggle = vi.fn()
vi.mock('../hooks/useTheme', () => ({
  useTheme: () => ({ theme: 'dark', toggle: mockToggle }),
}))

vi.mock('../hooks/useBackend', () => ({
  useBackend: () => ({ call: vi.fn().mockResolvedValue({ name: 'Hawkward', version: '1.0.0', goVersion: '1.26', uptime: '2h' }) }),
}))

describe('Settings Page', () => {
  it('renders settings heading', () => {
    render(<Settings />)
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('renders appearance section', () => {
    render(<Settings />)
    expect(screen.getByText('Appearance')).toBeInTheDocument()
  })

  it('theme toggle triggers useTheme toggle', () => {
    render(<Settings />)
    const lightBtn = screen.getByText('Light')
    fireEvent.click(lightBtn)
    expect(mockToggle).toHaveBeenCalledTimes(1)
  })
})
