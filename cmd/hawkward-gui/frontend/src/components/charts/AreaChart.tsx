import { useId } from 'react'
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  ReferenceLine,
} from 'recharts'

interface AreaChartProps {
  data: { time: string; value: number }[]
  title?: string
  color?: string
  threshold?: number
  unit?: string
  height?: number
}

export function AreaChart({
  data,
  title,
  color = '#38bdf8',
  threshold,
  unit = '%',
  height = 200,
}: AreaChartProps) {
  const gradientId = useId()

  return (
    <div className="w-full">
      {title && <h4 className="text-sm font-medium text-text-faint mb-2">{title}</h4>}
      <ResponsiveContainer width="100%" height={height}>
        <RechartsAreaChart data={data} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.3} />
              <stop offset="100%" stopColor={color} stopOpacity={0.05} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" strokeOpacity={0.3} />
          <XAxis
            dataKey="time"
            tick={{ fill: 'var(--color-text-faint)', fontSize: 10 }}
            axisLine={{ stroke: 'var(--color-border)' }}
            tickLine={false}
            interval="preserveStartEnd"
          />
          <YAxis
            tick={{ fill: 'var(--color-text-faint)', fontSize: 10 }}
            axisLine={false}
            tickLine={false}
            domain={[0, 'auto']}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--color-panel-3)',
              border: '1px solid var(--color-border)',
              borderRadius: '8px',
              color: 'var(--color-text)',
              fontSize: '12px',
            }}
            formatter={(value) => [`${Number(value).toFixed(1)}${unit}`, title || 'Value']}
            labelStyle={{ color: 'var(--color-text-faint)' }}
          />
          {threshold !== undefined && (
            <ReferenceLine
              y={threshold}
              stroke="#f87171"
              strokeDasharray="4 4"
              label={{
                value: `Threshold: ${threshold}${unit}`,
                fill: '#f87171',
                fontSize: 10,
                position: 'right',
              }}
            />
          )}
          <Area
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={2}
            fill={`url(#${gradientId})`}
            dot={false}
            activeDot={{ r: 4, fill: color, stroke: 'var(--color-panel-3)', strokeWidth: 2 }}
          />
        </RechartsAreaChart>
      </ResponsiveContainer>
    </div>
  )
}
