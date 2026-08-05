import { Suspense, useEffect, useState } from 'react'
import { AnimatePresence, motion } from 'motion/react'
import type { Page } from '../../stores/useSettingsStore'
import { ErrorBoundary } from '../ui/ErrorBoundary'
import { lazyPages, preloadSuggestedPages } from '@/lib/pageRegistry'

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
        {/* static skeleton */}
        {[1, 2, 3].map(i => (
          <div key={i} className="h-48 bg-[var(--color-panel-2)] rounded-[var(--radius-lg)]" />
        ))}
      </div>
      <div className="h-64 bg-[var(--color-panel-2)] rounded-[var(--radius-xl)]" />
    </div>
  )
}

export function MainContent({ currentPage }: { currentPage: Page }) {
  const [prefersReduced, setPrefersReduced] = useState(false)

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    setPrefersReduced(mq.matches)
    const handler = (e: MediaQueryListEvent) => setPrefersReduced(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])

  useEffect(() => {
    const connection = (navigator as Navigator & {
      connection?: { saveData?: boolean; effectiveType?: string }
    }).connection
    if (connection?.saveData || connection?.effectiveType === 'slow-2g') return

    const timer = window.setTimeout(() => {
      preloadSuggestedPages(currentPage)
    }, 250)

    return () => window.clearTimeout(timer)
  }, [currentPage])

  const PageComponent = lazyPages[currentPage] ?? lazyPages.dashboard

  return (
    <main className="flex-1 overflow-y-auto bg-[var(--color-bg)]">
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
              <PageComponent />
            </motion.div>
          </AnimatePresence>
        </Suspense>
      </ErrorBoundary>
    </main>
  )
}
