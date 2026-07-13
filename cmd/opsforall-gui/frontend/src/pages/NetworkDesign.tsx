import { useState, useCallback, useRef, useMemo, useEffect } from 'react'
import { nanoid } from 'nanoid'
import { cn } from '@/lib/utils'
import { DeviceNode } from '@/components/network/DeviceNode'
import { ConnectionLine } from '@/components/network/ConnectionLine'
import {
  Server,
  GitMerge,
  HardDrive,
  Monitor,
  Shield,
  Cloud,
  Plus,
  Minus,
  Maximize2,
  Trash2,
  Link2,
  MousePointer2,
  X,
  FileJson,
  Save,
  FolderOpen,
  Radar,
} from 'lucide-react'
import { AnalysisSidebar } from './networkdesign/AnalysisSidebar'
import type { TopologyDevice, TopologyConnection, DeviceType, TopologyStatus, ConnectionType } from '@/types'
import { SaveFileDialog, OpenFileDialog } from '../../wailsjs/go/app/App'
import { WriteFile, ReadFile } from '../../wailsjs/go/app/DevOps'

type CanvasMode = 'select' | 'connect' | 'delete'

const deviceTypes: { type: DeviceType; label: string; icon: React.ReactNode }[] = [
  { type: 'router', label: 'Router', icon: <Server size={16} /> },
  { type: 'switch', label: 'Switch', icon: <GitMerge size={16} /> },
  { type: 'server', label: 'Server', icon: <HardDrive size={16} /> },
  { type: 'workstation', label: 'Workstation', icon: <Monitor size={16} /> },
  { type: 'firewall', label: 'Firewall', icon: <Shield size={16} /> },
  { type: 'cloud', label: 'Cloud', icon: <Cloud size={16} /> },
]

function genId(): string {
  return `dev-${nanoid(8)}`
}

function genConnId(): string {
  return `conn-${nanoid(8)}`
}

// ── Properties Panel ──
function PropertiesPanel({
  device,
  onUpdate,
  onClose,
}: {
  device: TopologyDevice
  onUpdate: (id: string, updates: Partial<TopologyDevice>) => void
  onClose: () => void
}) {
  return (
    <div className="w-72 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg p-4 space-y-3 shrink-0 overflow-y-auto">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-[var(--color-text)]">Device Properties</h3>
        <button onClick={onClose} className="text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors">
          <X size={16} />
        </button>
      </div>

      <div className="space-y-3">
        <div>
          <label className="text-xs text-[var(--color-text-faint)] uppercase tracking-wider font-medium block mb-1.5">Name</label>
          <input
            type="text"
            value={device.label}
            onChange={(e) => onUpdate(device.id, { label: e.target.value })}
            className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded px-2.5 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] focus:outline-none focus:ring-1 focus:ring-[var(--color-accent)]"
          />
        </div>

        <div>
          <label className="text-xs text-[var(--color-text-faint)] uppercase tracking-wider font-medium block mb-1.5">Type</label>
          <p className="text-sm text-[var(--color-text)] font-medium capitalize">{device.type}</p>
        </div>

        <div>
          <label className="text-xs text-[var(--color-text-faint)] uppercase tracking-wider font-medium block mb-1.5">IP Address</label>
          <input
            type="text"
            value={device.ip || ''}
            onChange={(e) => onUpdate(device.id, { ip: e.target.value })}
            placeholder="e.g., 192.168.1.1"
            className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded px-2.5 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] focus:outline-none focus:ring-1 focus:ring-[var(--color-accent)]"
          />
        </div>

        <div>
          <label className="text-xs text-[var(--color-text-faint)] uppercase tracking-wider font-medium block mb-1.5">Subnet</label>
          <input
            type="text"
            value={device.subnet || ''}
            onChange={(e) => onUpdate(device.id, { subnet: e.target.value })}
            placeholder="e.g., 255.255.255.0"
            className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded px-2.5 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] focus:outline-none focus:ring-1 focus:ring-[var(--color-accent)]"
          />
        </div>

        <div>
          <label className="text-[10px] text-text-faint uppercase tracking-wider block mb-1">MAC Address</label>
          <input
            type="text"
            value={device.mac || ''}
            onChange={(e) => onUpdate(device.id, { mac: e.target.value })}
            placeholder="e.g., 00:1A:2B:3C:4D:5E"
            className="w-full bg-panel-2 border border-border rounded px-2.5 py-1.5 text-xs text-text placeholder:text-text-faint focus:outline-none focus:ring-1 focus:ring-accent"
          />
        </div>

        <div>
          <label className="text-[10px] text-text-faint uppercase tracking-wider block mb-1">Status</label>
          <select
            value={device.status}
            onChange={(e) => onUpdate(device.id, { status: e.target.value as TopologyStatus })}
            className="w-full bg-panel-2 border border-border rounded px-2.5 py-1.5 text-xs text-text focus:outline-none focus:ring-1 focus:ring-accent"
          >
            <option value="healthy">Healthy</option>
            <option value="warning">Warning</option>
            <option value="critical">Critical</option>
          </select>
        </div>

        <div>
          <label className="text-[10px] text-text-faint uppercase tracking-wider block mb-1">Notes</label>
          <textarea
            value={device.notes || ''}
            onChange={(e) => onUpdate(device.id, { notes: e.target.value })}
            placeholder="Custom notes..."
            rows={3}
            className="w-full bg-panel-2 border border-border rounded px-2.5 py-1.5 text-xs text-text placeholder:text-text-faint focus:outline-none focus:ring-1 focus:ring-accent resize-none"
          />
        </div>
      </div>
    </div>
  )
}

