import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Trash2,
  RefreshCw,
  Eraser,
  RotateCcw,
  CheckCircle2,
  XCircle,
  Zap,
} from 'lucide-react'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import type { NetworkActionResult } from '@/types'

// ── Action Result Banner ──

function ActionBanner({ result }: { result: NetworkActionResult }) {
  return (
    <div
      className={cn(
        'flex items-center gap-3 px-5 py-3 rounded-xl border text-sm font-medium',
        result.success
          ? 'bg-success/10 border-success/30 text-success'
          : 'bg-danger/10 border-danger/30 text-danger',
      )}
    >
      {result.success ? (
        <CheckCircle2 size={16} className="shrink-0" />
      ) : (
        <XCircle size={16} className="shrink-0" />
      )}
      <span>{result.message}</span>
    </div>
  )
}

// ── Action Card ──

interface ActionCardProps {
  title: string
  description: string
  icon: React.ReactNode
  buttonText: string
  loadingText: string
  onClick: () => void
  isRunning: boolean
  result: NetworkActionResult | null
  extraInput?: React.ReactNode
}

function ActionCard({
  title,
  description,
  icon,
  buttonText,
  loadingText,
  onClick,
  isRunning,
  result,
  extraInput,
}: ActionCardProps) {
  return (
    <div className="bg-panel-2 border border-border rounded-xl p-5 transition-all hover:border-accent/30">
      <div className="flex items-center gap-3 mb-3">
        <div className="w-10 h-10 rounded-xl bg-panel-3 flex items-center justify-center border border-border">
          {icon}
        </div>
        <div className="min-w-0">
          <h4 className="text-xs font-bold text-text uppercase tracking-wider">{title}</h4>
          <p className="text-[11px] font-medium text-text-faint mt-0.5">{description}</p>
        </div>
      </div>

      {extraInput && <div className="mb-3">{extraInput}</div>}

      <button
        onClick={onClick}
        disabled={isRunning}
        className={cn(
          'w-full flex items-center justify-center gap-2 px-4 py-2.5 text-xs font-bold rounded-xl transition-all',
          isRunning
            ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
            : 'bg-accent/15 text-accent border border-accent/30 hover:bg-accent/25',
        )}
      >
        {isRunning ? (
          <>
            <RefreshCw size={14} className="animate-spin" />
            {loadingText}
          </>
        ) : (
          buttonText
        )}
      </button>

      {result && <div className="mt-3"><ActionBanner result={result} /></div>}
    </div>
  )
}

// ── Main ActionsTab ──

export function ActionsTab() {
  const { call } = useBackend()
  const [resetIface, setResetIface] = useState('')
  const [results, setResults] = useState<Record<string, NetworkActionResult>>({})

  const setActionResult = (key: string, result: NetworkActionResult) => {
    setResults(prev => ({ ...prev, [key]: result }))
  }

  const flushDns = useMutation({
    mutationFn: async () => {
      return (await call('NetOps.RunNetworkAction', 'flush_dns', {})) as NetworkActionResult
    },
    onSuccess: (res) => setActionResult('flush_dns', res),
    onError: () => setActionResult('flush_dns', { action: 'flush_dns', success: false, message: 'DNS flush failed' }),
  })

  const renewDhcp = useMutation({
    mutationFn: async () => {
      return (await call('NetOps.RunNetworkAction', 'renew_dhcp', {})) as NetworkActionResult
    },
    onSuccess: (res) => setActionResult('renew_dhcp', res),
    onError: () => setActionResult('renew_dhcp', { action: 'renew_dhcp', success: false, message: 'DHCP renewal failed' }),
  })

  const clearArp = useMutation({
    mutationFn: async () => {
      return (await call('NetOps.RunNetworkAction', 'clear_arp_cache', {})) as NetworkActionResult
    },
    onSuccess: (res) => setActionResult('clear_arp_cache', res),
    onError: () => setActionResult('clear_arp_cache', { action: 'clear_arp_cache', success: false, message: 'ARP clear failed' }),
  })

  const resetInterface = useMutation({
    mutationFn: async () => {
      return (await call('NetOps.RunNetworkAction', 'reset_interface', { interface: resetIface })) as NetworkActionResult
    },
    onSuccess: (res) => setActionResult('reset_interface', res),
    onError: () => setActionResult('reset_interface', { action: 'reset_interface', success: false, message: 'Interface reset failed' }),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <SectionBriefing
        title="Network Actions"
        objective="Execute common network maintenance tasks including DNS cache flush, DHCP renewal, interface reset, and ARP cache clear."
        checklist={[
          'Flush DNS clears stale resolution cache',
          'DHCP renewal requests a fresh IP lease',
          'Interface reset can resolve connectivity issues',
          'ARP cache clear forces re-discovery of neighbors',
        ]}
      />

      {/* ── Action Cards Grid ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-panel-3 border border-border">
            <Zap size={18} className="text-accent" />
          </div>
          <h3 className="text-sm font-bold text-text uppercase tracking-widest">Maintenance Actions</h3>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          <ActionCard
            title="Flush DNS Cache"
            description="Clear the local DNS resolver cache"
            icon={<Trash2 size={20} className="text-accent" />}
            buttonText="Flush DNS"
            loadingText="Flushing..."
            onClick={() => flushDns.mutate()}
            isRunning={flushDns.isPending}
            result={results.flush_dns}
          />

          <ActionCard
            title="Renew DHCP"
            description="Request a fresh IP lease from the DHCP server"
            icon={<RefreshCw size={20} className="text-accent" />}
            buttonText="Renew DHCP"
            loadingText="Renewing..."
            onClick={() => renewDhcp.mutate()}
            isRunning={renewDhcp.isPending}
            result={results.renew_dhcp}
          />

          <ActionCard
            title="Clear ARP Cache"
            description="Flush the ARP neighbor cache"
            icon={<Eraser size={20} className="text-accent" />}
            buttonText="Clear ARP"
            loadingText="Clearing..."
            onClick={() => clearArp.mutate()}
            isRunning={clearArp.isPending}
            result={results.clear_arp_cache}
          />

          <ActionCard
            title="Reset Interface"
            description="Disable and re-enable the network interface"
            icon={<RotateCcw size={20} className="text-accent" />}
            buttonText="Reset"
            loadingText="Resetting..."
            onClick={() => {
              if (resetIface.trim()) resetInterface.mutate()
            }}
            isRunning={resetInterface.isPending}
            result={results.reset_interface}
            extraInput={
              <input
                type="text"
                value={resetIface}
                onChange={(e) => setResetIface(e.target.value)}
                placeholder="Interface name (e.g., Ethernet)"
                className={cn(
                  'w-full px-4 py-2.5 text-xs font-medium text-text bg-panel-3 border border-border rounded-xl',
                  'placeholder:text-text-faint focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/30 transition-all',
                )}
              />
            }
          />
        </div>
      </div>
    </div>
  )
}
