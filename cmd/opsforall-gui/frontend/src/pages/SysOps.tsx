import { useState, Suspense, lazy } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Cpu, Server, MemoryStick, Disc, Activity, Settings, FileText,
  HardDrive, Users, Stethoscope,
  Zap, Monitor,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import type { CPUInfo, MemoryInfo, SystemInfo, DiskInfo } from '@/types'

// Tab imports (lazy-loaded to reduce initial bundle)
const SystemInfoTab = lazy(() => import('./SysOps/SystemInfoTab').then(m => ({ default: m.SystemInfoTab })))
const CpuTab = lazy(() => import('./SysOps/CpuTab').then(m => ({ default: m.CpuTab })))
const MemoryTab = lazy(() => import('./SysOps/MemoryTab').then(m => ({ default: m.MemoryTab })))
const DiskTab = lazy(() => import('./SysOps/DiskTab').then(m => ({ default: m.DiskTab })))
const ProcessesTab = lazy(() => import('./SysOps/ProcessesTab').then(m => ({ default: m.ProcessesTab })))
const ServicesTab = lazy(() => import('./SysOps/ServicesTab').then(m => ({ default: m.ServicesTab })))
const LogsTab = lazy(() => import('./SysOps/LogsTab').then(m => ({ default: m.LogsTab })))
const UsersTab = lazy(() => import('./SysOps/UsersTab').then(m => ({ default: m.UsersTab })))
const DiagnosticsTab = lazy(() => import('./SysOps/DiagnosticsTab').then(m => ({ default: m.DiagnosticsTab })))
const ActionsTab = lazy(() => import('./SysOps/ActionsTab').then(m => ({ default: m.ActionsTab })))

type SysOpsCategory = 'system-info' | 'cpu' | 'memory' | 'disk' | 'processes' | 'services' | 'logs' | 'storage' | 'users' | 'diagnostics' | 'actions'

interface CategoryDef {
  id: SysOpsCategory
  label: string
  icon: React.ReactNode
  group: 'inspection' | 'diagnosis' | 'action'
}

const categories: CategoryDef[] = [
  { id: 'system-info', label: 'System Info', icon: <Server size={18} />, group: 'inspection' },
  { id: 'cpu', label: 'CPU', icon: <Cpu size={18} />, group: 'inspection' },
  { id: 'memory', label: 'Memory', icon: <MemoryStick size={18} />, group: 'inspection' },
  { id: 'disk', label: 'Disk', icon: <Disc size={18} />, group: 'inspection' },
  { id: 'processes', label: 'Processes', icon: <Activity size={18} />, group: 'inspection' },
  { id: 'services', label: 'Services', icon: <Settings size={18} />, group: 'inspection' },
  { id: 'logs', label: 'Logs', icon: <FileText size={18} />, group: 'inspection' },
  { id: 'storage', label: 'Storage', icon: <HardDrive size={18} />, group: 'inspection' },
  { id: 'users', label: 'Users', icon: <Users size={18} />, group: 'inspection' },
  { id: 'diagnostics', label: 'Diagnostics', icon: <Stethoscope size={18} />, group: 'diagnosis' },
  { id: 'actions', label: 'Actions', icon: <Zap size={18} />, group: 'action' },
]

