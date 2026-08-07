import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Monitor,
  Activity,
  Network,
  Info,
  Moon,
  Sun,
  ExternalLink,
  RotateCcw,
  Bell,
  Trash2,
  Plus,
  XCircle,
  Play,
  Pause,
  RefreshCw,
  BrainCircuit,
  ScrollText,
  Zap,
  Database,
  HardDrive,
} from 'lucide-react'
import { cn, safeDate } from '@/lib/utils'
import * as Slider from '@radix-ui/react-slider'
import * as Dialog from '@radix-ui/react-dialog'
import { useBackend } from '@/hooks/useBackend'
import { useThemeStore, useSettingsStore } from '@/stores'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { toast } from 'sonner'
import { useState } from 'react'
import type { AlertRuleInfo, CollectorStatus } from '@/types'
import { Panel, PanelHeader } from '@/components/ui/Panel'
import { SettingsSidebar, type SettingsTab } from '@/components/settings/SettingsSidebar'
import { useConfigStore } from '@/stores/useConfigStore'
import { DeploymentBar } from '@/components/settings/DeploymentBar'
import { CapabilityMatrix } from '@/components/settings/CapabilityMatrix'
import { ModelfileEditor } from '@/components/settings/ModelfileEditor'
import { ModelManager } from '@/components/settings/ModelManager'
import { ContextPreview } from '@/components/settings/ContextPreview'

// ── Setting Row ──

function SettingRow({ label, description, children }: { label: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4 py-1">
      <div className="min-w-0">
        <p className="text-sm font-bold text-[var(--color-text)]">{label}</p>
        {description && <p className="text-xs text-[var(--color-text-dim)] mt-0.5">{description}</p>}
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  )
}

function formatLastRun(date: Date): string {
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMinutes = Math.max(1, Math.round(Math.abs(diffMs) / 60000))

  const relative =
    diffMinutes < 60
      ? `${diffMinutes}m ago`
      : diffMinutes < 24 * 60
        ? `${Math.round(diffMinutes / 60)}h ago`
        : `${Math.round(diffMinutes / (60 * 24))}d ago`

  const absolute = date.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })

  return `${absolute} - ${relative}`
}

// ── Collector Manager ──

