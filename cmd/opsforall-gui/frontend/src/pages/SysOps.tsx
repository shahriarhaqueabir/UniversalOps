import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Cpu, Server, MemoryStick, Disc, Activity, Settings, FileText,
  HardDrive, Users, BarChart3, Package, Calendar, Stethoscope,
  Zap, Monitor,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import type { CPUInfo, MemoryInfo, SystemInfo, DiskInfo } from '@/types'

// Tab imports
import { SystemInfoTab } from './SysOps/SystemInfoTab'
import { CpuTab } from './SysOps/CpuTab'
import { MemoryTab } from './SysOps/MemoryTab'
import { DiskTab } from './SysOps/DiskTab'
import { ProcessesTab } from './SysOps/ProcessesTab'
import { ServicesTab } from './SysOps/ServicesTab'
import { LogsTab } from './SysOps/LogsTab'
import { UsersTab } from './SysOps/UsersTab'
import { PerformanceTab } from './SysOps/PerformanceTab'
import { PackageManagerTab } from './SysOps/PackageManagerTab'
import { SchedulerTab } from './SysOps/SchedulerTab'
import { DiagnosticsTab } from './SysOps/DiagnosticsTab'
import { ActionsTab } from './SysOps/ActionsTab'

export { Bar } from './SysOps/CpuTab'

type SysOpsCategory = 'system-info' | 'cpu' | 'memory' | 'disk' | 'processes' | 'services' | 'logs' | 'storage' | 'users' | 'performance' | 'packages' | 'scheduler' | 'diagnostics' | 'actions'

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
  { id: 'performance', label: 'Performance', icon: <BarChart3 size={18} />, group: 'diagnosis' },
  { id: 'packages', label: 'Packages', icon: <Package size={18} />, group: 'diagnosis' },
  { id: 'scheduler', label: 'Scheduler', icon: <Calendar size={18} />, group: 'diagnosis' },
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
      <div className="space-y-4 animate-pulse">
        <div className="h-8 w-48 bg-[var(--color-panel-2)] rounded" />
        <div className="grid grid-cols-2 gap-4">
          <div className="h-32 bg-[var(--color-panel-2)] rounded" />
          <div className="h-32 bg-[var(--color-panel-2)] rounded" />
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
      case 'performance': return <PerformanceTab />
      case 'packages': return <PackageManagerTab />
      case 'scheduler': return <SchedulerTab />
      case 'diagnostics': return <DiagnosticsTab />
      case 'actions': return <ActionsTab />
      default: return <DiagnosticsTab />
    }
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      {/* Header */}
      <div className="border-b border-[var(--color-border)] bg-[var(--color-panel-2)] py-4 px-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-4">
              <Monitor size={32} className="text-[var(--color-accent)]" /> SYSTEM OPERATIONS
            </h1>
            <p className="text-[var(--color-text-dim)] text-sm mt-1">Inspection, diagnosis, and action for system administrators.</p>
            <DataFreshnessIndicator lastUpdated={cpuUpdatedAt ? new Date(cpuUpdatedAt) : null} className="mt-1" />
          </div>
        </div>
      </div>

      {/* Content: Sidebar + Main */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div className="w-56 border-r border-[var(--color-border)] bg-[var(--color-panel-2)] overflow-y-auto p-3">
          <CategoryGroup label="INSPECTION" categories={inspectionCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="DIAGNOSIS" categories={diagnosisCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="ACTION" categories={actionCategories} active={activeCategory} onSelect={setActiveCategory} />
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto p-8">
          {renderContent()}
        </div>
      </div>
    </div>
  )
}

function CategoryGroup({ label, categories, active, onSelect }: { label: string; categories: CategoryDef[]; active: SysOpsCategory; onSelect: (id: SysOpsCategory) => void }) {
  return (
    <div className="mb-4">
      <p className="text-[10px] font-bold text-[var(--color-text-faint)] uppercase tracking-widest px-3 mb-2">{label}</p>
      {categories.map(cat => (
        <button
          key={cat.id}
          onClick={() => onSelect(cat.id)}
          className={cn(
            'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-bold transition-all mb-0.5',
            active === cat.id ? 'bg-[var(--color-accent)] text-white' : 'text-[var(--color-text-dim)] hover:text-[var(--color-text)] hover:bg-[var(--color-sidebar-hover)]'
          )}
        >
          {cat.icon}
          {cat.label}
        </button>
      ))}
    </div>
  )
}
