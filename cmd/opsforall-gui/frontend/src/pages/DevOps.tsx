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
  RefreshCw,
  FileText,
  X,
  PlayCircle,
  StopCircle,
  Zap,
  Activity,
  Globe,
  TerminalSquare,
  GitBranch,
  Box,
  Wrench,
  Container,
  Variable,
  Lightbulb,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Tabs from '@radix-ui/react-tabs'
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { nanoid } from 'nanoid'
import type { CommandResult, ServiceEntry, FileEntry, ToolInfo, ContainerSummary, GitSummary, LocalServer, EnvironmentInfo, DevOpsSuggestion, DockerStatus, KubernetesStatus, ServiceGroupSummary } from '@/types'

// Strip ANSI escape sequences from terminal output
function stripAnsi(text: string): string {
  // eslint-disable-next-line no-control-regex
  return text.replace(/\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g, '')
}

type TabId = 'overview' | 'terminal' | 'powershell-pro' | 'services' | 'file-browser' | 'toolbox' | 'containers' | 'git' | 'servers' | 'environment' | 'log-explorer' | 'ai-suggestions'

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
          Unified control center for shell access, diagnostics, and infrastructure state.
        </p>
      </div>

      <Tabs.Root defaultValue="terminal" onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-border bg-panel px-4 overflow-x-auto no-scrollbar">
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
                'flex items-center gap-3 px-6 py-4 text-base font-bold transition-all border-b-2 border-transparent whitespace-nowrap',
                activeTab === tab.id ? 'border-accent text-text bg-accent/5' : 'text-text-faint hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}
            >
              {tab.icon}
              {tab.label}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="overview" className="h-full"><OverviewTab /></Tabs.Content>
          <Tabs.Content value="terminal" className="h-full"><TerminalTab /></Tabs.Content>
          <Tabs.Content value="powershell-pro" className="h-full"><PowerShellProTab /></Tabs.Content>
          <Tabs.Content value="services" className="h-full"><ServicesTab /></Tabs.Content>
          <Tabs.Content value="file-browser" className="h-full"><FileBrowserTab /></Tabs.Content>
          <Tabs.Content value="toolbox" className="h-full"><ToolboxTab /></Tabs.Content>
          <Tabs.Content value="containers" className="h-full"><ContainersTab /></Tabs.Content>
          <Tabs.Content value="git" className="h-full"><GitTab /></Tabs.Content>
          <Tabs.Content value="servers" className="h-full"><ServersTab /></Tabs.Content>
          <Tabs.Content value="ai-suggestions" className="h-full"><AISuggestionsTab /></Tabs.Content>
          <Tabs.Content value="environment" className="h-full"><EnvironmentTab /></Tabs.Content>
          <Tabs.Content value="log-explorer" className="h-full"><LogExplorerTab /></Tabs.Content>
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

  const { data: dockerData } = useQuery<DockerStatus>({
    queryKey: ['devops-docker-status'],
    queryFn: async () => {
      try { return await call('DevOps.GetDockerStatus') as DockerStatus }
      catch { return { installed: false, running: false, version: '', containers: { running: 0, stopped: 0, failed: 0, total: 0, containers: [] } } }
    },
    refetchInterval: refreshInterval,
  })

  const { data: k8sData } = useQuery<KubernetesStatus>({
    queryKey: ['devops-k8s-status'],
    queryFn: async () => {
      try { return await call('DevOps.GetKubernetesStatus') as KubernetesStatus }
      catch { return { installed: false, connected: false, cluster: '', nodes: 0, pods: 0 } }
    },
    refetchInterval: refreshInterval,
  })

  const { data: summary } = useQuery<ServiceGroupSummary>({
    queryKey: ['devops-service-summary'],
    queryFn: async () => {
      try { return await call('DevOps.GetServiceGroupSummary') as ServiceGroupSummary }
      catch { return { databases: 0, messageQueues: 0, webServers: 0, containers: 0, other: 0, running: 0, stopped: 0 } }
    },
    refetchInterval: refreshInterval,
  })

  const summaryKpis = [
    { label: 'Databases', value: summary?.databases ?? 0, colorClass: 'text-accent' },
    { label: 'Queues', value: summary?.messageQueues ?? 0, colorClass: 'text-accent-2' },
    { label: 'Web Servers', value: summary?.webServers ?? 0, colorClass: 'text-success' },
    { label: 'Running', value: summary?.running ?? 0, colorClass: 'text-success' },
    { label: 'Stopped', value: summary?.stopped ?? 0, colorClass: 'text-danger' },
  ]

  return (
    <div className="flex flex-col h-full p-8 space-y-8 overflow-y-auto">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <Box size={28} className="text-accent" />
              <h3 className="text-xl font-bold text-text uppercase tracking-widest">Docker</h3>
            </div>
            <span className={cn('text-xs font-bold px-3 py-1 rounded-full uppercase tracking-widest border',
              dockerData?.installed ? (dockerData?.running ? 'bg-success/20 text-success border-success/30' : 'bg-warning/20 text-warning border-warning/30') : 'bg-text-faint/20 text-text-faint border-text-faint/30')}>
              {dockerData?.installed ? (dockerData?.running ? 'Running' : 'Installed') : 'Not Installed'}
            </span>
          </div>
          {dockerData?.installed ? (
            <div className="space-y-6">
              {dockerData.version && <p className="text-sm text-text-dim"><span className="text-text-faint">Version:</span> <span className="font-mono text-text font-medium">{dockerData.version}</span></p>}
              <div className="grid grid-cols-4 gap-4">
                {[{ label: 'Running', value: dockerData.containers?.running ?? 0, dot: 'bg-success', cls: 'text-success' },
                  { label: 'Stopped', value: dockerData.containers?.stopped ?? 0, dot: 'bg-warning', cls: 'text-warning' },
                  { label: 'Failed', value: dockerData.containers?.failed ?? 0, dot: 'bg-danger', cls: 'text-danger' },
                  { label: 'Total', value: dockerData.containers?.total ?? 0, dot: 'bg-text-faint', cls: 'text-text-dim' }].map(item => (
                    <div key={item.label} className="flex flex-col items-center gap-2 bg-panel-2 border border-border rounded-xl p-4">
                      <span className={cn('w-2.5 h-2.5 rounded-full', item.dot)} />
                      <span className={cn('text-2xl font-bold tabular-nums', item.cls)}>{item.value}</span>
                      <span className="text-[10px] font-semibold text-text-faint uppercase tracking-wider">{item.label}</span>
                    </div>
                ))}
              </div>
            </div>
          ) : <p className="text-text-faint text-sm">Docker binary not detected in PATH.</p>}
        </div>

        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <Server size={28} className="text-info" />
              <h3 className="text-xl font-bold text-text uppercase tracking-widest">Kubernetes</h3>
            </div>
            <span className={cn('text-xs font-bold px-3 py-1 rounded-full uppercase tracking-widest border',
              k8sData?.installed ? (k8sData?.connected ? 'bg-success/20 text-success border-success/30' : 'bg-warning/20 text-warning border-warning/30') : 'bg-text-faint/20 text-text-faint border-text-faint/30')}>
              {k8sData?.installed ? (k8sData?.connected ? 'Connected' : 'Not Connected') : 'Not Installed'}
            </span>
          </div>
          {k8sData?.installed ? (
            <div className="space-y-6">
              <p className="text-sm text-text-dim"><span className="text-text-faint">Cluster:</span> <span className="font-mono text-text font-medium">{k8sData.cluster || 'N/A'}</span></p>
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
          ) : <p className="text-text-faint text-sm">kubectl not detected in PATH.</p>}
        </div>
      </div>

      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-2xl">
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
          {summaryKpis.map((kpi) => (
            <div key={kpi.label} className="flex flex-col items-center gap-2 bg-panel-2 border border-border rounded-xl p-5">
              <span className={cn('text-3xl font-bold tabular-nums', kpi.colorClass)}>{kpi.value}</span>
              <span className="text-[10px] font-semibold text-text-faint uppercase tracking-wider">{kpi.label}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function TerminalTab() {
  const { call } = useBackend()
  const [input, setInput] = useState('')
  const [output, setOutput] = useState<string[]>([`OpsForAll Unified Terminal v1.1 — Integrated Live Output\nType a command and press Enter to begin.\n`])
  const [history, setHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [isRunning, setIsRunning] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingCmd, setPendingCmd] = useState('')
  const [currentCmdId, setCurrentCmdId] = useState<string | null>(null)
  const outputRef = useRef<HTMLDivElement>(null)

  useEvents('DevOps.EventCmdLine', (data) => {
    if (data.id === currentCmdId) {
      setOutput(prev => [...prev, data.line])
    }
  })

  useEvents('DevOps.EventCmdDone', (id) => {
    if (id === currentCmdId) {
      setIsRunning(false)
      setCurrentCmdId(null)
    }
  })

  const runCommand = useCallback(async (cmd: string) => {
    if (!cmd.trim() || isRunning) return
    const id = nanoid()
    setIsRunning(true)
    setCurrentCmdId(id)
    setHistory(prev => [...prev, cmd])
    setOutput(prev => [...prev, `$ ${cmd}`])

    try {
      const res = await call('DevOps.RunCommandLive', cmd, id) as CommandResult
      if (res.error) {
        setOutput(prev => [...prev, `\u001b[31mError: ${res.error}\u001b[0m`])
        setIsRunning(false)
      } else if (res.output) {
        setOutput(prev => [...prev, res.output])
      }
    } catch (err) {
      setOutput(prev => [...prev, `\u001b[31mExecution Error: ${String(err)}\u001b[0m`])
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
          setPendingCmd(input); setConfirmOpen(true)
        } else {
          runCommand(input); setInput(''); setHistoryIndex(-1)
        }
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (history.length > 0) {
        const newIdx = historyIndex === -1 ? history.length - 1 : Math.max(historyIndex - 1, 0)
        setHistoryIndex(newIdx); setInput(history[newIdx])
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIndex >= 0 && historyIndex < history.length - 1) {
        const newIdx = historyIndex + 1; setHistoryIndex(newIdx); setInput(history[newIdx])
      } else {
        setHistoryIndex(-1); setInput('')
      }
    }
  }

  useEffect(() => {
    if (outputRef.current) outputRef.current.scrollTop = outputRef.current.scrollHeight
  }, [output])

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <ConfirmDialog open={confirmOpen} title="Impactful Command Detected"
        description={`Run command: "${pendingCmd}"?`}
        type="danger" confirmText="Execute"
        onConfirm={() => { runCommand(pendingCmd); setInput(''); setPendingCmd('') }}
        onClose={() => setConfirmOpen(false)} />

      <div className="flex items-center gap-4">
        <div className="relative flex-1">
          <input type="text" value={input} onChange={(e) => setInput(e.target.value)} onKeyDown={handleKeyDown}
            placeholder="Enter shell command..."
            className="w-full bg-[var(--color-bg)] border border-border rounded-xl pl-12 pr-4 py-4 text-lg font-mono text-text focus:outline-none focus:border-accent shadow-inner"
            disabled={isRunning} />
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-success text-xl font-bold font-mono">$</span>
        </div>
        <button onClick={() => { if (input.trim()) runCommand(input) }} disabled={isRunning || !input.trim()}
          className="flex items-center gap-2 px-8 py-4 text-lg font-bold bg-accent text-white rounded-xl hover:bg-accent/90 disabled:opacity-50 transition-all shadow-lg active:scale-95">
          <Play size={20} /> Execute
        </button>
        <button onClick={() => setOutput([`Output cleared.\n`])} className="px-6 py-4 text-lg font-bold text-text-dim border border-border rounded-xl hover:bg-panel-3 transition-all">
          <Trash2 size={20} /> Clear
        </button>
      </div>

      <div ref={outputRef} className="flex-1 bg-black border border-border rounded-2xl p-8 overflow-y-auto font-mono text-lg leading-relaxed whitespace-pre-wrap shadow-inner">
        {output.map((block, i) => <div key={i} className="mb-1">{stripAnsi(block)}</div>)}
        {isRunning && <div className="flex items-center gap-2 mt-2"><span className="inline-block w-3 h-6 bg-success animate-pulse" /><span className="text-sm font-bold text-success uppercase tracking-widest animate-pulse">Running...</span></div>}
      </div>
    </div>
  )
}

function PowerShellProTab() {
  const { call } = useBackend()
  const [isRunning, setIsRunning] = useState(false)
  const [output, setOutput] = useState('')
  const [selectedWorkflow, setSelectedWorkflow] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)

  const { data: workflows = [] } = useQuery<string[]>({
    queryKey: ['devops-workflows'],
    queryFn: async () => (await call('DevOps.GetPowerShellWorkflows') as string[]) || []
  })

  const runWorkflow = async (name: string) => {
    setIsRunning(true); setOutput(`Running workflow: ${name}...\n`)
    try {
      const res = await call('DevOps.RunPowerShell', name) as CommandResult
      if (res.error) setOutput(prev => prev + `\u001b[31mError: ${res.error}\u001b[0m`)
      else setOutput(prev => prev + res.output)
    } catch (err) { setOutput(prev => prev + `Execution Error: ${String(err)}`) }
    finally { setIsRunning(false) }
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 h-full p-8 gap-8 overflow-hidden">
      <ConfirmDialog open={confirmOpen} title="Execute PowerShell Workflow" description={`Run diagnostic workflow "${selectedWorkflow}"?`}
        onConfirm={() => runWorkflow(selectedWorkflow)} onClose={() => setConfirmOpen(false)} />
      <div className="lg:col-span-1 space-y-4 overflow-y-auto pr-2">
        <h3 className="text-lg font-bold text-text-dim uppercase tracking-widest mb-4">Diagnostic Workflows</h3>
        {workflows.map(wf => (
          <button key={wf} onClick={() => { setSelectedWorkflow(wf); setConfirmOpen(true) }} disabled={isRunning}
            className="w-full text-left bg-panel border border-border rounded-2xl p-6 transition-all hover:border-accent/50 hover:bg-accent/5 group disabled:opacity-50">
            <div className="flex items-center gap-4 mb-3">
              <div className="w-12 h-12 rounded-xl bg-panel-3 flex items-center justify-center text-text-dim group-hover:text-accent border border-border"><Zap size={24} /></div>
              <span className="text-xl font-bold text-text group-hover:text-accent transition-colors">{wf.replace('Invoke-Hawk', '')}</span>
            </div>
          </button>
        ))}
      </div>
      <div className="lg:col-span-2 flex flex-col space-y-4">
        <div className="flex-1 bg-black border border-border rounded-2xl p-8 overflow-y-auto font-mono text-lg leading-relaxed whitespace-pre shadow-inner">
          {stripAnsi(output) || 'Select a workflow to begin diagnostic execution.'}
        </div>
        <button onClick={() => setOutput('')} className="self-end px-6 py-3 text-lg font-bold text-text-dim border border-border rounded-xl hover:bg-panel-3 transition-all">Clear</button>
      </div>
    </div>
  )
}

function ServicesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [search, setSearch] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ name: string; action: string } | null>(null)
  const queryClient = useQueryClient()

  const { data: services = [] } = useQuery<ServiceEntry[]>({
    queryKey: ['devops-services'],
    queryFn: async () => (await call('DevOps.GetServices') as ServiceEntry[]) || [],
    refetchInterval: refreshInterval,
  })

  const controlService = async (name: string, action: string) => {
    const success = await call('DevOps.ControlService', name, action)
    if (success) queryClient.invalidateQueries({ queryKey: ['devops-services'] })
    setPendingAction(null)
  }

  const filtered = services.filter(s => s.name.toLowerCase().includes(search.toLowerCase()) || s.display_name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <ConfirmDialog open={confirmOpen} title="Service Control Permission" description={`Perform ${pendingAction?.action} on "${pendingAction?.name}"?`}
        type={pendingAction?.action === 'stop' ? 'danger' : 'warning'} confirmText="Execute"
        onConfirm={() => pendingAction && controlService(pendingAction.name, pendingAction.action)} onClose={() => setConfirmOpen(false)} />

      <div className="flex items-center gap-6 flex-wrap">
        <div className="bg-panel border border-border px-6 py-3 rounded-2xl flex gap-8">
          <span className="text-lg font-bold text-success">{services.filter(s => s.status === 'Running').length} Running</span>
          <span className="text-lg font-bold text-danger">{services.filter(s => s.status === 'Stopped').length} Stopped</span>
        </div>
        <div className="relative flex-1 max-w-sm ml-auto">
          <Search size={20} className="absolute left-4 top-1/2 -translate-y-1/2 text-text-faint" />
          <input type="text" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search services..."
            className="w-full bg-panel border border-border rounded-xl pl-12 pr-4 py-3 text-lg focus:outline-none focus:border-accent shadow-inner" />
        </div>
      </div>

      <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl overflow-hidden shadow-inner">
        <div className="overflow-y-auto h-full">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-6 py-4 text-xs font-semibold text-text-dim uppercase">Name</th>
                <th className="px-6 py-4 text-xs font-semibold text-text-dim uppercase">Status</th>
                <th className="px-6 py-4 text-xs font-semibold text-text-dim uppercase text-right">Control</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(svc => (
                <tr key={svc.name} className="border-b border-border/20 hover:bg-panel-3 group transition-colors">
                  <td className="px-6 py-4"><p className="text-sm font-bold text-text">{svc.display_name}</p><p className="text-xs font-mono text-text-faint">{svc.name}</p></td>
                  <td className="px-6 py-4"><StatusBadge status={svc.status} /></td>
                  <td className="px-6 py-4 text-right opacity-0 group-hover:opacity-100 transition-opacity flex justify-end gap-2">
                    <button onClick={() => { setPendingAction({ name: svc.name, action: 'start' }); setConfirmOpen(true) }} className="p-2 text-success hover:bg-success/10 rounded-lg"><PlayCircle size={20} /></button>
                    <button onClick={() => { setPendingAction({ name: svc.name, action: 'stop' }); setConfirmOpen(true) }} className="p-2 text-danger hover:bg-danger/10 rounded-lg"><StopCircle size={20} /></button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

function FileBrowserTab() {
  const { call } = useBackend()
  const [currentPath, setCurrentPath] = useState('')
  const [entries, setEntries] = useState<FileEntry[]>([])
  const [history, setHistory] = useState<string[]>([])
  const [, setLoading] = useState(false)
  const [previewFile, setPreviewFile] = useState<FileEntry | null>(null)
  const [fileContent, setPreviewContent] = useState('')

  const fetchDir = useCallback(async (path: string) => {
    setLoading(true)
    try {
      const res = await call('DevOps.ListDirectory', path)
      setEntries((res as FileEntry[]) || [])
      setCurrentPath(path)
    } catch (err) { console.error(err) } finally { setLoading(false) }
  }, [call])

  useEffect(() => {
    if (!currentPath) call('DevOps.GetDefaultPath').then(p => { if (p) { setCurrentPath(p as string); fetchDir(p as string) } })
    else fetchDir(currentPath)
  }, [fetchDir, currentPath, call])

  const openFile = async (file: FileEntry) => {
    if (file.is_dir) {
      setHistory(prev => [...prev, currentPath]); fetchDir(file.path); setPreviewFile(null)
    } else {
      setPreviewFile(file); setPreviewContent(await call('DevOps.ReadFile', file.path) as string)
    }
  }

  return (
    <div className="flex h-full p-8 gap-8">
      <div className={cn('flex flex-col flex-1 min-w-0 space-y-4', previewFile ? 'w-1/2' : 'w-full')}>
        <div className="flex items-center gap-4 bg-panel border border-border rounded-xl px-4 py-2">
          <button onClick={() => { const prev = history.pop(); if (prev) fetchDir(prev) }} disabled={history.length === 0} className="p-2 hover:bg-panel-3 rounded-lg"><ChevronRight size={20} className="rotate-180" /></button>
          <span className="text-sm font-mono truncate">{currentPath}</span>
        </div>
        <div className="flex-1 bg-panel border border-border rounded-2xl overflow-y-auto">
          <table className="w-full text-left">
            <tbody>
              {entries.map(file => (
                <tr key={file.path} onClick={() => openFile(file)} className="border-b border-border/10 hover:bg-panel-3 cursor-pointer group">
                  <td className="px-6 py-3 flex items-center gap-3">
                    {file.is_dir ? <Folder size={18} className="text-accent" /> : <FileText size={18} className="text-text-faint" />}
                    <span className="text-sm font-medium group-hover:text-accent">{file.name}</span>
                  </td>
                  <td className="px-6 py-3 text-right text-xs font-mono text-text-faint">{file.size}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {previewFile && (
        <div className="w-1/2 flex flex-col space-y-4">
          <div className="flex items-center justify-between"><h3 className="font-bold truncate">{previewFile.name}</h3><button onClick={() => setPreviewFile(null)}><X /></button></div>
          <div className="flex-1 bg-black border border-border rounded-2xl p-6 overflow-auto font-mono text-sm">
            {previewFile.is_binary ? 'Binary content hidden' : fileContent}
          </div>
        </div>
      )}
    </div>
  )
}

function LogExplorerTab() {
  const { call } = useBackend()
  const [logPath, setLogPath] = useState('opsforall-gui.log')
  const [lines, setLines] = useState<string[]>([])
  const [loading, setLoading] = useState(false)

  const handleTail = async () => {
    setLoading(true); try { setLines(await call('DevOps.TailLog', logPath, 100) as string[]) } finally { setLoading(false) }
  }

  return (
    <div className="flex flex-col h-full p-8 space-y-6">
      <div className="flex gap-4 bg-panel border border-border p-6 rounded-2xl">
        <input type="text" value={logPath} onChange={e => setLogPath(e.target.value)} className="flex-1 bg-panel-2 border border-border rounded-lg px-4 py-2 text-sm" />
        <button onClick={handleTail} className="px-6 py-2 bg-accent text-white rounded-lg font-bold">Tail</button>
      </div>
      <div className="flex-1 bg-black border border-border rounded-2xl p-6 font-mono text-sm overflow-auto">
        {loading ? <RefreshCw className="animate-spin" /> : lines.map((l, i) => <div key={i} className="mb-1 text-text-dim">{l}</div>)}
      </div>
    </div>
  )
}

function AISuggestionsTab() {
  const { call } = useBackend()
  const { data: suggestions = [] } = useQuery<DevOpsSuggestion[]>({
    queryKey: ['devops-ai-suggestions'],
    queryFn: async () => (await call('DevOps.GetAISuggestions') as DevOpsSuggestion[]) ?? []
  })
  return (
    <div className="p-8 space-y-4 overflow-y-auto h-full">
      {suggestions.map((s, i) => (
        <div key={i} className={cn('p-6 rounded-xl border bg-panel', s.severity === 'critical' ? 'border-danger/30 bg-danger/5' : 'border-border')}>
          <h4 className="font-bold text-text mb-1 uppercase tracking-tight">{s.category}: {s.message}</h4>
          <p className="text-sm text-text-dim">{s.action}</p>
        </div>
      ))}
    </div>
  )
}

function ToolboxTab() {
  const { call } = useBackend()
  const { data: tools = [] } = useQuery<ToolInfo[]>({ queryKey: ['devops-tools'], queryFn: async () => (await call('DevOps.GetInstalledTools') as ToolInfo[]) || [] })
  return (
    <div className="p-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 overflow-y-auto h-full">
      {tools.map(t => (
        <div key={t.name} className="p-5 rounded-xl border border-border bg-panel flex justify-between items-center">
          <div><p className="font-bold">{t.name}</p><p className="text-xs font-mono text-text-faint">{t.version || 'not found'}</p></div>
          <StatusBadge status={t.status} />
        </div>
      ))}
    </div>
  )
}

function ContainersTab() {
  const { call } = useBackend()
  const { data } = useQuery<ContainerSummary>({ queryKey: ['devops-containers'], queryFn: async () => (await call('DevOps.GetContainers') as ContainerSummary) || { containers: [] } })
  return (
    <div className="p-8 overflow-y-auto h-full space-y-4">
      {(data?.containers ?? []).map(c => (
        <div key={c.id} className="p-5 rounded-xl border border-border bg-panel flex justify-between">
          <div><p className="font-bold text-accent">{c.name}</p><p className="text-xs text-text-dim">{c.image}</p></div>
          <StatusBadge status={c.state} />
        </div>
      ))}
    </div>
  )
}

function GitTab() {
  const { call } = useBackend()
  const { data } = useQuery<GitSummary>({ queryKey: ['devops-git'], queryFn: async () => (await call('DevOps.GetGitSummary') as GitSummary) || { repositories: [] } })
  return (
    <div className="p-8 space-y-4 overflow-y-auto h-full">
      {(data?.repositories ?? []).map(r => (
        <div key={r.path} className="p-5 rounded-xl border border-border bg-panel flex justify-between">
          <p className="font-bold text-sm truncate">{r.path.split(/[\\/]/).pop()}</p>
          <div className="flex gap-4 text-xs font-mono"><span>{r.branch}</span><span className={r.clean ? 'text-success' : 'text-warning'}>{r.clean ? 'Clean' : 'Dirty'}</span></div>
        </div>
      ))}
    </div>
  )
}

function ServersTab() {
  const { call } = useBackend()
  const { data: servers = [] } = useQuery<LocalServer[]>({ queryKey: ['devops-servers'], queryFn: async () => (await call('DevOps.GetLocalServers') as LocalServer[]) || [] })
  return (
    <div className="p-8 space-y-4 overflow-y-auto h-full">
      {servers.map(s => (
        <div key={`${s.port}-${s.pid}`} className="p-5 rounded-xl border border-border bg-panel flex justify-between items-center">
          <div><p className="font-bold text-lg">:{s.port}</p><p className="text-xs font-mono text-text-faint">{s.process}</p></div>
          <span className="text-xs px-2 py-1 bg-accent/20 text-accent rounded uppercase font-bold">{s.framework}</span>
        </div>
      ))}
    </div>
  )
}

function EnvironmentTab() {
  const { call } = useBackend()
  const { data: env } = useQuery<EnvironmentInfo>({ queryKey: ['devops-env'], queryFn: async () => (await call('DevOps.GetEnvironment') as EnvironmentInfo) })
  return (
    <div className="p-8 space-y-8 overflow-y-auto h-full">
      <section><h3 className="font-bold mb-4">Key Variables</h3><div className="space-y-1">{env?.key_vars.map(v => <div key={v.name} className="flex gap-4 text-sm font-mono"><span className="text-accent">{v.name}</span><span className="text-text-dim truncate">{v.value}</span></div>)}</div></section>
      <section><h3 className="font-bold mb-4">SDKs</h3><div className="flex gap-2 flex-wrap">{env?.sdks.map(s => <span key={s.name} className="px-3 py-1 bg-panel border border-border rounded-full text-xs font-mono">{s.name} {s.version}</span>)}</div></section>
    </div>
  )
}
