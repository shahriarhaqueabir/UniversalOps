import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { formatSafeDate } from '@/lib/utils'
import { Sidebar } from './components/layout/Sidebar'
import { TopBar } from './components/layout/TopBar'
import { MainContent } from './components/layout/MainContent'
import { HawkSidebar } from './components/layout/HawkSidebar'
import { OnboardingModal } from './components/dialogs/OnboardingModal'
import { useThemeStore, useAlertStore, useSettingsStore, useNavigationStore, type Page } from './stores'
import { getGo, getRuntime } from './hooks/useBackend'
import type { AlertInfo } from './types'

function App() {
  const { currentPage, navigate } = useNavigationStore()
  const [onboarded, setOnboarded] = useState<boolean | null>(null)
  const [showHawk, setShowHawk] = useState(false)
  const { theme } = useThemeStore()
  const refreshInterval = useSettingsStore((s) => s.refreshInterval)
  const pingCount = useSettingsStore((s) => s.pingCount)
  const dnsTimeout = useSettingsStore((s) => s.dnsTimeout)
  const companionName = useSettingsStore((s) => s.companionName)
  const addAlert = useAlertStore((s) => s.addAlert)

  // Check onboarding status on mount
  useEffect(() => {
    let cancelled = false
    let retryTimer: ReturnType<typeof setTimeout> | null = null

    const checkOnboarded = async (attempt = 0) => {
      try {
        const app = getGo()
        const method = app?.App?.IsOnboarded
        if (method) {
          const res = await method()
          if (!cancelled) setOnboarded(res as boolean)
        } else if (attempt < 20) {
          // Retry up to 20 times (10s total) then assume onboarded
          retryTimer = setTimeout(() => checkOnboarded(attempt + 1), 500)
        } else {
          if (!cancelled) setOnboarded(true)
        }
      } catch {
        if (!cancelled) setOnboarded(true)
      }
    }
    checkOnboarded()

    return () => {
      cancelled = true
      if (retryTimer) clearTimeout(retryTimer)
    }
  }, [])

  // Sync all settings to backend whenever they change
  useEffect(() => {
    const syncSettings = async () => {
      try {
        const app = getGo()
        if (app?.PipelineAPI?.UpdateSettings) {
          await app.PipelineAPI.UpdateSettings(refreshInterval, 0, pingCount, dnsTimeout)
        }
        if (app?.AIOps?.SetCompanionName) {
          await app.AIOps.SetCompanionName(companionName)
        }
      } catch { /* ignore — Backend not ready yet */ }
    }
    syncSettings()
  }, [refreshInterval, pingCount, dnsTimeout, companionName])

  // Apply theme on mount and when it changes
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  // Global keyboard shortcuts
  useEffect(() => {
    const pages: Page[] = ['dashboard', 'sysops', 'workflows', 'netops', 'secops', 'devops', 'aiops', 'reports', 'alerts', 'logs', 'settings']
    const handler = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey) {
        const num = parseInt(e.key)
        if (num >= 1 && num <= pages.length) {
          e.preventDefault()
          navigate(pages[num - 1])
        }
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [navigate])

  // Subscribe to Wails alert events → sonner toasts
  const handleAlertEvent = useCallback((...args: unknown[]) => {
    const data = args[0] as Record<string, unknown> | undefined
    if (!data?.alert) return
    const alert = data.alert as AlertInfo
    addAlert(alert)

    // Show toast based on severity
    const level = alert.level?.toLowerCase() || 'info'
    const msg = `${alert.metric}: ${alert.message} (${Math.round(alert.value * 10) / 10})`

    const ts = formatSafeDate(alert.timestamp)
    switch (level) {
      case 'critical':
      case 'error':
        toast.error(msg, { description: ts })
        break
      case 'warning':
        toast.warning(msg, { description: ts })
        break
      default:
        toast.info(msg, { description: ts })
    }
  }, [addAlert])

  useEffect(() => {
    const runtime = getRuntime()
    if (runtime?.EventsOn) {
      runtime.EventsOn('alert', handleAlertEvent)

      return () => {
        runtime.EventsOff('alert', handleAlertEvent)
      }
    }
  }, [handleAlertEvent])

  // Show a neutral loading state while checking onboarding status
  if (onboarded === null) {
    return (
      <div className="flex h-screen w-screen items-center justify-center bg-[var(--color-bg)]">
        <div className="flex flex-col items-center gap-3 text-[var(--color-text-faint)]">
          <div className="w-8 h-8 rounded-lg bg-[var(--color-accent)]/20 flex items-center justify-center">
            <svg className="w-4 h-4 text-[var(--color-accent)] animate-spin" viewBox="0 0 24 24" fill="none">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
          </div>
          <p className="text-[10px] font-bold uppercase tracking-[0.15em]">Initializing...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--color-bg)] noise-overlay">
      {onboarded === false && <OnboardingModal onComplete={() => setOnboarded(true)} />}
      <Sidebar currentPage={currentPage} onNavigate={navigate} />
      <div className="flex-1 flex flex-col overflow-hidden relative">
        <TopBar currentPage={currentPage} onToggleHawk={() => setShowHawk(!showHawk)} />
        <div className="flex-1 flex overflow-hidden">
          <MainContent currentPage={currentPage} />
          <HawkSidebar isOpen={showHawk} onClose={() => setShowHawk(false)} />
        </div>
      </div>
    </div>
  )
}

export default App
