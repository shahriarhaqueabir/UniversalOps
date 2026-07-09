import { Suspense, lazy } from 'react'
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
const NetworkDesign = lazy(() => import('../../pages/NetworkDesign').then(m => ({ default: m.NetworkDesign })))
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

export function MainContent({ currentPage, onNavigate }: MainContentProps) {
  const renderPage = () => {
    switch (currentPage) {
      case 'dashboard': return <Dashboard onNavigate={onNavigate} />
      case 'sysops': return <SysOps />
      case 'netops': return <NetOps />
      case 'secops': return <SecOps />
      case 'devops': return <DevOps />
      case 'aiops': return <AIOps />
      case 'network-design': return <NetworkDesign />
      case 'logs': return <Logs />
      case 'settings': return <Settings />
      default: return <Dashboard />
    }
  }

  return (
    <main className="flex-1 overflow-y-auto bg-[var(--color-bg)]">
      <ErrorBoundary>
        <Suspense fallback={
          <div className="flex h-full items-center justify-center">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-accent border-t-transparent" />
          </div>
        }>
          <AnimatePresence mode="wait">
            <motion.div
              key={currentPage}
              variants={pageVariants}
              initial="initial"
              animate="animate"
              exit="exit"
              transition={{ duration: 0.15, ease: 'easeOut' }}
            >
              {renderPage()}
            </motion.div>
          </AnimatePresence>
        </Suspense>
      </ErrorBoundary>
    </main>
  )
}
