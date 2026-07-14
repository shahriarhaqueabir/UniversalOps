import { MemoryStick } from 'lucide-react'
import type { MemoryInfo } from '@/types'

interface MemoryTabProps {
  memInfo: MemoryInfo
}

export function MemoryTab({ memInfo }: MemoryTabProps) {
  const availableGB = memInfo.total_gb - memInfo.used_gb
  const cachedGB = memInfo.cached_bytes / (1024 * 1024 * 1024)
  const swapUsedGB = memInfo.swap_used / (1024 * 1024 * 1024)
  const swapTotalGB = memInfo.swap_total / (1024 * 1024 * 1024)

  return (
    <div className="space-y-8">
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <MemoryStick size={20} className="text-[var(--color-success)]" />
            <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">RAM Usage</h3>
          </div>
          <span className="text-3xl font-bold text-[var(--color-success)] tabular-nums">{memInfo.used_percent.toFixed(1)}%</span>
        </div>
        <div className="h-6 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)] mb-6">
          <div className="h-full rounded-full bg-gradient-to-r from-[var(--color-success)]/60 to-[var(--color-success)] transition-all duration-700" style={{ width: `${memInfo.used_percent}%` }} />
        </div>
        <div className="grid grid-cols-4 gap-6">
          <StatBox label="Total" value={`${memInfo.total_gb.toFixed(1)} GB`} />
          <StatBox label="Used" value={`${memInfo.used_gb.toFixed(1)} GB`} color="text-[var(--color-success)]" />
          <StatBox label="Available" value={`${availableGB.toFixed(1)} GB`} />
          <StatBox label="Cached" value={`${cachedGB.toFixed(1)} GB`} />
        </div>
      </div>

      {memInfo.swap_total > 0 && (
        <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-bold text-[var(--color-text)] uppercase tracking-widest">Swap Usage</h3>
            <span className="text-2xl font-bold text-[var(--color-warning)] tabular-nums">{memInfo.swap_percent.toFixed(1)}%</span>
          </div>
          <div className="h-4 bg-[var(--color-panel-3)] rounded-full overflow-hidden border border-[var(--color-border)]">
            <div className="h-full rounded-full bg-gradient-to-r from-[var(--color-warning)]/60 to-[var(--color-warning)] transition-all duration-700" style={{ width: `${memInfo.swap_percent}%` }} />
          </div>
          <div className="flex justify-between mt-3 text-sm">
            <span className="text-[var(--color-text-dim)]">{swapUsedGB.toFixed(1)} GB used</span>
            <span className="text-[var(--color-text-faint)]">{swapTotalGB.toFixed(1)} GB total</span>
          </div>
        </div>
      )}
    </div>
  )
}

function StatBox({ label, value, color = 'text-[var(--color-text)]' }: { label: string; value: string; color?: string }) {
  return (
    <div className="text-center">
      <p className={`text-xl font-bold tabular-nums ${color}`}>{value}</p>
      <p className="text-xs font-bold text-[var(--color-text-faint)] uppercase">{label}</p>
    </div>
  )
}
