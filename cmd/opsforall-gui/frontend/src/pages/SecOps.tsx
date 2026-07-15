import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Shield, LayoutDashboard, Users, Wifi, Monitor, AlertTriangle,
  HardHat, ClipboardCheck, Siren,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'

// Tab imports
import { OverviewTab } from './secops/OverviewTab'
import { IdentityTab } from './secops/IdentityTab'
import { NetworkSecurityTab } from './secops/NetworkSecurityTab'
import { EndpointTab } from './secops/EndpointTab'
import { EventsTab } from './secops/EventsTab'
import { HardeningTab } from './secops/HardeningTab'
import { AuditTab } from './secops/AuditTab'
import { ResponseTab } from './secops/ResponseTab'

type SecOpsCategory =
  | 'overview' | 'identity' | 'network' | 'endpoint'
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
  { id: 'network', label: 'Network Security', icon: <Wifi size={18} />, group: 'assessment' },
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
      case 'network': return <NetworkSecurityTab />
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
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      {/* Header */}
      <div className="border-b border-[var(--color-border)] bg-[var(--color-panel-2)] py-4 px-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-4">
              <Shield size={32} className="text-[var(--color-danger)]" /> SECURITY OPERATIONS
            </h1>
            <p className="text-[var(--color-text-dim)] text-sm mt-1">Threat surface analysis, access control, and endpoint protection status.</p>
            <DataFreshnessIndicator lastUpdated={secUpdatedAt ? new Date(secUpdatedAt) : null} className="mt-1" />
          </div>
        </div>
      </div>

      {/* Content: Sidebar + Main */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div className="w-56 border-r border-[var(--color-border)] bg-[var(--color-panel-2)] overflow-y-auto p-3">
          <CategoryGroup label="ASSESSMENT" categories={assessmentCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="DETECTION" categories={detectionCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="RESPONSE" categories={responseCategories} active={activeCategory} onSelect={setActiveCategory} />
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto p-8">
          {renderContent()}
        </div>
      </div>
    </div>
  )
}

function CategoryGroup({ label, categories, active, onSelect }: { label: string; categories: CategoryDef[]; active: SecOpsCategory; onSelect: (id: SecOpsCategory) => void }) {
  return (
    <div className="mb-4">
      <p className="text-[10px] font-bold text-[var(--color-text-faint)] uppercase tracking-widest px-3 mb-2">{label}</p>
      {categories.map(cat => (
        <button
          key={cat.id}
          onClick={() => onSelect(cat.id)}
          className={cn(
            'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-bold transition-all mb-0.5',
            active === cat.id ? 'bg-[var(--color-danger)] text-white' : 'text-[var(--color-text-dim)] hover:text-[var(--color-text)] hover:bg-[var(--color-sidebar-hover)]'
          )}
        >
          {cat.icon}
          {cat.label}
        </button>
      ))}
    </div>
  )
}
