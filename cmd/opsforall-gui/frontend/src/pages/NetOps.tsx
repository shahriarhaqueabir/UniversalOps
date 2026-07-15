import { useState, Suspense, lazy } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Network, LayoutDashboard, Activity, Globe, Cable, Wifi, Map, Search,
  Signal, Radio, Route, Radar, Stethoscope, ShieldCheck, ShieldOff,
  Compass, Zap, WifiOff, Globe2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'

// Tab imports (lazy-loaded to reduce initial bundle)
const OverviewTab = lazy(() => import('./netops/OverviewTab').then(m => ({ default: m.OverviewTab })))
const PingTab = lazy(() => import('./netops/PingTab').then(m => ({ default: m.PingTab })))
const DnsTab = lazy(() => import('./netops/DnsTab').then(m => ({ default: m.DnsTab })))
const ConnectionsTab = lazy(() => import('./netops/ConnectionsTab').then(m => ({ default: m.ConnectionsTab })))
const InterfacesTab = lazy(() => import('./netops/InterfacesTab').then(m => ({ default: m.InterfacesTab })))
const TracerouteTab = lazy(() => import('./netops/TracerouteTab').then(m => ({ default: m.TracerouteTab })))
const PortScanTab = lazy(() => import('./netops/PortScanTab').then(m => ({ default: m.PortScanTab })))
const BandwidthTab = lazy(() => import('./netops/BandwidthTab').then(m => ({ default: m.BandwidthTab })))
const ArpTab = lazy(() => import('./netops/ArpTab').then(m => ({ default: m.ArpTab })))
const RoutingTab = lazy(() => import('./netops/RoutingTab').then(m => ({ default: m.RoutingTab })))
const WifiTab = lazy(() => import('./netops/WifiTab').then(m => ({ default: m.WifiTab })))
const DnsAdvancedTab = lazy(() => import('./netops/DnsAdvancedTab').then(m => ({ default: m.DnsAdvancedTab })))
const MultiPingTab = lazy(() => import('./netops/MultiPingTab').then(m => ({ default: m.MultiPingTab })))
const HealthCheckTab = lazy(() => import('./netops/HealthCheckTab').then(m => ({ default: m.HealthCheckTab })))
const VpnTab = lazy(() => import('./netops/VpnTab').then(m => ({ default: m.VpnTab })))
const FirewallTab = lazy(() => import('./netops/FirewallTab').then(m => ({ default: m.FirewallTab })))
const DiscoveryTab = lazy(() => import('./netops/DiscoveryTab').then(m => ({ default: m.DiscoveryTab })))
const ActionsTab = lazy(() => import('./netops/ActionsTab').then(m => ({ default: m.ActionsTab })))

type NetOpsCategory =
  | 'overview' | 'connections' | 'interfaces' | 'arp' | 'routing' | 'wifi'
  | 'ping' | 'dns' | 'traceroute' | 'portscan' | 'bandwidth'
  | 'dns-advanced' | 'multi-ping' | 'health' | 'vpn'
  | 'firewall' | 'discovery' | 'actions'

interface CategoryDef {
  id: NetOpsCategory
  label: string
  icon: React.ReactNode
  group: 'inspection' | 'diagnosis' | 'action'
}

