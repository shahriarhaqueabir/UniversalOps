import { Disc, HardDrive, Trash2, Zap } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { Panel } from '@/components/ui/Panel'
import type { DiskInfo, DiskIOData, ActionPreview } from '@/types'
import { ConfirmationModal } from '@/components/dialogs/ConfirmationModal'
import { useState } from 'react'
import { toast } from 'sonner'

export function DiskTab({ diskInfo }: { diskInfo: DiskInfo }) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [preview, setPreview] = useState<ActionPreview | null>(null)

  const { data: diskIO } = useQuery<DiskIOData>({
    queryKey: ['sysops-disk-io'],
    queryFn: async () => { const r = await call('SysOps.GetDiskIO'); return r as DiskIOData },
    refetchInterval: refreshInterval,
  })

  const requestAction = async (action: string) => {
    const p = await call('SysOps.RunSystemAction', action) as ActionPreview
    setPreview(p)
  }

  const handleConfirm = async () => {
    if (!preview) return
    const handshakeID = preview.handshake_id
    setPreview(null)
    const res = await call('App.ConfirmAction', handshakeID) as { success: boolean; error?: string }
    if (res.success) {
      toast.success('Action initiated successfully')
    } else {
      toast.error(res.error || 'Action failed')
    }
  }

  return (
    <div className="space-y-8">
      <ConfirmationModal
        preview={preview}
        onConfirm={handleConfirm}
        onCancel={() => setPreview(null)}
      />

      <div className="flex gap-4">
        <button onClick={() => requestAction('disk_cleanup')} className="flex-1 flex items-center justify-center gap-3 p-4 rounded-xl bg-orange-500/10 border border-orange-500/20 text-orange-500 hover:bg-orange-500/20 transition-all font-bold text-xs uppercase tracking-widest">
          <Trash2 size={16} /> Run Disk Cleanup
        </button>
        <button onClick={() => requestAction('defrag')} className="flex-1 flex items-center justify-center gap-3 p-4 rounded-xl bg-accent/10 border border-accent/20 text-accent hover:bg-accent/20 transition-all font-bold text-xs uppercase tracking-widest">
          <Zap size={16} /> Defragment Drive
        </button>
      </div>

      {/* Partition Usage */}
      <Panel variant="elevated" padding="lg" category="system">
        <div className="flex items-center gap-3 mb-6">
          <Disc size={20} className="text-[var(--color-accent)]" />
          <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Disk Usage</h3>
        </div>
        <div className="space-y-6">
          {diskInfo.partitions.map((p) => (
            <div key={p.mountpoint}>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-bold text-[var(--color-text)]">{p.mountpoint}</span>
                <span className="text-sm font-bold text-[var(--color-text)] tabular-nums">{Math.round(p.used_percent)}%</span>
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
      </Panel>

      {diskIO && (
        <Panel variant="elevated" padding="lg" category="system">
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
            {diskIO.disks.map((d) => (
              <div key={d.name} className="flex items-center justify-between py-2 border-b border-[var(--color-border)]/50 last:border-0">
                <span className="text-sm font-bold text-[var(--color-text)]">{d.name}</span>
                <div className="flex gap-6 text-xs">
                  <span className="text-[var(--color-accent)] tabular-nums">R: {(d.read_bytes / 1e6).toFixed(1)} MB</span>
                  <span className="text-[var(--color-warning)] tabular-nums">W: {(d.write_bytes / 1e6).toFixed(1)} MB</span>
                </div>
              </div>
            ))}
          </div>
        </Panel>
      )}
    </div>
  )
}
