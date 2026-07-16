import { useState, useEffect, useCallback } from 'react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Activity,
  Globe,
  Play,
  Square,
  Timer,
  Signal,
  ShieldCheck,
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
import type { PingEntry, PingStats } from '@/types'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'

export function PingTab() {
  const { call } = useBackend()
  const { pingCount } = useSettingsStore()
  const [pingTarget, setPingTarget] = useState('8.8.8.8')
  const [pingRunning, setPingRunning] = useState(false)
  const [pingEntries, setPingEntries] = useState<PingEntry[]>([])

  const executePing = useCallback(async () => {
    try {
      const res = await call('NetOps.Ping', pingTarget, pingCount) as PingStats & { error?: string; ip: string; ttl: number | null }
      if (res?.error) {
        setPingEntries(prev => [...prev.slice(-49), {
          seq: prev.length + 1,
          ip: pingTarget,
          rtt_ms: null,
          jitter_ms: null,
          ttl: null,
          status: 'timeout'
        } as PingEntry])
      } else if (res) {
        setPingEntries(prev => {
          const lastEntry = prev[prev.length - 1]
          let currentJitter = 0
          if (lastEntry && lastEntry.rtt_ms !== null && (res.avg_ms || 0) !== undefined) {
            currentJitter = Math.abs((res.avg_ms || 0) - lastEntry.rtt_ms)
          }
          return [...prev.slice(-49), {
            seq: prev.length + 1,
            ip: res.ip,
            rtt_ms: (res.avg_ms || 0) || res.min_ms,
            jitter_ms: currentJitter,
            ttl: res.ttl,
            status: res.lost > 0 ? 'timeout' : 'success'
          } as PingEntry]
        })
      }
    } catch (err: unknown) {
      console.error('Ping failed:', err)
    }
  }, [call, pingTarget, pingCount])

  useEffect(() => {
    if (pingRunning) {
      const t = setInterval(executePing, 1000)
      return () => clearInterval(t)
    }
  }, [pingRunning, executePing])

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="ICMP Probe Analysis"
        objective="Measure Round-Trip Time (RTT) to determine routing stability. Latency spikes correlate with packet-shaping or bottlenecked gateways."
        checklist={[
          "Ideal RTT: < 50ms for low-latency nodes.",
          "Jitter: Monitor for inconsistent response times.",
          "TTL: Verify hop-count to identify path changes.",
          "Packet Loss: 0% is the target for stable links."
        ]}
      />
      <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[var(--radius-lg)] shadow-inner">
        <div className="relative group flex-1">
          <Globe size={24} className="absolute left-5 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
          <input
            type="text"
            value={pingTarget}
            onChange={(e) => setPingTarget(e.target.value)}
            className="w-full bg-panel border border-border rounded-2xl pl-16 pr-4 py-3 text-sm font-medium text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-accent shadow-xl"
          />
        </div>
        <button onClick={() => {
          if (!pingRunning) setPingEntries([])
          setPingRunning(!pingRunning)
        }} className="flex items-center gap-3 px-5 py-2.5 text-sm font-semibold rounded-xl transition-all shadow-xl bg-accent text-white hover:bg-accent/90">
          {pingRunning ? <Square size={24} fill="currentColor" /> : <Play size={24} fill="currentColor" />}
          {pingRunning ? 'STOP PROBE' : 'START PROBE'}
        </button>
      </div>
      <div className="grid grid-cols-4 gap-6">
        <MiniStat label="Latency" value={pingEntries.length > 0 ? pingEntries[pingEntries.length - 1].rtt_ms?.toFixed(1) || '—' : '—'} unit="ms" icon={<Timer size={24} />} />
        <MiniStat
          label="Jitter"
          value={pingEntries.length > 1 ? (pingEntries.slice(-10).reduce((acc, curr, i, arr) => {
            if (i === 0 || curr.rtt_ms === null || arr[i - 1].rtt_ms === null) return acc
            return acc + Math.abs(curr.rtt_ms - arr[i - 1].rtt_ms!)
          }, 0) / (pingEntries.slice(-10).filter(e => e.rtt_ms !== null).length - 1 || 1)).toFixed(2) : '0.00'}
          unit="ms"
          icon={<Activity size={24} />}
        />
        <MiniStat label="Reliability" value={pingEntries.length > 0 ? (100 - (pingEntries.filter(e => e.status === 'timeout').length / pingEntries.length * 100)).toFixed(1) : '100'} unit="%" icon={<ShieldCheck size={24} />} />
        <MiniStat label="Signal" value={pingEntries.length > 0 ? pingEntries[pingEntries.length - 1].ttl || '—' : '—'} unit="ttl" icon={<Signal size={24} />} />
      </div>

      {/* Latency Chart */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
        <div className="flex items-center justify-between mb-8">
          <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3">
            <Activity size={20} className="text-accent" /> Latency History
          </h3>
          <div className="flex items-center gap-4 text-xs font-bold text-text-faint">
            <span className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-accent" /> RTT (ms)
            </span>
          </div>
        </div>
        <div className="h-[300px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={pingEntries.map(e => ({ seq: e.seq, rtt: e.rtt_ms || 0 }))}>
              <defs>
                <linearGradient id="colorRtt" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-accent)" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="var(--color-accent)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
              <XAxis dataKey="seq" hide />
              <YAxis
                stroke="rgba(255,255,255,0.3)"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v: number) => `${v}ms`}
              />
              <Tooltip
                contentStyle={{ backgroundColor: 'rgba(0,0,0,0.8)', border: '1px solid rgba(255,255,255,0.1)', borderRadius: '12px' }}
                itemStyle={{ color: 'var(--color-accent)' }}
                labelStyle={{ display: 'none' }}
              />
              <Area
                type="monotone"
                dataKey="rtt"
                stroke="var(--color-accent)"
                strokeWidth={3}
                fillOpacity={1}
                fill="url(#colorRtt)"
                animationDuration={300}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
