import { useState, useRef, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Send,
  Trash2,
  Bot,
  User,
  FileText,
  RefreshCw,
  AlertTriangle,
  Activity,
  Sparkles,
  ChevronRight,
  BrainCircuit,
  MessageSquare,
  ShieldCheck,
  Zap,
  Globe,
  Copy,
  Check,
} from 'lucide-react'
import { format } from 'date-fns'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { EmptyState } from '@/components/ui/EmptyState'
import * as Tabs from '@radix-ui/react-tabs'
import { useOllamaStore } from '@/stores/useOllamaStore'
import type { ChatMessage, AnomalyInfo, OllamaStatus } from '@/types'

type TabId = 'ai-chat' | 'reports' | 'anomalies'

// ── Inline helpers ──

function ChatBubble({ role, content }: { role: string; content: string }) {
  const isAssistant = role === 'assistant' || role === 'system'
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // clipboard API may fail in some contexts
    }
  }

  return (
    <div className={cn('flex gap-6 max-w-[90%] animate-in fade-in slide-in-from-bottom-2 duration-300', !isAssistant ? 'ml-auto flex-row-reverse' : '')}>
      <div
        className={cn(
          'w-12 h-12 rounded-xl flex items-center justify-center shrink-0 shadow-lg border',
          !isAssistant ? 'bg-[var(--color-accent)] border-[var(--color-accent)]/20' : 'bg-[var(--color-panel-3)] border-[var(--color-accent)]/20',
        )}
      >
        {!isAssistant ? (
          <User size={24} className="text-white" />
        ) : (
          <Bot size={24} className="text-[var(--color-accent)]" />
        )}
      </div>
      <div className="flex flex-col space-y-2">
        <div
          className={cn(
            'rounded-2xl px-6 py-4 text-lg shadow-xl relative group',
            !isAssistant
              ? 'bg-[var(--color-accent)] text-white rounded-tr-none'
              : 'bg-[var(--color-panel-2)] border border-[var(--color-border)] text-[var(--color-text)] rounded-tl-none',
          )}
        >
          <div className="whitespace-pre-wrap leading-relaxed tabular-nums">{content}</div>
          {isAssistant && (
            <button
              onClick={handleCopy}
              className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity p-1.5 rounded-lg hover:bg-[var(--color-panel-3)] text-[var(--color-text-faint)] hover:text-[var(--color-text)]"
              aria-label="Copy message"
            >
              {copied ? <Check size={14} className="text-[var(--color-success)]" /> : <Copy size={14} />}
            </button>
          )}
        </div>
        <span className={cn("text-xs font-semibold text-[var(--color-text-faint)] px-1", !isAssistant ? "text-right" : "text-left")}>
          {role} \u2022 {format(new Date(), 'HH:mm')}
        </span>
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    critical: 'bg-danger/20 text-danger border-danger/30',
    high: 'bg-warning/20 text-warning border-warning/30',
    medium: 'bg-warning/15 text-warning border-warning/20',
    low: 'bg-accent/20 text-accent border-accent/30',
    warning: 'bg-warning/20 text-warning border-warning/30',
    info: 'bg-accent/20 text-accent border-accent/30',
    healthy: 'bg-success/20 text-success border-success/30',
  }
  const key = status.toLowerCase()
  return (
    <span className={cn('px-3 py-1 rounded-full text-xs font-bold uppercase tracking-tighter border', colors[key] || 'bg-text-faint/20 text-text-faint border-border')}>
      {status}
    </span>
  )
}

// ══════════════════════════════════════════════
//  AIOps Page
// ══════════════════════════════════════════════

