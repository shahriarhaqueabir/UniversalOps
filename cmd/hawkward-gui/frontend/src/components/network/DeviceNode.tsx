import { useState, useRef, useCallback } from 'react'
import { cn } from '@/lib/utils'
import {
  Server,
  Monitor,
  Shield,
  Cloud,
  GitMerge,
} from 'lucide-react'
import type { TopologyDevice, DeviceType } from '@/types'

interface DeviceNodeProps {
  device: TopologyDevice
  isSelected: boolean
  isConnectMode: boolean
  onSelect: (id: string) => void
  onMove: (id: string, x: number, y: number) => void
  onLabelChange: (id: string, label: string) => void
  onStartConnect: (id: string) => void
}

const deviceIcons: Record<DeviceType, React.ReactNode> = {
  router: <Server size={24} />,
  switch: <GitMerge size={24} />,
  server: <Server size={24} />,
  workstation: <Monitor size={24} />,
  firewall: <Shield size={24} />,
  cloud: <Cloud size={24} />,
}

const deviceColors: Record<DeviceType, string> = {
  router: '#38bdf8',
  switch: '#4ade80',
  server: '#fbbf24',
  workstation: '#a78bfa',
  firewall: '#f87171',
  cloud: '#94a3b8',
}

const statusDotColors: Record<string, string> = {
  healthy: 'bg-success shadow-[0_0_6px_rgba(74,222,128,0.6)]',
  warning: 'bg-warning shadow-[0_0_6px_rgba(251,191,36,0.6)]',
  critical: 'bg-danger shadow-[0_0_6px_rgba(248,113,113,0.6)]',
}

export function DeviceNode({
  device,
  isSelected,
  isConnectMode,
  onSelect,
  onMove,
  onLabelChange,
  onStartConnect,
}: DeviceNodeProps) {
  const [editing, setEditing] = useState(false)
  const [editLabel, setEditLabel] = useState(device.label)
  const [dragging, setDragging] = useState(false)
  const dragRef = useRef({ startX: 0, startY: 0, devX: 0, devY: 0 })

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      e.stopPropagation()
      if (isConnectMode) {
        onStartConnect(device.id)
        return
      }
      onSelect(device.id)
      setDragging(true)
      dragRef.current = {
        startX: e.clientX,
        startY: e.clientY,
        devX: device.x,
        devY: device.y,
      }
    },
    [device.id, device.x, device.y, isConnectMode, onSelect, onStartConnect],
  )

  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (!dragging) return
      const dx = e.clientX - dragRef.current.startX
      const dy = e.clientY - dragRef.current.startY
      onMove(device.id, dragRef.current.devX + dx, dragRef.current.devY + dy)
    },
    [dragging, device.id, onMove],
  )

  const handleMouseUp = useCallback(() => {
    setDragging(false)
  }, [])

  const handleDoubleClick = useCallback(() => {
    setEditLabel(device.label)
    setEditing(true)
  }, [device.label])

  const handleBlur = useCallback(() => {
    setEditing(false)
    if (editLabel.trim() && editLabel !== device.label) {
      onLabelChange(device.id, editLabel.trim())
    }
  }, [editLabel, device.id, device.label, onLabelChange])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        (e.currentTarget as HTMLInputElement).blur()
      }
      if (e.key === 'Escape') {
        setEditLabel(device.label)
        setEditing(false)
      }
    },
    [device.label],
  )

  return (
    <div
      className={cn(
        'absolute flex flex-col items-center cursor-grab select-none transition-shadow',
        dragging && 'cursor-grabbing z-50',
        isSelected && 'z-40',
      )}
      style={{ left: device.x, top: device.y }}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onDoubleClick={handleDoubleClick}
    >
      {/* Device card */}
      <div
        className={cn(
          'flex flex-col items-center justify-center w-[120px] h-[90px] rounded-xl border-2 transition-all duration-150',
          isSelected
            ? 'border-primary bg-primary/10 shadow-[0_0_16px_rgba(56,189,248,0.25)]'
            : 'border-border bg-card hover:border-muted',
          dragging && 'scale-105 shadow-lg',
          isConnectMode && 'hover:border-success cursor-pointer',
        )}
      >
        {/* Icon */}
        <span style={{ color: deviceColors[device.type] }}>
          {deviceIcons[device.type]}
        </span>

        {/* Status dot */}
        <span
          className={cn(
            'absolute top-2 right-2 w-2.5 h-2.5 rounded-full',
            statusDotColors[device.status],
          )}
        />
      </div>

      {/* Label */}
      {editing ? (
        <input
          autoFocus
          className="mt-1.5 bg-card border border-border rounded px-2 py-0.5 text-xs text-text text-center w-[120px] focus:outline-none focus:ring-1 focus:ring-primary"
          value={editLabel}
          onChange={(e) => setEditLabel(e.target.value)}
          onBlur={handleBlur}
          onKeyDown={handleKeyDown}
        />
      ) : (
        <span className="mt-1.5 text-xs text-text text-center truncate w-[120px] block leading-tight">
          {device.label}
        </span>
      )}
    </div>
  )
}
