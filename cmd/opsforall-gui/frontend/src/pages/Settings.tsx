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
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Slider from '@radix-ui/react-slider'
import * as Dialog from '@radix-ui/react-dialog'
import { useBackend } from '@/hooks/useBackend'
import { useThemeStore, useSettingsStore } from '@/stores/useSettingsStore'
import { toast } from 'sonner'
import { useState } from 'react'
import type { AlertRuleInfo, CollectorStatus } from '@/types'

// ── Section Card (aligned with Squib design system) ──

function Section({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-6 shadow-lg">
      <div className="flex items-center gap-3 mb-4 pb-3 border-b border-[var(--color-border)]/50">
        <div className="w-8 h-8 rounded-lg bg-[var(--color-accent-soft)] flex items-center justify-center text-[var(--color-accent)]">
          {icon}
        </div>
        <h2 className="text-xs font-semibold text-[var(--color-text)] uppercase tracking-wider">{title}</h2>
      </div>
      <div className="space-y-4">{children}</div>
    </div>
  )
}

// ── Setting Row ──

function SettingRow({ label, description, children }: { label: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="min-w-0">
        <p className="text-sm font-medium text-[var(--color-text)]">{label}</p>
        {description && <p className="text-xs text-[var(--color-text-faint)] mt-0.5">{description}</p>}
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  )
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
      {collectors.map((c) => (
        <div
          key={c.id}
          className="flex items-center justify-between gap-3 p-3 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]/50"
        >
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="text-sm font-semibold text-[var(--color-text)]">{c.name}</p>
              <span
                className={cn(
                  'inline-block w-2 h-2 rounded-full',
                  c.enabled ? 'bg-green-500' : 'bg-gray-500',
                )}
              />
            </div>
            <p className="text-xs text-[var(--color-text-faint)] truncate">{c.description}</p>
            {c.last_run && (
              <p className="text-[10px] text-[var(--color-text-faint)] mt-0.5 font-mono">
                Last: {new Date(c.last_run).toLocaleTimeString()}
              </p>
            )}
          </div>

          <div className="flex items-center gap-2 shrink-0">
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
            >
              {c.enabled ? <Pause size={14} /> : <Play size={14} />}
            </button>
          </div>
        </div>
      ))}
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
  name: 'OpsForAll',
  version: '1.3.0',
  go_version: 'go1.26.5',
  uptime: '--',
}

