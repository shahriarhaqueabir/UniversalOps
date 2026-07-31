import { useState, useRef, useEffect } from 'react'
import { motion } from 'motion/react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
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
  ChevronDown,
  ChevronUp,
  BrainCircuit,
  MessageSquare,
  ShieldCheck,
  ShieldAlert,
  LockOpen,
  XCircle,
  Zap,
  Copy,
  Check,
  Lightbulb,
  Cpu,
  MemoryStick,
  HardDrive,
  ArrowUp,
  ArrowDown,
  Minus,
  Loader2,
  Workflow,
  BarChart3,
} from 'lucide-react'
import { format } from 'date-fns'
import { toast } from 'sonner'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { Panel } from '@/components/ui/Panel'
import { cn, formatSafeDate } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useEvents } from '@/hooks/useEvents'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { EmptyState } from '@/components/ui/EmptyState'
import * as Tabs from '@radix-ui/react-tabs'
import { useOllamaStore } from '@/stores/useOllamaStore'
import { useNavigationStore, type Page } from '@/stores/useSettingsStore'
import type { ChatMessage, AnomalyInfo, OllamaStatus, AIInsight, AIConfidence, ConversationMessage, ChatSession, DashboardData, ActionPreview, AIWorkflowEvent, DataStreamMetric } from '@/types'

type TabId = 'overview' | 'ai-chat' | 'anomalies' | 'insights'

interface ChatResponse {
  content: string
  actions?: ActionPreview[]
}

// ── Inline helpers ──

interface ChatBubbleProps extends ChatMessage {
  sessionID: string;
  onAssistantReply: (content: string, actions?: ActionPreview[]) => void;
}

