import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Shield, Radio, Activity, Lock, Unlock, AlertTriangle } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { ConfirmationModal } from '@/components/dialogs/ConfirmationModal'
import type { FirewallRule, ListeningPort, ActionPreview } from '@/types'
import { cn } from '@/lib/utils'

export function PerimeterTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const queryClient = useQueryClient()
  const [preview, setPreview] = useState<ActionPreview | null>(null)

  const { data: rules = [] } = useQuery<FirewallRule[]>({
    queryKey: ['secops-firewall-rules'],
    queryFn: async () => (await call('SecOps.GetFirewallRules') as FirewallRule[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: ports = [] } = useQuery<ListeningPort[]>({
    queryKey: ['secops-listening-ports'],
    queryFn: async () => (await call('SecOps.GetListeningPorts') as ListeningPort[]) || [],
    refetchInterval: refreshInterval,
  })

  const highRiskRules = rules.filter(r => r.is_high_risk && r.enabled).length
  const highRiskPorts = ports.filter(p => p.risk_level === 'high').length

  const requestToggle = async (rule: FirewallRule) => {
    const p = await call('SecOps.SetFirewallRuleHandshake', rule.name, !rule.enabled) as ActionPreview
    setPreview(p)
  }

  const handleConfirm = async () => {
    if (!preview) return
    const handshakeID = preview.handshake_id
    setPreview(null)
    await call('App.ConfirmAction', handshakeID)
    queryClient.invalidateQueries({ queryKey: ['secops-firewall-rules'] })
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Perimeter Security"
        objective="Analyze inbound/outbound access rules and identify processes listening on external network interfaces."
        checklist={['Firewall policy audit', 'Listening port exposure', 'High-risk rule detection', 'Process-to-port mapping']}
      />

      <ConfirmationModal
        preview={preview}
        onConfirm={handleConfirm}
        onCancel={() => setPreview(null)}
      />

      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="Total Rules" value={rules.length} icon={<Shield size={24} />} variant="default" />
        <MiniStat label="High Risk Rules" value={highRiskRules} icon={<AlertTriangle size={24} />} variant={highRiskRules === 0 ? 'success' : 'danger'} />
        <MiniStat label="Listening Ports" value={ports.length} icon={<Activity size={24} />} variant="default" />
        <MiniStat label="High Risk Ports" value={highRiskPorts} icon={<Radio size={24} />} variant={highRiskPorts === 0 ? 'success' : 'danger'} />
      </div>

      {/* Firewall Rules */}
      <div className="bg-panel border border-border rounded-[2rem] p-10 shadow-xl">
        <h3 className="text-xl font-black text-text uppercase tracking-widest mb-8 flex items-center gap-3">
          <div className="w-1.5 h-6 bg-accent rounded-full" />
          Active Firewall Policy
        </h3>
        <div className="space-y-3">
          {rules.map((rule, i) => (
            <div key={i} className={cn(
              "flex items-center justify-between p-5 rounded-2xl border transition-all duration-300 group",
              rule.is_high_risk && rule.enabled ? "bg-danger/5 border-danger/30" : "bg-panel-2 border-border/50 hover:border-accent/40"
            )}>
              <div className="flex items-center gap-5">
                <div className={cn(
                  "w-12 h-12 rounded-xl flex items-center justify-center border transition-transform duration-300 group-hover:rotate-3",
                  rule.enabled ? "bg-success/10 border-success/20 text-success" : "bg-panel-3 border-border text-text-faint"
                )}>
                  {rule.enabled ? <Lock size={20} /> : <Unlock size={20} />}
                </div>
                <div>
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-black text-text uppercase tracking-wider">{rule.name}</span>
                    {rule.is_high_risk && <span className="px-2 py-0.5 rounded-full bg-danger text-white text-[8px] font-black uppercase tracking-widest animate-pulse">High Risk</span>}
                  </div>
                  <p className="text-[10px] font-bold text-text-faint uppercase tracking-widest mt-1">
                    {rule.direction} \u2022 {rule.action} \u2022 {rule.protocol} \u2022 PORT {rule.local_port || 'ANY'}
                  </p>
                </div>
              </div>
              <button
                onClick={() => requestToggle(rule)}
                className={cn(
                  "px-6 py-2 rounded-xl text-[10px] font-black uppercase tracking-[0.2em] transition-all active:scale-95",
                  rule.enabled ? "bg-danger/10 text-danger border border-danger/20 hover:bg-danger/20" : "bg-success/10 text-success border border-success/20 hover:bg-success/20"
                )}
              >
                {rule.enabled ? 'Disable' : 'Enable'}
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Listening Ports */}
      <div className="bg-panel border border-border rounded-[2rem] p-10 shadow-xl">
        <h3 className="text-xl font-black text-text uppercase tracking-widest mb-8 flex items-center gap-3">
          <div className="w-1.5 h-6 bg-accent rounded-full" />
          Listening Ports
        </h3>
        <div className="max-h-[500px] overflow-y-auto pr-2 space-y-2">
          {ports.map((p, i) => {
            const riskStyle = p.risk_level === 'high'
              ? "bg-danger/5 border-danger/30"
              : p.risk_level === 'medium'
                ? "bg-warning/5 border-warning/30"
                : "bg-panel-2 border-border/50"
            const iconStyle = p.risk_level === 'high'
              ? "bg-danger/10 border-danger/20 text-danger"
              : p.risk_level === 'medium'
                ? "bg-warning/10 border-warning/20 text-warning"
                : "bg-success/10 border-success/20 text-success"
            return (
              <div key={i} className={cn(
                "flex items-center justify-between p-5 rounded-xl border transition-all duration-300",
                riskStyle
              )}>
                <div className="flex items-center gap-4">
                  <div className={cn(
                    "w-10 h-10 rounded-lg flex items-center justify-center border",
                    iconStyle
                  )}>
                    <Radio size={18} />
                  </div>
                  <div>
                    <span className="text-sm font-black text-text uppercase tabular-nums tracking-wider">PORT {p.port}</span>
                    {p.service_name && (
                      <span className="ml-2 text-[10px] font-bold text-accent uppercase tracking-widest">({p.service_name})</span>
                    )}
                    <p className="text-[10px] font-bold text-text-faint uppercase tracking-widest mt-0.5">{p.process_name} (PID {p.pid})</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-[10px] font-black text-text-faint uppercase tracking-widest">{p.protocol}</span>
                  {p.is_external && (
                    <span className="px-2 py-0.5 rounded-full text-[8px] font-black uppercase tracking-widest bg-warning/10 text-warning border border-warning/30">External</span>
                  )}
                  <StatusBadge status={p.risk_level === 'high' ? 'danger' : p.risk_level === 'medium' ? 'warning' : 'success'} />
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
