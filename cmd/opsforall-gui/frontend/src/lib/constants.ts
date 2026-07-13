export const opsLayers = [
  { id: 'sysops', icon: 'Monitor', title: 'System', description: 'CPU, memory, disk, processes', color: 'var(--color-accent)' },
  { id: 'netops', icon: 'Network', title: 'Network', description: 'Ping, DNS, ports, traceroute', color: 'var(--color-success)' },
  { id: 'secops', icon: 'Shield', title: 'Security', description: 'Firewall, users, defender', color: 'var(--color-danger)' },
  { id: 'devops', icon: 'Terminal', title: 'DevOps', description: 'Commands, logs, files, services', color: 'var(--color-warning)' },
  { id: 'aiops', icon: 'Brain', title: 'AI Ops', description: 'Ollama, chat, anomalies', color: 'var(--color-accent-2)' },
] as const