function ChatBubble({ role, content, actions, sessionID, onAssistantReply }: ChatBubbleProps) {
  const { call } = useBackend()
  const isAssistant = role === 'assistant' || role === 'system'
  const [copied, setCopied] = useState(false)
  const [authorizedIds, setAuthorizedIds] = useState<Set<string>>(new Set())
  const [abortedIds, setAbortedIds] = useState<Set<string>>(new Set())
  const [isExecuting, setIsExecuting] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { toast.error('Failed to copy to clipboard') }
  }

  const handleAuthorize = async (action: ActionPreview) => {
    if (isExecuting) return
    setIsExecuting(true)
    const tid = toast.loading(`Executing ${action.action}...`)
    try {
      const result = await call('App.ConfirmAction', action.handshake_id) as any
      setAuthorizedIds(prev => new Set(prev).add(action.handshake_id))
      toast.success('Action executed successfully', { id: tid })

      // Explicitly show result in chat
      const detail = result.message || result.error || 'No output'
      const status = result.success ? 'SUCCESS' : 'FAILED'

      // Notify the AI about the success and GET a summary response
      try {
        const response = await call('AIOps.NotifyActionResult', sessionID, action.action, status, detail, action.handshake_id) as ChatResponse
        if (response.content) {
          onAssistantReply(response.content, response.actions)
        }
      } catch (chatErr) {
        console.error('Failed to notify AI of action success:', chatErr)
      }
    } catch (err) {
      console.error(err)
      toast.error('Neural Authorization Failed', { id: tid })
    } finally {
      setIsExecuting(false)
    }
  }

  const handleAbort = async (action: ActionPreview) => {
    setAbortedIds(prev => new Set(prev).add(action.handshake_id))
    // Notify the AI about the abort
    try {
      const response = await call('AIOps.NotifyActionResult', sessionID, action.action, 'ABORTED', 'User cancelled the action.', action.handshake_id) as ChatResponse
      if (response.content) {
        onAssistantReply(response.content, response.actions)
      }
    } catch (chatErr) {
      console.error('Failed to notify AI of action abort:', chatErr)
    }
  }

  return (
    <div className={cn('flex gap-6 max-w-[95%] animate-in fade-in slide-in-from-bottom-2 duration-300', !isAssistant ? 'ml-auto flex-row-reverse' : '')}>
      <div
        className={cn(
          'w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 shadow-lg border transition-transform duration-500 hover:rotate-6',
          !isAssistant ? 'bg-[var(--color-accent)] border-white/20 shadow-accent/20' : 'bg-[var(--color-panel-3)] border-[var(--color-border)]',
        )}
      >
        {!isAssistant ? (
          <User size={24} className="text-white" />
        ) : (
          <Bot size={24} className="text-[var(--color-accent)]" />
        )}
      </div>
      <div className="flex flex-col space-y-3 flex-1 min-w-0">
        <div
          className={cn(
            'rounded-[2rem] px-8 py-6 text-base shadow-2xl relative group transition-all',
            !isAssistant
              ? 'bg-[var(--color-accent)] text-white rounded-tr-none ml-auto border-t-white/20'
              : 'bg-[var(--color-panel-2)] border border-[var(--color-border)] text-[var(--color-text)] rounded-tl-none',
          )}
        >
          <div className="whitespace-pre-wrap leading-relaxed tabular-nums font-medium text-wrap-pretty">{content}</div>
          {isAssistant && (
            <button
              onClick={handleCopy}
              className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-all p-2 rounded-xl bg-panel-3/50 hover:bg-panel-3 text-[var(--color-text-faint)] hover:text-text border border-border"
              aria-label="Copy message"
            >
              {copied ? <Check size={16} className="text-[var(--color-success)]" /> : <Copy size={16} />}
            </button>
          )}
        </div>

        {actions && actions.map((action) => (
          <div key={action.handshake_id}>
            {!authorizedIds.has(action.handshake_id) && !abortedIds.has(action.handshake_id) && (
              <div className="bg-panel border-2 border-warning/30 rounded-2xl overflow-hidden shadow-2xl animate-in zoom-in-95 duration-300 mb-4">
                <div className="bg-warning/10 px-6 py-3 border-b border-warning/20 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <ShieldAlert size={18} className="text-warning" />
                    <span className="text-xs font-black uppercase tracking-widest text-warning">Neural Authorization Required</span>
                  </div>
                  <span className="text-[10px] font-mono text-warning/60">{action.handshake_id.slice(0, 8)}</span>
                </div>
                <div className="p-6 space-y-4">
                  <div>
                    <p className="text-xs font-bold text-text-faint uppercase tracking-tighter mb-1">Proposed Action</p>
                    <p className="text-base font-bold text-text">{action.description}</p>
                  </div>
                  {action.command && (
                    <div className="bg-panel-3/50 border border-border p-3 rounded-xl">
                      <p className="text-[10px] font-bold text-text-faint uppercase tracking-widest mb-2">Technical command to be executed</p>
                      <code className="text-xs font-mono text-accent block break-all whitespace-pre-wrap">{action.command}</code>
                    </div>
                  )}
                  <div>
                    <p className="text-xs font-bold text-text-faint uppercase tracking-tighter mb-2">Technical Risks</p>
                    <ul className="space-y-1">
                      {action.risks.map((risk) => (
                        <li key={risk} className="text-xs text-danger/80 flex items-start gap-2">
                          <span className="mt-1.5 w-1 h-1 rounded-full bg-danger shrink-0" />
                          {risk}
                        </li>
                      ))}
                    </ul>
                  </div>
                  <div className="pt-2 flex gap-3">
                    <button
                      onClick={() => handleAuthorize(action)}
                      aria-label={`Authorize execution of ${action.action}: ${action.description}`}
                      className="flex-1 bg-accent text-white py-2.5 rounded-xl font-bold text-sm hover:bg-accent/90 transition-all active:scale-95 shadow-lg flex items-center justify-center gap-2"
                    >
                      <LockOpen size={16} aria-hidden="true" />
                      Authorize Execution
                    </button>
                    <button
                      onClick={() => handleAbort(action)}
                      aria-label="Abort this proposed action"
                      className="px-6 border border-border bg-panel-2 py-2.5 rounded-xl font-bold text-sm text-text-dim hover:bg-panel-3 transition-all active:scale-95"
                    >
                      Abort
                    </button>
                  </div>
                </div>
              </div>
            )}

            {authorizedIds.has(action.handshake_id) && (
              <div className="bg-success/10 border border-success/30 rounded-xl px-4 py-2 flex items-center gap-3 animate-in fade-in duration-500 max-w-fit mb-4">
                <ShieldCheck size={16} className="text-success" />
                <span className="text-[10px] font-black text-success uppercase tracking-widest">Handshake Complete: Action Executed</span>
              </div>
            )}

            {abortedIds.has(action.handshake_id) && (
              <div className="bg-panel-3 border border-border rounded-xl px-4 py-2 flex items-center gap-3 opacity-60 max-w-fit mb-4">
                <XCircle size={16} className="text-text-faint" />
                <span className="text-[10px] font-black text-text-faint uppercase tracking-widest">Neural Handshake Aborted</span>
              </div>
            )}
          </div>
        ))}

        <span className={cn("text-[10px] font-bold text-[var(--color-text-faint)] px-1 uppercase tracking-tighter", !isAssistant ? "text-right" : "text-left")}>
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
  const queryClient = useQueryClient()
  const [initializing, setInitializing] = useState(false)
  const [isModelDropdownOpen, setIsModelDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Ollama status via react-query
  const { data: ollamaStatusData } = useQuery<OllamaStatus>({
    queryKey: ['ollama-status'],
    queryFn: async () => {
      const res = await call('AIOps.GetOllamaStatus') as OllamaStatus
      return res
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  useEffect(() => {
    if (ollamaStatusData) {
      setOllamaStatus(ollamaStatusData)
    }
  }, [ollamaStatusData, setOllamaStatus])

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsModelDropdownOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  const ollamaStatus: OllamaStatus = ollamaStatusData ?? { available: false, binary_exists: false, model: '', version: '' }

  // Check if the universalops persona model exists in the available models list
  const universalopsModel = ollamaStatus.available_models?.find(m => m.startsWith('universalops'))
  const universalopsExistsInOllama = !!universalopsModel
  const isPersonaMissing = ollamaStatus.available && !ollamaStatus.model.startsWith('universalops')
  const canJustSwitch = isPersonaMissing && universalopsExistsInOllama

  const handleInitializePersona = async () => {
    if (initializing) return
    setInitializing(true)

    if (universalopsExistsInOllama && universalopsModel) {
      // Model already exists — just switch to it
      const tid = toast.loading(`Switching to ${universalopsModel}...`)
      try {
        await call('AIOps.SetOllamaModel', universalopsModel)
        toast.success(`Switched to persona model: ${universalopsModel}`, { id: tid })
        queryClient.invalidateQueries({ queryKey: ['ollama-status'] })
      } catch (err: any) {
        toast.error(`Failed to switch: ${err.message}`, { id: tid })
      } finally {
        setInitializing(false)
      }
      return
    }

    const tid = toast.loading('Creating specialized AI persona from modelfile...')
    try {
      await call('AIOps.CreateOpsPersona')
      toast.success('Persona initialized successfully', { id: tid })
      queryClient.invalidateQueries({ queryKey: ['ollama-status'] })
    } catch (err: any) {
      toast.error(`Initialization failed: ${err.message}`, { id: tid })
    } finally {
      setInitializing(false)
    }
  }

  const handleSetModel = async (model: string) => {
    try {
      await call('AIOps.SetOllamaModel', model)
      setIsModelDropdownOpen(false)
      toast.success(`Switched to model: ${model}`)
      queryClient.invalidateQueries({ queryKey: ['ollama-status'] })
    } catch (err: any) {
      toast.error(`Failed to switch model: ${err.message}`)
    }
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500">
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 flex items-center justify-between px-10">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
               <BrainCircuit size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em] text-wrap-balance">Neural Processing</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">AI Operations Analyst</p>
          <p className="text-[var(--color-text-dim)] text-xs font-semibold uppercase tracking-widest mt-2 text-wrap-pretty">Local intelligence for system diagnostics and trend analysis</p>
        </div>
        <div className="flex items-center gap-4">
          {isPersonaMissing && (
            <button
              onClick={handleInitializePersona}
              disabled={initializing}
              className={cn(
                "flex items-center gap-2.5 px-6 py-3 rounded-xl text-sm font-bold transition-all disabled:opacity-50 active:scale-95",
                canJustSwitch
                  ? "bg-accent/10 border border-accent/30 text-accent hover:bg-accent/20"
                  : "bg-warning/10 border border-warning/30 text-warning hover:bg-warning/20"
              )}
              title={canJustSwitch
                ? `Switch to the existing ${universalopsModel} persona model`
                : 'Create the universalops model from the modelfile in data/ directory'}
            >
              {initializing ? <RefreshCw size={14} className="animate-spin" /> : <Bot size={14} />}
              {canJustSwitch ? 'Switch to Persona' : 'Initialize Persona'}
            </button>
          )}

          <div className="relative" ref={dropdownRef}>
            <button
              onClick={() => setIsModelDropdownOpen(!isModelDropdownOpen)}
              className="flex items-center gap-5 bg-[var(--color-panel)] border border-[var(--color-border)] px-6 py-3.5 rounded-2xl shadow-xl hover:border-accent/30 transition-all outline-none focus:border-accent group"
            >
              <div className="flex items-center gap-3">
                <div className={cn("w-2.5 h-2.5 rounded-full", ollamaStatus.available ? "bg-success shadow-[0_0_8px_var(--color-success)]" : "bg-danger")} />
                <span className="text-[10px] font-black text-text-faint uppercase tracking-[0.15em] group-hover:text-text-dim transition-colors">
                  {ollamaStatus.available ? 'Ollama Online' : 'Ollama Offline'}
                </span>
              </div>
              <div className="w-px h-4 bg-border" />
              <div className="flex items-center gap-3">
                <span className="text-sm font-black text-accent tracking-tight">
                  {ollamaStatus.available ? ollamaStatus.model : (ollamaStatus.error || 'N/A')}
                </span>
                <ChevronDown size={14} className={cn("text-text-faint transition-transform", isModelDropdownOpen && "rotate-180")} />
              </div>
            </button>

            {isModelDropdownOpen && (
              <div className="absolute top-full right-0 mt-2 z-[300] min-w-[260px] bg-panel-2 border border-border rounded-xl p-2 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
                <p className="px-3 py-2 text-[10px] font-bold text-text-faint uppercase tracking-[0.2em]">Available Models</p>
                <div className="max-h-64 overflow-y-auto space-y-1">
                  {ollamaStatus.available_models?.map(m => {
                    const isCurrent = m === ollamaStatus.model
                    const isPersonaModel = m.startsWith('universalops')
                    return (
                      <button
                        key={m}
                        onClick={() => handleSetModel(m)}
                        className={cn(
                          "w-full px-3 py-2.5 rounded-lg text-sm font-bold flex items-center justify-between gap-2 transition-all",
                          isCurrent ? "bg-accent text-white" : "text-text-dim hover:bg-accent-soft hover:text-text"
                        )}
                      >
                        <span className="truncate flex items-center gap-2">
                          {m}
                          {isPersonaModel && <Bot size={12} className="shrink-0" />}
                        </span>
                        <span className="flex items-center gap-2 shrink-0">
                          {isCurrent && <Check size={16} />}
                          {isPersonaModel && !isCurrent && (
                            <span className="text-[8px] font-bold text-accent uppercase tracking-wider bg-accent/10 px-1.5 py-0.5 rounded">PERSONA</span>
                          )}
                        </span>
                      </button>
                    )
                  })}
                  {!ollamaStatus.available_models?.length && (
                    <div className="px-3 py-4 text-center">
                      <p className="text-xs text-text-faint italic">No models found</p>
                    </div>
                  )}
                </div>
                <div className="h-px bg-border my-2" />
                <button
                  onClick={() => {
                    queryClient.invalidateQueries({ queryKey: ['ollama-status'] })
                    setIsModelDropdownOpen(false)
                  }}
                  className="w-full px-3 py-2.5 rounded-lg text-sm font-bold text-accent hover:bg-accent-soft flex items-center gap-2 transition-all"
                >
                  <RefreshCw size={14} />
                  Refresh Status
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <Tabs.Root defaultValue="ai-chat" onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-[var(--color-border)] bg-[var(--color-panel)] px-6">
          {[
            { id: 'overview', label: 'Overview', icon: <BrainCircuit size={18} /> },
            { id: 'ai-chat', label: 'Analyst Chat', icon: <MessageSquare size={18} /> },
            { id: 'anomalies', label: 'Anomaly Detection', icon: <Activity size={18} /> },
            { id: 'insights', label: 'AI Insights', icon: <Lightbulb size={18} /> },
          ].map((tab) => (
            <Tabs.Trigger
              key={tab.id}
              value={tab.id}
              data-automation-id={`aiops-tab-${tab.id}`}
              className={cn(
                'flex items-center gap-3 px-10 py-5 text-sm font-bold transition-all border-b-2 border-transparent relative',
                activeTab === tab.id ? 'text-accent' : 'text-text-faint hover:text-text hover:bg-[var(--color-sidebar-hover)]',
              )}
            >
              {tab.icon}
              <span className="uppercase tracking-widest text-[10px] font-black">{tab.label}</span>
              {activeTab === tab.id && (
                <motion.div
                  layoutId="aiops-tab-indicator"
                  className="absolute bottom-0 left-0 right-0 h-0.5 bg-accent"
                  transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                />
              )}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="overview" className="h-full">
            <OverviewTab />
          </Tabs.Content>
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
//  Overview Tab
// ══════════════════════════════════════════════

function OverviewTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const { navigate } = useNavigationStore()

  const { data: ollamaStatus } = useQuery<OllamaStatus>({
    queryKey: ['ollama-status'],
    queryFn: async () => await call('AIOps.GetOllamaStatus') as OllamaStatus,
    refetchInterval: refreshInterval,
  })

  const { data: insights = [] } = useQuery<AIInsight[]>({
    queryKey: ['ai-insights'],
    queryFn: async () => {
      const res = await call('AIOps.GetAIInsights') as AIInsight[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: confidence } = useQuery<AIConfidence>({
    queryKey: ['ai-confidence'],
    queryFn: async () => await call('AIOps.GetConfidenceScore') as AIConfidence,
    refetchInterval: refreshInterval,
  })

  const { data: recentMessages = [] } = useQuery<ConversationMessage[]>({
    queryKey: ['recent-conversations'],
    queryFn: async () => {
      const sessions = await call('AIOps.ListSessions') as { session_id: string; msg_count: number; last_active: string }[]
      if (!sessions || sessions.length === 0) return []
      const recent = sessions.sort((a, b) => new Date(b.last_active).getTime() - new Date(a.last_active).getTime()).slice(0, 3)
      const all: ConversationMessage[] = []
      for (const s of recent) {
        const msgs = await call('AIOps.GetMessages', s.session_id) as ConversationMessage[]
        if (msgs && msgs.length > 0) {
          all.push(...msgs.slice(-3))
        }
      }
      return all.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()).slice(0, 10)
    },
    refetchInterval: refreshInterval,
  })

  const isAvailable = ollamaStatus?.available ?? false
  const modelName = ollamaStatus?.model ?? 'N/A'
  const modelVersion = ollamaStatus?.version ?? ''
  const availableModels = ollamaStatus?.available_models ?? []

  return (
    <div className="h-full overflow-y-auto p-8 space-y-8">
      {/* ── AI Status ── */}
      <Panel variant="elevated" padding="lg" category="ai">
        <h3 className="text-xs font-black text-text uppercase tracking-[0.2em] mb-6 flex items-center gap-4">
          <BrainCircuit size={18} className="text-accent" />
          AI Engine Status
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-panel-2 border border-border/50 rounded-xl p-5 flex flex-col gap-2">
            <span className="text-[9px] font-black text-text-faint uppercase tracking-widest">Model</span>
            <span className="text-sm font-black text-text truncate">{modelName}</span>
            {modelVersion && <span className="text-[10px] font-semibold text-text-dim">{modelVersion}</span>}
          </div>
          <div className="bg-panel-2 border border-border/50 rounded-xl p-5 flex flex-col gap-2">
            <span className="text-[9px] font-black text-text-faint uppercase tracking-widest">Status</span>
            <div className="flex items-center gap-2 mt-1">
              <span className={cn("w-2.5 h-2.5 rounded-full", isAvailable ? "bg-success shadow-[0_0_8px_var(--color-success)]" : "bg-danger shadow-[0_0_8px_var(--color-danger)]")} />
              <span className={cn("text-sm font-black", isAvailable ? "text-success" : "text-danger")}>
                {isAvailable ? 'Online' : 'Offline'}
              </span>
            </div>
          </div>
          <div className="bg-panel-2 border border-border/50 rounded-xl p-5 flex flex-col gap-2">
            <span className="text-[9px] font-black text-text-faint uppercase tracking-widest">Available Models</span>
            <span className="text-2xl font-black text-text tabular-nums">{availableModels.length}</span>
          </div>
          <div className="bg-panel-2 border border-border/50 rounded-xl p-5 flex flex-col gap-2">
            <span className="text-[9px] font-black text-text-faint uppercase tracking-widest">Confidence</span>
            <span className={cn(
              "text-2xl font-black tabular-nums",
              (confidence?.overall ?? 0) >= 80 ? "text-success" :
              (confidence?.overall ?? 0) >= 50 ? "text-warning" : "text-danger"
            )}>
              {confidence ? `${Math.round(confidence.overall)}%` : '—'}
            </span>
          </div>
        </div>
        {confidence && confidence.factors && Object.keys(confidence.factors).length > 0 && (
          <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-3">
            {Object.entries(confidence.factors).map(([key, val]) => (
              <div key={key} className="flex items-center justify-between bg-panel-2/50 border border-border/30 rounded-lg px-4 py-2.5">
                <span className="text-[10px] font-bold text-text-faint uppercase tracking-wider">{key}</span>
                <span className={cn(
                  "text-[11px] font-black tabular-nums",
                  val >= 80 ? "text-success" : val >= 50 ? "text-warning" : "text-danger"
                )}>{Math.round(val)}%</span>
              </div>
            ))}
          </div>
        )}
      </Panel>

      {/* ── Recent Conversations ── */}
      <Panel variant="elevated" padding="lg" category="ai">
        <h3 className="text-xs font-black text-text uppercase tracking-[0.2em] mb-6 flex items-center gap-4">
          <MessageSquare size={18} className="text-accent" />
          Recent Conversations
          <button
            onClick={() => navigate('aiops')}
            className="ml-auto text-[10px] font-black text-accent uppercase tracking-wider hover:underline"
          >
            Open Chat →
          </button>
        </h3>
        {recentMessages.length === 0 ? (
          <div className="bg-panel-2 border border-border/50 rounded-xl p-8 text-center">
            <MessageSquare size={24} className="text-text-faint mx-auto mb-3 opacity-50" />
            <p className="text-xs font-bold text-text-faint">No conversations yet</p>
            <p className="text-[10px] text-text-dim mt-1">Start a chat with the AI Analyst to see history here.</p>
          </div>
        ) : (
          <div className="space-y-2">
            {recentMessages.map((msg, i) => (
              <div key={msg.id ?? i} className="flex items-start gap-3 bg-panel-2 border border-border/50 rounded-xl p-4 hover:border-accent/20 transition-all">
                <div className={cn(
                  "w-8 h-8 rounded-lg flex items-center justify-center shrink-0 border",
                  msg.role === 'assistant' ? "bg-accent/10 border-accent/20 text-accent" :
                  msg.role === 'system' ? "bg-warning/10 border-warning/20 text-warning" :
                  "bg-panel-3 border-border text-text-dim"
                )}>
                  {msg.role === 'assistant' ? <Bot size={14} /> : msg.role === 'system' ? <Activity size={14} /> : <User size={14} />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-[10px] font-black uppercase tracking-wider text-text-dim">{msg.role}</span>
                    <span className="text-[9px] text-text-faint tabular-nums">
                      {formatSafeDate(msg.timestamp, (d) => format(d, 'MMM d, HH:mm'))}
                    </span>
                  </div>
                  <p className="text-xs font-medium text-text-dim leading-relaxed line-clamp-2">{msg.content}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </Panel>

      {/* ── Generated Reports / Insights Summary ── */}
      <Panel variant="elevated" padding="lg" category="ai">
        <h3 className="text-xs font-black text-text uppercase tracking-[0.2em] mb-6 flex items-center gap-4">
          <Lightbulb size={18} className="text-accent" />
          Active Insights
          <button
            onClick={() => navigate('aiops')}
            className="ml-auto text-[10px] font-black text-accent uppercase tracking-wider hover:underline"
          >
            View All →
          </button>
        </h3>
        {insights.length === 0 ? (
          <div className="bg-panel-2 border border-border/50 rounded-xl p-8 text-center">
            <Lightbulb size={24} className="text-text-faint mx-auto mb-3 opacity-50" />
            <p className="text-xs font-bold text-text-faint">No active insights</p>
            <p className="text-[10px] text-text-dim mt-1">System is operating within normal parameters.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {insights.slice(0, 5).map((insight, i) => (
              <div
                key={insight.title + '-' + i}
                className={cn(
                  "flex items-start gap-4 p-4 rounded-xl border transition-all cursor-pointer hover:border-accent/30",
                  insight.severity === 'critical' ? "bg-danger/5 border-danger/20" :
                  insight.severity === 'warning' ? "bg-warning/5 border-warning/20" :
                  "bg-panel-2 border-border/50"
                )}
                onClick={() => insight.actionPage && navigate(insight.actionPage as Page)}
              >
                <div className={cn(
                  "mt-0.5 w-2 h-2 rounded-full shrink-0",
                  insight.severity === 'critical' ? "bg-danger shadow-[0_0_6px_var(--color-danger)]" :
                  insight.severity === 'warning' ? "bg-warning shadow-[0_0_6px_var(--color-warning)]" :
                  "bg-accent"
                )} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <span className="text-xs font-bold text-text">{insight.title}</span>
                    <span className={cn(
                      "text-[9px] font-black uppercase tracking-wider px-1.5 py-0.5 rounded",
                      insight.severity === 'critical' ? "bg-danger/10 text-danger" :
                      insight.severity === 'warning' ? "bg-warning/10 text-warning" :
                      "bg-accent/10 text-accent"
                    )}>{insight.severity}</span>
                    <span className="text-[9px] font-semibold text-text-faint uppercase tracking-wider">{insight.category}</span>
                  </div>
                  <p className="text-xs font-medium text-text-dim leading-relaxed">{insight.message}</p>
                  {insight.action && (
                    <p className="text-[10px] font-bold text-accent mt-1.5">→ {insight.action}</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </Panel>
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
  const [streamingContent, setStreamingContent] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const chatRef = useRef<HTMLDivElement>(null)

  const { data: sessions = [], refetch: refetchSessions } = useQuery<ChatSession[]>({
    queryKey: ['chat-sessions'],
    queryFn: async () => {
      const res = await call('AIOps.ListSessions')
      return (res as ChatSession[]) || []
    }
  })

  // ── Streaming Events ────────────────────────────────────────────
  useEvents('chat:token', (data: unknown) => {
    const d = data as { sessionId: string; token: string }
    if (d.sessionId === activeSession) {
      setStreamingContent(prev => (prev ?? '') + d.token)
    }
  })

  useEvents('chat:done', (data: unknown) => {
    const d = data as { sessionId: string; content: string }
    if (d.sessionId === activeSession) {
      const assistantMsg: ChatMessage = {
        role: 'assistant',
        content: d.content,
      }
      setMessages(prev => [...prev, assistantMsg])
      setStreamingContent(null)
      setIsTyping(false)
      refetchSessions()
    }
  })

  useEvents('chat:error', (data: unknown) => {
    const d = data as { sessionId: string; error: string }
    if (d.sessionId === activeSession) {
      setMessages(prev => [...prev, { role: 'assistant', content: `Error: ${d.error}` }])
      setStreamingContent(null)
      setIsTyping(false)
    }
  })
  // ───────────────────────────────────────────────────────────────

  // Load persisted messages when activeSession changes
  useEffect(() => {
    let cancelled = false
    const loadMessages = async () => {
      setLoaded(false)
      setStreamingContent(null)
      setIsTyping(false)
      try {
        const msgs = await call('AIOps.GetMessages', activeSession) as ChatMessage[]
        if (!cancelled && msgs && msgs.length > 0) {
          setMessages(msgs)
        } else if (!cancelled) {
          setMessages([{ role: 'assistant', content: 'Greetings. I am the Universal-Ops AI Analyst. I have analyzed your system metrics and identified several areas for potential optimization. How can I assist you today?' }])
        }
      } catch { /* ignore */
        if (!cancelled) {
          setMessages([{ role: 'assistant', content: 'Greetings. I am the Universal-Ops AI Analyst. I have analyzed your system metrics and identified several areas for potential optimization. How can I assist you today?' }])
        }
      } finally {
        if (!cancelled) setLoaded(true)
      }
    }
    loadMessages()
    return () => { cancelled = true }
  }, [call, activeSession])

  const handleSend = async () => {
    if (!input.trim() || isTyping) return

    const userMsg: ChatMessage = { role: 'user', content: input }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setIsTyping(true)
    setStreamingContent('') // ← signals frontend to show live bubble

    // Persist user message + fire streaming request (response arrives via events)
    try {
      await call('AIOps.SaveMessage', activeSession, 'user', userMsg.content)
      refetchSessions()
    } catch { toast.error('Failed to save message') }

    try {
      await call('AIOps.ChatStream', activeSession, input)
    } catch (err) {
      setMessages(prev => [...prev, { role: 'assistant', content: `Analyst Error: ${String(err)}` }])
      setStreamingContent(null)
      setIsTyping(false)
    }
  }

  const handleNewSession = () => {
    setActiveSession(`sess-${Date.now()}`)
    setStreamingContent(null)
  }

  const handleDeleteSession = (sid: string, e: React.MouseEvent) => {
    e.stopPropagation()
    setDeleteTarget(sid)
  }

  const confirmDeleteSession = async () => {
    if (!deleteTarget) return
    await call('AIOps.DeleteSession', deleteTarget)
    refetchSessions()
    if (activeSession === deleteTarget) {
      handleNewSession()
    }
    setDeleteTarget(null)
  }

  const cancelDelete = () => setDeleteTarget(null)

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
  }, [messages, isTyping, streamingContent])

  return (
    <div className="flex h-full bg-[var(--color-bg)] overflow-hidden">
      {/* Sessions Sidebar */}
      <div className="w-80 border-r border-[var(--color-border)] bg-[var(--color-panel)] flex flex-col shrink-0">
        <div className="p-8 border-b border-[var(--color-border)]">
          <button
            onClick={handleNewSession}
            className="w-full flex items-center justify-center gap-2.5 py-4 px-4 bg-accent text-white rounded-xl font-black text-xs uppercase tracking-widest hover:bg-accent/90 transition-all shadow-lg active:scale-95 group"
          >
            <Sparkles size={16} className="group-hover:rotate-12 transition-transform" />
            New Intelligence Session
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          {sessions.length === 0 ? (
            <div className="py-16 px-4 text-center">
              <div className="w-12 h-12 rounded-full bg-panel-3 border border-border flex items-center justify-center mx-auto mb-4 opacity-50">
                <MessageSquare size={20} className="text-text-faint" />
              </div>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em] mb-2">No History</p>
              <p className="text-xs text-text-dim font-medium leading-relaxed">Persistent sessions will appear here.</p>
            </div>
          ) : (
            sessions.map((s) => (
              <div
                key={s.session_id}
                onClick={() => setActiveSession(s.session_id)}
                className={cn(
                  "group relative p-5 rounded-xl border transition-all cursor-pointer",
                  activeSession === s.session_id
                    ? "bg-accent-soft border-accent/30 shadow-sm"
                    : "border-transparent hover:bg-[var(--color-sidebar-hover)]"
                )}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <p className={cn("text-sm font-bold truncate tracking-tight", activeSession === s.session_id ? "text-text" : "text-text-dim group-hover:text-text")}>
                      {s.session_id}
                    </p>
                    <div className="flex items-center gap-3 mt-2 text-[10px] font-black text-text-faint uppercase tracking-tighter">
                      <span className="flex items-center gap-1.5"><MessageSquare size={10} className="text-accent/60" /> {s.msg_count}</span>
                      <span className="opacity-30">|</span>
                      <span>{formatSafeDate(s.last_active, (d) => format(d, 'MMM d, HH:mm'))}</span>
                    </div>
                  </div>
                  <button
                    onClick={(e) => handleDeleteSession(s.session_id, e)}
                    className="opacity-0 group-hover:opacity-100 p-2 rounded-lg hover:bg-danger/10 hover:text-danger text-text-faint transition-all"
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
      <div className="flex-1 flex flex-row relative min-w-0">
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
                  <ChatBubble
                    key={msg.role + '-' + i}
                    role={msg.role}
                    content={msg.content}
                    actions={msg.actions}
                    sessionID={activeSession}
                    onAssistantReply={(content, actions) => {
                      setMessages(prev => [...prev, { role: 'assistant', content, actions }])
                    }}
                  />
                ))}
                {streamingContent !== null ? (
                  <ChatBubble
                    role="assistant"
                    content={streamingContent}
                    sessionID={activeSession}
                    onAssistantReply={() => {}}
                  />
                ) : isTyping ? (
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
                ) : null}
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
                          className="px-4 py-2 bg-panel border border-border rounded-full text-sm font-bold text-text-dim hover:text-accent hover:border-accent/40 transition-all active:scale-95"
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
        <ContextSidebar />
      </div>

      <ConfirmDialog
        open={deleteTarget !== null}
        title="Delete Session"
        description="Delete this session and all its messages? This action cannot be undone."
        type="danger"
        confirmText="Delete"
        onConfirm={confirmDeleteSession}
        onClose={cancelDelete}
      />
    </div>
  )
}

function ContextSidebar() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: dashboard } = useQuery<DashboardData>({
    queryKey: ['dashboard-mini'],
    queryFn: async () => await call('Dashboard.GetDashboardData') as DashboardData,
    refetchInterval: refreshInterval,
  })

  // ── AI Workflow Activity Feed ──
  const [workflowEvents, setWorkflowEvents] = useState<AIWorkflowEvent[]>([])
  const [wfExpanded, setWfExpanded] = useState(true)
  useEvents('ai:workflow', (data) => {
    setWorkflowEvents(prev => {
      const evt = data as AIWorkflowEvent
      // Update existing event if same stage+session is still running (status change)
      const idx = prev.findIndex(e => e.sessionId === evt.sessionId && e.stage === evt.stage && e.status === 'running')
      if (idx >= 0 && evt.status !== 'running') {
        const updated = [...prev]
        updated[idx] = evt
        return updated.slice(0, 100)
      }
      return [evt, ...prev].slice(0, 100)
    })
  })

  // ── Data Stream Metrics ──
  const [dsExpanded, setDsExpanded] = useState(true)
  const { data: streamMetricsRaw } = useQuery<DataStreamMetric[]>({
    queryKey: ['data-stream'],
    queryFn: async () => await call('AIOps.GetDataStreamSnapshot') as DataStreamMetric[],
    refetchInterval: refreshInterval,
  })
  const streamMetrics: DataStreamMetric[] = Array.isArray(streamMetricsRaw) ? streamMetricsRaw : []

  if (!dashboard) return null

  return (
    <div className="w-80 border-l border-border bg-panel flex flex-col shrink-0 animate-in slide-in-from-right-4 duration-500">
      {/* Header */}
      <div className="p-6 border-b border-border bg-panel-2/50">
        <h3 className="text-xs font-black text-text-faint uppercase tracking-[0.2em] flex items-center gap-2">
          <Activity size={14} className="text-accent" />
          Live Context
        </h3>
      </div>

      {/* Scrollable Content */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6">

        {/* ── System Gauges ── */}
        <div className="space-y-4">
          <ContextItem
            icon={<Cpu size={18} />}
            label="Processor"
            value={dashboard.cpu.value}
            unit="%"
            color={dashboard.cpu.value > 80 ? 'text-danger' : 'text-success'}
          />
          <ContextItem
            icon={<MemoryStick size={18} />}
            label="Memory"
            value={dashboard.memory.value}
            unit="%"
            color={dashboard.memory.value > 85 ? 'text-warning' : 'text-success'}
          />
          <ContextItem
            icon={<HardDrive size={18} />}
            label="Storage"
            value={dashboard.disk.value}
            unit="%"
            color={dashboard.disk.value > 90 ? 'text-danger' : 'text-success'}
          />
        </div>

        {/* ── Neural Awareness ── */}
        <div className="pt-6 border-t border-border">
          <h4 className="text-[10px] font-bold text-text-faint uppercase tracking-widest mb-4">Neural Awareness</h4>
          <div className="p-4 rounded-xl bg-accent-soft border border-accent/20 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-[10px] font-bold text-text-dim">ANOMALIES</span>
              <span className="text-[10px] font-black text-accent tabular-nums">{dashboard.alerts > 0 ? dashboard.alerts : 'NONE'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[10px] font-bold text-text-dim">UPTIME</span>
              <span className="text-[10px] font-black text-text tabular-nums">{dashboard.uptime}</span>
            </div>
          </div>
        </div>

        {/* ── AI Workflow Activity Feed ── */}
        <div className="pt-6 border-t border-border">
          <button
            onClick={() => setWfExpanded(!wfExpanded)}
            className="w-full flex items-center justify-between text-[10px] font-bold text-text-faint uppercase tracking-widest mb-4 hover:text-accent transition-colors"
          >
            <span className="flex items-center gap-2">
              <Workflow size={12} />
              Workflow Activity
              {workflowEvents.filter(e => e.status === 'running').length > 0 && (
                <span className="inline-block w-1.5 h-1.5 rounded-full bg-accent animate-pulse" />
              )}
            </span>
            {wfExpanded ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
          </button>

          {wfExpanded && (
            <div className="space-y-1 max-h-[240px] overflow-y-auto">
              {workflowEvents.length === 0 ? (
                <p className="text-[10px] text-text-faint italic">No recent AI operations — send a message to begin.</p>
              ) : (
                workflowEvents.map((evt, i) => (
                  <div
                    key={`${evt.sessionId}-${evt.stage}-${evt.timestamp}-${i}`}
                    className="flex items-start gap-2 py-1.5 px-2 rounded-lg hover:bg-panel-2 transition-colors"
                  >
                    {/* Status icon */}
                    <div className="mt-0.5 shrink-0">
                      {evt.status === 'running' ? (
                        <Loader2 size={12} className="text-accent animate-spin" />
                      ) : evt.status === 'completed' ? (
                        <div className="w-3 h-3 rounded-full bg-success/20 flex items-center justify-center">
                          <div className="w-1.5 h-1.5 rounded-full bg-success" />
                        </div>
                      ) : (
                        <div className="w-3 h-3 rounded-full bg-danger/20 flex items-center justify-center">
                          <div className="w-1.5 h-1.5 rounded-full bg-danger" />
                        </div>
                      )}
                    </div>
                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1.5">
                        <span className="text-[10px] font-bold text-text uppercase">{evt.stage}</span>
                        <span className={cn(
                          'text-[9px] font-semibold uppercase tracking-wider',
                          evt.status === 'running' ? 'text-accent' :
                          evt.status === 'completed' ? 'text-success' : 'text-danger'
                        )}>
                          {evt.status}
                        </span>
                      </div>
                      <p className="text-[9px] text-text-faint truncate">{evt.detail}</p>
                    </div>
                    {/* Timestamp */}
                    <span className="text-[8px] text-text-faint/60 tabular-nums shrink-0">
                      {formatSafeDate(evt.timestamp, (d) => format(d, 'HH:mm:ss'))}
                    </span>
                  </div>
                ))
              )}
            </div>
          )}
        </div>

        {/* ── Data Stream Metrics ── */}
        <div className="pt-6 border-t border-border">
          <button
            onClick={() => setDsExpanded(!dsExpanded)}
            className="w-full flex items-center justify-between text-[10px] font-bold text-text-faint uppercase tracking-widest mb-4 hover:text-accent transition-colors"
          >
            <span className="flex items-center gap-2">
              <BarChart3 size={12} />
              Data Stream
              {streamMetrics && <span className="text-[9px] font-normal text-text-faint">({streamMetrics.length})</span>}
            </span>
            {dsExpanded ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
          </button>

          {dsExpanded && (
            <div className="space-y-1 max-h-[240px] overflow-y-auto">
              {!streamMetrics || streamMetrics.length === 0 ? (
                <p className="text-[10px] text-text-faint italic">No data stream metrics available yet.</p>
              ) : (
                streamMetrics.map((m) => (
                  <div
                    key={m.name}
                    className="flex items-center justify-between py-1.5 px-2 rounded-lg hover:bg-panel-2 transition-colors"
                  >
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1.5">
                        <span className="text-[10px] font-bold text-text truncate">{m.name}</span>
                        <span className="text-[8px] text-text-faint/60 tabular-nums">{m.samples} pts</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span className="text-[11px] font-black text-text tabular-nums">
                          {m.lastValue.toFixed(1)}
                        </span>
                        <span className="text-[9px] text-text-faint">{m.unit}</span>
                      </div>
                    </div>
                    {/* Trend indicator */}
                    <div className={cn(
                      'shrink-0',
                      m.trend === 'rising' ? 'text-success' :
                      m.trend === 'falling' ? 'text-danger' : 'text-text-faint'
                    )}>
                      {m.trend === 'rising' ? (
                        <ArrowUp size={14} />
                      ) : m.trend === 'falling' ? (
                        <ArrowDown size={14} />
                      ) : (
                        <Minus size={14} />
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          )}
        </div>

        {/* ── Footer Quote ── */}
        <div className="pt-4">
          <p className="text-[10px] text-text-faint leading-relaxed italic">
            "Analyst is continuously monitoring these heuristics to ground its responses in physical reality."
          </p>
        </div>
      </div>
    </div>
  )
}

function ContextItem({ icon, label, value, unit, color }: { icon: React.ReactNode, label: string, value: number, unit: string, color: string }) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-text-dim">
          {icon}
          <span className="text-[10px] font-black uppercase tracking-wider">{label}</span>
        </div>
        <span className={cn("text-sm font-black tabular-nums", color)}>
          {Math.round(value)}{unit}
        </span>
      </div>
      <div className="h-1 bg-panel-3 rounded-full overflow-hidden border border-border/50">
        <div
          className={cn("h-full transition-all duration-500", color.replace('text-', 'bg-'))}
          style={{ width: `${value}%` }}
        />
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

  const { data: anomalies = [], isFetching, refetch } = useQuery<AnomalyInfo[]>({
    queryKey: ['anomalies'],
    queryFn: async () => {
      const res = await call('AIOps.DetectAnomalies') as AnomalyInfo[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="flex flex-col h-full space-y-6 p-10">
      <div className="flex items-center justify-between bg-[var(--color-panel)] border border-[var(--color-border)] px-8 py-6 rounded-2xl shadow-xl">
        <div className="flex items-center gap-10">
          <div className="flex items-center gap-5">
            <div className={cn("w-12 h-12 rounded-xl flex items-center justify-center border-2 shadow-lg transition-transform hover:scale-105", anomalies.length > 0 ? "bg-danger/10 border-danger/20 text-danger" : "bg-success/10 border-success/20 text-success")}>
              {anomalies.length > 0 ? <AlertTriangle size={24} /> : <ShieldCheck size={24} />}
            </div>
            <div>
              <p className="text-3xl font-black text-[var(--color-text)] tabular-nums tracking-tighter">{anomalies.length}</p>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em]">Detected Anomalies</p>
            </div>
          </div>
          <div className="w-px h-10 bg-border/50" />
          <div className="text-text-dim text-xs leading-relaxed max-w-sm font-medium">
            Statistical deviation tracking compares current live metrics against a 12-minute rolling window of history.
          </div>
        </div>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-3 px-6 py-3 bg-accent text-white rounded-xl hover:opacity-90 font-black text-xs uppercase tracking-widest transition-all shadow-lg active:scale-95"
        >
          <RefreshCw size={18} className={cn(isFetching && "animate-spin")} />
          Deep Scan Now
        </button>
      </div>

      <div className="flex-1 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-2xl overflow-hidden shadow-2xl relative">
        <div className="overflow-y-auto h-full p-4">
          {anomalies.length === 0 ? (
            <div className="flex items-center justify-center h-full py-20">
              <EmptyState
                icon={<ShieldCheck size={32} className="text-success" />}
                title="System Equilibrium Established"
                description="System operating within established statistical baseline. All metrics are within normal operating parameters."
              />
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-3">
              {anomalies.map((a, i) => (
                <div key={a.metric + '-' + a.timestamp} className="flex items-center justify-between p-5 bg-panel-2 border border-border/50 rounded-xl hover:border-accent/30 transition-all group">
                  <div className="flex items-center gap-6">
                    <div className="w-10 h-10 rounded-lg bg-panel-3 border border-border flex items-center justify-center text-warning group-hover:scale-110 transition-transform">
                       <Zap size={20} />
                    </div>
                    <div>
                      <h4 className="text-xs font-black text-text uppercase tracking-widest mb-1">{a.metric.replace('.percent', '').toUpperCase()}</h4>
                      <div className="flex items-center gap-3 text-[10px] font-bold text-text-faint uppercase tracking-tighter">
                        <span>OBSERVED: <span className="text-text tabular-nums">{a.value.toFixed(2)}%</span></span>
                        <span className="opacity-30">|</span>
                        <span>EXPECTED: <span className="text-text-dim tabular-nums">{a.expected.toFixed(2)}%</span></span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-8">
                    <div className="text-right">
                      <p className="text-lg font-black text-danger tabular-nums">+{a.deviation.toFixed(1)}σ</p>
                      <p className="text-[10px] font-black text-text-faint uppercase tracking-widest">Deviation</p>
                    </div>
                    <StatusBadge status={a.severity} />
                  </div>
                </div>
              ))}
            </div>
          )}
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
  const { navigate } = useNavigationStore()

  const { data: insights = [], isFetching, refetch } = useQuery<AIInsight[]>({
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
    <div className="flex flex-col h-full space-y-6 p-10">
      <div className="flex items-center justify-between bg-[var(--color-panel)] border border-[var(--color-border)] px-8 py-6 rounded-2xl shadow-xl">
        <div className="flex items-center gap-10">
          <div className="flex items-center gap-5">
            <div className={cn("w-12 h-12 rounded-xl flex items-center justify-center border-2 shadow-lg bg-accent/10 border-accent/20 text-accent transition-transform hover:rotate-12")}>
              <Lightbulb size={24} />
            </div>
            <div>
              <p className="text-3xl font-black text-[var(--color-text)] tabular-nums tracking-tighter">{insights.length}</p>
              <p className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em]">Active Insights</p>
            </div>
          </div>
          <div className="w-px h-10 bg-border/50" />
          <div className="text-text-dim text-xs leading-relaxed max-w-sm font-medium">
            Synthesized from anomaly detection, metric trends, and active alert analysis.
          </div>
        </div>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-3 px-6 py-3 bg-[var(--color-panel-3)] border border-[var(--color-border)] rounded-xl hover:bg-panel hover:border-accent/40 text-[var(--color-text)] font-black text-xs uppercase tracking-widest transition-all shadow-lg active:scale-95"
        >
          <RefreshCw size={18} className={cn(isFetching && "animate-spin")} />
          Refresh Logic
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-4 pr-2">
        {insights.length === 0 ? (
          <div className="flex items-center justify-center h-full py-20">
            <EmptyState
              icon={<Lightbulb size={32} className="text-accent" />}
              title="Equilibrium Confirmed"
              description="System is operating within normal parameters. Insights will appear when anomalies, trends, or alerts require attention."
            />
          </div>
        ) : (
          insights.map((insight, i) => {
            const cfg = severityConfig[insight.severity] || severityConfig.info
            const SevIcon = cfg.icon
            return (
              <div
                key={insight.title + '-' + insight.timestamp}
                className={cn(
                  'bg-[var(--color-panel)] border rounded-2xl p-6 transition-all hover:border-accent/30 hover:shadow-2xl group relative overflow-hidden',
                  cfg.border,
                )}
              >
                <div className="absolute top-0 right-0 w-32 h-32 bg-accent/5 rounded-bl-full pointer-events-none group-hover:bg-accent/10 transition-colors" />
                <div className="flex items-start gap-6 relative z-10">
                  <div className={cn('w-12 h-12 rounded-xl flex items-center justify-center border-2 shrink-0 transition-transform group-hover:scale-110', cfg.bg, cfg.border)}>
                    <SevIcon size={24} className={cfg.color} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-3">
                      <h3 className="text-lg font-bold text-[var(--color-text)] group-hover:text-accent transition-colors tracking-tight">
                        {insight.title}
                      </h3>
                      <span className={cn('px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border', categoryColors[insight.category] || categoryColors.general)}>
                        {insight.category}
                      </span>
                      <StatusBadge status={insight.severity} />
                    </div>
                    <p className="text-sm text-[var(--color-text-dim)] leading-relaxed mb-4 font-medium max-w-3xl">{insight.message}</p>
                    <button
                      onClick={() => insight.actionPage && navigate(insight.actionPage as Page)}
                      disabled={!insight.actionPage}
                      className="flex items-center gap-2 text-xs text-accent font-black uppercase tracking-widest group-hover:translate-x-1 transition-transform disabled:opacity-60 disabled:cursor-default enabled:cursor-pointer"
                    >
                      {insight.action}
                      <ChevronRight size={14} />
                    </button>
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
