import { useState, useCallback, useRef, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useNavigationStore } from '@/stores'
import {
  Map,
  Save,
  Plus,
  Trash2,
  Link2,
  MousePointer2,
  RefreshCw,
  Network,
  Radio,
  Server,
  Monitor,
  Shield,
  Cloud,
  GitMerge,
  Zap,
  Download,
} from 'lucide-react'
import { DeviceNode } from '@/components/network/DeviceNode'
import { ConnectionLine } from '@/components/network/ConnectionLine'
import type {
  NetworkTopologyData,
  TopologyDeviceData,
  TopologyConnectionData,
  DiscoveryTemplateData,
  TopologyDevice,
  TopologyConnection,
  DeviceType,
  ConnectionType,
} from '@/types'

// ── Helpers ──

function dataToDevice(d: TopologyDeviceData): TopologyDevice {
  return {
    id: d.id,
    type: (d.type as DeviceType) || 'unknown',
    label: d.label,
    x: d.x,
    y: d.y,
    ip: d.ip,
    subnet: d.subnet,
    mac: d.mac,
    vendor: d.vendor,
    hostname: d.hostname,
    status: (d.status as TopologyDevice['status']) || 'healthy',
    online: d.online,
    notes: d.notes,
  }
}

function dataToConnection(c: TopologyConnectionData): TopologyConnection {
  return {
    id: c.id,
    sourceId: c.source_id,
    targetId: c.target_id,
    type: (c.type as ConnectionType) || 'ethernet',
    label: c.label,
    metric: c.metric,
  }
}

function deviceToData(d: TopologyDevice): TopologyDeviceData {
  return {
    id: d.id,
    type: d.type,
    label: d.label,
    x: d.x,
    y: d.y,
    ip: d.ip || '',
    subnet: d.subnet || '',
    mac: d.mac || '',
    vendor: d.vendor || '',
    hostname: d.hostname || '',
    status: d.status,
    online: d.online ?? true,
    notes: d.notes || '',
  }
}

function connectionToData(c: TopologyConnection): TopologyConnectionData {
  return {
    id: c.id,
    source_id: c.sourceId,
    target_id: c.targetId,
    type: c.type,
    label: c.label || '',
    metric: c.metric || 0,
  }
}

let idCounter = 1000
function genId(prefix: string) {
  idCounter++
  return `${prefix}-${idCounter}`
}

// ── Device type config for the add-device palette ──

interface DeviceTypeConfig {
  type: DeviceType
  label: string
  icon: React.ReactNode
  color: string
}

const deviceTypes: DeviceTypeConfig[] = [
  { type: 'gateway', label: 'Gateway', icon: <Network size={16} />, color: 'var(--color-accent)' },
  { type: 'router', label: 'Router', icon: <Server size={16} />, color: 'var(--color-accent)' },
  { type: 'switch', label: 'Switch', icon: <GitMerge size={16} />, color: 'var(--color-success)' },
  { type: 'firewall', label: 'Firewall', icon: <Shield size={16} />, color: 'var(--color-danger)' },
  { type: 'server', label: 'Server', icon: <Server size={16} />, color: 'var(--color-warning)' },
  { type: 'workstation', label: 'Workstation', icon: <Monitor size={16} />, color: 'var(--color-accent-2)' },
  { type: 'cloud', label: 'Cloud', icon: <Cloud size={16} />, color: 'var(--color-text-dim)' },
]

const connectionTypeOptions: { type: ConnectionType; label: string }[] = [
  { type: 'ethernet', label: 'Ethernet' },
  { type: 'fiber', label: 'Fiber' },
  { type: 'wireless', label: 'Wireless' },
  { type: 'vpn', label: 'VPN' },
  { type: 'direct', label: 'Direct' },
]

// ── NetworkDesigner Page ──

