import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { NetworkDesign } from './NetworkDesign'

let nanoidCounter = 0
vi.mock('nanoid', () => ({
  nanoid: vi.fn(() => {
    nanoidCounter++
    return nanoidCounter.toString().padStart(8, '0')
  }),
}))

vi.mock('@/components/network/DeviceNode', () => ({
  DeviceNode: ({ device, isSelected, isConnectMode, onSelect, onStartConnect }: {
    device: { id: string; label: string; type: string; status: string }
    isSelected: boolean
    isConnectMode: boolean
    onSelect: (id: string) => void
    onStartConnect: (id: string) => void
  }) => (
    <div
      data-testid={`device-node-${device.id}`}
      data-selected={isSelected}
      data-connect-mode={isConnectMode}
      onClick={() => {
        if (isConnectMode) onStartConnect(device.id)
        else onSelect(device.id)
      }}
    >
      {device.label}
    </div>
  ),
}))

vi.mock('@/components/network/ConnectionLine', () => ({
  ConnectionLine: ({ connection: conn }: { connection: { id: string; label: string } }) => (
    <div data-testid={`conn-line-${conn.id}`}>{conn.label}</div>
  ),
}))

vi.mock('./networkdesign/AnalysisSidebar', () => ({
  AnalysisSidebar: ({ isOpen, onToggle }: { isOpen: boolean; onToggle: () => void }) => (
    <div data-testid="analysis-sidebar" data-open={isOpen}>
      <button onClick={onToggle} data-testid="analysis-toggle-inner">Toggle</button>
    </div>
  ),
}))

vi.mock('../../wailsjs/go/app/App', () => ({
  SaveFileDialog: vi.fn().mockResolvedValue('saved.json'),
  OpenFileDialog: vi.fn().mockResolvedValue('loaded.json'),
}))

vi.mock('../../wailsjs/go/app/DevOps', () => ({
  WriteFile: vi.fn().mockResolvedValue(true),
  ReadFile: vi.fn().mockResolvedValue(JSON.stringify({
    devices: [{ id: 'dev-loaded', type: 'router', label: 'Loaded Router', x: 100, y: 100, ip: '10.0.0.1', subnet: '255.255.255.0', mac: '00:00:00:00:00:01', status: 'healthy' }],
    connections: [],
  })),
}))

