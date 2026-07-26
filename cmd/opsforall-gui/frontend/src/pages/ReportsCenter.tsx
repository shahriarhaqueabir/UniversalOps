import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  FileText, Shield, HeartPulse, Bot,
  Search, Trash2, Download, ChevronRight,
  AlertTriangle, CheckCircle, XCircle,
  RefreshCw, EyeOff, Plus, Clock,
} from 'lucide-react'
import { motion, AnimatePresence } from 'motion/react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { toast } from 'sonner'
import type { ReportRecord, PrebuiltTemplate } from '@/types'
import { app, common } from '../../wailsjs/go/models'

const TYPE_META: Record<string, { label: string; icon: React.ReactNode; color: string }> = {
  health: {
    label: 'Health',
    icon: <HeartPulse size={14} />,
    color: 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20',
  },
  security: {
    label: 'Security',
    icon: <Shield size={14} />,
    color: 'text-amber-500 bg-amber-500/10 border-amber-500/20',
  },
  auto_diag: {
    label: 'Auto-Diagnostic',
    icon: <Bot size={14} />,
    color: 'text-violet-500 bg-violet-500/10 border-violet-500/20',
  },
}

function getTypeMeta(type: string) {
  return TYPE_META[type] ?? {
    label: type,
    icon: <FileText size={14} />,
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
    })
  } catch {
    return ts
  }
}

function getScoreColor(score: number): string {
  if (score >= 90) return 'text-emerald-500'
  if (score >= 70) return 'text-amber-500'
  return 'text-red-500'
}

function getScoreIcon(score: number): React.ReactNode {
  if (score >= 90) return <CheckCircle size={16} className="text-emerald-500" />
  if (score >= 70) return <AlertTriangle size={16} className="text-amber-500" />
  return <XCircle size={16} className="text-red-500" />
}

type ReportFilter = 'all' | 'health' | 'security' | 'auto_diag'

