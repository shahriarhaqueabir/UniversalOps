import { Sun, Moon, Bell } from 'lucide-react'
import type { Page } from '../../App'
import { useTheme } from '../../hooks/useTheme'

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
  const { theme, toggle } = useTheme()

  return (
    <header
      className="h-[60px] flex items-center justify-between px-6 flex-shrink-0 border-b border-[var(--color-border)]"
      style={{
        background: 'var(--color-bg)',
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

        {/* Notification bell */}
        <button
          className="relative text-[var(--color-text-dim)] hover:text-[var(--color-text)] transition-colors"
          aria-label="Notifications"
        >
          <Bell size={18} />
          <span className="absolute -top-0.5 -right-0.5 w-[7px] h-[7px] rounded-full bg-[var(--color-danger)]" />
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
