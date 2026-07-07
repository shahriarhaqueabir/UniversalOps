import { useState, useRef, useCallback } from 'react'
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
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Tabs from '@radix-ui/react-tabs'

type TabId = 'terminal' | 'services' | 'file-browser'

// ── Inline helpers ──

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    running: 'bg-success/20 text-success',
    stopped: 'bg-danger/20 text-danger',
    auto: 'bg-primary/20 text-primary',
    manual: 'bg-warning/20 text-warning',
    disabled: 'bg-muted/20 text-muted',
  }
  return (
    <span className={cn('px-2 py-0.5 rounded text-xs font-medium', colors[status.toLowerCase()] || 'bg-muted/20 text-muted')}>
      {status}
    </span>
  )
}

// ── Mock data ──

interface ServiceInfo {
  name: string
  displayName: string
  status: 'running' | 'stopped'
  startType: 'auto' | 'manual' | 'disabled'
}

const mockServices: ServiceInfo[] = [
  { name: 'hawkward-api', displayName: 'Hawkward API Server', status: 'running', startType: 'auto' },
  { name: 'hawkward-worker', displayName: 'Hawkward Worker', status: 'running', startType: 'auto' },
  { name: 'postgresql', displayName: 'PostgreSQL Database', status: 'running', startType: 'auto' },
  { name: 'redis', displayName: 'Redis Cache', status: 'running', startType: 'auto' },
  { name: 'nginx', displayName: 'Nginx Reverse Proxy', status: 'running', startType: 'auto' },
  { name: 'prometheus', displayName: 'Prometheus Metrics', status: 'running', startType: 'auto' },
  { name: 'grafana', displayName: 'Grafana Dashboards', status: 'running', startType: 'auto' },
  { name: 'docker', displayName: 'Docker Engine', status: 'running', startType: 'auto' },
  { name: 'sshd', displayName: 'OpenSSH Server', status: 'stopped', startType: 'manual' },
  { name: 'cron', displayName: 'Cron Scheduler', status: 'running', startType: 'auto' },
  { name: 'syslog', displayName: 'System Logger', status: 'stopped', startType: 'disabled' },
]

interface FSNode {
  name: string
  type: 'file' | 'dir'
  size?: string
  modified: string
  children?: FSNode[]
  content?: string
}

function buildMockFileSystem(): FSNode[] {
  return [
    {
      name: 'src', type: 'dir', modified: '2026-07-07 10:30',
      children: [
        { name: 'main.go', type: 'file', size: '12.4 KB', modified: '2026-07-06 15:20', content: 'package main\n\nimport (\n\t"fmt"\n\t"log"\n)\n\nfunc main() {\n\tfmt.Println("Hello, Hawkward!")\n}\n' },
        { name: 'handler.go', type: 'file', size: '8.1 KB', modified: '2026-07-05 11:10', content: 'package main\n\nfunc handleRequest(req Request) Response {\n\treturn Response{Status: 200}\n}\n' },
        {
          name: 'utils', type: 'dir', modified: '2026-07-04 09:00',
          children: [
            { name: 'helpers.go', type: 'file', size: '3.2 KB', modified: '2026-07-03 16:45', content: 'package utils\n\nfunc Min(a, b int) int {\n\tif a < b { return a }\n\treturn b\n}\n' },
            { name: 'config.go', type: 'file', size: '5.7 KB', modified: '2026-07-02 14:30', content: 'package utils\n\ntype Config struct {\n\tPort int\n\tDebug bool\n}\n' },
          ],
        },
      ],
    },
    {
      name: 'cmd', type: 'dir', modified: '2026-07-07 09:00',
      children: [
        {
          name: 'hawkward', type: 'dir', modified: '2026-07-07 09:00',
          children: [
            { name: 'main.go', type: 'file', size: '2.1 KB', modified: '2026-07-07 09:00', content: 'package main\n\nfunc main() {\n\tapp := NewApp()\n\tapp.Run()\n}\n' },
          ],
        },
      ],
    },
    { name: 'go.mod', type: 'file', size: '156 B', modified: '2026-07-01 12:00', content: 'module github.com/hawkward\n\ngo 1.22\n' },
    { name: 'go.sum', type: 'file', size: '45.2 KB', modified: '2026-07-01 12:00', content: 'github.com/charmbracelet/bubbletea v0.26.0 h1:abc...\n' },
    { name: 'README.md', type: 'file', size: '3.1 KB', modified: '2026-07-07 08:30', content: '# Hawkward\n\nAll-in-one operations tool.\n' },
    { name: 'package.json', type: 'file', size: '890 B', modified: '2026-07-07 08:30', content: '{\n  "name": "hawkward",\n  "version": "0.1.0"\n}\n' },
    {
      name: 'scripts', type: 'dir', modified: '2026-07-06 16:00',
      children: [
        { name: 'build.sh', type: 'file', size: '420 B', modified: '2026-07-06 16:00', content: '#!/bin/bash\ngo build -o hawkward ./cmd/hawkward\n' },
        { name: 'test.sh', type: 'file', size: '180 B', modified: '2026-07-06 15:30', content: '#!/bin/bash\ngo test ./... -v\n' },
      ],
    },
  ]
}

