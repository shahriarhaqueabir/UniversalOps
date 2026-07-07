import { useState, useEffect, useCallback, useMemo } from 'react'
import {
  Cpu,
  MemoryStick,
  HardDrive,
  Activity,
  LayoutDashboard,
  Network,
  Shield,
  Terminal,
  Brain,
  Wifi,
  CheckCircle,
  XCircle,
  AlertTriangle,
} from 'lucide-react'
import {
  AreaChart as RechartsAreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import { opsLayers } from '@/lib/constants'
import type { DashboardData, AlertInfo, TimeSeriesPoint } from '@/types'

/* ───────────────────────────────────────────
   Helpers
   ─────────────────────────────────────────── */

function clamp(v: number, min = 0, max = 100) {
  return Math.min(max, Math.max(min, v))
}

function healthColor(pct: number) {
  if (pct >= 90) return 'var(--color-success)'
  if (pct >= 80) return 'var(--color-warning)'
  return 'var(--color-danger)'
}

function healthLabel(pct: number) {
  if (pct >= 90) return 'Operational'
  if (pct >= 80) return 'Degraded'
  return 'Critical'
}

/* ───────────────────────────────────────────
   Inline: HeroSection
   ─────────────────────────────────────────── */

function HeroSection({ stats }: { stats: DashboardData }) {
  const avgHealth = clamp((stats.cpu.value + stats.memory.value + stats.disk.value) / 3)

  // Derive up / degraded / down counts (out of 14 logical services)
  const totalChecks = 14
  const degradedCount = Math.round(((100 - avgHealth) / 100) * totalChecks * 0.6)
  const downCount = Math.round(((100 - avgHealth) / 100) * totalChecks * 0.3)
  const upCount = totalChecks - degradedCount - downCount

  // SVG donut geometry (r = 38, circumference ≈ 238.76)
  const r = 38
  const circumference = 2 * Math.PI * r
  const dash = (avgHealth / 100) * circumference
  const gap = circumference - dash

  // Simulated 90-bar uptime strip
  const uptimeBars = useMemo(() => {
    return Array.from({ length: 90 }, (_, i) => {
      // Use a pseudo-random seed based on index for stability
      const val = (Math.sin(i * 12.3) + 1) / 2
      if (val > 0.15) return 'up' as const
      if (val > 0.05) return 'degraded' as const
      return 'down' as const
    })
  }, [])

  const barColor: Record<string, string> = {
    up: 'var(--color-success)',
    degraded: 'var(--color-warning)',
    down: 'var(--color-danger)',
  }

  return (
    <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-5 flex items-center gap-6">
      {/* ── SVG Health Donut ── */}
      <div className="relative shrink-0">
        <svg width={96} height={96} viewBox="0 0 96 96" className="tabular-nums">
          <circle cx="48" cy="48" r={r} fill="none" stroke="var(--color-border)" strokeWidth="6" />
          <circle
            cx="48" cy="48" r={r}
            fill="none"
            stroke={healthColor(avgHealth)}
            strokeWidth="6"
            strokeLinecap="round"
            strokeDasharray={`${dash} ${gap}`}
            transform="rotate(-90 48 48)"
            style={{ transition: 'stroke-dasharray 0.6s ease, stroke 0.4s ease' }}
          />
          <text x="48" y="42" textAnchor="middle" fill="var(--color-text)" fontSize="20" fontWeight="700" dominantBaseline="middle">
            {Math.round(avgHealth)}%
          </text>
          <text x="48" y="62" textAnchor="middle" fill="var(--color-text-faint)" fontSize="10" dominantBaseline="middle">
            Health
          </text>
        </svg>
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span
            className="w-2.5 h-2.5 rounded-full inline-block"
            style={{
              backgroundColor: healthColor(avgHealth),
              boxShadow: `0 0 8px ${healthColor(avgHealth)}`,
            }}
          />
          <span className="text-text font-semibold text-lg">
            Overall Status: {healthLabel(avgHealth)}
          </span>
        </div>

        <p className="text-text-dim text-xs mb-3">
          {stats.alerts} active alarm{stats.alerts !== 1 ? 's' : ''}
        </p>

        <div className="flex items-center gap-4 text-xs mb-3">
          <span className="flex items-center gap-1">
            <CheckCircle size={12} className="text-[var(--color-success)]" />
            <span className="text-text-dim">Up</span>
            <span className="text-text tabular-nums">{upCount}</span>
          </span>
          <span className="flex items-center gap-1">
            <AlertTriangle size={12} className="text-[var(--color-warning)]" />
            <span className="text-text-dim">Degraded</span>
            <span className="text-text tabular-nums">{degradedCount}</span>
          </span>
          <span className="flex items-center gap-1">
            <XCircle size={12} className="text-[var(--color-danger)]" />
            <span className="text-text-dim">Down</span>
            <span className="text-text tabular-nums">{downCount}</span>
          </span>
        </div>

        <div className="flex gap-[2px] items-end h-[22px]">
          {uptimeBars.map((state, i) => (
            <div
              key={i}
              className="w-[6px] rounded-[1px] transition-colors"
              style={{
                height: state === 'up' ? '100%' : state === 'degraded' ? '60%' : '30%',
                backgroundColor: barColor[state],
                opacity: state === 'up' ? 0.7 : 1,
              }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}

/* ───────────────────────────────────────────
   Inline: KpiCard
   ─────────────────────────────────────────── */

function KpiCard({
  icon,
  label,
  value,
  unit,
  status,
  sparkline,
}: {
  icon: React.ReactNode
  label: string
  value: string
  unit?: string
  status: 'healthy' | 'warning' | 'critical'
  sparkline?: number[]
}) {
  const statusColor =
    status === 'healthy'
      ? 'var(--color-success)'
      : status === 'warning'
        ? 'var(--color-warning)'
        : 'var(--color-danger)'

  const polylinePoints = useMemo(() => {
    if (!sparkline || sparkline.length < 2) return ''
    const w = 80
    const h = 24
    const min = Math.min(...sparkline)
    const max = Math.max(...sparkline)
    const range = max - min || 1
    return sparkline
      .map((v, i) => `${(i / (sparkline.length - 1)) * w},${h - ((v - min) / range) * (h - 2)}`)
      .join(' ')
  }, [sparkline])

  return (
    <div
      className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4 transition-all duration-200 group hover:border-[var(--color-accent)]/30 hover:shadow-[0_0_16px_rgba(124,108,255,0.08)] hover:-translate-y-0.5 cursor-default"
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="text-text-dim group-hover:text-[var(--color-accent)] transition-colors">
            {icon}
          </span>
          <span className="text-text-dim text-xs font-medium uppercase tracking-wider">
            {label}
          </span>
        </div>
        <span
          className="w-2 h-2 rounded-full inline-block"
          style={{ backgroundColor: statusColor, boxShadow: `0 0 6px ${statusColor}` }}
        />
      </div>

      <div className="flex items-baseline gap-1 mb-2">
        <span className="text-text font-bold text-2xl tabular-nums">{value}</span>
        {unit && <span className="text-text-faint text-xs">{unit}</span>}
      </div>

      {sparkline && sparkline.length > 1 && (
        <svg
          width="100%"
          height={26}
          viewBox="0 0 80 26"
          className="opacity-60 group-hover:opacity-100 transition-opacity"
        >
          <polyline
            points={polylinePoints}
            fill="none"
            stroke={statusColor}
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      )}
    </div>
  )
}

/* ───────────────────────────────────────────
   Main Dashboard
   ─────────────────────────────────────────── */

const EMPTY_DATA: DashboardData = {
  cpu: { value: 0, unit: '%', history: [], forecast: [], trend: 'stable' },
  memory: { value: 0, unit: '%', history: [], forecast: [], trend: 'stable' },
  disk: { value: 0, unit: '%', history: [], forecast: [], trend: 'stable' },
  network: { rx_rate: 0, tx_rate: 0, unit: 'bps' },
  processes: 0,
  connections: 0,
  alerts: 0,
  uptime: '--',
}

export function Dashboard() {
  const { call } = useBackend()

  const [data, setData] = useState<DashboardData>(EMPTY_DATA)
  const [cpuHistory, setCpuHistory] = useState<TimeSeriesPoint[]>([])
  const [memHistory, setMemHistory] = useState<TimeSeriesPoint[]>([])
  const [alerts, setAlerts] = useState<AlertInfo[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    async function init() {
      try {
        const result = await call('Dash.GetDashboardData')
        if (!cancelled && result) {
          const d = result as DashboardData
          setData(d)
          const t = new Date().toLocaleTimeString()
          setCpuHistory([{ time: t, value: d.cpu.value }])
          setMemHistory([{ time: t, value: d.memory.value }])
        }
      } catch (err) {
        console.warn('Failed to fetch dashboard data:', err)
      }
      if (!cancelled) setLoading(false)
    }
    init()

    return () => { cancelled = true }
  }, [call])

  const handleMetrics = useCallback((payload: any) => {
    if (!payload) return
    const d: DashboardData = payload.data ?? payload
    setData(d)
    const t = new Date().toLocaleTimeString()
    setCpuHistory((prev) => [...prev.slice(-59), { time: t, value: d.cpu.value }])
    setMemHistory((prev) => [...prev.slice(-59), { time: t, value: d.memory.value }])
  }, [])

  const handleAlert = useCallback((payload: any) => {
    if (!payload) return
    const alert: AlertInfo = payload.data ?? payload
    setAlerts((prev) => [alert, ...prev].slice(0, 50))
  }, [])

  useEvents('metrics', handleMetrics)
  useEvents('alert', handleAlert)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="flex items-center gap-3">
          <div className="w-5 h-5 border-2 border-[var(--color-accent)] border-t-transparent rounded-full animate-spin" />
          <span className="text-text-dim text-sm">Loading dashboard...</span>
        </div>
      </div>
    )
  }

  const cpuStatus = data.cpu.value >= 90 ? 'critical' : data.cpu.value >= 70 ? 'warning' : 'healthy'
  const memStatus = data.memory.value >= 90 ? 'critical' : data.memory.value >= 70 ? 'warning' : 'healthy'
  const diskStatus = data.disk.value >= 90 ? 'critical' : data.disk.value >= 70 ? 'warning' : 'healthy'
  const netAvg = (data.network.rx_rate + data.network.tx_rate) / 2
  const netStatus = netAvg > 1000000 ? 'critical' : netAvg > 100000 ? 'warning' : 'healthy'

  const iconMap: Record<string, React.ReactNode> = {
    sysops: <Activity size={18} />,
    netops: <Network size={18} />,
    secops: <Shield size={18} />,
    devops: <Terminal size={18} />,
    aiops: <Brain size={18} />,
  }

  return (
    <div className="p-6 space-y-4 overflow-y-auto h-full">
      <div>
        <h1 className="text-text font-bold text-xl flex items-center gap-2">
          <LayoutDashboard size={22} /> Dashboard
        </h1>
        <p className="text-text-dim text-sm mt-0.5">
          System overview and key metrics at a glance
        </p>
      </div>

      <HeroSection stats={data} />

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard
          icon={<Cpu size={18} />}
          label="CPU"
          value={`${Math.round(data.cpu.value)}`}
          unit="%"
          status={cpuStatus}
          sparkline={data.cpu.history}
        />
        <KpiCard
          icon={<MemoryStick size={18} />}
          label="Memory"
          value={`${Math.round(data.memory.value)}`}
          unit="%"
          status={memStatus}
          sparkline={data.memory.history}
        />
        <KpiCard
          icon={<HardDrive size={18} />}
          label="Disk"
          value={`${Math.round(data.disk.value)}`}
          unit="%"
          status={diskStatus}
          sparkline={data.disk.history}
        />
        <KpiCard
          icon={<Wifi size={18} />}
          label="Network"
          value={(data.network.rx_rate / 1024).toFixed(1)}
          unit="KB/s"
          status={netStatus}
          sparkline={[]}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4">
          <h4 className="text-text-dim text-xs font-medium uppercase tracking-wider mb-3">
            CPU Over Time
          </h4>
          <ResponsiveContainer width="100%" height={200}>
            <RechartsAreaChart data={cpuHistory} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
              <defs>
                <linearGradient id="cpuGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--color-accent)" stopOpacity={0.25} />
                  <stop offset="100%" stopColor="var(--color-accent)" stopOpacity={0.04} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" strokeOpacity={0.4} />
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
                domain={[0, 100]}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--color-panel-2)',
                  border: '1px solid var(--color-border)',
                  borderRadius: '8px',
                  color: 'var(--color-text)',
                  fontSize: '12px',
                }}
                formatter={(value: any) => [`${Number(value).toFixed(1)}%`, 'CPU']}
                labelStyle={{ color: 'var(--color-text-dim)' }}
              />
              <Area
                type="monotone"
                dataKey="value"
                stroke="var(--color-accent)"
                strokeWidth={2}
                fill="url(#cpuGrad)"
                dot={false}
                activeDot={{ r: 4, fill: 'var(--color-accent)', stroke: 'var(--color-panel)', strokeWidth: 2 }}
              />
            </RechartsAreaChart>
          </ResponsiveContainer>
        </div>

        <div className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4">
          <h4 className="text-text-dim text-xs font-medium uppercase tracking-wider mb-3">
            Memory Over Time
          </h4>
          <ResponsiveContainer width="100%" height={200}>
            <RechartsAreaChart data={memHistory} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
              <defs>
                <linearGradient id="memGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--color-success)" stopOpacity={0.25} />
                  <stop offset="100%" stopColor="var(--color-success)" stopOpacity={0.04} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" strokeOpacity={0.4} />
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
                domain={[0, 100]}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--color-panel-2)',
                  border: '1px solid var(--color-border)',
                  borderRadius: '8px',
                  color: 'var(--color-text)',
                  fontSize: '12px',
                }}
                formatter={(value: any) => [`${Number(value).toFixed(1)}%`, 'Memory']}
                labelStyle={{ color: 'var(--color-text-dim)' }}
              />
              <Area
                type="monotone"
                dataKey="value"
                stroke="var(--color-success)"
                strokeWidth={2}
                fill="url(#memGrad)"
                dot={false}
                activeDot={{ r: 4, fill: 'var(--color-success)', stroke: 'var(--color-panel)', strokeWidth: 2 }}
              />
            </RechartsAreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
        <div className="lg:col-span-3 grid grid-cols-1 sm:grid-cols-3 gap-4">
          {opsLayers.map((layer) => (
            <a
              key={layer.id}
              href={`#${layer.id}`}
              className="bg-panel border border-[var(--color-border)] rounded-[12px] p-4 transition-all duration-200 group hover:border-[var(--color-accent)]/30 hover:shadow-[0_0_16px_rgba(124,108,255,0.06)]"
            >
              <div className="flex items-center gap-2 mb-2">
                <span className="text-text-dim group-hover:text-[var(--color-accent)] transition-colors">
                  {iconMap[layer.id]}
                </span>
                <span className="text-text-dim text-xs font-medium uppercase tracking-wider">
                  {layer.title}
                </span>
              </div>
              <p className="text-text-faint text-sm">{layer.description}</p>
              <div className="mt-2 flex items-center gap-1.5">
                <span
                  className="w-1.5 h-1.5 rounded-full bg-[var(--color-success)]"
                  style={{ boxShadow: '0 0 6px var(--color-success)' }}
                />
                <span className="text-text-faint text-xs">Online</span>
              </div>
            </a>
          ))}
        </div>

        <div className="lg:col-span-2 bg-panel border border-[var(--color-border)] rounded-[12px] p-4">
          <h4 className="text-text-dim text-xs font-medium uppercase tracking-wider mb-3 flex items-center gap-2">
            <AlertTriangle size={14} />
            Recent Alerts
          </h4>
          <div className="space-y-2 max-h-[240px] overflow-y-auto">
            {alerts.length === 0 && (
              <p className="text-text-faint text-sm">No active alerts</p>
            )}
            {alerts.map((alert) => (
              <div
                key={alert.id}
                className="flex items-start gap-2 py-1.5 px-2 rounded-[8px] hover:bg-[var(--color-panel-2)] transition-colors"
              >
                <span
                  className="w-2 h-2 rounded-full mt-0.5 shrink-0"
                  style={{
                    backgroundColor:
                      alert.level === 'critical'
                        ? 'var(--color-danger)'
                        : alert.level === 'warning'
                          ? 'var(--color-warning)'
                          : 'var(--color-accent)',
                    boxShadow: `0 0 6px ${alert.level === 'critical'
                      ? 'var(--color-danger)'
                      : alert.level === 'warning'
                        ? 'var(--color-warning)'
                        : 'var(--color-accent)'
                      }`,
                  }}
                />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-text text-xs font-medium truncate">
                      {alert.metric}
                    </span>
                    <span
                      className={`text-[10px] font-medium uppercase px-1 py-0.5 rounded ${alert.level === 'critical'
                        ? 'text-[var(--color-danger)] bg-[var(--color-danger)]/10'
                        : alert.level === 'warning'
                          ? 'text-[var(--color-warning)] bg-[var(--color-warning)]/10'
                          : 'text-[var(--color-accent)] bg-[var(--color-accent-soft)]'
                        }`}
                    >
                      {alert.level}
                    </span>
                  </div>
                  <p className="text-text-faint text-xs truncate mt-0.5">
                    {alert.message}
                  </p>
                  <p className="text-text-faint text-[10px] mt-0.5">
                    {new Date(alert.timestamp).toLocaleTimeString()}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