// ── Connection Properties Panel ──
function ConnectionPropertiesPanel({
  connection,
  devices,
  onUpdate,
  onClose,
}: {
  connection: TopologyConnection
  devices: TopologyDevice[]
  onUpdate: (id: string, updates: Partial<TopologyConnection>) => void
  onClose: () => void
}) {
  const source = devices.find((d) => d.id === connection.sourceId)
  const target = devices.find((d) => d.id === connection.targetId)

  return (
    <div className="w-72 bg-panel border border-border rounded-lg p-4 space-y-3 shrink-0 overflow-y-auto">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-text">Connection Properties</h3>
        <button onClick={onClose} className="text-text-faint hover:text-text transition-colors">
          <X size={16} />
        </button>
      </div>

      <div className="space-y-2.5">
        <div>
          <label className="text-[10px] text-text-faint uppercase tracking-wider block mb-1">Source</label>
          <p className="text-xs text-text font-medium">{source?.label || 'Unknown'}</p>
        </div>

        <div>
          <label className="text-[10px] text-text-faint uppercase tracking-wider block mb-1">Target</label>
          <p className="text-xs text-text font-medium">{target?.label || 'Unknown'}</p>
        </div>

        <div>
          <label className="text-[10px] text-text-faint uppercase tracking-wider block mb-1">Label / Bandwidth</label>
          <input
            type="text"
            value={connection.label || ''}
            onChange={(e) => onUpdate(connection.id, { label: e.target.value })}
            placeholder="e.g., 1 Gbps"
            className="w-full bg-panel-2 border border-border rounded px-2.5 py-1.5 text-xs text-text placeholder:text-text-faint focus:outline-none focus:ring-1 focus:ring-accent"
          />
        </div>

        <div>
          <label className="text-[10px] text-text-faint uppercase tracking-wider block mb-1">Connection Type</label>
          <select
            value={connection.type}
            onChange={(e) => onUpdate(connection.id, { type: e.target.value as ConnectionType })}
            className="w-full bg-panel-2 border border-border rounded px-2.5 py-1.5 text-xs text-text focus:outline-none focus:ring-1 focus:ring-accent"
          >
            <option value="ethernet">Ethernet</option>
            <option value="fiber">Fiber</option>
            <option value="wireless">Wireless</option>
          </select>
        </div>
      </div>
    </div>
  )
}

