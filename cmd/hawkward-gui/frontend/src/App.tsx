import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import { Sidebar } from './components/layout/Sidebar'
import { TopBar } from './components/layout/TopBar'
import { MainContent } from './components/layout/MainContent'
import { useThemeStore, useAlertStore } from './stores/useSettingsStore'
import type { AlertInfo } from './types'

export type Page = 'dashboard' | 'sysops' | 'netops' | 'secops' | 'devops' | 'aiops' | 'network-design' | 'logs' | 'settings'

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
  const { theme } = useThemeStore()
  const addAlert = useAlertStore((s) => s.addAlert)

  // Apply theme on mount and when it changes
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  // Global keyboard shortcuts
  useEffect(() => {
    const pages: Page[] = ['dashboard', 'sysops', 'netops', 'secops', 'devops', 'aiops', 'network-design', 'logs', 'settings']
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
      return () => {
        runtime.EventsOff('alert', handleAlertEvent)
      }
    }
  }, [handleAlertEvent])

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--color-bg)] noise-overlay">
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <div className="flex-1 flex flex-col overflow-hidden">
        <TopBar currentPage={currentPage} />
        <MainContent currentPage={currentPage} onNavigate={setCurrentPage} />
      </div>
    </div>
  )
}

export default App