// ══════════════════════════════════════════════
//  DevOps Page
// ══════════════════════════════════════════════

export function DevOps() {
  const [activeTab, setActiveTab] = useState<TabId>('terminal')

  return (
    <div className="flex h-full">
      <Tabs.Root value={activeTab} onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-border bg-background">
          <Tabs.Trigger
            value="terminal"
            className={cn(
              'flex items-center gap-2 px-4 py-3 text-sm border-b-2 border-transparent transition-colors',
              activeTab === 'terminal' ? 'border-primary text-text' : 'text-muted hover:text-text',
            )}
          >
            <Terminal size={16} />
            Terminal
          </Tabs.Trigger>
          <Tabs.Trigger
            value="services"
            className={cn(
              'flex items-center gap-2 px-4 py-3 text-sm border-b-2 border-transparent transition-colors',
              activeTab === 'services' ? 'border-primary text-text' : 'text-muted hover:text-text',
            )}
          >
            <Server size={16} />
            Services
          </Tabs.Trigger>
          <Tabs.Trigger
            value="file-browser"
            className={cn(
              'flex items-center gap-2 px-4 py-3 text-sm border-b-2 border-transparent transition-colors',
              activeTab === 'file-browser' ? 'border-primary text-text' : 'text-muted hover:text-text',
            )}
          >
            <Folder size={16} />
            File Browser
          </Tabs.Trigger>
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="terminal" className="h-full">
            <TerminalTab />
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

const mockOutputs: Record<string, string> = {
  docker: 'NAMES\t\t\tSTATUS\nhawkward-api\t\tUp 3 days\nhawkward-db\t\tUp 3 days\nprometheus\t\tUp 14 days\ngrafana\t\t\tUp 14 days',
  kubectl: 'NAME\t\t\t\t\tREADY\tSTATUS\tRESTARTS\tAGE\nhawkward-api-7d4f8c9b6-abc12\t1/1\tRunning\t0\t\t3d\nhawkward-worker-5e6f7a8b9-def34\t1/1\tRunning\t1\t\t3d',
  ping: 'PING 8.8.8.8 (8.8.8.8): 56 data bytes\n64 bytes from 8.8.8.8: icmp_seq=0 ttl=117 time=12.5 ms\n64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=11.8 ms\n64 bytes from 8.8.8.8: icmp_seq=2 ttl=117 time=13.2 ms\n\n--- 8.8.8.8 ping statistics ---\n4 packets transmitted, 4 received, 0% packet loss',
  build: '[1/5] Compiling core package... OK\n[2/5] Compiling sysops module... OK\n[3/5] Compiling netops module... OK\n[4/5] Compiling secops module... OK\n[5/5] Linking binary... OK\nBuild completed in 3.42s',
  error: 'Error: command not found: unknown-command\nExit code: 127',
  default: 'Command completed successfully (exit code: 0)',
}

function getSimulatedOutput(command: string): string {
  if (command.includes('docker')) return mockOutputs.docker
  if (command.includes('kubectl')) return mockOutputs.kubectl
  if (command.includes('ping')) return mockOutputs.ping
  if (command.includes('build') || command.includes('go build')) return mockOutputs.build
  if (Math.random() > 0.85 && command.length > 0) return mockOutputs.error
  return mockOutputs.default
}

function TerminalTab() {
  const [input, setInput] = useState('')
  const [output, setOutput] = useState<string[]>([`Hawkward Terminal v0.1\nType a command and press Enter to run.\n`])
  const [history, setHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [isRunning, setIsRunning] = useState(false)
  const [workingDir] = useState('/home/hawkward/projects/hawkward')
  const outputRef = useRef<HTMLDivElement>(null)

  const runCommand = useCallback((cmd: string) => {
    if (!cmd.trim() || isRunning) return
    setIsRunning(true)
    setHistory(prev => [...prev, cmd])
    setOutput(prev => [...prev, `$ ${cmd}`])

    setTimeout(() => {
      const result = getSimulatedOutput(cmd)
      setOutput(prev => [...prev, result])
      setIsRunning(false)
    }, 300 + Math.random() * 700)
  }, [isRunning])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (input.trim()) {
        runCommand(input)
        setInput('')
        setHistoryIndex(-1)
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

  const clearOutput = () => {
    setOutput([`Output cleared.\n`])
  }

  return (
    <div className="flex flex-col h-full p-4">
      {/* Working directory */}
      <div className="flex items-center gap-2 mb-3 text-xs text-muted">
        <Folder size={12} />
        <span className="font-[JetBrains_Mono]">{workingDir}</span>
      </div>

      {/* Command input */}
      <div className="flex items-center gap-2 mb-3">
        <div className="relative flex-1">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Enter command..."
            className="w-full bg-[#0b1120] border border-border rounded-lg pl-9 pr-3 py-2.5 text-sm font-[JetBrains_Mono] text-text placeholder-muted focus:outline-none focus:border-primary"
            disabled={isRunning}
          />
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-success text-sm font-[JetBrains_Mono]">$</span>
        </div>
        <button
          onClick={() => { if (input.trim()) runCommand(input) }}
          disabled={isRunning || !input.trim()}
          className="flex items-center gap-1.5 px-4 py-2.5 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <Play size={14} />
          Run
        </button>
        <button
          onClick={clearOutput}
          className="flex items-center gap-1.5 px-3 py-2.5 text-sm text-muted border border-border rounded-lg hover:bg-sidebar-hover transition-colors"
        >
          <Trash2 size={14} />
          Clear
        </button>
      </div>

      {/* Output */}
      <div
        ref={outputRef}
        className="flex-1 bg-[#0b1120] border border-border rounded-lg p-4 overflow-y-auto font-[JetBrains_Mono] text-sm leading-relaxed whitespace-pre-wrap"
        style={{ minHeight: 200 }}
      >
        {output.map((block, i) => (
          <div key={i} className="whitespace-pre-wrap break-all">
            {block}
          </div>
        ))}
        {isRunning && (
          <span className="inline-block w-2 h-4 bg-success animate-pulse ml-1" />
        )}
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Services Tab
// ══════════════════════════════════════════════

function ServicesTab() {
  const [services, setServices] = useState<ServiceInfo[]>(mockServices)
  const [search, setSearch] = useState('')

  const filtered = services.filter(s =>
    s.name.toLowerCase().includes(search.toLowerCase()) ||
    s.displayName.toLowerCase().includes(search.toLowerCase()),
  )

  const toggleService = (name: string, action: 'start' | 'stop') => {
    setServices(prev =>
      prev.map(s =>
        s.name === name ? { ...s, status: action === 'start' ? 'running' : 'stopped' } : s,
      ),
    )
  }

  const runningCount = services.filter(s => s.status === 'running').length
  const stoppedCount = services.filter(s => s.status === 'stopped').length

  return (
    <div className="flex flex-col h-full p-4">
      {/* Summary + Search */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <div className="flex items-center gap-3 text-xs text-muted">
          <span><span className="text-success font-medium">{runningCount}</span> running</span>
          <span><span className="text-danger font-medium">{stoppedCount}</span> stopped</span>
          <span className="text-border">|</span>
          <span>{services.length} total</span>
        </div>
        <div className="flex-1" />
        <div className="relative">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search services..."
            className="bg-[#0f172a] border border-border rounded-lg pl-8 pr-3 py-1.5 text-sm text-text placeholder-muted focus:outline-none focus:border-primary w-56"
          />
        </div>
      </div>

      {/* Table */}
      <div className="flex-1 bg-[#0b1120] border border-border rounded-lg overflow-y-auto min-h-0">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-xs text-muted">
              <th className="text-left px-4 py-2.5 font-medium">Name</th>
              <th className="text-left px-4 py-2.5 font-medium">Display Name</th>
              <th className="text-left px-4 py-2.5 font-medium">Status</th>
              <th className="text-left px-4 py-2.5 font-medium">Start Type</th>
              <th className="text-right px-4 py-2.5 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-muted text-sm">No services found</td>
              </tr>
            )}
            {filtered.map((svc) => (
              <tr key={svc.name} className="border-b border-border/30 hover:bg-white/5 transition-colors">
                <td className="px-4 py-2.5 font-[JetBrains_Mono] text-xs text-text">{svc.name}</td>
                <td className="px-4 py-2.5 text-text">{svc.displayName}</td>
                <td className="px-4 py-2.5"><StatusBadge status={svc.status} /></td>
                <td className="px-4 py-2.5"><StatusBadge status={svc.startType} /></td>
                <td className="px-4 py-2.5 text-right">
                  <div className="flex items-center justify-end gap-1">
                    {svc.status === 'stopped' ? (
                      <button
                        onClick={() => toggleService(svc.name, 'start')}
                        className="p-1.5 text-success hover:bg-success/10 rounded transition-colors"
                        title="Start"
                      >
                        <PlayCircle size={14} />
                      </button>
                    ) : (
                      <button
                        onClick={() => toggleService(svc.name, 'stop')}
                        className="p-1.5 text-danger hover:bg-danger/10 rounded transition-colors"
                        title="Stop"
                      >
                        <StopCircle size={14} />
                      </button>
                    )}
                    <button
                      onClick={() => {
                        toggleService(svc.name, 'stop')
                        setTimeout(() => toggleService(svc.name, 'start'), 300)
                      }}
                      className="p-1.5 text-warning hover:bg-warning/10 rounded transition-colors"
                      title="Restart"
                    >
                      <RotateCcw size={14} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  File Browser Tab
// ══════════════════════════════════════════════

interface PathSegment {
  name: string
  path: string
}

const mockPaths: string[] = [
  '/home/hawkward/projects/hawkward',
  '/var/log',
  '/etc/nginx',
  '/opt/hawkward',
]

function FileBrowserTab() {
  const [filesystem] = useState<FSNode[]>(buildMockFileSystem())
  const [currentPath, setCurrentPath] = useState<string[]>([])
  const [navHistory, setNavHistory] = useState<string[][]>([])
  const [previewFile, setPreviewFile] = useState<FSNode | null>(null)
  const [pathInput, setPathInput] = useState('')

  const getCurrentDir = (): FSNode[] => {
    let current: FSNode[] = filesystem
    for (const seg of currentPath) {
      const found = current.find(n => n.name === seg && n.type === 'dir')
      if (found && found.children) {
        current = found.children
      } else {
        return filesystem
      }
    }
    return current
  }

  const currentDir = getCurrentDir()

  const segments: PathSegment[] = [
    { name: 'root', path: '' },
    ...currentPath.map((seg, i) => ({ name: seg, path: currentPath.slice(0, i + 1).join('/') })),
  ]

  const navigateInto = (name: string) => {
    setNavHistory(prev => [...prev, currentPath])
    setCurrentPath(prev => [...prev, name])
    setPreviewFile(null)
  }

  const navigateTo = (index: number) => {
    setCurrentPath(index === 0 ? [] : currentPath.slice(0, index))
    setPreviewFile(null)
  }

  const goBack = () => {
    if (navHistory.length > 0) {
      const prev = navHistory[navHistory.length - 1]
      setNavHistory(prev => prev.slice(0, -1))
      setCurrentPath(prev)
      setPreviewFile(null)
    }
  }

  const handleBrowse = () => {
    if (!pathInput.trim()) return
    const parts = pathInput.trim().split('/').filter(Boolean)
    setCurrentPath(parts)
    setNavHistory([])
    setPreviewFile(null)
  }

  const isTextFile = (name: string): boolean => {
    const ext = name.split('.').pop()?.toLowerCase()
    return ['go', 'ts', 'tsx', 'js', 'jsx', 'json', 'md', 'yaml', 'yml', 'toml', 'sh', 'bat', 'py', 'css', 'html', 'txt', 'conf', 'mod', 'sum', 'env'].includes(ext || '')
  }

  const openPreview = (node: FSNode) => {
    if (node.type === 'dir') {
      navigateInto(node.name)
    } else if (isTextFile(node.name)) {
      setPreviewFile(node)
    }
  }

  return (
    <div className="flex h-full p-4 gap-4">
      {/* File list */}
      <div className={cn('flex flex-col flex-1 min-w-0', previewFile ? 'w-1/2' : 'w-full')}>
        {/* Path input */}
        <div className="flex items-center gap-2 mb-3">
          <div className="relative flex-1">
            <Folder size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted" />
            <input
              type="text"
              value={pathInput}
              onChange={(e) => setPathInput(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') handleBrowse() }}
              placeholder="Enter path to browse..."
              list="mock-paths"
              className="w-full bg-[#0f172a] border border-border rounded-lg pl-9 pr-3 py-2 text-sm text-text placeholder-muted focus:outline-none focus:border-primary"
            />
            <datalist id="mock-paths">
              {mockPaths.map(p => <option key={p} value={p} />)}
            </datalist>
          </div>
          <button
            onClick={handleBrowse}
            className="flex items-center gap-1.5 px-4 py-2 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
          >
            Browse
          </button>
        </div>

        {/* Breadcrumb */}
        <div className="flex items-center gap-1 mb-3 text-sm flex-wrap">
          <button onClick={() => { setCurrentPath([]); setPreviewFile(null) }} className="text-muted hover:text-text transition-colors">
            <Home size={14} />
          </button>
          {currentPath.length > 0 && (
            <button onClick={goBack} className="text-muted hover:text-text transition-colors ml-1 text-xs">
              &larr; Back
            </button>
          )}
          {segments.map((seg, i) => (
            <div key={i} className="flex items-center gap-1">
              <ChevronRight size={12} className="text-muted" />
              <button
                onClick={() => navigateTo(i)}
                className={cn(
                  'text-xs hover:text-text transition-colors',
                  i === segments.length - 1 ? 'text-text font-medium' : 'text-muted',
                )}
              >
                {seg.name}
              </button>
            </div>
          ))}
          <div className="flex-1" />
          <button
            onClick={() => { setPreviewFile(null); setCurrentPath([...currentPath]) }}
            className="text-xs text-muted hover:text-text transition-colors flex items-center gap-1"
          >
            <RefreshCw size={12} /> Refresh
          </button>
        </div>

        {/* File table */}
        <div className="flex-1 bg-[#0b1120] border border-border rounded-lg overflow-y-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-xs text-muted">
                <th className="text-left px-4 py-2 font-medium">Name</th>
                <th className="text-right px-4 py-2 font-medium w-24">Size</th>
                <th className="text-left px-4 py-2 font-medium w-20">Type</th>
                <th className="text-left px-4 py-2 font-medium w-40">Modified</th>
              </tr>
            </thead>
            <tbody>
              {currentDir.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-muted text-sm">Empty directory</td>
                </tr>
              )}
              {currentDir.map((node) => (
                <tr
                  key={node.name}
                  onClick={() => openPreview(node)}
                  className="border-b border-border/30 hover:bg-white/5 cursor-pointer transition-colors"
                >
                  <td className="px-4 py-2 flex items-center gap-2">
                    {node.type === 'dir' ? (
                      <span className="text-lg">📁</span>
                    ) : (
                      <span className="text-lg">📄</span>
                    )}
                    <span className="text-text">{node.name}</span>
                  </td>
                  <td className="px-4 py-2 text-right text-muted font-[JetBrains_Mono] text-xs">{node.size || '—'}</td>
                  <td className="px-4 py-2 text-muted text-xs">{node.type}</td>
                  <td className="px-4 py-2 text-muted text-xs">{node.modified}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Preview panel */}
      {previewFile && (
        <div className="w-1/2 flex flex-col">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-medium text-text flex items-center gap-2">
              <FileText size={14} className="text-primary" />
              {previewFile.name}
            </h3>
            <button
              onClick={() => setPreviewFile(null)}
              className="p-1 text-muted hover:text-text transition-colors"
            >
              <X size={14} />
            </button>
          </div>
          <div className="flex-1 bg-[#0b1120] border border-border rounded-lg overflow-y-auto p-4">
            <pre className="font-[JetBrains_Mono] text-xs text-text leading-relaxed whitespace-pre-wrap">
              {previewFile.content || '// No content available'}
            </pre>
          </div>
        </div>
      )}
    </div>
  )
}
