import { Sun, Moon, Bell } from 'lucide-react'
import type { Page } from '../../App'
import { useThemeStore, useAlertStore } from '../../stores/useSettingsStore'

interface TopBarProps {
  currentPage: Page
}

const pageLabels: Record<Page, string> = {
  dashboard: 'Dashboard',
  sysops: 'System Operations',
  netops: 'Network Operations',
  secops: 'Security Operations',
  devops: 'Development Operations',
  aiops: 'AI Operations',
  'network-design': 'Network Design',
  logs: 'Log Viewer',
  settings: 'Settings',
}

export function TopBar({ currentPage }: TopBarProps) {
  const { theme, toggle } = useThemeStore()
  const alertCount = useAlertStore((s) => s.alertCount)

  return (
    <header
      className="h-[60px] flex items-center justify-between px-6 flex-shrink-0 border-b border-[var(--color-border)] bg-[var(--color-bg)]"
      style={{
        backdropFilter: 'blur(14px)',
        WebkitBackdropFilter: 'blur(14px)',
      }}
    >
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm">
        <span className="text-[var(--color-text-faint)]">Hawkward</span>
        <span className="text-[var(--color-text-faint)] text-[10px]">/</span>
        <span className="text-[var(--color-text)] font-medium">{pageLabels[currentPage]}</span>
      </div>

      {/* Right section */}
      <div className="flex items-center gap-4">
        {/* Live status dot */}
        <div className="flex items-center gap-2">
          <span
            className="w-2 h-2 rounded-full bg-[var(--color-success)] animate-pulse"
            style={{
              boxShadow: '0 0 8px var(--color-success)',
            }}
          />
          <span className="text-xs text-[var(--color-text-faint)] hidden sm:inline">
            All Systems Nominal
          </span>
        </div>

        {/* Notification bell with alert badge */}
        <button
          className="relative text-[var(--color-text-dim)] hover:text-[var(--color-text)] transition-colors"
          aria-label={`Notifications${alertCount > 0 ? `, ${alertCount} active alerts` : ''}`}
        >
          <Bell size={18} />
          {alertCount > 0 && (
            <span className="absolute -top-1.5 -right-1.5 min-w-[16px] h-4 px-1 rounded-full bg-[var(--color-danger)] text-[10px] font-bold text-white flex items-center justify-center leading-none">
              {alertCount > 99 ? '99+' : alertCount}
            </span>
          )}
        </button>

        {/* Theme toggle */}
        <button
          onClick={toggle}
          className="text-[var(--color-text-dim)] hover:text-[var(--color-text)] transition-colors"
          aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
        </button>
      </div>
    </header>
  )
}
