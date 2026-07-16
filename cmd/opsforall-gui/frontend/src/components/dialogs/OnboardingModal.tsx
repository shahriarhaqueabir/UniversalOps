import { useState, useEffect, useCallback } from 'react'
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
      if (status.available) {
        // If everything is already there, we could skip to finished,
        // but we'll let them see the AI setup step if models are missing.
      }
    } catch (err) {
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

      setStep('finished')
    } catch (err: any) {
      setError(err?.message || 'AI setup failed. You can skip this for now and configure it later in AIOps.')
    } finally {
      setSetupRunning(false)
    }
  }

  const handleFinish = async () => {
    await call('App.MarkOnboarded')
    onComplete()
  }

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-black/80 backdrop-blur-xl animate-in fade-in duration-500">
      <div className="bg-panel border border-border rounded-[2.5rem] shadow-[0_0_100px_rgba(0,0,0,0.8)] w-full max-w-2xl overflow-hidden relative">
        {/* Background Glow */}
        <div className="absolute -top-24 -left-24 w-96 h-96 bg-accent/20 rounded-full blur-[100px] pointer-events-none" />
        <div className="absolute -bottom-24 -right-24 w-96 h-96 bg-accent/10 rounded-full blur-[100px] pointer-events-none" />

        <div className="relative z-10">
          {/* Header */}
          <div className="px-10 py-10 border-b border-border bg-panel-2/50 flex items-center justify-between">
            <div className="flex items-center gap-5">
              <div className="w-16 h-16 rounded-2xl bg-accent flex items-center justify-center shadow-lg shadow-accent/20">
                <ShieldCheck size={40} className="text-white" />
              </div>
              <div>
                <h2 className="text-3xl font-black text-text tracking-tight italic">OpsForAll</h2>
                <p className="text-accent font-bold uppercase tracking-[0.2em] text-xs mt-1">Universal Command Center</p>
              </div>
            </div>
            <div className="flex gap-2">
              {(['welcome', 'dependencies', 'ai-setup', 'finished'] as Step[]).map((s, i) => (
                <div
                  key={s}
                  className={cn(
                    "w-3 h-3 rounded-full border-2 transition-all duration-300",
                    step === s ? "bg-accent border-accent scale-125" :
                    i < ['welcome', 'dependencies', 'ai-setup', 'finished'].indexOf(step) ? "bg-success border-success" : "border-border"
                  )}
                />
              ))}
            </div>
          </div>

          {/* Body */}
          <div className="p-12 min-h-[400px] flex flex-col">
            {step === 'welcome' && (
              <div className="flex flex-col items-center text-center space-y-8 animate-in slide-in-from-bottom-4 duration-500">
                <div className="w-24 h-24 rounded-3xl bg-panel-3 border border-border flex items-center justify-center text-accent">
                  <Zap size={48} />
                </div>
                <div className="space-y-4">
                  <h3 className="text-4xl font-bold text-text">Welcome to the future of Ops.</h3>
                  <p className="text-text-dim text-xl max-w-md mx-auto leading-relaxed">
                    OpsForAll brings enterprise-grade telemetry and local AI intelligence to your workstation.
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-4 w-full pt-4">
                  {[
                    { icon: <Activity className="text-success" />, label: 'Real-time Metrics' },
                    { icon: <BrainCircuit className="text-accent" />, label: 'Local AI Analyst' },
                    { icon: <ShieldCheck className="text-warning" />, label: 'Security Auditing' },
                    { icon: <Terminal className="text-text" />, label: 'DevOps Automation' },
                  ].map((item, i) => (
                    <div key={i} className="flex items-center gap-3 p-4 rounded-2xl bg-panel-2 border border-border">
                      {item.icon}
                      <span className="font-bold text-text-dim">{item.label}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {step === 'dependencies' && (
              <div className="space-y-8 animate-in slide-in-from-right-4 duration-500">
                <div className="space-y-2">
                  <h3 className="text-3xl font-bold text-text">System Pre-flight</h3>
                  <p className="text-text-dim text-lg">Checking if your machine is ready for advanced operations.</p>
                </div>

                <div className="space-y-4">
                  <div className="p-6 rounded-2xl bg-panel-2 border border-border flex items-center justify-between">
                    <div className="flex items-center gap-5">
                      <div className="w-12 h-12 rounded-xl bg-panel-3 flex items-center justify-center">
                        <Bot size={24} className="text-accent" />
                      </div>
                      <div>
                        <p className="font-bold text-text text-lg">Ollama AI Runtime</p>
                        <p className="text-sm text-text-faint">Required for local intelligent analysis</p>
                      </div>
                    </div>
                    {checking ? (
                      <RefreshCw size={24} className="animate-spin text-accent" />
                    ) : ollamaStatus?.binary_exists ? (
                      <CheckCircle2 size={32} className="text-success" />
                    ) : (
                      <div className="flex flex-col items-end gap-1">
                        <div className="flex items-center gap-2 text-danger font-bold">
                          <AlertCircle size={20} />
                          Missing
                        </div>
                        <a href="https://ollama.com" target="_blank" rel="noreferrer" className="text-xs text-accent hover:underline">Download from ollama.com</a>
                      </div>
                    )}
                  </div>
                </div>

                {!checking && !ollamaStatus?.binary_exists && (
                  <div className="p-5 rounded-2xl bg-danger/10 border border-danger/20 flex items-start gap-4">
                    <AlertCircle size={20} className="text-danger mt-1 shrink-0" />
                    <div className="space-y-1">
                      <p className="font-bold text-danger">Ollama not detected</p>
                      <p className="text-sm text-danger/80">You'll need to install Ollama to use the AI features. You can continue, but the Analyst will be disabled.</p>
                    </div>
                  </div>
                )}
              </div>
            )}

            {step === 'ai-setup' && (
              <div className="space-y-8 animate-in slide-in-from-right-4 duration-500">
                <div className="space-y-2">
                  <h3 className="text-3xl font-bold text-text">Neural Core Initialization</h3>
                  <p className="text-text-dim text-lg">Setting up your local operations specialist.</p>
                </div>

                {setupRunning ? (
                  <div className="space-y-10 py-10">
                    <div className="space-y-4">
                      <div className="flex justify-between text-sm font-bold uppercase tracking-widest text-text-dim">
                        <span>{setupProgress?.status || 'Processing...'}</span>
                        <span>{Math.round(setupProgress?.percent || 0)}%</span>
                      </div>
                      <div className="h-4 bg-panel-3 rounded-full overflow-hidden border border-border">
                        <div
                          className="h-full bg-accent transition-all duration-500 shadow-[0_0_15px_rgba(var(--color-accent-rgb),0.5)]"
                          style={{ width: `${setupProgress?.percent || 0}%` }}
                        />
                      </div>
                    </div>
                    <div className="flex flex-col items-center text-center space-y-4">
                      <RefreshCw size={48} className="animate-spin text-accent" />
                      <p className="text-text-faint italic">This may take a few minutes depending on your connection. Downloads are saved locally.</p>
                    </div>
                  </div>
                ) : (
                  <div className="flex flex-col items-center text-center space-y-8 py-4">
                    <div className="w-20 h-20 rounded-full bg-accent/10 flex items-center justify-center text-accent">
                      <Download size={40} />
                    </div>
                    <p className="text-text-dim text-lg leading-relaxed">
                      We need to download the base LLM and configure the specialized <strong>OpsForAll</strong> persona.
                      This requires about 5GB of disk space.
                    </p>
                    <button
                      onClick={handleAISetup}
                      className="w-full py-5 rounded-2xl bg-accent text-white text-xl font-black shadow-xl shadow-accent/20 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-3"
                    >
                      <Zap size={24} />
                      Download & Initialize AI
                    </button>
                    <button onClick={() => setStep('finished')} className="text-text-faint font-bold hover:text-text transition-colors">Skip for now (AI features will be unavailable)</button>
                  </div>
                )}

                {error && (
                  <div className="p-4 rounded-xl bg-danger/10 border border-danger/20 text-danger text-sm font-bold flex items-center gap-3">
                    <AlertCircle size={18} />
                    {error}
                  </div>
                )}
              </div>
            )}

            {step === 'finished' && (
              <div className="flex flex-col items-center text-center space-y-10 animate-in zoom-in-95 duration-700">
                <div className="relative">
                  <div className="absolute inset-0 bg-success blur-[40px] opacity-20 animate-pulse" />
                  <div className="w-32 h-32 rounded-full bg-success flex items-center justify-center text-white relative z-10 shadow-2xl">
                    <CheckCircle2 size={64} />
                  </div>
                </div>
                <div className="space-y-4">
                  <h3 className="text-4xl font-bold text-text tracking-tight">System Fully Operational.</h3>
                  <p className="text-text-dim text-xl max-w-md mx-auto">
                    Your workstation is now equipped with the OpsForAll platform. All systems nominal.
                  </p>
                </div>
                <button
                  onClick={handleFinish}
                  className="w-full py-6 rounded-2xl bg-success text-white text-2xl font-black shadow-xl shadow-success/20 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-4"
                >
                  Enter Command Center
                  <ChevronRight size={28} />
                </button>
              </div>
            )}
          </div>

          {/* Footer Navigation */}
          {step !== 'finished' && !setupRunning && (
            <div className="p-10 bg-panel-2/50 border-t border-border flex justify-between items-center">
              <button
                onClick={() => {
                  if (step === 'dependencies') setStep('welcome')
                  if (step === 'ai-setup') setStep('dependencies')
                }}
                className={cn("text-text-faint font-bold hover:text-text transition-all", step === 'welcome' && "opacity-0 pointer-events-none")}
              >
                Back
              </button>

              <button
                onClick={() => {
                  if (step === 'welcome') setStep('dependencies')
                  if (step === 'dependencies') setStep('ai-setup')
                }}
                className="px-10 py-4 bg-panel-3 border border-border rounded-xl text-text font-black hover:bg-panel hover:border-accent/40 transition-all flex items-center gap-3 shadow-lg group"
              >
                Next Step
                <ChevronRight size={20} className="group-hover:translate-x-1 transition-transform" />
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
