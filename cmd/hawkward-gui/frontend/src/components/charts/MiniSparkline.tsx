import { useId } from 'react'
import { AreaChart, Area, ResponsiveContainer } from 'recharts'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'

interface MiniSparklineProps {
  data: { time: string; value: number }[]
  color?: string
  trend?: 'up' | 'down' | 'stable'
  height?: number
}

export function MiniSparkline({
  data,
  color = '#38bdf8',
  trend,
  height = 36,
}: MiniSparklineProps) {
  const gradientId = useId()
  const lastTwo = data.slice(-2)
  const computedTrend = trend ?? getTrend(lastTwo)

  return (
    <div className="flex items-center gap-1.5">
      <div className="flex-1 min-w-0">
        <ResponsiveContainer width="100%" height={height}>
          <AreaChart data={data}>
            <defs>
              <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity={0.4} />
                <stop offset="100%" stopColor={color} stopOpacity={0} />
              </linearGradient>
            </defs>
            <Area
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={1.5}
              fill={`url(#${gradientId})`}
              dot={false}
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
      {computedTrend === 'up' && <TrendingUp size={14} className="text-success shrink-0" />}
      {computedTrend === 'down' && <TrendingDown size={14} className="text-danger shrink-0" />}
      {computedTrend === 'stable' && <Minus size={14} className="text-text-faint shrink-0" />}
    </div>
  )
}

function getTrend(lastTwo: { value: number }[]): 'up' | 'down' | 'stable' {
  if (lastTwo.length < 2) return 'stable'
  const diff = lastTwo[1].value - lastTwo[0].value
  if (diff > 1) return 'up'
  if (diff < -1) return 'down'
  return 'stable'
}
