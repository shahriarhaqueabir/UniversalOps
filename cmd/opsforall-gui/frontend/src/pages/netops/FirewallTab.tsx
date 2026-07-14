import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  ShieldAlert,
  ShieldCheck,
  ArrowUpRight,
  ArrowDownRight,
  ToggleLeft,
  ToggleRight,
  ServerCrash,
} from 'lucide-react'
import { SectionBriefing } from './components'
import type { NetOpsFirewallRuleData } from '@/types'

// ── Main FirewallTab ──

export function FirewallTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: rules = [], isLoading } = useQuery<NetOpsFirewallRuleData[]>({
    queryKey: ['netops-firewall-rules'],
    queryFn: async () => {
      const res = (await call('NetOps.GetFirewallRules')) as NetOpsFirewallRuleData[]
      return res || []
    },
    refetchInterval: refreshInterval,
    retry: false,
  })

  const inboundCount = rules.filter(r => r.direction.toLowerCase() === 'inbound').length
  const outboundCount = rules.filter(r => r.direction.toLowerCase() === 'outbound').length
  const allowCount = rules.filter(r => r.action.toUpperCase() === 'ALLOW').length
  const denyCount = rules.filter(r => r.action.toUpperCase() === 'DENY').length
  const disabledCount = rules.filter(r => !r.enabled).length

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <SectionBriefing
        title="Firewall Rules"
        objective="Enumerate system firewall rules. Audit inbound/outbound policies, protocol coverage, and port exposure."
        checklist={[
          'Inbound rules control incoming traffic',
          'ALLOW rules permit traffic through the firewall',
          'DENY rules block traffic',
          'Disabled rules may create security gaps',
        ]}
      />

      {/* ── Rules Table ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
            <ShieldAlert size={18} className="text-accent" />
          </div>
          <h3 className="text-sm font-bold text-text uppercase tracking-widest">Firewall Rules</h3>
          <span className="ml-auto px-2.5 py-0.5 text-[10px] font-bold rounded-full bg-accent/15 text-accent border border-accent/30">
            {rules.length}
          </span>
        </div>

        {/* ── Summary Badges ── */}
        {rules.length > 0 && (
          <div className="flex items-center gap-3 mb-6">
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-panel-2 border border-border">
              <ArrowDownRight size={12} className="text-accent" />
              <span className="text-[10px] font-bold text-text-dim">{inboundCount} Inbound</span>
            </div>
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-panel-2 border border-border">
              <ArrowUpRight size={12} className="text-accent" />
              <span className="text-[10px] font-bold text-text-dim">{outboundCount} Outbound</span>
            </div>
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-success/10 border border-success/30">
              <ShieldCheck size={12} className="text-success" />
              <span className="text-[10px] font-bold text-success">{allowCount} Allow</span>
            </div>
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-danger/10 border border-danger/30">
              <ShieldAlert size={12} className="text-danger" />
              <span className="text-[10px] font-bold text-danger">{denyCount} Deny</span>
            </div>
            {disabledCount > 0 && (
              <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-warning/10 border border-warning/30">
                <ToggleLeft size={12} className="text-warning" />
                <span className="text-[10px] font-bold text-warning">{disabledCount} Disabled</span>
              </div>
            )}
          </div>
        )}

        {/* ── Table ── */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <span className="text-xs font-medium text-text-faint">Loading firewall rules...</span>
          </div>
        ) : rules.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-panel-3 flex items-center justify-center border border-border mb-4">
              <ServerCrash size={32} className="text-text-faint" />
            </div>
            <p className="text-sm font-bold text-text uppercase tracking-widest mb-1">
              No Rules Found
            </p>
            <p className="text-xs font-medium text-text-faint">
              No firewall rules were detected on this system
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-border">
                  {['Name', 'Direction', 'Action', 'Protocol', 'Ports', 'Enabled'].map(h => (
                    <th
                      key={h}
                      className="px-4 py-3 text-[10px] font-bold text-text-faint uppercase tracking-widest"
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rules.map((rule, idx) => {
                  const isAllow = rule.action.toUpperCase() === 'ALLOW'
                  const isInbound = rule.direction.toLowerCase() === 'inbound'

                  return (
                    <tr
                      key={idx}
                      className={cn(
                        'border-b border-border/50 transition-colors hover:bg-panel-2/50',
                        !rule.enabled && 'opacity-50',
                      )}
                    >
                      <td className="px-4 py-3">
                        <span className="text-xs font-bold text-text">{rule.name}</span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          {isInbound ? (
                            <ArrowDownRight size={12} className="text-accent" />
                          ) : (
                            <ArrowUpRight size={12} className="text-accent" />
                          )}
                          <span className="text-xs font-medium text-text-dim">{rule.direction}</span>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={cn(
                            'inline-block px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-widest rounded-full border',
                            isAllow
                              ? 'bg-success/15 text-success border-success/30'
                              : 'bg-danger/15 text-danger border-danger/30',
                          )}
                        >
                          {rule.action}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="text-xs font-medium text-text-dim">{rule.protocol}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="text-xs font-medium text-text-dim tabular-nums">{rule.ports}</span>
                      </td>
                      <td className="px-4 py-3">
                        {rule.enabled ? (
                          <ToggleRight size={16} className="text-success" />
                        ) : (
                          <ToggleLeft size={16} className="text-text-faint" />
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