export function NetworkDesigner() {
  const { call } = useBackend()
  const { navigate } = useNavigationStore()
  const canvasRef = useRef<HTMLDivElement>(null)

  // State
  const [devices, setDevices] = useState<TopologyDevice[]>([])
  const [connections, setConnections] = useState<TopologyConnection[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [connectMode, setConnectMode] = useState(false)
  const [connectSource, setConnectSource] = useState<string | null>(null)
  const [selectedTemplate, setSelectedTemplate] = useState('ping-sweep')
  const [hasRun, setHasRun] = useState(false)
  const [mode, setMode] = useState<'select' | 'connect' | 'add-device'>('select')
  const [addingDeviceType, setAddingDeviceType] = useState<DeviceType>('workstation')
  const [showSaveConfirm, setShowSaveConfirm] = useState(false)

  // ── Queries ──

  const { data: templates } = useQuery<DiscoveryTemplateData[]>({
    queryKey: ['netops-templates'],
    queryFn: async () => (await call('NetOps.GetDiscoveryTemplates')) as DiscoveryTemplateData[],
    staleTime: 300_000,
  })

  const { data: savedTopology, refetch: refetchSaved } = useQuery<NetworkTopologyData>({
    queryKey: ['netops-topology'],
    queryFn: async () => (await call('NetOps.GetTopology')) as NetworkTopologyData,
    enabled: false,
    retry: false,
  })

  const {
    data: discoveryResult,
    isFetching: isDiscovering,
    refetch: runDiscovery,
  } = useQuery<NetworkTopologyData>({
    queryKey: ['netops-auto-discover', selectedTemplate],
    queryFn: async () =>
      (await call('NetOps.RunAutoDiscovery', selectedTemplate)) as NetworkTopologyData,
    enabled: false,
    retry: false,
  })

  // Load discovery results into local state
  useEffect(() => {
    if (discoveryResult) {
      setDevices(discoveryResult.devices.map(dataToDevice))
      setConnections(discoveryResult.connections.map(dataToConnection))
      setHasRun(true)
      setSelectedId(null)
    }
  }, [discoveryResult])

  // Load saved topology
  useEffect(() => {
    if (savedTopology) {
      setDevices(savedTopology.devices.map(dataToDevice))
      setConnections(savedTopology.connections.map(dataToConnection))
      setHasRun(true)
    }
  }, [savedTopology])

  // ── Handlers ──

  const handleRunDiscovery = useCallback(() => {
    runDiscovery()
  }, [runDiscovery])

  const handleLoadSaved = useCallback(() => {
    refetchSaved()
  }, [refetchSaved])

  const handleSave = useCallback(async () => {
    const data: NetworkTopologyData = {
      devices: devices.map(deviceToData),
      connections: connections.map(connectionToData),
      generated_at: new Date().toISOString(),
      subnet: '',
    }
    await call('NetOps.SaveTopology', data)
    setShowSaveConfirm(true)
    setTimeout(() => setShowSaveConfirm(false), 2000)
  }, [devices, connections, call])

  const handleClear = useCallback(() => {
    setDevices([])
    setConnections([])
    setSelectedId(null)
    setHasRun(false)
  }, [])

  // Device CRUD
  const handleSelectDevice = useCallback((id: string) => {
    setSelectedId(id)
  }, [])

  const handleMoveDevice = useCallback((id: string, x: number, y: number) => {
    setDevices(prev => prev.map(d => (d.id === id ? { ...d, x, y } : d)))
  }, [])

  const handleLabelChange = useCallback((id: string, label: string) => {
    setDevices(prev => prev.map(d => (d.id === id ? { ...d, label } : d)))
  }, [])

  const handleStartConnect = useCallback((id: string) => {
    if (connectSource === null) {
      setConnectSource(id)
    } else if (connectSource !== id) {
      // Create connection
      const newConn: TopologyConnection = {
        id: genId('conn'),
        sourceId: connectSource,
        targetId: id,
        type: 'ethernet',
        label: 'LAN',
      }
      setConnections(prev => [...prev, newConn])
      setConnectSource(null)
      setConnectMode(false)
    } else {
      setConnectSource(null)
    }
  }, [connectSource])

  const handleDeleteSelected = useCallback(() => {
    if (!selectedId) return
    setDevices(prev => prev.filter(d => d.id !== selectedId))
    setConnections(prev => prev.filter(c => c.sourceId !== selectedId && c.targetId !== selectedId))
    setSelectedId(null)
  }, [selectedId])

  const handleDeleteConnection = useCallback((id: string) => {
    setConnections(prev => prev.filter(c => c.id !== id))
    setSelectedId(null)
  }, [])

  const handleAddDevice = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    const newDev: TopologyDevice = {
      id: genId('dev'),
      type: addingDeviceType,
      label: `New ${addingDeviceType.charAt(0).toUpperCase() + addingDeviceType.slice(1)}`,
      x: rect.width / 2 - 60 + (devices.length % 5) * 30,
      y: rect.height / 2 - 45 + Math.floor(devices.length / 5) * 30,
      status: 'healthy',
      online: true,
    }
    setDevices(prev => [...prev, newDev])
    setSelectedId(newDev.id)
  }, [addingDeviceType, devices.length])

  const handleCanvasClick = useCallback(() => {
    if (connectSource !== null) {
      setConnectSource(null)
      setConnectMode(false)
    }
    setSelectedId(null)
  }, [connectSource])

  // ── Canvas click handler ──
  const onCanvasMouseDown = useCallback(
    (e: React.MouseEvent) => {
      if ((e.target as HTMLElement).closest('[data-device-node]')) return
      if ((e.target as HTMLElement).closest('[data-connection-line]')) return
      handleCanvasClick()
    },
    [handleCanvasClick],
  )

  // ── Render ──

  const selectedDevice = devices.find(d => d.id === selectedId)
  const selectedConn = connections.find(c => c.id === selectedId)

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500">
      {/* Header */}
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 flex items-center justify-between px-10">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
              <Map size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Topology & Discovery</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">Network Designer</p>
          <p className="text-[var(--color-text-dim)] text-xs font-semibold uppercase tracking-widest mt-2">
            Auto-discover network topology or design manually with drag-and-drop.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate('netops', 'discovery')}
            className="flex items-center gap-2 px-4 py-2 text-xs font-bold rounded-xl bg-panel-3 border border-border text-text-dim hover:text-text hover:border-accent/30 transition-all"
          >
            <Radio size={14} />
            Discovery Tools
          </button>
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-2 px-10 py-3 border-b border-[var(--color-border)] bg-[var(--color-panel)]">
        {/* Mode toggles */}
        <div className="flex items-center gap-1 bg-panel-3 rounded-xl p-1 border border-border">
          <button
            onClick={() => { setMode('select'); setConnectMode(false); setConnectSource(null) }}
            className={cn(
              'flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-lg transition-all',
              mode === 'select' ? 'bg-accent text-white shadow-lg' : 'text-text-dim hover:text-text',
            )}
          >
            <MousePointer2 size={14} />
            Select
          </button>
          <button
            onClick={() => { setMode('connect'); setConnectMode(true); setConnectSource(null) }}
            className={cn(
              'flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-lg transition-all',
              mode === 'connect' ? 'bg-accent text-white shadow-lg' : 'text-text-dim hover:text-text',
            )}
          >
            <Link2 size={14} />
            Connect
          </button>
        </div>

        <div className="w-px h-6 bg-border mx-2" />

        {/* Template selector */}
        <select
          value={selectedTemplate}
          onChange={e => setSelectedTemplate(e.target.value)}
          className="bg-panel-3 border border-border rounded-lg px-3 py-1.5 text-xs font-bold text-text outline-none focus:border-accent/50"
        >
          {templates?.map(t => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>

        <button
          onClick={handleRunDiscovery}
          disabled={isDiscovering}
          className={cn(
            'flex items-center gap-1.5 px-4 py-1.5 text-xs font-bold rounded-xl transition-all',
            isDiscovering
              ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
              : 'bg-accent text-white hover:bg-accent/90 shadow-lg',
          )}
        >
          {isDiscovering ? <RefreshCw size={14} className="animate-spin" /> : <Zap size={14} />}
          {isDiscovering ? 'Discovering...' : 'Auto-Discover'}
        </button>

        <button
          onClick={handleLoadSaved}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-xl bg-panel-3 border border-border text-text-dim hover:text-text transition-all"
        >
          <Download size={14} />
          Load Saved
        </button>

        <div className="w-px h-6 bg-border mx-2" />

        {/* Add device */}
        <div className="flex items-center gap-1 bg-panel-3 rounded-xl p-1 border border-border">
          <select
            value={addingDeviceType}
            onChange={e => setAddingDeviceType(e.target.value as DeviceType)}
            className="bg-transparent border-none px-2 py-1 text-xs font-bold text-text outline-none"
          >
            {deviceTypes.map(dt => (
              <option key={dt.type} value={dt.type}>{dt.label}</option>
            ))}
          </select>
          <button
            onClick={handleAddDevice}
            className="flex items-center gap-1 px-3 py-1 text-xs font-bold rounded-lg bg-accent text-white hover:bg-accent/90 transition-all"
          >
            <Plus size={14} />
            Add
          </button>
        </div>

        <div className="flex-1" />

        {/* Actions */}
        {selectedId && (
          <>
            {selectedConn ? (
              <button
                onClick={() => handleDeleteConnection(selectedId)}
                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-xl bg-danger/20 text-danger border border-danger/30 hover:bg-danger/30 transition-all"
              >
                <Trash2 size={14} />
                Delete Connection
              </button>
            ) : (
              <button
                onClick={handleDeleteSelected}
                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-xl bg-danger/20 text-danger border border-danger/30 hover:bg-danger/30 transition-all"
              >
                <Trash2 size={14} />
                Delete Device
              </button>
            )}
          </>
        )}

        <button
          onClick={handleClear}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-xl bg-panel-3 border border-border text-text-dim hover:text-text transition-all"
        >
          <Trash2 size={14} />
          Clear
        </button>

        <button
          onClick={handleSave}
          className="flex items-center gap-1.5 px-4 py-1.5 text-xs font-bold rounded-xl bg-success/20 text-success border border-success/30 hover:bg-success/30 transition-all"
        >
          <Save size={14} />
          Save
        </button>

        {showSaveConfirm && (
          <span className="text-xs font-bold text-success animate-in fade-in">Saved!</span>
        )}
      </div>

      {/* Main Content: Canvas + Properties Panel */}
      <div className="flex-1 flex overflow-hidden">
        {/* Canvas */}
        <div
          ref={canvasRef}
          className="flex-1 relative overflow-hidden bg-[var(--color-bg)]"
          onMouseDown={onCanvasMouseDown}
        >
          {/* Grid background */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none" style={{ opacity: 0.06 }}>
            <defs>
              <pattern id="grid" width={40} height={40} patternUnits="userSpaceOnUse">
                <path d="M 40 0 L 0 0 0 40" fill="none" stroke="currentColor" strokeWidth={0.5} />
              </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#grid)" />
          </svg>

          {/* Empty state */}
          {!hasRun && devices.length === 0 && (
            <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
              <div className="w-16 h-16 rounded-2xl bg-panel-3 border border-border flex items-center justify-center mb-4">
                <Map size={32} className="text-text-faint" />
              </div>
              <p className="text-sm font-bold text-text-faint uppercase tracking-widest mb-1">Empty Canvas</p>
              <p className="text-xs font-medium text-text-faint max-w-md text-center">
                Run an auto-discovery scan or manually add devices to build your network topology.
              </p>
            </div>
          )}

          {/* Connections layer */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none">
            {connections.map(conn => {
              const source = devices.find(d => d.id === conn.sourceId)
              const target = devices.find(d => d.id === conn.targetId)
              if (!source || !target) return null
              return (
                <g key={conn.id} data-connection-line className="pointer-events-auto">
                  <ConnectionLine
                    connection={conn}
                    source={source}
                    target={target}
                    isSelected={selectedId === conn.id}
                    onSelect={handleDeleteConnection}
                  />
                </g>
              )
            })}
          </svg>

          {/* Devices layer */}
          {devices.map(device => (
            <div
              key={device.id}
              data-device-node
              className="absolute pointer-events-auto"
              style={{ left: device.x, top: device.y }}
            >
              <DeviceNode
                device={device}
                isSelected={selectedId === device.id}
                isConnectMode={connectMode}
                onSelect={handleSelectDevice}
                onMove={handleMoveDevice}
                onLabelChange={handleLabelChange}
                onStartConnect={handleStartConnect}
              />
            </div>
          ))}

          {/* Connection mode indicator */}
          {connectSource && (
            <div className="absolute top-4 left-1/2 -translate-x-1/2 px-4 py-2 bg-accent/20 border border-accent/30 rounded-xl shadow-xl">
              <p className="text-xs font-bold text-accent flex items-center gap-2">
                <Link2 size={14} />
                Click a target device to connect, or click empty space to cancel
              </p>
            </div>
          )}
        </div>

        {/* Properties Panel */}
        <div className="w-72 border-l border-[var(--color-border)] bg-[var(--color-panel)] overflow-y-auto p-4 space-y-4">
          <h3 className="text-[10px] font-black text-text-faint uppercase tracking-widest">Properties</h3>

          {selectedDevice && (
            <div className="space-y-3">
              <div className="flex items-center gap-2 mb-3">
                <div className="w-8 h-8 rounded-lg bg-panel-3 border border-border flex items-center justify-center" style={{ color: deviceTypes.find(dt => dt.type === selectedDevice.type)?.color }}>
                  {deviceTypes.find(dt => dt.type === selectedDevice.type)?.icon || <Monitor size={16} />}
                </div>
                <div>
                  <p className="text-xs font-bold text-text">{selectedDevice.label}</p>
                  <p className="text-[10px] font-medium text-text-faint uppercase">{selectedDevice.type}</p>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Label</label>
                <input
                  value={selectedDevice.label}
                  onChange={e => handleLabelChange(selectedDevice.id, e.target.value)}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50"
                />
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">IP Address</label>
                <input
                  value={selectedDevice.ip || ''}
                  onChange={e => setDevices(prev => prev.map(d => d.id === selectedDevice.id ? { ...d, ip: e.target.value } : d))}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50"
                  placeholder="192.168.1.x"
                />
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">MAC Address</label>
                <input
                  value={selectedDevice.mac || ''}
                  onChange={e => setDevices(prev => prev.map(d => d.id === selectedDevice.id ? { ...d, mac: e.target.value } : d))}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50"
                  placeholder="xx:xx:xx:xx:xx:xx"
                />
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Type</label>
                <select
                  value={selectedDevice.type}
                  onChange={e => setDevices(prev => prev.map(d => d.id === selectedDevice.id ? { ...d, type: e.target.value as DeviceType } : d))}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50"
                >
                  {deviceTypes.map(dt => (
                    <option key={dt.type} value={dt.type}>{dt.label}</option>
                  ))}
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Status</label>
                <select
                  value={selectedDevice.status}
                  onChange={e => setDevices(prev => prev.map(d => d.id === selectedDevice.id ? { ...d, status: e.target.value as TopologyDevice['status'] } : d))}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50"
                >
                  <option value="healthy">Healthy</option>
                  <option value="warning">Warning</option>
                  <option value="critical">Critical</option>
                </select>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={selectedDevice.online ?? true}
                  onChange={e => setDevices(prev => prev.map(d => d.id === selectedDevice.id ? { ...d, online: e.target.checked } : d))}
                  className="rounded border-border"
                />
                <span className="text-xs font-bold text-text">Online</span>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Notes</label>
                <textarea
                  value={selectedDevice.notes || ''}
                  onChange={e => setDevices(prev => prev.map(d => d.id === selectedDevice.id ? { ...d, notes: e.target.value } : d))}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50 resize-none h-20"
                  placeholder="Optional notes..."
                />
              </div>
            </div>
          )}

          {selectedConn && (
            <div className="space-y-3">
              <p className="text-xs font-bold text-text">Connection</p>
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Type</label>
                <select
                  value={selectedConn.type}
                  onChange={e => setConnections(prev => prev.map(c => c.id === selectedConn.id ? { ...c, type: e.target.value as ConnectionType } : c))}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50"
                >
                  {connectionTypeOptions.map(ct => (
                    <option key={ct.type} value={ct.type}>{ct.label}</option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Label</label>
                <input
                  value={selectedConn.label || ''}
                  onChange={e => setConnections(prev => prev.map(c => c.id === selectedConn.id ? { ...c, label: e.target.value } : c))}
                  className="w-full bg-panel-3 border border-border rounded-lg px-3 py-2 text-xs font-bold text-text outline-none focus:border-accent/50"
                  placeholder="e.g. LAN, WAN, Backbone"
                />
              </div>
            </div>
          )}

          {!selectedDevice && !selectedConn && (
            <div className="text-center py-8">
              <div className="w-10 h-10 rounded-xl bg-panel-3 border border-border flex items-center justify-center mx-auto mb-3">
                <MousePointer2 size={18} className="text-text-faint" />
              </div>
              <p className="text-xs font-bold text-text-faint">Select a device or connection</p>
              <p className="text-[10px] text-text-faint mt-1">Click on the canvas to inspect properties</p>
            </div>
          )}

          {/* Device count summary */}
          {devices.length > 0 && (
            <div className="pt-4 border-t border-border">
              <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-2">
                Topology Summary
              </p>
              <div className="space-y-1.5">
                <div className="flex justify-between text-xs">
                  <span className="text-text-dim">Devices</span>
                  <span className="font-bold text-text">{devices.length}</span>
                </div>
                <div className="flex justify-between text-xs">
                  <span className="text-text-dim">Connections</span>
                  <span className="font-bold text-text">{connections.length}</span>
                </div>
                <div className="flex justify-between text-xs">
                  <span className="text-text-dim">Device Types</span>
                  <span className="font-bold text-text">
                    {new Set(devices.map(d => d.type)).size}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}