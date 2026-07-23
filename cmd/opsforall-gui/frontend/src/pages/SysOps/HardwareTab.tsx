import { useState, useCallback } from 'react'
import { Monitor, Thermometer, Wind, Database, Play, Square, ShieldAlert, ShieldCheck, Loader2, CheckCircle2, AlertTriangle } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { Panel, PanelHeader } from '@/components/ui/Panel'
import { StatCard } from '@/components/ui/StatCard'
import type { HardwareInfo, SensorData, LHMStatusResult, LHMAuthorization } from '@/types'
import { cn } from '@/lib/utils'

// ── LHM Admin Elevation Dialog ──────────────────────────────────────────────

function LHMElevationDialog({
  auth,
  onConfirm,
  onCancel,
  isStarting,
}: {
  auth: LHMAuthorization
  onConfirm: () => void
  onCancel: () => void
  isStarting: boolean
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-lg mx-4 rounded-2xl border border-[var(--color-border)] bg-[var(--color-panel)] shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center gap-3 px-6 py-5 border-b border-[var(--color-border)] bg-warning/5">
          <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-warning/15">
            <ShieldAlert size={22} className="text-warning" />
          </div>
          <div>
            <h3 className="text-sm font-black text-[var(--color-text)] uppercase tracking-wide">Administrator Privileges Required</h3>
            <p className="text-[10px] font-bold text-[var(--color-text-dim)] mt-0.5">Windows User Account Control (UAC)</p>
          </div>
        </div>

        {/* Body */}
        <div className="px-6 py-5 space-y-5">
          {/* Reason */}
          <p className="text-xs leading-relaxed text-[var(--color-text-dim)]">{auth.reason}</p>

          {/* What you get */}
          <div>
            <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-2">What sensors will be available</p>
            <ul className="space-y-1.5">
              {auth.capabilities.map((cap, i) => (
                <li key={i} className="flex items-start gap-2">
                  <CheckCircle2 size={13} className="text-success mt-0.5 shrink-0" />
                  <span className="text-xs text-[var(--color-text)]">{cap}</span>
                </li>
              ))}
            </ul>
          </div>

          {/* Safety notes */}
          <div className="rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]/50 p-4">
            <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-2">Safety & transparency</p>
            <ul className="space-y-1.5">
              {auth.risks.map((risk, i) => (
                <li key={i} className="flex items-start gap-2">
                  <AlertTriangle size={13} className="text-[var(--color-text-dim)] mt-0.5 shrink-0" />
                  <span className="text-[11px] text-[var(--color-text-dim)]">{risk}</span>
                </li>
              ))}
            </ul>
          </div>

          {/* Binary info */}
          <div className="flex items-center gap-3 p-3 rounded-xl border border-[var(--color-border)]/50 bg-[var(--color-panel-2)]">
            <ShieldCheck size={18} className="text-success shrink-0" />
            <div className="min-w-0">
              <p className="text-[11px] font-bold text-[var(--color-text)] truncate">{auth.binaryName}</p>
              <p className="text-[10px] text-[var(--color-text-faint)]">Publisher: {auth.publisher}</p>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-[var(--color-border)]">
          <button
            onClick={onCancel}
            disabled={isStarting}
            className="px-4 py-2 text-xs font-bold uppercase tracking-wide rounded-lg border border-[var(--color-border)] text-[var(--color-text-dim)] hover:bg-[var(--color-panel-2)] transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={isStarting}
            className="flex items-center gap-2 px-5 py-2 text-xs font-black uppercase tracking-wide rounded-lg bg-warning/20 text-warning border border-warning/30 hover:bg-warning/30 transition-colors disabled:opacity-50"
          >
            {isStarting ? (
              <><Loader2 size={14} className="animate-spin" /> Starting...</>
            ) : (
              <><ShieldAlert size={14} /> Allow &amp; Elevate</>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

// ── LHM Status Panel ────────────────────────────────────────────────────────

function LHMStatusPanel() {
  const { call } = useBackend()
  const queryClient = useQueryClient()
  const [showElevationDialog, setShowElevationDialog] = useState(false)

  const { data: lhmStatus } = useQuery<LHMStatusResult>({
    queryKey: ['lhm-status'],
    queryFn: async () => { const r = await call('SysOps.GetLHMStatus'); return r as LHMStatusResult },
    refetchInterval: 5000,
  })

  const { data: authData } = useQuery<LHMAuthorization>({
    queryKey: ['lhm-authorization'],
    queryFn: async () => { const r = await call('SysOps.GetLHMAuthorization'); return r as LHMAuthorization },
    enabled: showElevationDialog,
  })

  const downloadMutation = useMutation({
    mutationFn: async () => {
      const r = await call('SysOps.DownloadLHM')
      return r as LHMStatusResult
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lhm-status'] })
    },
  })

  const startMutation = useMutation({
    mutationFn: async () => {
      const r = await call('SysOps.StartLHM')
      return r as LHMStatusResult
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lhm-status'] })
      setShowElevationDialog(false)
    },
  })

  const stopMutation = useMutation({
    mutationFn: async () => {
      const r = await call('SysOps.StopLHM')
      return r as LHMStatusResult
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lhm-status'] })
    },
  })

  const handleEnableClick = useCallback(async () => {
    try {
      // If not downloaded yet, download first
      if (lhmStatus && !lhmStatus.available) {
        const result = await downloadMutation.mutateAsync()
        // Check if download actually succeeded (binary now available)
        if (!result.available) {
          // Download reported failure — don't proceed to elevation
          return
        }
      }
      // Show the elevation dialog so the user sees exactly what admin is for
      setShowElevationDialog(true)
    } catch {
      // Error is surfaced via mutation state / lhmStatus.error
    }
  }, [lhmStatus, downloadMutation])

  const handleConfirmElevation = useCallback(() => {
    startMutation.mutate()
  }, [startMutation])

  if (!lhmStatus) return null

  const statusColor = lhmStatus.running
    ? 'text-success'
    : lhmStatus.available
      ? 'text-warning'
      : 'text-[var(--color-text-faint)]'
  const statusLabel = lhmStatus.running
    ? 'Active'
    : lhmStatus.available
      ? 'Installed'
      : 'Not Installed'

  return (
    <>
      <Panel variant="elevated" category="system" padding="md">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={cn(
              "flex items-center justify-center w-9 h-9 rounded-xl border",
              lhmStatus.running
                ? "bg-success/10 border-success/20"
                : lhmStatus.available
                  ? "bg-warning/10 border-warning/20"
                  : "bg-[var(--color-panel-2)] border-[var(--color-border)]"
            )}>
              {lhmStatus.running
                ? <Monitor size={18} className="text-success" />
                : <Monitor size={18} className="text-[var(--color-text-dim)]" />
              }
            </div>
            <div>
              <p className="text-xs font-black text-[var(--color-text)] uppercase tracking-wide">LibreHardwareMonitor</p>
              <div className="flex items-center gap-2 mt-0.5">
                <span className={cn("text-[10px] font-bold", statusColor)}>{statusLabel}</span>
                {lhmStatus.version && (
                  <span className="text-[9px] font-mono text-[var(--color-text-faint)]">v{lhmStatus.version}</span>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {lhmStatus.error && (
              <span className="text-[10px] font-bold text-danger max-w-[200px] truncate" title={lhmStatus.error}>
                {lhmStatus.error}
              </span>
            )}
            {downloadMutation.isError && (
              <span className="text-[10px] font-bold text-danger max-w-[200px] truncate" title={String(downloadMutation.error)}>
                Download failed
              </span>
            )}
            {lhmStatus.running ? (
              <button
                onClick={() => stopMutation.mutate()}
                disabled={stopMutation.isPending}
                className="flex items-center gap-1.5 px-3 py-1.5 text-[10px] font-black uppercase tracking-wide rounded-lg border border-danger/30 text-danger bg-danger/10 hover:bg-danger/20 transition-colors disabled:opacity-50"
              >
                <Square size={12} />
                Stop
              </button>
            ) : (
              <button
                onClick={handleEnableClick}
                disabled={downloadMutation.isPending}
                className="flex items-center gap-1.5 px-3 py-1.5 text-[10px] font-black uppercase tracking-wide rounded-lg border border-success/30 text-success bg-success/10 hover:bg-success/20 transition-colors disabled:opacity-50"
              >
                {downloadMutation.isPending ? (
                  <><Loader2 size={12} className="animate-spin" /> Downloading...</>
                ) : (
                  <><Play size={12} /> Enable Sensors</>
                )}
              </button>
            )}
          </div>
        </div>

        {lhmStatus.running && (
          <p className="mt-3 text-[10px] text-[var(--color-text-faint)] leading-relaxed">
            LHM is running in the background with admin privileges. It provides real-time CPU temperature,
            GPU temperature, fan speeds, and voltage sensors via local WMI. No data leaves this machine.
          </p>
        )}
      </Panel>

      {/* Admin Elevation Dialog */}
      {showElevationDialog && authData && (
        <LHMElevationDialog
          auth={authData}
          onConfirm={handleConfirmElevation}
          onCancel={() => setShowElevationDialog(false)}
          isStarting={startMutation.isPending}
        />
      )}
    </>
  )
}

// ── Main Hardware Tab ────────────────────────────────────────────────────────

export function HardwareTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: hw, isLoading, isError, error, refetch } = useQuery<HardwareInfo>({
    queryKey: ['sysops-hardware-info'],
    queryFn: async () => { const r = await call('SysOps.GetHardwareInfo'); return r as HardwareInfo },
    refetchInterval: refreshInterval,
    retry: 2,
    retryDelay: 1000,
  })

  if (isLoading || !hw) {
    if (isError) {
      return (
        <div className="flex flex-col items-center justify-center h-64 text-[var(--color-text-faint)]">
          <AlertTriangle className="mb-4 opacity-40 text-[var(--color-danger)]" size={48} />
          <p className="text-xs font-black uppercase tracking-widest mb-2 text-[var(--color-danger)]">Telemetry Collection Failed</p>
          <p className="text-[10px] text-[var(--color-text-dim)] mb-4 max-w-sm text-center">
            {error instanceof Error ? error.message : 'Hardware telemetry query failed. This may indicate a system-level issue.'}
          </p>
          <button
            onClick={() => refetch()}
            className="px-4 py-2 text-xs font-bold uppercase tracking-wider rounded-lg border border-[var(--color-border)] hover:border-[var(--color-accent)] transition-colors"
          >
            Retry
          </button>
        </div>
      )
    }
    return (
      <div className="flex flex-col items-center justify-center h-64 text-[var(--color-text-faint)] animate-pulse">
        <Database className="mb-4 opacity-20" size={48} />
        <p className="text-xs font-black uppercase tracking-widest">Collecting Workstation Telemetry...</p>
      </div>
    )
  }

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">

      {/* LHM Status Panel */}
      <LHMStatusPanel />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {/* GPU Information */}
        <Panel variant="elevated" category="system" padding="lg">
          <PanelHeader icon={<Monitor size={20} />} title="Graphics Matrix" subtitle="GPU Hardware & Drivers" />
          {hw.gpu.detected ? (
            <div className="mt-6 space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-bold text-[var(--color-text)]">{hw.gpu.name}</span>
                <span className="text-[10px] font-black text-accent border border-accent/30 rounded px-2 py-0.5 uppercase">{hw.gpu.vendor}</span>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <StatCard label="Video RAM" value={`${hw.gpu.memory_gb.toFixed(1)} GB`} />
                {hw.gpu.utilization > 0 && (
                  <StatCard label="Utilization" value={`${hw.gpu.utilization.toFixed(0)}%`} valueClassName="text-[var(--color-accent)]" />
                )}
                {hw.gpu.temperature > 0 && (
                  <StatCard label="Temperature" value={`${hw.gpu.temperature.toFixed(0)}°C`} valueClassName={cn(hw.gpu.temperature > 80 ? "text-danger" : "text-[var(--color-text)]")} />
                )}
                {hw.gpu.fan_speed > 0 && (
                  <StatCard label="Fan Speed" value={`${hw.gpu.fan_speed.toFixed(0)} RPM`} />
                )}
                {(hw.gpu.utilization === 0 && hw.gpu.temperature === 0 && hw.gpu.fan_speed === 0) && (
                  <div className="col-span-2 text-[10px] text-[var(--color-text-faint)] italic py-2">
                    Real-time metrics unavailable — enable LHM above or install NVIDIA drivers for GPU monitoring.
                  </div>
                )}
              </div>
              <div className="pt-4 border-t border-[var(--color-border)]/50">
                <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-1">Driver Version</p>
                <p className="text-xs font-mono text-[var(--color-text-dim)] truncate">{hw.gpu.driver}</p>
              </div>
            </div>
          ) : (
            <div className="mt-8 text-center py-6 border-2 border-dashed border-[var(--color-border)] rounded-2xl opacity-40">
              <p className="text-xs font-bold text-[var(--color-text-dim)] uppercase">No Discrete GPU Detected</p>
            </div>
          )}
        </Panel>

        {/* Thermal & Power */}
        <Panel variant="elevated" category="system" padding="lg">
          <PanelHeader icon={<Thermometer size={20} />} title="Thermal Envelope" subtitle="Sensors & Power Management" />
          <div className="mt-6 space-y-6">
            <div className="grid grid-cols-2 gap-6">
              <div className="flex flex-col gap-1">
                <span className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest">CPU Temperature</span>
                <div className="flex items-end gap-2">
                  <span className={cn(
                    "text-3xl font-black tabular-nums",
                    hw.cpu.temperature > 80 ? "text-danger" : hw.cpu.temperature > 65 ? "text-warning" : "text-accent"
                  )}>
                    {hw.cpu.temperature > 0 ? `${hw.cpu.temperature.toFixed(0)}°C` : '--'}
                  </span>
                  {hw.cpu.temperature > 0 && <span className="text-[10px] font-bold text-[var(--color-text-dim)] mb-1 uppercase">Package</span>}
                </div>
              </div>

              {hw.battery.detected && (
                <div className="flex flex-col gap-1">
                  <span className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest">Battery Status</span>
                  <div className="flex items-end gap-2">
                    <span className="text-3xl font-black text-[var(--color-success)] tabular-nums">{hw.battery.percent}%</span>
                    <span className="text-[10px] font-bold text-[var(--color-text-dim)] mb-1 uppercase">
                      {hw.battery.charging ? 'Charging' : 'Discharging'}
                    </span>
                  </div>
                </div>
              )}
            </div>

            <div className="h-px bg-[var(--color-border)]/50" />

            <div className="space-y-4">
              <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest">Active Hardware Sensors</p>
              {hw.sensors && hw.sensors.length > 0 ? (
                <div className="grid grid-cols-2 gap-4">
                  {hw.sensors.map((s: SensorData, i: number) => (
                    <div key={i} className="flex items-center justify-between p-3 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)]">
                      <div className="flex items-center gap-2">
                        {s.type === 'Fan' ? <Wind size={14} className="text-accent" /> : <Thermometer size={14} className="text-accent" />}
                        <span className="text-xs font-bold text-[var(--color-text)] truncate w-24">{s.name}</span>
                      </div>
                      <span className="text-xs font-black tabular-nums">{s.value > 0 ? `${s.value}${s.unit}` : 'Active'}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="space-y-3">
                  <p className="text-[10px] text-[var(--color-text-dim)] italic">
                    No sensor data available. Enable LHM above to read CPU temperature, GPU temperature, fan speeds, and voltages.
                  </p>
                </div>
              )}
            </div>
          </div>
        </Panel>
      </div>

      {/* Motherboard / Chassis */}
      <Panel variant="default" category="none" padding="lg">
        <PanelHeader icon={<Database size={20} />} title="Motherboard & Baseboard" />
        <div className="mt-6 grid grid-cols-2 md:grid-cols-4 gap-8">
          <InfoItem label="Manufacturer" value={hw.baseboard.manufacturer} />
          <InfoItem label="Product" value={hw.baseboard.product} />
          <InfoItem label="Version" value={hw.baseboard.version} />
          <InfoItem label="Serial Number" value={hw.baseboard.serial_number} />
        </div>
      </Panel>
    </div>
  )
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-[10px] font-black text-[var(--color-text-faint)] uppercase tracking-widest mb-1">{label}</p>
      <p className="text-sm font-bold text-[var(--color-text)] truncate">{value || 'Unknown'}</p>
    </div>
  )
}
