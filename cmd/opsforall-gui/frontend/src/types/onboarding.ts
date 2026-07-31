// ── Onboarding-specific types ──

export interface EnvironmentReport {
  hostname: string
  cpu: string
  memory: string
  os: string
  arch: string
  interfaces: string[]
}

export interface PerformanceProfile {
  category: 'low' | 'standard' | 'high' | 'workstation'
  cpu_threads: number
  memory_gb: number
}

export interface SystemCapability {
  id: string
  name: string
  description: string
  available: boolean
}

export type SafetyPolicy = 'standard' | 'tactical'

export interface OnboardingConfig {
  companionName: string
  storagePath: string
  logsPath: string
  modelPath: string
  safetyPolicy: SafetyPolicy
  engineProfile: string
}

export type OnboardingStep =
  | 'welcome'
  | 'privacy'
  | 'system-check'
  | 'snapshot'
  | 'ai-setup'
  | 'finished'

export const ONBOARDING_STEPS: OnboardingStep[] = [
  'welcome',
  'privacy',
  'system-check',
  'snapshot',
  'ai-setup',
  'finished',
]

// ── AI Onboarding Types ──

export interface ModelOption {
  name: string
  label: string
  size_gb: number
}

export interface AISetupRecommendation {
  can_run_qwythos: boolean
  recommended_model: string
  recommended_label: string
  qwythos_exists: boolean
  pull_required: boolean
  system_ram_gb: number
  system_cpu_threads: number
  timestamp: string
  fallback_models?: ModelOption[]
}

export interface OllamaProgress {
  status: string
  percent: number
  total: number
  completed: number
}
