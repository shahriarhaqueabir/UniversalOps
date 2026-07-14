import { Suspense, lazy, useEffect, useState } from 'react'
import { AnimatePresence, motion } from 'motion/react'
import type { Page } from '../../App'
import { ErrorBoundary } from '../ui/ErrorBoundary'

// Lazy load pages for code splitting
const Dashboard = lazy(() => import('../../pages/Dashboard').then(m => ({ default: m.Dashboard })))
const SysOps = lazy(() => import('../../pages/SysOps').then(m => ({ default: m.SysOps })))
const NetOps = lazy(() => import('../../pages/NetOps').then(m => ({ default: m.NetOps })))
const SecOps = lazy(() => import('../../pages/SecOps').then(m => ({ default: m.SecOps })))
const DevOps = lazy(() => import('../../pages/DevOps').then(m => ({ default: m.DevOps })))
const AIOps = lazy(() => import('../../pages/AIOps').then(m => ({ default: m.AIOps })))
const Logs = lazy(() => import('../../pages/Logs').then(m => ({ default: m.Logs })))
const Settings = lazy(() => import('../../pages/Settings').then(m => ({ default: m.Settings })))

interface MainContentProps {
  currentPage: Page
  onNavigate?: (page: Page) => void
}

const pageVariants = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -8 },
}

function PageSkeleton() {
  return (
    <div className="p-10 space-y-8 animate-pulse">
      <div className="flex items-center justify-between">
        <div className="space-y-3">
          <div className="h-8 w-72 bg-[var(--color-panel-2)] rounded-xl" />
          <div className="h-4 w-96 bg-[var(--color-panel-2)] rounded-lg" />
        </div>
        <div className="flex gap-3">
          <div className="h-10 w-36 bg-[var(--color-panel-2)] rounded-xl" />
          <div className="h-10 w-40 bg-[var(--color-panel-2)] rounded-xl" />
        </div>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {[1, 2, 3].map(i => (
          <div key={i} className="h-48 bg-[var(--color-panel-2)] rounded-[var(--radius-lg)]" />
        ))}
      </div>
      <div className="h-64 bg-[var(--color-panel-2)] rounded-[var(--radius-xl)]" />
    </div>
  )
}

export function MainContent({ currentPage, onNavigate }: MainContentProps) {
  const [prefersReduced, setPrefersReduced] = useState(false)

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    setPrefersReduced(mq.matches)
    const handler = (e: MediaQueryListEvent) => setPrefersReduced(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])
  const renderPage = () => {
    switch (currentPage) {
      case 'dashboard': return <Dashboard onNavigate={onNavigate} />
      case 'sysops': return <SysOps />
      case 'netops': return <NetOps />
      case 'secops': return <SecOps />
      case 'devops': return <DevOps />
      case 'aiops': return <AIOps />
      case 'logs': return <Logs />
      case 'settings': return <Settings />
      default: return <Dashboard />
    }
  }

  return (
    <main className="flex-1 overflow-y-auto bg-[var(--color-bg)] p-8">
      <ErrorBoundary>
        <Suspense fallback={<PageSkeleton />}>
          <AnimatePresence mode="wait">
            <motion.div
              key={currentPage}
              variants={pageVariants}
              initial="initial"
              animate="animate"
              exit="exit"
              transition={{ duration: prefersReduced ? 0 : 0.15, ease: 'easeOut' }}
            >
              {renderPage()}
            </motion.div>
          </AnimatePresence>
        </Suspense>
      </ErrorBoundary>
    </main>
  )
}
