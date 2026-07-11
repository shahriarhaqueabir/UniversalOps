import { useState, useRef, useCallback, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Terminal,
  Server,
  Folder,
  Play,
  Trash2,
  Search,
  ChevronRight,
  Home,
  RefreshCw,
  FileText,
  X,
  PlayCircle,
  StopCircle,
  RotateCcw,
  Zap,
  Activity,
  ShieldCheck,
  Globe,
  Crosshair,
  History,
  FileCheck,
  TerminalSquare,
  Lock,
  GitBranch,
  Box,
  Code,
  Wrench,
  Container,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Package,
  Variable,
  Cpu,
  Lightbulb,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Tabs from '@radix-ui/react-tabs'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { EmptyState } from '@/components/ui/EmptyState'

// Strip ANSI escape sequences from terminal output
function stripAnsi(text: string): string {
  // Matches ANSI escape sequences: ESC[<params>m, ESC[<params>K, ESC[<params>J, etc.
  // eslint-disable-next-line no-control-regex
  return text.replace(/\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g, '')
}
import type { CommandResult, ServiceEntry, FileEntry, ToolInfo, ContainerSummary, GitSummary, LocalServer, EnvironmentInfo, DevOpsSuggestion, DockerStatus, KubernetesStatus, ServiceCategory, ServiceGroupSummary } from '@/types'

type TabId = 'overview' | 'terminal' | 'powershell-pro' | 'services' | 'file-browser' | 'toolbox' | 'containers' | 'git' | 'servers' | 'environment' | 'log-explorer'

// ── Inline helpers ──

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    running: 'bg-success/20 text-success',
    stopped: 'bg-danger/20 text-danger',
    auto: 'bg-accent/20 text-accent',
    manual: 'bg-warning/20 text-warning',
    disabled: 'bg-text-faint/20 text-text-faint',
  }
  return (
    <span className={cn('px-2 py-0.5 rounded text-xs font-medium border border-current opacity-80', colors[status.toLowerCase()] || 'bg-text-faint/20 text-text-faint')}>
      {status}
    </span>
  )
}

// ══════════════════════════════════════════════
//  DevOps Page
// ══════════════════════════════════════════════

