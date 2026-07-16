import { useQuery, useMutation } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import {
  RefreshCw,
  Wifi,
  Radio,
  Signal,
  ListChecks,
  Info,
  HelpCircle,
  Scan,
  Activity,
  Zap,
} from 'lucide-react'
import { SectionHeader } from '@/components/ui/SectionHeader'
import type { WiFiInfoData, WiFiNetworkData } from '@/types'

function SignalBar({ percent }: { percent: number }) {
  const clamped = Math.max(0, Math.min(100, percent))
  const bars = 5
  const filledBars = Math.round((clamped / 100) * bars)

  return (
    <div className="flex items-end gap-0.5" title={`${clamped}%`}>
      {Array.from({ length: bars }).map((_, i) => (
        <div
          key={i}
          className={cn(
            'w-1 rounded-full transition-all',
            i < filledBars ? 'bg-success' : 'bg-panel-3',
          )}
          style={{ height: `${(i + 1) * 3 + 2}px` }}
        />
      ))}
    </div>
  )
}

function StatCard({ icon, label, value, color }: { icon: React.ReactNode; label: string; value: string; color: string }) {
  return (
    <div className="bg-panel-2 border border-border rounded-xl p-4 hover:border-accent/30 transition-colors">
      <div className="flex items-center gap-3 mb-2">
        <div
          className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0"
          style={{ backgroundColor: `color-mix(in srgb, ${color} 15%, transparent)`, color }}
        >
          {icon}
        </div>
        <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">{label}</p>
      </div>
      <p className="text-sm font-bold text-text truncate">{value}</p>
    </div>
  )
}

// ── Main WifiTab ──

