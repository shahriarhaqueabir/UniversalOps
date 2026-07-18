import { create } from 'zustand'
import type { AlertInfo } from '@/types'

// ── Settings Store ──

interface SettingsState {
  refreshInterval: number
  pingCount: number
  dnsTimeout: number
  companionName: string
  setRefreshInterval: (val: number) => void
  setPingCount: (val: number) => void
  setDnsTimeout: (val: number) => void
  setCompanionName: (name: string) => void
}

function loadSetting<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    return raw !== null ? (JSON.parse(raw) as T) : fallback
  } catch { return fallback }
}

function saveSetting<T>(key: string, value: T): void {
  try { localStorage.setItem(key, JSON.stringify(value)) } catch { /* ignore */ }
}

export const useSettingsStore = create<SettingsState>((set) => ({
  refreshInterval: loadSetting('opsforall_refreshInterval', 5000),
  pingCount: loadSetting('opsforall_pingCount', 4),
  dnsTimeout: loadSetting('opsforall_dnsTimeout', 2000),
  companionName: loadSetting('opsforall_companionName', 'Hawk'),

  setRefreshInterval: (val) => {
    saveSetting('opsforall_refreshInterval', val)
    set({ refreshInterval: val })
  },
  setPingCount: (val) => {
    saveSetting('opsforall_pingCount', val)
    set({ pingCount: val })
  },
  setDnsTimeout: (val) => {
    saveSetting('opsforall_dnsTimeout', val)
    set({ dnsTimeout: val })
  },
  setCompanionName: (name) => {
    saveSetting('opsforall_companionName', name)
    set({ companionName: name })
  },
}))

// ── Alert Store ──

interface AlertState {
  alerts: AlertInfo[]
  alertCount: number
  setAlerts: (alerts: AlertInfo[]) => void
  addAlert: (alert: AlertInfo) => void
  clearAlerts: () => void
}

export const useAlertStore = create<AlertState>((set) => ({
  alerts: [],
  alertCount: 0,
  setAlerts: (alerts) => set({ alerts, alertCount: alerts.length }),
  addAlert: (alert) => set((state) => ({
    alerts: [alert, ...state.alerts].slice(0, 100),
    alertCount: state.alertCount + 1,
  })),
  clearAlerts: () => set({ alerts: [], alertCount: 0 }),
}))

// ── Theme Store ──

type Theme = 'dark' | 'light'

interface ThemeState {
  theme: Theme
  toggle: () => void
  setTheme: (t: Theme) => void
}

const THEME_KEY = 'opsforall-theme'

export const useThemeStore = create<ThemeState>((set) => ({
  theme: (() => {
    try {
      const stored = typeof window !== 'undefined' && typeof localStorage !== 'undefined'
        ? localStorage.getItem(THEME_KEY)
        : null
      if (stored === 'dark' || stored === 'light') return stored as Theme
    } catch { /* localStorage unavailable */ }
    return 'dark' as Theme
  })(),

  toggle: () => set((state) => {
    const next = state.theme === 'dark' ? 'light' : 'dark'
    localStorage.setItem(THEME_KEY, next)
    document.documentElement.setAttribute('data-theme', next)
    return { theme: next }
  }),

  setTheme: (t) => {
    localStorage.setItem(THEME_KEY, t)
    document.documentElement.setAttribute('data-theme', t)
    set({ theme: t })
  },
}))
