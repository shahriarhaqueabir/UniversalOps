import { useState, useRef, useCallback, useEffect } from 'react'
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
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Tabs from '@radix-ui/react-tabs'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'

// Strip ANSI escape sequences from terminal output
function stripAnsi(text: string): string {
  // Matches ANSI escape sequences: ESC[<params>m, ESC[<params>K, ESC[<params>J, etc.
  // eslint-disable-next-line no-control-regex
  return text.replace(/\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g, '')
}
import type { CommandResult, ServiceEntry, FileEntry } from '@/types'

type TabId = 'terminal' | 'powershell-pro' | 'services' | 'file-browser'

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
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const t = setTimeout(() => setLoading(false), 50)
    return () => clearTimeout(t)
  }, [])

  if (loading) {
    return (
      <div className="p-6 space-y-4 animate-pulse">
        <div className="h-8 w-48 bg-panel-2 rounded" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="h-32 bg-panel-2 rounded" />
          <div className="h-32 bg-panel-2 rounded" />
        </div>
      </div>
    )
  }

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
            { id: 'terminal', label: 'Interactive Terminal', icon: <Terminal size={20} /> },
            { id: 'powershell-pro', label: 'PowerShell Pro', icon: <Zap size={20} className="text-warning" /> },
            { id: 'services', label: 'System Services', icon: <Server size={20} /> },
            { id: 'file-browser', label: 'File Explorer', icon: <Folder size={20} /> },
          ].map((tab) => (
            <Tabs.Trigger
              key={tab.id}
              value={tab.id}
              className={cn(
                'flex items-center gap-3 px-6 py-4 text-base font-bold transition-all border-b-2 border-transparent',
                activeTab === tab.id ? 'border-accent text-text bg-accent/5' : 'text-text-faint hover:text-text hover:bg-white/5',
              )}
            >
              {tab.icon}
              {tab.label}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
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
        </div>
      </Tabs.Root>
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
            className="w-full bg-[var(--color-bg)] border border-border rounded-xl pl-12 pr-4 py-4 text-lg font-[JetBrains_Mono] text-text placeholder-text-faint focus:outline-none focus:border-primary shadow-inner"
            disabled={isRunning}
          />
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-success text-xl font-bold font-[JetBrains_Mono]">$</span>
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
        className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl p-8 overflow-y-auto font-[JetBrains_Mono] text-lg leading-relaxed whitespace-pre-wrap shadow-inner"
      >
        {output.map((block, i) => (
          <div key={i} className="whitespace-pre-wrap break-all mb-2">
            {stripAnsi(block)}
          </div>
        ))}
        {isRunning && (
          <div className="flex items-center gap-2 mt-2">
            <span className="inline-block w-3 h-6 bg-success animate-pulse" />
            <span className="text-sm font-bold text-success uppercase tracking-widest animate-pulse">Running...</span>
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
  const [workflows, setWorkflows] = useState<string[]>([])
  const [isRunning, setIsRunning] = useState(false)
  const [output, setOutput] = useState('')
  const [selectedWorkflow, setSelectedWorkflow] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)

  useEffect(() => {
    call('DevOps.GetPowerShellWorkflows').then(res => setWorkflows(res as string[]))
  }, [call])

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
        <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl p-8 overflow-y-auto font-[JetBrains_Mono] text-lg leading-relaxed whitespace-pre shadow-inner">
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
  const [services, setServices] = useState<ServiceEntry[]>([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ name: string; action: string } | null>(null)

  const fetchServices = useCallback(async () => {
    setLoading(true)
    try {
      const res = await call('DevOps.GetServices')
      setServices((res as ServiceEntry[]) || [])
    } catch (err) {
      console.error(err)
      setServices([])
    } finally {
      setLoading(false)
    }
  }, [call])

  const controlService = async (name: string, action: string) => {
    const success = await call('DevOps.ControlService', name, action)
    if (success) {
      fetchServices()
    }
    setPendingAction(null)
  }

  useEffect(() => {
    fetchServices()
  }, [fetchServices])

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
          onClick={fetchServices}
          className="p-3.5 bg-panel border border-border rounded-xl hover:bg-panel-3 text-text-dim hover:text-text transition-all shadow-md active:rotate-180 duration-500"
        >
          <RefreshCw size={24} />
        </button>
      </div>

      <div className="flex-1 bg-[#0b1120] border border-border rounded-2xl overflow-hidden shadow-inner">
        <div className="overflow-y-auto h-full">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-8 py-5 text-sm font-bold text-text-dim uppercase tracking-widest">Name</th>
                <th className="px-8 py-5 text-sm font-bold text-text-dim uppercase tracking-widest">Display Name</th>
                <th className="px-8 py-5 text-sm font-bold text-text-dim uppercase tracking-widest">Status</th>
                <th className="px-8 py-5 text-sm font-bold text-text-dim uppercase tracking-widest">Startup</th>
                <th className="px-8 py-5 text-sm font-bold text-text-dim uppercase tracking-widest text-right">Control</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-8 py-20 text-center">
                    <div className="flex flex-col items-center gap-4 text-text-faint">
                      <RefreshCw size={48} className="animate-spin opacity-20" />
                      <p className="text-xl">Enumerating System Services...</p>
                    </div>
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-8 py-20 text-center text-text-faint text-xl italic">No services matching your criteria</td>
                </tr>
              ) : (
                filtered.map((svc) => (
                  <tr key={svc.name} className="border-b border-border/20 hover:bg-white/5 transition-colors group">
                    <td className="px-8 py-4 font-[JetBrains_Mono] text-base text-accent font-medium">{svc.name}</td>
                    <td className="px-8 py-4 text-lg text-text">{svc.display_name}</td>
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
      setEntries(res as FileEntry[])
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
              className="p-1.5 hover:bg-white/10 rounded-lg text-text-faint hover:text-accent transition-all shrink-0"
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
                      isLast ? "text-text cursor-default" : "text-text-dim hover:text-accent hover:bg-white/5"
                    )}
                  >
                    {part}
                  </button>
                </div>
              )
            })}
          </div>
        </div>

        <div className="flex-1 bg-[#0b1120] border border-border rounded-2xl overflow-hidden shadow-inner">
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
                      className="border-b border-border/20 hover:bg-white/5 cursor-pointer transition-colors group"
                    >
                      <td className="px-8 py-4 flex items-center gap-4">
                        {file.is_dir ? <Folder size={24} className="text-accent" /> : <FileText size={24} className="text-text-faint" />}
                        <span className="text-lg text-text group-hover:text-accent transition-colors">{file.name}</span>
                      </td>
                      <td className="px-8 py-4 text-right font-[JetBrains_Mono] text-base text-text-dim">{file.size}</td>
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
          <div className="flex-1 bg-[#0b1120] border border-border rounded-2xl p-8 overflow-y-auto shadow-inner relative group">
            {previewFile.is_binary ? (
              <div className="flex flex-col items-center justify-center h-full text-center space-y-4 opacity-50">
                <Lock size={48} className="text-warning" />
                <div>
                  <p className="text-xl font-bold text-text">Binary Data Encrypted</p>
                  <p className="text-base text-text-dim">Direct preview is disabled for safety.</p>
                </div>
              </div>
            ) : (
              <pre className="font-[JetBrains_Mono] text-base text-text leading-relaxed whitespace-pre-wrap">
                {fileContent || '// No readable content or file is empty.'}
              </pre>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