const categories: CategoryDef[] = [
  { id: 'overview', label: 'Overview', icon: <LayoutDashboard size={18} />, group: 'inspection' },
  { id: 'connections', label: 'Connections', icon: <Cable size={18} />, group: 'inspection' },
  { id: 'interfaces', label: 'Interfaces', icon: <Wifi size={18} />, group: 'inspection' },
  { id: 'arp', label: 'ARP Table', icon: <Radio size={18} />, group: 'inspection' },
  { id: 'routing', label: 'Routing', icon: <Route size={18} />, group: 'inspection' },
  { id: 'wifi', label: 'WiFi', icon: <WifiOff size={18} />, group: 'inspection' },
  { id: 'ping', label: 'Ping', icon: <Activity size={18} />, group: 'diagnosis' },
  { id: 'dns', label: 'DNS', icon: <Globe size={18} />, group: 'diagnosis' },
  { id: 'traceroute', label: 'Traceroute', icon: <Map size={18} />, group: 'diagnosis' },
  { id: 'portscan', label: 'Port Scan', icon: <Search size={18} />, group: 'diagnosis' },
  { id: 'bandwidth', label: 'Bandwidth', icon: <Signal size={18} />, group: 'diagnosis' },
  { id: 'dns-advanced', label: 'DNS Advanced', icon: <Globe2 size={18} />, group: 'diagnosis' },
  { id: 'multi-ping', label: 'Multi-Ping', icon: <Radar size={18} />, group: 'diagnosis' },
  { id: 'health', label: 'Health Check', icon: <Stethoscope size={18} />, group: 'diagnosis' },
  { id: 'vpn', label: 'VPN', icon: <ShieldCheck size={18} />, group: 'diagnosis' },
  { id: 'firewall', label: 'Firewall', icon: <ShieldOff size={18} />, group: 'action' },
  { id: 'discovery', label: 'Discovery', icon: <Compass size={18} />, group: 'action' },
  { id: 'actions', label: 'Actions', icon: <Zap size={18} />, group: 'action' },
]

export function NetOps() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [activeCategory, setActiveCategory] = useState<NetOpsCategory>('overview')

  const { dataUpdatedAt: ifacesUpdatedAt } = useQuery({
    queryKey: ['netops-interfaces-sidebar'],
    queryFn: async () => {
      const res = await call('NetOps.GetInterfaces')
      return res
    },
    refetchInterval: refreshInterval,
  })

  const renderContent = () => {
    switch (activeCategory) {
      case 'overview': return <OverviewTab />
      case 'connections': return <ConnectionsTab />
      case 'interfaces': return <InterfacesTab />
      case 'arp': return <ArpTab />
      case 'routing': return <RoutingTab />
      case 'wifi': return <WifiTab />
      case 'ping': return <PingTab />
      case 'dns': return <DnsTab />
      case 'traceroute': return <TracerouteTab />
      case 'portscan': return <PortScanTab />
      case 'bandwidth': return <BandwidthTab />
      case 'dns-advanced': return <DnsAdvancedTab />
      case 'multi-ping': return <MultiPingTab />
      case 'health': return <HealthCheckTab />
      case 'vpn': return <VpnTab />
      case 'firewall': return <FirewallTab />
      case 'discovery': return <DiscoveryTab />
      case 'actions': return <ActionsTab />
      default: return <OverviewTab />
    }
  }

  const inspectionCategories = categories.filter(c => c.group === 'inspection')
  const diagnosisCategories = categories.filter(c => c.group === 'diagnosis')
  const actionCategories = categories.filter(c => c.group === 'action')

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      {/* Header */}
      <div className="border-b border-[var(--color-border)] bg-[var(--color-panel-2)] py-4 px-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-4">
              <Network size={32} className="text-[var(--color-accent)]" /> NETWORK OPERATIONS
            </h1>
            <p className="text-[var(--color-text-dim)] text-sm mt-1">Inspection, diagnosis, and action for network engineering workflows.</p>
            <DataFreshnessIndicator lastUpdated={ifacesUpdatedAt ? new Date(ifacesUpdatedAt) : null} className="mt-1" />
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
          <Suspense fallback={<div className="flex items-center justify-center h-32 text-[var(--color-text-faint)] text-sm">Loading...</div>}>
            {renderContent()}
          </Suspense>
        </div>
      </div>
    </div>
  )
}

function CategoryGroup({ label, categories, active, onSelect }: { label: string; categories: CategoryDef[]; active: NetOpsCategory; onSelect: (id: NetOpsCategory) => void }) {
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
