import { Disc, HardDrive } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { DiskInfo, DiskIOData } from '@/types'

export function DiskTab({ diskInfo }: { diskInfo: DiskInfo }) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: diskIO } = useQuery<DiskIOData>({
    queryKey: ['sysops-disk-io'],
    queryFn: async () => { const r = await call('SysOps.GetDiskIO'); return r as DiskIOData },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-8">
      {/* Partition Usage */}
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Disc size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Disk Usage</h3>
        </div>
        <div className="space-y-6">
          {diskInfo.partitions.map((p, i) => (
            <div key={i}>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-bold text-[var(--color-text)]">{p.mountpoint}</span>
                <span className="text-sm font-bold text-[var(--color-text)] tabular-nums">{p.used_percent.toFixed(1)}%</span>
              </div>
              <div className="h-4 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)]">
                <div className="h-full rounded-full bg-gradient-to-r from-[var(--color-accent)]/60 to-[var(--color-accent)] transition-all duration-700" style={{ width: `${p.used_percent}%` }} />
              </div>
              <div className="flex justify-between mt-1 text-xs text-[var(--color-text-faint)]">
                <span>{(p.total_bytes / 1e9).toFixed(1)} GB total · {(p.free_bytes / 1e9).toFixed(1)} GB free</span>
                <span>{p.fs_type} · {p.device}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Disk I/O */}
      {diskIO && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center gap-3 mb-6">
            <HardDrive size={20} className="text-[var(--color-warning)]" />
            <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Disk I/O</h3>
          </div>
          <div className="grid grid-cols-2 gap-6 mb-6">
            <div className="text-center">
              <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums">{(diskIO.total_read_bytes / 1e9).toFixed(2)} GB</p>
              <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">Total Read</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-[var(--color-text)] tabular-nums">{(diskIO.total_write_bytes / 1e9).toFixed(2)} GB</p>
              <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">Total Write</p>
            </div>
          </div>
          <div className="space-y-3">
            {diskIO.disks.map((d, i) => (
              <div key={i} className="flex items-center justify-between py-2 border-b border-[var(--color-border)]/50 last:border-0">
                <span className="text-sm font-bold text-[var(--color-text)]">{d.name}</span>
                <div className="flex gap-6 text-xs">
                  <span className="text-[var(--color-accent)] tabular-nums">R: {(d.read_bytes / 1e6).toFixed(1)} MB</span>
                  <span className="text-[var(--color-warning)] tabular-nums">W: {(d.write_bytes / 1e6).toFixed(1)} MB</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
