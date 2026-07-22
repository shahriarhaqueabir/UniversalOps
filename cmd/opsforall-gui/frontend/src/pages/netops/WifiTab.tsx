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
import { Panel, PanelHeader } from '@/components/ui/Panel'
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
    <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl p-5 flex items-center gap-4 transition-all hover:scale-105 active:scale-95 shadow-sm group">
      <div
        className="w-11 h-11 rounded-xl flex items-center justify-center shrink-0 border shadow-inner group-hover:scale-110 transition-transform"
        style={{ backgroundColor: `color-mix(in srgb, ${color} 15%, transparent)`, color, borderColor: `color-mix(in srgb, ${color} 30%, transparent)` }}
      >
        {icon}
      </div>
      <div>
        <p className="text-[10px] font-black text-text-faint uppercase tracking-[0.2em] mb-0.5">{label}</p>
        <p className="text-sm font-black text-text truncate uppercase tracking-tight">{value}</p>
      </div>
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
      <Panel variant="default" padding="md">
        <div className="flex items-start gap-4">
          <div className="w-11 h-11 rounded-xl flex items-center justify-center bg-[var(--color-panel-3)] border border-[var(--color-border)] shrink-0 shadow-inner">
            <Info size={18} className="text-accent" />
          </div>
          <div className="pt-1">
            <p className="text-xs font-black text-[var(--color-text)] uppercase tracking-[0.2em] mb-1">Knowledge Node</p>
            <p className="text-sm text-text-dim leading-relaxed font-medium">
              WiFi Scanner — Scan for nearby wireless networks and view current connection details. Use to assess signal strength, channel congestion, and security.
            </p>
          </div>
        </div>
      </Panel>

      {/* ── Checklist ── */}
      <Panel variant="elevated" padding="md">
        <PanelHeader icon={<ListChecks size={20} />} title="Signal Integrity Checklist" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[
            'Signal strength (dBm) determines connection quality',
            'Channel overlap causes interference',
            'Security type should be WPA2/WPA3',
            'BSSID identifies the access point',
          ].map(item => (
            <div key={item} className="flex items-center gap-3 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl px-5 py-3 shadow-sm">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-[11px] font-bold text-text-dim uppercase tracking-wider">{item}</span>
            </div>
          ))}
        </div>
      </Panel>

      {/* ── Current Connection ── */}
      <Panel variant="default" padding="md" category="network">
        <PanelHeader icon={<Wifi size={20} />} title="Active Handshake" category="network" />
        {infoLoading ? (
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-24 bg-panel-2 border border-border rounded-2xl animate-pulse" />
            ))}
          </div>
        ) : !wifiInfo?.ssid ? (
          <div className="flex flex-col items-center justify-center py-12 opacity-30">
            <Wifi size={40} className="text-text-faint mb-4" />
            <p className="text-xs font-black uppercase tracking-widest text-text-faint">No Wireless Uplink Detected</p>
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
              label="Rate"
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
      </Panel>

      {/* ── Available Networks ── */}
      <Panel variant="default" padding="none" category="network" className="overflow-hidden">
        <div className="p-8">
          <PanelHeader
            icon={<Radio size={20} />}
            title="Available Spectrums"
            category="network"
            action={
              <div className="flex items-center gap-4">
                <span className="text-xs font-black px-4 py-1.5 rounded-full bg-panel-3 text-accent border border-border tabular-nums shadow-inner">
                  {networks.length} SSIDs
                </span>
                <button
                  onClick={() => scanMutation.mutate()}
                  disabled={scanMutation.isPending}
                  className={cn(
                    'flex items-center gap-2 px-4 py-2 text-[10px] font-black uppercase tracking-widest rounded-xl transition-all shadow-lg',
                    scanMutation.isPending
                      ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                      : 'bg-accent text-white hover:opacity-90 active:scale-95',
                  )}
                >
                  {scanMutation.isPending ? (
                    <RefreshCw size={12} className="animate-spin" />
                  ) : (
                    <Scan size={12} />
                  )}
                  {scanMutation.isPending ? 'Probing...' : 'Initiate Scan'}
                </button>
              </div>
            }
          />

          {networksLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-14 bg-panel-2 border border-border rounded-xl animate-pulse" />
              ))}
            </div>
          ) : networks.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 opacity-30">
              <Radio size={40} className="text-text-faint mb-4" />
              <p className="text-xs font-black uppercase tracking-widest text-text-faint">No Spectrums Discovered</p>
            </div>
          ) : (
            <div className="overflow-x-auto -mx-8">
              <table className="w-full text-left border-collapse">
                <thead className="bg-[var(--color-panel-2)] border-y border-[var(--color-border)]/50">
                  <tr>
                    {['SSID', 'Signal', 'Channel', 'Security', 'BSSID'].map(h => (
                      <th key={h} className="px-8 py-4 text-[10px] font-black text-text-faint uppercase tracking-[0.2em]">{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-[var(--color-border)]/30">
                  {networks.map((net, idx) => (
                    <tr
                      key={`${net.ssid}-${idx}`}
                      className="hover:bg-[var(--color-sidebar-hover)] transition-colors group"
                    >
                      <td className="px-8 py-4 text-xs font-black text-text uppercase tracking-tight">{net.ssid || '<HIDDEN>'}</td>
                      <td className="px-8 py-4">
                        <div className="flex items-center gap-3">
                          <SignalBar percent={net.signal} />
                          <span className="text-[10px] font-black text-text-dim tabular-nums tracking-widest">{net.signal}%</span>
                        </div>
                      </td>
                      <td className="px-8 py-4 text-xs font-black text-accent tabular-nums uppercase">{net.channel}</td>
                      <td className="px-8 py-4">
                        <span className={cn(
                          'text-[9px] font-black px-3 py-1 rounded-lg border uppercase tracking-widest',
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
                      <td className="px-8 py-4 text-[10px] font-mono font-bold text-text-faint">{net.bssid}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </Panel>
    </div>
  )
}
