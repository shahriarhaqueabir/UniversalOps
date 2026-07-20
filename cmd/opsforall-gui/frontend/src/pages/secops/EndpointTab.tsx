import { useQuery } from '@tanstack/react-query'
import { HardDrive, ShieldCheck, Server, Lock } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { Panel } from '@/components/ui/Panel'
import type { DiskEncryption, SecureBoot, SystemService } from '@/types'

export function EndpointTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: disks = [] } = useQuery<DiskEncryption[]>({
    queryKey: ['secops-disk-encryption'],
    queryFn: async () => (await call('SecOps.GetDiskEncryptionStatus') as DiskEncryption[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: secureBoot } = useQuery<SecureBoot>({
    queryKey: ['secops-secure-boot'],
    queryFn: async () => await call('SecOps.GetSecureBootStatus') as SecureBoot,
    refetchInterval: refreshInterval,
  })

  const { data: services = [] } = useQuery<SystemService[]>({
    queryKey: ['secops-services'],
    queryFn: async () => (await call('SecOps.GetRunningServices') as SystemService[]) || [],
    refetchInterval: refreshInterval,
  })

  const encryptedCount = disks.filter(d => d.encrypted).length
  const runningServices = services.filter(s => s.status.toLowerCase() === 'running').length

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Endpoint Security"
        objective="Verify disk encryption, secure boot, and running services to ensure endpoint hardening."
        checklist={['Full disk encryption status', 'Secure Boot verification', 'Running services inventory', 'Service startup types']}
        why="Endpoints are the front line of physical and network security. Ensuring they are 'locked' prevents local data exfiltration and persistent boot-level threats."
        risks={['Insecure boot can allow rootkits', 'Unencrypted drives are vulnerable to theft', 'Rogue services can bypass firewalls']}
        typicalValues="Secure Boot: ON, Disk: Encrypted (BitLocker/FileVault), Services: No unknown auto-start items"
        recommendedActions={['Enable BitLocker/FileVault', 'Review auto-start services', 'Enable TPM 2.0 if available']}
      />

      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="Disk Volumes" value={disks.length} icon={<HardDrive size={24} />} variant="default" />
        <MiniStat label="Encrypted" value={encryptedCount} icon={<Lock size={24} />} variant={encryptedCount === disks.length && disks.length > 0 ? 'success' : 'warning'} />
        <MiniStat label="Secure Boot" value={secureBoot?.enabled ? 'On' : 'Off'} icon={<ShieldCheck size={24} />} variant={secureBoot?.enabled ? 'success' : 'warning'} />
        <MiniStat label="Running Services" value={runningServices} icon={<Server size={24} />} variant="default" />
      </div>

      {/* Disk Encryption */}
      <Panel padding="lg" category="security">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
          <HardDrive size={22} className="text-accent" /> Disk Encryption
        </h3>
        <div className="space-y-3">
          {disks.map((d, i) => (
            <div key={i} className="flex items-center justify-between bg-panel-2 border border-border rounded-xl px-5 py-3">
              <span className="text-sm font-bold text-text">{d.volume}</span>
              <StatusBadge status={d.encrypted ? 'enabled' : 'disabled'} />
              <span className="text-sm text-text-dim">{d.method}</span>
              <span className="text-sm text-text-faint">{d.status}</span>
            </div>
          ))}
        </div>
      </Panel>

      {/* Secure Boot */}
      {secureBoot && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
            <ShieldCheck size={22} className="text-success" /> Secure Boot
          </h3>
          <div className="flex items-center gap-4">
            <StatusBadge status={secureBoot.enabled ? 'enabled' : 'disabled'} />
            <span className="text-sm text-text-dim">{secureBoot.state}</span>
          </div>
        </Panel>
      )}

      {/* Running Services */}
      <Panel padding="lg" category="security">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 flex items-center gap-3">
          <Server size={22} className="text-accent" /> Running Services
        </h3>
        <div className="max-h-[400px] overflow-y-auto">
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Display Name</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase tracking-wider">Startup</th>
              </tr>
            </thead>
            <tbody>
              {services.slice(0, 50).map((s, i) => (
                <tr key={i} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)]">
                  <td className="px-4 py-3 text-sm font-bold text-accent">{s.name}</td>
                  <td className="px-4 py-3 text-sm text-text">{s.display_name}</td>
                  <td className="px-4 py-3"><StatusBadge status={s.status} /></td>
                  <td className="px-4 py-3 text-sm text-text-dim">{s.startup_type}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  )
}
