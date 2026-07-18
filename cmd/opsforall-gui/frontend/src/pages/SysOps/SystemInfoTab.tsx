import { Cpu, Users, Copy, Globe } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { SystemInfo, CPUInfo, CPUExtendedInfo, LoggedInUserData } from '@/types'

interface SystemInfoTabProps {
  sysInfo: SystemInfo
  cpuInfo: CPUInfo
}

export function SystemInfoTab({ sysInfo, cpuInfo }: SystemInfoTabProps) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: cpuExtended } = useQuery<CPUExtendedInfo>({
    queryKey: ['sysops-cpu-extended'],
    queryFn: async () => { const r = await call('SysOps.GetCPUExtended'); return r as CPUExtendedInfo },
    refetchInterval: refreshInterval,
  })

  const { data: users = [] } = useQuery<LoggedInUserData[]>({
    queryKey: ['sysops-users'],
    queryFn: async () => { const r = await call('SysOps.GetLoggedInUsers'); return (r as LoggedInUserData[]) || [] },
    refetchInterval: refreshInterval,
  })

  const copyHostname = () => navigator.clipboard.writeText(sysInfo.hostname)

  return (
    <div className="space-y-8">
      {/* OS Information */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Globe size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">OS Information</h3>
        </div>
        <div className="grid grid-cols-2 gap-x-12 gap-y-6">
          <InfoRow label="Hostname" value={sysInfo.hostname} action={<button onClick={copyHostname} className="p-1.5 hover:bg-panel-3 rounded-lg active:scale-90 transition-transform"><Copy size={14} className="text-text-faint hover:text-accent transition-colors" /></button>} />
          <InfoRow label="Platform" value={sysInfo.platform} />
          <InfoRow label="Kernel Version" value={sysInfo.kernel_version} />
          <InfoRow label="Architecture" value={sysInfo.kernel_arch} />
          <InfoRow label="Uptime" value={sysInfo.uptime} />
          <InfoRow label="Virtualization" value={sysInfo.virtualization || 'None'} />
        </div>
      </div>

      {/* Hardware Summary */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Cpu size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Hardware Summary</h3>
        </div>
        <div className="grid grid-cols-2 gap-x-12 gap-y-6">
          <InfoRow label="CPU Model" value={cpuExtended?.model_name || 'N/A'} />
          <InfoRow label="CPU Frequency" value={cpuExtended ? `${cpuExtended.frequency_mhz.toFixed(0)} MHz` : 'N/A'} isNumeric />
          <InfoRow label="Physical Cores" value={cpuInfo.physical_cores} isNumeric />
          <InfoRow label="Logical Cores" value={cpuInfo.logical_cores} isNumeric />
          <InfoRow label="Cache Size" value={cpuExtended ? `${cpuExtended.cache_size_kb} KB` : 'N/A'} isNumeric />
          <InfoRow label="Processes" value={sysInfo.process_count} isNumeric />
        </div>
      </div>

      {/* Logged-in Users */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Users size={20} className="text-[var(--color-success)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Logged-in Users</h3>
        </div>
        {users.length === 0 ? (
          <p className="text-[var(--color-text-dim)] text-sm">No users detected</p>
        ) : (
          <div className="space-y-3">
            {users.map((u, i) => (
              <div key={i} className="flex items-center justify-between py-2 border-b border-[var(--color-border)] last:border-0">
                <span className="text-sm font-bold text-[var(--color-text)]">{u.user}</span>
                <span className="text-xs text-[var(--color-text-faint)]">{u.terminal} from {u.host || 'local'}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function InfoRow({ label, value, action, isNumeric }: { label: string; value: string | number; action?: React.ReactNode, isNumeric?: boolean }) {
  return (
    <div className="flex items-center justify-between py-3 border-b border-border/40 last:border-0 group">
      <span className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em]">{label}</span>
      <div className="flex items-center gap-2">
        <span className={cn("text-sm font-bold text-text group-hover:text-accent transition-colors", isNumeric && "tabular-nums")}>
          {String(value)}
        </span>
        {action}
      </div>
    </div>
  )
}
