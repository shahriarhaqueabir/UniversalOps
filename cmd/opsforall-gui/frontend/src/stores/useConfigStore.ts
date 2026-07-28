import { create } from 'zustand'
import Ajv from 'ajv'
import { useSettingsStore } from './useSettingsStore'
import schema from '@/lib/ConfigSchema.json'

// ── Singleton AJV instance ──

const ajv = new Ajv()
const validate = ajv.compile(schema)

// ── Standalone utility: pruneDefaults ──

/**
 * pruneDefaults — strips keys whose values match the schema-defined default.
 *
 * Pass the result to the deploy/persist layer so only non-default overrides
 * are written, keeping the stored config lean and the diff clean.
 *
 * @example
 * ```ts
 * const pruned = pruneDefaults(schema, { refreshInterval: 5000, pingCount: 8 })
 * // => { pingCount: 8 }  (refreshInterval matches default, dropped)
 * ```
 */
export function pruneDefaults(
  configSchema: Record<string, any>,
  config: Record<string, any>,
): Record<string, any> {
  if (!configSchema?.properties) return { ...config }

  const result: Record<string, any> = {}
  for (const [key, value] of Object.entries(config)) {
    const propDef = configSchema.properties[key]
    if (propDef && 'default' in propDef && value === propDef.default) {
      continue // Skip — matches default
    }
    result[key] = value
  }
  return result
}

// ── Helpers ──

/**
 * validateChange — validates a single key-value pair against the schema.
 * Returns null on success, or an error message string on failure.
 */
function validateChange(key: string, value: unknown): string | null {
  const valid = validate({ [key]: value })
  if (!valid && validate.errors) {
    const err = validate.errors.find(
      (e) => e.instancePath === `/${key}` || e.params?.additionalProperty === key,
    )
    if (err) return err.message || 'Invalid value'
  }
  return null
}

// ── Types ──

interface ConfigState {
  stagedChanges: Map<string, any>
  validationErrors: Record<string, string>
  stageChange: (key: string, value: any) => boolean
  stageBatch: (changes: Record<string, any>) => Record<string, string>
  discardAll: () => void
  clearValidationError: (key: string) => void
  getOriginalValue: (key: string) => any
  getRiskLevel: (key: string, value: any) => StagedRisk
}

/**
 * useConfigStore — manages the "Shadow State" for configuration changes.
 * Allows users to stage modifications and review them before deployment.
 * Includes AJV-backed schema validation and pruneDefaults integration.
 */
export interface StagedRisk {
  level: 'low' | 'med' | 'high'
  message: string
}

export const useConfigStore = create<ConfigState>((set, get) => ({
  stagedChanges: new Map(),
  validationErrors: {},

  stageChange: (key, value) => {
    // Validate before staging
    const error = validateChange(key, value)
    if (error) {
      set((s) => ({
        validationErrors: { ...s.validationErrors, [key]: error },
      }))
      return false
    }

    const next = new Map(get().stagedChanges)
    next.set(key, value)

    // Clear any prior validation error for this key
    const { [key]: _omit, ...rest } = get().validationErrors
    set({ stagedChanges: next, validationErrors: rest })
    return true
  },

  stageBatch: (changes) => {
    const errors: Record<string, string> = {}
    const next = new Map(get().stagedChanges)

    for (const [key, value] of Object.entries(changes)) {
      const error = validateChange(key, value)
      if (error) {
        errors[key] = error
      } else {
        next.set(key, value)
      }
    }

    set((s) => ({
      stagedChanges: next,
      validationErrors: { ...s.validationErrors, ...errors },
    }))
    return errors
  },

  discardAll: () => {
    set({ stagedChanges: new Map(), validationErrors: {} })
  },

  clearValidationError: (key) => {
    set((s) => {
      const { [key]: _omit, ...rest } = s.validationErrors
      return { validationErrors: rest }
    })
  },

  getOriginalValue: (key) => {
    const settings = useSettingsStore.getState()
    return (settings as any)[key]
  },

  getRiskLevel: (key: string, value: any): StagedRisk => {
    switch (key) {
      case 'refreshInterval':
        if (value < 1000) return { level: 'high', message: 'Sub-second refresh will significantly increase CPU & Disk I/O load.' }
        if (value < 3000) return { level: 'med', message: 'Frequent updates may cause minor UI jitter on lower-end hardware.' }
        return { level: 'low', message: 'Standard monitoring frequency.' }
      case 'pingCount':
        if (value > 15) return { level: 'med', message: 'High ping counts may be flagged as suspicious by some firewalls.' }
        return { level: 'low', message: 'Normal network diagnostic load.' }
      case 'dnsTimeout':
        if (value > 8000) return { level: 'med', message: 'High timeouts may slow down NetOps page responsiveness.' }
        return { level: 'low', message: 'Standard lookup timeout.' }
      case 'modelfile':
        return { level: 'med', message: 'Rebuilding the neural core may take 10-30 seconds depending on hardware.' }
      default:
        return { level: 'low', message: 'Safe configuration change.' }
    }
  },
}))