export function WifiTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: wifiInfo, isLoading: infoLoading } = useQuery<WiFiInfoData>({
    queryKey: ['wifi-info'],
    queryFn: async () => {
      const res = await call('NetOps.GetWiFiInfo') as WiFiInfoData
      return res
    },
    refetchInterval: refreshInterval,
  })

  const { data: networks = [], isLoading: networksLoading, refetch: refetchNetworks } = useQuery<WiFiNetworkData[]>({
    queryKey: ['wifi-networks'],
    queryFn: async () => {
      const res = await call('NetOps.ScanWiFiNetworks') as WiFiNetworkData[]
      return res || []
    },
  })

  const scanMutation = useMutation({
    mutationFn: async () => {
      const res = await call('NetOps.ScanWiFiNetworks') as WiFiNetworkData[]
      return res || []
    },
    onSuccess: () => {
      refetchNetworks()
    },
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">

      {/* ── Briefing ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-start gap-3">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border shrink-0 mt-0.5">
            <Info size={18} className="text-accent" />
          </div>
          <p className="text-sm text-text-dim leading-relaxed">
            WiFi Scanner — Scan for nearby wireless networks and view current connection details. Use to assess signal strength, channel congestion, and security.
          </p>
        </div>
      </div>

      {/* ── Checklist ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
            <ListChecks size={18} className="text-accent" />
          </div>
          <h3 className="text-sm font-bold text-text uppercase tracking-widest">What to Look For</h3>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {[
            'Signal strength (dBm) determines connection quality',
            'Channel overlap causes interference',
            'Security type should be WPA2/WPA3',
            'BSSID identifies the access point',
          ].map(item => (
            <div key={item} className="flex items-center gap-2.5 bg-panel-2 border border-border rounded-xl px-4 py-2.5">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-xs font-medium text-text-dim">{item}</span>
            </div>
          ))}
        </div>
      </div>

      {/* ── Current Connection ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <SectionHeader icon={<Wifi size={18} className="text-accent" />} title="Current Connection" />
        {infoLoading ? (
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-24 bg-panel-2 border border-border rounded-xl animate-pulse" />
            ))}
          </div>
        ) : !wifiInfo?.ssid ? (
          <div className="flex flex-col items-center justify-center py-8">
            <Wifi size={28} className="text-text-faint mb-3" />
            <p className="text-sm font-medium text-text-faint">Not connected to a WiFi network.</p>
            <p className="text-xs text-text-faint mt-1">WiFi info is unavailable on wired connections.</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <StatCard
              icon={<Radio size={18} />}
              label="Interface"
              value={wifiInfo.interface || 'N/A'}
              color="var(--color-accent)"
            />
            <StatCard
              icon={<Wifi size={18} />}
              label="SSID"
              value={wifiInfo.ssid}
              color="var(--color-accent)"
            />
            <StatCard
              icon={<Signal size={18} />}
              label="Signal"
              value={`${wifiInfo.signal}%`}
              color={wifiInfo.signal > 70 ? 'var(--color-success)' : wifiInfo.signal > 40 ? 'var(--color-warning)' : 'var(--color-danger)'}
            />
            <StatCard
              icon={<Zap size={18} />}
              label="Speed"
              value={wifiInfo.speed || 'N/A'}
              color="var(--color-success)"
            />
            <StatCard
              icon={<Activity size={18} />}
              label="Channel"
              value={String(wifiInfo.channel)}
              color="var(--color-accent)"
            />
          </div>
        )}
      </div>

      {/* ── Available Networks ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
            <Radio size={18} className="text-accent" />
          </div>
          <h3 className="text-sm font-bold text-text uppercase tracking-widest">Available Networks</h3>
          <span className="ml-auto px-2.5 py-0.5 text-[10px] font-bold rounded-full bg-accent/15 text-accent border border-accent/30">
            {networks.length}
          </span>
          <button
            onClick={() => scanMutation.mutate()}
            disabled={scanMutation.isPending}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 text-[11px] font-bold rounded-lg transition-all',
              scanMutation.isPending
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:bg-accent/90',
            )}
          >
            {scanMutation.isPending ? (
              <RefreshCw size={12} className="animate-spin" />
            ) : (
              <Scan size={12} />
            )}
            {scanMutation.isPending ? 'Scanning...' : 'Scan'}
          </button>
        </div>

        {networksLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-12 bg-panel-2 border border-border rounded-xl animate-pulse" />
            ))}
          </div>
        ) : networks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8">
            <Radio size={28} className="text-text-faint mb-3" />
            <p className="text-sm font-medium text-text-faint">No networks found.</p>
            <p className="text-xs text-text-faint mt-1">Click Scan to search for nearby networks.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-border">
                  {['SSID', 'Signal', 'Channel', 'Security', 'BSSID'].map(h => (
                    <th key={h} className="px-4 py-2.5 text-[10px] font-bold text-text-faint uppercase tracking-widest">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {networks.map((net, idx) => (
                  <tr
                    key={`${net.ssid}-${idx}`}
                    className={cn(
                      'border-b border-border last:border-0 hover:bg-panel-2 transition-colors',
                      idx % 2 === 0 ? 'bg-panel-2/30' : 'bg-transparent',
                    )}
                  >
                    <td className="px-4 py-3 text-xs font-bold text-text">{net.ssid || '<Hidden>'}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <SignalBar percent={net.signal} />
                        <span className="text-xs font-bold text-text-dim tabular-nums">{net.signal}%</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs font-bold text-text-dim tabular-nums">{net.channel}</td>
                    <td className="px-4 py-3">
                      <span className={cn(
                        'text-[10px] font-bold px-2 py-0.5 rounded-full border uppercase tracking-wider',
                        /wpa3/i.test(net.security)
                          ? 'bg-success/15 text-success border-success/30'
                          : /wpa2/i.test(net.security)
                            ? 'bg-accent/15 text-accent border-accent/30'
                            : /wep|open/i.test(net.security)
                              ? 'bg-danger/15 text-danger border-danger/30'
                              : 'bg-warning/15 text-warning border-warning/30',
                      )}>
                        {net.security}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-[10px] font-mono text-text-faint">{net.bssid}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
