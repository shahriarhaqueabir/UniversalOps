import { useState, useCallback } from 'react'
import { useBackend } from '@/hooks/useBackend'
import { cn } from '@/lib/utils'
import {
  Globe,
  Network,
  Search,
  RefreshCw,
  AlertTriangle,
} from 'lucide-react'
import type { PortResult } from '@/types'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { SearchInput } from '@/components/ui/SearchInput'

export function PortScanTab() {
  const { call } = useBackend()
  const [portScanTarget, setPortScanTarget] = useState('127.0.0.1')
  const [portScanPorts, setPortScanPorts] = useState('21,22,23,25,53,80,110,143,443,445,993,1433,1521,3306,3389,5432,6379,8080,8443,27017')
  const [portScanResults, setPortScanResults] = useState<PortResult[]>([])
  const [portScanLoading, setPortScanLoading] = useState(false)
  const [portScanError, setPortScanError] = useState<string | null>(null)

  const handlePortScan = useCallback(async () => {
    setPortScanLoading(true); setPortScanResults([]); setPortScanError(null)
    try {
      const ports = portScanPorts.split(',').map(p => parseInt(p.trim(), 10)).filter(p => !isNaN(p))
      const res = await call('NetOps.PortScan', portScanTarget, ports) as PortResult[]
      setPortScanResults(res || [])
    } catch (err: unknown) {
      console.error('Port scan failed:', err)
      setPortScanError(err instanceof Error ? err.message : String(err))
    } finally {
      setPortScanLoading(false)
    }
  }, [call, portScanTarget, portScanPorts])

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Port Scan"
        objective="Scan a target host for open ports to identify running services, exposed attack surfaces, or unauthorized listeners."
        checklist={[
          "Open ports indicate services accepting connections.",
          "Common ports (21,22,23,25,53,80,110,143,443,445,993,1433,1521,3306,3389,5432,6379,8080,8443,27017) are pre-configured.",
          "Unrecognized open ports may signal unauthorized services or malware.",
          "Closed or filtered ports are typically invisible to scanners."
        ]}
      />

      {/* Input + Scan Button */}
      <div className="flex items-center gap-4 bg-[var(--color-panel-2)] border border-[var(--color-border)] p-5 rounded-[var(--radius-lg)]">
        <SearchInput
          size="lg"
          icon={<Globe size={18} />}
          value={portScanTarget}
          onChange={(e) => setPortScanTarget(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handlePortScan()}
          placeholder="Target hostname or IP"
          className="flex-1"
        />
        <SearchInput
          size="lg"
          icon={<Network size={18} />}
          value={portScanPorts}
          onChange={(e) => setPortScanPorts(e.target.value)}
          placeholder="Ports (e.g. 80,443,8080)"
          className="w-80"
        />
        <button onClick={handlePortScan} disabled={portScanLoading} className="flex items-center gap-2 px-5 py-2.5 bg-[var(--color-accent)] text-white text-sm font-bold rounded-xl hover:bg-[var(--color-accent)]/90 transition-all h-12 shrink-0 disabled:opacity-50">
          {portScanLoading ? <RefreshCw size={16} className="animate-spin" /> : <Search size={16} />}
          {portScanLoading ? 'SCANNING...' : 'SCAN'}
        </button>
      </div>

      {/* Results Table */}
      {portScanResults.length > 0 && !portScanLoading && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] overflow-hidden shadow-2xl">
          <div className="px-8 py-6 bg-panel-2 border-b border-border flex items-center justify-between">
            <h3 className="text-xl font-bold text-text uppercase tracking-widest flex items-center gap-3">
              <Globe size={20} className="text-accent" /> {portScanTarget}
            </h3>
            <span className="px-4 py-1.5 text-sm font-bold text-text-dim bg-panel-3 rounded-full border border-border/30 uppercase tracking-widest">
              {portScanResults.filter(p => p.open).length}/{portScanResults.length} OPEN
            </span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-panel-2 border-b border-border">
                <tr>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Port</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider">Service</th>
                  <th className="px-8 py-6 text-xs font-semibold text-[var(--color-text-dim)] uppercase tracking-wider text-right">Status</th>
                </tr>
              </thead>
              <tbody>
                {portScanResults.map((p, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-[var(--color-sidebar-hover)] transition-all group">
                    <td className="px-8 py-5">
                      <span className="text-2xl font-bold text-accent tabular-nums">{p.port}</span>
                    </td>
                    <td className="px-8 py-5">
                      <span className="text-sm font-medium text-[var(--color-text)]">{p.service || 'Unknown'}</span>
                    </td>
                    <td className="px-8 py-5 text-right">
                      <span className={cn(
                        "px-4 py-1.5 rounded-full text-sm font-bold uppercase tracking-widest",
                        p.open
                          ? "bg-success/10 text-success border border-success/30"
                          : "bg-danger/10 text-danger border border-danger/30"
                      )}>
                        {p.open ? 'OPEN' : 'CLOSED'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Error state */}
      {portScanError && !portScanLoading && (
        <div className="bg-danger/10 border border-danger/30 rounded-[var(--radius-lg)] p-6 flex items-start gap-3">
          <AlertTriangle size={20} className="text-danger shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold text-danger">Port scan failed</p>
            <p className="text-sm text-[var(--color-text-dim)] mt-1">{portScanError}</p>
          </div>
        </div>
      )}

      {/* Empty state */}
      {portScanResults.length === 0 && !portScanLoading && !portScanError && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-12 shadow-xl text-center">
          <Search size={48} className="mx-auto mb-4 text-text-faint" />
          <p className="text-sm font-medium text-[var(--color-text-dim)]">
            Enter a target host and click SCAN to begin.
          </p>
        </div>
      )}
    </div>
  )
}
