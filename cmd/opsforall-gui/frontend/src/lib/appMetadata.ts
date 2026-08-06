import type { Page } from '@/stores'

export const APP_NAME = 'Universal-Ops'
export const APP_DESCRIPTION = 'High-performance native operations platform for systems, network, security, and AI operations.'
const THEME_COLOR: Record<'dark' | 'light', string> = {
  dark: '#080a0f',
  light: '#f5f6fa',
}

export interface PageMetadata {
  title: string
  description: string
  keywords: string[]
}

export const PAGE_METADATA: Record<Page, PageMetadata> = {
  dashboard: {
    title: 'Dashboard',
    description: 'Operational overview with health score, active alerts, and cross-pillar status.',
    keywords: ['overview', 'health score', 'active alerts', 'operations'],
  },
  sysops: {
    title: 'System Operations',
    description: 'CPU, memory, disk, process, and hardware telemetry for workstation health.',
    keywords: ['cpu', 'memory', 'disk', 'processes', 'hardware', 'telemetry'],
  },
  workflows: {
    title: 'Operational Workflows',
    description: 'Browse reusable operational workflows and guided action playbooks.',
    keywords: ['workflows', 'playbooks', 'automation', 'operations'],
  },
  'network-designer': {
    title: 'Network Designer',
    description: 'Visually map and edit network topology, links, and device relationships.',
    keywords: ['network topology', 'diagram', 'design', 'links'],
  },
  netops: {
    title: 'Network Operations',
    description: 'Ping, DNS, ports, traceroute, routing, and interface diagnostics.',
    keywords: ['ping', 'dns', 'traceroute', 'ports', 'network diagnostics'],
  },
  secops: {
    title: 'Security Operations',
    description: 'Firewall, endpoints, users, audit, and incident response monitoring.',
    keywords: ['security', 'audit', 'firewall', 'endpoints', 'incidents'],
  },
  devops: {
    title: 'DevOps',
    description: 'Services, logs, files, packages, and local automation controls.',
    keywords: ['devops', 'services', 'logs', 'packages', 'automation'],
  },
  aiops: {
    title: 'AI Operations',
    description: 'Local AI companion, anomalies, insights, and model management.',
    keywords: ['ai', 'ollama', 'anomalies', 'insights', 'models'],
  },
  reports: {
    title: 'Reports Center',
    description: 'Generate, search, and review operational reports and audit outputs.',
    keywords: ['reports', 'audit', 'search', 'export'],
  },
  alerts: {
    title: 'Alerts Dashboard',
    description: 'Review active alerts, resolve incidents, and manage alerting rules.',
    keywords: ['alerts', 'incident response', 'alert rules', 'monitoring'],
  },
  logs: {
    title: 'Log Viewer',
    description: 'Search live and historical logs with filters, timestamps, and audit context.',
    keywords: ['logs', 'search', 'audit trail', 'log viewer'],
  },
  settings: {
    title: 'Settings',
    description: 'Configure themes, pipelines, thresholds, storage, and system maintenance.',
    keywords: ['settings', 'preferences', 'configuration', 'storage'],
  },
}

function setMetaTag(attr: 'name' | 'property', key: string, content: string) {
  const selector = `meta[${attr}="${key}"]`
  let el = document.head.querySelector(selector) as HTMLMetaElement | null
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    document.head.appendChild(el)
  }
  el.setAttribute('content', content)
}

export function buildDocumentMetadata(page: Page, activeAlertCount = 0) {
  const meta = PAGE_METADATA[page]
  const title = activeAlertCount > 0 && page === 'alerts'
    ? `${meta.title} (${activeAlertCount})`
    : meta.title

  const description = activeAlertCount > 0 && page === 'alerts'
    ? `${meta.description} ${activeAlertCount} active alert${activeAlertCount === 1 ? '' : 's'} require attention.`
    : meta.description

  const keywords = [
    APP_NAME,
    meta.title.toLowerCase(),
    ...meta.keywords,
    page,
    page.replace('-', ' '),
  ].join(', ')

  return { title: `${title} · ${APP_NAME}`, description, keywords }
}

export function applyDocumentMetadata(page: Page, activeAlertCount = 0, theme: 'dark' | 'light' = 'dark') {
  if (typeof document === 'undefined') return

  const metadata = buildDocumentMetadata(page, activeAlertCount)
  document.title = metadata.title
  document.documentElement.lang = 'en'
  setMetaTag('name', 'description', metadata.description)
  setMetaTag('name', 'keywords', metadata.keywords)
  setMetaTag('name', 'application-name', APP_NAME)
  setMetaTag('name', 'apple-mobile-web-app-title', APP_NAME)
  setMetaTag('name', 'theme-color', THEME_COLOR[theme])
  setMetaTag('name', 'color-scheme', 'dark light')
  setMetaTag('name', 'robots', 'index,follow')
  setMetaTag('property', 'og:title', metadata.title)
  setMetaTag('property', 'og:description', metadata.description)
  setMetaTag('property', 'og:site_name', APP_NAME)
  setMetaTag('property', 'og:type', 'website')
}
