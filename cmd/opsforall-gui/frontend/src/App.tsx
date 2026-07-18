import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { Sidebar } from './components/layout/Sidebar'
import { TopBar } from './components/layout/TopBar'
import { MainContent } from './components/layout/MainContent'
import { OnboardingModal } from './components/dialogs/OnboardingModal'
import { useThemeStore, useAlertStore, useSettingsStore, useMetricsStore } from './stores'
import type { AlertInfo, DashboardData } from './types'

export type Page = 'dashboard' | 'sysops' | 'netops' | 'secops' | 'devops' | 'aiops' | 'logs' | 'settings'

interface WailsRuntime {
  EventsOn: (event: string, handler: (...args: unknown[]) => void) => void
  EventsOff: (event: string, handler: (...args: unknown[]) => void) => void
}

function getRuntime(): WailsRuntime | null {
  const w = window as { runtime?: WailsRuntime }
  return w.runtime ?? null
}

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard')
  const [onboarded, setOnboarded] = useState<boolean | null>(null)
  const { theme } = useThemeStore()
  const refreshInterval = useSettingsStore((s) => s.refreshInterval)
  const pingCount = useSettingsStore((s) => s.pingCount)
  const dnsTimeout = useSettingsStore((s) => s.dnsTimeout)
  const addAlert = useAlertStore((s) => s.addAlert)
  const setMetrics = useMetricsStore((s) => s.setMetrics)

  // Check onboarding status on mount
  useEffect(() => {
    const checkOnboarded = async () => {
      try {
        const go = (window as any).go
        const method = go?.app?.App?.IsOnboarded
        if (method) {
          const res = await method()
          setOnboarded(res)
        } else {
          // Retry logic or wait for runtime
          setTimeout(checkOnboarded, 500)
        }
      } catch {
        setOnboarded(true)
      }
    }
    checkOnboarded()
  }, [])

  // Sync all settings to backend whenever they change
  useEffect(() => {
    const syncSettings = async () => {
      try {
        const go = (window as any).go
        if (go?.app?.PipelineAPI?.UpdateSettings) {
          await go.app.PipelineAPI.UpdateSettings(refreshInterval, 0, pingCount, dnsTimeout)
        }
      } catch { /* ignore — Backend not ready yet; synced on next Settings page visit */ }
    }
    syncSettings()
  }, [refreshInterval, pingCount, dnsTimeout])

  // Apply theme on mount and when it changes
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  // Global keyboard shortcuts
  useEffect(() => {
    const pages: Page[] = ['dashboard', 'sysops', 'netops', 'secops', 'devops', 'aiops', 'logs', 'settings']
    const handler = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey) {
        const num = parseInt(e.key)
        if (num >= 1 && num <= pages.length) {
          e.preventDefault()
          setCurrentPage(pages[num - 1])
        }
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  // Subscribe to Wails alert events → sonner toasts
  const handleAlertEvent = useCallback((...args: unknown[]) => {
    const data = args[0] as Record<string, unknown> | undefined
    if (!data?.alert) return
    const alert = data.alert as AlertInfo
    addAlert(alert)

    // Show toast based on severity
    const level = alert.level?.toLowerCase() || 'info'
    const msg = `${alert.metric}: ${alert.message} (${Math.round(alert.value * 10) / 10})`

    switch (level) {
      case 'critical':
      case 'error':
        toast.error(msg, { description: new Date(alert.timestamp).toLocaleString() })
        break
      case 'warning':
        toast.warning(msg, { description: new Date(alert.timestamp).toLocaleString() })
        break
      default:
        toast.info(msg, { description: new Date(alert.timestamp).toLocaleString() })
    }
  }, [addAlert])

  useEffect(() => {
    const runtime = getRuntime()
    if (runtime?.EventsOn) {
      runtime.EventsOn('alert', handleAlertEvent)

      const handleMetrics = (payload: unknown) => {
        const d = (payload as any)?.data ?? payload
        if (d && d.cpu) setMetrics(d as DashboardData)
      }
      runtime.EventsOn('metrics', handleMetrics)

      return () => {
        runtime.EventsOff('alert', handleAlertEvent)
        runtime.EventsOff('metrics', handleMetrics)
      }
    }
  }, [handleAlertEvent, setMetrics])

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--color-bg)] noise-overlay">
      {onboarded === false && <OnboardingModal onComplete={() => setOnboarded(true)} />}
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <div className="flex-1 flex flex-col overflow-hidden">
        <TopBar currentPage={currentPage} />
        <MainContent currentPage={currentPage} onNavigate={setCurrentPage} />
      </div>
    </div>
  )
}

export default App
