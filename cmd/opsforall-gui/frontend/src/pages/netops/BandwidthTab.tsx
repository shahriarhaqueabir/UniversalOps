import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Wifi,
  Signal,
} from 'lucide-react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import type { InterfaceInfo } from '@/types'
import { SectionBriefing, MiniStat } from './components'

export function BandwidthTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: interfaces = [] } = useQuery<InterfaceInfo[]>({
    queryKey: ['netops-interfaces'],
    queryFn: async () => {
      const res = await call('NetOps.GetInterfaces') as InterfaceInfo[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Traffic Analysis"
        objective="Monitor bandwidth utilization across all network interfaces. Track real-time throughput and historical traffic patterns to identify congestion and plan capacity."
        checklist={[
          "Combined throughput shows aggregate bandwidth demand across all links.",
          "RX/TX balance indicates asymmetric link usage or duplex mismatches.",
          "Sustained saturation suggests capacity planning is needed.",
          "Spike patterns may correlate with large transfers or anomalies."
        ]}
      />

      {interfaces.length === 0 ? (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-12 shadow-xl text-center">
          <Signal size={48} className="mx-auto mb-4 text-text-faint" />
          <p className="text-sm font-medium text-[var(--color-text-dim)]">
            No interface data available. Visit the Hardware tab first.
          </p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-3 gap-6">
            <MiniStat
              label="Total Interfaces"
              value={interfaces.length}
              icon={<Wifi size={24} />}
            />
            <MiniStat
              label="Combined RX"
              value={(interfaces.reduce((sum, i) => sum + i.rx_rate_bps, 0) / 1_000_000).toFixed(2)}
              unit="Mbps"
              icon={<Signal size={24} />}
            />
            <MiniStat
              label="Combined TX"
              value={(interfaces.reduce((sum, i) => sum + i.tx_rate_bps, 0) / 1_000_000).toFixed(2)}
              unit="Mbps"
              icon={<Signal size={24} />}
            />
          </div>

          <div className="grid grid-cols-2 gap-8">
            {interfaces.map((iface) => {
              const hasHistory = iface.rx_history?.length > 0
              const chartData = hasHistory
                ? iface.rx_history.map((_, idx) => ({
                  point: idx,
                  rx: +(iface.rx_history[idx] / 1_000_000).toFixed(4),
                  tx: +(iface.tx_history[idx] / 1_000_000).toFixed(4),
                }))
                : []
              const gradId = `bw-${iface.name.replace(/[^a-zA-Z0-9]/g, '')}`

              return (
                <div key={iface.name} className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
                  <div className="flex items-center justify-between mb-6 flex-wrap gap-3">
                    <h3 className="text-lg font-bold text-text uppercase tracking-wider">{iface.name}</h3>
                    <div className="flex items-center gap-4 text-sm font-bold tabular-nums">
                      <span className="flex items-center gap-1.5">
                        <span className="w-2.5 h-2.5 rounded-full bg-[var(--color-success)]" />
                        <span className="text-text-dim">RX <span className="text-text">{(iface.rx_rate_bps / 1_000_000).toFixed(2)}</span> Mbps</span>
                      </span>
                      <span className="flex items-center gap-1.5">
                        <span className="w-2.5 h-2.5 rounded-full bg-[var(--color-accent)]" />
                        <span className="text-text-dim">TX <span className="text-text">{(iface.tx_rate_bps / 1_000_000).toFixed(2)}</span> Mbps</span>
                      </span>
                    </div>
                  </div>
                  {hasHistory ? (
                    <ResponsiveContainer width="100%" height={200}>
                      <AreaChart data={chartData}>
                        <defs>
                          <linearGradient id={`${gradId}-rx`} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="var(--color-success)" stopOpacity={0.35} />
                            <stop offset="95%" stopColor="var(--color-success)" stopOpacity={0} />
                          </linearGradient>
                          <linearGradient id={`${gradId}-tx`} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="var(--color-accent)" stopOpacity={0.35} />
                            <stop offset="95%" stopColor="var(--color-accent)" stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
                        <XAxis dataKey="point" tick={false} axisLine={false} />
                        <YAxis tick={{ fill: 'rgba(255,255,255,0.3)', fontSize: 11 }} tickFormatter={(v: number) => `${v}`} width={40} axisLine={false} />
                        <Tooltip
                          contentStyle={{
                            background: 'var(--color-panel)',
                            border: '1px solid var(--color-border)',
                            borderRadius: '12px',
                            backdropFilter: 'blur(8px)',
                          }}
                          labelStyle={{ display: 'none' }}
                          formatter={(value: any) => [`${Number(value ?? 0).toFixed(2)} Mbps`]}
                        />
                        <Area type="monotone" dataKey="rx" stackId="1" stroke="var(--color-success)" fill={`url(#${gradId}-rx)`} strokeWidth={2} dot={false} />
                        <Area type="monotone" dataKey="tx" stackId="1" stroke="var(--color-accent)" fill={`url(#${gradId}-tx)`} strokeWidth={2} dot={false} />
                      </AreaChart>
                    </ResponsiveContainer>
                  ) : (
                    <div className="h-[200px] flex items-center justify-center">
                      <p className="text-sm font-bold text-text-faint">No traffic history available</p>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </>
      )}
    </div>
  )
}
