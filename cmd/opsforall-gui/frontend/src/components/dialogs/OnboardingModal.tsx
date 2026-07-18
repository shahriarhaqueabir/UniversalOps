import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import {
  Bot,
  CheckCircle2,
  AlertCircle,
  Download,
  Zap,
  BrainCircuit,
  ShieldCheck,
  ChevronRight,
  RefreshCw,
  Terminal,
  Activity
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import type { OllamaStatus, OllamaProgress } from '@/types'

interface OnboardingModalProps {
  onComplete: () => void
}

type Step = 'welcome' | 'dependencies' | 'ai-setup' | 'finished'

export function OnboardingModal({ onComplete }: OnboardingModalProps) {
  const { call } = useBackend()
  const [step, setStep] = useState<Step>('welcome')
  const [ollamaStatus, setOllamaStatus] = useState<OllamaStatus | null>(null)
  const [checking, setChecking] = useState(false)
  const [setupProgress, setSetupProgress] = useState<OllamaProgress | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [setupRunning, setSetupRunning] = useState(false)

  const checkDependencies = useCallback(async () => {
    setChecking(true)
    setError(null)
    try {
      const status = await call('AIOps.GetOllamaStatus') as OllamaStatus
      setOllamaStatus(status)
    } catch {
      setError('Failed to check system dependencies. Please ensure the backend is reachable.')
    } finally {
      setChecking(false)
    }
  }, [call])

  useEffect(() => {
    if (step === 'dependencies') {
      checkDependencies()
    }
  }, [step, checkDependencies])

  // Auto-retry dependency check when user returns to app
  useEffect(() => {
    const handleFocus = () => {
      if (step === 'dependencies') {
        checkDependencies()
      }
    }
    window.addEventListener('focus', handleFocus)
    return () => window.removeEventListener('focus', handleFocus)
  }, [step, checkDependencies])

  // Listen for progress events
  useEffect(() => {
    const runtime = (window as any).runtime
    if (runtime?.EventsOn) {
      const handleProgress = (data: OllamaProgress) => {
        setSetupProgress(data)
      }
      runtime.EventsOn('ollama:progress', handleProgress)
      return () => runtime.EventsOff('ollama:progress', handleProgress)
    }
  }, [])

  const handleAISetup = async () => {
    setSetupRunning(true)
    setError(null)
    try {
      // 1. Pull base model
      const baseModel = "hf.co/empero-ai/Qwythos-9B-Claude-Mythos-5-1M-GGUF:Q6_K"
      setSetupProgress({ status: 'Initiating pull...', percent: 0, total: 0, completed: 0 })
      await call('AIOps.PullModel', baseModel)

      // 2. Create specialized persona
      setSetupProgress({ status: 'Creating specialized persona...', percent: 100, total: 0, completed: 0 })
      await call('AIOps.CreateOpsPersona')

      toast.success('AI Core initialized successfully')
      setStep('finished')
    } catch (err: any) {
      const msg = err?.message || 'AI setup failed'
      setError(msg)
      toast.error(msg, {
        description: 'You can skip this for now and configure it later in AIOps.'
      })
    } finally {
      setSetupRunning(false)
    }
  }

  const handleFinish = async () => {
    await call('App.MarkOnboarded')
    onComplete()
  }

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center p-6 bg-black/90 backdrop-blur-2xl animate-in fade-in duration-700">
      <div className="bg-panel border border-white/10 rounded-[2.5rem] shadow-[0_0_120px_rgba(0,0,0,1)] w-full max-w-3xl min-h-[680px] flex flex-col overflow-hidden relative border-t-white/20">
        {/* Background Atmosphere */}
        <div className="absolute -top-32 -left-32 w-[32rem] h-[32rem] bg-accent/15 rounded-full blur-[128px] pointer-events-none" />
        <div className="absolute -bottom-32 -right-32 w-[32rem] h-[32rem] bg-accent/5 rounded-full blur-[128px] pointer-events-none" />

        <div className="relative z-10 flex flex-col h-full grow">
          {/* Header */}
          <div className="px-10 py-8 border-b border-white/5 bg-panel-2/30 flex items-center justify-between">
            <div className="flex items-center gap-4 group">
              <div className="w-14 h-14 rounded-2xl bg-accent flex items-center justify-center shadow-[0_0_30px_rgba(var(--color-accent-rgb),0.3)] transition-all duration-700 group-hover:scale-105">
                <ShieldCheck size={32} className="text-white" />
              </div>
              <div>
                <h2 className="text-2xl font-black text-white tracking-tight italic">OpsForAll</h2>
                <p className="text-accent font-bold uppercase tracking-[0.25em] text-[10px] mt-1 opacity-80">Universal Command Center</p>
              </div>
            </div>
            <div className="flex gap-2">
              {(['welcome', 'dependencies', 'ai-setup', 'finished'] as Step[]).map((s, i) => (
                <div
                  key={s}
                  className={cn(
                    "w-2.5 h-2.5 rounded-full border-2 transition-all duration-500",
                    step === s ? "bg-accent border-accent scale-125" :
                    i < ['welcome', 'dependencies', 'ai-setup', 'finished'].indexOf(step) ? "bg-success border-success" : "border-white/10"
                  )}
                />
              ))}
            </div>
          </div>

          {/* Body */}
          <div className="flex-1 p-10 flex flex-col justify-center overflow-y-auto">
            {step === 'welcome' && (
              <div className="flex flex-col items-center text-center space-y-10 animate-in slide-in-from-bottom-8 duration-700 ease-out">
                <div className="w-20 h-20 rounded-[1.75rem] bg-panel-3 border border-white/10 flex items-center justify-center text-accent shadow-inner relative group">
                   <div className="absolute inset-0 bg-accent/20 blur-xl rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-700" />
                  <Zap size={36} className="relative z-10" />
                </div>
                <div className="space-y-4">
                  <h3 className="text-4xl font-black text-white leading-[1.1] tracking-tight max-w-lg">
                    Welcome to the<br />future of Operations.
                  </h3>
                  <p className="text-text-dim text-lg max-w-md mx-auto leading-relaxed font-medium">
                    Enterprise-grade telemetry meets local AI intelligence at your workstation.
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-4 w-full max-w-2xl pt-4">
                  {[
                    { icon: <Activity className="text-success" size={20} />, label: 'Real-time Metrics' },
                    { icon: <BrainCircuit className="text-accent" size={20} />, label: 'Local AI Analyst' },
                    { icon: <ShieldCheck className="text-warning" size={20} />, label: 'Security Auditing' },
                    { icon: <Terminal className="text-white" size={20} />, label: 'DevOps Automation' },
                  ].map((item, i) => (
                    <div key={i} className="flex items-center gap-4 p-4.5 rounded-2xl bg-panel-2/50 border border-white/5 hover:border-white/10 transition-colors group">
                      <div className="shrink-0 group-hover:scale-110 transition-transform">{item.icon}</div>
                      <span className="font-bold text-sm text-text-dim group-hover:text-white transition-colors">{item.label}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {step === 'dependencies' && (
              <div className="space-y-8 animate-in slide-in-from-right-8 duration-700 ease-out max-w-xl mx-auto w-full">
                <div className="space-y-2">
                  <h3 className="text-3xl font-bold text-white tracking-tight">System Pre-flight</h3>
                  <p className="text-text-dim text-lg font-medium">Checking machine readiness for advanced operations.</p>
                </div>

                <div className="space-y-4">
                  <div className="p-6 rounded-2xl bg-panel-2/50 border border-white/5 flex items-center justify-between group hover:bg-panel-2 transition-colors">
                    <div className="flex items-center gap-5">
                      <div className="w-14 h-14 rounded-xl bg-panel-3 border border-white/5 flex items-center justify-center text-accent group-hover:scale-105 transition-transform shadow-inner">
                        <Bot size={28} />
                      </div>
                      <div>
                        <p className="font-bold text-white text-xl">Ollama AI Runtime</p>
                        <p className="text-sm text-text-faint font-medium">Local intelligence engine requirement</p>
                      </div>
                    </div>
                    {checking ? (
                      <RefreshCw size={24} className="animate-spin text-accent" />
                    ) : ollamaStatus?.binary_exists ? (
                      <div className="w-10 h-10 rounded-full bg-success/10 flex items-center justify-center text-success border border-success/20">
                        <CheckCircle2 size={24} />
                      </div>
                    ) : (
                      <div className="flex flex-col items-end gap-1.5">
                        <div className="flex items-center gap-2 text-danger font-black text-xs uppercase tracking-widest">
                          <AlertCircle size={16} />
                          Missing
                        </div>
                        <a href="https://ollama.com" target="_blank" rel="noreferrer" className="text-xs text-accent font-bold hover:underline transition-all hover:opacity-80">Download Runtime</a>
                      </div>
                    )}
                  </div>
                </div>

                {!checking && !ollamaStatus?.binary_exists && (
                  <div className="p-6 rounded-2xl bg-danger/5 border border-danger/20 flex items-start gap-4 animate-in shake-in duration-500">
                    <AlertCircle size={24} className="text-danger mt-1 shrink-0" />
                    <div className="space-y-1">
                      <p className="font-bold text-danger">Ollama not detected</p>
                      <p className="text-sm text-danger/80 leading-relaxed font-medium">You'll need to install Ollama to use AI features. You can continue, but the Analyst core will be inactive.</p>
                    </div>
                  </div>
                )}
              </div>
            )}

            {step === 'ai-setup' && (
              <div className="space-y-10 animate-in slide-in-from-right-8 duration-700 ease-out max-w-xl mx-auto w-full">
                <div className="space-y-2">
                  <h3 className="text-3xl font-bold text-white tracking-tight">Neural Core Initialization</h3>
                  <p className="text-text-dim text-lg font-medium">Configuring your local operations specialist.</p>
                </div>

                {setupRunning ? (
                  <div className="space-y-12 py-6">
                    <div className="space-y-6">
                      <div className="flex justify-between text-xs font-black uppercase tracking-[0.2em] text-accent">
                        <span>{setupProgress?.status || 'Processing...'}</span>
                        <span className="tabular-nums">{Math.round(setupProgress?.percent || 0)}%</span>
                      </div>
                      <div className="h-2.5 bg-panel-3 rounded-full overflow-hidden border border-white/5 shadow-inner">
                        <div
                          className="h-full bg-accent transition-all duration-700 shadow-[0_0_20px_rgba(var(--color-accent-rgb),0.5)]"
                          style={{ width: `${setupProgress?.percent || 0}%` }}
                        />
                      </div>
                    </div>
                    <div className="flex flex-col items-center text-center space-y-6">
                      <div className="w-16 h-16 rounded-full bg-accent/10 flex items-center justify-center">
                        <RefreshCw size={32} className="animate-spin text-accent" />
                      </div>
                      <p className="text-text-faint text-sm italic font-medium max-w-xs mx-auto">
                        Fetching neural weights. This may take a few minutes depending on your connection.
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="flex flex-col items-center text-center space-y-8 py-4">
                    <div className="w-24 h-24 rounded-full bg-accent/5 flex items-center justify-center text-accent border border-accent/10 shadow-inner group">
                      <Download size={48} className="group-hover:translate-y-1 transition-transform" />
                    </div>
                    <p className="text-text-dim text-lg leading-relaxed font-medium">
                      Downloading the base intelligence core and configuring the specialized <strong className="text-white">OpsForAll</strong> persona (approx. 5GB).
                    </p>

                    {!ollamaStatus?.binary_exists ? (
                      <div className="w-full space-y-5">
                        <div className="p-5 rounded-2xl bg-danger/5 border border-danger/10 text-danger text-sm font-bold flex items-center justify-center gap-3">
                          <AlertCircle size={20} />
                          Ollama runtime is required to proceed.
                        </div>
                        <button
                          disabled
                          className="w-full py-5 rounded-2xl bg-white/5 text-white/20 text-xl font-black border border-white/5 cursor-not-allowed transition-all flex items-center justify-center gap-3"
                        >
                          <Bot size={24} />
                          Core Inactive
                        </button>
                      </div>
                    ) : (
                      <button
                        onClick={handleAISetup}
                        className="w-full py-6 rounded-2xl bg-accent text-white text-xl font-black shadow-[0_0_40px_rgba(var(--color-accent-rgb),0.3)] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-4 group"
                      >
                        <Zap size={24} className="group-hover:scale-110 transition-transform" />
                        Initialize Neural Core
                      </button>
                    )}

                    <button onClick={() => setStep('finished')} className="text-text-faint font-bold text-sm uppercase tracking-widest hover:text-white transition-all hover:opacity-80">Skip Activation</button>
                  </div>
                )}

                {error && (
                  <div className="p-4 rounded-xl bg-danger/10 border border-danger/20 text-danger text-xs font-black uppercase tracking-widest flex items-center justify-center gap-3 animate-in fade-in duration-300">
                    <AlertCircle size={18} />
                    {error}
                  </div>
                )}
              </div>
            )}

            {step === 'finished' && (
              <div className="flex flex-col items-center text-center space-y-12 animate-in zoom-in-95 duration-700 ease-out">
                <div className="relative">
                  <div className="absolute inset-0 bg-success blur-[64px] opacity-20 animate-pulse" />
                  <div className="w-36 h-36 rounded-full bg-success flex items-center justify-center text-white relative z-10 shadow-[0_0_50px_rgba(var(--color-success-rgb),0.4)]">
                    <CheckCircle2 size={80} />
                  </div>
                </div>
                <div className="space-y-4">
                  <h3 className="text-4xl font-black text-white tracking-tight leading-tight">System Fully<br />Operational.</h3>
                  <p className="text-text-dim text-xl max-w-sm mx-auto font-medium leading-relaxed">
                    Your terminal is now equipped with the OpsForAll platform. All systems nominal.
                  </p>
                </div>
                <button
                  onClick={handleFinish}
                  className="w-full max-w-md py-7 rounded-2xl bg-success text-white text-2xl font-black shadow-[0_0_40px_rgba(var(--color-success-rgb),0.3)] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-5 group"
                >
                  Enter Ops Center
                  <ChevronRight size={32} className="group-hover:translate-x-1.5 transition-transform duration-500" />
                </button>
              </div>
            )}
          </div>

          {/* Footer Navigation */}
          {step !== 'finished' && !setupRunning && (
            <div className="p-10 bg-panel-2/20 border-t border-white/5 flex justify-between items-center backdrop-blur-md">
              <button
                onClick={() => {
                  if (step === 'dependencies') setStep('welcome')
                  if (step === 'ai-setup') setStep('dependencies')
                }}
                className={cn("text-text-faint font-bold text-xs uppercase tracking-[0.2em] hover:text-white transition-all", step === 'welcome' && "opacity-0 pointer-events-none")}
              >
                Back
              </button>

              <button
                onClick={() => {
                  if (step === 'welcome') setStep('dependencies')
                  if (step === 'dependencies') setStep('ai-setup')
                }}
                className="px-10 py-5 bg-white/5 border border-white/10 rounded-2xl text-white text-xs font-black uppercase tracking-[0.25em] hover:bg-accent hover:border-accent hover:shadow-[0_0_30px_rgba(var(--color-accent-rgb),0.3)] transition-all flex items-center gap-4 group active:scale-95"
              >
                Next Phase
                <ChevronRight size={18} className="group-hover:translate-x-1 transition-transform" />
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
