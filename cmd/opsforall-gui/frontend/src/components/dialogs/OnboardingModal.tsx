import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import {
  CheckCircle2,
  Zap,
  BrainCircuit,
  ShieldCheck,
  ChevronRight,
  RefreshCw,
  Terminal,
  Activity,
  User,
  HardDrive,
  Database,
  Globe,
  Settings,
  Bot,
  Info
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { CapabilityInfo } from '@/types'

interface OnboardingModalProps {
  onComplete: () => void
}

type Step =
  | 'welcome'
  | 'identity'
  | 'sovereignty'
  | 'ecosystem'
  | 'governance'
  | 'capabilities'
  | 'pipeline'
  | 'ai-setup'
  | 'finished'

const STEPS: Step[] = [
  'welcome', 'identity', 'sovereignty', 'ecosystem', 'governance',
  'capabilities', 'pipeline', 'ai-setup', 'finished'
]

export function OnboardingModal({ onComplete }: OnboardingModalProps) {
  const { call } = useBackend()
  const { setCompanionName } = useSettingsStore()

  const [step, setStep] = useState<Step>('welcome')
  const [pendingConfig, setPendingConfig] = useState({
    companionName: 'Hawk',
    storagePath: '',
    modelPath: '',
    safetyPolicy: 'standard', // sentinel | standard | tactical
    engineProfile: 'standard', // eco | standard | burst
    homeSubnet: 'Detecting...',
  })

  const [capabilities, setCapabilities] = useState<CapabilityInfo[]>([])
  const [setupRunning, setSetupRunning] = useState(false)

  // ── Data Fetching ──

  const refreshData = useCallback(async () => {
    try {
      const caps = await call('App.GetSystemCapabilities')
      setCapabilities(caps as CapabilityInfo[])
    } catch (err) {
      console.error('Onboarding refresh failed:', err)
    }
  }, [call])

  useEffect(() => {
    if (['capabilities', 'ai-setup'].includes(step)) {
      refreshData()
    }
  }, [step, refreshData])

  // ── Actions ──

  const handleBrowse = async (field: 'storagePath' | 'modelPath') => {
    try {
      const path = await call('App.OpenFileDialog', 'Select Directory', [])
      if (path) setPendingConfig(prev => ({ ...prev, [field]: path }))
    } catch { toast.error('Failed to select directory') }
  }

  const handleAISetup = async () => {
    setSetupRunning(true)
    try {
      const baseModel = "hf.co/empero-ai/Qwythos-9B-Claude-Mythos-5-1M-GGUF:Q6_K"
      await call('AIOps.PullModel', baseModel)
      await call('AIOps.CreateOpsPersona')
      setStep('finished')
    } catch (err: any) {
      toast.error(err?.message || 'AI setup failed')
    } finally {
      setSetupRunning(false)
    }
  }

  const handleFinalDeploy = async () => {
    try {
      // Apply all pending configs
      setCompanionName(pendingConfig.companionName)
      await call('App.ApplyOperationalProfile', pendingConfig.engineProfile)
      if (pendingConfig.storagePath) {
        await call('App.UpdateStorageConfig', pendingConfig.storagePath)
      }

      await call('App.MarkOnboarded')
      onComplete()
    } catch (err) {
      toast.error('Final deployment failed. Check logs.')
    }
  }

  const nextStep = () => {
    const idx = STEPS.indexOf(step)
    if (idx < STEPS.length - 1) setStep(STEPS[idx + 1])
  }

  const prevStep = () => {
    const idx = STEPS.indexOf(step)
    if (idx > 0) setStep(STEPS[idx - 1])
  }

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center p-6 bg-black/95 backdrop-blur-3xl animate-in fade-in duration-700">
      <div className="bg-panel border border-white/10 rounded-[3rem] shadow-[0_0_150px_rgba(0,0,0,1)] w-full max-w-4xl min-h-[720px] flex flex-col overflow-hidden relative border-t-white/20">

        {/* Animated Background Gradients */}
        <div className="absolute -top-40 -left-40 w-[40rem] h-[40rem] bg-accent/20 rounded-full blur-[160px] animate-pulse pointer-events-none" />
        <div className="absolute -bottom-40 -right-40 w-[40rem] h-[40rem] bg-accent/10 rounded-full blur-[160px] pointer-events-none" />

        {/* Header Rail */}
        <div className="px-12 py-8 border-b border-white/5 bg-panel-2/40 flex items-center justify-between">
          <div className="flex items-center gap-5">
            <div className="w-16 h-16 rounded-[1.5rem] bg-accent flex items-center justify-center shadow-[0_0_40px_rgba(var(--color-accent-rgb),0.4)]">
              <ShieldCheck size={36} className="text-white" />
            </div>
            <div>
              <h2 className="text-2xl font-black text-white tracking-tighter uppercase italic">Grand Initialization</h2>
              <p className="text-accent font-black uppercase tracking-[0.3em] text-[10px] mt-1 opacity-90">Workstation Setup Phase {STEPS.indexOf(step) + 1} of {STEPS.length}</p>
            </div>
          </div>
          <div className="flex gap-1.5">
            {STEPS.map((s, i) => (
              <div
                key={s}
                className={cn(
                  "h-1 rounded-full transition-all duration-700",
                  step === s ? "w-8 bg-accent" :
                  i < STEPS.indexOf(step) ? "w-4 bg-success/60" : "w-4 bg-white/10"
                )}
              />
            ))}
          </div>
        </div>

        {/* Workspace */}
        <div className="flex-1 p-12 overflow-y-auto">
          {step === 'welcome' && (
            <div className="flex flex-col items-center text-center space-y-12 animate-in slide-in-from-bottom-12 duration-1000 ease-out">
              <div className="w-24 h-24 rounded-3xl bg-panel-3 border border-white/10 flex items-center justify-center text-accent shadow-2xl relative group">
                <Zap size={48} className="relative z-10" />
                <div className="absolute inset-0 bg-accent/30 blur-2xl rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-1000" />
              </div>
              <div className="space-y-6">
                <h3 className="text-5xl font-black text-white leading-tight tracking-tight">
                  The future is local.<br />The future is <span className="text-accent italic">yours.</span>
                </h3>
                <p className="text-text-dim text-xl max-w-xl mx-auto leading-relaxed font-medium">
                  Welcome to AllOpsFull. We are about to initialize your professional operations control center.
                </p>
              </div>
              <div className="grid grid-cols-3 gap-6 w-full max-w-3xl pt-6">
                {[
                  { icon: <Globe className="text-blue-400" />, title: 'Zero Telemetry', desc: 'No data leaves this machine.' },
                  { icon: <BrainCircuit className="text-violet-400" />, title: 'Local AI', desc: 'Private workstation intelligence.' },
                  { icon: <Activity className="text-emerald-400" />, title: 'High Density', desc: 'Mechanics-first telemetry.' },
                ].map((item, i) => (
                  <div key={i} className="p-6 rounded-3xl bg-panel-2/50 border border-white/5 space-y-3 text-left hover:bg-panel-2 transition-all group">
                    <div className="w-10 h-10 rounded-xl bg-white/5 flex items-center justify-center group-hover:scale-110 transition-transform">{item.icon}</div>
                    <p className="font-black text-xs uppercase tracking-widest text-white">{item.title}</p>
                    <p className="text-[10px] text-text-faint font-bold leading-relaxed">{item.desc}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {step === 'identity' && (
            <div className="max-w-xl mx-auto space-y-10 animate-in slide-in-from-right-12 duration-700">
              <div className="space-y-3">
                <h3 className="text-3xl font-black text-white uppercase tracking-tight">Co-Pilot Identity</h3>
                <p className="text-text-dim text-lg font-medium leading-relaxed">What should we call your primary AI assistant?</p>
              </div>

              <div className="relative group">
                <input
                  type="text"
                  value={pendingConfig.companionName}
                  onChange={(e) => setPendingConfig(prev => ({ ...prev, companionName: e.target.value }))}
                  className="w-full bg-panel-2 border-2 border-white/10 rounded-[2rem] px-8 py-6 text-4xl font-black text-white focus:outline-none focus:border-accent transition-all shadow-inner"
                  placeholder="Companion Name"
                />
                <User size={32} className="absolute right-8 top-8 text-white/10 group-focus-within:text-accent transition-colors" />
              </div>

              <div className="p-6 rounded-3xl bg-accent/5 border border-accent/10 flex items-start gap-5">
                <div className="w-12 h-12 rounded-2xl bg-accent/10 flex items-center justify-center text-accent shrink-0"><Bot size={24} /></div>
                <p className="text-sm text-text-dim font-medium leading-relaxed italic">
                  "I will use this name for all briefings, reports, and real-time alerts. You can change this later in the Control Plane."
                </p>
              </div>
            </div>
          )}

          {step === 'sovereignty' && (
            <div className="max-w-2xl mx-auto space-y-10 animate-in slide-in-from-right-12 duration-700">
              <div className="space-y-3 text-center">
                <h3 className="text-3xl font-black text-white uppercase tracking-tight">Portable Sovereignty</h3>
                <p className="text-text-dim text-lg font-medium leading-relaxed">AllOpsFull is strictly self-contained. No system registry entries or hidden folders.</p>
              </div>

              <div className="space-y-4">
                <div className="p-6 rounded-3xl bg-panel-2/50 border border-white/5 flex items-center justify-between group transition-all hover:border-white/20">
                  <div className="flex items-center gap-5">
                    <div className="w-14 h-14 rounded-2xl bg-panel-3 flex items-center justify-center text-blue-400 shadow-inner"><Database size={28} /></div>
                    <div>
                      <p className="font-bold text-white text-lg leading-none">Telemetry Core</p>
                      <p className="text-xs text-accent mt-1.5 font-mono">./data/allopsfull.db</p>
                      <p className="text-[10px] text-text-faint mt-1 leading-relaxed">
                        A portable SQLite instance containing all metrics, alerts, and chat history.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="p-6 rounded-3xl bg-panel-2/50 border border-white/5 flex items-center justify-between group transition-all hover:border-white/20">
                  <div className="flex items-center gap-5">
                    <div className="w-14 h-14 rounded-2xl bg-panel-3 flex items-center justify-center text-violet-400 shadow-inner"><HardDrive size={28} /></div>
                    <div>
                      <p className="font-bold text-white text-lg leading-none">AI Identity</p>
                      <p className="text-xs text-accent mt-1.5 font-mono">./data/allopsfull.modelfile</p>
                      <p className="text-[10px] text-text-faint mt-1 leading-relaxed">
                        Your private intelligence configuration, forced into the local data folder.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="p-6 rounded-3xl bg-panel-2/50 border border-white/5 flex items-center justify-between group transition-all hover:border-white/20">
                  <div className="flex items-center gap-5">
                    <div className="w-14 h-14 rounded-2xl bg-panel-3 flex items-center justify-center text-emerald-400 shadow-inner"><ScrollText size={28} /></div>
                    <div>
                      <p className="font-bold text-white text-lg leading-none">Session Logs</p>
                      <p className="text-xs text-accent mt-1.5 font-mono">./logs/opsforall.log</p>
                      <p className="text-[10px] text-text-faint mt-1 leading-relaxed">
                        Human-readable diagnostic logs for the current session.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {step === 'ecosystem' && (
            <div className="max-w-3xl mx-auto space-y-8 animate-in slide-in-from-right-12 duration-700">
              <div className="space-y-3 text-center">
                <h3 className="text-3xl font-black text-white uppercase tracking-tight">Built-in Capabilities</h3>
                <p className="text-text-dim text-lg font-medium leading-relaxed">Everything needed for core operation is already inside the binary.</p>
              </div>

              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-4">
                  <h4 className="text-[10px] font-black text-accent uppercase tracking-[0.2em] px-2">Core Tools (Native Go)</h4>
                  <div className="space-y-2">
                    {[
                      { label: 'System Monitoring', desc: 'CPU, RAM, Disk, Process analytics.' },
                      { label: 'Network Discovery', desc: 'Native ARP & ICMP subnet sweep.' },
                      { label: 'Security Auditing', desc: 'Windows Defender & Firewall status.' },
                      { label: 'Port Scanner', desc: 'Concurrent TCP socket probing.' },
                    ].map((item, i) => (
                      <div key={i} className="p-4 rounded-2xl bg-panel-2/30 border border-white/5 flex gap-4">
                        <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center shrink-0"><CheckCircle2 size={16} className="text-success" /></div>
                        <div>
                          <p className="text-xs font-bold text-white">{item.label}</p>
                          <p className="text-[9px] text-text-faint mt-0.5">{item.desc}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="space-y-4">
                  <h4 className="text-[10px] font-black text-warning uppercase tracking-[0.2em] px-2">Extension Powers (External)</h4>
                  <div className="space-y-2 opacity-80">
                    {[
                      { label: 'Ollama', desc: 'Drives the AI Analyst (Required for AI features).' },
                      { label: 'Docker / K8s', desc: 'Enables container orchestration views.' },
                      { label: 'Nmap / Git', desc: 'Unlocks advanced auditing & repo management.' },
                      { label: 'PowerShell 7', desc: 'Runs high-fidelity diagnostic workflows.' },
                    ].map((item, i) => (
                      <div key={i} className="p-4 rounded-2xl bg-panel-2/30 border border-dashed border-white/10 flex gap-4">
                        <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center shrink-0"><Zap size={16} className="text-warning" /></div>
                        <div>
                          <p className="text-xs font-bold text-white">{item.label}</p>
                          <p className="text-[9px] text-text-faint mt-0.5">{item.desc}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="p-4 rounded-2xl bg-warning/5 border border-warning/10">
                    <p className="text-[9px] text-warning font-medium leading-relaxed italic">
                      Note: You don't need to install these now. The system will alert you only if an action specifically requires them.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {step === 'governance' && (
            <div className="max-w-2xl mx-auto space-y-10 animate-in slide-in-from-right-12 duration-700">
              <div className="space-y-3 text-center">
                <h3 className="text-3xl font-black text-white uppercase tracking-tight">System Governance</h3>
                <p className="text-text-dim text-lg font-medium leading-relaxed">Integrated tools are 100% optional. The system works out-of-the-box.</p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-6 rounded-3xl bg-panel-2/50 border border-white/5 space-y-4">
                  <div className="flex items-center gap-3">
                    <Wrench size={18} className="text-accent" />
                    <p className="text-xs font-black uppercase tracking-widest text-white">Integrated Power</p>
                  </div>
                  <p className="text-[10px] text-text-dim leading-relaxed">
                    AllOpsFull detects tools like <span className="text-accent font-bold">Nmap</span> or <span className="text-accent font-bold">Docker</span> to unlock <span className="italic">extra</span> features. If they are missing, we use native Go code for core diagnostics. No installation is required for normal operation.
                  </p>
                </div>

                <div className="p-6 rounded-3xl bg-panel-2/50 border border-white/5 space-y-4">
                  <div className="flex items-center gap-3">
                    <ShieldCheck size={18} className="text-success" />
                    <p className="text-xs font-black uppercase tracking-widest text-white">Safety Gateway</p>
                  </div>
                  <p className="text-[10px] text-text-dim leading-relaxed">
                    Choose your AI authority. <span className="text-white font-bold">Standard</span> mode requires you to manually click "Confirm" for every AI-suggested remediation.
                  </p>
                  <div className="flex gap-2">
                    {['standard', 'tactical'].map(p => (
                      <button key={p} onClick={() => setPendingConfig(prev => ({ ...prev, safetyPolicy: p }))}
                        className={cn("px-3 py-1.5 rounded-lg text-[9px] font-black uppercase border transition-all",
                        pendingConfig.safetyPolicy === p ? "bg-accent border-accent text-white" : "bg-white/5 border-white/10 text-text-faint")}>{p}</button>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          {step === 'capabilities' && (
            <div className="max-w-4xl mx-auto space-y-10 animate-in slide-in-from-right-12 duration-700">
              <div className="space-y-3 text-center">
                <h3 className="text-3xl font-black text-white uppercase tracking-tight">Command Encyclopedia</h3>
                <p className="text-text-dim text-lg font-medium leading-relaxed">The following methods and workflows are available for local orchestration.</p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-3">
                  <p className="text-[10px] font-black text-accent uppercase tracking-widest px-2">Go Backend Bindings</p>
                  <div className="bg-panel-2/50 border border-white/5 rounded-2xl p-4 font-mono text-[9px] text-text-dim space-y-1.5 overflow-hidden">
                    <p className="text-white font-bold opacity-100">// Domain Logic Facades</p>
                    <p>App.GetSystemCapabilities()</p>
                    <p>SysOps.GetCPUStats()</p>
                    <p>NetOps.Ping(host, count)</p>
                    <p>SecOps.executeBlockIP(ip)</p>
                    <p>AIOps.Chat(session, prompt)</p>
                    <p>Timeline.ExplainEvents(ids)</p>
                  </div>
                </div>

                <div className="space-y-3">
                  <p className="text-[10px] font-black text-warning uppercase tracking-widest px-2">Integrated Workflows</p>
                  <div className="bg-panel-2/50 border border-white/5 rounded-2xl p-4 font-mono text-[9px] text-text-dim space-y-1.5">
                    <p className="text-white font-bold opacity-100">// Handshake Orchestration</p>
                    <p>IR: Kill Process -> Human Auth -> Exec</p>
                    <p>Net: Isolate Host -> Restricted Token</p>
                    <p>AIOps: Anomaly Scan -> Pearson R -> Toast</p>
                    <p>DevOps: Script Runner -> System Sandbox</p>
                  </div>
                </div>
              </div>

              <div className="p-5 rounded-2xl bg-white/5 border border-dashed border-white/10 flex items-start gap-4">
                <Terminal size={16} className="text-text-faint mt-1" />
                <p className="text-[10px] text-text-faint leading-relaxed italic">
                  All commands are executed via the <span className="text-white font-bold">Unified Sandbox Infrastructure</span>, ensuring that even under AI request, no process runs without predefined boundaries (CPU/RAM/Network).
                </p>
              </div>
            </div>
          )}

          {step === 'pipeline' && (
            <div className="max-w-3xl mx-auto space-y-10 animate-in slide-in-from-right-12 duration-700">
              <div className="space-y-3 text-center">
                <h3 className="text-3xl font-black text-white uppercase tracking-tight">Execution Pipeline</h3>
                <p className="text-text-dim text-lg font-medium leading-relaxed">How AllOpsFull processes your workstation data.</p>
              </div>

              <div className="flex items-center justify-between gap-4 relative">
                <div className="absolute top-1/2 left-0 right-0 h-px bg-white/10 -z-10" />
                {[
                  { icon: <Activity />, label: 'Inquiry', desc: 'Collectors poll OS stats.' },
                  { icon: <BrainCircuit />, label: 'Analysis', desc: 'AI evaluates trends.' },
                  { icon: <RefreshCw />, label: 'Handshake', desc: 'You review & authorize.' },
                  { icon: <Zap />, label: 'Action', desc: 'Sandboxed remediation.' },
                ].map((item, i) => (
                  <div key={i} className="flex flex-col items-center text-center space-y-3 bg-panel p-4 rounded-2xl border border-white/5 shadow-2xl">
                    <div className="w-12 h-12 rounded-xl bg-panel-2 flex items-center justify-center text-accent border border-white/10">{item.icon}</div>
                    <div>
                      <p className="text-[10px] font-black uppercase text-white">{item.label}</p>
                      <p className="text-[8px] text-text-faint font-bold w-24 leading-tight mt-1">{item.desc}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {step === 'ai-setup' && (
            <div className="max-w-2xl mx-auto space-y-10 animate-in slide-in-from-right-12 duration-700">
              <div className="space-y-3 text-center">
                <h3 className="text-3xl font-black text-white uppercase tracking-tight">Neural Core Initialization</h3>
                <p className="text-text-dim text-lg font-medium leading-relaxed">Configuring the local intelligence layer.</p>
              </div>

              {setupRunning ? (
                <div className="space-y-10 py-6">
                  <div className="space-y-4">
                    <div className="flex justify-between text-[10px] font-black uppercase tracking-widest text-accent">
                      <span>Initializing Weights...</span>
                      <span>50%</span>
                    </div>
                    <div className="h-2 bg-panel-3 rounded-full overflow-hidden border border-white/5">
                      <div
                        className="h-full bg-accent transition-all duration-500 shadow-[0_0_20px_rgba(var(--color-accent-rgb),0.5)]"
                        style={{ width: `50%` }}
                      />
                    </div>
                  </div>
                  <div className="flex flex-col items-center gap-4">
                    <RefreshCw size={32} className="animate-spin text-accent/50" />
                    <p className="text-[10px] text-text-faint font-black uppercase tracking-widest">Inference engine is assembling weights...</p>
                  </div>
                </div>
              ) : (
                <div className="space-y-6">
                  <div className="p-6 rounded-3xl bg-violet-400/5 border border-violet-400/20 space-y-4">
                    <div className="flex items-center gap-3">
                      <BrainCircuit size={18} className="text-violet-400" />
                      <p className="text-[10px] font-black uppercase tracking-widest text-violet-400">Persona Profile: {pendingConfig.companionName}</p>
                    </div>
                    <div className="bg-black/20 rounded-xl p-4 border border-white/5">
                      <pre className="text-[9px] font-mono text-text-dim leading-relaxed">
                        FROM hf.co/empero-ai/Qwythos-9B-Claude-Mythos-5-1M-GGUF:Q6_K{"\n"}
                        SYSTEM You are the OpsForAll Technical Auditor...{"\n"}
                        PARAMETER temperature 0.1
                      </pre>
                    </div>
                  </div>

                  <button
                    onClick={handleAISetup}
                    className="w-full py-6 rounded-2xl bg-accent text-white font-black uppercase tracking-widest shadow-xl hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-4"
                  >
                    <Zap size={20} /> Initialize Intelligence
                  </button>
                  <button onClick={nextStep} className="w-full text-[10px] font-black uppercase tracking-widest text-text-faint hover:text-white transition-colors">Skip Activation</button>
                </div>
              )}
            </div>
          )}

          {step === 'finished' && (
            <div className="max-w-2xl mx-auto flex flex-col items-center text-center space-y-12 animate-in zoom-in-95 duration-1000">
              <div className="relative">
                <div className="absolute inset-0 bg-success blur-[80px] opacity-20 animate-pulse" />
                <div className="w-40 h-40 rounded-full bg-success flex items-center justify-center text-white relative z-10 shadow-[0_0_60px_rgba(var(--color-success-rgb),0.4)]">
                  <CheckCircle2 size={80} />
                </div>
              </div>

              <div className="space-y-4">
                <h3 className="text-5xl font-black text-white uppercase tracking-tighter">System Locked.</h3>
                <p className="text-text-dim text-xl font-medium max-w-md mx-auto">
                  Workstation successfully initialized. {pendingConfig.companionName} is now standing by.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4 w-full pt-4">
                <div className="p-4 rounded-2xl bg-white/5 border border-white/5 text-left">
                  <p className="text-[10px] font-black text-text-faint uppercase tracking-widest">Sovereignty</p>
                  <p className="text-xs font-bold text-white mt-1">Local Mode Active</p>
                </div>
                <div className="p-4 rounded-2xl bg-white/5 border border-white/5 text-left">
                  <p className="text-[10px] font-black text-text-faint uppercase tracking-widest">Safety Policy</p>
                  <p className="text-xs font-bold text-white mt-1 uppercase">{pendingConfig.safetyPolicy}</p>
                </div>
              </div>

              <button
                onClick={handleFinalDeploy}
                className="w-full py-8 rounded-[2rem] bg-success text-white text-2xl font-black shadow-[0_0_50px_rgba(var(--color-success-rgb),0.3)] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-6"
              >
                ENTER OPS CENTER
                <ChevronRight size={32} />
              </button>
            </div>
          )}
        </div>

        {/* Footer Navigation */}
        {step !== 'finished' && (
          <div className="p-10 bg-panel-2/30 border-t border-white/5 flex justify-between items-center backdrop-blur-3xl">
            <button
              onClick={prevStep}
              className={cn("text-text-faint font-black text-[10px] uppercase tracking-[0.3em] hover:text-white transition-all flex items-center gap-2", step === 'welcome' && "opacity-0 pointer-events-none")}
            >
              <div className="rotate-180"><ChevronRight size={14} /></div> Previous Phase
            </button>

            <button
              onClick={nextStep}
              className="px-12 py-5 bg-accent text-white text-[10px] font-black uppercase tracking-[0.3em] shadow-[0_0_40px_rgba(var(--color-accent-rgb),0.3)] hover:scale-[1.02] transition-all flex items-center gap-4 group active:scale-95"
            >
              Progress Initialization
              <ChevronRight size={18} className="group-hover:translate-x-1.5 transition-transform duration-500" />
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
