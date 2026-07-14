import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Cable,
  Server,
  Timer,
} from 'lucide-react'
import type { ConnectionInfo } from '@/types'
import { SectionBriefing, MiniStat, StatusBadge } from './components'

export function ConnectionsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: connections = [] } = useQuery<ConnectionInfo[]>({
    queryKey: ['netops-connections'],
    queryFn: async () => {
      const res = await call('NetOps.GetConnections')
      return (res as ConnectionInfo[]) || []
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="grid grid-cols-4 gap-12 animate-in fade-in duration-500">
      <div className="col-span-1 space-y-8">
        <SectionBriefing
          title="Endpoint Matrix"
          objective="Audit established sockets to detect unauthorized data exfiltration or active C2 channels."
          checklist={[
            "ESTABLISHED: Active traffic session.",
            "LISTENING: Local port waiting for input.",
            "TIME_WAIT: Socket closing sequence.",
            "CLOSE_WAIT: Pending remote termination."
          ]}
        />
      </div>
      <div className="col-span-3 space-y-4">
        <div className="grid grid-cols-3 gap-4">
          <MiniStat label="Established" value={connections.filter(c => c.state === 'ESTABLISHED').length} icon={<Cable size={24} />} />
          <MiniStat label="Listening" value={connections.filter(c => c.state === 'LISTEN').length} icon={<Server size={24} />} />
          <MiniStat label="Time Wait" value={connections.filter(c => c.state === 'TIME_WAIT').length} icon={<Timer size={24} />} />
        </div>
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
          <div className="max-h-[800px] overflow-y-auto">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                <tr>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Protocol</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Endpoint Node</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Process Origin</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">State</th>
                </tr>
              </thead>
              <tbody>
                {connections.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-8 py-16 text-center">
                      <p className="text-sm font-semibold text-[var(--color-text-dim)]">No active connections detected.</p>
                      <p className="text-xs text-[var(--color-text-faint)] mt-1">Connections will appear as network activity is observed.</p>
                    </td>
                  </tr>
                ) : connections.map((c, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all group">
                    <td className="px-8 py-4 font-bold text-accent">{c.protocol}</td>
                    <td className="px-8 py-4">
                      <div className="flex flex-col">
                        <span className="text-lg font-bold text-text">{c.remote_addr}:{c.remote_port}</span>
                        <span className="text-sm font-bold text-text-faint uppercase tabular-nums">LOCAL: {c.local_addr}:{c.local_port}</span>
                      </div>
                    </td>
                    <td className="px-8 py-4">
                      <div className="flex flex-col">
                        <span className="text-sm font-medium text-[var(--color-text)]">{c.process_name || 'System Core'}</span>
                        <span className="text-xs font-bold text-text-faint uppercase tracking-widest">PID: {c.pid}</span>
                      </div>
                    </td>
                    <td className="px-8 py-4 text-right"><StatusBadge status={c.state} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