// ── Default topology seed data ──
function initialDevices(): TopologyDevice[] {
  return [
    { id: genId(), type: 'router', label: 'Core Router', x: 550, y: 40, ip: '10.0.0.1', subnet: '255.255.255.0', mac: '00:1A:2B:00:00:01', status: 'healthy' },
    { id: genId(), type: 'firewall', label: 'Firewall', x: 550, y: 180, ip: '10.0.0.2', subnet: '255.255.255.0', mac: '00:1A:2B:00:00:02', status: 'healthy' },
    { id: genId(), type: 'switch', label: 'Core Switch', x: 550, y: 320, ip: '10.0.0.3', subnet: '255.255.255.0', mac: '00:1A:2B:00:00:03', status: 'healthy' },
    { id: genId(), type: 'server', label: 'Web Server', x: 300, y: 460, ip: '10.0.0.10', subnet: '255.255.255.0', mac: '00:1A:2B:00:00:10', status: 'healthy' },
    { id: genId(), type: 'server', label: 'DB Server', x: 550, y: 460, ip: '10.0.0.11', subnet: '255.255.255.0', mac: '00:1A:2B:00:00:11', status: 'warning' },
    { id: genId(), type: 'workstation', label: 'Dev Workstation', x: 800, y: 460, ip: '10.0.0.20', subnet: '255.255.255.0', mac: '00:1A:2B:00:00:20', status: 'healthy' },
    { id: genId(), type: 'cloud', label: 'AWS Cloud', x: 100, y: 40, ip: '', subnet: '', mac: '', status: 'healthy' },
  ]
}

function initialConnections(devices: TopologyDevice[]): TopologyConnection[] {
  const router = devices[0]
  const firewall = devices[1]
  const sw = devices[2]
  const web = devices[3]
  const db = devices[4]
  const ws = devices[5]
  const cloud = devices[6]
  return [
    { id: genConnId(), sourceId: cloud.id, targetId: router.id, label: '1 Gbps', type: 'fiber' },
    { id: genConnId(), sourceId: router.id, targetId: firewall.id, label: '10 Gbps', type: 'fiber' },
    { id: genConnId(), sourceId: firewall.id, targetId: sw.id, label: '10 Gbps', type: 'ethernet' },
    { id: genConnId(), sourceId: sw.id, targetId: web.id, label: '1 Gbps', type: 'ethernet' },
    { id: genConnId(), sourceId: sw.id, targetId: db.id, label: '1 Gbps', type: 'ethernet' },
    { id: genConnId(), sourceId: sw.id, targetId: ws.id, label: '1 Gbps', type: 'ethernet' },
  ]
}

// ── Persistence helpers ──
const STORAGE_KEY = 'opsforall-topology'

function loadSavedTopology(): { devices: TopologyDevice[]; connections: TopologyConnection[] } | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    return JSON.parse(raw) as { devices: TopologyDevice[]; connections: TopologyConnection[] }
  } catch { /* sync quota error */ return null }
}

function saveTopologyState(devices: TopologyDevice[], connections: TopologyConnection[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ devices, connections }))
  } catch { /* quota exceeded, ignore */ }
}

