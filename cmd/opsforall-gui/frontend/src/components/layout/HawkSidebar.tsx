import { X, Send, Sparkles } from 'lucide-react'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { cn } from '@/lib/utils'

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
        <button
          onClick={onClose}
          className="p-1.5 rounded-lg hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-all"
        >
          <X size={18} />
        </button>
      </div>

      {/* Chat Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        <div className="flex gap-3">
          <div className="w-8 h-8 rounded-full bg-[var(--color-accent)]/20 border border-[var(--color-accent)]/30 flex items-center justify-center shrink-0">
            <Sparkles size={14} className="text-[var(--color-accent)]" />
          </div>
          <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl p-3 shadow-sm">
            <p className="text-xs leading-relaxed text-[var(--color-text)]">
              Hello! I am <span className="font-bold text-[var(--color-accent)]">{companionName}</span>, your operations co-pilot.
              I can help you analyze system health, optimize collection intervals, or explain security events.
            </p>
          </div>
        </div>
      </div>

      {/* Footer / Input */}
      <div className="p-4 border-t border-[var(--color-border)] bg-[var(--color-panel-3)]">
        <div className="relative group">
          <input
            type="text"
            placeholder={`Message ${companionName}...`}
            className="w-full bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl pl-4 pr-10 py-2.5 text-xs text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] focus:ring-1 focus:ring-[var(--color-accent)] transition-all"
          />
          <button
            className="absolute right-2 top-1.5 p-1.5 rounded-lg bg-[var(--color-accent)] text-white hover:opacity-90 transition-all active:scale-95"
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
