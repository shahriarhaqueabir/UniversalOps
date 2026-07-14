import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  Wifi,
  Globe,
  Network,
  ArrowUpCircle,
  ArrowDownCircle,
  PlusCircle,
  MinusCircle,
  CircleDot,
  CircleOff,
} from 'lucide-react'
import type { InterfaceInfo, NetworkChange } from '@/types'

export function InterfacesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: interfaces = [] } = useQuery<InterfaceInfo[]>({
    queryKey: ['netops-interfaces'],
    queryFn: async () => {
      const res = await call('NetOps.GetInterfaces') as InterfaceInfo[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  const { data: recentChanges = [] } = useQuery<NetworkChange[]>({
    queryKey: ['netops-recent-changes'],
    queryFn: async () => {
      const res = await call('NetOps.GetRecentChanges') as NetworkChange[]
      return res || []
    },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      {interfaces.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="w-16 h-16 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center text-[var(--color-text-faint)] mb-4">
            <Wifi size={28} />
          </div>
          <h3 className="text-base font-bold text-[var(--color-text)] mb-1">No Interfaces Found</h3>
          <p className="text-sm text-[var(--color-text-dim)]">Network interfaces will appear once detected by the system.</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-8">
          {interfaces.map(iface => (
            <div key={iface.name} className="bg-panel border border-border rounded-[var(--radius-lg)] p-10 shadow-xl relative overflow-hidden group">
              <div className="absolute top-0 right-0 w-32 h-32 bg-accent/5 rounded-bl-full pointer-events-none" />
              <div className="flex items-center gap-6 mb-8">
                <div className={cn("w-16 h-16 rounded-2xl flex items-center justify-center border shadow-inner transition-all", iface.is_up ? "bg-success/10 border-success/30 text-success" : "bg-danger/10 border-danger/30 text-danger")}>
                  <Wifi size={32} />
                </div>
                <div>
                  <h3 className="text-2xl font-bold text-text uppercase tracking-tighter">{iface.name}</h3>
                  <div className="flex items-center gap-2 mt-1">
                    <span className={cn("w-2 h-2 rounded-full", iface.is_up ? "bg-success shadow-[0_0_8px_var(--color-success)]" : "bg-danger")} />
                    <span className="text-sm font-bold text-text-faint uppercase tracking-widest">{iface.is_up ? 'ACTIVE NODE' : 'DISCONNECTED'}</span>
                  </div>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <p className="text-xs font-bold text-text-faint uppercase mb-2">Physical MAC</p>
                  <p className="text-sm font-medium text-[var(--color-text)] tabular-nums">{iface.mac}</p>
                </div>
                <div>
                  <p className="text-xs font-bold text-text-faint uppercase mb-2">Link Capacity</p>
                  <p className="text-xl font-bold text-accent">{iface.speed}</p>
                </div>
              </div>
              <div className="mt-8 pt-8 border-t border-border flex items-center gap-4 flex-wrap">
                {iface.ips.map((ip, idx) => (
                  <div key={idx} className="px-4 py-1.5 bg-panel-3 border border-border rounded-full text-sm font-bold text-accent tabular-nums flex items-center gap-2">
                    <Globe size={14} /> {ip}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Recent Network Changes */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
        <div className="flex items-center gap-4 mb-6">
          <Network size={20} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Recent State Changes</h3>
          {recentChanges.length > 0 && (
            <span className="ml-auto px-2.5 py-0.5 text-xs font-bold rounded-full bg-accent/15 text-accent border border-accent/30">
              {recentChanges.length}
            </span>
          )}
        </div>
        {recentChanges.length === 0 ? (
          <p className="text-sm font-medium text-text-faint">No interface changes detected yet. Changes appear after the first two polling cycles.</p>
        ) : (
          <div className="space-y-3">
            {recentChanges.map((change, idx) => {
              const iconMap: Record<string, React.ReactNode> = {
                up: <ArrowUpCircle size={16} className="text-success" />,
                down: <ArrowDownCircle size={16} className="text-danger" />,
                ip_added: <PlusCircle size={16} className="text-accent" />,
                ip_removed: <MinusCircle size={16} className="text-warning" />,
                appeared: <CircleDot size={16} className="text-success" />,
                disappeared: <CircleOff size={16} className="text-danger" />,
              }
              const labelMap: Record<string, string> = {
                up: 'UP',
                down: 'DOWN',
                ip_added: 'IP ADDED',
                ip_removed: 'IP REMOVED',
                appeared: 'APPEARED',
                disappeared: 'GONE',
              }
              const colorMap: Record<string, string> = {
                up: 'bg-success/15 text-success border-success/30',
                down: 'bg-danger/15 text-danger border-danger/30',
                ip_added: 'bg-accent/15 text-accent border-accent/30',
                ip_removed: 'bg-warning/15 text-warning border-warning/30',
                appeared: 'bg-success/15 text-success border-success/30',
                disappeared: 'bg-danger/15 text-danger border-danger/30',
              }
              return (
                <div key={idx} className="flex items-center gap-4 p-4 bg-panel-3 border border-border rounded-xl">
                  {iconMap[change.type]}
                  <span className={cn('px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-widest rounded-full border', colorMap[change.type] || 'bg-text-faint/20 text-text-faint border-border')}>
                    {labelMap[change.type] || change.type}
                  </span>
                  <span className="text-sm font-bold text-text uppercase tracking-tight">{change.interface}</span>
                  <span className="text-sm font-medium text-text-dim flex-1">{change.detail}</span>
                  <span className="text-[11px] font-medium text-text-faint tabular-nums whitespace-nowrap">
                    {new Date(change.timestamp).toLocaleTimeString()}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
