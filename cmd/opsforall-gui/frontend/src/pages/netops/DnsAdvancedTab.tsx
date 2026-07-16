import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import {
  Trash2,
  ArrowLeftRight,
  Shield,
  ListChecks,
  Info,
  HelpCircle,
  CheckCircle2,
  XCircle,
  Loader2,
  Server,
  Clock,
  Globe,
} from 'lucide-react'
import { SectionHeader } from '@/components/ui/SectionHeader'
import type { DoHResultData } from '@/types'

export function DnsAdvancedTab() {
  const { call } = useBackend()

  // ── Flush DNS ──
  const flushMutation = useMutation({
    mutationFn: async () => {
      const res = await call('NetOps.FlushDNSCache') as { success?: boolean; message?: string; error?: string }
      return res
    },
  })

  // ── Reverse Lookup ──
  const [revIp, setRevIp] = useState('')
  const reverseLookupMutation = useMutation({
    mutationFn: async (ip: string) => {
      const res = await call('NetOps.ReverseLookup', ip) as { hostname?: string; error?: string }
      return res
    },
  })

  // ── DoH Test ──
  const [dohServer, setDohServer] = useState('https://dns.google/dns-query')
  const dohMutation = useMutation({
    mutationFn: async (server: string) => {
      const res = await call('NetOps.TestDoH', server) as DoHResultData
      return res
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
            DNS Advanced — Flush the local DNS cache, perform reverse lookups, and test DNS-over-HTTPS resolvers for privacy and performance.
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
            'DNS cache flush clears stale entries',
            'Reverse lookup resolves IP to hostname via PTR',
            'DoH encrypts DNS queries for privacy',
            'Compare latency between DoH servers',
          ].map(item => (
            <div key={item} className="flex items-center gap-2.5 bg-panel-2 border border-border rounded-xl px-4 py-2.5">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-xs font-medium text-text-dim">{item}</span>
            </div>
          ))}
        </div>
      </div>

      {/* ── Flush DNS Cache ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <SectionHeader icon={<Trash2 size={18} className="text-accent" />} title="Flush DNS Cache" />
        <div className="flex items-center gap-4">
          <button
            onClick={() => flushMutation.mutate()}
            disabled={flushMutation.isPending}
            className={cn(
              'flex items-center gap-3 px-5 py-2.5 text-sm font-semibold rounded-xl transition-all shadow-xl',
              flushMutation.isPending
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-danger text-white hover:bg-danger/90',
            )}
          >
            {flushMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <Trash2 size={16} />}
            {flushMutation.isPending ? 'Flushing...' : 'Flush DNS Cache'}
          </button>
        </div>

        {flushMutation.isSuccess && flushMutation.data && (
          <div className={cn(
            'mt-4 flex items-center gap-3 rounded-xl p-4',
            flushMutation.data.success
              ? 'bg-success/10 border border-success/30'
              : 'bg-danger/10 border border-danger/30',
          )}>
            {flushMutation.data.success ? (
              <CheckCircle2 size={16} className="text-success shrink-0" />
            ) : (
              <XCircle size={16} className="text-danger shrink-0" />
            )}
            <p className={cn('text-sm font-medium', flushMutation.data.success ? 'text-success' : 'text-danger')}>
              {flushMutation.data.message || flushMutation.data.error || 'Operation completed'}
            </p>
          </div>
        )}

        {flushMutation.isError && (
          <div className="mt-4 flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-4">
            <XCircle size={16} className="text-danger shrink-0" />
            <p className="text-sm font-medium text-danger">
              {flushMutation.error?.message || 'Failed to flush DNS cache'}
            </p>
          </div>
        )}
      </div>

      {/* ── Reverse Lookup ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <SectionHeader icon={<ArrowLeftRight size={18} className="text-accent" />} title="Reverse Lookup" />
        <div className="flex items-center gap-3">
          <input
            type="text"
            value={revIp}
            onChange={(e) => setRevIp(e.target.value)}
            placeholder="Enter IP address (e.g. 8.8.8.8)"
            className="flex-1 bg-panel-2 border border-border rounded-xl px-4 py-2.5 text-sm text-text placeholder:text-text-faint focus:outline-none focus:border-accent/50 transition-colors"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && revIp.trim()) {
                reverseLookupMutation.mutate(revIp.trim())
              }
            }}
          />
          <button
            onClick={() => {
              if (revIp.trim()) reverseLookupMutation.mutate(revIp.trim())
            }}
            disabled={!revIp.trim() || reverseLookupMutation.isPending}
            className={cn(
              'flex items-center gap-2 px-5 py-2.5 text-sm font-semibold rounded-xl transition-all',
              !revIp.trim() || reverseLookupMutation.isPending
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:bg-accent/90',
            )}
          >
            {reverseLookupMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <ArrowLeftRight size={16} />}
            {reverseLookupMutation.isPending ? 'Looking up...' : 'Lookup'}
          </button>
        </div>

        {reverseLookupMutation.isSuccess && reverseLookupMutation.data && (
          <div className="mt-4">
            {reverseLookupMutation.data.error ? (
              <div className="flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-4">
                <XCircle size={16} className="text-danger shrink-0" />
                <p className="text-sm font-medium text-danger">{reverseLookupMutation.data.error}</p>
              </div>
            ) : (
              <div className="bg-success/10 border border-success/30 rounded-xl p-4">
                <div className="flex items-center gap-3 mb-2">
                  <CheckCircle2 size={16} className="text-success shrink-0" />
                  <p className="text-xs font-bold text-success uppercase tracking-wider">PTR Record Found</p>
                </div>
                <div className="bg-panel-2 border border-border rounded-lg p-3">
                  <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider mb-1">Hostname</p>
                  <p className="text-sm font-mono font-bold text-text">{reverseLookupMutation.data.hostname || 'N/A'}</p>
                </div>
              </div>
            )}
          </div>
        )}

        {reverseLookupMutation.isError && (
          <div className="mt-4 flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-4">
            <XCircle size={16} className="text-danger shrink-0" />
            <p className="text-sm font-medium text-danger">
              {reverseLookupMutation.error?.message || 'Reverse lookup failed'}
            </p>
          </div>
        )}
      </div>

      {/* ── DoH Test ── */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-6 shadow-xl">
        <SectionHeader icon={<Shield size={18} className="text-accent" />} title="DNS-over-HTTPS Test" />
        <div className="flex items-center gap-3">
          <input
            type="text"
            value={dohServer}
            onChange={(e) => setDohServer(e.target.value)}
            placeholder="DoH server URL (e.g. https://dns.google/dns-query)"
            className="flex-1 bg-panel-2 border border-border rounded-xl px-4 py-2.5 text-sm text-text font-mono placeholder:text-text-faint focus:outline-none focus:border-accent/50 transition-colors"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && dohServer.trim()) {
                dohMutation.mutate(dohServer.trim())
              }
            }}
          />
          <button
            onClick={() => {
              if (dohServer.trim()) dohMutation.mutate(dohServer.trim())
            }}
            disabled={!dohServer.trim() || dohMutation.isPending}
            className={cn(
              'flex items-center gap-2 px-5 py-2.5 text-sm font-semibold rounded-xl transition-all',
              !dohServer.trim() || dohMutation.isPending
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:bg-accent/90',
            )}
          >
            {dohMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <Shield size={16} />}
            {dohMutation.isPending ? 'Testing...' : 'Test DoH'}
          </button>
        </div>

        {dohMutation.isSuccess && dohMutation.data && (
          <div className="mt-4">
            {dohMutation.data.success ? (
              <div className="bg-success/10 border border-success/30 rounded-xl p-5">
                <div className="flex items-center gap-3 mb-4">
                  <CheckCircle2 size={16} className="text-success shrink-0" />
                  <p className="text-xs font-bold text-success uppercase tracking-wider">DoH Test Passed</p>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="bg-panel-2 border border-border rounded-lg p-3">
                    <div className="flex items-center gap-2 mb-1">
                      <Server size={12} className="text-text-faint" />
                      <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Server</p>
                    </div>
                    <p className="text-xs font-mono font-bold text-text truncate">{dohMutation.data.server}</p>
                  </div>
                  <div className="bg-panel-2 border border-border rounded-lg p-3">
                    <div className="flex items-center gap-2 mb-1">
                      <Clock size={12} className="text-text-faint" />
                      <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Latency</p>
                    </div>
                    <p className={cn(
                      'text-xs font-bold tabular-nums',
                      dohMutation.data.latency_ms < 100 ? 'text-success' : dohMutation.data.latency_ms < 300 ? 'text-warning' : 'text-danger',
                    )}>
                      {dohMutation.data.latency_ms.toFixed(1)}ms
                    </p>
                  </div>
                  <div className="bg-panel-2 border border-border rounded-lg p-3">
                    <div className="flex items-center gap-2 mb-1">
                      <Globe size={12} className="text-text-faint" />
                      <p className="text-[10px] font-bold text-text-faint uppercase tracking-wider">Resolved IP</p>
                    </div>
                    <p className="text-xs font-mono font-bold text-text">{dohMutation.data.resolved_ip || 'N/A'}</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-4">
                <XCircle size={16} className="text-danger shrink-0" />
                <p className="text-sm font-medium text-danger">DoH test failed for {dohMutation.data.server}</p>
              </div>
            )}
          </div>
        )}

        {dohMutation.isError && (
          <div className="mt-4 flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-4">
            <XCircle size={16} className="text-danger shrink-0" />
            <p className="text-sm font-medium text-danger">
              {dohMutation.error?.message || 'DoH test failed'}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
