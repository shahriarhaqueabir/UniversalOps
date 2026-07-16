import { useQuery } from '@tanstack/react-query'
import { Shield, Globe, AlertTriangle, Lock } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'
import { StatusBadge } from '@/components/ui/StatusBadge'
import type { TLSCertificate, PublicExposure } from '@/types'

export function NetworkSecurityTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: certs = [] } = useQuery<TLSCertificate[]>({
    queryKey: ['secops-tls-certs'],
    queryFn: async () => (await call('SecOps.GetTLSCertificates') as TLSCertificate[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: exposed = [] } = useQuery<PublicExposure[]>({
    queryKey: ['secops-public-exposure'],
    queryFn: async () => (await call('SecOps.GetPublicExposure') as PublicExposure[]) || [],
    refetchInterval: refreshInterval,
  })

  const expiringCount = certs.filter(c => c.is_expiring).length
  const criticalPorts = exposed.filter(e => e.severity === 'critical').length

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Network Security"
        objective="Inspect TLS certificates and identify publicly exposed ports that could be attack vectors."
        checklist={['TLS certificate validity', 'Public port exposure', 'Critical service exposure (RDP, SSH, Telnet)', 'Certificate expiry monitoring']}
      />

      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="TLS Certificates" value={certs.length} icon={<Lock size={24} />} variant="default" />
        <MiniStat label="Expiring Soon" value={expiringCount} icon={<AlertTriangle size={24} />} variant={expiringCount === 0 ? 'success' : 'danger'} />
        <MiniStat label="Publicly Exposed" value={exposed.length} icon={<Globe size={24} />} variant={exposed.length === 0 ? 'success' : 'warning'} />
        <MiniStat label="Critical Ports" value={criticalPorts} icon={<Shield size={24} />} variant={criticalPorts === 0 ? 'success' : 'danger'} />
      </div>

      {/* Public Exposure */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
          <Globe size={22} className="text-warning" /> Public Port Exposure
        </h3>
        {exposed.length === 0 ? (
          <p className="text-text-dim text-sm">No externally-exposed ports detected.</p>
        ) : (
          <div className="space-y-3">
            {exposed.map((e, i) => (
              <div key={i} className="flex items-center justify-between bg-panel-2 border border-border rounded-xl px-4 py-3">
                <div className="flex items-center gap-4">
                  <span className="text-sm font-bold text-accent tabular-nums">{e.port}</span>
                  <StatusBadge status={e.severity} />
                </div>
                <span className="text-sm text-text-dim">{e.protocol}</span>
                <span className="text-sm font-bold text-text">{e.process_name}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* TLS Certificates */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
          <Lock size={22} className="text-accent" /> TLS Certificates
        </h3>
        {certs.length === 0 ? (
          <p className="text-text-dim text-sm">No TLS certificates found.</p>
        ) : (
          <div className="space-y-3">
            {certs.slice(0, 20).map((c, i) => (
              <div key={i} className="flex items-center justify-between bg-panel-2 border border-border rounded-xl px-4 py-3">
                <span className="text-sm font-bold text-text truncate max-w-xs">{c.subject}</span>
                <span className="text-sm text-text-dim">{c.issuer}</span>
                <StatusBadge status={c.is_expiring ? 'warning' : 'active'} />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
