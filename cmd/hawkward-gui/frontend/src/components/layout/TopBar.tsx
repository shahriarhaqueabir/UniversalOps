import { useState, useRef, useEffect } from 'react'
import { Sun, Moon, Bell, Trash2 } from 'lucide-react'
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
  const alerts = useAlertStore((s) => s.alerts)
  const clearAlerts = useAlertStore((s) => s.clearAlerts)
  const [showAlertPanel, setShowAlertPanel] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)

  // Close panel on outside click
  useEffect(() => {
    if (!showAlertPanel) return
    const handler = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setShowAlertPanel(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showAlertPanel])

  return (
    <header
      className="h-12 flex items-center justify-between px-4 flex-shrink-0 border-b border-[var(--color-border)] bg-[var(--color-bg)]"
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
      <div className="flex items-center gap-3">
        {/* Live status dot */}
        <div className="flex items-center gap-2">
          <span
            className="w-2 h-2 rounded-full bg-[var(--color-success)] animate-pulse"
            style={{
              boxShadow: '0 0 8px var(--color-success)',
            }}
          />
          <span className="text-xs text-[var(--color-text-faint)] hidden sm:inline">
            {alertCount > 0 ? `${alertCount} Alert${alertCount > 1 ? 's' : ''} Active` : 'All Systems Nominal'}
          </span>
        </div>

        {/* Notification bell with alert badge */}
        <div className="relative" ref={panelRef}>
          <button
            onClick={() => setShowAlertPanel(!showAlertPanel)}
            className="relative text-[var(--color-text-dim)] hover:text-[var(--color-text)] transition-colors p-1.5 rounded-lg hover:bg-[var(--color-sidebar-hover)]"
            aria-label={`Notifications${alertCount > 0 ? `, ${alertCount} active alerts` : ''}`}
            aria-expanded={showAlertPanel}
          >
            <Bell size={18} />
            {alertCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 min-w-[16px] h-4 px-1 rounded-full bg-[var(--color-danger)] text-[10px] font-bold text-white flex items-center justify-center leading-none">
                {alertCount > 99 ? '99+' : alertCount}
              </span>
            )}
          </button>

          {/* Alert dropdown panel */}
          {showAlertPanel && (
            <div className="absolute top-full right-0 mt-2 w-80 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl shadow-2xl z-50 overflow-hidden">
              <div className="p-3 border-b border-[var(--color-border)] flex items-center justify-between">
                <span className="text-sm font-bold text-[var(--color-text)]">Alerts</span>
                {alerts.length > 0 && (
                  <button
                    onClick={() => { clearAlerts(); setShowAlertPanel(false) }}
                    className="flex items-center gap-1 text-xs text-[var(--color-text-faint)] hover:text-[var(--color-danger)] transition-colors"
                  >
                    <Trash2 size={12} /> Clear All
                  </button>
                )}
              </div>
              <div className="max-h-64 overflow-y-auto">
                {alerts.length === 0 ? (
                  <div className="p-8 text-center text-sm text-[var(--color-text-faint)]">No active alerts</div>
                ) : (
                  alerts.slice(0, 20).map((a, i) => (
                    <div key={i} className="px-4 py-3 border-b border-[var(--color-border)] last:border-0 hover:bg-[var(--color-sidebar-hover)] transition-colors">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-warning)]" />
                        <span className="text-xs font-bold text-[var(--color-text)]">{a.metric}</span>
                      </div>
                      <div className="text-xs text-[var(--color-text-dim)] ml-3.5">{a.message}</div>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}
        </div>

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
