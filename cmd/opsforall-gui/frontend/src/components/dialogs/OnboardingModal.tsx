import { useState, useEffect, useRef, useCallback, type KeyboardEvent as ReactKeyboardEvent } from 'react'
import { toast } from 'sonner'
import {
  ShieldCheck,
  ChevronRight,
  RefreshCw,
  Activity,
  User,
  Database,
  Globe,
  Bot,
  ScrollText,
  CheckCircle2,
  Zap,
  BrainCircuit,
  HardDrive,
  Cpu,
  MemoryStick,
  Server,
  Eye,
  X,
  SkipForward,
  AlertTriangle,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { useOllamaStore } from '@/stores/useOllamaStore'
import { DependencyChecklist } from '../onboarding/DependencyChecklist'
import type {
  EnvironmentReport,
  PerformanceProfile,
  SystemCapability,
  SafetyPolicy,
  OnboardingConfig,
  OnboardingStep,
  AISetupRecommendation,
} from '@/types/onboarding'
import { ONBOARDING_STEPS } from '@/types/onboarding'

// ── Constants ──

const MAX_NAME_LENGTH = 32
const NAME_PATTERN = /^[a-zA-Z0-9\s_-]+$/
const DEFAULT_CONFIG: OnboardingConfig = {
  companionName: 'Hawk',
  storagePath: 'data',
  logsPath: 'logs',
  modelPath: '',
  safetyPolicy: 'standard',
  engineProfile: 'standard',
}

// ── Step metadata ──

const STEP_META: Record<OnboardingStep, { label: string; title: string; subtitle: string }> = {
  welcome: { label: 'Welcome', title: 'Welcome', subtitle: '' },
  privacy: { label: 'Privacy', title: 'Data & Privacy', subtitle: 'Your configuration and system history are stored locally.' },
  'system-check': { label: 'System', title: 'System Overview', subtitle: 'Reviewing your workstation resources and installed toolchains.' },
  snapshot: { label: 'Snapshot', title: 'Baseline Snapshot', subtitle: 'Saving current system configuration for future monitoring.' },
  'ai-setup': { label: 'AI', title: 'AI Assistant', subtitle: 'Configuring your local assistant for system analysis.' },
  finished: { label: 'Done', title: 'Setup Complete', subtitle: 'Platform initialization is finished. Your environment is ready.' },
}

// ── Component ──

interface OnboardingModalProps {
  onComplete: () => void
}

export function OnboardingModal({ onComplete }: OnboardingModalProps) {
  const { call } = useBackend()
  const { setCompanionName } = useSettingsStore()
  const modalRef = useRef<HTMLDivElement>(null)
  const continueBtnRef = useRef<HTMLButtonElement>(null)
  const previousFocusRef = useRef<HTMLElement | null>(null)

  const [step, setStep] = useState<OnboardingStep>('welcome')
  const [envReport, setEnvReport] = useState<EnvironmentReport | null>(null)
  const [dependencies, setDependencies] = useState<SystemCapability[]>([])
  const [perfProfile, setPerfProfile] = useState<PerformanceProfile | null>(null)
  const [config, setConfig] = useState<OnboardingConfig>(DEFAULT_CONFIG)
  const [setupRunning, setSetupRunning] = useState(false)
  const [snapshotDone, setSnapshotDone] = useState(false)
  const [aiRecommendation, setAiRecommendation] = useState<AISetupRecommendation | null>(null)
  const aiProgress = useOllamaStore((s) => s.progress)
  const [isCapturing, setIsCapturing] = useState(false)
  const [nameError, setNameError] = useState('')
  const [discoveryError, setDiscoveryError] = useState(false)
  const [depError, setDepError] = useState(false)

  const stepIndex = ONBOARDING_STEPS.indexOf(step)
  const isLastStep = stepIndex === ONBOARDING_STEPS.length - 1

  // ── Focus trap + restore ──

  useEffect(() => {
    previousFocusRef.current = document.activeElement as HTMLElement
    const timer = setTimeout(() => modalRef.current?.focus(), 50)
    return () => {
      clearTimeout(timer)
      previousFocusRef.current?.focus()
    }
  }, [])

  useEffect(() => {
    if (!isLastStep) {
      const timer = setTimeout(() => continueBtnRef.current?.focus(), 100)
      return () => clearTimeout(timer)
    }
  }, [step, isLastStep])

  // ── Actions (declared before keyboard handler) ──

  const handleSkipAll = useCallback(() => {
    setCompanionName(DEFAULT_CONFIG.companionName)
    call('App.ApplyOperationalProfile', DEFAULT_CONFIG.engineProfile).catch(() => {})
    call('App.MarkOnboarded').catch(() => {})
    onComplete()
  }, [call, setCompanionName, onComplete])

  // ── Keyboard: Escape to skip, Tab trap ──

  const handleKeyDown = useCallback((e: ReactKeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault()
      handleSkipAll()
      return
    }
    if (e.key === 'Tab' && modalRef.current) {
      const focusable = modalRef.current.querySelectorAll<HTMLElement>(
        'button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex="-1"])'
      )
      if (focusable.length === 0) return
      const first = focusable[0]
      const last = focusable[focusable.length - 1]
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault()
        last.focus()
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault()
        first.focus()
      }
    }
  }, [handleSkipAll])

  // ── Data fetching ──

  const fetchDiscovery = async () => {
    setDiscoveryError(false)
    try {
      const [env, profile] = await Promise.all([
        call('App.DiscoverEnvironment'),
        call('App.GetPerformanceProfile'),
      ])
      setEnvReport(env as EnvironmentReport)
      const p = profile as PerformanceProfile
      setPerfProfile(p)
      if (p?.category) {
        setConfig((prev) => ({ ...prev, engineProfile: p.category }))
      }
    } catch {
      setDiscoveryError(true)
    }
  }

  const fetchDependencies = async () => {
    setDepError(false)
    try {
      const res = await call('App.GetSystemCapabilities')
      // Map Go response to internal frontend types if necessary,
      // though DependencyChecklist handles the Go CapabilityInfo format now.
      setDependencies(res as any[])
    } catch {
      setDepError(true)
    }
  }

  const generateSnapshot = async () => {
    if (snapshotDone) return
    setIsCapturing(true)
    try {
      await call('App.GenerateBaselineSnapshot')
      await new Promise((r) => setTimeout(r, 1200))
      setSnapshotDone(true)
    } catch {
      toast.error('Baseline snapshot failed. You can retry or skip.')
    } finally {
      setIsCapturing(false)
    }
  }

  // ── Folder/file pickers ──

  const pickFolder = async (key: 'storagePath' | 'logsPath') => {
    try {
      const path = await call('App.SelectFolderDialog', `Select ${key === 'storagePath' ? 'Data' : 'Logs'} Folder`)
      if (typeof path === 'string' && path) {
        setConfig((prev) => ({ ...prev, [key]: path }))
      }
    } catch {
      toast.error('Folder selection failed')
    }
  }

  const pickModelfile = async () => {
    try {
      const path = await call('App.OpenFileDialog', 'Select Modelfile', ['Modelfile|*.modelfile', 'All Files|*.*'])
      if (typeof path === 'string' && path) {
        setConfig((prev) => ({ ...prev, modelPath: path }))
      }
    } catch {
      toast.error('File selection failed')
    }
  }

  // ── Name validation ──

  const updateName = (value: string) => {
    if (value.length > MAX_NAME_LENGTH) {
      setNameError(`Max ${MAX_NAME_LENGTH} characters`)
      return
    }
    if (value && !NAME_PATTERN.test(value)) {
      setNameError('Letters, numbers, spaces, _ or - only')
      return
    }
    setNameError('')
    setConfig((prev) => ({ ...prev, companionName: value }))
  }

  // ── AI setup ──

  const handleAISetup = async () => {
    const rec = aiRecommendation
    if (!rec) {
      toast.error('No AI recommendation available. Run system check first.')
      return
    }
    setSetupRunning(true)
    useOllamaStore.getState().setProgress(null)
    try {
      await call('AIOps.SetupOllamaPersona', rec.recommended_model)
      setStep('finished')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'AI setup failed'
      toast.error(msg)
    } finally {
      setSetupRunning(false)
      useOllamaStore.getState().setProgress(null)
    }
  }

  const fetchAiRecommendation = useCallback(async () => {
    try {
      const rec = await call('AIOps.GetAISetupRecommendation')
      setAiRecommendation(rec as AISetupRecommendation)
    } catch {
      setAiRecommendation(null)
    }
  }, [call])

  // ── Navigation ──

  const nextStep = () => {
    const idx = ONBOARDING_STEPS.indexOf(step)
    if (idx < ONBOARDING_STEPS.length - 1) {
      const next = ONBOARDING_STEPS[idx + 1]
      if (next === 'system-check') {
        fetchDiscovery()
        fetchDependencies()
      }
      if (next === 'snapshot' && !snapshotDone) generateSnapshot()
      if (next === 'ai-setup') fetchAiRecommendation()
      setStep(next)
    }
  }

  const prevStep = () => {
    const idx = ONBOARDING_STEPS.indexOf(step)
    if (idx > 0) setStep(ONBOARDING_STEPS[idx - 1])
  }

  const handleFinalDeploy = async () => {
    try {
      setCompanionName(config.companionName)
      await call('App.ApplyOperationalProfile', config.engineProfile)
      if (config.storagePath !== 'data') {
        await call('App.UpdateStorageConfig', config.storagePath)
      }
      if (config.logsPath !== 'logs') {
        await call('App.UpdateLogsConfig', config.logsPath)
      }
      await call('App.MarkOnboarded')
      onComplete()
    } catch {
      toast.error('Setup finalization failed')
    }
  }

  const canAdvance = (): boolean => {
    if (step === 'welcome') return true
    if (step === 'privacy') return !nameError && config.companionName.trim().length > 0
    if (step === 'system-check') return true
    if (step === 'snapshot') return snapshotDone || !isCapturing
    if (step === 'ai-setup') return !setupRunning
    return true
  }

  // ── Render ──

  return (
    <div
      className="fixed inset-0 z-[200] flex items-center justify-center p-4 sm:p-6 bg-[var(--color-overlay)] backdrop-blur-xl"
      role="dialog"
      aria-modal="true"
      aria-label="Platform Setup"
      ref={modalRef}
      tabIndex={-1}
      onKeyDown={handleKeyDown}
    >
      <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-2xl shadow-2xl w-full max-w-4xl h-[85dvh] max-h-[800px] flex flex-col overflow-hidden relative">

        {/* ── Fixed Header ── */}
        <header className="px-6 sm:px-10 py-4 sm:py-5 border-b border-[var(--color-border)] flex items-center justify-between shrink-0 bg-[var(--color-panel)]">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-[var(--color-accent)] flex items-center justify-center">
              <ShieldCheck size={20} className="text-white" />
            </div>
            <div>
              <h2 className="text-base sm:text-lg font-bold text-[var(--color-text)] tracking-tight">
                Platform Setup
              </h2>
              <p className="text-[10px] text-[var(--color-text-faint)] font-semibold uppercase tracking-[0.15em]">
                Step {stepIndex + 1} of {ONBOARDING_STEPS.length}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            {/* Progress dots */}
            <div
              className="flex gap-1"
              role="progressbar"
              aria-valuenow={stepIndex + 1}
              aria-valuemin={1}
              aria-valuemax={ONBOARDING_STEPS.length}
              aria-label={`Step ${stepIndex + 1} of ${ONBOARDING_STEPS.length}`}
            >
              {ONBOARDING_STEPS.map((s, i) => (
                <div
                  key={s}
                  className={cn(
                    'h-1.5 rounded-full transition-all duration-300',
                    i === stepIndex ? 'w-6 bg-[var(--color-accent)]' : i < stepIndex ? 'w-3 bg-[var(--color-accent)] opacity-50' : 'w-3 bg-[var(--color-text-faint)] opacity-20'
                  )}
                />
              ))}
            </div>
            {/* Skip / close button */}
            <button
              onClick={handleSkipAll}
              data-automation-id="onboarding-skip"
              className="ml-2 p-2.5 min-h-[40px] min-w-[40px] rounded-lg text-[var(--color-text-faint)] hover:text-[var(--color-text)] hover:bg-[var(--color-panel-2)] transition-colors flex items-center justify-center"
              aria-label="Skip setup and use defaults"
              title="Skip setup (Esc)"
            >
              <X size={16} />
            </button>
          </div>
        </header>

        {/* ── Scrollable Content ── */}
        <div className="flex-1 overflow-y-auto px-6 sm:px-10 py-8 sm:py-12">

          {/* ── WELCOME ── */}
          {step === 'welcome' && (
            <div
              className="flex flex-col items-center text-center space-y-8 sm:space-y-10 h-full justify-center max-w-2xl mx-auto"
              style={{ animation: 'onb-enter 0.6s cubic-bezier(0.32,0.72,0,1) both' }}
            >
              <div className="w-16 h-16 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center text-[var(--color-accent)] shadow-md">
                <Zap size={32} />
              </div>
              <div className="space-y-3">
                <h3 className="text-2xl sm:text-3xl font-black text-[var(--color-text)]">
                  Private. Local.{' '}
                  <span className="text-[var(--color-accent)] italic">Professional.</span>
                </h3>
                <p className="text-[var(--color-text-dim)] text-base max-w-md mx-auto">
                  Welcome to Universal-Ops. Let&apos;s configure your local operations environment.
                </p>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 w-full">
                {[
                  { icon: <Globe size={18} />, title: 'Private', desc: 'No data leaves this machine. Zero cloud sync or tracking.' },
                  { icon: <BrainCircuit size={18} />, title: 'Local AI', desc: 'Intelligence powered by models running on your hardware.' },
                  { icon: <Activity size={18} />, title: 'Telemetry', desc: 'Real-time monitoring of local system resources.' },
                ].map((item) => (
                  <div
                    key={item.title}
                    className="p-6 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] space-y-3 text-left"
                  >
                    <div className="text-[var(--color-accent)]">{item.icon}</div>
                    <p className="font-bold text-[var(--color-text)] text-[11px] uppercase tracking-[0.12em]">
                      {item.title}
                    </p>
                    <p className="text-[11px] text-[var(--color-text-faint)] leading-relaxed font-medium">
                      {item.desc}
                    </p>
                  </div>
                ))}
              </div>
              <button
                onClick={handleSkipAll}
                className="flex items-center gap-2 text-[11px] font-semibold text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors"
              >
                <SkipForward size={14} />
                Skip to defaults
              </button>
            </div>
          )}

          {/* ── PRIVACY ── */}
          {step === 'privacy' && (
            <div
              className="max-w-3xl mx-auto space-y-8 h-full flex flex-col justify-center"
              style={{ animation: 'onb-enter 0.6s cubic-bezier(0.32,0.72,0,1) both' }}
            >
              <div className="text-center space-y-1">
                <h3 className="text-xl sm:text-2xl font-black text-[var(--color-text)]">
                  {STEP_META.privacy.title}
                </h3>
                <p className="text-sm text-[var(--color-text-dim)]">{STEP_META.privacy.subtitle}</p>
              </div>

              {/* Assistant name */}
              <div className="space-y-2">
                <label htmlFor="companion-name" className="text-[11px] font-bold text-[var(--color-text-faint)] uppercase tracking-[0.12em]">
                  Assistant Name
                </label>
                <div className="relative">
                  <input
                    id="companion-name"
                    type="text"
                    value={config.companionName}
                    onChange={(e) => updateName(e.target.value)}
                    maxLength={MAX_NAME_LENGTH}
                    aria-required="true"
                    aria-invalid={!!nameError}
                    aria-describedby={nameError ? 'name-error' : 'name-hint'}
                    className={cn(
                      'w-full bg-[var(--color-panel-2)] border rounded-xl px-6 py-4 text-lg font-bold text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] transition-all',
                      nameError ? 'border-[var(--color-danger)]' : 'border-[var(--color-border)]'
                    )}
                    placeholder="e.g. Hawk"
                  />
                  <User size={18} className="absolute left-5 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)] opacity-40" />
                  <span className="absolute right-4 top-1/2 -translate-y-1/2 text-[10px] text-[var(--color-text-faint)] font-mono">
                    {config.companionName.length}/{MAX_NAME_LENGTH}
                  </span>
                </div>
                {nameError ? (
                  <p id="name-error" className="text-[11px] text-[var(--color-danger)] font-medium" role="alert">
                    {nameError}
                  </p>
                ) : (
                  <p id="name-hint" className="text-[11px] text-[var(--color-text-faint)]">
                    <Bot size={12} className="inline mr-1 opacity-60" />
                    Used for system briefings, reports, and security actions.
                  </p>
                )}
              </div>

              {/* Storage paths */}
              <div className="space-y-3">
                <p className="text-[11px] font-bold text-[var(--color-text-faint)] uppercase tracking-[0.12em]">
                  Storage Locations
                </p>
                {[
                  { icon: <Database size={16} />, title: 'Database', path: `${config.storagePath}/universalops.db`, desc: 'Local SQLite database for metrics and alerts.', action: () => pickFolder('storagePath'), pathKey: 'storagePath' as const },
                  { icon: <HardDrive size={16} />, title: 'AI Configuration', path: config.modelPath || './data/universalops.modelfile', desc: 'Local assistant settings and rules.', action: pickModelfile, pathKey: 'modelPath' as const },
                  { icon: <ScrollText size={16} />, title: 'Logs', path: `${config.logsPath}/universalops.log`, desc: 'System event logs for auditing.', action: () => pickFolder('logsPath'), pathKey: 'logsPath' as const },
                ].map((item) => (
                  <div
                    key={item.title}
                    className="p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-between gap-4"
                  >
                    <div className="flex items-center gap-4 min-w-0">
                      <div className="w-9 h-9 rounded-lg bg-[var(--color-panel-3)] flex items-center justify-center shrink-0 border border-[var(--color-border)] text-[var(--color-accent)]">
                        {item.icon}
                      </div>
                      <div className="min-w-0">
                        <p className="font-bold text-[var(--color-text)] text-[11px] uppercase tracking-wider">{item.title}</p>
                        <p className="text-[10px] text-[var(--color-accent)] font-mono truncate mt-0.5">{item.path}</p>
                        <p className="text-[10px] text-[var(--color-text-faint)] mt-0.5 font-medium">{item.desc}</p>
                      </div>
                    </div>
                    <button
                      onClick={item.action}
                      className="px-5 py-3 min-h-[44px] rounded-lg bg-[var(--color-panel-2)] hover:bg-[var(--color-panel-3)] text-[10px] font-bold uppercase tracking-[0.1em] text-[var(--color-text)] border border-[var(--color-border)] transition-all shrink-0"
                    >
                      Browse
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* ── SYSTEM CHECK (merged discovery + dependency) ── */}
          {step === 'system-check' && (
            <div
              className="max-w-4xl mx-auto space-y-8"
              style={{ animation: 'onb-enter 0.6s cubic-bezier(0.32,0.72,0,1) both' }}
            >
              <div className="text-center space-y-1">
                <h3 className="text-xl sm:text-2xl font-black text-[var(--color-text)]">
                  {STEP_META['system-check'].title}
                </h3>
                <p className="text-sm text-[var(--color-text-dim)]">{STEP_META['system-check'].subtitle}</p>
              </div>

              {/* Environment section */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <p className="text-[11px] font-bold text-[var(--color-text-faint)] uppercase tracking-[0.12em]">
                    Workstation Resources
                  </p>
                  {discoveryError && (
                    <button
                      onClick={fetchDiscovery}
                      className="flex items-center gap-1.5 text-[11px] font-semibold text-[var(--color-accent)] hover:text-[var(--color-accent-2)] transition-colors"
                    >
                      <RefreshCw size={12} />
                      Retry
                    </button>
                  )}
                </div>

                {!envReport && !discoveryError ? (
                  <div className="py-16 flex flex-col items-center gap-3 text-[var(--color-text-faint)]">
                    <RefreshCw size={32} className="animate-spin opacity-20" />
                    <p className="text-[11px] font-bold uppercase tracking-[0.12em]">Checking System...</p>
                  </div>
                ) : discoveryError ? (
                  <div className="py-12 flex flex-col items-center gap-3 text-[var(--color-text-faint)]">
                    <AlertTriangle size={32} className="text-[var(--color-warning)] opacity-60" />
                    <p className="text-[11px] font-bold uppercase tracking-[0.12em]">System Check Failed</p>
                    <button onClick={fetchDiscovery} className="text-[11px] text-[var(--color-accent)] font-semibold hover:underline">
                      Tap to retry
                    </button>
                  </div>
                ) : (
                  <div className="p-6 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] grid grid-cols-1 sm:grid-cols-2 gap-x-10 gap-y-5">
                    {[
                      { label: 'Hostname', value: envReport!.hostname, icon: <Server size={12} /> },
                      { label: 'Processor', value: envReport!.cpu, icon: <Cpu size={12} />, color: 'text-[var(--color-accent)]' },
                      { label: 'Memory', value: envReport!.memory, icon: <MemoryStick size={12} /> },
                      { label: 'OS Platform', value: `${envReport!.os} (${envReport!.arch})`, icon: <Globe size={12} /> },
                    ].map((item) => (
                      <div key={item.label} className="space-y-1 min-w-0">
                        <div className="flex items-center gap-2 opacity-40">
                          {item.icon}
                          <span className="text-[10px] font-bold uppercase tracking-[0.1em]">{item.label}</span>
                        </div>
                        <p className={cn('text-sm font-bold truncate', item.color || 'text-[var(--color-text)]')}>
                          {item.value}
                        </p>
                      </div>
                    ))}
                  </div>
                )}

                {envReport && (
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div className="p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] space-y-3">
                      <div className="flex items-center gap-2">
                        <Zap size={14} className="text-[var(--color-warning)]" />
                        <span className="text-[10px] font-bold uppercase tracking-[0.1em] text-[var(--color-text)]">Performance Profile</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-medium text-[var(--color-text-faint)]">Category</span>
                        <span className="px-2 py-0.5 bg-[var(--color-warning)]/10 text-[var(--color-warning)] rounded text-[9px] font-bold uppercase border border-[var(--color-warning)]/20">
                          {perfProfile?.category || '...'}
                        </span>
                      </div>
                      {perfProfile && (
                        <p className="text-[10px] text-[var(--color-text-dim)] leading-relaxed">
                          Detected {perfProfile.cpu_threads} threads and {perfProfile.memory_gb.toFixed(0)}GB RAM.
                        </p>
                      )}
                    </div>
                    <div className="p-5 rounded-xl bg-[var(--color-panel-2)] border border-[var(--color-border)] space-y-3">
                      <div className="flex items-center gap-2">
                        <Globe size={14} className="text-[var(--color-accent)]" />
                        <span className="text-[10px] font-bold uppercase tracking-[0.1em] text-[var(--color-text)]">Network Interfaces</span>
                      </div>
                      <div className="flex flex-wrap gap-1.5">
                        {envReport.interfaces?.slice(0, 4).map((iface: string) => (
                          <span key={iface} className="px-2 py-0.5 bg-[var(--color-panel-2)] rounded text-[9px] font-mono text-[var(--color-text-dim)] border border-[var(--color-border)]">
                            {iface}
                          </span>
                        ))}
                        {(!envReport.interfaces || envReport.interfaces.length === 0) && (
                          <span className="text-[10px] text-[var(--color-text-faint)]">No interfaces detected</span>
                        )}
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {/* Tools section */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <p className="text-[11px] font-bold text-[var(--color-text-faint)] uppercase tracking-[0.12em]">
                    System Tools
                  </p>
                  {depError && (
                    <button
                      onClick={fetchDependencies}
                      className="flex items-center gap-1.5 text-[11px] font-semibold text-[var(--color-accent)] hover:text-[var(--color-accent-2)] transition-colors"
                    >
                      <RefreshCw size={12} />
                      Retry
                    </button>
                  )}
                </div>

                {depError ? (
                  <div className="py-8 flex flex-col items-center gap-3 text-[var(--color-text-faint)]">
                    <AlertTriangle size={24} className="text-[var(--color-warning)] opacity-60" />
                    <p className="text-[11px] font-bold uppercase tracking-[0.12em]">Tool discovery failed</p>
                    <button onClick={fetchDependencies} className="text-[11px] text-[var(--color-accent)] font-semibold hover:underline">
                      Tap to retry
                    </button>
                  </div>
                ) : dependencies.length === 0 && !depError ? (
                  <div className="py-8 flex flex-col items-center gap-3 text-[var(--color-text-faint)]">
                    <RefreshCw size={24} className="animate-spin opacity-20" />
                    <p className="text-[11px] font-bold uppercase tracking-[0.12em]">Scanning tools...</p>
                  </div>
                ) : (
                  <DependencyChecklist
                    dependencies={dependencies as any}
                    onRefresh={fetchDependencies}
                  />
                )}

                <div className="p-4 rounded-xl bg-[var(--color-panel-3)] border border-[var(--color-border)] flex items-center gap-3 max-w-2xl">
                  <Eye size={14} className="text-[var(--color-text-faint)] shrink-0" />
                  <p className="text-[10px] text-[var(--color-text-dim)] font-medium">
                    Missing tools will not block startup. We use built-in system collectors.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* ── SNAPSHOT ── */}
          {step === 'snapshot' && (
            <div
              className="max-w-xl mx-auto h-full flex flex-col items-center text-center justify-center space-y-8"
              style={{ animation: 'onb-enter 0.6s cubic-bezier(0.32,0.72,0,1) both' }}
            >
              <div
                className={cn(
                  'w-20 h-20 rounded-2xl border-2 flex items-center justify-center transition-all duration-500',
                  snapshotDone
                    ? 'bg-[var(--color-success)]/10 border-[var(--color-success)]/40 text-[var(--color-success)]'
                    : 'bg-[var(--color-accent)]/10 border-[var(--color-accent)]/40 text-[var(--color-accent)]'
                )}
              >
                {snapshotDone ? <CheckCircle2 size={40} /> : <RefreshCw size={40} className="animate-spin" />}
              </div>
              <div className="space-y-3">
                <h3 className="text-2xl sm:text-3xl font-black text-[var(--color-text)]">
                  {snapshotDone ? 'Status Saved' : 'Saving State'}
                </h3>
                <p className="text-[var(--color-text-dim)] text-base leading-relaxed">
                  {snapshotDone
                    ? 'System baseline captured for drift detection.'
                    : 'Saving current system configuration as a baseline for future monitoring.'}
                </p>
              </div>
              <div className="grid grid-cols-2 gap-3 w-full">
                {['Hardware', 'Software', 'Network', 'Security'].map((label) => (
                  <div
                    key={label}
                    className="p-4 rounded-lg bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-between"
                  >
                    <span className="text-[10px] font-bold uppercase tracking-[0.1em] text-[var(--color-text-faint)]">{label}</span>
                    <span
                      className={cn(
                        'text-[10px] font-bold uppercase',
                        snapshotDone ? 'text-[var(--color-success)]' : 'text-[var(--color-warning)]'
                      )}
                    >
                      {snapshotDone ? 'Verified' : 'Saving...'}
                    </span>
                  </div>
                ))}
              </div>
              {!snapshotDone && !isCapturing && (
                <button
                  onClick={generateSnapshot}
                  className="flex items-center gap-2 text-[11px] font-semibold text-[var(--color-accent)] hover:underline"
                >
                  <RefreshCw size={12} />
                  Retry snapshot
                </button>
              )}
            </div>
          )}

          {/* ── AI SETUP (recommendation + progress) ── */}
          {step === 'ai-setup' && (
            <div
              className="max-w-2xl mx-auto h-full flex flex-col justify-center space-y-6"
              style={{ animation: 'onb-enter 0.6s cubic-bezier(0.32,0.72,0,1) both' }}
            >
              <div className="text-center space-y-1">
                <h3 className="text-xl sm:text-2xl font-black text-[var(--color-text)]">
                  {STEP_META['ai-setup'].title}
                </h3>
                <p className="text-sm text-[var(--color-text-dim)]">{STEP_META['ai-setup'].subtitle}</p>
              </div>

              {setupRunning ? (
                <div className="space-y-6 flex flex-col items-center">
                  <RefreshCw size={40} className="animate-spin text-[var(--color-accent)]" />
                  <div className="w-full max-w-sm space-y-3">
                    <div className="flex justify-between text-[10px] font-bold uppercase tracking-[0.1em]">
                      <span className="text-[var(--color-accent)]">
                        {aiProgress?.status || 'Downloading model...'}
                      </span>
                      <span className="text-[var(--color-text-faint)]">
                        {aiProgress && aiProgress.total > 0
                          ? `${Math.round(aiProgress.percent)}%`
                          : 'In progress'}
                      </span>
                    </div>
                    <div className="h-2 bg-[var(--color-panel-3)] rounded-full overflow-hidden">
                      <div
                        className="h-full bg-[var(--color-accent)] rounded-full transition-all duration-300 ease-out"
                        style={{
                          width: aiProgress && aiProgress.total > 0
                            ? `${aiProgress.percent}%`
                            : '30%',
                          animation: aiProgress?.total
                            ? 'none'
                            : 'onb-progress-indeterminate 2s ease-in-out infinite',
                        }}
                      />
                    </div>
                    {aiProgress && aiProgress.total > 0 && (
                      <p className="text-[10px] text-[var(--color-text-faint)] text-center font-mono">
                        {Math.round(aiProgress.completed / 1_000_000)} MB / {Math.round(aiProgress.total / 1_000_000)} MB
                      </p>
                    )}
                    {aiRecommendation && (
                      <p className="text-[10px] text-[var(--color-text-dim)] text-center">
                        Setting up <span className="font-bold text-[var(--color-accent)]">{aiRecommendation.recommended_label}</span> ...
                      </p>
                    )}
                  </div>
                </div>
              ) : (
                <div className="space-y-6">
                  {/* Safety policy selector */}
                  <div className="space-y-3">
                    <p className="text-[11px] font-bold text-[var(--color-text-faint)] uppercase tracking-[0.12em]">
                      Security Policy
                    </p>
                    <p className="text-[11px] text-[var(--color-text-dim)]">
                      Define how the AI assistant interacts with system settings.
                    </p>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      {([
                        { id: 'standard' as SafetyPolicy, title: 'Standard', desc: 'Manual control. Human confirmation required for all actions.', icon: <CheckCircle2 size={16} /> },
                        { id: 'tactical' as SafetyPolicy, title: 'Tactical', desc: 'Faster response. Best for professional power users.', icon: <Zap size={16} /> },
                      ]).map((p) => (
                        <button
                          key={p.id}
                          onClick={() => setConfig((prev) => ({ ...prev, safetyPolicy: p.id }))}
                          className={cn(
                            'p-6 rounded-xl border-2 transition-all text-left space-y-2',
                            config.safetyPolicy === p.id
                              ? 'bg-[var(--color-accent)]/10 border-[var(--color-accent)]'
                              : 'bg-[var(--color-panel-2)] border-[var(--color-border)] hover:border-[var(--color-border)] hover:brightness-125'
                          )}
                          role="radio"
                          aria-checked={config.safetyPolicy === p.id}
                        >
                          <div className="flex items-center justify-between">
                            <span className="font-bold text-sm uppercase tracking-[0.1em] text-[var(--color-text)]">{p.title}</span>
                            <div className={config.safetyPolicy === p.id ? 'text-[var(--color-accent)]' : 'text-[var(--color-text-faint)]'}>
                              {p.icon}
                            </div>
                          </div>
                          <p className="text-[11px] text-[var(--color-text-dim)] font-medium leading-relaxed">{p.desc}</p>
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* AI recommendation card */}
                  {aiRecommendation ? (
                    <div className="p-6 rounded-xl bg-[var(--color-accent)]/5 border border-[var(--color-accent)]/20 space-y-4">
                      <div className="flex items-center gap-3">
                        <BrainCircuit size={20} className="text-[var(--color-accent)]" />
                        <p className="text-[11px] font-bold text-[var(--color-text)] uppercase">AI Model Recommendation</p>
                      </div>
                      <div className="bg-[var(--color-bg)]/60 p-5 rounded-lg border border-[var(--color-border)] font-mono text-[10px] leading-relaxed space-y-2">
                        <div className="grid grid-cols-3 gap-2">
                          <span className="text-[var(--color-text-faint)]">MODEL:</span>
                          <span className="col-span-2 text-[var(--color-accent)] font-bold">{aiRecommendation.recommended_label}</span>
                          <span className="text-[var(--color-text-faint)]">RAM:</span>
                          <span className="col-span-2">{aiRecommendation.system_ram_gb.toFixed(0)} GB</span>
                          <span className="text-[var(--color-text-faint)]">CPU:</span>
                          <span className="col-span-2">{aiRecommendation.system_cpu_threads} threads</span>
                          <span className="text-[var(--color-text-faint)]">STATUS:</span>
                          <span className="col-span-2">
                            {aiRecommendation.pull_required
                              ? <span className="text-[var(--color-warning)]">Download required</span>
                              : <span className="text-[var(--color-success)]">Ready to use</span>}
                          </span>
                        </div>
                      </div>
                      {aiRecommendation.fallback_models && aiRecommendation.fallback_models.length > 0 && (
                        <div className="space-y-1.5">
                          <p className="text-[9px] font-bold text-[var(--color-text-faint)] uppercase tracking-[0.12em]">Fallback options</p>
                          <div className="flex flex-wrap gap-1.5">
                            {aiRecommendation.fallback_models.map((m) => (
                              <span key={m.name} className="px-2 py-0.5 bg-[var(--color-panel-2)] rounded text-[9px] font-mono text-[var(--color-text-dim)] border border-[var(--color-border)]">
                                {m.label} ({m.size_gb.toFixed(1)} GB)
                              </span>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="py-8 flex flex-col items-center gap-3 text-[var(--color-text-faint)]">
                      <RefreshCw size={24} className="animate-spin opacity-20" />
                      <p className="text-[11px] font-bold uppercase tracking-[0.12em]">Analyzing system...</p>
                    </div>
                  )}

                  <div className="flex flex-col gap-3">
                    <button
                      onClick={handleAISetup}
                      disabled={!aiRecommendation}
                      className={cn(
                        'w-full py-4 min-h-[48px] rounded-xl font-bold uppercase tracking-[0.1em] text-xs shadow-lg transition-all',
                        aiRecommendation
                          ? 'bg-[var(--color-accent)] text-white hover:brightness-110 active:scale-[0.99] cursor-pointer'
                          : 'bg-[var(--color-panel-3)] text-[var(--color-text-faint)] cursor-not-allowed'
                      )}
                    >
                      {aiRecommendation?.pull_required ? 'Download & Initialize' : 'Initialize Assistant'}
                    </button>
                    <button
                      onClick={nextStep}
                      className="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-all text-center py-2"
                    >
                      Skip AI Setup
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* ── FINISHED ── */}
          {step === 'finished' && (
            <div
              className="max-w-xl mx-auto h-full flex flex-col items-center text-center justify-center space-y-8"
              style={{ animation: 'onb-enter 0.6s cubic-bezier(0.32,0.72,0,1) both' }}
            >
              <div
                className="w-24 h-24 rounded-full bg-[var(--color-success)] flex items-center justify-center text-white shadow-lg"
                style={{ animation: 'onb-scale-in 0.5s cubic-bezier(0.32,0.72,0,1) 0.2s both' }}
              >
                <CheckCircle2 size={48} />
              </div>
              <div className="space-y-3">
                <h3 className="text-3xl sm:text-4xl font-black text-[var(--color-text)]">
                  Setup Complete
                </h3>
                <p className="text-[var(--color-text-dim)] text-base">
                  Platform initialization is finished. Your environment is ready.
                </p>
              </div>
              <div className="grid grid-cols-2 gap-3 w-full">
                <div className="p-5 rounded-xl bg-[var(--color-panel-3)] border border-[var(--color-border)] text-left">
                  <p className="text-[10px] font-bold text-[var(--color-text-faint)] uppercase mb-1">Status</p>
                  <p className="text-xs font-bold text-[var(--color-success)] uppercase">Local Mode Active</p>
                </div>
                <div className="p-5 rounded-xl bg-[var(--color-panel-3)] border border-[var(--color-border)] text-left">
                  <p className="text-[10px] font-bold text-[var(--color-text-faint)] uppercase mb-1">Policy</p>
                  <p className="text-xs font-bold text-[var(--color-accent)] uppercase">{config.safetyPolicy}</p>
                </div>
              </div>
              <button
                onClick={handleFinalDeploy}
                className="w-full py-5 min-h-[56px] rounded-xl bg-[var(--color-accent)] text-white text-lg font-bold uppercase tracking-[0.08em] shadow-lg hover:brightness-110 active:scale-[0.99] transition-all flex items-center justify-center gap-3"
              >
                Enter Control Center
                <ChevronRight size={20} />
              </button>
            </div>
          )}
        </div>

        {/* ── Fixed Footer Navigation ── */}
        {!isLastStep && (
          <footer className="px-6 sm:px-10 py-5 sm:py-6 bg-[var(--color-panel)] border-t border-[var(--color-border)] flex justify-between items-center shrink-0">
            <button
              onClick={prevStep}
              className={cn(
                'text-[var(--color-text-faint)] font-bold text-[11px] uppercase tracking-[0.1em] hover:text-[var(--color-text)] transition-all min-h-[44px] px-4',
                step === 'welcome' && 'opacity-0 pointer-events-none'
              )}
            >
              Back
            </button>
            <button
              ref={continueBtnRef}
              onClick={nextStep}
              disabled={!canAdvance()}
              className={cn(
                'px-8 py-3 min-h-[44px] text-white text-[11px] font-bold uppercase tracking-[0.1em] rounded-xl transition-all flex items-center gap-3',
                canAdvance()
                  ? 'bg-[var(--color-accent)] hover:brightness-110 active:scale-95 shadow-md cursor-pointer'
                  : 'bg-[var(--color-panel-3)] text-[var(--color-text-faint)] cursor-not-allowed'
              )}
            >
              {isCapturing ? 'Processing...' : 'Continue'}
              {!isCapturing && <ChevronRight size={14} />}
            </button>
          </footer>
        )}
      </div>
    </div>
  )
}
