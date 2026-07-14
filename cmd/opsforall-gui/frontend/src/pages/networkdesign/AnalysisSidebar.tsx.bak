import { useMemo, useState, useEffect } from 'react'
import { cn } from '@/lib/utils'
import {
  Activity,
  AlertTriangle,
  CheckCircle,
  CheckCircle2,
  ChevronRight,
  ChevronLeft,
  Cpu,
  FileText,
  Globe,
  HardDrive,
  Lightbulb,
  ListChecks,
  Monitor,
  Network,
  Radar,
  Server,
  Shield,
  Wifi,
  X,
} from 'lucide-react'
import type { TopologyDevice, TopologyConnection, TopologyHealth } from '@/types'
import { useBackend } from '@/hooks/useBackend'
import { useQuery } from '@tanstack/react-query'
import {
  groupByType,
  findDocumentationGaps,
  computeHealthScore,
  discoverLocalNetwork,
} from './topologyApi'

// ── Types ──

interface InventorySummary {
  missingIPs: number
  missingMACs: number
  vendors?: Record<string, number>
  types?: Record<string, number>
}

// ── Device type icon mapping ──

const deviceIcons: Record<string, React.ReactNode> = {
  router: <Server size={14} />,
  switch: <Network size={14} />,
  server: <HardDrive size={14} />,
  workstation: <Monitor size={14} />,
  firewall: <Shield size={14} />,
  cloud: <Globe size={14} />,
}

function getDeviceIcon(type: string): React.ReactNode {
  return deviceIcons[type] ?? <Cpu size={14} />
}

// ── Collapsible section ──

function Section({
  title,
  icon,
  count,
  countColor,
  defaultOpen,
  children,
}: {
  title: string
  icon: React.ReactNode
  count?: number
  countColor?: string
  defaultOpen?: boolean
  children: React.ReactNode
}) {
  const [open, setOpen] = useState(defaultOpen ?? true)

  return (
    <div className="border-b border-[var(--color-border)]">
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center gap-2 px-3 py-2.5 text-left hover:bg-[var(--color-sidebar-hover)] transition-colors"
      >
        <span className="text-[var(--color-text-faint)]">{icon}</span>
        <span className="text-xs font-semibold text-[var(--color-text)] uppercase tracking-wider flex-1">
          {title}
        </span>
        {count !== undefined && (
          <span
            className={cn(
              'text-xs font-mono px-1.5 py-0.5 rounded',
              countColor ?? 'text-[var(--color-text-faint)] bg-[var(--color-panel-3)]',
            )}
          >
            {count}
          </span>
        )}
        <ChevronRight
          size={12}
          className={cn(
            'text-[var(--color-text-faint)] transition-transform',
            open && 'rotate-90',
          )}
        />
      </button>
      {open && <div className="px-3 pb-3 space-y-2">{children}</div>}
    </div>
  )
}

// ── Stat row ──

function StatRow({
  label,
  value,
  color,
}: {
  label: string
  value: string | number
  color?: string
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-xs text-[var(--color-text-dim)]">{label}</span>
      <span className={cn('text-xs font-mono font-semibold', color ?? 'text-[var(--color-text)]')}>
        {value}
      </span>
    </div>
  )
}

// ── Main sidebar component ──

interface AnalysisSidebarProps {
  devices: TopologyDevice[]
  connections: TopologyConnection[]
  isOpen: boolean
  onToggle: () => void
  onDevicesDiscovered: (devices: TopologyDevice[]) => void
}