export function Settings() {
  const { call } = useBackend()
  const queryClient = useQueryClient()

  // Theme — from zustand store
  const { theme, toggle } = useThemeStore()

  // Settings — from zustand store (auto-persisted to localStorage)
  const { refreshInterval, pingCount, dnsTimeout, setRefreshInterval, setPingCount, setDnsTimeout } = useSettingsStore()

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

  const isDark = theme === 'dark'

  return (
    <div className="h-full overflow-y-auto space-y-5 max-w-3xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-3">
          <Monitor size={24} className="text-[var(--color-accent)]" /> Settings
        </h1>
        <p className="text-sm text-[var(--color-text-dim)] mt-1">
          Configure theme, collection intervals, and network parameters.
        </p>
      </div>

      {/* ── Theme ── */}
      <Section title="Theme" icon={<Monitor size={20} />}>
        <SettingRow label="Appearance" description="Switch between dark and light theme">
          <div className="flex items-center rounded-xl overflow-hidden border border-border">
            <button
              onClick={() => { if (!isDark) toggle() }}
              className={cn(
                'flex items-center gap-2 px-4 py-2 text-sm font-bold transition-all',
                isDark
                  ? 'bg-accent text-white shadow-lg'
                  : 'bg-panel-2 text-text-dim hover:text-text',
              )}
            >
              <Moon size={16} /> Dark
            </button>
            <button
              onClick={() => { if (isDark) toggle() }}
              className={cn(
                'flex items-center gap-2 px-4 py-2 text-sm font-bold transition-all',
                !isDark
                  ? 'bg-accent text-white shadow-lg'
                  : 'bg-panel-2 text-text-dim hover:text-text',
              )}
            >
              <Sun size={16} /> Light
            </button>
          </div>
        </SettingRow>
      </Section>

      {/* ── Collection Interval ── */}
      <Section title="Collection" icon={<Activity size={20} />}>
        <SettingRow
          label="Refresh Interval"
          description="How often the dashboard and metrics refresh"
        >
          <select
            value={refreshInterval}
            onChange={(e) => {
              const val = Number(e.target.value)
              setRefreshInterval(val)
              call('PipelineAPI.UpdateSettings', val, 0, pingCount, dnsTimeout)
              toast.success(`Refresh interval set to ${val / 1000}s`)
            }}
            className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm font-medium text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] transition-colors"
          >
            {intervalOptions.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </SettingRow>
      </Section>


      {/* ── Collectors ── */}
      <Section title="Collectors" icon={<Activity size={20} />}>
        <p className="text-sm text-[var(--color-text-faint)] mb-3">
          Enable or disable individual data collectors, adjust intervals, or trigger a manual collection.
        </p>
        <CollectorList call={call} />
      </Section>

      {/* ── Alert Rules ── */}
      <Section title="Alert Rules" icon={<Bell size={20} />}>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-text-faint">Configured system-wide alert thresholds.</p>
            <Dialog.Root open={addOpen} onOpenChange={setAddOpen}>
              <Dialog.Trigger asChild>
                <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-accent text-white text-xs font-bold hover:opacity-90 transition-all shadow-md active:scale-95">
                  <Plus size={14} /> Add Rule
                </button>
              </Dialog.Trigger>
              <Dialog.Portal>
                <Dialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" />
                <Dialog.Content className="fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[20px] p-8 w-full max-w-md shadow-2xl">
                  <div className="flex items-center justify-between mb-6">
                    <Dialog.Title className="text-xl font-bold text-text">New Alert Rule</Dialog.Title>
                    <Dialog.Close className="text-text-faint hover:text-text transition-colors"><XCircle size={20} /></Dialog.Close>
                  </div>
                  <div className="space-y-4">
                    <div>
                      <label className="text-xs font-bold text-text-dim uppercase tracking-widest block mb-1.5">Metric</label>
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
                        <label className="text-xs font-bold text-text-dim uppercase tracking-widest block mb-1.5">Condition</label>
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
                        <label className="text-xs font-bold text-text-dim uppercase tracking-widest block mb-1.5">Threshold</label>
                        <input
                          type="number"
                          value={newRule.threshold}
                          onChange={(e) => setNewRule({ ...newRule, threshold: Number(e.target.value) })}
                          className="w-full bg-[var(--color-panel-2)] border border-border rounded-lg px-3 py-2 text-sm text-text focus:outline-none focus:border-accent"
                        />
                      </div>
                    </div>
                    <div>
                      <label className="text-xs font-bold text-text-dim uppercase tracking-widest block mb-1.5">Severity</label>
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

          <div className="bg-panel-2 border border-border rounded-xl overflow-hidden shadow-inner">
            <table className="w-full text-left border-collapse">
              <thead className="bg-panel-3 border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-[10px] font-black uppercase text-text-faint">Metric</th>
                  <th className="px-4 py-3 text-[10px] font-black uppercase text-text-faint">Condition</th>
                  <th className="px-4 py-3 text-[10px] font-black uppercase text-text-faint">Severity</th>
                  <th className="px-4 py-3 text-right"></th>
                </tr>
              </thead>
              <tbody>
                {rules.map((rule, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-panel transition-colors group">
                    <td className="px-4 py-3">
                      <p className="text-xs font-bold text-text">{rule.metric.replace('.percent', '').replace('_', ' ').toUpperCase()}</p>
                    </td>
                    <td className="px-4 py-3">
                      <span className="text-xs font-mono font-bold text-text-dim">{rule.condition} {rule.threshold}</span>
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
                        className="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-danger/10 hover:text-danger text-text-faint transition-all"
                      >
                        <Trash2 size={14} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </Section>

      {/* ── Network ── */}
      <Section title="Network" icon={<Network size={20} />}>
        <SettingRow
          label="Default Ping Count"
          description="Number of echo requests per ping (1–20)"
        >
          <div className="flex items-center gap-3 w-44">
            <Slider.Root
              value={[pingCount]}
              onValueChange={([v]: number[]) => { setPingCount(v); call('PipelineAPI.UpdateSettings', 0, 0, v, dnsTimeout); toast.success(`Ping count set to ${v}`) }}
              min={1}
              max={20}
              step={1}
              className="relative flex items-center flex-1 h-5 cursor-pointer"
            >
              <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-panel-2 border border-border">
                <Slider.Range className="absolute h-full rounded-full bg-accent" />
              </Slider.Track>
              <Slider.Thumb className="block w-4 h-4 bg-text rounded-full shadow-lg focus:outline-none focus:ring-2 focus:ring-accent" />
            </Slider.Root>
            <span className="text-xs font-semibold font-[Geist_Mono] text-[var(--color-text)] w-6 text-right tabular-nums">{pingCount}</span>
          </div>
        </SettingRow>

        <SettingRow
          label="DNS Timeout"
          description="Timeout in milliseconds for DNS lookups"
        >
          <div className="flex items-center gap-3 w-44">
            <Slider.Root
              value={[dnsTimeout]}
              onValueChange={([v]: number[]) => { setDnsTimeout(v); call('PipelineAPI.UpdateSettings', 0, 0, pingCount, v); toast.success(`DNS timeout set to ${v}ms`) }}
              min={500}
              max={10000}
              step={100}
              className="relative flex items-center flex-1 h-5 cursor-pointer"
            >
              <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-panel-2 border border-border">
                <Slider.Range className="absolute h-full rounded-full bg-accent" />
              </Slider.Track>
              <Slider.Thumb className="block w-4 h-4 bg-text rounded-full shadow-lg focus:outline-none focus:ring-2 focus:ring-accent" />
            </Slider.Root>
            <span className="text-xs font-semibold font-[Geist_Mono] text-[var(--color-text)] w-14 text-right tabular-nums">{dnsTimeout}ms</span>
          </div>
        </SettingRow>
      </Section>

      {/* ── Management ── */}
      <Section title="Management" icon={<Monitor size={20} />}>
        <div className="space-y-4">
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between p-4 rounded-xl bg-panel-2 border border-border">
              <div>
                <p className="text-sm font-bold text-text">Factory Reset</p>
                <p className="text-xs text-text-faint">Clear all settings and restore defaults.</p>
              </div>
              <button
                onClick={() => {
                  setRefreshInterval(DEFAULT_SETTINGS.refreshInterval)
                  setPingCount(DEFAULT_SETTINGS.pingCount)
                  setDnsTimeout(DEFAULT_SETTINGS.dnsTimeout)
                  localStorage.setItem('opsforall_refreshInterval', JSON.stringify(DEFAULT_SETTINGS.refreshInterval))
                  localStorage.setItem('opsforall_pingCount', JSON.stringify(DEFAULT_SETTINGS.pingCount))
                  localStorage.setItem('opsforall_dnsTimeout', JSON.stringify(DEFAULT_SETTINGS.dnsTimeout))
                  call('PipelineAPI.UpdateSettings', DEFAULT_SETTINGS.refreshInterval, 0, DEFAULT_SETTINGS.pingCount, DEFAULT_SETTINGS.dnsTimeout)
                  toast.success('All settings reset to defaults')
                }}
                className="px-4 py-2 bg-danger/10 hover:bg-danger/20 text-danger text-xs font-bold rounded-lg border border-danger/30 transition-all"
              >
                <RotateCcw size={14} className="inline mr-2" />
                Reset Defaults
              </button>
            </div>

            <div className="flex items-center justify-between p-4 rounded-xl bg-panel-2 border border-border">
              <div>
                <p className="text-sm font-bold text-text">Onboarding Wizard</p>
                <p className="text-xs text-text-faint">Re-run the first-time setup experience.</p>
              </div>
              <button
                onClick={async () => {
                  if (confirm('Are you sure you want to reset onboarding? The app will reload.')) {
                    // We don't have a direct 'ClearOnboarded' binding yet, but we can add one
                    // or just use a generic command if DevOps allowed it.
                    // Better to use the App facade.
                    try {
                      // We'll assume the user has access to App.ClearOnboarded once we add it
                      await call('App.ClearOnboarded')
                      window.location.reload()
                    } catch {
                      toast.error('Failed to reset onboarding')
                    }
                  }
                }}
                className="px-4 py-2 bg-accent/10 hover:bg-accent/20 text-accent text-xs font-bold rounded-lg border border-accent/30 transition-all"
              >
                <RefreshCw size={14} className="inline mr-2" />
                Reset Onboarding
              </button>
            </div>
          </div>
        </div>
      </Section>

      {/* ── About ── */}
      <Section title="About" icon={<Info size={20} />}>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Application</p>
            <p className="text-[var(--color-text)] font-medium">{appInfo.name}</p>
          </div>
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Version</p>
            <p className="text-[var(--color-text)] font-medium font-[Geist_Mono]">{appInfo.version}</p>
          </div>
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Go Version</p>
            <p className="text-[var(--color-text)] font-medium font-[Geist_Mono]">{appInfo.go_version}</p>
          </div>
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Uptime</p>
            <p className="text-[var(--color-text)] font-medium font-[Geist_Mono]">{appInfo.uptime}</p>
          </div>
        </div>
        <div className="pt-6 border-t border-border/50 mt-6">
          <div className="flex items-center gap-6">
            <a href="https://github.com/shahriarhaqueabir/AllOpsFull" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-sm font-bold text-accent hover:text-accent-2 transition-colors">
              <ExternalLink size={16} /> GitHub
            </a>
            <a href="https://github.com/shahriarhaqueabir/AllOpsFull#readme" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-sm font-bold text-accent hover:text-accent-2 transition-colors">
              <ExternalLink size={16} /> Documentation
            </a>
          </div>
        </div>
      </Section>
    </div>
  )
}
