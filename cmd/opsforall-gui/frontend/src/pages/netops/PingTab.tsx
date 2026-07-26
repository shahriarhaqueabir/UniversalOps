import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { toast } from 'sonner'
import {
  Activity,
  Globe,
  Play,
  Square,
  Timer,
  Signal,
  ShieldCheck,
  Loader2,
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
import type { PingEntry, PingResult } from '@/types'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'
import { SearchInput } from '@/components/ui/SearchInput'

export function PingTab() {
  const { call } = useBackend()
  const [pingTarget, setPingTarget] = useState('8.8.8.8')
  const [pingEntries, setPingEntries] = useState<PingEntry[]>([])

  const pingMutation = useMutation({
    mutationFn: (host: string) => call('NetOps.Ping', host, 10) as Promise<PingResult>,
    onSuccess: (result) => {
      if (result.error) {
        toast.error(`Ping to ${result.target} failed`, { description: result.error })
        return
      }

      const entries: PingEntry[] = []
      const rttGap = result.received > 1 ? (result.max_ms - result.min_ms) / Math.max(result.received - 1, 1) : 0

      for (let i = 0; i < result.received; i++) {
        entries.push({
          seq: i + 1,
          ip: result.ip,
          rtt_ms: Math.round((result.min_ms + rttGap * i) * 10) / 10 || result.avg_ms,
          jitter_ms: result.jitter_ms,
          ttl: result.ttl,
          status: 'success',
        })
      }
      for (let i = 0; i < result.lost; i++) {
        entries.push({
          seq: result.received + i + 1,
          ip: result.ip,
          rtt_ms: null,
          jitter_ms: null,
          ttl: null,
          status: 'timeout',
        })
      }

      entries.sort((a, b) => a.seq - b.seq)
      setPingEntries(entries)
    },
    onError: (err: Error) => {
      toast.error('Ping failed', { description: err.message })
    },
  })

  const handleToggle = () => {
    if (pingMutation.isPending) {
      // Can't abort the one-shot — just let it finish
      return
    }
    if (pingEntries.length > 0) {
      setPingEntries([])
      return
    }
    const target = pingTarget.trim()
    if (!target) {
      toast.error('Enter a target hostname or IP')
      return
    }
    pingMutation.mutate(target)
  }

  const isRunning = pingMutation.isPending
  const lastEntry = pingEntries.length > 0 ? pingEntries[pingEntries.length - 1] : null

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="ICMP Probe Analysis"
        objective="Measure packet loss and round-trip latency to a target host."
        checklist={['Enter target hostname or IP', 'Click start to begin probing', 'Review latency and loss statistics']}
      />
      <div className="flex items-center gap-6 bg-panel-2 border border-border p-6 rounded-[var(--radius-lg)] shadow-inner">
        <div className="flex-1">
          <SearchInput
            icon={<Globe size={18} />}
            value={pingTarget}
            onChange={(e) => setPingTarget(e.target.value)}
            placeholder="Target hostname or IP"
            size="lg"
            className="shadow-xl"
          />
        </div>
        <button
          onClick={handleToggle}
          disabled={isRunning}
          className="flex items-center gap-3 px-8 py-3.5 text-sm font-black uppercase tracking-widest rounded-xl transition-all shadow-xl bg-accent text-white hover:bg-accent/90 active:scale-95 h-12 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isRunning ? (
            <Loader2 size={18} className="animate-spin" />
          ) : pingEntries.length > 0 ? (
            <Square size={18} fill="currentColor" />
          ) : (
            <Play size={18} fill="currentColor" />
          )}
          {isRunning ? 'PROBING\u2026' : pingEntries.length > 0 ? 'CLEAR' : 'START PROBE'}
        </button>
      </div>
      <div className="grid grid-cols-4 gap-6">
        <MiniStat label="Latency" value={lastEntry ? (lastEntry.rtt_ms?.toFixed(1) ?? '\u2014') : '\u2014'} unit="ms" icon={<Timer size={24} />} />
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
        <MiniStat label="Signal" value={lastEntry ? (lastEntry.ttl ?? '\u2014').toString() : '\u2014'} unit="ttl" icon={<Signal size={24} />} />
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
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
