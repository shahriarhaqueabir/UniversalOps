import { lazy, type ComponentType, type LazyExoticComponent } from 'react'
import type { Page } from '../stores'

type PageLoader = () => Promise<{ default: ComponentType<any> }>

const loadDashboard: PageLoader = () => import('../pages/Dashboard').then((m) => ({ default: m.Dashboard }))
const loadSysOps: PageLoader = () => import('../pages/SysOps').then((m) => ({ default: m.SysOps }))
const loadWorkflowCenter: PageLoader = () => import('../pages/WorkflowCenter').then((m) => ({ default: m.WorkflowCenter }))
const loadNetOps: PageLoader = () => import('../pages/NetOps').then((m) => ({ default: m.NetOps }))
const loadNetworkDesigner: PageLoader = () => import('../pages/NetworkDesigner').then((m) => ({ default: m.NetworkDesigner }))
const loadSecOps: PageLoader = () => import('../pages/SecOps').then((m) => ({ default: m.SecOps }))
const loadDevOps: PageLoader = () => import('../pages/DevOps').then((m) => ({ default: m.DevOps }))
const loadAIOps: PageLoader = () => import('../pages/AIOps').then((m) => ({ default: m.AIOps }))
const loadReportsCenter: PageLoader = () => import('../pages/ReportsCenter').then((m) => ({ default: m.ReportsCenter }))
const loadAlertsDashboard: PageLoader = () => import('../pages/AlertsDashboard').then((m) => ({ default: m.AlertsDashboard }))
const loadLogs: PageLoader = () => import('../pages/Logs').then((m) => ({ default: m.Logs }))
const loadSettings: PageLoader = () => import('../pages/Settings').then((m) => ({ default: m.Settings }))

const pageLoaders = {
  dashboard: loadDashboard,
  sysops: loadSysOps,
  workflows: loadWorkflowCenter,
  netops: loadNetOps,
  'network-designer': loadNetworkDesigner,
  secops: loadSecOps,
  devops: loadDevOps,
  aiops: loadAIOps,
  reports: loadReportsCenter,
  alerts: loadAlertsDashboard,
  logs: loadLogs,
  settings: loadSettings,
} as const satisfies Record<Page, PageLoader>

export const lazyPages = {
  dashboard: lazy(loadDashboard),
  sysops: lazy(loadSysOps),
  workflows: lazy(loadWorkflowCenter),
  netops: lazy(loadNetOps),
  'network-designer': lazy(loadNetworkDesigner),
  secops: lazy(loadSecOps),
  devops: lazy(loadDevOps),
  aiops: lazy(loadAIOps),
  reports: lazy(loadReportsCenter),
  alerts: lazy(loadAlertsDashboard),
  logs: lazy(loadLogs),
  settings: lazy(loadSettings),
} as const satisfies Record<Page, LazyExoticComponent<ComponentType<any>>>

const preloadHints: Record<Page, Page[]> = {
  dashboard: ['sysops', 'alerts', 'reports', 'logs'],
  sysops: ['dashboard', 'netops', 'devops'],
  workflows: ['dashboard', 'reports', 'aiops'],
  netops: ['dashboard', 'sysops', 'logs'],
  'network-designer': ['netops', 'dashboard'],
  secops: ['dashboard', 'alerts', 'reports'],
  devops: ['dashboard', 'logs', 'settings'],
  aiops: ['dashboard', 'settings', 'reports'],
  reports: ['dashboard', 'alerts', 'logs'],
  alerts: ['dashboard', 'logs', 'reports'],
  logs: ['dashboard', 'alerts', 'reports'],
  settings: ['dashboard', 'devops', 'aiops'],
}

const preloadedPages = new Set<Page>()

export function preloadPage(page: Page) {
  if (preloadedPages.has(page)) return
  preloadedPages.add(page)
  void pageLoaders[page]()
}

export function preloadSuggestedPages(page: Page) {
  for (const nextPage of preloadHints[page] ?? []) {
    preloadPage(nextPage)
  }
}

export function getSuggestedPages(page: Page): Page[] {
  return preloadHints[page] ?? []
}