export function AnalysisSidebar({
  devices,
  connections,
  isOpen,
  onToggle,
  onDevicesDiscovered,
}: AnalysisSidebarProps) {
  const { call } = useBackend()
  const [discovering, setDiscovering] = useState(false)

  // ── Sync topology to backend ──
  useEffect(() => {
    const sync = async () => {
      try {
        await call('NetDesign.SetTopology', JSON.stringify(devices), JSON.stringify(connections))
      } catch (err: unknown) {
        console.error('[NetworkDesign] Sync failed:', err)
      }
    }
    sync()
  }, [devices, connections, call])

  // ── Backend Analysis ──
  const { data: health = { totalNodes: 0, totalEdges: 0, brokenLinks: 0, missingLabels: 0, orphanNodes: [], duplicateIPs: [], subnetErrors: [], suggestions: [] } } = useQuery<TopologyHealth>({
    queryKey: ['topology-health', devices.length, connections.length],
    queryFn: async () => {
      const res = await call('NetDesign.AnalyzeTopology')
      return res as TopologyHealth
    },
  })

  const { data: inventorySummary } = useQuery<InventorySummary>({
    queryKey: ['topology-inventory', devices.length],
    queryFn: async () => {
      return await call('NetDesign.GetInventory') as any
    }
  })

  const healthScore = useMemo(() => computeHealthScore(health), [health])
  const inventory = useMemo(() => groupByType(devices), [devices])
  const docGaps = useMemo(() => findDocumentationGaps(devices, connections), [devices, connections])

  const scoreColor =
    healthScore >= 80
      ? 'text-[var(--color-success)]'
      : healthScore >= 50
        ? 'text-[var(--color-warning)]'
        : 'text-[var(--color-danger)]'

  const scoreBg =
    healthScore >= 80
      ? 'bg-[var(--color-success)]'
      : healthScore >= 50
        ? 'bg-[var(--color-warning)]'
        : 'bg-[var(--color-danger)]'

  const handleDiscover = async () => {
    setDiscovering(true)
    try {
      const discovered = await discoverLocalNetwork()
      if (discovered.length > 0) {
        onDevicesDiscovered(discovered)
      }
    } finally {
      setDiscovering(false)
    }
  }

  return (
    <>
      {/* Toggle button (visible when closed) */}
      {!isOpen && (
        <button
          onClick={onToggle}
          className="fixed right-0 top-1/2 -translate-y-1/2 z-40 bg-[var(--color-panel)] border border-[var(--color-border)] border-r-0 rounded-l-lg px-1.5 py-6 text-[var(--color-text-faint)] hover:text-[var(--color-accent)] hover:bg-[var(--color-sidebar-hover)] transition-colors"
          title="Show Analysis Panel"
        >
          <ChevronLeft size={16} />
        </button>
      )}

      {/* Sidebar panel */}
      <div
        className={cn(
          'shrink-0 bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg overflow-hidden transition-all duration-300 flex flex-col',
          isOpen ? 'w-80 opacity-100' : 'w-0 opacity-0 pointer-events-none',
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-3 py-2.5 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]">
          <div className="flex items-center gap-2">
            <Radar size={16} className="text-[var(--color-accent)]" />
            <span className="text-sm font-semibold text-[var(--color-text)]">Analysis</span>
          </div>
          <button
            onClick={onToggle}
            className="text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors p-1 rounded hover:bg-[var(--color-sidebar-hover)]"
            title="Close panel"
          >
            <X size={14} />
          </button>
        </div>

        {/* Scrollable content */}
        <div className="flex-1 overflow-y-auto">
          {/* ── Section 1: Topology Summary ── */}
          <Section
            title="Summary"
            icon={<Activity size={14} />}
            defaultOpen={true}
          >
            <div className="space-y-2">
              <StatRow label="Total Devices" value={health.totalNodes} />
              <StatRow label="Total Connections" value={health.totalEdges} />
              {inventorySummary && (
                <>
                  <StatRow label="Missing IPs" value={inventorySummary.missingIPs} color={inventorySummary.missingIPs > 0 ? 'text-warning' : ''} />
                  <StatRow label="Missing MACs" value={inventorySummary.missingMACs} color={inventorySummary.missingMACs > 0 ? 'text-warning' : ''} />
                </>
              )}
              <div className="flex items-center justify-between">
                <span className="text-xs text-[var(--color-text-dim)]">Health Score</span>
                <div className="flex items-center gap-2">
                  <div className="w-16 h-1.5 rounded-full bg-[var(--color-panel-3)] overflow-hidden">
                    <div
                      className={cn('h-full rounded-full transition-all', scoreBg)}
                      style={{ width: `${healthScore}%` }}
                    />
                  </div>
                  <span className={cn('text-xs font-mono font-bold', scoreColor)}>
                    {healthScore}
                  </span>
                </div>
              </div>
            </div>
          </Section>

          {/* ── Section 2: Inventory Statistics (NEW) ── */}
          {inventorySummary && (
            <Section
              title="Inventory Stats"
              icon={<ListChecks size={14} />}
              defaultOpen={false}
            >
              <div className="space-y-3">
                {inventorySummary.vendors && Object.entries(inventorySummary.vendors).length > 0 && (
                  <div>
                    <span className="text-[10px] font-bold text-text-faint uppercase tracking-widest block mb-1.5">Top Vendors</span>
                    <div className="space-y-1">
                      {Object.entries(inventorySummary.vendors).map(([v, count]: [string, any]) => (
                        <StatRow key={v} label={v} value={count} />
                      ))}
                    </div>
                  </div>
                )}
                {inventorySummary.types && (
                  <div>
                    <span className="text-[10px] font-bold text-text-faint uppercase tracking-widest block mb-1.5">Device Types</span>
                    <div className="space-y-1">
                      {Object.entries(inventorySummary.types).map(([t, count]: [string, any]) => (
                        <StatRow key={t} label={t.charAt(0).toUpperCase() + t.slice(1)} value={count} />
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </Section>
          )}

          {/* ── Section 3: Topology Health ── */}
          <Section
            title="Health"
            icon={<Activity size={14} />}
            count={health.brokenLinks + health.missingLabels + health.orphanNodes.length + health.duplicateIPs.length + health.subnetErrors.length}
            countColor={
              health.brokenLinks + health.duplicateIPs.length > 0
                ? 'text-[var(--color-danger)] bg-[var(--color-danger)]/10'
                : health.missingLabels + health.orphanNodes.length > 0
                  ? 'text-[var(--color-warning)] bg-[var(--color-warning)]/10'
                  : 'text-[var(--color-success)] bg-[var(--color-success)]/10'
            }
            defaultOpen={true}
          >
            {/* Broken links */}
            <div className="flex items-center justify-between">
              <span className="text-xs text-[var(--color-text-dim)]">Broken Links</span>
              <span
                className={cn(
                  'text-xs font-mono font-semibold',
                  health.brokenLinks > 0 ? 'text-[var(--color-danger)]' : 'text-[var(--color-text-faint)]',
                )}
              >
                {health.brokenLinks}
              </span>
            </div>

            {/* Missing labels */}
            <div className="flex items-center justify-between">
              <span className="text-xs text-[var(--color-text-dim)]">Missing Labels</span>
              <span
                className={cn(
                  'text-xs font-mono font-semibold',
                  health.missingLabels > 0 ? 'text-[var(--color-warning)]' : 'text-[var(--color-text-faint)]',
                )}
              >
                {health.missingLabels}
              </span>
            </div>

            {/* Orphan nodes */}
            {health.orphanNodes.length > 0 && (
              <div>
                <div className="flex items-center gap-1.5 mb-1">
                  <AlertTriangle size={12} className="text-[var(--color-warning)]" />
                  <span className="text-xs text-[var(--color-warning)] font-medium">Orphan Nodes</span>
                </div>
                {health.orphanNodes.map((name) => (
                  <div
                    key={name}
                    className="text-xs text-[var(--color-text-dim)] pl-5 py-0.5"
                  >
                    {name}
                  </div>
                ))}
              </div>
            )}

            {/* Duplicate IPs */}
            {health.duplicateIPs.length > 0 && (
              <div>
                <div className="flex items-center gap-1.5 mb-1">
                  <AlertTriangle size={12} className="text-[var(--color-danger)]" />
                  <span className="text-xs text-[var(--color-danger)] font-medium">Duplicate IPs</span>
                </div>
                {health.duplicateIPs.map((dup) => (
                  <div
                    key={dup.ip}
                    className="text-xs text-[var(--color-text-dim)] pl-5 py-0.5"
                  >
                    {dup.ip} — {dup.nodes.join(', ')}
                  </div>
                ))}
              </div>
            )}

            {/* Subnet errors */}
            {health.subnetErrors.length > 0 && (
              <div>
                <div className="flex items-center gap-1.5 mb-1">
                  <AlertTriangle size={12} className="text-[var(--color-warning)]" />
                  <span className="text-xs text-[var(--color-warning)] font-medium">Subnet Mismatch</span>
                </div>
                {health.subnetErrors.map((err) => (
                  <div
                    key={err}
                    className="text-xs text-[var(--color-text-dim)] pl-5 py-0.5"
                  >
                    {err}
                  </div>
                ))}
              </div>
            )}

            {/* All clear */}
            {health.brokenLinks === 0 &&
              health.missingLabels === 0 &&
              health.orphanNodes.length === 0 &&
              health.duplicateIPs.length === 0 &&
              health.subnetErrors.length === 0 && (
                <div className="flex items-center gap-2 text-[var(--color-success)]">
                  <CheckCircle size={14} />
                  <span className="text-xs font-medium">All checks passed</span>
                </div>
              )}
          </Section>

          {/* ── Section 3: Device Inventory ── */}
          <Section
            title="Inventory"
            icon={<Cpu size={14} />}
            count={devices.length}
            defaultOpen={devices.length > 0}
          >
            {inventory.map((group) => (
              <div key={group.type} className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="text-[var(--color-text-faint)]">{getDeviceIcon(group.type)}</span>
                  <span className="text-xs font-medium text-[var(--color-text)] capitalize">
                    {group.type}s
                  </span>
                  <span className="text-xs text-[var(--color-text-faint)] font-mono">
                    ({group.count})
                  </span>
                </div>
                {group.devices.map((dev) => (
                  <div
                    key={dev.id}
                    className="flex items-center gap-2 pl-6 py-0.5"
                  >
                    <div
                      className={cn(
                        'w-1.5 h-1.5 rounded-full shrink-0',
                        dev.status === 'healthy'
                          ? 'bg-[var(--color-success)]'
                          : dev.status === 'warning'
                            ? 'bg-[var(--color-warning)]'
                            : 'bg-[var(--color-danger)]',
                      )}
                    />
                    <span className="text-xs text-[var(--color-text)] truncate flex-1">
                      {dev.label}
                    </span>
                    {dev.ip && (
                      <span className="text-[10px] text-[var(--color-text-faint)] font-mono shrink-0">
                        {dev.ip}
                      </span>
                    )}
                  </div>
                ))}
              </div>
            ))}

            {devices.length === 0 && (
              <p className="text-xs text-[var(--color-text-faint)] italic">
                No devices on canvas
              </p>
            )}
          </Section>

          {/* ── Section 4: Documentation Gaps ── */}
          <Section
            title="Docs Gaps"
            icon={<ListChecks size={14} />}
            count={docGaps.reduce((sum, g) => sum + g.items.length, 0)}
            countColor={
              docGaps.length > 0
                ? 'text-[var(--color-warning)] bg-[var(--color-warning)]/10'
                : 'text-[var(--color-success)] bg-[var(--color-success)]/10'
            }
            defaultOpen={false}
          >
            {docGaps.map((gap) => (
              <div key={gap.category}>
                <span className="text-xs text-[var(--color-warning)] font-medium">
                  {gap.category}
                </span>
                <ul className="mt-0.5 space-y-0.5">
                  {gap.items.map((item) => (
                    <li
                      key={item}
                      className="text-xs text-[var(--color-text-dim)] pl-4 relative before:content-[''] before:absolute before:left-1.5 before:top-[7px] before:w-1 before:h-1 before:rounded-full before:bg-[var(--color-text-faint)]"
                    >
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
            ))}

            {docGaps.length === 0 && (
              <div className="flex items-center gap-2 text-[var(--color-success)]">
                <CheckCircle size={14} />
                <span className="text-xs font-medium">Fully documented</span>
              </div>
            )}
          </Section>

          {/* ── Section 5: Documentation Details (NEW) ── */}
          <Section
            title="Documentation"
            icon={<FileText size={14} />}
            count={devices.filter(d => !d.ip || !d.mac || !d.notes).length}
            defaultOpen={false}
          >
            <div className="space-y-4">
              {devices.map(d => {
                const missing = []
                if (!d.ip) missing.push('IP')
                if (!d.mac) missing.push('MAC')
                if (!d.notes) missing.push('Notes')
                if (!d.subnet) missing.push('Subnet')

                if (missing.length === 0) return null

                return (
                  <div key={d.id} className="space-y-1">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-text truncate">{d.label}</span>
                      <span className="text-[9px] font-black uppercase text-warning bg-warning/10 px-1.5 py-0.5 rounded">Incomplete</span>
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {missing.map(m => (
                        <span key={m} className="text-[9px] font-mono text-text-faint bg-panel-3 px-1 rounded border border-border">Missing {m}</span>
                      ))}
                    </div>
                  </div>
                )
              })}
              {devices.every(d => d.ip && d.mac && d.notes && d.subnet) && (
                <div className="flex items-center gap-2 text-success py-2">
                  <CheckCircle2 size={14} />
                  <span className="text-xs font-bold uppercase tracking-widest">Full Coverage</span>
                </div>
              )}
            </div>
          </Section>

          {/* ── Section 6: AI Suggestions ── */}
          <Section
            title="Suggestions"
            icon={<Lightbulb size={14} />}
            count={health.suggestions.length}
            countColor="text-[var(--color-accent)] bg-[var(--color-accent-soft)]"
            defaultOpen={health.suggestions.length > 0}
          >
            <div className="space-y-1.5">
              {health.suggestions.map((suggestion, i) => (
                <div
                  key={i}
                  className="flex items-start gap-2 p-2 rounded bg-[var(--color-panel-2)] border border-[var(--color-border)]"
                >
                  <Lightbulb
                    size={12}
                    className="text-[var(--color-accent)] shrink-0 mt-0.5"
                  />
                  <span className="text-xs text-[var(--color-text-dim)] leading-relaxed">
                    {suggestion}
                  </span>
                </div>
              ))}
            </div>
          </Section>
        </div>

        {/* ── Section 6: Actions (sticky footer) ── */}
        <div className="border-t border-[var(--color-border)] p-3 space-y-2 bg-[var(--color-panel-2)]">
          <div className="flex gap-2">
            <button
              onClick={handleDiscover}
              disabled={discovering}
              className={cn(
                'flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-md text-xs font-medium transition-colors',
                discovering
                  ? 'bg-[var(--color-panel-3)] text-[var(--color-text-faint)] cursor-not-allowed'
                  : 'bg-[var(--color-accent)] text-[var(--color-bg)] hover:opacity-90',
              )}
            >
              <Radar size={14} className={cn(discovering && 'animate-spin')} />
              {discovering ? 'Scanning...' : 'Discover Network'}
            </button>
          </div>
          <div className="flex gap-2">
            <div className="flex-1 flex items-center gap-1 px-3 py-1.5 rounded-md text-xs text-[var(--color-text-faint)] bg-[var(--color-panel-3)] font-mono">
              <Wifi size={12} />
              {devices.length}d · {connections.length}c
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
