import { useState, useRef, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Sun, Moon, Bell, RefreshCw, AlertTriangle, AlertOctagon, Info, CheckCircle2, Clock, Sparkles, ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import type { Page } from '../../stores/useSettingsStore'
import { useThemeStore, useSettingsStore, useNavigationStore } from '../../stores'
import { PAGE_METADATA } from '@/lib/appMetadata'
import type { AlertInfo } from '@/types'

interface TopBarProps {
  currentPage: Page
  onToggleHawk?: () => void
}

const severityIcon: Record<string, { icon: typeof Info; color: string }> = {
  critical: { icon: AlertOctagon, color: 'text-[var(--color-danger)]' },
  warning: { icon: AlertTriangle, color: 'text-[var(--color-warning)]' },
  info: { icon: Info, color: 'text-[var(--color-accent)]' },
}

export function TopBar({ currentPage, onToggleHawk }: TopBarProps) {
  const { theme, toggle } = useThemeStore()
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const { history, goBack } = useNavigationStore()
  const queryClient = useQueryClient()
  const [showAlertPanel, setShowAlertPanel] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)
  const pageMeta = PAGE_METADATA[currentPage]

  // ── Fetch active alerts from backend when panel is open ──
  const { data: activeAlerts = [], isLoading: alertsLoading } = useQuery<AlertInfo[]>({
    queryKey: ['alertAPI', 'active'],
    queryFn: async () => {
      const res = await call('AlertAPI.GetActiveAlerts')
      return (res as AlertInfo[]) || []
    },
    enabled: true,
    refetchInterval: refreshInterval,
  })
  const activeAlertCount = activeAlerts.length

  const resolveMutation = useMutation({
    mutationFn: async (id: string) => { await call('AlertAPI.ResolveAlert', id) },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alertAPI', 'active'] }),
  })

  const evaluateMutation = useMutation({
    mutationFn: async () => { await call('AlertAPI.EvaluateNow') },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alertAPI', 'active'] }),
  })

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
      className="h-12 flex items-center justify-between px-4 flex-shrink-0 border-b border-[var(--color-border)] bg-[var(--color-bg)]/80 backdrop-blur-md wails-drag"
    >
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm wails-no-drag">
        {history.length > 0 && (
          <button
            onClick={goBack}
            className="p-1 rounded-md text-[var(--color-text-dim)] hover:text-[var(--color-text)] hover:bg-[var(--color-sidebar-hover)] transition-all mr-1"
            title="Go Back"
          >
            <ArrowLeft size={16} />
          </button>
        )}
        <span className="text-[var(--color-text-faint)]">Universal-Ops</span>
        <span className="text-[var(--color-text-faint)] text-[10px]">/</span>
        <span className="text-[var(--color-text)] font-medium" title={pageMeta.description}>
          {pageMeta.title}
        </span>
      </div>

      {/* Right section */}
      <div className="flex items-center gap-3 wails-no-drag">
        {/* Live status dot */}
        <div className="flex items-center gap-2">
          <span
            className="w-2 h-2 rounded-full bg-[var(--color-success)] animate-pulse"
            style={{
              boxShadow: '0 0 8px var(--color-success)',
            }}
          />
          <span className="text-xs text-[var(--color-text-faint)] hidden sm:inline">
            {alertsLoading
              ? 'Checking alerts...'
              : activeAlertCount > 0
                ? `${activeAlertCount} Alert${activeAlertCount > 1 ? 's' : ''} Active`
                : 'All Systems Nominal'}
          </span>
        </div>

        {/* Notification bell with alert badge */}
        <div className="relative" ref={panelRef}>
          <button
            onClick={() => setShowAlertPanel(!showAlertPanel)}
            className="relative text-[var(--color-text-dim)] hover:text-[var(--color-text)] transition-colors p-1.5 rounded-lg hover:bg-[var(--color-sidebar-hover)]"
            aria-label={`Notifications${activeAlertCount > 0 ? `, ${activeAlertCount} active alerts` : ''}`}
            aria-expanded={showAlertPanel}
          >
            <Bell size={18} />
            {activeAlertCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 min-w-[16px] h-4 px-1 rounded-full bg-[var(--color-danger)] text-[10px] font-bold text-white flex items-center justify-center leading-none">
                {activeAlertCount > 99 ? '99+' : activeAlertCount}
              </span>
            )}
          </button>

          {/* Alert dropdown panel */}
          {showAlertPanel && (
            <div className="absolute top-full right-0 mt-2 w-96 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl shadow-2xl z-50 overflow-hidden">
              <div className="p-4 border-b border-[var(--color-border)] flex items-center justify-between">
                <span className="text-sm font-bold text-[var(--color-text)]">
                  Active Alerts
                  <span className="ml-2 text-xs font-normal text-[var(--color-text-faint)]">({activeAlertCount})</span>
                </span>
                <div className="flex items-center gap-1.5">
                  <button
                    onClick={() => evaluateMutation.mutate()}
                    disabled={evaluateMutation.isPending}
                    className="flex items-center gap-1 text-xs px-2 py-1 rounded-md text-[var(--color-text-faint)] hover:text-[var(--color-accent)] hover:bg-[var(--color-accent)]/10 transition-colors"
                    title="Force re-evaluation"
                    >
                    <RefreshCw size={12} className={cn(evaluateMutation.isPending && 'animate-spin')} /> Evaluate
                  </button>
                  <button
                    onClick={() => queryClient.invalidateQueries({ queryKey: ['alertAPI', 'active'] })}
                    className="flex items-center gap-1 text-xs px-2 py-1 rounded-md text-[var(--color-text-faint)] hover:text-[var(--color-danger)] hover:bg-[var(--color-danger)]/10 transition-colors"
                  >
                    <RefreshCw size={12} /> Refresh
                  </button>
                </div>
              </div>
              <div className="max-h-80 overflow-y-auto">
                {alertsLoading ? (
                  <div className="p-8 text-center text-sm text-[var(--color-text-faint)]">Loading...</div>
                ) : activeAlerts.length === 0 ? (
                  <div className="p-8 text-center">
                    <CheckCircle2 size={24} className="mx-auto mb-2 text-[var(--color-success)]" />
                    <p className="text-sm text-[var(--color-text-faint)]">No active alerts</p>
                  </div>
                ) : (
                  activeAlerts.map((a) => {
                    const sev = severityIcon[a.level] || severityIcon.info
                    const Icon = sev.icon
                    return (
                      <div key={a.id} className="px-4 py-3 border-b border-[var(--color-border)] last:border-0 hover:bg-[var(--color-sidebar-hover)] transition-colors">
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex items-start gap-2 min-w-0">
                            <Icon size={14} className={cn('mt-0.5 flex-shrink-0', sev.color)} />
                            <div className="min-w-0">
                              <div className="text-xs font-bold text-[var(--color-text)] truncate">{a.message}</div>
                              <div className="flex items-center gap-2 mt-1 text-[10px] text-[var(--color-text-faint)]">
                                <Clock size={10} />
                                {a.timestamp}
                                <span className="opacity-40">|</span>
                                <span className="font-mono uppercase">{a.metric}</span>
                              </div>
                            </div>
                          </div>
                          <button
                            onClick={() => resolveMutation.mutate(a.id)}
                            disabled={resolveMutation.isPending}
                            className="flex-shrink-0 text-[10px] px-2 py-1 rounded-md font-semibold text-[var(--color-success)] hover:bg-[var(--color-success)]/10 border border-[var(--color-success)]/20 transition-colors"
                          >
                            Resolve
                          </button>
                        </div>
                      </div>
                    )
                  })
                )}
              </div>
            </div>
          )}
        </div>

        {/* Theme toggle */}
        <button
          onClick={toggle}
          className="text-[var(--color-text-dim)] hover:text-[var(--color-text)] transition-colors p-1.5 rounded-lg hover:bg-[var(--color-sidebar-hover)]"
          aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
        </button>

        {/* Hawk Toggle */}
        <button
          onClick={onToggleHawk}
          className="text-[var(--color-text-dim)] hover:text-[var(--color-accent)] transition-colors p-1.5 rounded-lg hover:bg-[var(--color-accent)]/10"
          aria-label="Toggle Hawk Companion"
        >
          <Sparkles size={18} />
        </button>
      </div>
    </header>
  )
}