describe('NetworkDesign Page', () => {
  beforeEach(() => {
    nanoidCounter = 0
    window.localStorage.clear()
  })

  it('renders page header', () => {
    render(<NetworkDesign />)
    expect(screen.getByText(/Network Topology Designer/i)).toBeInTheDocument()
  })

  it('renders toolbar with all device type buttons', () => {
    render(<NetworkDesign />)
    expect(screen.getByTitle(/Add Router/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Add Switch/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Add Server/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Add Workstation/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Add Firewall/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Add Cloud/i)).toBeInTheDocument()
  })

  it('renders default 7 devices', () => {
    render(<NetworkDesign />)
    expect(screen.getByText(/Core Router/i)).toBeInTheDocument()
    expect(screen.getAllByText(/Firewall/i).length).toBe(2)
    expect(screen.getByText(/Core Switch/i)).toBeInTheDocument()
    expect(screen.getByText(/Web Server/i)).toBeInTheDocument()
    expect(screen.getByText(/DB Server/i)).toBeInTheDocument()
    expect(screen.getByText(/Dev Workstation/i)).toBeInTheDocument()
    expect(screen.getByText(/AWS Cloud/i)).toBeInTheDocument()
  })

  it('shows stats bar with device and connection counts', () => {
    render(<NetworkDesign />)
    expect(screen.getByText(/7 devices/i)).toBeInTheDocument()
    expect(screen.getByText(/6 connections/i)).toBeInTheDocument()
  })

  it('switches to Connect mode', () => {
    render(<NetworkDesign />)
    const connectBtn = screen.getByTitle(/Connect devices/i)
    fireEvent.click(connectBtn)
    expect(connectBtn.className).toContain('bg-success')
  })

  it('switches to Delete mode', () => {
    render(<NetworkDesign />)
    const deleteBtn = screen.getByTitle(/Delete mode/i)
    fireEvent.click(deleteBtn)
    expect(deleteBtn.className).toContain('bg-danger')
  })

  it('switches back to Select mode after Connect', () => {
    render(<NetworkDesign />)
    fireEvent.click(screen.getByTitle(/Connect devices/i))
    const selectBtn = screen.getByTitle(/Select \/ Move/i)
    fireEvent.click(selectBtn)
    expect(selectBtn.className).toContain('bg-accent')
  })

  it('adds a new device when toolbar button is clicked', () => {
    render(<NetworkDesign />)
    const addRouterBtn = screen.getByTitle(/Add Router/i)
    fireEvent.click(addRouterBtn)
    expect(screen.getByText(/8 devices/i)).toBeInTheDocument()
  })

  it('shows zoom percentage', () => {
    render(<NetworkDesign />)
    expect(screen.getByText(/100%/i)).toBeInTheDocument()
  })

  it('zoom in increases zoom', () => {
    render(<NetworkDesign />)
    fireEvent.click(screen.getByTitle(/Zoom In/i))
    expect(screen.getByText(/110%/i)).toBeInTheDocument()
  })

  it('zoom out decreases zoom', () => {
    render(<NetworkDesign />)
    fireEvent.click(screen.getByTitle(/Zoom Out/i))
    expect(screen.getByText(/90%/i)).toBeInTheDocument()
  })

  it('zoom reset goes back to 100%', () => {
    render(<NetworkDesign />)
    fireEvent.click(screen.getByTitle(/Zoom In/i))
    fireEvent.click(screen.getByTitle(/Zoom In/i))
    expect(screen.getByText(/120%/i)).toBeInTheDocument()
    fireEvent.click(screen.getByTitle(/Reset Zoom/i))
    expect(screen.getByText(/100%/i)).toBeInTheDocument()
  })

  it('exports JSON via web download', () => {
    const createObjectURL = vi.fn(() => 'blob:url')
    const revokeObjectURL = vi.fn()
    URL.createObjectURL = createObjectURL
    URL.revokeObjectURL = revokeObjectURL

    render(<NetworkDesign />)
    const exportBtn = screen.getByTitle(/Export JSON/i)
    fireEvent.click(exportBtn)
    expect(createObjectURL).toHaveBeenCalled()
  })

  it('clears canvas and shows empty state', () => {
    render(<NetworkDesign />)
    fireEvent.click(screen.getByText(/Clear/i))
    expect(screen.getByText(/0 devices/i)).toBeInTheDocument()
    expect(screen.getByText(/Canvas is empty/i)).toBeInTheDocument()
  })

  it('shows selected device label in stats bar', () => {
    render(<NetworkDesign />)
    const coreRouter = screen.getByText(/Core Router/i)
    fireEvent.click(coreRouter)
    expect(screen.getByText(/Selected:/i)).toBeInTheDocument()
  })

  it('shows connect mode hint when source is selected', () => {
    render(<NetworkDesign />)
    fireEvent.click(screen.getByTitle(/Connect devices/i))
    const coreRouter = screen.getByText(/Core Router/i)
    fireEvent.click(coreRouter)
    expect(screen.getByText(/Click a target device to connect/i)).toBeInTheDocument()
  })

  it('toggles analysis sidebar', () => {
    render(<NetworkDesign />)
    expect(screen.getByTestId('analysis-sidebar')).toHaveAttribute('data-open', 'true')
    fireEvent.click(screen.getByTitle(/Toggle Analysis Panel/i))
    expect(screen.getByTestId('analysis-sidebar')).toHaveAttribute('data-open', 'false')
  })

  it('reopens analysis sidebar when toggled again', () => {
    render(<NetworkDesign />)
    const toggleBtn = screen.getByTitle(/Toggle Analysis Panel/i)
    fireEvent.click(toggleBtn)
    expect(screen.getByTestId('analysis-sidebar')).toHaveAttribute('data-open', 'false')
    fireEvent.click(toggleBtn)
    expect(screen.getByTestId('analysis-sidebar')).toHaveAttribute('data-open', 'true')
  })

  it('saves topology to localStorage on change', async () => {
    render(<NetworkDesign />)
    const addBtn = screen.getByTitle(/Add Router/i)
    fireEvent.click(addBtn)
    await waitFor(() => {
      const saved = window.localStorage.getItem('opsforall-topology')
      expect(saved).not.toBeNull()
      const parsed = JSON.parse(saved!)
      expect(parsed.devices.length).toBe(8)
    })
  })

  it('loads topology from localStorage on mount', () => {
    const savedData = {
      devices: [
        { id: 'dev-saved', type: 'router', label: 'Saved Router', x: 100, y: 100, ip: '10.0.0.1', subnet: '255.255.255.0', mac: '00:00:00:00:00:01', status: 'healthy' },
      ],
      connections: [],
    }
    window.localStorage.setItem('opsforall-topology', JSON.stringify(savedData))
    render(<NetworkDesign />)
    expect(screen.getByText(/Saved Router/i)).toBeInTheDocument()
    expect(screen.getByText(/1 device/i)).toBeInTheDocument()
    expect(screen.getByText(/0 connections/i)).toBeInTheDocument()
  })

  it('handles save topology file dialog', async () => {
    const { SaveFileDialog } = await import('../../wailsjs/go/app/App')
    const { WriteFile } = await import('../../wailsjs/go/app/DevOps')
    render(<NetworkDesign />)
    const saveBtn = screen.getByTitle(/Save Topology/i)
    fireEvent.click(saveBtn)
    await waitFor(() => {
      expect(SaveFileDialog).toHaveBeenCalled()
      expect(WriteFile).toHaveBeenCalled()
    })
  })

  it('handles load topology file dialog', async () => {
    const { OpenFileDialog } = await import('../../wailsjs/go/app/App')
    const { ReadFile } = await import('../../wailsjs/go/app/DevOps')
    render(<NetworkDesign />)
    const loadBtn = screen.getByTitle(/Load Topology/i)
    fireEvent.click(loadBtn)
    await waitFor(() => {
      expect(OpenFileDialog).toHaveBeenCalled()
      expect(ReadFile).toHaveBeenCalled()
      expect(screen.getByText(/Loaded Router/i)).toBeInTheDocument()
    })
  })

  it('adds multiple device types', () => {
    render(<NetworkDesign />)
    fireEvent.click(screen.getByTitle(/Add Router/i))
    fireEvent.click(screen.getByTitle(/Add Switch/i))
    fireEvent.click(screen.getByTitle(/Add Server/i))
    expect(screen.getByText(/10 devices/i)).toBeInTheDocument()
  })
})
