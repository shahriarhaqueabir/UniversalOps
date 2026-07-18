import { useState } from 'react'
import { ShieldOff, Skull, Ban, Lock, Camera, Download } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { Panel } from '@/components/ui/Panel'
import { ConfirmationModal } from '@/components/dialogs/ConfirmationModal'
import type { SecActionResult, ActionPreview } from '@/types'

export function ResponseTab() {
  const { call } = useBackend()
  const [lastResult, setLastResult] = useState<SecActionResult | null>(null)
  const [preview, setPreview] = useState<ActionPreview | null>(null)
  const [blockIP, setBlockIP] = useState('')
  const [killPID, setKillPID] = useState('')
  const [disableUser, setDisableUser] = useState('')

  const requestAction = async (action: string) => {
    let p: ActionPreview
    switch (action) {
      case 'isolate':
        p = await call('SecOps.IsolateHost', true, 3600) as ActionPreview
        break
      case 'capture':
        p = await call('SecOps.CaptureEvidence') as ActionPreview
        break
      case 'export':
        p = await call('SecOps.ExportForensicBundle') as ActionPreview
        break
      case 'blockip':
        p = await call('SecOps.BlockIP', blockIP) as ActionPreview
        break
      case 'kill':
        p = await call('SecOps.KillProcess', parseInt(killPID)) as ActionPreview
        break
      case 'disable':
        p = await call('SecOps.DisableAccount', disableUser) as ActionPreview
        break
      default:
        return
    }
    setPreview(p)
  }

  const handleConfirm = async () => {
    if (!preview) return
    const handshakeID = preview.handshake_id
    setPreview(null)
    const result = await call('App.ConfirmAction', handshakeID) as SecActionResult
    setLastResult(result)
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Incident Response"
        objective="Take immediate action during a security incident. These actions are destructive or disruptive — use with caution."
        checklist={['Isolate host from network', 'Block malicious IP addresses', 'Kill suspicious processes', 'Disable compromised accounts', 'Capture forensic evidence']}
      />

      <ConfirmationModal
        preview={preview}
        onConfirm={handleConfirm}
        onCancel={() => setPreview(null)}
      />

      {/* Quick Actions */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <button
          onClick={() => requestAction('isolate')}
          className="bg-panel border border-danger/30 rounded-[var(--radius-lg)] p-8 shadow-xl text-left hover:bg-danger/5 transition-all group active:scale-95"
        >
          <ShieldOff size={32} className="text-danger mb-4 group-hover:scale-110 transition-transform" />
          <h3 className="text-lg font-bold text-text mb-2">Isolate Host</h3>
          <p className="text-sm text-text-dim">Block all inbound network traffic immediately.</p>
        </button>

        <button
          onClick={() => requestAction('capture')}
          className="bg-panel border border-accent/30 rounded-[var(--radius-lg)] p-8 shadow-xl text-left hover:bg-accent/5 transition-all group active:scale-95"
        >
          <Camera size={32} className="text-accent mb-4 group-hover:scale-110 transition-transform" />
          <h3 className="text-lg font-bold text-text mb-2">Capture Evidence</h3>
          <p className="text-sm text-text-dim">Collect running processes and listening ports.</p>
        </button>

        <button
          onClick={() => requestAction('export')}
          className="bg-panel border border-accent/30 rounded-[var(--radius-lg)] p-8 shadow-xl text-left hover:bg-accent/5 transition-all group active:scale-95"
        >
          <Download size={32} className="text-accent mb-4 group-hover:scale-110 transition-transform" />
          <h3 className="text-lg font-bold text-text mb-2">Export Bundle</h3>
          <p className="text-sm text-text-dim">Export forensic evidence to file.</p>
        </button>
      </div>

      {/* Targeted Actions */}
      <Panel padding="lg" category="security">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">Targeted Actions</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {/* Block IP */}
          <div className="bg-panel-2 border border-border rounded-xl p-6">
            <Ban size={24} className="text-danger mb-4" />
            <h4 className="text-sm font-bold text-text mb-3">Block IP Address</h4>
            <input
              type="text"
              placeholder="e.g. 192.168.1.100"
              value={blockIP}
              onChange={e => setBlockIP(e.target.value)}
              className="w-full bg-panel border border-border rounded-lg px-3 py-2 text-sm text-text placeholder-text-faint mb-3 focus:outline-none focus:border-accent"
            />
            <button
              onClick={() => requestAction('blockip')}
              disabled={!blockIP}
              className="w-full px-4 py-2 rounded-lg bg-danger text-white text-sm font-bold uppercase tracking-wider disabled:opacity-50 transition-all active:scale-95"
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
              className="w-full bg-panel border border-border rounded-lg px-3 py-2 text-sm text-text placeholder-text-faint mb-3 focus:outline-none focus:border-accent"
            />
            <button
              onClick={() => requestAction('kill')}
              disabled={!killPID}
              className="w-full px-4 py-2 rounded-lg bg-warning text-black text-sm font-bold uppercase tracking-wider disabled:opacity-50 transition-all active:scale-95"
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
              className="w-full bg-panel border border-border rounded-lg px-3 py-2 text-sm text-text placeholder-text-faint mb-3 focus:outline-none focus:border-accent"
            />
            <button
              onClick={() => requestAction('disable')}
              disabled={!disableUser}
              className="w-full px-4 py-2 rounded-lg bg-danger text-white text-sm font-bold uppercase tracking-wider disabled:opacity-50 transition-all active:scale-95"
            >
              Disable Account
            </button>
          </div>
        </div>
      </Panel>

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
