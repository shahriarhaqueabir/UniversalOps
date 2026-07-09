import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { Sidebar } from './components/layout/Sidebar'
import { TopBar } from './components/layout/TopBar'
import { MainContent } from './components/layout/MainContent'
import { useThemeStore, useAlertStore } from './stores/useSettingsStore'
import type { AlertInfo } from './types'

export type Page = 'dashboard' | 'sysops' | 'netops' | 'secops' | 'devops' | 'aiops' | 'network-design' | 'logs' | 'settings'

function getRuntime(): any {
  return (window as any)?.runtime
}

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard')
  const { theme } = useThemeStore()
  const addAlert = useAlertStore((s) => s.addAlert)

  // Apply theme on mount and when it changes
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  // Subscribe to Wails alert events → sonner toasts
  const handleAlertEvent = useCallback((data: any) => {
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
      return () => {
        runtime.EventsOff('alert', handleAlertEvent)
      }
    }
  }, [handleAlertEvent])

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--color-bg)]">
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <div className="flex-1 flex flex-col overflow-hidden">
        <TopBar currentPage={currentPage} />
        <MainContent currentPage={currentPage} onNavigate={setCurrentPage} />
      </div>
    </div>
  )
}

export default App