export function SysOps() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [activeCategory, setActiveCategory] = useState<SysOpsCategory>('diagnostics')

  // Shared queries
  const { data: cpuInfo, dataUpdatedAt: cpuUpdatedAt } = useQuery<CPUInfo>({
    queryKey: ['sysops-cpu'],
    queryFn: async () => { const r = await call('SysOps.GetCPUInfo'); return r as CPUInfo },
    refetchInterval: refreshInterval,
  })

  const { data: memInfo } = useQuery<MemoryInfo>({
    queryKey: ['sysops-mem'],
    queryFn: async () => { const r = await call('SysOps.GetMemoryInfo'); return r as MemoryInfo },
    refetchInterval: refreshInterval,
  })

  const { data: sysInfo } = useQuery<SystemInfo>({
    queryKey: ['sysops-sys'],
    queryFn: async () => { const r = await call('SysOps.GetSystemInfo'); return r as SystemInfo },
    refetchInterval: refreshInterval,
  })

  const { data: diskInfo } = useQuery<DiskInfo>({
    queryKey: ['sysops-disk'],
    queryFn: async () => { const r = await call('SysOps.GetDiskInfo'); return r as DiskInfo },
    refetchInterval: refreshInterval,
  })

  if (!cpuInfo || !memInfo || !sysInfo || !diskInfo) {
    return (
      <div className="space-y-12 animate-pulse p-10">
        <div className="h-10 w-64 bg-[var(--color-panel-2)] rounded-xl" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="h-48 bg-[var(--color-panel-2)] rounded-[24px]" />
          <div className="h-48 bg-[var(--color-panel-2)] rounded-[24px]" />
        </div>
      </div>
    )
  }

  const inspectionCategories = categories.filter(c => c.group === 'inspection')
  const diagnosisCategories = categories.filter(c => c.group === 'diagnosis')
  const actionCategories = categories.filter(c => c.group === 'action')

  const renderContent = () => {
    switch (activeCategory) {
      case 'system-info': return <SystemInfoTab sysInfo={sysInfo} cpuInfo={cpuInfo} />
      case 'cpu': return <CpuTab cpuInfo={cpuInfo} />
      case 'memory': return <MemoryTab memInfo={memInfo} />
      case 'disk': return <DiskTab diskInfo={diskInfo} />
      case 'processes': return <ProcessesTab />
      case 'services': return <ServicesTab />
      case 'logs': return <LogsTab />
      case 'storage': return <DiskTab diskInfo={diskInfo} />
      case 'users': return <UsersTab />
      case 'diagnostics': return <DiagnosticsTab />
      case 'actions': return <ActionsTab />
      default: return <DiagnosticsTab />
    }
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500">
      {/* Header */}
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 flex items-center justify-between px-10">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
               <Monitor size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Subsystem Diagnostics</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">System Operations</p>
          <p className="text-[var(--color-text-dim)] text-xs font-semibold uppercase tracking-widest mt-2">Inspection, diagnosis, and action for system administrators</p>
          <DataFreshnessIndicator lastUpdated={cpuUpdatedAt ? new Date(cpuUpdatedAt) : null} className="mt-4" />
        </div>
      </div>

      {/* Content: Sidebar + Main */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div className="w-64 border-r border-[var(--color-border)] bg-[var(--color-panel)] overflow-y-auto p-4 scroll-smooth">
          <CategoryGroup label="INSPECTION" categories={inspectionCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="DIAGNOSIS" categories={diagnosisCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="ACTION" categories={actionCategories} active={activeCategory} onSelect={setActiveCategory} />
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto p-10 relative">
          <div className="absolute inset-0 overflow-hidden pointer-events-none opacity-10">
            <div className="absolute -top-24 -left-24 w-96 h-96 bg-accent rounded-full blur-[120px]" />
          </div>
          <Suspense fallback={<div className="flex items-center justify-center h-full text-text-faint text-xs font-black uppercase tracking-widest animate-pulse">Initializing Subsystem...</div>}>
            {renderContent()}
          </Suspense>
        </div>
      </div>
    </div>
  )
}

function CategoryGroup({ label, categories, active, onSelect }: { label: string; categories: CategoryDef[]; active: SysOpsCategory; onSelect: (id: SysOpsCategory) => void }) {
  return (
    <div className="mb-5">
      <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-[0.2em] px-2.5 mb-2.5">{label}</p>
      <div className="space-y-1">
        {categories.map(cat => (
          <button
            key={cat.id}
            onClick={() => onSelect(cat.id)}
            className={cn(
              'w-full flex items-center gap-3 px-2.5 py-2 rounded-xl text-sm font-bold transition-all active:scale-[0.97]',
              active === cat.id
                ? 'bg-accent text-white shadow-lg shadow-accent/20'
                : 'text-text-dim hover:text-text hover:bg-[var(--color-sidebar-hover)]'
            )}
          >
            <div className={cn("transition-colors", active === cat.id ? "text-white" : "text-accent")}>
              {cat.icon}
            </div>
            {cat.label}
          </button>
        ))}
      </div>
    </div>
  )
}
