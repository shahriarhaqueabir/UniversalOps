import { cn } from '@/lib/utils'
import type { TopologyConnection, TopologyDevice, ConnectionType } from '@/types'

interface ConnectionLineProps {
  connection: TopologyConnection
  source: TopologyDevice
  target: TopologyDevice
  isSelected: boolean
  onSelect: (id: string) => void
}

const connectionColors: Record<ConnectionType, string> = {
  ethernet: 'var(--color-accent)',
  fiber: 'var(--color-warning)',
  wireless: 'var(--color-success)',
  vpn: 'var(--color-danger)',
  direct: 'var(--color-text-dim)',
}

const DEVICE_WIDTH = 120
const DEVICE_HEIGHT = 90

export function ConnectionLine({
  connection,
  source,
  target,
  isSelected,
  onSelect,
}: ConnectionLineProps) {
  const sx = source.x + DEVICE_WIDTH / 2
  const sy = source.y + DEVICE_HEIGHT / 2
  const tx = target.x + DEVICE_WIDTH / 2
  const ty = target.y + DEVICE_HEIGHT / 2

  const mx = (sx + tx) / 2
  const my = (sy + ty) / 2

  const color = connectionColors[connection.type]
  const strokeWidth = isSelected ? 3 : 2

  return (
    <g
      className="cursor-pointer"
      onClick={(e) => {
        e.stopPropagation()
        onSelect(connection.id)
      }}
    >
      {/* Invisible wider hit area */}
      <line
        x1={sx}
        y1={sy}
        x2={tx}
        y2={ty}
        stroke="transparent"
        strokeWidth={14}
        className="cursor-pointer"
      />
      {/* Visible line */}
      <line
        x1={sx}
        y1={sy}
        x2={tx}
        y2={ty}
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={connection.type === 'wireless' ? '6 4' : undefined}
        className={cn(
          'transition-all duration-150',
          isSelected ? 'opacity-100' : 'opacity-70 hover:opacity-100',
        )}
      />
      {/* Label */}
      {connection.label && (
        <g>
          <rect
            x={mx - 30}
            y={my - 9}
            width={60}
            height={18}
            rx={4}
            fill="var(--color-panel-3)"
            stroke={color}
            strokeWidth={1}
            opacity={0.9}
          />
          <text
            x={mx}
            y={my + 4}
            textAnchor="middle"
            fill="var(--color-text)"
            fontSize={9}
            fontFamily="monospace"
          >
            {connection.label}
          </text>
        </g>
      )}
    </g>
  )
}
