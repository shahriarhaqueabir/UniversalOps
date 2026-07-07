import { X, Sun, RefreshCw, Brain } from 'lucide-react'
import { Switch } from '@radix-ui/react-switch'

interface SettingsDialogProps {
  open: boolean
  onClose: () => void
}

export function SettingsDialog({ open, onClose }: SettingsDialogProps) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="bg-card border border-border rounded-xl shadow-2xl w-[480px] max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold text-text">Quick Settings</h2>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-sidebar-hover text-muted hover:text-text transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Settings */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {/* Theme */}
          <div className="flex items-center justify-between py-2">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-sidebar-hover flex items-center justify-center">
                <Sun size={16} className="text-warning" />
              </div>
              <div>
                <p className="text-sm text-text font-medium">Dark Mode</p>
                <p className="text-xs text-muted">Toggle dark/light theme</p>
              </div>
            </div>
            <Switch
              checked={true}
              onCheckedChange={() => { }}
              className="w-10 h-6 bg-[#0f172a] border border-border rounded-full relative data-[state=checked]:bg-primary cursor-pointer"
            >
              <span className="block w-4 h-4 bg-text rounded-full transition-transform translate-x-1 data-[state=checked]:translate-x-5" />
            </Switch>
          </div>

          {/* Auto-refresh */}
          <div className="flex items-center justify-between py-2">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-sidebar-hover flex items-center justify-center">
                <RefreshCw size={16} className="text-primary" />
              </div>
              <div>
                <p className="text-sm text-text font-medium">Auto-refresh</p>
                <p className="text-xs text-muted">Automatically refresh data</p>
              </div>
            </div>
            <Switch
              checked={true}
              onCheckedChange={() => { }}
              className="w-10 h-6 bg-[#0f172a] border border-border rounded-full relative data-[state=checked]:bg-primary cursor-pointer"
            >
              <span className="block w-4 h-4 bg-text rounded-full transition-transform translate-x-1 data-[state=checked]:translate-x-5" />
            </Switch>
          </div>

          {/* AI Auto-report */}
          <div className="flex items-center justify-between py-2">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-sidebar-hover flex items-center justify-center">
                <Brain size={16} className="text-[#a78bfa]" />
              </div>
              <div>
                <p className="text-sm text-text font-medium">AI Auto-report</p>
                <p className="text-xs text-muted">Auto-generate AI reports</p>
              </div>
            </div>
            <Switch
              checked={false}
              onCheckedChange={() => { }}
              className="w-10 h-6 bg-[#0f172a] border border-border rounded-full relative data-[state=checked]:bg-primary cursor-pointer"
            >
              <span className="block w-4 h-4 bg-text rounded-full transition-transform translate-x-1 data-[state=checked]:translate-x-5" />
            </Switch>
          </div>

          <div className="border-t border-border pt-4">
            <p className="text-xs text-muted text-center">
              Extended settings available on the Settings page
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