export function AIOps() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [activeTab, setActiveTab] = useState<TabId>('ai-chat')
  const setOllamaStatus = useOllamaStore((s) => s.setStatus)

  // Ollama status via react-query, synced to zustand store
  const { data: ollamaStatus = { available: false, model: '', version: '' } } = useQuery<OllamaStatus>({
    queryKey: ['ollama-status'],
    queryFn: async () => {
      const res = await call('AIOps.GetOllamaStatus') as OllamaStatus
      setOllamaStatus(res)
      return res
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      <div className="p-8 border-b border-border bg-panel-2 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text flex items-center gap-3">
            <BrainCircuit size={32} className="text-accent" />
            AI Operations Analyst
          </h1>
          <p className="text-text-dim text-lg mt-2">
            Local intelligence for system diagnostics, trend analysis, and anomaly detection.
          </p>
        </div>
        <div className="flex items-center gap-4 bg-panel border border-border px-6 py-3 rounded-2xl shadow-inner">
          <div className="flex items-center gap-2">
            <div className={cn("w-2 h-2 rounded-full", ollamaStatus.available ? "bg-success animate-pulse shadow-[0_0_8px_var(--color-success)]" : "bg-danger")} />
            <span className="text-sm font-bold text-text-dim uppercase tracking-widest">
              {ollamaStatus.available ? 'Ollama Online' : 'Ollama Offline'}
            </span>
          </div>
          <div className="w-px h-4 bg-border" />
          <span className="text-sm font-bold text-accent">
            {ollamaStatus.available ? ollamaStatus.model : (ollamaStatus.error || '—')}
          </span>
        </div>
      </div>

      <Tabs.Root value={activeTab} onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-border bg-panel px-4">
          {[
            { id: 'ai-chat', label: 'Analyst Chat', icon: <MessageSquare size={20} /> },
            { id: 'reports', label: 'Intelligence Reports', icon: <FileText size={20} /> },
            { id: 'anomalies', label: 'Anomaly Detection', icon: <Activity size={20} /> },
          ].map((tab) => (
            <Tabs.Trigger
              key={tab.id}
              value={tab.id}
              className={cn(
                'flex items-center gap-3 px-8 py-5 text-base font-bold transition-all border-b-2 border-transparent',
                activeTab === tab.id ? 'border-accent text-text bg-accent-soft' : 'text-text-faint hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}
            >
              {tab.icon}
              {tab.label}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="ai-chat" className="h-full">
            <ChatTab />
          </Tabs.Content>
          <Tabs.Content value="reports" className="h-full">
            <ReportsTab />
          </Tabs.Content>
          <Tabs.Content value="anomalies" className="h-full">
            <AnomaliesTab />
          </Tabs.Content>
        </div>
      </Tabs.Root>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Chat Tab
// ══════════════════════════════════════════════

function ChatTab() {
  const { call } = useBackend()
  const [messages, setMessages] = useState<ChatMessage[]>([
    { role: 'assistant', content: 'Greeting, Operator. I am the Hawkward AI Analyst. I have full visibility into your system metrics, network state, and security logs. How can I assist you with your operations today?' }
  ])
  const [input, setInput] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const chatRef = useRef<HTMLDivElement>(null)

  const handleSend = async () => {
    if (!input.trim() || isTyping) return

    const userMsg: ChatMessage = { role: 'user', content: input }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setIsTyping(true)

    try {
      const response = await call('AIOps.Chat', input) as string
      setMessages(prev => [...prev, { role: 'assistant', content: response }])
    } catch (err) {
      setMessages(prev => [...prev, { role: 'assistant', content: `Analyst Error: ${String(err)}` }])
    } finally {
      setIsTyping(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  useEffect(() => {
    if (chatRef.current) {
      chatRef.current.scrollTop = chatRef.current.scrollHeight
    }
  }, [messages, isTyping])

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] relative">
      {/* Background Decor */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none opacity-20">
        <div className="absolute -top-24 -left-24 w-96 h-96 bg-accent rounded-full blur-[120px]" />
        <div className="absolute bottom-0 right-0 w-[500px] h-[500px] bg-accent rounded-full blur-[150px] opacity-30" />
      </div>

      {/* Messages */}
      <div ref={chatRef} className="flex-1 overflow-y-auto p-10 space-y-8 relative z-10 scroll-smooth">
        {messages.map((msg, i) => (
          <ChatBubble key={i} role={msg.role} content={msg.content} />
        ))}
        {isTyping && (
          <div className="flex gap-6 max-w-[80%] animate-pulse">
            <div className="w-12 h-12 rounded-xl bg-panel-3 border border-accent/20 flex items-center justify-center shrink-0">
              <Bot size={24} className="text-accent" />
            </div>
            <div className="bg-panel-2 border border-border rounded-2xl rounded-tl-none px-6 py-4">
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 bg-accent rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <span className="w-2 h-2 bg-accent rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <span className="w-2 h-2 bg-accent rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                <span className="ml-2 text-sm font-bold text-text-faint uppercase tracking-widest">Analyzing System Context...</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Input Area */}
      <div className="p-8 bg-panel-2 border-t border-border relative z-20">
        <div className="max-w-5xl mx-auto space-y-6">
          {/* Suggested Prompts */}
          {messages.length < 3 && (
            <div className="flex flex-wrap gap-3">
              {[
                "Analyze recent system health",
                "Check for security anomalies",
                "Review network connection density",
                "Summarize resource usage trends"
              ].map(prompt => (
                <button
                  key={prompt}
                  onClick={() => { setInput(prompt); }}
                  className="px-4 py-2 bg-panel border border-border rounded-full text-sm font-bold text-text-dim hover:text-accent hover:border-accent/40 transition-all"
                >
                  {prompt}
                </button>
              ))}
            </div>
          )}

          <div className="flex items-end gap-4">
            <div className="relative flex-1 group">
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Ask about system health, anomalies, or network state..."
                rows={1}
                className="w-full bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl px-6 py-5 text-xl text-text placeholder-text-faint focus:outline-none focus:border-accent transition-all shadow-inner resize-none min-h-[64px] max-h-40"
                style={{ height: 'auto' }}
              />
              <div className="absolute right-4 bottom-4 flex items-center gap-3 text-text-faint">
                <span className="text-xs font-bold uppercase tracking-tighter group-focus-within:text-accent transition-colors">Shift+Enter for newline</span>
                <div className="w-px h-3 bg-border" />
                <Zap size={14} className="group-focus-within:text-warning transition-colors" />
              </div>
            </div>
            <button
              onClick={handleSend}
              disabled={!input.trim() || isTyping}
              className="h-16 w-16 flex items-center justify-center bg-accent text-white rounded-2xl hover:bg-accent/90 disabled:opacity-30 disabled:scale-95 transition-all shadow-lg active:scale-90"
            >
              <Send size={28} />
            </button>
            <button
              onClick={() => setMessages([])}
              className="h-16 w-16 flex items-center justify-center text-text-faint border border-border rounded-2xl hover:bg-danger/10 hover:text-danger transition-all"
              title="Clear History"
            >
              <Trash2 size={24} />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Reports Tab
// ══════════════════════════════════════════════

function ReportsTab() {
  const { call } = useBackend()
  const [reportType, setReportType] = useState('System Health')
  const [reportContent, setReportContent] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)

  const templates = [
    { id: 'health', label: 'System Health', desc: 'Aggregate performance, uptime, and resource usage.', icon: <Activity size={20} /> },
    { id: 'security', label: 'Security Audit', desc: 'Failed logins, firewall status, and user changes.', icon: <ShieldCheck size={20} /> },
    { id: 'network', label: 'Network Triage', desc: 'Latency trends, DNS health, and connection density.', icon: <Globe size={20} /> },
  ]

  const generate = async (name: string) => {
    setReportType(name)
    setIsGenerating(true)
    setReportContent('')
    try {
      const res = await call('AIOps.GenerateReport', [name]) as string
      setReportContent(res)
    } catch (err) {
      setReportContent(`Generation Error: ${String(err)}`)
    } finally {
      setIsGenerating(false)
    }
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 h-full p-8 gap-8 overflow-hidden">
      <div className="lg:col-span-1 space-y-4 overflow-y-auto pr-2">
        <h3 className="text-lg font-bold text-text-dim uppercase tracking-widest mb-4">Intelligence Templates</h3>
        {templates.map(t => (
          <button
            key={t.id}
            onClick={() => generate(t.label)}
            disabled={isGenerating}
            className={cn(
              "w-full text-left border rounded-2xl p-6 transition-all group disabled:opacity-50",
              reportType === t.label ? "bg-accent-soft border-accent/50" : "bg-panel border-border hover:bg-[var(--color-sidebar-hover)]"
            )}
          >
            <div className="flex items-center gap-4 mb-3">
              <div className="w-10 h-10 rounded-lg bg-panel-3 flex items-center justify-center text-accent border border-border">
                {t.icon}
              </div>
              <span className="text-xl font-bold text-text">{t.label}</span>
            </div>
            <p className="text-text-dim text-base leading-relaxed mb-4">{t.desc}</p>
            <div className="flex items-center justify-between text-xs font-bold uppercase tracking-widest text-accent opacity-0 group-hover:opacity-100 transition-opacity">
              <span>Draft Intel Report</span>
              <ChevronRight size={14} />
            </div>
          </button>
        ))}
      </div>

      <div className="lg:col-span-3 flex flex-col space-y-4 min-w-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <FileText size={20} className="text-text-faint" />
            <h3 className="text-lg font-bold text-text-dim uppercase tracking-widest">{reportType} Document</h3>
          </div>
          {isGenerating && (
            <div className="flex items-center gap-2 text-accent animate-pulse">
              <Sparkles size={16} />
              <span className="text-sm font-bold uppercase tracking-tighter">AI Synthesizing Report...</span>
            </div>
          )}
        </div>
        <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl p-10 overflow-y-auto font-[Geist_Mono] text-lg leading-relaxed whitespace-pre shadow-inner">
          {reportContent || (isGenerating ? 'Collecting system metrics and performing heuristic analysis...' : 'Select a template to generate a professional intelligence report.')}
        </div>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Anomalies Tab
// ══════════════════════════════════════════════

function AnomaliesTab() {
  const { call } = useBackend()

  const { data: anomalies = [], isLoading, refetch } = useQuery<AnomalyInfo[]>({
    queryKey: ['anomalies'],
    queryFn: async () => {
      const res = await call('AIOps.DetectAnomalies') as AnomalyInfo[]
      return res || []
    },
    refetchInterval: 10000,
  })

  return (
    <div className="flex flex-col h-full p-8 space-y-8">
      <div className="flex items-center justify-between bg-panel border border-border px-6 py-4 rounded-xl shadow-lg">
        <div className="flex items-center gap-8">
          <div className="flex items-center gap-4">
            <div className={cn("w-10 h-10 rounded-lg flex items-center justify-center border", anomalies.length > 0 ? "bg-danger/10 border-danger/30 text-danger" : "bg-success/10 border-success/30 text-success")}>
              {anomalies.length > 0 ? <AlertTriangle size={32} /> : <ShieldCheck size={32} />}
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums">{anomalies.length}</p>
              <p className="text-sm font-bold text-text-dim uppercase tracking-widest">Detected Anomalies</p>
            </div>
          </div>
          <div className="w-px h-12 bg-border" />
          <div className="text-text-dim text-sm leading-relaxed max-w-md italic">
            Statistical deviation tracking compares current live metrics against a 12-minute rolling window of history.
          </div>
        </div>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-3 px-5 py-2.5 bg-[var(--color-panel-3)] border border-[var(--color-border)] rounded-lg hover:bg-panel hover:border-accent/40 text-text font-bold transition-all shadow-lg active:scale-95"
        >
          <RefreshCw size={20} className={cn(isLoading && "animate-spin")} />
          Deep Scan Now
        </button>
      </div>

      <div className="flex-1 bg-[var(--color-bg)] border border-border rounded-2xl overflow-hidden shadow-inner">
        <div className="overflow-y-auto h-full">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Heuristic Signal</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Observed</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Expected (Mean)</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Deviation</th>
                <th className="px-6 py-4 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Severity</th>
              </tr>
            </thead>
            <tbody>
              {anomalies.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-16 text-center">
                    <EmptyState
                      icon={<ShieldCheck size={28} />}
                      title="No Anomalies Detected"
                      description="System operating within established statistical baseline. All metrics are within normal operating parameters."
                    />
                  </td>
                </tr>
              ) : (
                anomalies.map((a, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-colors group">
                    <td className="px-6 py-4 font-semibold text-sm text-[var(--color-text)] group-hover:text-[var(--color-accent)] transition-colors flex items-center gap-4">
                      <Zap size={20} className="text-warning" />
                      {a.metric.replace('.percent', '').toUpperCase()}
                    </td>
                    <td className="px-10 py-6 text-sm font-semibold text-[var(--color-text)] tabular-nums">{a.value.toFixed(2)}%</td>
                    <td className="px-6 py-4 text-sm text-[var(--color-text-faint)] tabular-nums">{a.expected.toFixed(2)}%</td>
                    <td className="px-6 py-4 text-sm font-semibold text-[var(--color-danger)] tabular-nums">+{a.deviation.toFixed(1)}σ</td>
                    <td className="px-6 py-4"><StatusBadge status={a.severity} /></td>
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
