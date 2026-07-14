import { useState, useRef, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Send,
  Trash2,
  Bot,
  User,
  RefreshCw,
  AlertTriangle,
  Activity,
  Sparkles,
  ChevronRight,
  BrainCircuit,
  MessageSquare,
  ShieldCheck,
  Zap,
  Copy,
  Check,
  Lightbulb,
} from 'lucide-react'
import { format } from 'date-fns'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { EmptyState } from '@/components/ui/EmptyState'
import * as Tabs from '@radix-ui/react-tabs'
import { useOllamaStore } from '@/stores/useOllamaStore'
import type { ChatMessage, AnomalyInfo, OllamaStatus, AIInsight, ChatSession } from '@/types'

type TabId = 'ai-chat' | 'anomalies' | 'insights'

// ── Inline helpers ──

function ChatBubble({ role, content }: { role: string; content: string }) {
  const isAssistant = role === 'assistant' || role === 'system'
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* ignore */ }
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
  const { data: ollamaStatusData } = useQuery<OllamaStatus>({
    queryKey: ['ollama-status'],
    queryFn: async () => {
      const res = await call('AIOps.GetOllamaStatus') as OllamaStatus
      setOllamaStatus(res)
      return res
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  const ollamaStatus = ollamaStatusData ?? { available: false, model: '', version: '' }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      <div className="py-4 border-b border-border bg-panel-2 flex items-center justify-between">
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

      <Tabs.Root defaultValue="ai-chat" onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-border bg-panel px-4">
          {[
            { id: 'ai-chat', label: 'Analyst Chat', icon: <MessageSquare size={20} /> },
            { id: 'anomalies', label: 'Anomaly Detection', icon: <Activity size={20} /> },
            { id: 'insights', label: 'AI Insights', icon: <Lightbulb size={20} /> },
          ].map((tab) => (
            <Tabs.Trigger
              key={tab.id}
              value={tab.id}
              data-automation-id={`aiops-tab-${tab.id}`}
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
          <Tabs.Content value="anomalies" className="h-full">
            <AnomaliesTab />
          </Tabs.Content>
          <Tabs.Content value="insights" className="h-full">
            <InsightsTab />
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
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const [loaded, setLoaded] = useState(false)
  const [activeSession, setActiveSession] = useState<string>(`sess-${Date.now()}`)
  const chatRef = useRef<HTMLDivElement>(null)

  const { data: sessions = [], refetch: refetchSessions } = useQuery<ChatSession[]>({
    queryKey: ['chat-sessions'],
    queryFn: async () => {
      const res = await call('AIOps.ListSessions')
      return (res as ChatSession[]) || []
    }
  })

  // Load persisted messages when activeSession changes
  useEffect(() => {
    let cancelled = false
    const loadMessages = async () => {
      setLoaded(false)
      try {
        const msgs = await call('AIOps.GetMessages', activeSession) as ChatMessage[]
        if (!cancelled && msgs && msgs.length > 0) {
          setMessages(msgs)
        } else if (!cancelled) {
          setMessages([{ role: 'assistant', content: 'Greetings. I am the OpsForAll AI Analyst. I have analyzed your system metrics and identified several areas for potential optimization. How can I assist you today?' }])
        }
      } catch { /* ignore */
        if (!cancelled) {
          setMessages([{ role: 'assistant', content: 'Greetings. I am the OpsForAll AI Analyst. I have analyzed your system metrics and identified several areas for potential optimization. How can I assist you today?' }])
        }
      } finally {
        if (!cancelled) setLoaded(true)
      }
    }
    loadMessages()
    return () => { cancelled = false }
  }, [call, activeSession])

  const handleSend = async () => {
    if (!input.trim() || isTyping) return

    const userMsg: ChatMessage = { role: 'user', content: input }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setIsTyping(true)

    // Persist user message
    try {
      await call('AIOps.SaveMessage', activeSession, 'user', userMsg.content)
      refetchSessions()
    } catch { /* ignore */ }

    try {
      const response = await call('AIOps.Chat', input) as string
      const assistantMsg: ChatMessage = { role: 'assistant', content: response }
      setMessages(prev => [...prev, assistantMsg])
      // Persist assistant message
      try {
        await call('AIOps.SaveMessage', activeSession, 'assistant', response)
        refetchSessions()
      } catch { /* ignore */ }
    } catch {
      const errMsg = 'Analyst Error: request failed'
      setMessages(prev => [...prev, { role: 'assistant', content: errMsg }])
      try {
        await call('AIOps.SaveMessage', activeSession, 'assistant', errMsg)
      } catch { /* ignore */ }
    } finally {
      setIsTyping(false)
    }
  }

  const handleNewSession = () => {
    setActiveSession(`sess-${Date.now()}`)
  }

  const handleDeleteSession = async (sid: string, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!window.confirm('Delete this session and all its messages?')) return
    await call('AIOps.DeleteSession', sid)
    refetchSessions()
    if (activeSession === sid) {
      handleNewSession()
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
    <div className="flex h-full bg-[var(--color-bg)] overflow-hidden">
      {/* Sessions Sidebar */}
      <div className="w-80 border-r border-border bg-panel flex flex-col shrink-0">
        <div className="p-6 border-b border-border">
          <button
            onClick={handleNewSession}
            className="w-full flex items-center justify-center gap-2 py-3 px-4 bg-accent text-white rounded-xl font-bold hover:bg-accent/90 transition-all shadow-lg active:scale-95"
          >
            <Sparkles size={18} />
            New Intelligence Session
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          {sessions.length === 0 ? (
            <div className="py-10 px-4 text-center">
              <p className="text-sm font-bold text-text-faint uppercase tracking-widest mb-2">No History</p>
              <p className="text-xs text-text-dim">Persistent sessions will appear here.</p>
            </div>
          ) : (
            sessions.map((s) => (
              <div
                key={s.session_id}
                onClick={() => setActiveSession(s.session_id)}
                className={cn(
                  "group relative p-4 rounded-xl border transition-all cursor-pointer",
                  activeSession === s.session_id
                    ? "bg-accent-soft border-accent/40 shadow-md"
                    : "border-transparent hover:bg-[var(--color-sidebar-hover)]"
                )}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <p className={cn("text-sm font-bold truncate", activeSession === s.session_id ? "text-text" : "text-text-dim group-hover:text-text")}>
                      {s.session_id}
                    </p>
                    <div className="flex items-center gap-3 mt-1.5 text-[10px] font-bold text-text-faint uppercase tracking-tighter">
                      <span className="flex items-center gap-1"><MessageSquare size={10} /> {s.msg_count}</span>
                      <span>\u2022</span>
                      <span>{format(new Date(s.last_active), 'MMM d, HH:mm')}</span>
                    </div>
                  </div>
                  <button
                    onClick={(e) => handleDeleteSession(s.session_id, e)}
                    className="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-danger/10 hover:text-danger text-text-faint transition-all"
                    title="Delete Session"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Chat Area */}
      <div className="flex-1 flex flex-col relative min-w-0">
        {!loaded ? (
          <div className="flex items-center justify-center h-full bg-[var(--color-bg)]">
            <div className="flex items-center gap-3 text-text-dim">
              <RefreshCw size={20} className="animate-spin" />
              <span className="text-sm font-bold uppercase tracking-widest">Retrieving logs...</span>
            </div>
          </div>
        ) : (
          <>
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
                    {["Analyze recent system health", "Check for security anomalies", "Review network connection density", "Summarize resource usage trends"].map(prompt => (
                      <button
                        key={prompt}
                        onClick={() => { setInput(prompt) }}
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
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

// Reports tab removed

// ══════════════════════════════════════════════
//  Anomalies Tab
// ══════════════════════════════════════════════

function AnomaliesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: anomalies = [], isLoading, refetch } = useQuery<AnomalyInfo[]>({
    queryKey: ['anomalies'],
    queryFn: async () => {
      const res = await call('AIOps.DetectAnomalies') as AnomalyInfo[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="flex flex-col h-full space-y-8">
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

// ══════════════════════════════════════════════
//  Insights Tab
// ══════════════════════════════════════════════

function InsightsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: insights = [], isLoading, refetch } = useQuery<AIInsight[]>({
    queryKey: ['ai-insights'],
    queryFn: async () => {
      const res = await call('AIOps.GetAIInsights') as AIInsight[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  const severityConfig: Record<string, { icon: typeof Zap; color: string; border: string; bg: string }> = {
    critical: { icon: AlertTriangle, color: 'text-danger', border: 'border-danger/30', bg: 'bg-danger/10' },
    warning: { icon: Zap, color: 'text-warning', border: 'border-warning/30', bg: 'bg-warning/10' },
    info: { icon: ShieldCheck, color: 'text-accent', border: 'border-accent/30', bg: 'bg-accent/10' },
  }

  const categoryColors: Record<string, string> = {
    performance: 'bg-warning/15 text-warning border-warning/20',
    storage: 'bg-accent/15 text-accent border-accent/20',
    network: 'bg-success/15 text-success border-success/20',
    alerts: 'bg-danger/15 text-danger border-danger/20',
    general: 'bg-text-faint/15 text-text-faint border-border',
  }

  return (
    <div className="flex flex-col h-full space-y-8">
      <div className="flex items-center justify-between bg-panel border border-border px-6 py-4 rounded-xl shadow-lg">
        <div className="flex items-center gap-8">
          <div className="flex items-center gap-4">
            <div className={cn("w-10 h-10 rounded-lg flex items-center justify-center border", insights.length > 0 ? "bg-warning/10 border-warning/30 text-warning" : "bg-success/10 border-success/30 text-success")}>
              <Lightbulb size={32} />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums">{insights.length}</p>
              <p className="text-sm font-bold text-text-dim uppercase tracking-widest">Active Insights</p>
            </div>
          </div>
          <div className="w-px h-12 bg-border" />
          <div className="text-text-dim text-sm leading-relaxed max-w-md italic">
            Synthesized from anomaly detection, metric trends, and active alert analysis.
          </div>
        </div>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-3 px-5 py-2.5 bg-[var(--color-panel-3)] border border-[var(--color-border)] rounded-lg hover:bg-panel hover:border-accent/40 text-text font-bold transition-all shadow-lg active:scale-95"
        >
          <RefreshCw size={20} className={cn(isLoading && "animate-spin")} />
          Refresh
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-4">
        {insights.length === 0 ? (
          <div className="flex items-center justify-center h-64">
            <EmptyState
              icon={<Lightbulb size={28} />}
              title="No Insights Available"
              description="System is operating within normal parameters. Insights will appear when anomalies, trends, or alerts require attention."
            />
          </div>
        ) : (
          insights.map((insight, i) => {
            const cfg = severityConfig[insight.severity] || severityConfig.info
            const SevIcon = cfg.icon
            return (
              <div
                key={i}
                className={cn(
                  'bg-panel border rounded-xl p-6 transition-all hover:shadow-lg group',
                  cfg.border,
                )}
              >
                <div className="flex items-start gap-4">
                  <div className={cn('w-10 h-10 rounded-lg flex items-center justify-center border shrink-0', cfg.bg, cfg.border)}>
                    <SevIcon size={22} className={cfg.color} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="text-base font-bold text-[var(--color-text)] group-hover:text-[var(--color-accent)] transition-colors">
                        {insight.title}
                      </h3>
                      <span className={cn('px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-tighter border', categoryColors[insight.category] || categoryColors.general)}>
                        {insight.category}
                      </span>
                      <StatusBadge status={insight.severity} />
                    </div>
                    <p className="text-sm text-[var(--color-text-dim)] leading-relaxed mb-3">{insight.message}</p>
                    <div className="flex items-center gap-2 text-xs text-[var(--color-accent)] font-semibold">
                      <ChevronRight size={14} />
                      {insight.action}
                    </div>
                  </div>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
