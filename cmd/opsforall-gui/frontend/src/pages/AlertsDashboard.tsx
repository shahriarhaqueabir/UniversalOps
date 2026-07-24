import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Bell, AlertTriangle, AlertOctagon, Info,
  CheckCircle2, RefreshCw, Trash2,
  Search, Activity,
  Shield, ShieldOff,
} from 'lucide-react'
import { motion, AnimatePresence } from 'motion/react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { toast } from 'sonner'
import type { AlertInfo, AlertRuleInfo } from '@/types'

const LEVEL_META: Record<string, { label: string; icon: React.ReactNode; color: string }> = {
  critical: {
    label: 'Critical',
    icon: <AlertOctagon size={14} />,
    color: 'text-red-500 bg-red-500/10 border-red-500/20',
  },
  warning: {
    label: 'Warning',
    icon: <AlertTriangle size={14} />,
    color: 'text-amber-500 bg-amber-500/10 border-amber-500/20',
  },
  info: {
    label: 'Info',
    icon: <Info size={14} />,
    color: 'text-blue-500 bg-blue-500/10 border-blue-500/20',
  },
}

function getLevelMeta(level: string) {
  return LEVEL_META[level.toLowerCase()] ?? {
    label: level,
    icon: <Bell size={14} />,
    color: 'text-neutral-500 bg-neutral-500/10 border-neutral-500/20',
  }
}

function formatTimestamp(ts: string) {
  try {
    const d = new Date(ts)
    return d.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return ts
  }
}