export function ReportsCenter() {
  const { call } = useBackend()
  const queryClient = useQueryClient()
  const [filter, setFilter] = useState<ReportFilter>('all')
  const [search, setSearch] = useState('')
  const [selectedReport, setSelectedReport] = useState<ReportRecord | null>(null)
  const [showDetail, setShowDetail] = useState(false)
  const [showRules, setShowRules] = useState(false)
  const [showGenerateDropdown, setShowGenerateDropdown] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  // Auto-report rule form state
  const [ruleForm, setRuleForm] = useState({ name: '', description: '', metric: '', condition: 'GT', threshold: 0, reportType: 'health', schedule: 'on_alert', templateId: '' })

  const { data: reports = [], isLoading } = useQuery<ReportRecord[]>({
    queryKey: ['reports'],
    queryFn: async () => (await call('ReportsAPI.ListAllReports')) as ReportRecord[] || [],
    refetchInterval: 15_000,
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const ok = await call('ReportsAPI.DeleteReport', id)
      if (!ok) throw new Error('Delete failed')
    },
    onSuccess: (_data, id) => {
      queryClient.setQueryData<ReportRecord[]>(['reports'], (old) =>
        old?.filter((r) => r.id !== id) ?? []
      )
      if (selectedReport?.id === id) {
        setSelectedReport(null)
        setShowDetail(false)
      }
      toast.success('Report deleted')
    },
    onError: () => toast.error('Failed to delete report'),
  })

  // Generate report mutation
  const generateMutation = useMutation({
    mutationFn: async (type: string) => {
      const result = await call('ReportsAPI.GenerateReport', type)
      return result as app.ReportGenerationResult
    },
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['reports'] })
      if (result) {
        toast.success(`${result.type} report generated (score: ${result.score})`)
      } else {
        toast.success('Report generated')
      }
    },
    onError: (err: any) => toast.error(`Generation failed: ${err?.message ?? 'unknown error'}`),
  })

  // Report types (for the generate dropdown)
  const { data: reportTypes = [] } = useQuery<app.ReportTypeMeta[]>({
    queryKey: ['reportTypes'],
    queryFn: async () => (await call('ReportsAPI.GetReportTypes')) as app.ReportTypeMeta[] || [],
    staleTime: 60_000,
  })

  // Auto-report rules
  const { data: rules = [] } = useQuery<common.AutoReportRule[]>({
    queryKey: ['reportRules'],
    queryFn: async () => (await call('ReportsAPI.GetReportRules')) as common.AutoReportRule[] || [],
    refetchInterval: 30_000,
  })

  // Prebuilt report templates
  const { data: templates = [] } = useQuery<PrebuiltTemplate[]>({
    queryKey: ['prebuiltTemplates'],
    queryFn: async () => (await call('ReportsAPI.GetPrebuiltTemplates')) as PrebuiltTemplate[] || [],
    staleTime: 120_000,
  })

  // Group templates by category for the dropdown
  const templatesByCategory = templates.reduce<Record<string, PrebuiltTemplate[]>>((acc, t) => {
    (acc[t.category] ??= []).push(t)
    return acc
  }, {})

  // Handle template selection: auto-fill rule form fields
  const handleTemplateSelect = useCallback((templateId: string) => {
    if (!templateId) {
      setRuleForm((f) => ({ ...f, templateId: '', name: '', description: '', metric: '' }))
      return
    }
    const t = templates.find((tmpl) => tmpl.id === templateId)
    if (!t) return
    setRuleForm({
      name: t.preset_name,
      description: t.description,
      metric: t.metric,
      condition: t.condition,
      threshold: t.threshold,
      reportType: t.report_type,
      schedule: t.schedule,
      templateId: t.id,
    })
  }, [templates])

  const addRuleMutation = useMutation({
    mutationFn: async () => {
      const r = ruleForm
      await call('ReportsAPI.AddReportRule', r.name, r.description, r.metric, r.condition, r.threshold, r.reportType, r.schedule)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reportRules'] })
      setRuleForm({ name: '', description: '', metric: '', condition: 'GT', threshold: 0, reportType: 'health', schedule: 'on_alert', templateId: '' })
      toast.success('Auto-report rule created')
    },
    onError: (err: any) => toast.error(`Failed to create rule: ${err?.message ?? 'unknown error'}`),
  })

  const toggleRuleMutation = useMutation({
    mutationFn: async (rule: common.AutoReportRule) => {
      await call('ReportsAPI.UpdateReportRule', rule.id, rule.name, rule.description, !rule.enabled)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reportRules'] })
      toast.success('Rule updated')
    },
    onError: (err: any) => toast.error(`Failed to update rule: ${err?.message ?? 'unknown error'}`),
  })

  const deleteRuleMutation = useMutation({
    mutationFn: async (id: string) => {
      const ok = await call('ReportsAPI.DeleteReportRule', id)
      if (!ok) throw new Error('Delete failed')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reportRules'] })
      toast.success('Rule deleted')
    },
    onError: (err: any) => toast.error(`Failed to delete rule: ${err?.message ?? 'unknown error'}`),
  })

  const filtered = reports.filter((r) => {
    if (filter !== 'all' && r.type !== filter) return false
    if (search) {
      const q = search.toLowerCase()
      if (!r.id.toLowerCase().includes(q) && !r.type.toLowerCase().includes(q)) return false
    }
    return true
  })

  const handleExport = useCallback((report: ReportRecord) => {
    try {
      const data = {
        id: report.id,
        timestamp: report.timestamp,
        type: report.type,
        score: report.score,
        data_json: report.data_json ? JSON.parse(report.data_json) : null,
      }
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `report-${report.type}-${report.id.replace(/[^a-zA-Z0-9_-]/g, '_')}.json`
      a.click()
      URL.revokeObjectURL(url)
      toast.success('Report exported')
    } catch {
      toast.error('Export failed')
    }
  }, [])

  const handleExportAll = useCallback(() => {
    try {
      const data = filtered.map((r) => ({
        id: r.id,
        timestamp: r.timestamp,
        type: r.type,
        score: r.score,
      }))
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `all-reports-${new Date().toISOString().slice(0, 10)}.json`
      a.click()
      URL.revokeObjectURL(url)
      toast.success('All reports exported')
    } catch {
      toast.error('Export failed')
    }
  }, [filtered])

  // Aggregate stats
  const totalCount = reports.length
  const avgScore = totalCount > 0
    ? Math.round(reports.reduce((sum, r) => sum + r.score, 0) / totalCount)
    : 0
  const typeCounts = {
    health: reports.filter((r) => r.type === 'health').length,
    security: reports.filter((r) => r.type === 'security').length,
    auto_diag: reports.filter((r) => r.type === 'auto_diag').length,
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500 overflow-hidden">
      {/* Header */}
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 px-10 flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
              <FileText size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Insights</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">Reports Center</p>
          <p className="text-xs text-[var(--color-text-faint)] mt-1 font-medium tracking-wide">
            {totalCount} total reports &middot; Avg score {avgScore}%
          </p>
        </div>
        <div className="flex items-center gap-2">
          {/* Generate Report dropdown */}
          <div className="relative">
            <button
              onClick={() => setShowGenerateDropdown(!showGenerateDropdown)}
              disabled={generateMutation.isPending}
              className="flex items-center gap-2 px-4 py-2 rounded-xl bg-accent text-white text-xs font-bold hover:brightness-110 transition-all disabled:opacity-40"
            >
              {generateMutation.isPending ? (
                <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <Plus size={14} />
              )}
              Generate Report
            </button>
            {showGenerateDropdown && (
              <>
                <div className="fixed inset-0 z-10" onClick={() => setShowGenerateDropdown(false)} />
                <div className="absolute right-0 top-full mt-2 z-20 w-64 p-2 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] shadow-xl shadow-black/20">
                  {reportTypes.length === 0 && (
                    <p className="px-3 py-4 text-[10px] text-[var(--color-text-faint)] text-center">No report types available</p>
                  )}
                  {reportTypes.map((rt) => (
                    <button
                      key={rt.type}
                      disabled={!rt.available}
                      onClick={() => {
                        setShowGenerateDropdown(false)
                        generateMutation.mutate(rt.type)
                      }}
                      className={cn(
                        'w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-xs transition-all text-left',
                        rt.available
                          ? 'text-[var(--color-text)] hover:bg-accent/10'
                          : 'text-[var(--color-text-faint)] opacity-40 cursor-not-allowed'
                      )}
                    >
                      <span className={cn('p-1 rounded-lg border', getTypeMeta(rt.type).color)}>
                        {getTypeMeta(rt.type).icon}
                      </span>
                      <div className="flex-1 min-w-0">
                        <p className="font-bold">{rt.label}</p>
                        <p className="text-[10px] text-[var(--color-text-faint)] truncate">{rt.description}</p>
                      </div>
                    </button>
                  ))}
                </div>
              </>
            )}
          </div>

          {/* Rules toggle */}
          <button
            onClick={() => setShowRules(!showRules)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-xl border text-xs font-bold transition-all',
              showRules
                ? 'bg-accent/10 border-accent/30 text-accent'
                : 'bg-[var(--color-panel-2)] border-[var(--color-border)] text-[var(--color-text)] hover:bg-[var(--color-panel-hover)]'
            )}
          >
            <Clock size={14} />
            Rules
          </button>

          <button
            onClick={handleExportAll}
            disabled={filtered.length === 0}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs font-bold text-[var(--color-text)] hover:bg-[var(--color-panel-hover)] transition-all disabled:opacity-40"
          >
            <Download size={14} />
            Export All
          </button>
          <button
            onClick={() => { queryClient.invalidateQueries({ queryKey: ['reports'] }) }}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs font-bold text-[var(--color-text)] hover:bg-[var(--color-panel-hover)] transition-all"
          >
            <RefreshCw size={14} />
            Refresh
          </button>
        </div>
      </div>

      {/* Stats bar */}
      <div className="grid grid-cols-4 gap-4 px-10 py-5 border-b border-[var(--color-border)] bg-[var(--color-panel-1)]/30">
        {(['all', 'health', 'security', 'auto_diag'] as ReportFilter[]).map((key) => {
          const label = key === 'all' ? 'All Reports' : getTypeMeta(key).label
          const count = key === 'all' ? totalCount : typeCounts[key]
          const isActive = filter === key
          const meta = key !== 'all' ? getTypeMeta(key) : null
          return (
            <button
              key={key}
              onClick={() => { setFilter(key); setSelectedReport(null); setShowDetail(false) }}
              className={cn(
                'flex items-center gap-3 px-4 py-3 rounded-xl border transition-all text-left',
                isActive
                  ? 'bg-accent/10 border-accent/30 text-accent'
                  : 'bg-[var(--color-panel-2)] border-[var(--color-border)] text-[var(--color-text)] hover:bg-[var(--color-panel-hover)]'
              )}
            >
              {meta && (
                <span className={cn('p-1.5 rounded-lg border', meta.color)}>
                  {meta.icon}
                </span>
              )}
              {!meta && (
                <span className="p-1.5 rounded-lg border text-accent border-accent/20 bg-accent/10">
                  <FileText size={14} />
                </span>
              )}
              <div>
                <p className="text-xs font-bold uppercase tracking-wide">{label}</p>
                <p className={cn('text-lg font-black', isActive ? 'text-accent' : 'text-[var(--color-text)]')}>
                  {count}
                </p>
              </div>
            </button>
          )
        })}
      </div>

      {/* Auto-Report Rules panel */}
      <AnimatePresence>
        {showRules && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
            className="overflow-hidden border-b border-[var(--color-border)] bg-[var(--color-panel-1)]/20"
          >
            <div className="px-10 py-5 space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Clock size={14} className="text-accent" />
                  <span className="text-xs font-bold uppercase tracking-wider text-[var(--color-text)]">Auto-Report Rules</span>
                  <span className="text-[10px] text-[var(--color-text-faint)] ml-2">
                    {rules.filter((r) => r.enabled).length} enabled &middot; {rules.length} total
                  </span>
                </div>
              </div>

              {/* Existing rules */}
              {rules.length > 0 && (
                <div className="space-y-2">
                  {rules.map((rule) => (
                    <div key={rule.id} className="flex items-center gap-4 px-4 py-3 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                      <button
                        onClick={() => toggleRuleMutation.mutate(rule)}
                        className={cn(
                          'p-1 rounded-lg transition-colors',
                          rule.enabled
                            ? 'text-emerald-500 hover:bg-emerald-500/10'
                            : 'text-[var(--color-text-faint)] hover:bg-[var(--color-panel-hover)]'
                        )}
                        title={rule.enabled ? 'Disable rule' : 'Enable rule'}
                      >
                        {rule.enabled ? <CheckCircle size={14} /> : <XCircle size={14} />}
                      </button>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <p className="text-xs font-bold text-[var(--color-text)]">{rule.name}</p>
                          <span className={cn(
                            'px-1.5 py-0.5 rounded text-[9px] font-bold uppercase',
                            getTypeMeta(rule.report_type).color
                          )}>
                            {rule.report_type}
                          </span>
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-[var(--color-panel-1)] text-[var(--color-text-faint)]">
                            {rule.schedule}
                          </span>
                        </div>
                        {rule.description && (
                          <p className="text-[10px] text-[var(--color-text-faint)] mt-0.5 truncate">{rule.description}</p>
                        )}
                      </div>
                      <button
                        onClick={() => deleteRuleMutation.mutate(rule.id)}
                        className="p-1.5 rounded-lg text-[var(--color-text-faint)] hover:text-red-500 hover:bg-red-500/10 transition-colors"
                        title="Delete rule"
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  ))}
                </div>
              )}

              {/* New rule form */}
              <div className="grid grid-cols-7 gap-3 items-end">
                {/* Template selector — replaces manual name/metric entry */}
                <div className="col-span-4">
                  <label className="block text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Prebuilt Template</label>
                  <select
                    value={ruleForm.templateId}
                    onChange={(e) => handleTemplateSelect(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] outline-none focus:border-accent/40 transition-colors"
                  >
                    <option value="">Custom Rule (manual)</option>
                    {(['health', 'security', 'performance'] as const).map((cat) => {
                      const catLabel = cat.charAt(0).toUpperCase() + cat.slice(1)
                      const items = templatesByCategory[cat]
                      if (!items?.length) return null
                      return (
                        <optgroup key={cat} label={catLabel}>
                          {items.map((t) => (
                            <option key={t.id} value={t.id}>{t.preset_name}</option>
                          ))}
                        </optgroup>
                      )
                    })}
                  </select>
                </div>
                <div className="col-span-3">
                  <label className="block text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Name</label>
                  <input
                    value={ruleForm.name}
                    onChange={(e) => setRuleForm((f) => ({ ...f, name: e.target.value, templateId: '' }))}
                    placeholder="e.g. High CPU alert"
                    className="w-full px-3 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] outline-none focus:border-accent/40 transition-colors"
                  />
                </div>
                <div className="col-span-2">
                  <label className="block text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Metric</label>
                  <input
                    value={ruleForm.metric}
                    onChange={(e) => setRuleForm((f) => ({ ...f, metric: e.target.value, templateId: '' }))}
                    placeholder="e.g. cpu_usage"
                    className="w-full px-3 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] outline-none focus:border-accent/40 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Condition</label>
                  <select
                    value={ruleForm.condition}
                    onChange={(e) => setRuleForm((f) => ({ ...f, condition: e.target.value }))}
                    className="w-full px-3 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] outline-none focus:border-accent/40 transition-colors"
                  >
                    <option value="GT">&gt; (above)</option>
                    <option value="LT">&lt; (below)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Threshold</label>
                  <input
                    type="number"
                    value={ruleForm.threshold}
                    onChange={(e) => setRuleForm((f) => ({ ...f, threshold: parseFloat(e.target.value) || 0 }))}
                    className="w-full px-3 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] outline-none focus:border-accent/40 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Report Type</label>
                  <select
                    value={ruleForm.reportType}
                    onChange={(e) => setRuleForm((f) => ({ ...f, reportType: e.target.value }))}
                    className="w-full px-3 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] outline-none focus:border-accent/40 transition-colors"
                  >
                    <option value="health">Health</option>
                    <option value="security">Security</option>
                    <option value="auto_diag">Auto-Diagnostic</option>
                  </select>
                </div>
                <div>
                  <label className="block text-[9px] font-bold uppercase tracking-wider text-[var(--color-text-faint)] mb-1">Schedule</label>
                  <select
                    value={ruleForm.schedule}
                    onChange={(e) => setRuleForm((f) => ({ ...f, schedule: e.target.value }))}
                    className="w-full px-3 py-2 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] outline-none focus:border-accent/40 transition-colors"
                  >
                    <option value="on_alert">On Alert</option>
                    <option value="hourly">Hourly</option>
                    <option value="daily">Daily</option>
                  </select>
                </div>
                <div>
                  <button
                    onClick={() => addRuleMutation.mutate()}
                    disabled={!ruleForm.name || addRuleMutation.isPending}
                    className="w-full flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg bg-accent text-white text-xs font-bold hover:brightness-110 transition-all disabled:opacity-40"
                  >
                    {addRuleMutation.isPending ? (
                      <div className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    ) : (
                      <Plus size={12} />
                    )}
                    Add Rule
                  </button>
                </div>
              </div>
              <p className="text-[9px] text-[var(--color-text-faint)] leading-relaxed">
                Rules with <strong className="text-[var(--color-text)]">on_alert</strong> schedule trigger a report when an alert fires.
                <strong className="text-[var(--color-text)]"> Hourly</strong> and <strong className="text-[var(--color-text)]">daily</strong> rules generate reports on a timer.
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Search + content */}
      <div className="flex-1 flex overflow-hidden">
        {/* List */}
        <div className={cn(
          'flex flex-col border-r border-[var(--color-border)] overflow-hidden transition-all duration-300',
          showDetail ? 'w-1/2' : 'flex-1'
        )}>
          {/* Search */}
          <div className="px-6 py-4 border-b border-[var(--color-border)]">
            <div className="relative">
              <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)]" />
              <input
                type="text"
                placeholder="Search reports by ID or type..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full pl-9 pr-4 py-2 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs text-[var(--color-text)] placeholder:text-[var(--color-text-faint)] outline-none focus:border-accent/40 transition-colors"
              />
            </div>
          </div>

          {/* Report list */}
          <div className="flex-1 overflow-y-auto p-4 space-y-2">
            {isLoading && (
              <div className="flex items-center justify-center py-20">
                <div className="w-6 h-6 border-2 border-accent/30 border-t-accent rounded-full animate-spin" />
              </div>
            )}

            {!isLoading && filtered.length === 0 && (
              <div className="flex flex-col items-center justify-center py-20 text-[var(--color-text-faint)]">
                <FileText size={40} className="mb-3 opacity-30" />
                <p className="text-xs font-bold uppercase tracking-wider">No reports found</p>
                <p className="text-[10px] mt-1 opacity-60">
                  {search ? 'Try a different search term' : 'Run a diagnostic or security audit to generate reports'}
                </p>
              </div>
            )}

            <AnimatePresence>
              {filtered.map((report) => {
                const meta = getTypeMeta(report.type)
                const isSelected = selectedReport?.id === report.id
                return (
                  <motion.button
                    key={report.id}
                    layout
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, x: -20 }}
                    onClick={() => {
                      setSelectedReport(report)
                      setShowDetail(true)
                    }}
                    className={cn(
                      'w-full flex items-center gap-4 px-4 py-3 rounded-xl border transition-all text-left group',
                      isSelected
                        ? 'bg-accent/10 border-accent/30'
                        : 'bg-[var(--color-panel-2)] border-[var(--color-border)] hover:bg-[var(--color-panel-hover)] hover:border-[var(--color-border-hover)]'
                    )}
                  >
                    <span className={cn('p-1.5 rounded-lg border shrink-0', meta.color)}>
                      {meta.icon}
                    </span>

                    <div className="flex-1 min-w-0">
                      <p className="text-xs font-bold text-[var(--color-text)] truncate">
                        {meta.label} Report
                      </p>
                      <p className="text-[10px] text-[var(--color-text-faint)] font-mono truncate mt-0.5">
                        {formatTimestamp(report.timestamp)}
                      </p>
                    </div>

                    <div className="flex items-center gap-3 shrink-0">
                      <span className={cn('text-xs font-black flex items-center gap-1', getScoreColor(report.score))}>
                        {getScoreIcon(report.score)}
                        {report.score}
                      </span>

                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleExport(report)
                        }}
                        className="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity text-[var(--color-text-faint)] hover:text-accent hover:bg-accent/10"
                        title="Export report"
                      >
                        <Download size={12} />
                      </button>

                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          setDeleteTarget(report.id)
                        }}
                        className="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity text-[var(--color-text-faint)] hover:text-red-500 hover:bg-red-500/10"
                        title="Delete report"
                      >
                        <Trash2 size={12} />
                      </button>

                      <ChevronRight size={14} className={cn(
                        'text-[var(--color-text-faint)] transition-transform',
                        isSelected && 'rotate-90'
                      )} />
                    </div>
                  </motion.button>
                )
              })}
            </AnimatePresence>
          </div>
        </div>

        {/* Detail panel */}
        <AnimatePresence>
          {showDetail && selectedReport && (
            <motion.div
              initial={{ width: 0, opacity: 0 }}
              animate={{ width: '50%', opacity: 1 }}
              exit={{ width: 0, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 300, damping: 30 }}
              className="flex flex-col overflow-hidden border-l border-[var(--color-border)]"
            >
              <div className="px-6 py-4 border-b border-[var(--color-border)] flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className={cn('p-1.5 rounded-lg border', getTypeMeta(selectedReport.type).color)}>
                    {getTypeMeta(selectedReport.type).icon}
                  </span>
                  <div>
                    <p className="text-xs font-bold text-[var(--color-text)]">
                      {getTypeMeta(selectedReport.type).label} Report
                    </p>
                    <p className="text-[10px] text-[var(--color-text-faint)] font-mono">
                      {selectedReport.id}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => setShowDetail(false)}
                  className="p-1.5 rounded-lg text-[var(--color-text-faint)] hover:text-[var(--color-text)] hover:bg-[var(--color-panel-hover)]"
                >
                  <EyeOff size={14} />
                </button>
              </div>

              <div className="flex-1 overflow-y-auto p-6 space-y-6">
                {/* Score card */}
                <div className="p-5 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                  <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--color-text-faint)] mb-3">Report Score</p>
                  <div className="flex items-center gap-4">
                    <div className={cn(
                      'w-16 h-16 rounded-2xl flex items-center justify-center text-2xl font-black border',
                      selectedReport.score >= 90 ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-500' :
                      selectedReport.score >= 70 ? 'bg-amber-500/10 border-amber-500/30 text-amber-500' :
                      'bg-red-500/10 border-red-500/30 text-red-500'
                    )}>
                      {selectedReport.score}
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs font-bold text-[var(--color-text)]">
                        {selectedReport.score >= 90 ? 'Excellent' :
                         selectedReport.score >= 70 ? 'Fair' :
                         'Critical'}
                      </p>
                      <p className="text-[10px] text-[var(--color-text-faint)]">
                        Recorded {formatTimestamp(selectedReport.timestamp)}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Raw data */}
                <div className="space-y-3">
                  <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--color-text-faint)]">Raw Data</p>
                  <div className="p-4 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] overflow-x-auto">
                    <pre className="text-[10px] leading-relaxed text-[var(--color-text-dim)] font-mono whitespace-pre-wrap max-h-[400px] overflow-y-auto">
                      {(() => {
                        try {
                          return JSON.stringify(JSON.parse(selectedReport.data_json || '{}'), null, 2)
                        } catch {
                          return selectedReport.data_json || 'No data available'
                        }
                      })()}
                    </pre>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-3 pt-2">
                  <button
                    onClick={() => handleExport(selectedReport)}
                    className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-accent text-white text-xs font-bold hover:brightness-110 transition-all"
                  >
                    <Download size={14} />
                    Export JSON
                  </button>
                  <button
                    onClick={() => setDeleteTarget(selectedReport.id)}
                    className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-500 text-xs font-bold hover:bg-red-500/20 transition-all"
                  >
                    <Trash2 size={14} />
                    Delete
                  </button>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
      <ConfirmDialog
        open={deleteTarget !== null}
        title="Delete Report"
        description="Delete this report? This action cannot be undone."
        type="danger"
        confirmText="Delete"
        onConfirm={() => {
          if (deleteTarget) deleteMutation.mutate(deleteTarget)
        }}
        onClose={() => setDeleteTarget(null)}
      />
    </div>
  )
}
