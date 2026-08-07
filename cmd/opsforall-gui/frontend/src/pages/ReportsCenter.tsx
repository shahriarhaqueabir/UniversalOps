import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  FileText, Shield, HeartPulse, Bot,
  Search, Trash2, Download, ChevronRight,
  AlertTriangle, CheckCircle, XCircle,
  RefreshCw, EyeOff, Plus, Clock, FileSpreadsheet,
} from 'lucide-react'
import { motion, AnimatePresence } from 'motion/react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { toast } from 'sonner'
import { toCSV, downloadText, safeFilename, reportToCSVRows } from '@/lib/export'
import type { ReportRecord, PrebuiltTemplate } from '@/types'
import { app, common } from '@wailsjs/go/models'

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

// ── Human-Readable Report Formatters ─────────────────────────────────────────

interface HealthCheckData {
  name: string
  status: string
  message: string
  value: string
}

interface HealthReportData {
  id: string
  checks: HealthCheckData[]
  score: number
  timestamp: string
}

interface SecAuditItem {
  category: string
  check: string
  passed: boolean
  description: string
  remediation: string
}

interface SecurityReportData {
  id: string
  score: number
  total: number
  passed: number
  failed: number
  items: SecAuditItem[]
  timestamp: string
}

interface AutoDiagReportData {
  sections: string[]
  content: string
  generated_by: string
}

/** Render a health report in human-readable format. */
function HealthReportView({ data }: { data: HealthReportData }) {
  const statusIcon = (s: string) => {
    switch (s) {
      case 'pass': return <CheckCircle size={12} className="text-emerald-500 shrink-0" />
      case 'warn': return <AlertTriangle size={12} className="text-amber-500 shrink-0" />
      default: return <XCircle size={12} className="text-red-500 shrink-0" />
    }
  }
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3 mb-4">
        <span className="text-lg font-black">{data.score}</span>
        <span className="text-[10px] text-text-dim font-medium">/ 100 — {data.checks?.length ?? 0} checks</span>
      </div>
      {(data.checks ?? []).map((c) => (
        <div key={c.name} className="flex items-start gap-3 px-3 py-2.5 rounded-lg bg-black/20 border border-white/5">
          {statusIcon(c.status)}
          <div className="flex-1 min-w-0">
            <p className="text-[11px] font-bold text-white">{c.name}</p>
            <p className="text-[10px] text-text-dim mt-0.5 leading-relaxed">{c.message}</p>
          </div>
          {c.value && (
            <span className="text-[10px] font-mono font-bold text-accent shrink-0">{c.value}</span>
          )}
        </div>
      ))}
    </div>
  )
}

