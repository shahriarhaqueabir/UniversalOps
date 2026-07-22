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
import { Panel, PanelHeader } from '@/components/ui/Panel'
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
      <Panel variant="default" padding="md">
        <div className="flex items-start gap-4">
          <div className="w-11 h-11 rounded-xl flex items-center justify-center bg-[var(--color-panel-3)] border border-[var(--color-border)] shrink-0 shadow-inner">
            <Info size={18} className="text-accent" />
          </div>
          <div className="pt-1">
            <p className="text-xs font-black text-[var(--color-text)] uppercase tracking-[0.2em] mb-1">Expert Briefing</p>
            <p className="text-sm text-text-dim leading-relaxed font-medium">
              DNS Advanced — Flush the local DNS cache, perform reverse lookups, and test DNS-over-HTTPS resolvers for privacy and performance.
            </p>
          </div>
        </div>
      </Panel>

      {/* ── Checklist ── */}
      <Panel variant="elevated" padding="md">
        <PanelHeader icon={<ListChecks size={20} />} title="DNS Diagnostics Checklist" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[
            'DNS cache flush clears stale entries',
            'Reverse lookup resolves IP to hostname via PTR',
            'DoH encrypts DNS queries for privacy',
            'Compare latency between DoH servers',
          ].map(item => (
            <div key={item} className="flex items-center gap-3 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-2xl px-5 py-3 shadow-sm">
              <HelpCircle size={14} className="text-accent shrink-0" />
              <span className="text-[11px] font-bold text-text-dim uppercase tracking-wider">{item}</span>
            </div>
          ))}
        </div>
      </Panel>

      {/* ── Flush DNS Cache ── */}
      <Panel variant="default" padding="md" category="security">
        <PanelHeader icon={<Trash2 size={20} />} title="Flush DNS Cache" category="security" />
        <div className="flex items-center gap-4">
          <button
            onClick={() => flushMutation.mutate()}
            disabled={flushMutation.isPending}
            className={cn(
              'flex items-center gap-3 px-6 py-3 text-xs font-black uppercase tracking-widest rounded-xl transition-all shadow-xl',
              flushMutation.isPending
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-danger text-white hover:opacity-90 active:scale-95',
            )}
          >
            {flushMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <Trash2 size={16} />}
            {flushMutation.isPending ? 'Flushing Substrate...' : 'Initiate Flush'}
          </button>
        </div>

        {flushMutation.isSuccess && flushMutation.data && (
          <div className={cn(
            'mt-6 flex items-center gap-3 rounded-xl p-5 border shadow-inner animate-in slide-in-from-top-2',
            flushMutation.data.success
              ? 'bg-success/10 border-success/30 text-success'
              : 'bg-danger/10 border-danger/30 text-danger',
          )}>
            {flushMutation.data.success ? (
              <CheckCircle2 size={18} className="shrink-0" />
            ) : (
              <XCircle size={18} className="shrink-0" />
            )}
            <p className="text-sm font-bold uppercase tracking-tight">
              {flushMutation.data.message || flushMutation.data.error || 'Operation completed'}
            </p>
          </div>
        )}
      </Panel>

      {/* ── Reverse Lookup ── */}
      <Panel variant="default" padding="md" category="network">
        <PanelHeader icon={<ArrowLeftRight size={20} />} title="Reverse Lookup" category="network" />
        <div className="flex items-center gap-3">
          <input
            type="text"
            value={revIp}
            onChange={(e) => setRevIp(e.target.value)}
            placeholder="Enter target IP (e.g. 8.8.8.8)"
            className="flex-1 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl px-5 py-3 text-sm font-bold text-text placeholder:text-text-faint focus:outline-none focus:border-accent transition-all shadow-inner"
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
              'flex items-center gap-3 px-6 py-3 text-xs font-black uppercase tracking-widest rounded-xl transition-all shadow-lg',
              !revIp.trim() || reverseLookupMutation.isPending
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:opacity-90 active:scale-95',
            )}
          >
            {reverseLookupMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <ArrowLeftRight size={16} />}
            RESOLVE
          </button>
        </div>

        {reverseLookupMutation.isSuccess && reverseLookupMutation.data && (
          <div className="mt-6 animate-in slide-in-from-top-2">
            {reverseLookupMutation.data.error ? (
              <div className="flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-5 text-danger font-bold uppercase text-xs">
                <XCircle size={18} className="shrink-0" />
                {reverseLookupMutation.data.error}
              </div>
            ) : (
              <div className="bg-success/10 border border-success/30 rounded-xl p-6 shadow-inner">
                <div className="flex items-center gap-3 mb-4">
                  <CheckCircle2 size={18} className="text-success shrink-0" />
                  <p className="text-[10px] font-black text-success uppercase tracking-widest">PTR Mapping Success</p>
                </div>
                <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-5 shadow-sm">
                  <p className="text-[10px] font-black text-text-faint uppercase tracking-widest mb-1">Resolved Hostname</p>
                  <p className="text-sm font-mono font-black text-accent">{reverseLookupMutation.data.hostname || 'N/A'}</p>
                </div>
              </div>
            )}
          </div>
        )}
      </Panel>

      {/* ── DoH Test ── */}
      <Panel variant="default" padding="md" category="network">
        <PanelHeader icon={<Shield size={20} />} title="DNS-over-HTTPS Integrity" category="network" />
        <div className="flex items-center gap-3">
          <input
            type="text"
            value={dohServer}
            onChange={(e) => setDohServer(e.target.value)}
            placeholder="DoH Endpoint URL"
            className="flex-1 bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl px-5 py-3 text-sm font-bold text-text font-mono focus:outline-none focus:border-accent transition-all shadow-inner"
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
              'flex items-center gap-3 px-6 py-3 text-xs font-black uppercase tracking-widest rounded-xl transition-all shadow-lg',
              !dohServer.trim() || dohMutation.isPending
                ? 'bg-panel-3 text-text-faint border border-border cursor-not-allowed'
                : 'bg-accent text-white hover:opacity-90 active:scale-95',
            )}
          >
            {dohMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <Shield size={16} />}
            PROBE
          </button>
        </div>

        {dohMutation.isSuccess && dohMutation.data && (
          <div className="mt-6 animate-in slide-in-from-top-2">
            {dohMutation.data.success ? (
              <div className="bg-success/10 border border-success/30 rounded-2xl p-6 shadow-inner">
                <div className="flex items-center gap-3 mb-6">
                  <CheckCircle2 size={18} className="text-success shrink-0" />
                  <p className="text-[10px] font-black text-success uppercase tracking-widest">DoH Handshake Verified</p>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-4 shadow-sm">
                    <div className="flex items-center gap-2 mb-2 opacity-50">
                      <Server size={12} />
                      <p className="text-[10px] font-black uppercase tracking-widest">Resolver</p>
                    </div>
                    <p className="text-xs font-mono font-bold text-text truncate">{dohMutation.data.server}</p>
                  </div>
                  <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-4 shadow-sm">
                    <div className="flex items-center gap-2 mb-2 opacity-50">
                      <Clock size={12} />
                      <p className="text-[10px] font-black uppercase tracking-widest">Latency</p>
                    </div>
                    <p className={cn(
                      'text-sm font-black tabular-nums',
                      dohMutation.data.latency_ms < 100 ? 'text-success' : dohMutation.data.latency_ms < 300 ? 'text-warning' : 'text-danger',
                    )}>
                      {dohMutation.data.latency_ms.toFixed(1)}ms
                    </p>
                  </div>
                  <div className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-xl p-4 shadow-sm">
                    <div className="flex items-center gap-2 mb-2 opacity-50">
                      <Globe size={12} />
                      <p className="text-[10px] font-black uppercase tracking-widest">Resolved IP</p>
                    </div>
                    <p className="text-xs font-mono font-bold text-accent">{dohMutation.data.resolved_ip || 'N/A'}</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex items-center gap-3 bg-danger/10 border border-danger/30 rounded-xl p-5 text-danger font-bold uppercase text-xs">
                <XCircle size={18} className="shrink-0" />
                Handshake failed for {dohMutation.data.server}
              </div>
            )}
          </div>
        )}
      </Panel>
    </div>
  )
}
