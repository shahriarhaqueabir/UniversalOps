import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Shield,
  ShieldOff,
  Globe,
  Fingerprint,
  ArrowRightLeft,
  Network,
} from 'lucide-react'
import { SectionBriefing } from './components'
import type { VPNStatusData } from '@/types'

// ── Status Card ──

function VpnFieldCard({
  label,
  value,
  icon,
  highlight,
}: {
  label: string
  value: string
  icon: React.ReactNode
  highlight?: boolean
}) {
  return (
    <div
      className={cn(
        'bg-panel-2 border rounded-xl p-5 transition-all',
        highlight ? 'border-accent/30 hover:border-accent/50' : 'border-border hover:border-accent/20',
      )}
    >
      <div className="flex items-center gap-3 mb-3">
        <div className="w-9 h-9 rounded-lg bg-panel-3 flex items-center justify-center border border-border">
          {icon}
        </div>
        <span className="text-[10px] font-bold text-text-faint uppercase tracking-widest">
          {label}
        </span>
      </div>
      <p className="text-lg font-bold text-text truncate">{value || 'N/A'}</p>
    </div>
  )
}

// ── Main VpnTab ──

export function VpnTab() {
  const { call } = useBackend()

  const { data: vpn, isLoading } = useQuery<VPNStatusData>({
    queryKey: ['netops-vpn-status'],
    queryFn: async () => {
      const res = (await call('NetOps.GetVPNStatus')) as VPNStatusData
      return res
    },
    refetchInterval: 15000,
    retry: false,
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <SectionBriefing
        title="VPN Status"
        objective="Detect active VPN connections including WireGuard, OpenVPN, and other tunnel interfaces. Monitor tunnel health and configuration."
        checklist={[
          'Active VPN adapters are auto-detected',
          'Interface name reveals VPN type',
          'Local IP is assigned by the VPN tunnel',
          'Protocol indicates the tunnel mechanism',
        ]}
      />

      {/* ── VPN Status ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
            <Shield size={18} className="text-accent" />
          </div>
          <h3 className="text-sm font-bold text-text uppercase tracking-widest">VPN Status</h3>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <div className="flex flex-col items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-panel-3 flex items-center justify-center border border-border">
                <Shield size={20} className="text-text-faint animate-pulse" />
              </div>
              <span className="text-xs font-medium text-text-faint">Detecting VPN adapters...</span>
            </div>
          </div>
        ) : vpn?.active ? (
          <div className="space-y-6">
            {/* ── Active Badge ── */}
            <div className="flex items-center gap-3">
              <span className="w-3 h-3 rounded-full bg-success shadow-[0_0_8px_var(--color-success)] animate-pulse" />
              <span className="text-sm font-bold text-success uppercase tracking-widest">
                VPN Active
              </span>
            </div>

            {/* ── VPN Detail Cards ── */}
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
              <VpnFieldCard
                label="VPN Active"
                value="Yes"
                icon={<Shield size={18} className="text-success" />}
                highlight
              />
              <VpnFieldCard
                label="VPN Type"
                value={vpn.type}
                icon={<Fingerprint size={18} className="text-accent" />}
              />
              <VpnFieldCard
                label="Interface"
                value={vpn.interface}
                icon={<Network size={18} className="text-accent" />}
              />
              <VpnFieldCard
                label="Local IP"
                value={vpn.local_ip}
                icon={<Globe size={18} className="text-accent" />}
              />
              <VpnFieldCard
                label="Protocol"
                value={vpn.protocol}
                icon={<ArrowRightLeft size={18} className="text-accent" />}
              />
            </div>
          </div>
        ) : (
          /* ── No VPN Active ── */
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-panel-3 flex items-center justify-center border border-border mb-4">
              <ShieldOff size={32} className="text-text-faint" />
            </div>
            <p className="text-sm font-bold text-text uppercase tracking-widest mb-1">
              No Active VPN Detected
            </p>
            <p className="text-xs font-medium text-text-faint max-w-sm">
              No VPN tunnel interfaces were found. Connect to a VPN and return to this tab to monitor tunnel status.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
