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
  const [showAddRule, setShowAddRule] = useState(false)
  const [newRule, setNewRule] = useState({
    metric: '',
    threshold: '',
    severity: 'warning',
    condition: 'gt',
    flapCount: '2',
    message: '',
  })

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

  const addRuleMutation = useMutation({
    mutationFn: async (input: {
      metric: string
      threshold: number
      severity: string
      condition: string
      flapCount: number
      message: string
    }) => {
      await call(
        'AlertAPI.AddRule',
        input.metric,
        input.threshold,
        input.severity,
        input.condition,
        input.flapCount,
        input.message,
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alertRules'] })
      toast.success('Rule added')
    },
    onError: () => toast.error('Failed to add rule'),
  })

  // ── Derived data ──

  const displayedAlerts = showResolved ? allAlerts : allAlerts.filter((a) => !a.resolved)
  const hasActiveFilters = filterLevel !== 'all' || showResolved || search.trim().length > 0

  const resetFilters = () => {
    setFilterLevel('all')
    setShowResolved(false)
    setSearch('')
  }

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
      <div className="grid grid-cols-2 gap-4 px-10 py-5 border-b border-[var(--color-border)] bg-[var(--color-panel-1)]/30 lg:grid-cols-3 xl:grid-cols-5">
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
      <div className="flex flex-col gap-3 px-10 py-3 border-b border-[var(--color-border)] bg-[var(--color-panel-1)]/20 lg:flex-row lg:items-center lg:justify-between">
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
            <span className="ml-2 rounded-full bg-white/15 px-1.5 py-0.5 text-[10px] font-black">{stats.active}</span>
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
            <span className="ml-2 rounded-full bg-white/15 px-1.5 py-0.5 text-[10px] font-black">{rules.length}</span>
          </button>
        </div>

        {tab === 'alerts' && (
          <div className="flex flex-wrap items-center gap-3">
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
              aria-pressed={showResolved}
            >
              <CheckCircle2 size={12} />
              Show resolved
            </button>

            {/* Search */}
            <div className="relative">
              <Search size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)]" />
              <input
                type="text"
                placeholder="Search metric, message, or id..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-60 max-w-full pl-8 pr-3 py-1.5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] text-[11px] text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] outline-none focus:border-accent/40 transition-colors"
              />
            </div>

            {hasActiveFilters && (
              <button
                onClick={resetFilters}
                className="px-3 py-1.5 rounded-xl text-[10px] font-bold uppercase tracking-wider border border-[var(--color-border)] bg-[var(--color-panel-2)] text-[var(--color-text-dim)] hover:text-[var(--color-text)] hover:bg-[var(--color-panel-hover)] transition-all"
              >
                Clear filters
              </button>
            )}
          </div>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {tab === 'alerts' && (
          <div className="p-6 space-y-2">
            {hasActiveFilters && (
              <div className="flex items-center justify-between gap-3 rounded-xl border border-[var(--color-border)] bg-[var(--color-panel-2)] px-4 py-3 text-[11px] text-[var(--color-text-dim)]">
                <span>
                  Showing {filteredAlerts.length} alert{filteredAlerts.length === 1 ? '' : 's'} in scope from {displayedAlerts.length} visible alert{displayedAlerts.length === 1 ? '' : 's'}
                </span>
                <button
                  onClick={resetFilters}
                  className="font-bold text-accent hover:underline"
                >
                  Reset
                </button>
              </div>
            )}

            {historyLoading && (
              <div className="flex items-center justify-center py-20">
                <div className="w-6 h-6 border-2 border-accent/30 border-t-accent rounded-full animate-spin" />
              </div>
            )}

            {!historyLoading && filteredAlerts.length === 0 && (
              <div className="flex flex-col items-center justify-center py-20 text-[var(--color-text-faint)]">
                <Bell size={40} className="mb-3 opacity-30" />
                <p className="text-xs font-bold uppercase tracking-wider">No alerts</p>
                <p className="text-[10px] mt-1 opacity-60">
                  {hasActiveFilters ? 'No alerts match the current filters' : 'All clear — no matching alerts'}
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
                        aria-label={`Resolve alert ${alert.id}`}
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
                <div className="flex-1">
                  <p className="text-xs font-bold text-[var(--color-text)] mb-1">Alert Rules</p>
                  <p className="text-[10px] text-[var(--color-text-faint)] leading-relaxed">
                    Rules define when alerts fire. Each rule monitors a metric (e.g. cpu, memory, disk)
                    and triggers when the value crosses the threshold. Default rules are added at startup.
                    Custom rules are persisted and survive restarts.
                  </p>
                </div>
                <button
                  onClick={() => setShowAddRule((v) => !v)}
                  className="px-3 py-1.5 rounded-lg text-[10px] font-bold bg-accent/15 text-accent hover:bg-accent/25 transition-all shrink-0"
                >
                  {showAddRule ? 'Cancel' : '+ Add Rule'}
                </button>
              </div>
            </div>

            {/* Add Rule form */}
            {showAddRule && (
              <div className="mb-6 p-4 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                <p className="text-xs font-bold text-[var(--color-text)] mb-3">New Alert Rule</p>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                  <label className="flex flex-col gap-1">
                    <span className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)]">Metric</span>
                    <input
                      value={newRule.metric}
                      onChange={(e) => setNewRule({ ...newRule, metric: e.target.value })}
                      placeholder="cpu.percent"
                      className="px-2.5 py-1.5 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] text-xs text-[var(--color-text)] focus:outline-none focus:border-accent/50"
                    />
                  </label>
                  <label className="flex flex-col gap-1">
                    <span className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)]">Condition</span>
                    <select
                      value={newRule.condition}
                      onChange={(e) => setNewRule({ ...newRule, condition: e.target.value })}
                      className="px-2.5 py-1.5 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] text-xs text-[var(--color-text)] focus:outline-none focus:border-accent/50"
                    >
                      <option value="gt">&gt; (greater than)</option>
                      <option value="lt">&lt; (less than)</option>
                    </select>
                  </label>
                  <label className="flex flex-col gap-1">
                    <span className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)]">Threshold</span>
                    <input
                      type="number"
                      value={newRule.threshold}
                      onChange={(e) => setNewRule({ ...newRule, threshold: e.target.value })}
                      placeholder="90"
                      className="px-2.5 py-1.5 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] text-xs text-[var(--color-text)] focus:outline-none focus:border-accent/50"
                    />
                  </label>
                  <label className="flex flex-col gap-1">
                    <span className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)]">Severity</span>
                    <select
                      value={newRule.severity}
                      onChange={(e) => setNewRule({ ...newRule, severity: e.target.value })}
                      className="px-2.5 py-1.5 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] text-xs text-[var(--color-text)] focus:outline-none focus:border-accent/50"
                    >
                      <option value="info">Info</option>
                      <option value="warning">Warning</option>
                      <option value="critical">Critical</option>
                    </select>
                  </label>
                  <label className="flex flex-col gap-1">
                    <span className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)]">Flap Count</span>
                    <input
                      type="number"
                      min={1}
                      value={newRule.flapCount}
                      onChange={(e) => setNewRule({ ...newRule, flapCount: e.target.value })}
                      placeholder="2"
                      className="px-2.5 py-1.5 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] text-xs text-[var(--color-text)] focus:outline-none focus:border-accent/50"
                    />
                  </label>
                  <label className="flex flex-col gap-1 col-span-2 md:col-span-3">
                    <span className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)]">Message (optional)</span>
                    <input
                      value={newRule.message}
                      onChange={(e) => setNewRule({ ...newRule, message: e.target.value })}
                      placeholder="CPU usage high: {value}% (threshold {threshold})"
                      className="px-2.5 py-1.5 rounded-lg bg-[var(--color-bg)] border border-[var(--color-border)] text-xs text-[var(--color-text)] focus:outline-none focus:border-accent/50"
                    />
                  </label>
                </div>
                <div className="mt-4 flex items-center justify-end gap-2">
                  <button
                    onClick={() => setShowAddRule(false)}
                    className="px-3 py-1.5 rounded-lg text-[10px] font-bold text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-all"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => {
                      const threshold = parseFloat(newRule.threshold)
                      const flapCount = parseInt(newRule.flapCount, 10)
                      if (!newRule.metric.trim()) {
                        toast.error('Metric is required')
                        return
                      }
                      if (Number.isNaN(threshold)) {
                        toast.error('Threshold must be a number')
                        return
                      }
                      addRuleMutation.mutate({
                        metric: newRule.metric.trim(),
                        threshold,
                        severity: newRule.severity,
                        condition: newRule.condition,
                        flapCount: Number.isNaN(flapCount) || flapCount < 1 ? 1 : flapCount,
                        message: newRule.message.trim(),
                      })
                      setNewRule({ metric: '', threshold: '', severity: 'warning', condition: 'gt', flapCount: '2', message: '' })
                      setShowAddRule(false)
                    }}
                    disabled={addRuleMutation.isPending}
                    className="px-4 py-1.5 rounded-lg text-[10px] font-bold bg-accent text-white hover:bg-accent/90 transition-all disabled:opacity-50"
                  >
                    {addRuleMutation.isPending ? 'Adding…' : 'Add Rule'}
                  </button>
                </div>
              </div>
            )}

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
                    <div className="flex-1 grid grid-cols-2 gap-4 text-xs lg:grid-cols-5">
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
                      aria-label={`Remove alert rule for ${rule.metric} at ${rule.threshold}`}
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