/** Render a security audit report in human-readable format. */
function SecurityReportView({ data }: { data: SecurityReportData }) {
  const byCategory = (data.items ?? []).reduce<Record<string, SecAuditItem[]>>((acc, item) => {
    (acc[item.category] ??= []).push(item)
    return acc
  }, {})
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-4 mb-4">
        <span className="text-lg font-black">{data.score}</span>
        <span className="text-[10px] text-text-dim font-medium">/ 100</span>
        <span className="text-[10px] font-bold text-emerald-500">{data.passed} passed</span>
        <span className="text-[10px] font-bold text-red-500">{data.failed} failed</span>
      </div>
      {Object.entries(byCategory).map(([category, items]) => (
        <div key={category}>
          <p className="text-[9px] font-black uppercase tracking-widest text-text-faint mb-2">{category}</p>
          <div className="space-y-1.5">
            {items.map((item) => (
              <div key={item.check} className="flex items-start gap-2.5 px-3 py-2 rounded-lg bg-black/20 border border-white/5">
                {item.passed
                  ? <CheckCircle size={11} className="text-emerald-500 shrink-0 mt-0.5" />
                  : <XCircle size={11} className="text-red-500 shrink-0 mt-0.5" />
                }
                <div className="flex-1 min-w-0">
                  <p className="text-[10px] font-bold text-white">{item.check}</p>
                  {item.description && (
                    <p className="text-[9px] text-text-dim mt-0.5">{item.description}</p>
                  )}
                  {!item.passed && item.remediation && (
                    <p className="text-[9px] text-amber-400/80 mt-0.5 italic">
                      Fix: {item.remediation}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

/** Render an auto-diagnostic report in human-readable format. */
function AutoDiagReportView({ data }: { data: AutoDiagReportData }) {
  // Split content on section headers (## ...)
  const parts = (data.content ?? '').split(/(?=## )/).filter(Boolean)
  // If content has markdown sections, render them; otherwise show sections list
  return (
    <div className="space-y-4">
      {parts.length > 1 ? (
        parts.map((part, i) => {
          // Extract the header line (## Title) from the part
          const headerMatch = part.match(/^##\s+(.+)/m)
          const title = headerMatch?.[1] ?? `Section ${i + 1}`
          const body = part.replace(/^##\s+.+/, '').trim()
          return (
            <div key={title}>
              <p className="text-[10px] font-bold text-white uppercase tracking-wider mb-1.5">{title}</p>
              <div className="text-[10px] text-text-dim leading-relaxed whitespace-pre-wrap font-sans">
                {body || 'No content'}
              </div>
            </div>
          )
        })
      ) : (
        <>
          <p className="text-[9px] font-black uppercase tracking-widest text-text-faint mb-2">Sections</p>
          {(data.sections ?? []).map((s) => (
            <div key={s} className="px-3 py-2 rounded-lg bg-black/20 border border-white/5">
              <p className="text-[10px] font-bold text-white">{s}</p>
            </div>
          ))}
          {data.content && (
            <>
              <p className="text-[9px] font-black uppercase tracking-widest text-text-faint mt-4 mb-2">Analysis</p>
              <div className="px-3 py-3 rounded-lg bg-black/20 border border-white/5 text-[10px] text-text-dim leading-relaxed whitespace-pre-wrap font-sans">
                {data.content}
              </div>
            </>
          )}
        </>
      )}
      <div className="pt-2 text-[8px] text-text-faint font-mono">
        Generated by {data.generated_by}
      </div>
    </div>
  )
}

/** Parse raw data_json string into a human-readable component based on report type. */
function ReportHumanReadable({ type, rawJson }: { type: string; rawJson: string }) {
  if (!rawJson || rawJson === 'No data available') {
    return <p className="text-[10px] text-text-faint italic">No data available</p>
  }
  let parsed: unknown
  try {
    parsed = JSON.parse(rawJson)
  } catch {
    return <p className="text-[10px] text-text-dim italic">{rawJson}</p>
  }
  switch (type) {
    case 'health':
      return <HealthReportView data={parsed as HealthReportData} />
    case 'security':
      return <SecurityReportView data={parsed as SecurityReportData} />
    case 'auto_diag':
      return <AutoDiagReportView data={parsed as AutoDiagReportData} />
    default:
      return (
        <pre className="text-[10px] leading-relaxed text-text-dim font-mono whitespace-pre-wrap">
          {JSON.stringify(parsed, null, 2)}
        </pre>
      )
  }
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

  // Fetch the full report (with data_json) when a report is selected for detail view
  const { data: fullReport } = useQuery<ReportRecord | null>({
    queryKey: ['report', selectedReport?.id],
    queryFn: async () => {
      if (!selectedReport?.id) return null
      return (await call('ReportsAPI.GetReport', selectedReport.id)) as ReportRecord | null
    },
    enabled: !!selectedReport?.id && showDetail,
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
  // Using regular function — React Compiler handles optimization automatically
  const handleTemplateSelect = (templateId: string) => {
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
  }

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

  // EXP-01: CSV export of a single report (flattened data_json).
  const handleExportCSV = useCallback((report: ReportRecord) => {
    try {
      const rows = reportToCSVRows(report)
      downloadText(
        `report-${report.type}-${safeFilename(report.id)}.csv`,
        toCSV(rows),
        'text/csv'
      )
      toast.success('Report exported as CSV')
    } catch {
      toast.error('CSV export failed')
    }
  }, [])

  // EXP-01: PDF export via browser print-to-PDF on a printable view.
  const handleExportPDF = useCallback((report: ReportRecord) => {
    try {
      const win = window.open('', '_blank', 'width=800,height=900')
      if (!win) {
        toast.error('Popup blocked — allow popups to export PDF')
        return
      }
      const meta = getTypeMeta(report.type)
      win.document.write(`<!doctype html><html><head><title>${meta.label} Report</title>
        <style>
          body { font-family: -apple-system, Segoe UI, Roboto, sans-serif; padding: 40px; color: #111; }
          h1 { font-size: 20px; margin: 0 0 4px; }
          .meta { color: #666; font-size: 12px; margin-bottom: 24px; }
          .score { font-size: 28px; font-weight: 800; margin-bottom: 24px; }
          pre { background: #f5f5f5; padding: 16px; border-radius: 8px; font-size: 11px; white-space: pre-wrap; word-break: break-word; }
        </style></head><body>
        <h1>${meta.label} Report</h1>
        <div class="meta">ID: ${report.id}<br>Generated: ${formatTimestamp(report.timestamp)}</div>
        <div class="score">Score: ${report.score}/100</div>
        <pre>${report.data_json ? JSON.stringify(JSON.parse(report.data_json), null, 2) : 'No data available'}</pre>
        <script>window.onload = () => setTimeout(() => window.print(), 300)</script>
        </body></html>`)
      win.document.close()
      toast.success('PDF export opened — use print dialog to save')
    } catch {
      toast.error('PDF export failed')
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

  // EXP-01: CSV export of all filtered reports (one row per report).
  const handleExportAllCSV = useCallback(() => {
    try {
      const rows = filtered.flatMap((r) => reportToCSVRows(r))
      downloadText(
        `all-reports-${new Date().toISOString().slice(0, 10)}.csv`,
        toCSV(rows),
        'text/csv'
      )
      toast.success('All reports exported as CSV')
    } catch {
      toast.error('CSV export failed')
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
            onClick={handleExportAllCSV}
            disabled={filtered.length === 0}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] text-xs font-bold text-[var(--color-text)] hover:bg-[var(--color-panel-hover)] transition-all disabled:opacity-40"
          >
            <FileSpreadsheet size={14} />
            Export CSV
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
                        title="Export report (JSON)"
                      >
                        <Download size={12} />
                      </button>

                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleExportCSV(report)
                        }}
                        className="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity text-[var(--color-text-faint)] hover:text-emerald-500 hover:bg-emerald-500/10"
                        title="Export report (CSV)"
                      >
                        <FileSpreadsheet size={12} />
                      </button>

                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleExportPDF(report)
                        }}
                        className="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity text-[var(--color-text-faint)] hover:text-red-500 hover:bg-red-500/10"
                        title="Export report (PDF)"
                      >
                        <FileText size={12} />
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

                {/* Report data — human readable */}
                <div className="space-y-3">
                  <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--color-text-faint)]">
                    {getTypeMeta(selectedReport.type).label} Report Data
                  </p>
                  <div className="p-4 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] max-h-[400px] overflow-y-auto">
                    <ReportHumanReadable
                      type={selectedReport.type}
                      rawJson={fullReport?.data_json ?? selectedReport?.data_json ?? ''}
                    />
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
