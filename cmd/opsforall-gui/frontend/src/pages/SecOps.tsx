import { useState, Suspense, lazy } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Shield, LayoutDashboard, Users, Wifi, Monitor, AlertTriangle,
  HardHat, ClipboardCheck, Siren,
} from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import { CategoryGroup } from '@/components/ui/CategoryGroup'

// Tab imports (lazy-loaded to reduce initial bundle)
const OverviewTab = lazy(() => import('./secops/OverviewTab').then(m => ({ default: m.OverviewTab })))
const IdentityTab = lazy(() => import('./secops/IdentityTab').then(m => ({ default: m.IdentityTab })))
const PerimeterTab = lazy(() => import('./secops/PerimeterTab').then(m => ({ default: m.PerimeterTab })))
const EndpointTab = lazy(() => import('./secops/EndpointTab').then(m => ({ default: m.EndpointTab })))
const EventsTab = lazy(() => import('./secops/EventsTab').then(m => ({ default: m.EventsTab })))
const HardeningTab = lazy(() => import('./secops/HardeningTab').then(m => ({ default: m.HardeningTab })))
const AuditTab = lazy(() => import('./secops/AuditTab').then(m => ({ default: m.AuditTab })))
const ResponseTab = lazy(() => import('./secops/ResponseTab').then(m => ({ default: m.ResponseTab })))

type SecOpsCategory =
  | 'overview' | 'identity' | 'perimeter' | 'endpoint'
  | 'events' | 'hardening'
  | 'audit' | 'response'

interface CategoryDef {
  id: SecOpsCategory
  label: string
  icon: React.ReactNode
  group: 'assessment' | 'detection' | 'response'
}

const categories: CategoryDef[] = [
  { id: 'overview', label: 'Overview', icon: <LayoutDashboard size={18} />, group: 'assessment' },
  { id: 'identity', label: 'Identity & Access', icon: <Users size={18} />, group: 'assessment' },
  { id: 'perimeter', label: 'Perimeter Security', icon: <Wifi size={18} />, group: 'assessment' },
  { id: 'endpoint', label: 'Endpoint Security', icon: <Monitor size={18} />, group: 'assessment' },
  { id: 'events', label: 'Log & Events', icon: <AlertTriangle size={18} />, group: 'detection' },
  { id: 'hardening', label: 'Security Hardening', icon: <HardHat size={18} />, group: 'detection' },
  { id: 'audit', label: 'Security Audit', icon: <ClipboardCheck size={18} />, group: 'response' },
  { id: 'response', label: 'Incident Response', icon: <Siren size={18} />, group: 'response' },
]

export function SecOps() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [activeCategory, setActiveCategory] = useState<SecOpsCategory>('overview')

  const { dataUpdatedAt: secUpdatedAt } = useQuery({
    queryKey: ['secops-health'],
    queryFn: async () => {
      const res = await call('SecOps.GetSecurityScore')
      return res
    },
    refetchInterval: refreshInterval,
    staleTime: 15000,
  })

  const renderContent = () => {
    switch (activeCategory) {
      case 'overview': return <OverviewTab />
      case 'identity': return <IdentityTab />
      case 'perimeter': return <PerimeterTab />
      case 'endpoint': return <EndpointTab />
      case 'events': return <EventsTab />
      case 'hardening': return <HardeningTab />
      case 'audit': return <AuditTab />
      case 'response': return <ResponseTab />
      default: return <OverviewTab />
    }
  }

  const assessmentCategories = categories.filter(c => c.group === 'assessment')
  const detectionCategories = categories.filter(c => c.group === 'detection')
  const responseCategories = categories.filter(c => c.group === 'response')

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500">
      {/* Header */}
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 flex items-center justify-between px-10">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-danger/10 flex items-center justify-center text-danger border border-danger/20">
               <Shield size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Threat & Perimeter</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">Security Operations</p>
          <p className="text-[var(--color-text-dim)] text-xs font-semibold uppercase tracking-widest mt-2">Threat surface analysis, access control, and endpoint protection status.</p>
          <DataFreshnessIndicator lastUpdated={secUpdatedAt ? new Date(secUpdatedAt) : null} className="mt-4" />
        </div>
      </div>

      {/* Content: Sidebar + Main */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div className="w-64 border-r border-[var(--color-border)] bg-[var(--color-panel)] overflow-y-auto p-4 scroll-smooth">
          <CategoryGroup label="ASSESSMENT" group="assessment" page="secops" categories={assessmentCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="DETECTION" group="detection" page="secops" categories={detectionCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="RESPONSE" group="response" page="secops" categories={responseCategories} active={activeCategory} onSelect={setActiveCategory} />
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto p-10 relative">
          <div className="absolute inset-0 overflow-hidden pointer-events-none opacity-10">
            <div className="absolute -top-24 -left-24 w-96 h-96 bg-danger rounded-full blur-[120px]" />
          </div>
          <Suspense fallback={<div className="flex items-center justify-center h-full text-text-faint text-xs font-black uppercase tracking-widest animate-pulse">Initializing Security Core...</div>}>
            {renderContent()}
          </Suspense>
        </div>
      </div>
    </div>
  )
}


