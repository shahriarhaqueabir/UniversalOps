import { useQuery } from '@tanstack/react-query'
import { Package, ArrowUpDown, ArrowUp, ArrowDown, Copy, X, AlertCircle } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { PackageManagerData } from '@/types'
import { useState, useRef, useMemo } from 'react'
import { SearchInput } from '@/components/ui/SearchInput'
import { Panel } from '@/components/ui/Panel'
import { useVirtualizer } from '@tanstack/react-virtual'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

type SortKey = 'name' | 'version'
type SortDir = 'asc' | 'desc'

export function PackageManagerTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [search, setSearch] = useState('')
  const [sortKey, setSortKey] = useState<SortKey>('name')
  const [sortDir, setSortDir] = useState<SortDir>('asc')

  const { data: managers = [] } = useQuery<PackageManagerData[]>({
    queryKey: ['sysops-packages'],
    queryFn: async () => { const r = await call('SysOps.GetInstalledPackages'); return (r as PackageManagerData[]) || [] },
    refetchInterval: refreshInterval,
  })

  const displayName = (name: string) => {
    const labels: Record<string, string> = {
      'windows-installed': 'Windows Apps (Registry)',
      'winget': 'WinGet',
      'choco': 'Chocolatey',
    }
    return labels[name] || name.charAt(0).toUpperCase() + name.slice(1)
  }

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDir('asc')
    }
  }

  const SortIcon = ({ column }: { column: SortKey }) => {
    if (sortKey !== column) return <ArrowUpDown size={12} className="opacity-30 group-hover:opacity-60 transition-opacity" />
    return sortDir === 'asc' ? <ArrowUp size={12} /> : <ArrowDown size={12} />
  }

  const activeManager = managers.find(m => m.found)

  const filteredPackages = useMemo(() => {
    if (!activeManager) return []
    const list = activeManager.packages.filter(p =>
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.version?.toLowerCase().includes(search.toLowerCase())
    )
    list.sort((a, b) => {
      const aVal = (sortKey === 'name' ? a.name : a.version || '')
      const bVal = (sortKey === 'name' ? b.name : b.version || '')
      const cmp = aVal.localeCompare(bVal, undefined, { numeric: true })
      return sortDir === 'asc' ? cmp : -cmp
    })
    return list
  }, [activeManager, search, sortKey, sortDir])

  const totalCount = activeManager?.packages.length ?? 0

  const copyName = (name: string) => {
    navigator.clipboard.writeText(name)
    toast.success(`Copied "${name}"`)
  }

  const parentRef = useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: filteredPackages.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 41,
    overscan: 15,
  })

  return (
    <div className="space-y-5">
      {/* ── Header bar ── */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2.5 px-4 py-2 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl">
          <Package size={15} className="text-[var(--color-accent)]" />
          <span className="text-sm font-bold text-[var(--color-text)]" data-testid="manager-name">{activeManager ? displayName(activeManager.name) : 'None detected'}</span>
        </div>
        <div className="flex items-center gap-3 text-sm text-[var(--color-text-faint)]">
          <span className="tabular-nums">{totalCount}</span>
          <span>packages</span>
          {search && filteredPackages.length < totalCount && (
            <>
              <span className="text-[var(--color-text-dim)]">·</span>
              <span className="tabular-nums text-[var(--color-accent)]">{filteredPackages.length}</span>
              <span>filtered</span>
            </>
          )}
        </div>
      </div>

      {activeManager?.found ? (
        <>
          {/* ── Search ── */}
          <div className="relative">
            <SearchInput
              size="md"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search by name or version…"
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                title="Clear"
                className="absolute right-2.5 top-1/2 -translate-y-1/2 p-1 rounded-md text-[var(--color-text-faint)] hover:text-[var(--color-text)] hover:bg-[var(--color-panel-3)] transition-colors"
              >
                <X size={14} />
              </button>
            )}
          </div>

          {/* ── Table ── */}
          <Panel variant="default" padding="none" category="system">
            <div ref={parentRef} className="max-h-[560px] overflow-y-auto">
              {/* Sticky header — rendered inside the scroll container for perfect alignment */}
              <table className="w-full text-left" style={{ tableLayout: 'fixed' }}>
                <colgroup>
                  <col className="w-10" />
                  <col />
                  <col className="w-[180px]" />
                  <col className="w-12" />
                </colgroup>
                <thead className="sticky top-0 z-10 bg-[var(--color-panel-2)] border-b border-[var(--color-border)]">
                  <tr>
                    <th className="px-2 py-3 text-right text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-wider select-none">#</th>
                    <th className="px-4 py-3">
                      <button
                        onClick={() => toggleSort('name')}
                        className="group flex items-center gap-1.5 text-xs font-bold text-[var(--color-text-faint)] uppercase tracking-wider select-none hover:text-[var(--color-text)] transition-colors"
                      >
                        Package <SortIcon column="name" />
                      </button>
                    </th>
                    <th className="px-4 py-3">
                      <button
                        onClick={() => toggleSort('version')}
                        className="group flex items-center gap-1.5 text-xs font-bold text-[var(--color-text-faint)] uppercase tracking-wider select-none hover:text-[var(--color-text)] transition-colors"
                      >
                        Version <SortIcon column="version" />
                      </button>
                    </th>
                    <th className="px-2 py-3" />
                  </tr>
                </thead>
                <tbody
                  style={{
                    height: `${rowVirtualizer.getTotalSize()}px`,
                    width: '100%',
                    position: 'relative',
                  }}
                >
                  {filteredPackages.length === 0 ? (
                    <tr>
                      <td colSpan={4} className="px-4 py-16 text-center">
                        <div className="flex flex-col items-center gap-2">
                          <AlertCircle size={24} className="text-[var(--color-text-faint)]" />
                          <p className="text-sm text-[var(--color-text-dim)]">No packages match your search</p>
                          <button
                            onClick={() => setSearch('')}
                            className="text-xs text-[var(--color-accent)] hover:underline"
                          >
                            Clear filter
                          </button>
                        </div>
                      </td>
                    </tr>
                  ) : (
                    rowVirtualizer.getVirtualItems().map(virtualRow => {
                      const p = filteredPackages[virtualRow.index]
                      const rowNum = virtualRow.index + 1
                      return (
                        <tr
                          key={virtualRow.key}
                          data-index={virtualRow.index}
                          data-testid="package-row"
                          ref={rowVirtualizer.measureElement}
                          style={{
                            position: 'absolute',
                            top: 0,
                            left: 0,
                            width: '100%',
                            height: `${virtualRow.size}px`,
                            transform: `translateY(${virtualRow.start}px)`,
                          }}
                          className={cn(
                            "border-b border-[var(--color-border)]/10 group transition-colors",
                            "hover:bg-[var(--color-sidebar-hover)]"
                          )}
                        >
                          <td className="px-2 py-2.5 text-right text-[11px] font-mono text-[var(--color-text-faint)] tabular-nums select-none" data-testid="row-number">
                            {rowNum}
                          </td>
                          <td className="px-4 py-2.5 min-w-0">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium text-[var(--color-text)] truncate" data-testid="package-name">{p.name}</span>
                            </div>
                          </td>
                          <td className="px-4 py-2.5">
                            {p.version ? (
                              <span className="inline-block px-2.5 py-0.5 text-[11px] font-semibold font-mono bg-[var(--color-panel-3)] text-[var(--color-text-dim)] border border-[var(--color-border)]/30 rounded-md leading-normal" data-testid="version-badge">
                                {p.version}
                              </span>
                            ) : (
                              <span className="text-[11px] text-[var(--color-text-faint)] italic" data-testid="version-badge">—</span>
                            )}
                          </td>
                          <td className="px-2 py-2.5 text-right">
                            <button
                              onClick={() => copyName(p.name)}
                              data-testid="copy-button"
                              className="opacity-0 group-hover:opacity-100 p-1.5 rounded-md text-[var(--color-text-faint)] hover:text-[var(--color-accent)] hover:bg-[var(--color-panel-3)] transition-all"
                              title="Copy package name"
                            >
                              <Copy size={13} />
                            </button>
                          </td>
                        </tr>
                      )
                    })
                  )}
                </tbody>
              </table>
            </div>
          </Panel>

          {/* ── Footer stats ── */}
          {filteredPackages.length > 0 && (
            <div className="flex items-center justify-between text-[11px] text-[var(--color-text-faint)] px-1">
              <span>
                Showing <span className="tabular-nums font-medium text-[var(--color-text-dim)]">{filteredPackages.length}</span> of{' '}
                <span className="tabular-nums font-medium text-[var(--color-text-dim)]">{totalCount}</span> packages
              </span>
              <span className="tabular-nums">
                Sorted by {sortKey} ({sortDir === 'asc' ? 'A→Z' : 'Z→A'})
              </span>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-16">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] mb-5">
            <Package size={28} className="text-[var(--color-text-faint)]" />
          </div>
          <p className="text-[var(--color-text-dim)] font-medium">No package manager detected</p>
          <p className="text-xs text-[var(--color-text-faint)] mt-2 max-w-xs mx-auto leading-relaxed">
            Install winget, chocolatey, or a Linux package manager to view installed packages.
          </p>
        </div>
      )}
    </div>
  )
}
