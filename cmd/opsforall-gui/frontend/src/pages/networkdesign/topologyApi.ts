import type {
  TopologyDevice,
  TopologyConnection,
  TopologyHealth,
  TopologyNode,
  TopologyEdge,
  DuplicateIPEntry,
} from '@/types'

// ── Convert existing canvas types to analysis types ──

function devicesToNodes(devices: TopologyDevice[]): TopologyNode[] {
  return devices.map((d) => ({
    id: d.id,
    type: d.type,
    label: d.label,
    ip: d.ip ?? '',
    mac: d.mac ?? '',
    vendor: '',
    notes: d.notes ?? '',
    vlan: '',
    online: d.status !== 'critical',
    props: {},
  }))
}

function connectionsToEdges(connections: TopologyConnection[]): TopologyEdge[] {
  return connections.map((c) => ({
    id: c.id,
    source: c.sourceId,
    target: c.targetId,
    label: c.label ?? '',
    type: c.type,
    bandwidth: c.label ?? '',
    status: 'active',
  }))
}

// ── Compute topology health from canvas data (client-side analysis) ──

export function analyzeTopology(
  devices: TopologyDevice[],
  connections: TopologyConnection[],
): TopologyHealth {
  const nodes = devicesToNodes(devices)
  const edges = connectionsToEdges(connections)

  // Find orphan nodes (no connections)
  const connectedIds = new Set<string>()
  edges.forEach((e) => {
    connectedIds.add(e.source)
    connectedIds.add(e.target)
  })
  const orphanNodes = nodes
    .filter((n) => !connectedIds.has(n.id))
    .map((n) => n.label)

  // Find duplicate IPs
  const ipMap = new Map<string, string[]>()
  nodes.forEach((n) => {
    if (n.ip) {
      const existing = ipMap.get(n.ip) ?? []
      existing.push(n.label)
      ipMap.set(n.ip, existing)
    }
  })
  const duplicateIPs: DuplicateIPEntry[] = []
  ipMap.forEach((nodeNames, ip) => {
    if (nodeNames.length > 1) {
      duplicateIPs.push({ ip, nodes: nodeNames })
    }
  })

  // Find subnet errors (devices on different subnets that are directly connected)
  const subnetErrors: string[] = []
  edges.forEach((e) => {
    const src = devices.find((d) => d.id === e.source)
    const tgt = devices.find((d) => d.id === e.target)
    if (src && tgt && src.subnet && tgt.subnet && src.subnet !== tgt.subnet) {
      subnetErrors.push(`${src.label} (${src.subnet}) ↔ ${tgt.label} (${tgt.subnet})`)
    }
  })

  // Missing labels on connections
  const missingLabels = edges.filter((e) => !e.label).length

  // Generate suggestions
  const suggestions: string[] = []
  if (orphanNodes.length > 0) {
    suggestions.push(`Connect orphan devices: ${orphanNodes.join(', ')}`)
  }
  if (duplicateIPs.length > 0) {
    suggestions.push('Resolve duplicate IP addresses to prevent conflicts')
  }
  if (subnetErrors.length > 0) {
    suggestions.push('Review cross-subnet connections for routing issues')
  }
  if (missingLabels > 0) {
    suggestions.push('Label all connections with bandwidth info')
  }
  nodes.forEach((n) => {
    if (!n.ip) {
      suggestions.push(`Assign IP to "${n.label}"`)
    }
  })
  if (devices.length > 0 && connections.length === 0) {
    suggestions.push('No connections exist — link devices together')
  }
  if (suggestions.length === 0) {
    suggestions.push('Topology looks healthy — no issues found')
  }

  return {
    totalNodes: nodes.length,
    totalEdges: edges.length,
    brokenLinks: 0, // Would need live checks from backend
    missingLabels,
    orphanNodes,
    duplicateIPs,
    subnetErrors,
    suggestions,
  }
}

// ── Group devices by type for inventory ──

export function groupByType(devices: TopologyDevice[]) {
  const groups = new Map<string, TopologyDevice[]>()
  devices.forEach((d) => {
    const existing = groups.get(d.type) ?? []
    existing.push(d)
    groups.set(d.type, existing)
  })
  return Array.from(groups.entries()).map(([type, devs]) => ({
    type,
    count: devs.length,
    devices: devs,
  }))
}

