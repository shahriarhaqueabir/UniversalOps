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
