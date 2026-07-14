import { useQuery } from '@tanstack/react-query'
import { Package, Search } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { PackageManagerData } from '@/types'
import { useState } from 'react'

export function PackageManagerTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [search, setSearch] = useState('')

  const { data: managers = [] } = useQuery<PackageManagerData[]>({
    queryKey: ['sysops-packages'],
    queryFn: async () => { const r = await call('SysOps.GetInstalledPackages'); return (r as PackageManagerData[]) || [] },
    refetchInterval: refreshInterval,
  })

  const activeManager = managers.find(m => m.found)

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 px-4 py-2 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl">
          <Package size={14} className="text-[var(--color-accent)]" />
          <span className="text-sm font-bold text-[var(--color-text)]">{activeManager?.name || 'None detected'}</span>
        </div>
        <span className="text-sm text-[var(--color-text-faint)]">{activeManager?.packages.length || 0} packages</span>
      </div>

      {activeManager?.found && (
        <>
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)]" />
            <input
              type="text"
              placeholder="Search packages..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl pl-10 pr-4 py-2 text-sm text-[var(--color-text)] placeholder-[var(--color-text-faint)] focus:outline-none focus:border-[var(--color-accent)]"
            />
          </div>
          <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-xl overflow-hidden">
            <div className="max-h-[500px] overflow-y-auto">
              <table className="w-full text-left">
                <thead className="sticky top-0 bg-[var(--color-panel-2)] border-b border-[var(--color-border)]">
                  <tr>
                    <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Package</th>
                    <th className="px-4 py-3 text-xs font-bold text-[var(--color-text-faint)] uppercase">Version</th>
                  </tr>
                </thead>
                <tbody>
                  {activeManager.packages
                    .filter(p => p.name.toLowerCase().includes(search.toLowerCase()))
                    .map((p, i) => (
                      <tr key={i} className="border-b border-[var(--color-border)]/20 hover:bg-[var(--color-sidebar-hover)]">
                        <td className="px-4 py-2 text-sm font-medium text-[var(--color-text)]">{p.name}</td>
                        <td className="px-4 py-2 text-sm text-[var(--color-text-dim)]">{p.version}</td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {!activeManager?.found && (
        <div className="text-center py-12">
          <Package size={48} className="text-[var(--color-text-faint)] mx-auto mb-4" />
          <p className="text-[var(--color-text-dim)]">No package manager detected</p>
          <p className="text-xs text-[var(--color-text-faint)] mt-2">Supported: apt, dnf, pacman, winget, choco</p>
        </div>
      )}
    </div>
  )
}