function CollectorList({ call }: { call: ReturnType<typeof useBackend>['call'] }) {
  const { refreshInterval } = useSettingsStore()
  const { data: collectors, refetch } = useQuery<CollectorStatus[]>({
    queryKey: ['collectors'],
    queryFn: async () => {
      const res = await call('App.ListCollectors')
      return (res as CollectorStatus[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const toggleEnabled = async (id: string, current: boolean) => {
    await call('App.SetCollectorEnabled', id, !current)
    refetch()
    toast.success(current ? `Paused ${id} collection` : `Resumed ${id} collection`)
  }

  const setInterval = async (id: string, ms: number) => {
    await call('App.SetCollectorInterval', id, ms)
    refetch()
    toast.info(`${id} interval set to ${ms / 1000}s`)
  }

  const trigger = async (id: string) => {
    try {
      await call('App.TriggerCollector', id)
      toast.success(`${id} collection triggered`)
      refetch()
    } catch {
      toast.error(`Failed to trigger ${id}`)
    }
  }

  if (!collectors || collectors.length === 0) {
    return <p className="text-sm text-[var(--color-text-faint)]">No collectors registered.</p>
  }

  return (
    <div className="space-y-2">
      {collectors.map((c) => {
        const lastRun = c.last_run ? safeDate(c.last_run) : null
        const lastRunLabel = lastRun ? formatLastRun(lastRun) : null

        return (
          <div
            key={c.id}
            className="flex flex-col gap-4 p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]/50 md:flex-row md:items-start md:justify-between"
          >
            <div className="min-w-0 flex-1 space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                <p className="text-sm font-semibold text-[var(--color-text)]">{c.name}</p>
                <span className={cn(
                  'inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider',
                  c.enabled ? 'border-green-500/20 bg-green-500/10 text-green-500' : 'border-[var(--color-border)] bg-[var(--color-panel)] text-[var(--color-text-faint)]'
                )}>
                  {c.enabled ? 'Enabled' : 'Paused'}
                </span>
              </div>
              <p className="text-xs text-[var(--color-text-faint)] truncate">{c.description}</p>
              <div className="flex flex-wrap items-center gap-2 text-[10px] text-[var(--color-text-faint)]">
                <span className="rounded-full border border-[var(--color-border)] bg-[var(--color-panel)] px-2 py-0.5 font-mono uppercase tracking-wider">
                  Every {c.interval_ms / 1000}s
                </span>
                {lastRunLabel && (
                  <p className="font-mono">
                    Last run: {lastRunLabel}
                  </p>
                )}
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-2 shrink-0 md:justify-end">
              {c.enabled && (
                <select
                  value={c.interval_ms}
                  onChange={(e) => setInterval(c.id, Number(e.target.value))}
                  className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg px-2 py-1 text-xs font-medium text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)]"
                >
                  {[1000, 3000, 5000, 10000, 15000, 30000].map((ms) => (
                    <option key={ms} value={ms}>
                      {ms / 1000}s
                    </option>
                  ))}
                </select>
              )}

              <button
                onClick={() => trigger(c.id)}
                className="p-2 rounded-lg hover:bg-[var(--color-accent-soft)] text-[var(--color-text-dim)] hover:text-[var(--color-accent)] transition-all"
                title="Collect now"
                aria-label={`Collect ${c.name} now`}
              >
                <RefreshCw size={14} />
              </button>

              <button
                onClick={() => toggleEnabled(c.id, c.enabled)}
                className={cn(
                  'p-2 rounded-lg transition-all',
                  c.enabled
                    ? 'hover:bg-[var(--color-danger)]/10 text-[var(--color-danger)]'
                    : 'hover:bg-green-500/10 text-green-500',
                )}
                title={c.enabled ? 'Pause' : 'Resume'}
                aria-label={c.enabled ? `Pause ${c.name}` : `Resume ${c.name}`}
              >
                {c.enabled ? <Pause size={14} /> : <Play size={14} />}
              </button>
            </div>
          </div>
        )
      })}
    </div>
  )
}

// ══════════════════════════════════
//  Settings Page
// ══════════════════════════════════

const intervalOptions = [
  { value: 1000, label: '1s' },
  { value: 3000, label: '3s' },
  { value: 5000, label: '5s' },
  { value: 10000, label: '10s' },
  { value: 30000, label: '30s' },
]

const DEFAULT_SETTINGS = {
  refreshInterval: 5000,
  pingCount: 4,
  dnsTimeout: 2000,
}

interface AppInfo {
  name: string
  version: string
  go_version: string
  uptime: string
}

const DEFAULT_APP_INFO: AppInfo = {
  name: 'UniversalOps',
  version: '1.6.2',
  go_version: 'go1.26.5',
  uptime: '--',
}

export function Settings() {
  const { call } = useBackend()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<SettingsTab>('general')

  // Theme — from zustand store
  const { theme, toggle } = useThemeStore()

  // Settings — from zustand store (auto-persisted to localStorage)
  const {
    refreshInterval,
    pingCount,
    dnsTimeout,
    companionName,
    autoEcoMode,
    setRefreshInterval,
    setPingCount,
    setDnsTimeout,
    setCompanionName,
    setAutoEcoMode,
  } = useSettingsStore()

  const { stagedChanges, stageChange, stageBatch } = useConfigStore()

  // Helper to determine active value (staged or original)
  const getVal = (key: string, original: any) => {
    return stagedChanges.has(key) ? stagedChanges.get(key) : original
  }

  const [optimizing, setOptimizing] = useState(false)
  const [resetConfirm, setResetConfirm] = useState(false)
  const [onboardingConfirm, setOnboardingConfirm] = useState(false)
  const handleHawkOptimize = async () => {
    if (optimizing) return
    setOptimizing(true)
    const tid = toast.loading('Hawk is analyzing system load and pipeline efficiency...')
    try {
      const response = await call('AIOps.RequestOptimization') as { payload?: Record<string, any>, content: string }
      if (response.payload) {
        stageBatch(response.payload)
        toast.success('Hawk has staged optimization proposals', { id: tid, description: response.content })
      } else {
        toast.error('Hawk could not determine optimal settings', { id: tid })
      }
    } catch (err: any) {
      toast.error(`Optimization failed: ${err.message}`, { id: tid })
    } finally {
      setOptimizing(false)
    }
  }

  // Rules — via react-query
  const { data: rules = [] } = useQuery<AlertRuleInfo[]>({
    queryKey: ['alert-rules'],
    queryFn: async () => {
      const res = await call('AlertAPI.GetRules')
      return (res as AlertRuleInfo[]) || []
    },
  })

  const [addOpen, setAddOpen] = useState(false)
  const [newRule, setNewRule] = useState({ metric: 'cpu.percent', threshold: 90, severity: 'critical', condition: 'gt' })

  const handleAddRule = async () => {
    await call('AlertAPI.AddRule', newRule.metric, newRule.threshold, newRule.severity, newRule.condition)
    queryClient.invalidateQueries({ queryKey: ['alert-rules'] })
    setAddOpen(false)
    toast.success('Alert rule added')
  }

  const handleRemoveRule = async (metric: string, threshold: number) => {
    await call('AlertAPI.RemoveRule', metric, threshold)
    queryClient.invalidateQueries({ queryKey: ['alert-rules'] })
    toast.info('Alert rule removed')
  }

  // About — via react-query
  const { data: appInfo = DEFAULT_APP_INFO } = useQuery<AppInfo>({
    queryKey: ['app-info'],
    queryFn: async () => {
      try {
        const result = await call('App.GetAppInfo') as Partial<AppInfo>
        if (result) {
          return {
            name: result.name || DEFAULT_APP_INFO.name,
            version: result.version || DEFAULT_APP_INFO.version,
            go_version: result.go_version || DEFAULT_APP_INFO.go_version,
            uptime: result.uptime || DEFAULT_APP_INFO.uptime,
          }
        }
      } catch { /* ignore */
        console.warn('[Settings] Failed to fetch AppInfo, using defaults')
      }
      return DEFAULT_APP_INFO
    },
    staleTime: 60000,
  })

  const { data: dataDir = 'data' } = useQuery({
    queryKey: ['data-dir'],
    queryFn: async () => (await call('App.GetDataDir')) as string,
  })

  const { data: logsDir = 'logs' } = useQuery({
    queryKey: ['logs-dir'],
    queryFn: async () => (await call('App.GetLogsDir')) as string,
  })

  const handleRelocateData = async () => {
    try {
      const path = await call('App.SelectFolderDialog', 'Select New Data Folder')
      if (path) {
        const id = toast.loading(`Relocating telemetry to ${path}...`)
        await call('App.UpdateStorageConfig', path)
        queryClient.invalidateQueries({ queryKey: ['data-dir'] })
        toast.success('Telemetry relocated successfully', { id })
      }
    } catch (err: any) {
      toast.error(err?.message || 'Relocation failed')
    }
  }

  const handleRelocateLogs = async () => {
    try {
      const path = await call('App.SelectFolderDialog', 'Select New Logs Folder')
      if (path) {
        const id = toast.loading(`Relocating logs to ${path}...`)
        await call('App.UpdateLogsConfig', path)
        queryClient.invalidateQueries({ queryKey: ['logs-dir'] })
        toast.success('Logs relocated successfully', { id })
      }
    } catch (err: any) {
      toast.error(err?.message || 'Relocation failed')
    }
  }

  const handlePickModelfile = async () => {
    try {
      const path = await call('App.OpenFileDialog', 'Select Persona Modelfile', ['Modelfile|*.modelfile', 'All Files|*.*'])
      if (path) {
        // Implementation for changing active modelfile path can go here
        toast.info(`Selected modelfile: ${path}`)
      }
    } catch (err: any) {
      toast.error(err?.message || 'Selection failed')
    }
  }

  const isDark = theme === 'dark'

  return (
    <div className="h-full flex overflow-hidden bg-[var(--color-bg)]/50 relative">
      <SettingsSidebar activeTab={activeTab} onTabChange={setActiveTab} />

      <main className="flex-1 overflow-y-auto p-8 pb-32 space-y-8 scroll-smooth">
        <div className="mb-2 max-w-4xl">
          <h1 className="text-2xl font-black text-[var(--color-text)] uppercase tracking-tight">
            Control Plane
          </h1>
          <p className="text-sm text-[var(--color-text-dim)] mt-1">
            Configure your local operations control center.
          </p>
        </div>

        {/* ── General ── */}
        {activeTab === 'general' && (
          <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl">
            <Panel category="none">
              <PanelHeader
                icon={<Monitor size={20} />}
                title="Appearance"
                subtitle="Configure the visual look and feel"
              />
              <div className="mt-6">
                <SettingRow label="Theme" description="Switch between dark and light appearance">
                  <div className="flex items-center rounded-xl overflow-hidden border border-[var(--color-border)] bg-[var(--color-panel-2)] p-1">
                    <button
                      onClick={() => { if (!isDark) toggle() }}
                      className={cn(
                        'flex items-center gap-2 px-4 py-1.5 text-xs font-bold transition-all rounded-lg',
                        isDark ? 'bg-[var(--color-accent)] text-white shadow-lg' : 'text-[var(--color-text-dim)] hover:text-[var(--color-text)]',
                      )}
                    >
                      <Moon size={14} /> Dark
                    </button>
                    <button
                      onClick={() => { if (isDark) toggle() }}
                      className={cn(
                        'flex items-center gap-2 px-4 py-1.5 text-xs font-bold transition-all rounded-lg',
                        !isDark ? 'bg-[var(--color-accent)] text-white shadow-lg' : 'text-[var(--color-text-dim)] hover:text-[var(--color-text)]',
                      )}
                    >
                      <Sun size={14} /> Light
                    </button>
                  </div>
                </SettingRow>

                <div className="h-px bg-[var(--color-border)]/50 my-4" />

                <SettingRow label="Companion Name" description="Set a custom name for your AI co-pilot">
                  <input
                    type="text"
                    value={getVal('companionName', companionName)}
                    onChange={(e) => stageChange('companionName', e.target.value)}
                    className={cn(
                      'bg-[var(--color-panel-2)] border rounded-lg px-3 py-1.5 text-sm font-bold text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] transition-colors w-32',
                      stagedChanges.has('companionName') ? 'border-[var(--color-accent)]' : 'border-[var(--color-border)]'
                    )}
                  />
                </SettingRow>
              </div>
            </Panel>

            <Panel category="none" variant="flat">
              <PanelHeader icon={<Info size={20} />} title="About" />
              <div className="mt-6 grid grid-cols-2 gap-6 text-sm">
                <div>
                  <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-1">Application</p>
                  <p className="text-[var(--color-text)] font-bold">{appInfo.name}</p>
                </div>
                <div>
                  <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-1">Version</p>
                  <p className="text-[var(--color-text)] font-mono font-bold">{appInfo.version}</p>
                </div>
                <div>
                  <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-1">Go Engine</p>
                  <p className="text-[var(--color-text)] font-mono font-bold">{appInfo.go_version}</p>
                </div>
                <div>
                  <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-1">Uptime</p>
                  <p className="text-[var(--color-text)] font-mono font-bold">{appInfo.uptime}</p>
                </div>
              </div>
              <div className="pt-6 border-t border-[var(--color-border)]/50 mt-6 flex gap-6">
                <a href="https://github.com/shahriarhaqueabir/UniversalOps" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-xs font-bold text-[var(--color-accent)] hover:opacity-80 transition-colors">
                  <ExternalLink size={14} /> SOURCE CODE
                </a>
                <a href="https://github.com/shahriarhaqueabir/UniversalOps#readme" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-xs font-bold text-[var(--color-accent)] hover:opacity-80 transition-colors">
                  <ExternalLink size={14} /> DOCUMENTATION
                </a>
              </div>
            </Panel>
          </div>
        )}

        {/* ── Intelligence ── */}
        {activeTab === 'intelligence' && (
          <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl">
            <Panel category="ai">
              <PanelHeader
                icon={<BrainCircuit size={20} />}
                title={`${companionName} Neural Core`}
                subtitle="Configure specialized intelligence and persona instructions"
              />
              <div className="mt-6">
                <ModelfileEditor />
              </div>
            </Panel>

            <Panel category="ai">
              <PanelHeader
                icon={<HardDrive size={20} />}
                title="AI Identity Location"
                subtitle="Manage where your companion's persona is stored"
              />
              <div className="mt-6">
                <div className="flex items-center justify-between p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-bold text-[var(--color-text)]">Active Modelfile</p>
                    <p className="text-xs text-accent mt-1 font-mono truncate">{dataDir}/universalops.modelfile</p>
                  </div>
                  <button
                    onClick={handlePickModelfile}
                    className="ml-4 px-3 py-1.5 bg-[var(--color-accent)]/10 hover:bg-[var(--color-accent)]/20 text-[var(--color-accent)] text-[10px] font-black uppercase rounded-lg border border-[var(--color-accent)]/30 transition-all"
                  >
                    CHOOSE FILE
                  </button>
                </div>
              </div>
            </Panel>

            <Panel category="ai">
              <PanelHeader
                icon={<Database size={20} />}
                title="Model Inventory"
                subtitle="Manage local AI artifacts and pull new intelligence"
              />
              <div className="mt-6">
                <ModelManager />
              </div>
            </Panel>
          </div>
        )}

        {/* ── Engine ── */}
        {activeTab === 'engine' && (
          <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl">
            <Panel category="system" variant="elevated" className="border-accent/20">
              <div className="flex items-center justify-between p-6">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-2xl bg-accent/10 border border-accent/20 flex items-center justify-center text-accent shadow-sm">
                    <BrainCircuit size={24} />
                  </div>
                  <div>
                    <h3 className="text-sm font-black text-text uppercase tracking-widest">Neural Optimization</h3>
                    <p className="text-xs text-text-dim mt-0.5">Let Hawk analyze your workload and propose ideal engine settings.</p>
                  </div>
                </div>
                <button
                  onClick={handleHawkOptimize}
                  disabled={optimizing}
                  className="px-6 py-3 bg-accent text-white rounded-xl font-black text-xs uppercase tracking-widest hover:scale-105 transition-all active:scale-95 shadow-lg flex items-center gap-3 disabled:opacity-50"
                >
                  {optimizing ? <RefreshCw size={14} className="animate-spin" /> : <Zap size={14} />}
                  Ask Hawk to Optimize
                </button>
              </div>
            </Panel>

            <Panel category="system">
              <PanelHeader
                icon={<Activity size={20} />}
                title="Data Pipeline"
                subtitle="Configure collection and refresh frequencies"
              />
              <div className="mt-6 space-y-6">
                <SettingRow label="Global Refresh" description="Frequency for dashboard and metric updates">
                  <select
                    value={getVal('refreshInterval', refreshInterval)}
                    onChange={(e) => stageChange('refreshInterval', Number(e.target.value))}
                    className={cn(
                      "bg-[var(--color-panel-2)] border rounded-lg px-3 py-2 text-sm font-bold text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] transition-colors",
                      stagedChanges.has('refreshInterval') ? 'border-[var(--color-accent)]' : 'border-[var(--color-border)]'
                    )}
                  >
                    {intervalOptions.map((o) => (
                      <option key={o.value} value={o.value}>{o.label}</option>
                    ))}
                  </select>
                </SettingRow>

                <div className="h-px bg-[var(--color-border)]/50 my-4" />

                <SettingRow label="Auto Eco-Mode" description="Throttles telemetry to preserve battery when discharging">
                  <button
                    onClick={() => stageChange('autoEcoMode', !getVal('autoEcoMode', autoEcoMode))}
                    className={cn(
                      "relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none",
                      getVal('autoEcoMode', autoEcoMode) ? "bg-accent" : "bg-[var(--color-panel-3)] border border-[var(--color-border)]",
                      stagedChanges.has('autoEcoMode') && "ring-2 ring-accent ring-offset-2 ring-offset-[var(--color-bg)]"
                    )}
                  >
                    <span className={cn(
                      "inline-block h-4 w-4 transform rounded-full bg-white transition-transform",
                      getVal('autoEcoMode', autoEcoMode) ? "translate-x-6" : "translate-x-1"
                    )} />
                  </button>
                </SettingRow>

                <div className="h-px bg-[var(--color-border)]/50 my-4" />

                <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-3">Active Collectors</p>
                <CollectorList call={call} />
              </div>
            </Panel>

            <Panel category="system">
              <PanelHeader
                icon={<Monitor size={20} />}
                title="System Capabilities"
                subtitle="Discovered workstation powers and tool integrations"
              />
              <div className="mt-6">
                <CapabilityMatrix />
              </div>
            </Panel>

            <Panel category="system">
              <PanelHeader
                icon={<Zap size={20} />}
                title="Core Maintenance"
                subtitle="Optimize and verify the underlying data structures"
              />
              <div className="mt-6 flex flex-col gap-3">
                <div className="flex items-center justify-between p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                  <div>
                    <p className="text-sm font-bold text-[var(--color-text)]">SQLite Optimization</p>
                    <p className="text-xs text-[var(--color-text-dim)]">Rebuild database to reclaim space and improve performance.</p>
                  </div>
                  <button
                    onClick={async () => {
                      const id = toast.loading('Optimizing database...')
                      try {
                        await call('App.VacuumDatabase')
                        toast.success('Database optimized', { id })
                      } catch { toast.error('Optimization failed', { id }) }
                    }}
                    className="px-4 py-2 bg-[var(--color-accent)]/10 hover:bg-[var(--color-accent)]/20 text-[var(--color-accent)] text-[10px] font-black uppercase rounded-lg border border-[var(--color-accent)]/30 transition-all"
                  >
                    OPTIMIZE
                  </button>
                </div>

                <div className="flex items-center justify-between p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                  <div>
                    <p className="text-sm font-bold text-[var(--color-text)]">Integrity Check</p>
                    <p className="text-xs text-[var(--color-text-dim)]">Verify table structures and update query statistics.</p>
                  </div>
                  <button
                    onClick={async () => {
                      const id = toast.loading('Running integrity check...')
                      try {
                        await call('App.AnalyzeDatabase')
                        toast.success('Integrity verified', { id })
                      } catch { toast.error('Check failed', { id }) }
                    }}
                    className="px-4 py-2 bg-[var(--color-accent)]/10 hover:bg-[var(--color-accent)]/20 text-[var(--color-accent)] text-[10px] font-black uppercase rounded-lg border border-[var(--color-accent)]/30 transition-all"
                  >
                    ANALYZE
                  </button>
                </div>
              </div>
            </Panel>

            <Panel category="none">
              <PanelHeader icon={<Network size={20} />} title="Network Diagnostics" />
              <div className="mt-6 space-y-6">
                <SettingRow label="Default Ping Count" description="Number of echo requests per ping (1–20)">
                  <div className="flex items-center gap-4 w-48">
                    <Slider.Root
                      value={[getVal('pingCount', pingCount)]}
                      onValueChange={([v]: number[]) => stageChange('pingCount', v)}
                      min={1} max={20} step={1}
                      className="relative flex items-center flex-1 h-5 cursor-pointer"
                    >
                      <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                        <Slider.Range className="absolute h-full rounded-full bg-[var(--color-accent)]" />
                      </Slider.Track>
                      <Slider.Thumb className="block w-4 h-4 bg-[var(--color-text)] rounded-full shadow-lg" />
                    </Slider.Root>
                    <span className="text-xs font-bold font-mono text-[var(--color-text)] w-6 text-right">{getVal('pingCount', pingCount)}</span>
                  </div>
                </SettingRow>

                <SettingRow label="DNS Timeout" description="Timeout in milliseconds for lookups">
                  <div className="flex items-center gap-4 w-48">
                    <Slider.Root
                      value={[getVal('dnsTimeout', dnsTimeout)]}
                      onValueChange={([v]: number[]) => stageChange('dnsTimeout', v)}
                      min={500} max={10000} step={100}
                      className="relative flex items-center flex-1 h-5 cursor-pointer"
                    >
                      <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                        <Slider.Range className="absolute h-full rounded-full bg-[var(--color-accent)]" />
                      </Slider.Track>
                      <Slider.Thumb className="block w-4 h-4 bg-[var(--color-text)] rounded-full shadow-lg" />
                    </Slider.Root>
                    <span className="text-xs font-bold font-mono text-[var(--color-text)] w-14 text-right">{getVal('dnsTimeout', dnsTimeout)}ms</span>
                  </div>
                </SettingRow>
              </div>
            </Panel>
          </div>
        )}

        {/* ── Security ── */}
        {activeTab === 'security' && (
          <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl">
            <Panel category="security">
              <PanelHeader
                icon={<Bell size={20} />}
                title="Alert Thresholds"
                subtitle="Configured system-wide triggers"
                action={
                  <button onClick={() => setAddOpen(true)} className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[var(--color-accent)] text-white text-[10px] font-black uppercase hover:opacity-90 transition-all shadow-md active:scale-95">
                    <Plus size={12} /> Add Rule
                  </button>
                }
              />
              <div className="mt-6">
                <table className="w-full text-left border-collapse">
                  <thead className="bg-[var(--color-panel-3)]/50">
                    <tr>
                      <th className="px-4 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Metric</th>
                      <th className="px-4 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Condition</th>
                      <th className="px-4 py-3 text-[10px] font-black uppercase text-[var(--color-text-faint)] tracking-widest">Severity</th>
                      <th className="px-4 py-3"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {rules.map((rule) => (
                      <tr key={`${rule.metric}-${rule.threshold}`} className="border-b border-[var(--color-border)]/20 hover:bg-[var(--color-panel-2)] transition-colors group">
                        <td className="px-4 py-3">
                          <p className="text-xs font-bold text-[var(--color-text)] uppercase">{rule.metric.replace('.percent', '')}</p>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-xs font-mono font-bold text-[var(--color-text-dim)]">{rule.condition} {rule.threshold}</span>
                        </td>
                        <td className="px-4 py-3">
                          <span className={cn(
                            'px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-tighter border',
                            rule.severity === 'CRITICAL' ? 'bg-danger/10 text-danger border-danger/30' :
                              rule.severity === 'WARNING' ? 'bg-warning/10 text-warning border-warning/30' :
                                'bg-accent/10 text-accent border-accent/30'
                          )}>
                            {rule.severity}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          <button
                            onClick={() => handleRemoveRule(rule.metric, rule.threshold)}
                            className="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-danger/10 hover:text-danger text-[var(--color-text-faint)] transition-all"
                          >
                            <Trash2 size={14} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Panel>
          </div>
        )}

        {/* ── Journal ── */}
        {activeTab === 'journal' && (
          <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl">
            <Panel category="none">
              <PanelHeader
                icon={<ScrollText size={20} />}
                title="Verbosity & Logging"
                subtitle="Control backend log output"
              />
              <div className="mt-6 space-y-6">
                <SettingRow label="Backend Log Level" description="Set the granularity of Go system logs">
                  <select
                    className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm font-bold text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] transition-colors"
                    defaultValue="info"
                    onChange={(e) => {
                      call('App.SetLogLevel', e.target.value)
                      toast.info(`Log level set to ${e.target.value}`)
                    }}
                  >
                    <option value="trace">Trace</option>
                    <option value="debug">Debug</option>
                    <option value="info">Info</option>
                    <option value="warn">Warning</option>
                    <option value="error">Error</option>
                  </select>
                </SettingRow>
              </div>
            </Panel>

            <Panel category="none">
              <PanelHeader
                icon={<Database size={20} />}
                title="Sovereignty & Storage"
                subtitle="Manage portable data and log locations"
              />
              <div className="mt-6 space-y-6">
                <div className="flex items-center justify-between p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-bold text-[var(--color-text)]">Telemetry Core</p>
                    <p className="text-xs text-accent mt-1 font-mono truncate">{dataDir}/universalops.db</p>
                  </div>
                  <button
                    onClick={handleRelocateData}
                    className="ml-4 px-3 py-1.5 bg-[var(--color-accent)]/10 hover:bg-[var(--color-accent)]/20 text-[var(--color-accent)] text-[10px] font-black uppercase rounded-lg border border-[var(--color-accent)]/30 transition-all"
                  >
                    RELOCATE
                  </button>
                </div>

                <div className="flex items-center justify-between p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-bold text-[var(--color-text)]">Diagnostic Logs</p>
                    <p className="text-xs text-accent mt-1 font-mono truncate">{logsDir}/universalops.log</p>
                  </div>
                  <button
                    onClick={handleRelocateLogs}
                    className="ml-4 px-3 py-1.5 bg-[var(--color-accent)]/10 hover:bg-[var(--color-accent)]/20 text-[var(--color-accent)] text-[10px] font-black uppercase rounded-lg border border-[var(--color-accent)]/30 transition-all"
                  >
                    RELOCATE
                  </button>
                </div>
              </div>
            </Panel>

            <Panel category="none">
              <PanelHeader icon={<ScrollText size={20} />} title="System Management" />
              <div className="mt-6 space-y-6">
                <div className="flex flex-col gap-4">
                  <div className="flex items-center justify-between p-5 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] shadow-inner">
                    <div>
                      <p className="text-sm font-black text-[var(--color-text)] uppercase tracking-tight">Factory Reset</p>
                      <p className="text-xs text-[var(--color-text-dim)] mt-1">Restore all settings to original defaults.</p>
                    </div>
                    <button
                      onClick={() => setResetConfirm(true)}
                      className="px-4 py-2 bg-danger/10 hover:bg-danger/20 text-danger text-xs font-bold rounded-xl border border-danger/30 transition-all"
                    >
                      <RotateCcw size={14} className="inline mr-2" />
                      RESET ALL
                    </button>
                  </div>

                  <div className="flex items-center justify-between p-5 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] shadow-inner">
                    <div>
                      <p className="text-sm font-black text-[var(--color-text)] uppercase tracking-tight">Onboarding</p>
                      <p className="text-xs text-[var(--color-text-dim)] mt-1">Re-run the initial setup wizard.</p>
                    </div>
                    <button
                      onClick={() => setOnboardingConfirm(true)}
                      className="px-4 py-2 bg-[var(--color-accent)]/10 hover:bg-[var(--color-accent)]/20 text-[var(--color-accent)] text-xs font-bold rounded-xl border border-[var(--color-accent)]/30 transition-all"
                    >
                      <RefreshCw size={14} className="inline mr-2" />
                      RE-RUN WIZARD
                    </button>
                  </div>
                </div>
              </div>
            </Panel>
          </div>
        )}
      </main>

      <ContextPreview />
      <DeploymentBar />

      <ConfirmDialog
        open={resetConfirm}
        title="Factory Reset"
        description="Reset all settings to defaults? This will restore refresh intervals, ping counts, DNS timeouts, and companion name to their original values."
        type="warning"
        confirmText="Reset All"
        onConfirm={() => {
          setRefreshInterval(DEFAULT_SETTINGS.refreshInterval)
          setPingCount(DEFAULT_SETTINGS.pingCount)
          setDnsTimeout(DEFAULT_SETTINGS.dnsTimeout)
          setCompanionName('Hawk')
          setAutoEcoMode(true)
          call('PipelineAPI.UpdateSettings', DEFAULT_SETTINGS.refreshInterval, 0, DEFAULT_SETTINGS.pingCount, DEFAULT_SETTINGS.dnsTimeout)
          toast.success('All settings reset')
        }}
        onClose={() => setResetConfirm(false)}
      />

      <ConfirmDialog
        open={onboardingConfirm}
        title="Re-run Onboarding"
        description="Re-run the initial setup wizard? The application will reload after clearing the onboarding state."
        type="info"
        confirmText="Re-run Wizard"
        onConfirm={async () => {
          try {
            await call('App.ClearOnboarded')
            window.location.reload()
          } catch { toast.error('Failed to reset onboarding') }
        }}
        onClose={() => setOnboardingConfirm(false)}
      />

      {/* Alert Rule Dialog (Legacy) */}
      <Dialog.Root open={addOpen} onOpenChange={setAddOpen}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" />
          <Dialog.Content className="fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[20px] p-6 w-full max-w-md shadow-2xl">
            <div className="flex items-center justify-between mb-5">
              <Dialog.Title className="text-xl font-bold text-text">New Alert Rule</Dialog.Title>
              <Dialog.Close className="text-text-faint hover:text-text transition-colors"><XCircle size={20} /></Dialog.Close>
            </div>
            <div className="space-y-4">
              <div>
                <label className="text-[10px] font-black text-text-dim uppercase tracking-widest block mb-1.5">Metric</label>
                <select
                  value={newRule.metric}
                  onChange={(e) => setNewRule({ ...newRule, metric: e.target.value })}
                  className="w-full bg-[var(--color-panel-2)] border border-border rounded-lg px-3 py-2 text-sm text-text focus:outline-none focus:border-accent"
                >
                  <option value="cpu.percent">CPU Usage (%)</option>
                  <option value="memory.percent">Memory Usage (%)</option>
                  <option value="disk.percent">Disk Usage (%)</option>
                  <option value="cpu.temperature">CPU Temp (°C)</option>
                  <option value="network.rx.rate">Network RX (bps)</option>
                  <option value="process.count">Process Count</option>
                </select>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-[10px] font-black text-text-dim uppercase tracking-widest block mb-1.5">Condition</label>
                  <select
                    value={newRule.condition}
                    onChange={(e) => setNewRule({ ...newRule, condition: e.target.value })}
                    className="w-full bg-[var(--color-panel-2)] border border-border rounded-lg px-3 py-2 text-sm text-text focus:outline-none focus:border-accent"
                  >
                    <option value="gt">Greater than</option>
                    <option value="lt">Less than</option>
                  </select>
                </div>
                <div>
                  <label className="text-[10px] font-black text-text-dim uppercase tracking-widest block mb-1.5">Threshold</label>
                  <input
                    type="number"
                    value={newRule.threshold}
                    onChange={(e) => setNewRule({ ...newRule, threshold: Number(e.target.value) })}
                    className="w-full bg-[var(--color-panel-2)] border border-border rounded-lg px-3 py-2 text-sm text-text focus:outline-none focus:border-accent"
                  />
                </div>
              </div>
              <div>
                <label className="text-[10px] font-black text-text-dim uppercase tracking-widest block mb-1.5">Severity</label>
                <select
                  value={newRule.severity}
                  onChange={(e) => setNewRule({ ...newRule, severity: e.target.value })}
                  className="w-full bg-[var(--color-panel-2)] border border-border rounded-lg px-3 py-2 text-sm text-text focus:outline-none focus:border-accent"
                >
                  <option value="info">Info</option>
                  <option value="warning">Warning</option>
                  <option value="critical">Critical</option>
                </select>
              </div>
              <button
                onClick={handleAddRule}
                className="w-full py-3 bg-accent text-white rounded-xl font-bold hover:opacity-90 transition-all shadow-lg active:scale-95 mt-4"
              >
                Create Rule
              </button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  )
}