export function DevOps() {
  const [activeTab, setActiveTab] = useState<TabId>('terminal')

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      <div className="p-8 border-b border-border bg-panel-2">
        <h1 className="text-3xl font-bold text-text flex items-center gap-3">
          <TerminalSquare size={32} className="text-accent" />
          DevOps Console
        </h1>
        <p className="text-text-dim text-lg mt-2">
          Unified command center for terminal, automated workflows, and system services.
        </p>
      </div>

      <Tabs.Root value={activeTab} onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-border bg-panel px-4">
          {[
            { id: 'overview', label: 'Overview', icon: <Activity size={20} className="text-accent" /> },
            { id: 'terminal', label: 'Interactive Terminal', icon: <Terminal size={20} /> },
            { id: 'powershell-pro', label: 'PowerShell Pro', icon: <Zap size={20} className="text-warning" /> },
            { id: 'services', label: 'System Services', icon: <Server size={20} /> },
            { id: 'containers', label: 'Containers', icon: <Container size={20} /> },
            { id: 'git', label: 'Git', icon: <GitBranch size={20} /> },
            { id: 'servers', label: 'Servers', icon: <Globe size={20} /> },
            { id: 'ai-suggestions', label: 'AI Insights', icon: <Lightbulb size={20} className="text-warning" /> },
            { id: 'environment', label: 'Environment', icon: <Variable size={20} /> },
            { id: 'file-browser', label: 'File Explorer', icon: <Folder size={20} /> },
            { id: 'toolbox', label: 'Toolbox', icon: <Wrench size={20} /> },
            { id: 'log-explorer', label: 'Log Explorer', icon: <FileText size={20} /> },
          ].map((tab) => (
            <Tabs.Trigger
              key={tab.id}
              value={tab.id}
              className={cn(
                'flex items-center gap-3 px-6 py-4 text-base font-bold transition-all border-b-2 border-transparent',
                activeTab === tab.id ? 'border-accent text-text bg-accent/5' : 'text-text-faint hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}
            >
              {tab.icon}
              {tab.label}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="overview" className="h-full">
            <OverviewTab />
          </Tabs.Content>
          <Tabs.Content value="terminal" className="h-full">
            <TerminalTab />
          </Tabs.Content>
          <Tabs.Content value="powershell-pro" className="h-full">
            <PowerShellProTab />
          </Tabs.Content>
          <Tabs.Content value="services" className="h-full">
            <ServicesTab />
          </Tabs.Content>
          <Tabs.Content value="file-browser" className="h-full">
            <FileBrowserTab />
          </Tabs.Content>
          <Tabs.Content value="toolbox" className="h-full">
            <ToolboxTab />
          </Tabs.Content>
          <Tabs.Content value="containers" className="h-full">
            <ContainersTab />
          </Tabs.Content>
          <Tabs.Content value="git" className="h-full">
            <GitTab />
          </Tabs.Content>
          <Tabs.Content value="servers" className="h-full">
            <ServersTab />
          </Tabs.Content>
          <Tabs.Content value="ai-suggestions" className="h-full">
            <AISuggestionsTab />
          </Tabs.Content>
          <Tabs.Content value="environment" className="h-full">
            <EnvironmentTab />
          </Tabs.Content>
          <Tabs.Content value="log-explorer" className="h-full">
            <LogExplorerTab />
          </Tabs.Content>
        </div>
      </Tabs.Root>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Overview Tab
// ══════════════════════════════════════════════

function OverviewTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [refreshKey] = useState(0)

  // ── Docker status ──
  const { data: dockerData } = useQuery<DockerStatus>({
    queryKey: ['devops-docker-status', refreshKey],
    queryFn: async () => {
      try {
        const res = await call('DevOps.GetDockerStatus')
        return res as DockerStatus
      } catch {
        return { installed: false, running: false, version: '', containers: { running: 0, stopped: 0, failed: 0, total: 0, containers: [] } }
      }
    },
    refetchInterval: refreshInterval,
  })

  // ── Kubernetes status ──
  const { data: k8sData } = useQuery<KubernetesStatus>({
    queryKey: ['devops-k8s-status', refreshKey],
    queryFn: async () => {
      try {
        const res = await call('DevOps.GetKubernetesStatus')
        return res as KubernetesStatus
      } catch {
        return { installed: false, connected: false, cluster: '', nodes: 0, pods: 0 }
      }
    },
    refetchInterval: refreshInterval,
  })

  // ── Service categories ──
  const { data: categories = [], isLoading: categoriesLoading } = useQuery<ServiceCategory[]>({
    queryKey: ['devops-service-categories', refreshKey],
    queryFn: async () => {
      try {
        const res = await call('DevOps.GetServiceCategories')
        return (res as ServiceCategory[]) || []
      } catch {
        return []
      }
    },
    refetchInterval: refreshInterval,
  })

  // ── Service group summary ──
  const { data: summary } = useQuery<ServiceGroupSummary>({
    queryKey: ['devops-service-summary', refreshKey],
    queryFn: async () => {
      try {
        const res = await call('DevOps.GetServiceGroupSummary')
        return res as ServiceGroupSummary
      } catch {
        return { databases: 0, messageQueues: 0, webServers: 0, containers: 0, other: 0, running: 0, stopped: 0 }
      }
    },
    refetchInterval: refreshInterval,
  })

  // ── AI Suggestions ──
  const { data: suggestions } = useQuery<DevOpsSuggestion[]>({
    queryKey: ['devops-suggestions-overview', refreshKey],
    queryFn: async () => {
      try {
        const res = await call('DevOps.GetAISuggestions')
        return (res as DevOpsSuggestion[]) || []
      } catch {
        return []
      }
    },
    refetchInterval: 120000,
  })

  // ── Git Summary ──
  const { data: gitSummary } = useQuery<GitSummary>({
    queryKey: ['devops-git-summary-overview', refreshKey],
    queryFn: async () => {
      try {
        const res = await call('DevOps.GetGitSummary')
        return res as GitSummary
      } catch {
        return { repositories: [], total_repos: 0 }
      }
    },
    refetchInterval: refreshInterval,
  })

  // ── Environment Summary ──
  const { data: environment } = useQuery<EnvironmentInfo>({
    queryKey: ['devops-env-overview', refreshKey],
    queryFn: async () => {
      try {
        const res = await call('DevOps.GetEnvironment')
        return (res as EnvironmentInfo) || { path_dirs: [], key_vars: [], sdks: [], package_managers: [] }
      } catch {
        return { path_dirs: [], key_vars: [], sdks: [], package_managers: [] }
      }
    },
    refetchInterval: refreshInterval,
  })

  const categoryIcons: Record<string, string> = {
    Databases: '\uD83D\uDDC4\uFE0F',
    'Message Queues': '\uD83D\uDCE8',
    'Web Servers': '\uD83C\uDF10',
    Containers: '\uD83D\uDCE6',
    Other: '\u2699\uFE0F',
  }

  const summaryKpis = [
    { label: 'Databases', value: summary?.databases ?? 0, colorClass: 'text-accent' },
    { label: 'Message Queues', value: summary?.messageQueues ?? 0, colorClass: 'text-accent-2' },
    { label: 'Web Servers', value: summary?.webServers ?? 0, colorClass: 'text-success' },
    { label: 'Running', value: summary?.running ?? 0, colorClass: 'text-success' },
    { label: 'Stopped', value: summary?.stopped ?? 0, colorClass: 'text-danger' },
  ]

  return (
    <div className="flex flex-col h-full p-8 space-y-8 overflow-y-auto">
      {/* ── Docker & Kubernetes Row ── */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Docker Status Card */}
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <Box size={28} className="text-accent" />
              <h3 className="text-xl font-bold text-text uppercase tracking-widest">Docker</h3>
            </div>
            <span
              className={cn(
                'text-xs font-bold px-3 py-1 rounded-full uppercase tracking-widest border',
                dockerData?.installed
                  ? dockerData?.running
                    ? 'bg-success/20 text-success border-success/30'
                    : 'bg-warning/20 text-warning border-warning/30'
                  : 'bg-text-faint/20 text-text-faint border-text-faint/30',
              )}
            >
              {dockerData?.installed ? (dockerData?.running ? 'Running' : 'Installed') : 'Not Installed'}
            </span>
          </div>

          {!dockerData?.installed ? (
            <p className="text-text-faint text-sm">Docker is not installed on this system.</p>
          ) : (
            <div className="space-y-6">
              {dockerData.version && (
                <div className="flex items-center gap-2 text-sm text-text-dim">
                  <span className="text-text-faint">Version:</span>
                  <span className="font-[Geist_Mono] text-text font-medium">{dockerData.version}</span>
                </div>
              )}
              <div className="grid grid-cols-4 gap-4">
                {[
                  { label: 'Running', value: dockerData.containers?.running ?? 0, dotClass: 'bg-success', textClass: 'text-success' },
                  { label: 'Stopped', value: dockerData.containers?.stopped ?? 0, dotClass: 'bg-warning', textClass: 'text-warning' },
                  { label: 'Failed', value: dockerData.containers?.failed ?? 0, dotClass: 'bg-danger', textClass: 'text-danger' },
                  { label: 'Total', value: dockerData.containers?.total ?? 0, dotClass: 'bg-text-faint', textClass: 'text-text-dim' },
                ].map((item) => (
                  <div key={item.label} className="flex flex-col items-center gap-2 bg-panel-2 border border-border rounded-xl p-4">
                    <span className={cn('w-2.5 h-2.5 rounded-full', item.dotClass)} />
                    <span className={cn('text-2xl font-bold tabular-nums', item.textClass)}>{item.value}</span>
                    <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">{item.label}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Kubernetes Status Card */}
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <Server size={28} className="text-info" />
              <h3 className="text-xl font-bold text-text uppercase tracking-widest">Kubernetes</h3>
            </div>
            <span
              className={cn(
                'text-xs font-bold px-3 py-1 rounded-full uppercase tracking-widest border',
                k8sData?.installed
                  ? k8sData?.connected
                    ? 'bg-success/20 text-success border-success/30'
                    : 'bg-warning/20 text-warning border-warning/30'
                  : 'bg-text-faint/20 text-text-faint border-text-faint/30',
              )}
            >
              {k8sData?.installed ? (k8sData?.connected ? 'Connected' : 'Not Connected') : 'Not Installed'}
            </span>
          </div>

          {!k8sData?.installed ? (
            <p className="text-text-faint text-sm">Kubernetes is not installed on this system.</p>
          ) : !k8sData?.connected ? (
            <div className="space-y-2">
              <p className="text-warning text-sm font-medium">Installed but not connected to a cluster.</p>
              {k8sData.cluster && (
                <p className="text-text-faint text-xs">Last known cluster: <span className="font-[Geist_Mono] text-text-dim">{k8sData.cluster}</span></p>
              )}
            </div>
          ) : (
            <div className="space-y-6">
              <div className="flex items-center gap-2 text-sm text-text-dim">
                <span className="text-text-faint">Cluster:</span>
                <span className="font-[Geist_Mono] text-text font-medium">{k8sData.cluster}</span>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col items-center gap-2 bg-panel-2 border border-border rounded-xl p-4">
                  <Server size={20} className="text-info" />
                  <span className="text-2xl font-bold tabular-nums text-info">{k8sData.nodes}</span>
                  <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">Nodes</span>
                </div>
                <div className="flex flex-col items-center gap-2 bg-panel-2 border border-border rounded-xl p-4">
                  <Container size={20} className="text-accent" />
                  <span className="text-2xl font-bold tabular-nums text-accent">{k8sData.pods}</span>
                  <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">Pods</span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* ── Service Group Summary Row ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Activity size={22} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Service Overview</h3>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
          {summaryKpis.map((kpi) => (
            <div key={kpi.label} className="flex flex-col items-center gap-2 bg-panel-2 border border-border rounded-xl p-5">
              <span className={cn('text-3xl font-bold tabular-nums', kpi.colorClass)}>{kpi.value}</span>
              <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">{kpi.label}</span>
            </div>
          ))}
        </div>
      </div>

      {/* ── Service Categories ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-8">
          <Package size={22} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Service Categories</h3>
        </div>

        {categoriesLoading ? (
          <div className="flex items-center justify-center py-12">
            <RefreshCw size={24} className="text-text-faint animate-spin" />
          </div>
        ) : categories.length === 0 ? (
          <EmptyState
            icon={<Package size={28} />}
            title="No Categories Available"
            description="Service categories are not yet available. The backend may still be initializing."
          />
        ) : (
          <div className="space-y-8">
            {categories.map((cat) => (
              <div key={cat.category} className="space-y-3">
                <div className="flex items-center gap-3">
                  <span className="text-lg">{categoryIcons[cat.category] ?? '\u2699\uFE0F'}</span>
                  <h4 className="text-base font-bold text-text uppercase tracking-wider">{cat.category}</h4>
                  <span className="px-2 py-0.5 rounded-full bg-accent-soft text-accent text-xs font-bold">{cat.services.length}</span>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                  {cat.services.map((svc) => (
                    <div
                      key={svc.name}
                      className="flex items-center justify-between bg-panel-2 border border-border rounded-xl px-4 py-3 hover:bg-[var(--color-sidebar-hover)] transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <span
                          className={cn(
                            'w-2.5 h-2.5 rounded-full',
                            svc.status.toLowerCase() === 'running' ? 'bg-success shadow-[0_0_6px_var(--color-success)]' : 'bg-danger shadow-[0_0_6px_var(--color-danger)]',
                          )}
                        />
                        <span className="text-sm font-medium text-text font-[Geist_Mono]">{svc.name}</span>
                      </div>
                      <span className="text-xs text-text-faint">{svc.port > 0 ? `:${svc.port}` : ''}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ── AI Suggestions Compact ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
        <div className="flex items-center gap-3 mb-4">
          <Lightbulb size={22} className="text-warning" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">AI Suggestions</h3>
        </div>
        {(!suggestions || suggestions.length === 0) ? (
          <p className="text-text-faint text-sm">No suggestions available.</p>
        ) : (
          <div className="space-y-3">
            {(() => {
              const critical = suggestions.filter((s) => s.severity === 'critical').length
              const warnings = suggestions.filter((s) => s.severity === 'warning').length
              const infos = suggestions.filter((s) => s.severity === 'info').length
              return (
                <div className="flex items-center gap-6">
                  {critical > 0 && (
                    <div className="flex items-center gap-2">
                      <XCircle size={16} className="text-danger" />
                      <span className="text-sm font-bold text-danger">{critical} Critical</span>
                    </div>
                  )}
                  {warnings > 0 && (
                    <div className="flex items-center gap-2">
                      <AlertTriangle size={16} className="text-warning" />
                      <span className="text-sm font-bold text-warning">{warnings} Warning{warnings !== 1 ? 's' : ''}</span>
                    </div>
                  )}
                  {infos > 0 && (
                    <div className="flex items-center gap-2">
                      <CheckCircle2 size={16} className="text-success" />
                      <span className="text-sm font-bold text-success">{infos} Info</span>
                    </div>
                  )}
                  <span className="text-xs text-text-faint">{suggestions.length} total</span>
                </div>
              )
            })()}
            {suggestions.slice(0, 3).map((s, i) => (
              <div key={`${s.category}-${i}`} className="flex items-start gap-3 bg-panel-2 border border-border rounded-xl px-4 py-3">
                <div className="flex-shrink-0 mt-0.5">
                  {s.severity === 'critical' ? <XCircle size={14} className="text-danger" /> : s.severity === 'warning' ? <AlertTriangle size={14} className="text-warning" /> : <CheckCircle2 size={14} className="text-success" />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-text truncate">{s.message}</p>
                  <p className="text-xs text-text-faint mt-0.5">{s.category}</p>
                </div>
              </div>
            ))}
            {suggestions.length > 3 && (
              <p className="text-xs text-text-faint text-center">+{suggestions.length - 3} more — see AI Insights tab</p>
            )}
          </div>
        )}
      </div>

      {/* ── Git Summary Compact ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
        <div className="flex items-center gap-3 mb-4">
          <GitBranch size={22} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Git Summary</h3>
        </div>
        {!gitSummary || gitSummary.total_repos === 0 ? (
          <p className="text-text-faint text-sm">No git repositories found.</p>
        ) : (
          <div className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
              <div className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
                <span className="text-2xl font-bold tabular-nums text-accent">{gitSummary.total_repos}</span>
                <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">Repos</span>
              </div>
              <div className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
                <span className="text-2xl font-bold tabular-nums text-warning">
                  {gitSummary.repositories.reduce((sum, r) => sum + r.modified_files, 0)}
                </span>
                <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">Modified</span>
              </div>
              <div className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
                <span className="text-2xl font-bold tabular-nums text-text-dim">
                  {gitSummary.repositories.filter((r) => r.clean).length}
                </span>
                <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">Clean</span>
              </div>
            </div>
            <div className="space-y-1.5">
              {gitSummary.repositories.slice(0, 4).map((repo) => (
                <div key={repo.path} className="flex items-center justify-between bg-panel-2 border border-border rounded-lg px-4 py-2">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className={cn('w-2 h-2 rounded-full flex-shrink-0', repo.clean ? 'bg-success' : 'bg-warning')} />
                    <span className="text-sm font-[Geist_Mono] text-text truncate" title={repo.path}>{repo.path.split(/[/\\]/).pop()}</span>
                  </div>
                  <div className="flex items-center gap-3 text-xs text-text-faint">
                    <span className="font-[Geist_Mono]">{repo.branch}</span>
                    {(repo.ahead > 0 || repo.behind > 0) && (
                      <span className="text-warning font-bold">↑{repo.ahead} ↓{repo.behind}</span>
                    )}
                  </div>
                </div>
              ))}
              {gitSummary.total_repos > 4 && (
                <p className="text-xs text-text-faint text-center">+{gitSummary.total_repos - 4} more — see Git tab</p>
              )}
            </div>
          </div>
        )}
      </div>

      {/* ── Environment Summary Compact ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
        <div className="flex items-center gap-3 mb-4">
          <Terminal size={22} className="text-success" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Environment</h3>
        </div>
        {!environment ? (
          <p className="text-text-faint text-sm">Loading environment...</p>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
              <span className="text-2xl font-bold tabular-nums text-accent">{environment.path_dirs.length}</span>
              <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">PATH Dirs</span>
            </div>
            <div className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
              <span className="text-2xl font-bold tabular-nums text-warning">{environment.key_vars.length}</span>
              <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">Key Vars</span>
            </div>
            <div className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
              <span className="text-2xl font-bold tabular-nums text-success">{environment.sdks.length}</span>
              <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">SDKs</span>
            </div>
            <div className="flex flex-col items-center gap-1 bg-panel-2 border border-border rounded-xl p-4">
              <span className="text-2xl font-bold tabular-nums text-info">{environment.package_managers.length}</span>
              <span className="text-xs font-semibold text-text-faint uppercase tracking-wider">Pkg Mgrs</span>
            </div>
          </div>
        )}
        {environment && environment.sdks.length > 0 && (
          <div className="mt-4 flex flex-wrap gap-2">
            {environment.sdks.map((sdk) => (
              <span key={sdk.name} className="text-xs font-[Geist_Mono] font-medium bg-success/10 text-success border border-success/20 rounded-full px-3 py-1">
                {sdk.name} {sdk.version}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Terminal Tab
// ══════════════════════════════════════════════

function TerminalTab() {
  const { call } = useBackend()
  const [input, setInput] = useState('')
  const [output, setOutput] = useState<string[]>([`Hawkward Terminal v1.0\nType a command and press Enter to run. Permissions will be requested for impactful actions.\n`])
  const [history, setHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [isRunning, setIsRunning] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingCmd, setPendingCmd] = useState('')
  const outputRef = useRef<HTMLDivElement>(null)

  const runCommand = useCallback(async (cmd: string) => {
    if (!cmd.trim() || isRunning) return
    setIsRunning(true)
    setHistory(prev => [...prev, cmd])
    setOutput(prev => [...prev, `$ ${cmd}`])

    try {
      const res = await call('DevOps.RunCommand', cmd) as CommandResult
      if (res.error) {
        setOutput(prev => [...prev, `\u001b[31mError: ${res.error}\u001b[0m`])
      } else {
        setOutput(prev => [...prev, res.output || 'Command completed successfully (no output).'])
      }
    } catch (err) {
      setOutput(prev => [...prev, `\u001b[31mExecution Error: ${String(err)}\u001b[0m`])
    } finally {
      setIsRunning(false)
    }
  }, [call, isRunning])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (input.trim()) {
        const lowerCmd = input.trim().toLowerCase()
        const dangerous = ['rm ', 'del ', 'format', 'mkfs', 'kill', 'stop-process', 'shutdown', 'restart', 'set-service', 'remove-item', 'stop-service']
        if (dangerous.some(d => lowerCmd.includes(d))) {
          setPendingCmd(input)
          setConfirmOpen(true)
        } else {
          runCommand(input)
          setInput('')
          setHistoryIndex(-1)
        }
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (history.length > 0) {
        const newIdx = historyIndex === -1 ? history.length - 1 : Math.max(historyIndex - 1, 0)
        setHistoryIndex(newIdx)
        setInput(history[newIdx])
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIndex >= 0 && historyIndex < history.length - 1) {
        const newIdx = historyIndex + 1
        setHistoryIndex(newIdx)
        setInput(history[newIdx])
      } else {
        setHistoryIndex(-1)
        setInput('')
      }
    }
  }

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight
    }
  }, [output])

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <ConfirmDialog
        open={confirmOpen}
        title="Impactful Command Detected"
        description={`You are about to run a command that may modify system state or delete data: "${pendingCmd}". Are you sure you want to proceed?`}
        type="danger"
        confirmText="Execute Command"
        onConfirm={() => {
          runCommand(pendingCmd)
          setInput('')
          setPendingCmd('')
        }}
        onClose={() => setConfirmOpen(false)}
      />

      {/* Command input */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Enter shell command (cmd/sh)..."
            className="w-full bg-[var(--color-bg)] border border-border rounded-xl pl-12 pr-4 py-4 text-lg font-[Geist_Mono] text-text placeholder-text-faint focus:outline-none focus:border-primary shadow-inner"
            disabled={isRunning}
          />
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-success text-xl font-bold font-[Geist_Mono]">$</span>
        </div>
        <button
          onClick={() => { if (input.trim()) runCommand(input) }}
          disabled={isRunning || !input.trim()}
          className="flex items-center gap-2 px-8 py-4 text-lg font-bold bg-accent text-white rounded-xl hover:bg-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg active:scale-95"
        >
          <Play size={20} />
          Execute
        </button>
        <button
          onClick={() => setOutput([`Output cleared.\n`])}
          className="flex items-center gap-2 px-6 py-4 text-lg font-bold text-text-dim border border-border rounded-xl hover:bg-panel-3 transition-all"
        >
          <Trash2 size={20} />
          Clear
        </button>
      </div>

      {/* Output */}
      <div
        ref={outputRef}
        className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl p-8 overflow-y-auto font-[Geist_Mono] text-lg leading-relaxed whitespace-pre-wrap shadow-inner"
      >
        {output.map((block, i) => (
          <div key={i} className="whitespace-pre-wrap break-all mb-2">
            {stripAnsi(block)}
          </div>
        ))}
        {!isRunning && output.length > 0 && (
          <div className="inline-flex items-center">
            <span className="text-[var(--color-success)] text-lg font-bold font-[Geist_Mono]">$</span>
            <span className="inline-block w-2.5 h-5 bg-[var(--color-success)] ml-1 animate-pulse" style={{ animationDuration: '1s' }} />
          </div>
        )}
        {isRunning && (
          <div className="flex items-center gap-2 mt-2">
            <span className="inline-block w-3 h-6 bg-[var(--color-success)] animate-pulse" />
            <span className="text-sm font-bold text-[var(--color-success)] uppercase tracking-widest animate-pulse">Running...</span>
          </div>
        )}
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  PowerShell Pro Tab
// ══════════════════════════════════════════════

function PowerShellProTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [isRunning, setIsRunning] = useState(false)
  const [output, setOutput] = useState('')
  const [selectedWorkflow, setSelectedWorkflow] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)

  const { data: workflows = [] } = useQuery<string[]>({
    queryKey: ['devops-workflows'],
    queryFn: async () => {
      const res = await call('DevOps.GetPowerShellWorkflows')
      return (res as string[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const runWorkflow = async (name: string) => {
    setIsRunning(true)
    setOutput(`Running workflow: ${name}...\n`)
    try {
      const res = await call('DevOps.RunPowerShell', name) as CommandResult
      if (res.error) setOutput(prev => prev + `\u001b[31mError: ${res.error}\u001b[0m`)
      else setOutput(prev => prev + res.output)
    } catch (err) {
      setOutput(prev => prev + `Execution Error: ${String(err)}`)
    } finally {
      setIsRunning(false)
    }
  }

  const workflowIcons: Record<string, React.ReactNode> = {
    'Invoke-HawkDailyOps': <Activity size={24} />,
    'Invoke-HawkSystemReview': <Search size={24} />,
    'Invoke-HawkSecurityAudit': <ShieldCheck size={24} />,
    'Invoke-HawkNetworkDiagnostics': <Globe size={24} />,
    'Invoke-HawkThreatHunt': <Crosshair size={24} />,
    'Invoke-HawkChangeAudit': <History size={24} />,
    'Invoke-HawkComplianceCheck': <FileCheck size={24} />,
  }

  const workflowDescs: Record<string, string> = {
    'Invoke-HawkDailyOps': 'Aggregate health check, storage audit, and network connectivity.',
    'Invoke-HawkSystemReview': 'Deep dive into hardware specs, performance metrics, and licensing.',
    'Invoke-HawkSecurityAudit': 'Full audit of firewall, startup entries, tasks, and defender status.',
    'Invoke-HawkNetworkDiagnostics': 'Verify internet, DNS resolvers, interfaces, and network shares.',
    'Invoke-HawkThreatHunt': 'Search for suspicious processes, ghost ports, and file anomalies.',
    'Invoke-HawkChangeAudit': 'Review recent files, patches, driver changes, and crash dumps.',
    'Invoke-HawkComplianceCheck': 'CIS-inspired baseline verification and compliance scoring.',
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 h-full p-8 gap-8 overflow-hidden">
      <ConfirmDialog
        open={confirmOpen}
        title="Execute PowerShell Workflow"
        description={`This will run the advanced diagnostic workflow "${selectedWorkflow}" which performs multiple system queries. Proceed?`}
        onConfirm={() => runWorkflow(selectedWorkflow)}
        onClose={() => setConfirmOpen(false)}
      />

      {/* Workflow Selection */}
      <div className="lg:col-span-1 space-y-4 overflow-y-auto pr-2">
        <h3 className="text-lg font-bold text-text-dim uppercase tracking-widest mb-4">Diagnostic Workflows</h3>
        {workflows.map(wf => (
          <button
            key={wf}
            onClick={() => {
              setSelectedWorkflow(wf)
              setConfirmOpen(true)
            }}
            disabled={isRunning}
            className="w-full text-left bg-panel border border-border rounded-2xl p-6 transition-all hover:border-accent/50 hover:bg-accent/5 group disabled:opacity-50"
          >
            <div className="flex items-center gap-4 mb-3">
              <div className="w-12 h-12 rounded-xl bg-panel-3 flex items-center justify-center text-text-dim group-hover:text-accent transition-colors border border-border">
                {workflowIcons[wf] || <Zap size={24} />}
              </div>
              <span className="text-xl font-bold text-text group-hover:text-accent transition-colors">{wf.replace('Invoke-Hawk', '')}</span>
            </div>
            <p className="text-text-dim text-base leading-relaxed">{workflowDescs[wf]}</p>
          </button>
        ))}
      </div>

      {/* Output Console */}
      <div className="lg:col-span-2 flex flex-col space-y-4 min-w-0">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-text-dim uppercase tracking-widest">Workflow Output</h3>
          {isRunning && (
            <div className="flex items-center gap-2 text-warning animate-pulse">
              <Zap size={16} />
              <span className="text-sm font-bold uppercase tracking-tighter">Processing...</span>
            </div>
          )}
        </div>
        <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl p-8 overflow-y-auto font-[Geist_Mono] text-lg leading-relaxed whitespace-pre shadow-inner">
          {stripAnsi(output) || 'Select a workflow to begin diagnostic execution.'}
        </div>
        <button
          onClick={() => setOutput('')}
          className="self-end px-6 py-3 text-lg font-bold text-text-dim border border-border rounded-xl hover:bg-panel-3 transition-all"
        >
          Clear Console
        </button>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Services Tab
// ══════════════════════════════════════════════

function ServicesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [search, setSearch] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ name: string; action: string } | null>(null)
  const queryClient = useQueryClient()

  const { data: services = [], isLoading: loading, refetch: refetchServices } = useQuery<ServiceEntry[]>({
    queryKey: ['devops-services'],
    queryFn: async () => {
      const res = await call('DevOps.GetServices')
      return (res as ServiceEntry[]) || []
    },
    refetchInterval: refreshInterval,
  })

  const controlService = async (name: string, action: string) => {
    const success = await call('DevOps.ControlService', name, action)
    if (success) {
      queryClient.invalidateQueries({ queryKey: ['devops-services'] })
    }
    setPendingAction(null)
  }

  const filtered = services.filter(s =>
    s.name.toLowerCase().includes(search.toLowerCase()) ||
    s.display_name.toLowerCase().includes(search.toLowerCase()),
  )

  const runningCount = services.filter(s => s.status === 'Running').length
  const stoppedCount = services.filter(s => s.status === 'Stopped').length

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <ConfirmDialog
        open={confirmOpen}
        title="Service Control Permission"
        description={`You are about to perform a ${pendingAction?.action} operation on the service "${pendingAction?.name}". This may affect system availability.`}
        type={pendingAction?.action === 'stop' ? 'danger' : 'warning'}
        confirmText={`${pendingAction?.action?.toUpperCase()} SERVICE`}
        onConfirm={() => {
          if (pendingAction) controlService(pendingAction.name, pendingAction.action)
        }}
        onClose={() => setConfirmOpen(false)}
      />

      <div className="flex items-center gap-6 flex-wrap">
        <div className="flex items-center gap-8 bg-panel border border-border px-6 py-3 rounded-2xl">
          <div className="flex items-center gap-3">
            <span className="w-3 h-3 rounded-full bg-success shadow-[0_0_8px_var(--color-success)]" />
            <span className="text-lg font-bold"><span className="text-success">{runningCount}</span> Running</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="w-3 h-3 rounded-full bg-danger shadow-[0_0_8px_var(--color-danger)]" />
            <span className="text-lg font-bold"><span className="text-danger">{stoppedCount}</span> Stopped</span>
          </div>
          <div className="text-text-faint text-lg font-medium">| &nbsp; {services.length} Total</div>
        </div>

        <div className="flex-1" />

        <div className="relative group w-80">
          <Search size={20} className="absolute left-4 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search services..."
            className="w-full bg-panel border border-border rounded-xl pl-12 pr-4 py-3 text-lg text-text placeholder-text-faint focus:outline-none focus:border-accent transition-all shadow-inner"
          />
        </div>

        <button
          onClick={() => refetchServices()}
          className="p-3.5 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md active:rotate-180 duration-500"
        >
          <RefreshCw size={24} />
        </button>
      </div>

      <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl overflow-hidden shadow-inner">
        <div className="overflow-y-auto h-full">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Name</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Display Name</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Status</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Startup</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">Control</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-16 text-center">
                    <div className="flex flex-col items-center gap-3">
                      <div className="w-10 h-10 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center">
                        <RefreshCw size={20} className="text-[var(--color-text-faint)] animate-spin" />
                      </div>
                      <p className="text-sm font-semibold text-[var(--color-text-dim)]">Enumerating services...</p>
                    </div>
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-16 text-center">
                    <EmptyState
                      icon={<Server size={28} />}
                      title="No Services Found"
                      description={search ? 'No services match your search query. Try adjusting your filter.' : 'No system services were detected on this machine.'}
                    />
                  </td>
                </tr>
              ) : (
                filtered.map((svc) => (
                  <tr key={svc.name} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-colors group">
                    <td className="px-6 py-4 font-[Geist_Mono] text-sm text-accent font-medium">{svc.name}</td>
                    <td className="px-8 py-4 text-sm text-[var(--color-text)]">{svc.display_name}</td>
                    <td className="px-8 py-4"><StatusBadge status={svc.status} /></td>
                    <td className="px-8 py-4"><StatusBadge status={svc.start_type} /></td>
                    <td className="px-8 py-4 text-right">
                      <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                          onClick={() => {
                            setPendingAction({ name: svc.name, action: 'start' })
                            setConfirmOpen(true)
                          }}
                          className="p-2 text-success hover:bg-success/10 rounded-lg"
                          title="Start"
                        >
                          <PlayCircle size={20} />
                        </button>
                        <button
                          onClick={() => {
                            setPendingAction({ name: svc.name, action: 'stop' })
                            setConfirmOpen(true)
                          }}
                          className="p-2 text-danger hover:bg-danger/10 rounded-lg"
                          title="Stop"
                        >
                          <StopCircle size={20} />
                        </button>
                        <button
                          onClick={() => {
                            setPendingAction({ name: svc.name, action: 'restart' })
                            setConfirmOpen(true)
                          }}
                          className="p-2 text-warning hover:bg-warning/10 rounded-lg"
                          title="Restart"
                        >
                          <RotateCcw size={20} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  File Browser Tab
// ══════════════════════════════════════════════

function FileBrowserTab() {
  const { call } = useBackend()
  const [currentPath, setCurrentPath] = useState('')
  const [entries, setEntries] = useState<FileEntry[]>([])
  const [history, setHistory] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [previewFile, setPreviewFile] = useState<FileEntry | null>(null)
  const [fileContent, setPreviewContent] = useState('')

  const fetchDir = useCallback(async (path: string) => {
    setLoading(true)
    try {
      const res = await call('DevOps.ListDirectory', path)
      setEntries((res as FileEntry[]) || [])
      setCurrentPath(path)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [call])

  useEffect(() => {
    if (!currentPath) {
      call('DevOps.GetDefaultPath').then((p) => {
        const path = (p as string) || ''
        setCurrentPath(path)
        if (path) fetchDir(path)
      }).catch(() => {
        setCurrentPath('')
      })
    } else {
      fetchDir(currentPath)
    }
  }, [fetchDir, currentPath, call])

  const navigate = (path: string) => {
    setHistory(prev => [...prev, currentPath])
    fetchDir(path)
    setPreviewFile(null)
  }

  const goBack = () => {
    if (history.length > 0) {
      const prev = history[history.length - 1]
      setHistory(h => h.slice(0, -1))
      fetchDir(prev)
      setPreviewFile(null)
    }
  }

  const openFile = async (file: FileEntry) => {
    if (file.is_dir) {
      navigate(file.path)
    } else {
      setPreviewFile(file)
      const content = await call('DevOps.ReadFile', file.path) as string
      setPreviewContent(content)
    }
  }

  return (
    <div className="flex h-full p-8 gap-8">
      {/* File list */}
      <div className={cn('flex flex-col flex-1 min-w-0 space-y-4', previewFile ? 'w-1/2' : 'w-full')}>
        <div className="flex items-center gap-4">
          <button
            onClick={goBack}
            disabled={history.length === 0}
            className="p-3 bg-panel border border-border rounded-xl hover:bg-panel-3 disabled:opacity-30 transition-all shadow-md"
          >
            <ChevronRight size={24} className="rotate-180" />
          </button>
          <div className="flex-1 bg-panel border border-border rounded-xl px-6 py-2.5 flex items-center gap-1 overflow-x-auto no-scrollbar shadow-inner">
            <button
              onClick={() => call('DevOps.GetDefaultPath').then(p => navigate(p as string))}
              className="p-1.5 hover:bg-[var(--color-sidebar-hover)] rounded-lg text-text-faint hover:text-accent transition-all shrink-0"
            >
              <Home size={18} />
            </button>
            {currentPath.split(/[\\/]/).filter(Boolean).map((part, i, arr) => {
              const fullPath = currentPath.split(/[\\/]/).slice(0, i + 2).join('\\') // Handle Windows paths
              const isLast = i === arr.length - 1
              return (
                <div key={i} className="flex items-center gap-1 shrink-0">
                  <ChevronRight size={14} className="text-text-dim" />
                  <button
                    onClick={() => !isLast && navigate(fullPath)}
                    className={cn(
                      "px-2 py-1 rounded-lg text-sm font-bold transition-all",
                      isLast ? "text-text cursor-default" : "text-text-dim hover:text-accent hover:bg-[var(--color-sidebar-hover)]"
                    )}
                  >
                    {part}
                  </button>
                </div>
              )
            })}
          </div>
        </div>

        <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl overflow-hidden shadow-inner">
          <div className="overflow-y-auto h-full">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                <tr>
                  <th className="px-8 py-4 text-sm font-bold text-text-dim uppercase tracking-widest">Name</th>
                  <th className="px-8 py-4 text-sm font-bold text-text-dim uppercase tracking-widest text-right">Size</th>
                  <th className="px-8 py-4 text-sm font-bold text-text-dim uppercase tracking-widest">Modified</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={3} className="px-8 py-20 text-center animate-pulse text-xl text-text-faint">Reading File System...</td></tr>
                ) : (
                  entries.map((file) => (
                    <tr
                      key={file.path}
                      onClick={() => openFile(file)}
                      className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] cursor-pointer transition-colors group"
                    >
                      <td className="px-8 py-4 flex items-center gap-4">
                        {file.is_dir ? <Folder size={24} className="text-accent" /> : <FileText size={24} className="text-text-faint" />}
                        <span className="text-lg text-text group-hover:text-accent transition-colors">{file.name}</span>
                      </td>
                      <td className="px-8 py-4 text-right font-[Geist_Mono] text-base text-text-dim">{file.size}</td>
                      <td className="px-8 py-4 text-base text-text-faint">{new Date(file.mod_time).toLocaleString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Preview panel */}
      {previewFile && (
        <div className="w-1/2 flex flex-col space-y-4 animate-in slide-in-from-right-8 duration-300">
          <div className="flex items-center justify-between p-2">
            <h3 className="text-xl font-bold text-text flex items-center gap-3 min-w-0">
              <FileText size={24} className="text-accent shrink-0" />
              <span className="truncate">{previewFile.name}</span>
            </h3>
            <div className="flex items-center gap-2">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(previewFile.path)
                  // Could add a toast here
                }}
                className="p-2.5 rounded-lg bg-panel border border-border hover:bg-panel-3 text-text-faint hover:text-text transition-all"
                title="Copy full path"
              >
                <Folder size={20} />
              </button>
              <button
                onClick={() => setPreviewFile(null)}
                className="p-2.5 rounded-lg hover:bg-danger/10 text-text-faint hover:text-danger transition-all"
              >
                <X size={28} />
              </button>
            </div>
          </div>
          <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl p-8 overflow-y-auto shadow-inner relative group">
            {previewFile.is_binary ? (
              <div className="flex flex-col items-center justify-center h-full text-center space-y-4 opacity-50">
                <Lock size={48} className="text-warning" />
                <div>
                  <p className="text-xl font-bold text-text">Binary Data Encrypted</p>
                  <p className="text-base text-text-dim">Direct preview is disabled for safety.</p>
                </div>
              </div>
            ) : (
              <pre className="font-[Geist_Mono] text-base text-text leading-relaxed whitespace-pre-wrap">
                {fileContent || '// No readable content or file is empty.'}
              </pre>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

// ══════════════════════════════════════════════
//  Log Explorer Tab
// ══════════════════════════════════════════════

function LogExplorerTab() {
  const { call } = useBackend()
  const [logPath, setLogPath] = useState('hawkward-gui.log')
  const [lines, setLines] = useState<string[]>([])
  const [searchPattern, setSearchPattern] = useState('')
  const [loading, setLoading] = useState(false)

  const handleTail = async () => {
    setLoading(true)
    try {
      const res = await call('DevOps.TailLog', logPath, 100)
      setLines((res as string[]) || [])
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = async () => {
    if (!searchPattern) return
    setLoading(true)
    try {
      const res = await call('DevOps.SearchLog', logPath, searchPattern)
      setLines((res as string[]) || [])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <div className="flex items-end gap-4 bg-panel border border-border p-6 rounded-2xl shadow-lg">
        <div className="flex-1 space-y-2">
          <label className="text-xs font-bold text-text-faint uppercase tracking-widest">Log File Path</label>
          <input
            type="text"
            value={logPath}
            onChange={(e) => setLogPath(e.target.value)}
            className="w-full bg-panel-2 border border-border rounded-lg px-4 py-2 text-sm text-text focus:outline-none focus:border-accent"
          />
        </div>
        <div className="flex-1 space-y-2">
          <label className="text-xs font-bold text-text-faint uppercase tracking-widest">Search Pattern</label>
          <div className="flex gap-2">
            <input
              type="text"
              value={searchPattern}
              onChange={(e) => setSearchPattern(e.target.value)}
              placeholder="Regex or text..."
              className="w-full bg-panel-2 border border-border rounded-lg px-4 py-2 text-sm text-text focus:outline-none focus:border-accent"
            />
            <button onClick={handleSearch} className="px-4 py-2 bg-accent text-white rounded-lg font-bold hover:opacity-90 transition-all shadow-md">
              Search
            </button>
          </div>
        </div>
        <button onClick={handleTail} className="px-6 py-2 bg-panel-3 border border-border rounded-lg font-bold text-text hover:bg-panel transition-all mb-0.5">
          Tail (Last 100)
        </button>
      </div>

      <div className="flex-1 bg-black/40 border border-border rounded-2xl p-6 font-[Geist_Mono] text-sm overflow-auto shadow-inner">
        {loading ? (
          <div className="flex items-center justify-center h-full gap-3 text-text-faint">
            <RefreshCw size={20} className="animate-spin" />
            <span className="font-bold uppercase tracking-widest">Accessing File...</span>
          </div>
        ) : lines.length === 0 ? (
          <div className="flex items-center justify-center h-full text-text-faint italic">
            No entries to display. Enter a path and tail or search.
          </div>
        ) : (
          <div className="space-y-1">
            {lines.map((line, i) => (
              <div key={i} className="whitespace-pre-wrap break-all text-text-dim border-b border-white/5 pb-1 last:border-0">{line}</div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  AI Suggestions Tab
// ══════════════════════════════════════════════

function AISuggestionsTab() {
  const { call } = useBackend()
  const [refreshKey, setRefreshKey] = useState(0)

  const { data: suggestions, isLoading } = useQuery<DevOpsSuggestion[]>({
    queryKey: ['devops-ai-suggestions', refreshKey],
    queryFn: async () => {
      const res = await call('DevOps.GetAISuggestions')
      return (res as DevOpsSuggestion[]) ?? []
    },
  })

  const severityConfig: Record<string, { icon: React.ReactNode; color: string; bg: string; border: string }> = {
    critical: {
      icon: <XCircle size={20} className="text-danger" />,
      color: 'text-danger',
      bg: 'bg-danger/5',
      border: 'border-danger/30',
    },
    warning: {
      icon: <AlertTriangle size={20} className="text-warning" />,
      color: 'text-warning',
      bg: 'bg-warning/5',
      border: 'border-warning/30',
    },
    info: {
      icon: <CheckCircle2 size={20} className="text-success" />,
      color: 'text-success',
      bg: 'bg-success/5',
      border: 'border-success/30',
    },
  }

  const categoryIcon: Record<string, React.ReactNode> = {
    docker: <Container size={16} className="text-accent" />,
    git: <GitBranch size={16} className="text-accent" />,
    node: <Code size={16} className="text-success" />,
    general: <Wrench size={16} className="text-text-dim" />,
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <RefreshCw size={32} className="text-text-dim animate-spin" />
      </div>
    )
  }

  const items = suggestions ?? []
  const critical = items.filter((s) => s.severity === 'critical').length
  const warnings = items.filter((s) => s.severity === 'warning').length
  const infos = items.filter((s) => s.severity === 'info').length

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <div className="flex items-center gap-6 flex-wrap">
        <div className="flex items-center gap-8 bg-panel border border-border px-6 py-3 rounded-2xl">
          <div className="flex items-center gap-3">
            <Lightbulb size={20} className="text-warning" />
            <span className="text-lg font-bold">{items.length} Suggestion{items.length !== 1 ? 's' : ''}</span>
          </div>
          {critical > 0 && <div className="text-text-faint text-lg font-medium">| &nbsp; <span className="text-danger">{critical} Critical</span></div>}
          {warnings > 0 && <div className="text-text-faint text-lg font-medium">| &nbsp; <span className="text-warning">{warnings} Warning{warnings !== 1 ? 's' : ''}</span></div>}
          {infos > 0 && <div className="text-text-faint text-lg font-medium">| &nbsp; <span className="text-success">{infos} Info</span></div>}
        </div>
        <div className="flex-1" />
        <button
          onClick={() => setRefreshKey((k) => k + 1)}
          className="p-3.5 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md active:rotate-180 duration-500"
        >
          <RefreshCw size={24} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-3">
        {items.length === 0 ? (
          <div className="py-16">
            <EmptyState
              icon={<Lightbulb size={28} />}
              title="No Suggestions"
              description="Everything looks good across Docker, Git, and development tools."
            />
          </div>
        ) : (
          items.map((s, i) => {
            const cfg = severityConfig[s.severity] ?? severityConfig.info
            return (
              <div
                key={`${s.category}-${s.message}-${i}`}
                className={cn('flex items-start gap-4 p-5 rounded-xl border bg-panel transition-colors hover:bg-panel-2', cfg.border)}
              >
                <div className="flex-shrink-0 mt-0.5">{cfg.icon}</div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    {categoryIcon[s.category]}
                    <span className="text-xs font-bold uppercase tracking-wider text-text-faint">{s.category}</span>
                    <span className={cn('px-2 py-0.5 rounded text-[10px] font-bold uppercase', cfg.bg, cfg.color)}>
                      {s.severity}
                    </span>
                  </div>
                  <p className="text-sm font-semibold text-text">{s.message}</p>
                  <p className="text-sm text-text-dim mt-1">{s.action}</p>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Toolbox Tab
// ══════════════════════════════════════════════

const toolIcons: Record<string, React.ReactNode> = {
  Git: <GitBranch size={24} className="text-accent" />,
  Docker: <Box size={24} className="text-accent" />,
  'Node.js': <Code size={24} className="text-success" />,
  Go: <Code size={24} className="text-info" />,
  Python: <Code size={24} className="text-warning" />,
  Java: <Code size={24} className="text-danger" />,
  Rust: <Code size={24} className="text-accent" />,
  '.NET': <TerminalSquare size={24} className="text-info" />,
}

function ToolboxTab() {
  const { call } = useBackend()

  const { data: tools, isLoading } = useQuery<ToolInfo[]>({
    queryKey: ['devops-tools'],
    queryFn: async () => {
      const res = await call('DevOps.GetInstalledTools')
      return (res as ToolInfo[]) || []
    },
    staleTime: 30_000,
  })

  const installed = tools?.filter((t) => t.status === 'installed') ?? []
  const others = tools?.filter((t) => t.status !== 'installed') ?? []
  const sorted = [...installed, ...others]

  const statusBadge = (status: string) => {
    const colors: Record<string, string> = {
      installed: 'bg-success/20 text-success',
      'not-found': 'bg-text-faint/20 text-text-faint',
      error: 'bg-danger/20 text-danger',
    }
    const labels: Record<string, string> = {
      installed: 'Installed',
      'not-found': 'Not installed',
      error: 'Error',
    }
    return (
      <span className={cn('inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-bold', colors[status])}>
        <span className={cn('w-2 h-2 rounded-full', {
          'bg-success': status === 'installed',
          'bg-text-faint': status === 'not-found',
          'bg-danger': status === 'error',
        })} />
        {labels[status]}
      </span>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <RefreshCw size={32} className="text-text-dim animate-spin" />
      </div>
    )
  }

  if (!tools || tools.length === 0) {
    return <EmptyState icon={<Wrench size={48} />} title="No tools detected" description="Unable to detect installed development tools." />
  }

  return (
    <div className="p-8 space-y-8 overflow-y-auto h-full">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-text">Installed Development Tools</h2>
          <p className="text-text-dim text-sm mt-1">
            {installed.length} of {tools.length} tools detected on this system.
          </p>
        </div>
        <ToolsRefreshButton />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {sorted.map((tool) => (
          <div
            key={tool.name}
            className={cn(
              'flex flex-col gap-3 p-5 rounded-[var(--radius-lg)] border border-border bg-panel transition-all',
              tool.status === 'installed' && 'hover:border-accent/30',
              tool.status === 'not-found' && 'opacity-60',
            )}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                {toolIcons[tool.name] ?? <Code size={24} className="text-text-dim" />}
                <span className="text-base font-bold text-text">{tool.name}</span>
              </div>
              {statusBadge(tool.status)}
            </div>
            {tool.status === 'installed' && (
              <div className="space-y-1">
                <p className="text-sm text-text-dim font-mono">{tool.version}</p>
                {tool.path && (
                  <p className="text-xs text-text-faint truncate" title={tool.path}>
                    {tool.path}
                  </p>
                )}
              </div>
            )}
            {tool.status === 'error' && (
              <p className="text-sm text-danger font-mono">{tool.version}</p>
            )}
            {tool.status === 'not-found' && (
              <p className="text-sm text-text-faint">Not found in PATH</p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

// Simple refresh trigger for the tools query
function ToolsRefreshButton() {
  const queryClient = useQueryClient()
  return (
    <button
      onClick={() => queryClient.invalidateQueries({ queryKey: ['devops-tools'] })}
      className="flex items-center gap-2 px-4 py-2 rounded-[var(--radius-lg)] border border-border bg-panel hover:bg-[var(--color-sidebar-hover)] text-text-dim hover:text-text transition-all text-sm font-bold"
    >
      <RefreshCw size={16} />
      Refresh
    </button>
  )
}

// ══════════════════════════════════════════════
//  Containers Tab
// ══════════════════════════════════════════════

function ContainersTab() {
  const { call } = useBackend()
  const [refreshKey, setRefreshKey] = useState(0)

  const { data, isLoading } = useQuery<ContainerSummary>({
    queryKey: ['devops-containers', refreshKey],
    queryFn: async () => {
      const res = await call('DevOps.GetContainers')
      return (res as ContainerSummary) || { running: 0, stopped: 0, failed: 0, total: 0, containers: [] }
    },
  })

  const stateBadge = (state: string) => {
    const colors: Record<string, string> = {
      running: 'bg-success/20 text-success',
      exited: 'bg-danger/20 text-danger',
      created: 'bg-warning/20 text-warning',
      paused: 'bg-warning/20 text-warning',
    }
    return (
      <span className={cn('px-2 py-0.5 rounded text-xs font-medium border border-current opacity-80', colors[state.toLowerCase()] || 'bg-text-faint/20 text-text-faint')}>
        {state}
      </span>
    )
  }

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <div className="flex items-center gap-6 flex-wrap">
        <div className="flex items-center gap-8 bg-panel border border-border px-6 py-3 rounded-2xl">
          <div className="flex items-center gap-3">
            <span className="w-3 h-3 rounded-full bg-success shadow-[0_0_8px_var(--color-success)]" />
            <span className="text-lg font-bold"><span className="text-success">{data?.running ?? 0}</span> Running</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="w-3 h-3 rounded-full bg-text-faint shadow-[0_0_8px_var(--color-text-faint)]" />
            <span className="text-lg font-bold"><span className="text-text-faint">{data?.stopped ?? 0}</span> Stopped</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="w-3 h-3 rounded-full bg-danger shadow-[0_0_8px_var(--color-danger)]" />
            <span className="text-lg font-bold"><span className="text-danger">{data?.failed ?? 0}</span> Failed</span>
          </div>
          <div className="text-text-faint text-lg font-medium">| &nbsp; {data?.total ?? 0} Total</div>
        </div>
        <div className="flex-1" />
        <button
          onClick={() => setRefreshKey(k => k + 1)}
          className="p-3.5 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md active:rotate-180 duration-500"
        >
          <RefreshCw size={24} />
        </button>
      </div>

      <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl overflow-hidden shadow-inner">
        <div className="overflow-y-auto h-full">
          {isLoading ? (
            <div className="flex flex-col items-center gap-3 py-16">
              <RefreshCw size={20} className="text-[var(--color-text-faint)] animate-spin" />
              <p className="text-sm font-semibold text-[var(--color-text-dim)]">Loading containers...</p>
            </div>
          ) : (data?.containers ?? []).length === 0 ? (
            <div className="py-16">
              <EmptyState
                icon={<Container size={28} />}
                title="No Containers Found"
                description="Docker is not installed or no containers are present."
              />
            </div>
          ) : (
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                <tr>
                  <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Name</th>
                  <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Image</th>
                  <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">State</th>
                  <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Status</th>
                  <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Ports</th>
                </tr>
              </thead>
              <tbody>
                {data?.containers.map((c) => (
                  <tr key={c.id} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-colors">
                    <td className="px-6 py-4 font-[Geist_Mono] text-sm text-accent font-medium">{c.name}</td>
                    <td className="px-6 py-4 text-sm text-[var(--color-text)] font-[Geist_Mono]">{c.image}</td>
                    <td className="px-6 py-4">{stateBadge(c.state)}</td>
                    <td className="px-6 py-4 text-sm text-[var(--color-text-dim)]">{c.status}</td>
                    <td className="px-6 py-4 text-sm text-[var(--color-text-dim)] font-[Geist_Mono] max-w-xs truncate">{c.ports || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Git Tab
// ══════════════════════════════════════════════

function GitTab() {
  const { call } = useBackend()
  const [refreshKey, setRefreshKey] = useState(0)

  const { data, isLoading } = useQuery<GitSummary>({
    queryKey: ['devops-git', refreshKey],
    queryFn: async () => {
      const res = await call('DevOps.GetGitSummary')
      return (res as GitSummary) || { repositories: [], total_repos: 0 }
    },
  })

  const totalModified = (data?.repositories ?? []).reduce((s, r) => s + r.modified_files, 0)
  const totalUntracked = (data?.repositories ?? []).reduce((s, r) => s + r.untracked_files, 0)

  const truncatePath = (p: string, maxLen = 50) => (p.length > maxLen ? '...' + p.slice(-(maxLen - 3)) : p)

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <div className="flex items-center gap-6 flex-wrap">
        <div className="flex items-center gap-8 bg-panel border border-border px-6 py-3 rounded-2xl">
          <div className="flex items-center gap-3">
            <GitBranch size={18} className="text-accent" />
            <span className="text-lg font-bold"><span className="text-accent">{data?.total_repos ?? 0}</span> Repos</span>
          </div>
          <div className="text-text-faint text-lg font-medium">| &nbsp; <span className="text-warning">{totalModified}</span> Modified</div>
          <div className="text-text-faint text-lg font-medium">| &nbsp; <span className="text-text-dim">{totalUntracked}</span> Untracked</div>
        </div>
        <div className="flex-1" />
        <button
          onClick={() => setRefreshKey(k => k + 1)}
          className="p-3.5 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md active:rotate-180 duration-500"
        >
          <RefreshCw size={24} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-4">
        {isLoading ? (
          <div className="flex flex-col items-center gap-3 py-16">
            <RefreshCw size={20} className="text-[var(--color-text-faint)] animate-spin" />
            <p className="text-sm font-semibold text-[var(--color-text-dim)]">Scanning repositories...</p>
          </div>
        ) : (data?.repositories ?? []).length === 0 ? (
          <div className="py-16">
            <EmptyState
              icon={<GitBranch size={28} />}
              title="No Git Repositories Found"
              description="No git repositories were detected in common locations."
            />
          </div>
        ) : (
          (data?.repositories ?? []).map((repo) => (
            <div key={repo.path} className="bg-panel border border-border rounded-xl p-5 hover:border-accent/30 transition-colors">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-2">
                    <span className="text-sm font-[Geist_Mono] text-accent truncate" title={repo.path}>{truncatePath(repo.path)}</span>
                    <span className="px-2 py-0.5 rounded text-xs font-medium bg-accent/20 text-accent border border-accent/30">
                      {repo.branch}
                    </span>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-text-dim">
                    <span>{repo.modified_files} modified</span>
                    <span>{repo.untracked_files} untracked</span>
                    {repo.ahead > 0 && <span className="text-success">{repo.ahead} ahead</span>}
                    {repo.behind > 0 && <span className="text-danger">{repo.behind} behind</span>}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {repo.clean ? (
                    <CheckCircle2 size={20} className="text-success" />
                  ) : (
                    <AlertTriangle size={20} className="text-warning" />
                  )}
                  <span className={cn('text-sm font-medium', repo.clean ? 'text-success' : 'text-warning')}>
                    {repo.clean ? 'Clean' : 'Dirty'}
                  </span>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Servers Tab
// ══════════════════════════════════════════════

function ServersTab() {
  const { call } = useBackend()
  const [refreshKey, setRefreshKey] = useState(0)

  const { data, isLoading } = useQuery<LocalServer[]>({
    queryKey: ['devops-servers', refreshKey],
    queryFn: async () => {
      const res = await call('DevOps.GetLocalServers')
      return (res as LocalServer[]) || []
    },
  })

  const healthBadge = (health: string) => {
    const colors: Record<string, string> = {
      healthy: 'bg-success/20 text-success',
      unknown: 'bg-warning/20 text-warning',
      error: 'bg-danger/20 text-danger',
    }
    const icons: Record<string, React.ReactNode> = {
      healthy: <CheckCircle2 size={14} />,
      unknown: <AlertTriangle size={14} />,
      error: <XCircle size={14} />,
    }
    return (
      <span className={cn('inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium', colors[health] || 'bg-text-faint/20 text-text-faint')}>
        {icons[health] || null}
        {health}
      </span>
    )
  }

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <div className="flex items-center gap-6 flex-wrap">
        <div className="bg-panel border border-border px-6 py-3 rounded-2xl">
          <span className="text-lg font-bold"><span className="text-accent">{data?.length ?? 0}</span> Listening Servers</span>
        </div>
        <div className="flex-1" />
        <button
          onClick={() => setRefreshKey(k => k + 1)}
          className="p-3.5 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md active:rotate-180 duration-500"
        >
          <RefreshCw size={24} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-4">
        {isLoading ? (
          <div className="flex flex-col items-center gap-3 py-16">
            <RefreshCw size={20} className="text-[var(--color-text-faint)] animate-spin" />
            <p className="text-sm font-semibold text-[var(--color-text-dim)]">Scanning listening ports...</p>
          </div>
        ) : (data ?? []).length === 0 ? (
          <div className="py-16">
            <EmptyState
              icon={<Globe size={28} />}
              title="No Local Servers Found"
              description="No listening servers detected on localhost."
            />
          </div>
        ) : (
          (data ?? []).map((srv) => (
            <div key={`${srv.port}-${srv.pid}`} className="bg-panel border border-border rounded-xl p-5 hover:border-accent/30 transition-colors">
              <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-5">
                  <div className="w-14 h-14 rounded-xl bg-accent/10 flex items-center justify-center">
                    <Globe size={24} className="text-accent" />
                  </div>
                  <div>
                    <div className="flex items-center gap-3 mb-1">
                      <span className="text-lg font-bold text-text">:{srv.port}</span>
                      <span className="px-2 py-0.5 rounded text-xs font-medium bg-panel-3 text-text-dim border border-border">
                        {srv.protocol.toUpperCase()}
                      </span>
                      {srv.framework && (
                        <span className="px-2 py-0.5 rounded text-xs font-medium bg-accent/20 text-accent border border-accent/30">
                          {srv.framework}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-4 text-sm text-text-dim">
                      <span className="font-[Geist_Mono]">{srv.process || 'unknown'}</span>
                      {srv.pid > 0 && <span>PID {srv.pid}</span>}
                    </div>
                  </div>
                </div>
                {healthBadge(srv.health)}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Environment Tab
// ══════════════════════════════════════════════

function EnvironmentTab() {
  const { call } = useBackend()
  const [refreshKey, setRefreshKey] = useState(0)

  const { data, isLoading } = useQuery<EnvironmentInfo>({
    queryKey: ['devops-environment', refreshKey],
    queryFn: async () => {
      const res = await call('DevOps.GetEnvironment')
      return (res as EnvironmentInfo) || { path_dirs: [], key_vars: [], sdks: [], package_managers: [] }
    },
  })

  const truncatePath = (p: string, maxLen = 70) => (p.length > maxLen ? p.slice(0, maxLen - 3) + '...' : p)

  return (
    <div className="flex flex-col h-full p-8 space-y-6 overflow-y-auto">
      <div className="flex items-center gap-6">
        <h2 className="text-xl font-bold text-text flex items-center gap-2">
          <Variable size={22} className="text-accent" />
          System Environment
        </h2>
        <div className="flex-1" />
        <button
          onClick={() => setRefreshKey(k => k + 1)}
          className="p-3.5 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md active:rotate-180 duration-500"
        >
          <RefreshCw size={24} />
        </button>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center gap-3 py-16">
          <RefreshCw size={20} className="text-[var(--color-text-faint)] animate-spin" />
          <p className="text-sm font-semibold text-[var(--color-text-dim)]">Loading environment...</p>
        </div>
      ) : (
        <div className="space-y-8">
          {/* PATH Directories */}
          <section>
            <h3 className="text-lg font-bold text-text mb-3 flex items-center gap-2">
              <Folder size={18} className="text-accent" />
              PATH Directories
              <span className="text-sm font-normal text-text-faint">({(data?.path_dirs ?? []).length})</span>
            </h3>
            <div className="bg-panel border border-border rounded-xl max-h-64 overflow-y-auto">
              {(data?.path_dirs ?? []).map((dir, i) => (
                <div key={i} className={cn('px-5 py-2.5 text-sm font-[Geist_Mono] text-text-dim hover:bg-[var(--color-sidebar-hover)] transition-colors', i > 0 && 'border-t border-border/20')} title={dir}>
                  {truncatePath(dir)}
                </div>
              ))}
            </div>
          </section>

          {/* Key Variables */}
          <section>
            <h3 className="text-lg font-bold text-text mb-3 flex items-center gap-2">
              <Lock size={18} className="text-warning" />
              Key Variables
              <span className="text-sm font-normal text-text-faint">({(data?.key_vars ?? []).length})</span>
            </h3>
            <div className="bg-panel border border-border rounded-xl overflow-hidden">
              <table className="w-full text-left border-collapse">
                <thead className="bg-panel-2 border-b border-border">
                  <tr>
                    <th className="px-5 py-3 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Name</th>
                    <th className="px-5 py-3 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Value</th>
                  </tr>
                </thead>
                <tbody>
                  {(data?.key_vars ?? []).map((v) => (
                    <tr key={v.name} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-colors">
                      <td className="px-5 py-3 font-[Geist_Mono] text-sm text-accent font-medium">{v.name}</td>
                      <td className="px-5 py-3 font-[Geist_Mono] text-sm text-text-dim max-w-md truncate" title={v.value}>{truncatePath(v.value, 80)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>

          {/* SDKs & Package Managers side by side */}
          <div className="grid grid-cols-2 gap-6">
            <section>
              <h3 className="text-lg font-bold text-text mb-3 flex items-center gap-2">
                <Cpu size={18} className="text-success" />
                SDKs
              </h3>
              <div className="space-y-2">
                {(data?.sdks ?? []).length === 0 ? (
                  <p className="text-sm text-text-faint py-4">No SDKs detected</p>
                ) : (
                  (data?.sdks ?? []).map((sdk) => (
                    <div key={sdk.name} className="flex items-center justify-between bg-panel border border-border rounded-lg px-5 py-3">
                      <span className="text-sm font-bold text-text">{sdk.name}</span>
                      <span className="text-sm font-[Geist_Mono] text-success bg-success/10 px-2.5 py-0.5 rounded">{sdk.version}</span>
                    </div>
                  ))
                )}
              </div>
            </section>

            <section>
              <h3 className="text-lg font-bold text-text mb-3 flex items-center gap-2">
                <Package size={18} className="text-warning" />
                Package Managers
              </h3>
              <div className="space-y-2">
                {(data?.package_managers ?? []).length === 0 ? (
                  <p className="text-sm text-text-faint py-4">No package managers detected</p>
                ) : (
                  (data?.package_managers ?? []).map((pm) => (
                    <div key={pm.name} className="flex items-center justify-between bg-panel border border-border rounded-lg px-5 py-3">
                      <span className="text-sm font-bold text-text">{pm.name}</span>
                      <span className="text-sm font-[Geist_Mono] text-warning bg-warning/10 px-2.5 py-0.5 rounded">{pm.version}</span>
                    </div>
                  ))
                )}
              </div>
            </section>
          </div>
        </div>
      )}
    </div>
  )
}

// ══════════════════════════════════════════════
//  Log Explorer Tab
// ══════════════════════════════════════════════

function LogExplorerTab() {
  const { call } = useBackend()
  const [logPath, setLogPath] = useState('hawkward-gui.log')
  const [lines, setLines] = useState<string[]>([])
  const [searchPattern, setSearchPattern] = useState('')
  const [loading, setLoading] = useState(false)

  const handleTail = async () => {
    setLoading(true)
    try {
      const res = await call('DevOps.TailLog', logPath, 100)
      setLines((res as string[]) || [])
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = async () => {
    if (!searchPattern) return
    setLoading(true)
    try {
      const res = await call('DevOps.SearchLog', logPath, searchPattern)
      setLines((res as string[]) || [])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <div className="flex items-end gap-4 bg-panel border border-border p-6 rounded-2xl shadow-lg">
        <div className="flex-1 space-y-2">
          <label className="text-xs font-bold text-text-faint uppercase tracking-widest">Log File Path</label>
          <input
            type="text"
            value={logPath}
            onChange={(e) => setLogPath(e.target.value)}
            className="w-full bg-panel-2 border border-border rounded-lg px-4 py-2 text-sm text-text focus:outline-none focus:border-accent"
          />
        </div>
        <div className="flex-1 space-y-2">
          <label className="text-xs font-bold text-text-faint uppercase tracking-widest">Search Pattern</label>
          <div className="flex gap-2">
            <input
              type="text"
              value={searchPattern}
              onChange={(e) => setSearchPattern(e.target.value)}
              placeholder="Regex or text..."
              className="w-full bg-panel-2 border border-border rounded-lg px-4 py-2 text-sm text-text focus:outline-none focus:border-accent"
            />
            <button onClick={handleSearch} className="px-4 py-2 bg-accent text-white rounded-lg font-bold hover:opacity-90 transition-all shadow-md">
              Search
            </button>
          </div>
        </div>
        <button onClick={handleTail} className="px-6 py-2 bg-panel-3 border border-border rounded-lg font-bold text-text hover:bg-panel transition-all mb-0.5">
          Tail (Last 100)
        </button>
      </div>

      <div className="flex-1 bg-black/40 border border-border rounded-2xl p-6 font-[Geist_Mono] text-sm overflow-auto shadow-inner">
        {loading ? (
          <div className="flex items-center justify-center h-full gap-3 text-text-faint">
            <RefreshCw size={20} className="animate-spin" />
            <span className="font-bold uppercase tracking-widest">Accessing File...</span>
          </div>
        ) : lines.length === 0 ? (
          <div className="flex items-center justify-center h-full text-text-faint italic">
            No entries to display. Enter a path and tail or search.
          </div>
        ) : (
          <div className="space-y-1">
            {lines.map((line, i) => (
              <div key={i} className="whitespace-pre-wrap break-all text-text-dim border-b border-white/5 pb-1 last:border-0">{line}</div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
