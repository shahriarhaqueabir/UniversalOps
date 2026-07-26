import { useState, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { History, Clock, AlertTriangle } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { EmptyState } from '@/components/ui/EmptyState'
import { Panel } from '@/components/ui/Panel'
import type { SecurityEvent, ScheduledTask, PrivilegeEvent } from '@/types'
import { cn } from '@/lib/utils'
import { useVirtualizer } from '@tanstack/react-virtual'

export function EventsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [activeFilter, setActiveFilter] = useState<string>('all')

  const { data: events = [] } = useQuery<SecurityEvent[]>({
    queryKey: ['secops-events'],
    queryFn: async () => (await call('SecOps.GetSecurityEvents') as SecurityEvent[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: tasks = [] } = useQuery<ScheduledTask[]>({
    queryKey: ['secops-tasks'],
    queryFn: async () => (await call('SecOps.GetScheduledTasks') as ScheduledTask[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: privEvents = [] } = useQuery<PrivilegeEvent[]>({
    queryKey: ['secops-privilege-events'],
    queryFn: async () => (await call('SecOps.GetPrivilegeEvents') as PrivilegeEvent[]) || [],
    refetchInterval: refreshInterval,
  })

  // Timeline query reserved for future event timeline visualization
  // useQuery<SecTimelineEvent[]>({
  //   queryKey: ['secops-timeline'],
  //   queryFn: async () => (await call('SecOps.GetEventTimeline') as SecTimelineEvent[]) || [],
  //   refetchInterval: refreshInterval,
  // })

  const categories = [
    { id: 'all', label: 'All', ids: null },
    { id: 'failed-logins', label: 'Failed Logins', ids: [4625] },
    { id: 'elevation', label: 'Elevation', ids: [4672, 4673] },
    { id: 'policy', label: 'Policy Changes', ids: [1102, 4719] },
    { id: 'account', label: 'Account Changes', ids: [4720, 4722, 4725, 4726] },
  ]

  const filtered = activeFilter === 'all'
    ? events
    : events.filter(e => {
      const cat = categories.find(c => c.id === activeFilter)
      return cat?.ids?.includes(e.id)
    })

  const parentRef = useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: filtered.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60,
    overscan: 5,
  })

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Log & Event Analysis"
        objective="Analyze security event logs, scheduled tasks, and privilege escalation events for anomalies."
        checklist={['Security event log review', 'Scheduled task audit', 'Privilege escalation detection', 'Event timeline analysis']}
      />

      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="Total Events" value={events.length} icon={<History size={24} />} variant="default" />
        <MiniStat label="Failed Logins" value={events.filter(e => e.id === 4625).length} icon={<AlertTriangle size={24} />} variant={events.filter(e => e.id === 4625).length === 0 ? 'success' : 'danger'} />
        <MiniStat label="Scheduled Tasks" value={tasks.length} icon={<Clock size={24} />} variant="default" />
        <MiniStat label="Privilege Events" value={privEvents.length} icon={<AlertTriangle size={24} />} variant={privEvents.length === 0 ? 'success' : 'warning'} />
      </div>

      {/* Filter Buttons */}
      <div className="flex items-center gap-3 flex-wrap">
        {categories.map(c => (
          <button
            key={c.id}
            onClick={() => setActiveFilter(c.id)}
            className={cn(
              'px-4 py-2 rounded-full text-sm font-bold uppercase tracking-wider border transition-all',
              activeFilter === c.id
                ? 'bg-accent text-white border-accent shadow-lg'
                : 'bg-panel border-border text-text-dim hover:text-text hover:bg-[var(--color-sidebar-hover)]'
            )}
          >
            {c.label}
            {c.ids && (
              <span className="ml-2 text-xs tabular-nums">
                {events.filter(e => c.ids.includes(e.id)).length}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Events Table */}
      <Panel padding="none" category="security" className="overflow-hidden">
        <div ref={parentRef} className="max-h-[600px] overflow-y-auto">
          {filtered.length === 0 ? (
            <div className="py-12">
              <EmptyState
                icon={<History size={28} />}
                title="No Security Events"
                description={activeFilter === 'all' ? 'No security events recorded.' : 'No events match the selected filter.'}
              />
            </div>
          ) : (
            <table className="w-full text-left border-collapse table-fixed">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                <tr>
                  <th className="w-[100px] px-6 py-4 text-xs font-bold text-text-faint uppercase tracking-wider">ID</th>
                  <th className="w-[120px] px-6 py-4 text-xs font-bold text-text-faint uppercase tracking-wider">Level</th>
                  <th className="w-[200px] px-6 py-4 text-xs font-bold text-text-faint uppercase tracking-wider">Origin</th>
                  <th className="px-6 py-4 text-xs font-bold text-text-faint uppercase tracking-wider">Message</th>
                </tr>
              </thead>
              <tbody
                style={{
                  height: `${rowVirtualizer.getTotalSize()}px`,
                  width: '100%',
                  position: 'relative',
                }}
              >
                {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                  const e = filtered[virtualRow.index]
                  return (
                    <tr
                      key={virtualRow.key}
                      data-index={virtualRow.index}
                      ref={rowVirtualizer.measureElement}
                      className={cn(
                        'border-b border-border/20 hover:bg-[var(--color-sidebar-hover)]',
                        e.level === 'Error' ? 'bg-danger/5' : ''
                      )}
                      style={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        height: `${virtualRow.size}px`,
                        transform: `translateY(${virtualRow.start}px)`,
                      }}
                    >
                      <td className="px-6 py-4 text-sm font-bold text-text-faint tabular-nums">{e.id}</td>
                      <td className="px-6 py-4"><StatusBadge status={e.level} /></td>
                      <td className="px-6 py-4 text-sm text-text truncate">{e.provider}</td>
                      <td className="px-6 py-4 text-sm text-text-dim truncate">{e.message}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>
      </Panel>

      {/* Scheduled Tasks */}
      <Panel padding="lg" category="security">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
          <Clock size={22} className="text-warning" /> Scheduled Tasks
        </h3>
        {tasks.length === 0 ? (
          <p className="text-text-dim text-sm">No scheduled tasks detected.</p>
        ) : (
          <div className="max-h-[300px] overflow-y-auto">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Name</th>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Status</th>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Last Run</th>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Next Run</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map((t, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)]">
                    <td className="px-4 py-3 text-sm font-bold text-text">{t.name}</td>
                    <td className="px-4 py-3"><StatusBadge status={t.status} /></td>
                    <td className="px-4 py-3 text-sm text-text-dim tabular-nums">{t.last_run || 'Never'}</td>
                    <td className="px-4 py-3 text-sm text-accent font-bold tabular-nums">{t.next_run || 'N/A'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Panel>
    </div>
  )
}
