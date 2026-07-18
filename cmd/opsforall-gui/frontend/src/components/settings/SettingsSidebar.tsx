import { Settings, BrainCircuit, Cpu, ShieldAlert, ScrollText } from 'lucide-react'
import { cn } from '@/lib/utils'

export type SettingsTab = 'general' | 'intelligence' | 'engine' | 'security' | 'journal'

interface SettingsSidebarProps {
  activeTab: SettingsTab
  onTabChange: (tab: SettingsTab) => void
}

const tabs: { id: SettingsTab; label: string; icon: any; color: string }[] = [
  { id: 'general', label: 'General', icon: Settings, color: 'text-[var(--color-text-faint)]' },
  { id: 'intelligence', label: 'Intelligence', icon: BrainCircuit, color: 'text-violet-400' },
  { id: 'engine', label: 'Engine', icon: Cpu, color: 'text-indigo-400' },
  { id: 'security', label: 'Security', icon: ShieldAlert, color: 'text-red-400' },
  { id: 'journal', label: 'Journal', icon: ScrollText, color: 'text-[var(--color-text-faint)]' },
]

export function SettingsSidebar({ activeTab, onTabChange }: SettingsSidebarProps) {
  return (
    <nav className="w-52 flex flex-col gap-1 p-3 border-r border-[var(--color-border)] bg-[var(--color-bg)]/30">
      <div className="mb-4 px-3 py-2">
        <h2 className="text-[10px] font-black uppercase tracking-[0.25em] text-[var(--color-text-faint)] opacity-50">
          Control Plane
        </h2>
      </div>
      {tabs.map((t) => {
        const Icon = t.icon
        const isActive = activeTab === t.id
        return (
          <button
            key={t.id}
            onClick={() => onTabChange(t.id)}
            className={cn(
              'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-bold transition-all group',
              isActive
                ? 'bg-[var(--color-accent)]/10 text-[var(--color-accent)] shadow-sm'
                : 'text-[var(--color-text-dim)] hover:bg-[var(--color-sidebar-hover)] hover:text-[var(--color-text)]',
            )}
          >
            <Icon
              size={18}
              className={cn('transition-colors', isActive ? t.color : 'text-[var(--color-text-faint)] group-hover:text-[var(--color-text-dim)]')}
            />
            {t.label}
          </button>
        )
      })}
    </nav>
  )
}