// ── Documentation gaps ──

export interface DocGap {
  category: string
  items: string[]
}

export function findDocumentationGaps(
  devices: TopologyDevice[],
  connections: TopologyConnection[],
): DocGap[] {
  const gaps: DocGap[] = []

  const missingIP = devices.filter((d) => !d.ip).map((d) => d.label)
  if (missingIP.length > 0) {
    gaps.push({ category: 'Missing IP Addresses', items: missingIP })
  }

  const missingNotes = devices.filter((d) => !d.notes).map((d) => d.label)
  if (missingNotes.length > 0) {
    gaps.push({ category: 'Missing Descriptions', items: missingNotes })
  }

  const missingConnLabels = connections.filter((c) => !c.label).map((c) => {
    const src = devices.find((d) => d.id === c.sourceId)
    const tgt = devices.find((d) => d.id === c.targetId)
    return `${src?.label ?? '?'} ↔ ${tgt?.label ?? '?'}`
  })
  if (missingConnLabels.length > 0) {
    gaps.push({ category: 'Connections Missing Labels', items: missingConnLabels })
  }

  return gaps
}

// ── Compute health score (0–100) ──

export function computeHealthScore(health: TopologyHealth): number {
  if (health.totalNodes === 0) return 0
  let score = 100
  score -= health.brokenLinks * 15
  score -= health.missingLabels * 5
  score -= health.orphanNodes.length * 10
  score -= health.duplicateIPs.length * 20
  score -= health.subnetErrors.length * 10
  return Math.max(0, Math.min(100, score))
}

// ── Backend calls via Wails runtime ──

function getGo(): Record<string, Record<string, unknown>> | undefined {
  const w = window as { go?: { app?: Record<string, Record<string, unknown>> } }
  return w.go?.app
}

async function wailsCall(method: string, ...args: unknown[]): Promise<unknown> {
  const go = getGo()
  if (!go) return null
  const parts = method.split('.')
  let target: unknown = go
  for (const part of parts) {
    if (!target) break
    target = (target as Record<string, unknown>)[part]
  }
  if (typeof target === 'function') {
    return await (target as (...args: unknown[]) => Promise<unknown>)(...args)
  }
  return null
}

export async function discoverLocalNetwork(): Promise<TopologyDevice[]> {
  try {
    const nodes = await wailsCall('NetDesign.DiscoverLocalNetwork')
    if (!Array.isArray(nodes)) return []
    return nodes.map((n: Record<string, unknown>) => ({
      id: (n.id as string) || `discovered-${Math.random().toString(36).slice(2, 8)}`,
      type: (n.type as string) || 'client',
      label: (n.label as string) || (n.ip as string) || 'Unknown',
      ip: (n.ip as string) || '',
      mac: (n.mac as string) || '',
      status: (n as { online?: boolean }).online ? 'healthy' : 'critical',
      notes: (n.vendor as string) ? `Vendor: ${n.vendor}` : '',
      subnet: '',
      x: 0,
      y: 0,
    }))
  } catch /* ignore */ {
    return []
  }
}

export async function saveTopologyToFile(
  devices: TopologyDevice[],
  connections: TopologyConnection[],
): Promise<boolean> {
  try {
    await wailsCall('NetDesign.SetTopology', JSON.stringify(devices), JSON.stringify(connections))
    await wailsCall('NetDesign.SaveTopology')
    return true
  } catch /* ignore */ {
    return false
  }
}

export async function loadTopologyFromFile(): Promise<{ devices: TopologyDevice[]; connections: TopologyConnection[] } | null> {
  try {
    const result = await wailsCall('NetDesign.LoadTopology') as Record<string, unknown> | null
    if (!result) return null
    const devices = typeof result.nodes === 'string' ? JSON.parse(result.nodes) : ((result.nodes as TopologyDevice[]) || [])
    const connections = typeof result.edges === 'string' ? JSON.parse(result.edges) : ((result.edges as TopologyConnection[]) || [])
    return { devices, connections }
  } catch /* ignore */ {
    return null
  }
}