export function AlertsDashboard() {
  const { call } = useBackend()
  const queryClient = useQueryClient()
  const [filterLevel, setFilterLevel] = useState<string>('all')
  const [showResolved, setShowResolved] = useState(false)
  const [search, setSearch] = useState('')
  const [tab, setTab] = useState<'alerts' | 'rules'>('alerts')

  // ── Queries ──

  const { data: allAlerts = [], isLoading: historyLoading } = useQuery<AlertInfo[]>({
    queryKey: ['alertHistory'],
    queryFn: async () => (await call('AlertAPI.GetAlertHistory')) as AlertInfo[] || [],
    refetchInterval: 15_000,
  })

  const { data: rules = [], isLoading: rulesLoading } = useQuery<AlertRuleInfo[]>({
    queryKey: ['alertRules'],
    queryFn: async () => (await call('AlertAPI.GetRules')) as AlertRuleInfo[] || [],
  })

  // ── Mutations ──

  const resolveMutation = useMutation({
    mutationFn: async (id: string) => { await call('AlertAPI.ResolveAlert', id) },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alertHistory'] })
      toast.success('Alert resolved')
    },
    onError: () => toast.error('Failed to resolve alert'),
  })

  const evaluateMutation = useMutation({
    mutationFn: async () => { await call('AlertAPI.EvaluateNow') },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alertHistory'] })
      toast.success('Evaluation complete')
    },
    onError: () => toast.error('Evaluation failed'),
  })

  const removeRuleMutation = useMutation({
    mutationFn: async ({ metric, threshold }: { metric: string; threshold: number }) => {
      await call('AlertAPI.RemoveRule', metric, threshold)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alertRules'] })
      toast.success('Rule removed')
    },
    onError: () => toast.error('Failed to remove rule'),
  })

  // ── Derived data ──

  const displayedAlerts = showResolved ? allAlerts : allAlerts.filter((a) => !a.resolved)

  const filteredAlerts = displayedAlerts.filter((a) => {
    if (filterLevel !== 'all' && a.level?.toLowerCase() !== filterLevel) return false
    if (search) {
      const q = search.toLowerCase()
      if (
        !a.metric?.toLowerCase().includes(q) &&
        !a.message?.toLowerCase().includes(q) &&
        !a.id?.toLowerCase().includes(q)
      ) return false
    }
    return true
  })

  const stats = {
    critical: allAlerts.filter((a) => a.level?.toLowerCase() === 'critical' && !a.resolved).length,
    warning: allAlerts.filter((a) => a.level?.toLowerCase() === 'warning' && !a.resolved).length,
    info: allAlerts.filter((a) => a.level?.toLowerCase() === 'info' && !a.resolved).length,
    total: allAlerts.length,
    active: allAlerts.filter((a) => !a.resolved).length,
    resolved: allAlerts.filter((a) => a.resolved).length,
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500 overflow-hidden">
      {/* Header */}
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 px-10 flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
              <Bell size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Monitoring</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">Alerts Dashboard</p>
          <p className="text-xs text-[var(--color-text-faint)] mt-1 font-medium tracking-wide">
            {stats.active} active &middot; {stats.resolved} resolved &middot; {stats.total} total
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => evaluateMutation.mutate()}
            disabled={evaluateMutation.isPending}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-accent text-white text-xs font-bold hover:brightness-110 transition-all disabled:opacity-50"
          >
            <Activity size={14} />
            Evaluate Now
          </button>
          <button
            onClick={() => {
              queryClient.invalidateQueries({ queryKey: ['alertHistory'] })
              queryClient.invalidateQueries({ queryKey: ['alertRules'] })
            }}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs font-bold text-[var(--color-text)] hover:bg-[var(--color-panel-hover)] transition-all"
          >
            <RefreshCw size={14} />
            Refresh
          </button>
        </div>
      </div>

      {/* Stats bar */}
      <div className="grid grid-cols-5 gap-4 px-10 py-5 border-b border-[var(--color-border)] bg-[var(--color-panel-1)]/30">
        {[
          { key: 'active', label: 'Active', count: stats.active, color: 'text-accent border-accent/20 bg-accent/10' },
          { key: 'critical', label: 'Critical', count: stats.critical, color: 'text-red-500 border-red-500/20 bg-red-500/10' },
          { key: 'warning', label: 'Warning', count: stats.warning, color: 'text-amber-500 border-amber-500/20 bg-amber-500/10' },
          { key: 'info', label: 'Info', count: stats.info, color: 'text-blue-500 border-blue-500/20 bg-blue-500/10' },
          { key: 'total', label: 'Total', count: stats.total, color: 'text-[var(--color-text)] border-[var(--color-border)] bg-[var(--color-panel-2)]' },
        ].map((s) => (
          <div key={s.key} className={cn('flex items-center gap-3 px-4 py-3 rounded-xl border', s.color)}>
            <div>
              <p className="text-xs font-bold uppercase tracking-wide">{s.label}</p>
              <p className={cn('text-lg font-black', s.key !== 'total' && s.key !== 'active' && `text-${s.key === 'critical' ? 'red' : s.key === 'warning' ? 'amber' : 'blue'}-500`)}>
                {s.count}
              </p>
            </div>
          </div>
        ))}
      </div>

      {/* Tabs + Filters */}
      <div className="flex items-center justify-between px-10 py-3 border-b border-[var(--color-border)] bg-[var(--color-panel-1)]/20">
        <div className="flex items-center gap-1">
          <button
            onClick={() => setTab('alerts')}
            className={cn(
              'px-4 py-2 rounded-lg text-xs font-bold transition-all',
              tab === 'alerts' ? 'bg-accent text-white' : 'text-[var(--color-text-dim)] hover:text-[var(--color-text)] hover:bg-[var(--color-panel-hover)]'
            )}
          >
            <Bell size={14} className="inline mr-1.5 -mt-0.5" />
            Alerts
          </button>
          <button
            onClick={() => setTab('rules')}
            className={cn(
              'px-4 py-2 rounded-lg text-xs font-bold transition-all',
              tab === 'rules' ? 'bg-accent text-white' : 'text-[var(--color-text-dim)] hover:text-[var(--color-text)] hover:bg-[var(--color-panel-hover)]'
            )}
          >
            <Shield size={14} className="inline mr-1.5 -mt-0.5" />
            Rules
          </button>
        </div>

        {tab === 'alerts' && (
          <div className="flex items-center gap-3">
            {/* Level filter */}
            <div className="flex items-center gap-1 bg-[var(--color-panel-2)] rounded-xl p-1 border border-[var(--color-border)]">
              {['all', 'critical', 'warning', 'info'].map((lvl) => (
                <button
                  key={lvl}
                  onClick={() => setFilterLevel(lvl)}
                  className={cn(
                    'px-3 py-1.5 rounded-lg text-[10px] font-bold uppercase tracking-wider transition-all',
                    filterLevel === lvl ? 'bg-accent text-white' : 'text-[var(--color-text-faint)] hover:text-[var(--color-text)]'
                  )}
                >
                  {lvl}
                </button>
              ))}
            </div>

            {/* Resolved toggle */}
            <button
              onClick={() => setShowResolved(!showResolved)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-[10px] font-bold uppercase tracking-wider border transition-all',
                showResolved
                  ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-500'
                  : 'bg-[var(--color-panel-2)] border-[var(--color-border)] text-[var(--color-text-faint)]'
              )}
            >
              <CheckCircle2 size={12} />
              Resolved
            </button>

            {/* Search */}
            <div className="relative">
              <Search size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)]" />
              <input
                type="text"
                placeholder="Search alerts..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-48 pl-8 pr-3 py-1.5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] text-[11px] text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] outline-none focus:border-accent/40 transition-colors"
              />
            </div>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {tab === 'alerts' && (
          <div className="p-6 space-y-2">
            {(_activeLoading || historyLoading) && (
              <div className="flex items-center justify-center py-20">
                <div className="w-6 h-6 border-2 border-accent/30 border-t-accent rounded-full animate-spin" />
              </div>
            )}

            {!(_activeLoading || historyLoading) && filteredAlerts.length === 0 && (
              <div className="flex flex-col items-center justify-center py-20 text-[var(--color-text-faint)]">
                <Bell size={40} className="mb-3 opacity-30" />
                <p className="text-xs font-bold uppercase tracking-wider">No alerts</p>
                <p className="text-[10px] mt-1 opacity-60">
                  {search ? 'Try a different search' : 'All clear — no matching alerts'}
                </p>
              </div>
            )}

            <AnimatePresence>
              {filteredAlerts.map((alert) => {
                const meta = getLevelMeta(alert.level || 'info')
                return (
                  <motion.div
                    key={alert.id}
                    layout
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, x: -20 }}
                    className={cn(
                      'flex items-start gap-4 px-5 py-4 rounded-xl border transition-all',
                      alert.resolved
                        ? 'bg-[var(--color-panel-2)]/50 border-[var(--color-border)] opacity-60'
                        : 'bg-[var(--color-panel-2)] border-[var(--color-border)]'
                    )}
                  >
                    <span className={cn('p-2 rounded-lg border shrink-0 mt-0.5', meta.color)}>
                      {meta.icon}
                    </span>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className={cn(
                          'text-[10px] font-bold uppercase px-1.5 py-0.5 rounded',
                          alert.resolved ? 'bg-emerald-500/10 text-emerald-500' : meta.color
                        )}>
                          {alert.resolved ? 'Resolved' : meta.label}
                        </span>
                        <span className="text-[10px] font-mono text-[var(--color-text-faint)]">
                          {alert.metric}
                        </span>
                      </div>
                      <p className="text-xs font-bold text-[var(--color-text)]">
                        {alert.message}
                      </p>
                      <div className="flex items-center gap-3 mt-1.5">
                        <span className="text-[10px] text-[var(--color-text-faint)] font-mono">
                          Value: {alert.value?.toFixed(1)} (threshold: {alert.threshold})
                        </span>
                        <span className="text-[10px] text-[var(--color-text-faint)]">
                          {formatTimestamp(alert.timestamp)}
                        </span>
                      </div>
                    </div>

                    {!alert.resolved && (
                      <button
                        onClick={() => resolveMutation.mutate(alert.id)}
                        className="p-1.5 rounded-lg text-[var(--color-text-faint)] hover:text-emerald-500 hover:bg-emerald-500/10 transition-all"
                        title="Resolve alert"
                      >
                        <CheckCircle2 size={14} />
                      </button>
                    )}
                  </motion.div>
                )
              })}
            </AnimatePresence>
          </div>
        )}

        {tab === 'rules' && (
          <div className="p-6">
            {/* Rules explanation */}
            <div className="mb-6 p-4 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
              <div className="flex items-start gap-3">
                <Shield size={16} className="text-accent mt-0.5 shrink-0" />
                <div>
                  <p className="text-xs font-bold text-[var(--color-text)] mb-1">Alert Rules</p>
                  <p className="text-[10px] text-[var(--color-text-faint)] leading-relaxed">
                    Rules define when alerts fire. Each rule monitors a metric (e.g. cpu, memory, disk)
                    and triggers when the value crosses the threshold. Default rules are added at startup.
                    Custom rules can be added from here or programmatically.
                  </p>
                </div>
              </div>
            </div>

            {rulesLoading && (
              <div className="flex items-center justify-center py-20">
                <div className="w-6 h-6 border-2 border-accent/30 border-t-accent rounded-full animate-spin" />
              </div>
            )}

            {!rulesLoading && rules.length === 0 && (
              <div className="flex flex-col items-center justify-center py-20 text-[var(--color-text-faint)]">
                <ShieldOff size={40} className="mb-3 opacity-30" />
                <p className="text-xs font-bold uppercase tracking-wider">No rules defined</p>
              </div>
            )}

            {!rulesLoading && rules.length > 0 && (
              <div className="space-y-2">
                {rules.map((rule, idx) => (
                  <div
                    key={`${rule.metric}-${rule.threshold}-${idx}`}
                    className="flex items-center gap-4 px-5 py-3 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]"
                  >
                    <div className="flex-1 grid grid-cols-5 gap-4 text-xs">
                      <div>
                        <p className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Metric</p>
                        <p className="font-mono font-bold text-[var(--color-text)]">{rule.metric}</p>
                      </div>
                      <div>
                        <p className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Condition</p>
                        <p className="font-bold text-[var(--color-text)]">{rule.condition}</p>
                      </div>
                      <div>
                        <p className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Threshold</p>
                        <p className="font-mono font-bold text-[var(--color-text)]">{rule.threshold}</p>
                      </div>
                      <div>
                        <p className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Severity</p>
                        <span className={cn(
                          'inline-block text-[10px] font-bold px-1.5 py-0.5 rounded',
                          getLevelMeta(rule.severity).color
                        )}>
                          {rule.severity}
                        </span>
                      </div>
                      <div>
                        <p className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Flap Count</p>
                        <p className="font-mono font-bold text-[var(--color-text)]">{rule.flap_count}</p>
                      </div>
                    </div>
                    <button
                      onClick={() => removeRuleMutation.mutate({ metric: rule.metric, threshold: rule.threshold })}
                      className="p-1.5 rounded-lg text-[var(--color-text-faint)] hover:text-red-500 hover:bg-red-500/10 transition-all shrink-0"
                      title="Remove rule"
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
