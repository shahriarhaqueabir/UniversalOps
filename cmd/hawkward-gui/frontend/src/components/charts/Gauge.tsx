import { useEffect, useState } from 'react'

interface GaugeProps {
  value: number
  max?: number
  label: string
  unit?: string
  size?: number
  thresholds?: { warn: number; crit: number }
}

function getColor(value: number, warn: number, crit: number): string {
  if (value >= crit) return '#f87171' // danger
  if (value >= warn) return '#fbbf24' // warning
  return '#4ade80' // success
}

export function Gauge({
  value,
  max = 100,
  label,
  unit = '%',
  size = 160,
  thresholds = { warn: 70, crit: 90 },
}: GaugeProps) {
  const [animatedValue, setAnimatedValue] = useState(0)
  const radius = size * 0.35
  const strokeWidth = size * 0.08
  const center = size / 2
  const arcLength = Math.PI * 0.75
  const startAngle = -Math.PI * 0.625

  useEffect(() => {
    const timer = setTimeout(() => setAnimatedValue(value), 50)
    return () => clearTimeout(timer)
  }, [value])

  const clampedValue = Math.min(animatedValue, max)
  const fraction = clampedValue / max
  const endAngle = startAngle + arcLength * fraction

  const x1 = center + radius * Math.cos(startAngle)
  const y1 = center + radius * Math.sin(startAngle)
  const x2 = center + radius * Math.cos(startAngle + arcLength)
  const y2 = center + radius * Math.sin(startAngle + arcLength)
  const xEnd = center + radius * Math.cos(endAngle)
  const yEnd = center + radius * Math.sin(endAngle)

  const largeArc = fraction > 0.5 ? 1 : 0

  const d = `M ${x1} ${y1} A ${radius} ${radius} 0 ${largeArc} 1 ${x2} ${y2}`

  const valueColor = getColor(value, thresholds.warn, thresholds.crit)

  return (
    <div className="flex flex-col items-center justify-center">
      <svg width={size} height={size * 0.7} viewBox={`0 0 ${size} ${size * 0.7}`}>
        {/* Background arc */}
        <path
          d={d}
          fill="none"
          stroke="#334155"
          strokeWidth={strokeWidth}
          strokeLinecap="round"
        />
        {/* Value arc */}
        <path
          d={`M ${x1} ${y1} A ${radius} ${radius} 0 ${largeArc ? 1 : 0} 1 ${xEnd} ${yEnd}`}
          fill="none"
          stroke={valueColor}
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          style={{
            transition: 'stroke-dashoffset 0.8s ease-in-out, stroke 0.3s ease',
          }}
        />
        {/* Center value text */}
        <text
          x={center}
          y={center + size * 0.05}
          textAnchor="middle"
          fill="#f8fafc"
          fontSize={size * 0.12}
          fontWeight="bold"
          fontFamily="Inter, sans-serif"
        >
          {Math.round(clampedValue)}
        </text>
        <text
          x={center}
          y={center + size * 0.15}
          textAnchor="middle"
          fill="#94a3b8"
          fontSize={size * 0.06}
          fontFamily="Inter, sans-serif"
        >
          {unit}
        </text>
      </svg>
      <span className="text-sm text-muted -mt-2">{label}</span>
    </div>
  )
}
