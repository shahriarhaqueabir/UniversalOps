import { useState } from 'react'
import { ShieldOff, Skull, Ban, Lock, Camera, Download } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { SectionBriefing } from './components'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import type { SecActionResult } from '@/types'

export function ResponseTab() {
  const { call } = useBackend()
  const [lastResult, setLastResult] = useState<SecActionResult | null>(null)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingAction, setPendingAction] = useState<string>('')
  const [blockIP, setBlockIP] = useState('')
  const [killPID, setKillPID] = useState('')
  const [disableUser, setDisableUser] = useState('')

  const executeAction = async (action: string) => {
    let result: SecActionResult
    switch (action) {
      case 'isolate':
        result = await call('SecOps.IsolateHost') as SecActionResult
        break
      case 'capture':
        result = await call('SecOps.CaptureEvidence') as SecActionResult
        break
      case 'export':
        result = await call('SecOps.ExportForensicBundle') as SecActionResult
        break
      case 'blockip':
        result = await call('SecOps.BlockIP', blockIP) as SecActionResult
        break
      case 'kill':
        result = await call('SecOps.KillProcess', parseInt(killPID)) as SecActionResult
        break
      case 'disable':
        result = await call('SecOps.DisableAccount', disableUser) as SecActionResult
        break
      default:
        return
    }
    setLastResult(result)
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Incident Response"
        objective="Take immediate action during a security incident. These actions are destructive or disruptive — use with caution."
        checklist={['Isolate host from network', 'Block malicious IP addresses', 'Kill suspicious processes', 'Disable compromised accounts', 'Capture forensic evidence']}
      />

      <ConfirmDialog
        open={confirmOpen}
        title="Confirm Destructive Action"
        description={`You are about to execute "${pendingAction}". This action may disrupt network connectivity or terminate processes. Proceed with caution.`}
        type="warning"
        confirmText="Execute"
        onConfirm={() => {
          setConfirmOpen(false)
          executeAction(pendingAction)
        }}
        onClose={() => setConfirmOpen(false)}
      />

      {/* Quick Actions */}
      <div className="grid grid-cols-3 gap-6">
        <button
          onClick={() => { setPendingAction('isolate'); setConfirmOpen(true) }}
          className="bg-panel border border-danger/30 rounded-[var(--radius-lg)] p-8 shadow-xl text-left hover:bg-danger/5 transition-all group"
        >
          <ShieldOff size={32} className="text-danger mb-4 group-hover:scale-110 transition-transform" />
          <h3 className="text-lg font-bold text-text mb-2">Isolate Host</h3>
          <p className="text-sm text-text-dim">Block all inbound network traffic immediately.</p>
        </button>

        <button
          onClick={() => { setPendingAction('capture'); setConfirmOpen(true) }}
          className="bg-panel border border-accent/30 rounded-[var(--radius-lg)] p-8 shadow-xl text-left hover:bg-accent/5 transition-all group"
        >
          <Camera size={32} className="text-accent mb-4 group-hover:scale-110 transition-transform" />
          <h3 className="text-lg font-bold text-text mb-2">Capture Evidence</h3>
          <p className="text-sm text-text-dim">Collect running processes and listening ports.</p>
        </button>

        <button
          onClick={() => { setPendingAction('export'); setConfirmOpen(true) }}
          className="bg-panel border border-accent/30 rounded-[var(--radius-lg)] p-8 shadow-xl text-left hover:bg-accent/5 transition-all group"
        >
          <Download size={32} className="text-accent mb-4 group-hover:scale-110 transition-transform" />
          <h3 className="text-lg font-bold text-text mb-2">Export Bundle</h3>
          <p className="text-sm text-text-dim">Export forensic evidence to file.</p>
        </button>
      </div>

      {/* Targeted Actions */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-xl">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">Targeted Actions</h3>
        <div className="grid grid-cols-3 gap-6">
          {/* Block IP */}
          <div className="bg-panel-2 border border-border rounded-xl p-6">
            <Ban size={24} className="text-danger mb-4" />
            <h4 className="text-sm font-bold text-text mb-3">Block IP Address</h4>
            <input
              type="text"
              placeholder="e.g. 192.168.1.100"
              value={blockIP}
              onChange={e => setBlockIP(e.target.value)}
              className="w-full bg-panel border border-border rounded-lg px-3 py-2 text-sm text-text placeholder-text-faint mb-3"
            />
            <button
              onClick={() => { setPendingAction('blockip'); setConfirmOpen(true) }}
              disabled={!blockIP}
              className="w-full px-4 py-2 rounded-lg bg-danger text-white text-sm font-bold uppercase tracking-wider disabled:opacity-50"
            >
              Block IP
            </button>
          </div>

          {/* Kill Process */}
          <div className="bg-panel-2 border border-border rounded-xl p-6">
            <Skull size={24} className="text-warning mb-4" />
            <h4 className="text-sm font-bold text-text mb-3">Kill Process</h4>
            <input
              type="number"
              placeholder="PID"
              value={killPID}
              onChange={e => setKillPID(e.target.value)}
              className="w-full bg-panel border border-border rounded-lg px-3 py-2 text-sm text-text placeholder-text-faint mb-3"
            />
            <button
              onClick={() => { setPendingAction('kill'); setConfirmOpen(true) }}
              disabled={!killPID}
              className="w-full px-4 py-2 rounded-lg bg-warning text-black text-sm font-bold uppercase tracking-wider disabled:opacity-50"
            >
              Kill Process
            </button>
          </div>

          {/* Disable Account */}
          <div className="bg-panel-2 border border-border rounded-xl p-6">
            <Lock size={24} className="text-danger mb-4" />
            <h4 className="text-sm font-bold text-text mb-3">Disable Account</h4>
            <input
              type="text"
              placeholder="Username"
              value={disableUser}
              onChange={e => setDisableUser(e.target.value)}
              className="w-full bg-panel border border-border rounded-lg px-3 py-2 text-sm text-text placeholder-text-faint mb-3"
            />
            <button
              onClick={() => { setPendingAction('disable'); setConfirmOpen(true) }}
              disabled={!disableUser}
              className="w-full px-4 py-2 rounded-lg bg-danger text-white text-sm font-bold uppercase tracking-wider disabled:opacity-50"
            >
              Disable Account
            </button>
          </div>
        </div>
      </div>

      {/* Result */}
      {lastResult && (
        <div className={`bg-panel border rounded-[var(--radius-lg)] p-6 shadow-xl ${lastResult.success ? 'border-success/30' : 'border-danger/30'}`}>
          <div className="flex items-center gap-3">
            {lastResult.success ? (
              <span className="text-success font-bold text-sm">SUCCESS</span>
            ) : (
              <span className="text-danger font-bold text-sm">FAILED</span>
            )}
            <span className="text-text-dim text-sm">{lastResult.message || lastResult.error}</span>
          </div>
        </div>
      )}
    </div>
  )
}