// ── NetworkDesign Page ──
export function NetworkDesign() {
  const [devices, setDevices] = useState<TopologyDevice[]>(() => loadSavedTopology()?.devices ?? initialDevices())
  const [connections, setConnections] = useState<TopologyConnection[]>(() => {
    const saved = loadSavedTopology()
    if (saved?.connections) return saved.connections
    return initialConnections(devices)
  })
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null)
  const [selectedConnectionId, setSelectedConnectionId] = useState<string | null>(null)
  const [mode, setMode] = useState<CanvasMode>('select')
  const [connectSourceId, setConnectSourceId] = useState<string | null>(null)
  const [zoom, setZoom] = useState(1)
  const [analysisOpen, setAnalysisOpen] = useState(true)
  const canvasRef = useRef<HTMLDivElement>(null)

  const selectedDevice = useMemo(
    () => devices.find((d) => d.id === selectedDeviceId) || null,
    [devices, selectedDeviceId],
  )

  const selectedConnection = useMemo(
    () => connections.find((c) => c.id === selectedConnectionId) || null,
    [connections, selectedConnectionId],
  )

  const updateConnection = useCallback((id: string, updates: Partial<TopologyConnection>) => {
    setConnections((prev) => prev.map((c) => (c.id === id ? { ...c, ...updates } : c)))
  }, [])

  // ── Auto-save to localStorage (debounced) ──
  useEffect(() => {
    const timer = setTimeout(() => saveTopologyState(devices, connections), 500)
    return () => clearTimeout(timer)
  }, [devices, connections])

  // ── Device actions ──
  const addDevice = useCallback((type: DeviceType) => {
    const newDevice: TopologyDevice = {
      id: genId(),
      type,
      label: type.charAt(0).toUpperCase() + type.slice(1),
      x: 100 + (devices.length % 4) * 160,
      y: 400 + Math.floor(devices.length / 4) * 120,
      ip: '',
      subnet: '',
      mac: '',
      status: 'healthy',
    }
    setDevices((prev) => [...prev, newDevice])
    setSelectedDeviceId(newDevice.id)
    setSelectedConnectionId(null)
  }, [devices.length])

  const moveDevice = useCallback((id: string, x: number, y: number) => {
    setDevices((prev) =>
      prev.map((d) => (d.id === id ? { ...d, x: Math.max(0, x), y: Math.max(0, y) } : d)),
    )
  }, [])

  const updateDevice = useCallback((id: string, updates: Partial<TopologyDevice>) => {
    setDevices((prev) => prev.map((d) => (d.id === id ? { ...d, ...updates } : d)))
  }, [])

  const clearCanvas = useCallback(() => {
    setDevices([])
    setConnections([])
    setSelectedDeviceId(null)
    setSelectedConnectionId(null)
  }, [])

  // ── Connection actions ──

  const handleCanvasClick = useCallback((e: React.MouseEvent) => {
    // Only deselect when clicking the canvas background itself
    if (e.target === e.currentTarget || (e.target as HTMLElement).tagName === 'svg') {
      setSelectedDeviceId(null)
      setSelectedConnectionId(null)
      if (mode === 'connect') {
        setConnectSourceId(null)
      }
    }
  }, [mode])

  const handleDeviceSelect = useCallback((id: string) => {
    if (mode === 'delete') {
      setDevices((prev) => prev.filter((d) => d.id !== id))
      setConnections((prev) => prev.filter((c) => c.sourceId !== id && c.targetId !== id))
      if (selectedDeviceId === id) setSelectedDeviceId(null)
      return
    }

    if (mode === 'connect') {
      if (connectSourceId === null) {
        setConnectSourceId(id)
      } else if (connectSourceId !== id) {
        // Check if connection already exists
        const exists = connections.some(
          (c) =>
            (c.sourceId === connectSourceId && c.targetId === id) ||
            (c.sourceId === id && c.targetId === connectSourceId),
        )
        if (!exists) {
          const newConn: TopologyConnection = {
            id: genConnId(),
            sourceId: connectSourceId,
            targetId: id,
            label: '1 Gbps',
            type: 'ethernet',
          }
          setConnections((prev) => [...prev, newConn])
        }
        setConnectSourceId(null)
      }
      return
    }

    setSelectedDeviceId(id)
    setSelectedConnectionId(null)
  }, [mode, connectSourceId, connections, selectedDeviceId])

  const handleConnectionSelect = useCallback((id: string) => {
    if (mode === 'delete') {
      setConnections((prev) => prev.filter((c) => c.id !== id))
      setSelectedConnectionId(null)
      return
    }
    setSelectedConnectionId(id)
    setSelectedDeviceId(null)
  }, [mode])

  const handleLabelChange = useCallback((id: string, label: string) => {
    updateDevice(id, { label })
  }, [updateDevice])

  // ── Persistence ──
  const saveTopology = useCallback(async () => {
    try {
      const path = await SaveFileDialog(
        'Save Topology',
        'topology.json',
        ['JSON Files (*.json)|*.json']
      )
      if (path) {
        const data = JSON.stringify({ devices, connections }, null, 2)
        const success = await WriteFile(path, data)
        if (success) {
          console.log('Topology saved to', path)
        }
      }
    } catch (err: unknown) {
      console.error('Failed to save topology:', err)
    }
  }, [devices, connections])

  const loadTopology = useCallback(async () => {
    try {
      const path = await OpenFileDialog(
        'Open Topology',
        ['JSON Files (*.json)|*.json']
      )
      if (path) {
        const content = await ReadFile(path)
        if (content) {
          const data = JSON.parse(content)
          if (data.devices && data.connections) {
            setDevices(data.devices)
            setConnections(data.connections)
            setSelectedDeviceId(null)
            setSelectedConnectionId(null)
          }
        }
      }
    } catch (err: unknown) {
      console.error('Failed to load topology:', err)
    }
  }, [])

  // ── Export ──
  const exportJSON = useCallback(() => {
    const data = { devices, connections }
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'topology.json'
    a.click()
    URL.revokeObjectURL(url)
  }, [devices, connections])

  // ── Zoom ──
  const zoomIn = useCallback(() => setZoom((z) => Math.min(z + 0.1, 2)), [])
  const zoomOut = useCallback(() => setZoom((z) => Math.max(z - 0.1, 0.3)), [])
  const zoomReset = useCallback(() => setZoom(1), [])

  const deviceMap = useMemo(() => {
    const map = new Map<string, TopologyDevice>()
    devices.forEach((d) => map.set(d.id, d))
    return map
  }, [devices])

  return (
    <div className="p-6 space-y-4 overflow-y-auto h-full">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-accent flex items-center gap-2">
          <GitMerge size={24} /> Network Topology Designer
        </h1>
        <p className="text-text-faint text-sm mt-1">Design, visualize, and export network topologies</p>
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-2 flex-wrap bg-panel border border-border rounded-lg p-2">
        {/* Add devices */}
        <div className="flex items-center gap-1 border-r border-border pr-2">
          {deviceTypes.map((dt) => (
            <button
              key={dt.type}
              onClick={() => addDevice(dt.type)}
              className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium text-text-faint hover:text-text hover:bg-sidebar-hover transition-colors"
              title={`Add ${dt.label}`}
            >
              {dt.icon}
              <span className="hidden sm:inline">{dt.label}</span>
            </button>
          ))}
        </div>

        {/* Modes */}
        <div className="flex items-center gap-1 border-r border-border pr-2">
          <button
            onClick={() => { setMode('select'); setConnectSourceId(null) }}
            className={cn(
              'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium transition-colors',
              mode === 'select' ? 'bg-accent/10 text-accent' : 'text-text-faint hover:text-text hover:bg-sidebar-hover',
            )}
            title="Select / Move"
          >
            <MousePointer2 size={14} /> Select
          </button>
          <button
            onClick={() => { setMode('connect'); setSelectedDeviceId(null); setSelectedConnectionId(null) }}
            className={cn(
              'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium transition-colors',
              mode === 'connect' ? 'bg-success/10 text-success' : 'text-text-faint hover:text-text hover:bg-sidebar-hover',
            )}
            title="Connect devices"
          >
            <Link2 size={14} /> Connect
          </button>
          <button
            onClick={() => { setMode('delete'); setConnectSourceId(null) }}
            className={cn(
              'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium transition-colors',
              mode === 'delete' ? 'bg-danger/10 text-danger' : 'text-text-faint hover:text-text hover:bg-sidebar-hover',
            )}
            title="Delete mode"
          >
            <Trash2 size={14} /> Delete
          </button>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1 border-r border-border pr-2">
          <button onClick={saveTopology} className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium text-text-faint hover:text-text hover:bg-sidebar-hover transition-colors" title="Save Topology">
            <Save size={14} /> Save
          </button>
          <button onClick={loadTopology} className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium text-text-faint hover:text-text hover:bg-sidebar-hover transition-colors" title="Load Topology">
            <FolderOpen size={14} /> Open
          </button>
          <button onClick={exportJSON} className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium text-text-faint hover:text-text hover:bg-sidebar-hover transition-colors" title="Export JSON (Web)">
            <FileJson size={14} /> Web Export
          </button>
        </div>

        {/* Zoom */}
        <div className="flex items-center gap-1 border-r border-border pr-2">
          <button onClick={zoomOut} className="px-2 py-1.5 rounded-md text-xs text-text-faint hover:text-text hover:bg-sidebar-hover transition-colors" title="Zoom Out">
            <Minus size={14} />
          </button>
          <span className="text-xs text-text-faint font-mono w-10 text-center">{Math.round(zoom * 100)}%</span>
          <button onClick={zoomIn} className="px-2 py-1.5 rounded-md text-xs text-text-faint hover:text-text hover:bg-sidebar-hover transition-colors" title="Zoom In">
            <Plus size={14} />
          </button>
          <button onClick={zoomReset} className="px-2 py-1.5 rounded-md text-xs text-text-faint hover:text-text hover:bg-sidebar-hover transition-colors" title="Reset Zoom">
            <Maximize2 size={14} />
          </button>
        </div>

        {/* Analysis toggle */}
        <div className="flex items-center gap-1 border-r border-border pr-2">
          <button
            onClick={() => setAnalysisOpen(!analysisOpen)}
            className={cn(
              'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium transition-colors',
              analysisOpen ? 'bg-accent/10 text-accent' : 'text-text-faint hover:text-text hover:bg-sidebar-hover',
            )}
            title="Toggle Analysis Panel"
          >
            <Radar size={14} /> Analyze
          </button>
        </div>

        <button
          onClick={clearCanvas}
          className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium text-text-faint hover:text-danger hover:bg-danger/10 transition-colors"
        >
          <Trash2 size={14} /> Clear
        </button>

        {/* Connect mode hint */}
        {mode === 'connect' && connectSourceId && (
          <span className="text-xs text-success font-medium ml-2">
            Click a target device to connect
          </span>
        )}
      </div>

      {/* Stats bar */}
      <div className="flex items-center gap-3 text-xs text-text-faint">
        <span>{devices.length} devices</span>
        <span className="text-border">|</span>
        <span>{connections.length} connections</span>
        {selectedDeviceId && (
          <>
            <span className="text-border">|</span>
            <span className="text-accent">Selected: {devices.find((d) => d.id === selectedDeviceId)?.label}</span>
          </>
        )}
      </div>

      {/* Main canvas area */}
      <div className="flex gap-4 h-[calc(100vh-280px)] min-h-[500px]">
        <div
          ref={canvasRef}
          className="flex-1 bg-panel border border-border rounded-lg overflow-auto relative min-h-[600px]"
          onClick={handleCanvasClick}
          style={{
            backgroundImage: 'radial-gradient(circle, rgba(255,255,255,0.06) 1px, transparent 1px)',
            backgroundSize: '20px 20px',
          }}
        >
          <div
            style={{
              width: 1200,
              height: 800,
              transform: `scale(${zoom})`,
              transformOrigin: 'top left',
              position: 'relative',
            }}
          >
            {/* SVG connections layer */}
            <svg
              className="absolute inset-0 pointer-events-none"
              style={{ width: 1200, height: 800, zIndex: 5 }}
            >
              {connections.map((conn) => {
                const source = deviceMap.get(conn.sourceId)
                const target = deviceMap.get(conn.targetId)
                if (!source || !target) return null
                return (
                  <ConnectionLine
                    key={conn.id}
                    connection={conn}
                    source={source}
                    target={target}
                    isSelected={selectedConnectionId === conn.id}
                    onSelect={handleConnectionSelect}
                  />
                )
              })}
            </svg>

            {/* Devices */}
            {devices.map((device) => (
              <DeviceNode
                key={device.id}
                device={device}
                isSelected={selectedDeviceId === device.id}
                isConnectMode={mode === 'connect'}
                onSelect={handleDeviceSelect}
                onMove={moveDevice}
                onLabelChange={handleLabelChange}
                onStartConnect={(id) => {
                  if (mode === 'connect') {
                    if (connectSourceId === null) {
                      setConnectSourceId(id)
                    }
                  }
                }}
              />
            ))}

            {/* Empty state */}
            {devices.length === 0 && (
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="text-center">
                  <GitMerge size={48} className="text-text-faint/30 mx-auto mb-3" />
                  <p className="text-sm text-text-faint">Canvas is empty</p>
                  <p className="text-xs text-text-faint/60 mt-1">Use the toolbar above to add network devices</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Properties panel */}
        {selectedDevice && (
          <PropertiesPanel
            device={selectedDevice}
            onUpdate={updateDevice}
            onClose={() => setSelectedDeviceId(null)}
          />
        )}
        {selectedConnection && (
          <ConnectionPropertiesPanel
            connection={selectedConnection}
            devices={devices}
            onUpdate={updateConnection}
            onClose={() => setSelectedConnectionId(null)}
          />
        )}

        {/* Analysis sidebar */}
        <AnalysisSidebar
          devices={devices}
          connections={connections}
          isOpen={analysisOpen}
          onToggle={() => setAnalysisOpen(!analysisOpen)}
          onDevicesDiscovered={(discovered) => {
            // Merge discovered devices with existing, avoiding duplicates
            setDevices((prev) => {
              const existingIds = new Set(prev.map((d) => d.id))
              const newDevices = discovered.filter((d) => !existingIds.has(d.id))
              return [...prev, ...newDevices]
            })
          }}
        />
      </div>
    </div>
  )
}
