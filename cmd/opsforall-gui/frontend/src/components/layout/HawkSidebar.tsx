import { X, Send, Sparkles, Zap, Loader2, Monitor } from 'lucide-react'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useState } from 'react'
import { ProposalCard } from './ProposalCard'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  proposal?: any
}

interface HawkSidebarProps {
  isOpen: boolean
  onClose: () => void
}

/**
 * HawkSidebar — Global companion panel.
 * Provides a dedicated space for "Hawk" (or custom-named) AI interactions.
 */
export function HawkSidebar({ isOpen, onClose }: HawkSidebarProps) {
  const { companionName } = useSettingsStore()
  const { call } = useBackend()
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content: `Hello! I am ${companionName}, your operations co-pilot. I can help you analyze system health, optimize collection intervals, or explain security events.`,
    },
  ])
  const [loading, setLoading] = useState(false)
  const [input, setInput] = useState('')

  const handleSend = async () => {
    if (!input.trim() || loading) return

    const userMsg: Message = { id: Date.now().toString(), role: 'user', content: input }
    setMessages((prev) => [...prev, userMsg])
    setInput('')
    setLoading(true)

    try {
      const res = await call('AIOps.Chat', input) as { content: string; actions?: any }
      setMessages((prev) => [
        ...prev,
        {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: res.content,
        },
      ])
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        { id: (Date.now() + 1).toString(), role: 'assistant', content: 'Chat failed: ' + err },
      ])
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleOptimize = async () => {
    setLoading(true)
    try {
      const res = await call('AIOps.RequestOptimization') as { content: string; payload: any }
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          role: 'assistant',
          content: res.content,
          proposal: res.payload,
        },
      ])
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        { id: Date.now().toString(), role: 'assistant', content: 'Optimization analysis failed: ' + err },
      ])
    } finally {
      setLoading(false)
    }
  }

  return (
    <aside
      className={cn(
        'fixed inset-y-0 right-0 z-40 w-80 bg-[var(--color-panel)] border-l border-[var(--color-border)] shadow-2xl flex flex-col transition-transform duration-300 ease-in-out transform',
        isOpen ? 'translate-x-0' : 'translate-x-full',
      )}
    >
      {/* Header */}
      <div className="h-12 px-4 border-b border-[var(--color-border)] flex items-center justify-between bg-[var(--color-bg)]/50 backdrop-blur-md">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-lg bg-[var(--color-accent)]/10 flex items-center justify-center text-[var(--color-accent)] border border-[var(--color-accent)]/20">
            <Sparkles size={16} />
          </div>
          <h2 className="text-sm font-bold uppercase tracking-[0.15em] text-[var(--color-text)]">
            {companionName}
          </h2>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={handleOptimize}
            disabled={loading}
            className="p-1.5 rounded-lg hover:bg-[var(--color-accent)]/10 text-[var(--color-accent)] transition-all title='Optimize Engine'"
          >
            {loading ? <Loader2 size={16} className="animate-spin" /> : <Zap size={16} />}
          </button>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-all"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      {/* Chat Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-6">
        {messages.map((m) => (
          <div key={m.id} className="space-y-4">
            <div className="flex gap-3">
              <div className={cn(
                "w-8 h-8 rounded-full border flex items-center justify-center shrink-0",
                m.role === 'assistant'
                  ? "bg-[var(--color-accent)]/20 border-[var(--color-accent)]/30"
                  : "bg-[var(--color-panel-3)] border-[var(--color-border)]"
              )}>
                {m.role === 'assistant' ? <Sparkles size={14} className="text-[var(--color-accent)]" /> : <Monitor size={14} className="text-[var(--color-text-faint)]" />}
              </div>
              <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl p-3 shadow-sm max-w-[85%]">
                <p className="text-xs leading-relaxed text-[var(--color-text)] whitespace-pre-wrap">
                  {m.content}
                </p>
              </div>
            </div>
            {m.proposal && (
              <div className="ml-11">
                <ProposalCard reasoning={m.content} payload={m.proposal} />
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Footer / Input */}
      <div className="p-4 border-t border-[var(--color-border)] bg-[var(--color-panel-3)]">
        <div className="relative group">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={loading}
            placeholder={`Message ${companionName}...`}
            className="w-full bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl pl-4 pr-10 py-2.5 text-xs text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] focus:ring-1 focus:ring-[var(--color-accent)] transition-all disabled:opacity-50"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || loading}
            className="absolute right-2 top-1.5 p-1.5 rounded-lg bg-[var(--color-accent)] text-white hover:opacity-90 transition-all active:scale-95 disabled:opacity-50"
          >
            <Send size={14} />
          </button>
        </div>
        <p className="text-[10px] text-[var(--color-text-faint)] mt-2 text-center italic">
          Hawk AI is currently in "Observer" mode.
        </p>
      </div>
    </aside>
  )
}
