import type { DashboardData, CPUInfo, MemoryInfo, DiskInfo, ProcessInfo, SystemInfo, PingEntry, PingStats, DNSResult, PortResult, TraceHop, TraceResult, ConnectionInfo, InterfaceInfo, FirewallRule, UserInfo, ListeningPort, DefenderStatus, ScheduledTask, SecurityEvent } from '@/types'

/** Generate a mock time-series array */
export function generateMockTimeSeries(points: number, min: number, max: number) {
  const data = []
  const now = Date.now()
  for (let i = 0; i < points; i++) {
    data.push({
      time: new Date(now - (points - i) * 3000).toLocaleTimeString(),
      value: min + Math.random() * (max - min),
    })
  }
  return data
}

/** Generate a flat number array (for history) */
export function generateMockHistory(points: number, min: number, max: number): number[] {
  return Array.from({ length: points }, () => min + Math.random() * (max - min))
}

/** Mock DashboardData snapshot */
export function mockDashboardData(): DashboardData {
  return {
    cpu: {
      value: 35 + Math.random() * 40,
      unit: '%',
      history: generateMockHistory(20, 10, 90),
      forecast: [],
      trend: Math.random() > 0.6 ? 'rising' : 'stable',
    },
    memory: {
      value: 40 + Math.random() * 30,
      unit: '%',
      history: generateMockHistory(20, 30, 80),
      forecast: [],
      trend: 'stable',
    },
    disk: {
      value: 50 + Math.random() * 20,
      unit: '%',
      history: generateMockHistory(20, 40, 80),
      forecast: [],
      trend: 'stable',
    },
    network: {
      rx_rate: 1.2 + Math.random() * 8,
      tx_rate: 0.5 + Math.random() * 4,
      unit: 'MB/s',
    },
    processes: 180 + Math.floor(Math.random() * 60),
    connections: 42 + Math.floor(Math.random() * 20),
    alerts: Math.floor(Math.random() * 5),
    uptime: `${Math.floor(Math.random() * 30) + 1}d ${Math.floor(Math.random() * 24)}h ${Math.floor(Math.random() * 60)}m`,
  }
}

/** Mock CPU info */
export function mockCPUInfo(): CPUInfo {
  return {
    percent: 35 + Math.random() * 40,
    per_cpu: Array.from({ length: 8 }, () => Math.random() * 100),
    model_name: 'AMD Ryzen 7 5800X 8-Core',
    core_count: 8,
    load_avg_1: 2.5 + Math.random() * 3,
    load_avg_5: 2.0 + Math.random() * 2,
    load_avg_15: 1.5 + Math.random() * 1.5,
  }
}

/** Mock memory info */
export function mockMemoryInfo(): MemoryInfo {
  const totalGB = 32
  const usedGB = 12 + Math.random() * 8
  return {
    total_bytes: totalGB * 1024 * 1024 * 1024,
    available_bytes: (totalGB - usedGB) * 1024 * 1024 * 1024,
    used_bytes: usedGB * 1024 * 1024 * 1024,
    used_percent: (usedGB / totalGB) * 100,
    total_gb: totalGB,
    used_gb: usedGB,
    swap_total: 8 * 1024 * 1024 * 1024,
    swap_used: 0.5 * 1024 * 1024 * 1024,
    swap_percent: 6.25,
  }
}

/** Mock disk info */
export function mockDiskInfo(): DiskInfo {
  return {
    partitions: [
      {
        mountpoint: 'C:\\',
        total_bytes: 500 * 1024 * 1024 * 1024,
        free_bytes: 250 * 1024 * 1024 * 1024,
        used_bytes: 250 * 1024 * 1024 * 1024,
        used_percent: 50,
        fs_type: 'NTFS',
        device: '\\Device\\HarddiskVolume1',
      },
      {
        mountpoint: 'D:\\',
        total_bytes: 1000 * 1024 * 1024 * 1024,
        free_bytes: 600 * 1024 * 1024 * 1024,
        used_bytes: 400 * 1024 * 1024 * 1024,
        used_percent: 40,
        fs_type: 'NTFS',
        device: '\\Device\\HarddiskVolume2',
      },
    ],
  }
}

/** Mock process list (top 20) */
export function mockProcesses(): ProcessInfo[] {
  const names = [
    'chrome.exe', 'explorer.exe', 'code.exe', 'node.exe', 'python.exe',
    'discord.exe', 'spotify.exe', 'powershell.exe', 'msedge.exe', 'Teams.exe',
    'slack.exe', 'docker.exe', 'git.exe', 'nvim.exe', 'obsidian.exe',
    'firefox.exe', 'vscode.exe', 'npm.exe', 'go.exe', 'windbg.exe',
    'outlook.exe', 'onedrive.exe', 'zoom.exe', 'vlc.exe', 'notepad++.exe',
  ]
  return names.slice(0, 20).map((name) => ({
    pid: 4000 + Math.floor(Math.random() * 20000),
    name,
    cpu: parseFloat((Math.random() * 20).toFixed(1)),
    memory: parseFloat((50 + Math.random() * 500).toFixed(1)),
    mem_pct: parseFloat((Math.random() * 5).toFixed(2)),
    status: Math.random() > 0.2 ? 'running' : 'sleeping',
    num_fds: Math.floor(Math.random() * 200),
  }))
}

/** Mock system info */
export function mockSystemInfo(): SystemInfo {
  return {
    hostname: 'HAWKWARD-PC',
    os: 'windows',
    platform: 'Microsoft Windows 11 Pro',
    platform_version: '10.0.26100',
    kernel_version: '10.0.26100.1',
    kernel_arch: 'x86_64',
    uptime: `${Math.floor(Math.random() * 14) + 1}d ${Math.floor(Math.random() * 24)}h ${Math.floor(Math.random() * 60)}m`,
    process_count: 180 + Math.floor(Math.random() * 60),
    virtualization: 'Hyper-V Enabled',
  }
}

/** Mock alerts */
export function mockAlerts() {
  const severities = ['critical', 'warning', 'info'] as const
  const metrics = ['CPU', 'Memory', 'Disk', 'Network', 'Processes']
  return Array.from({ length: 3 + Math.floor(Math.random() * 4) }, (_, i) => {
    const severity = severities[Math.floor(Math.random() * severities.length)]
    const metric = metrics[Math.floor(Math.random() * metrics.length)]
    return {
      id: `alert-${i}`,
      level: severity,
      metric,
      message: `${metric} usage threshold crossed`,
      value: 70 + Math.random() * 30,
      threshold: 90,
      timestamp: new Date(Date.now() - Math.random() * 3600000).toISOString(),
      resolved: Math.random() > 0.7,
    }
  })
}

/** Ops layer quick actions */
// ── NetOps Mock Data ──

const commonServices: Record<number, string> = {
  22: 'SSH', 23: 'Telnet', 25: 'SMTP', 53: 'DNS', 80: 'HTTP',
  110: 'POP3', 143: 'IMAP', 443: 'HTTPS', 445: 'SMB', 993: 'IMAPS',
  995: 'POP3S', 1433: 'MSSQL', 3306: 'MySQL', 3389: 'RDP', 5432: 'PostgreSQL',
  6379: 'Redis', 8080: 'HTTP-Alt', 8443: 'HTTPS-Alt', 27017: 'MongoDB', 50070: 'HDFS',
}

const states = ['LISTENING', 'ESTABLISHED', 'TIME_WAIT', 'CLOSE_WAIT'] as const

export function mockPingEntry(seq: number, ip: string): PingEntry {
  const timeout = Math.random() < 0.05
  return {
    seq,
    ip,
    rtt_ms: timeout ? null : Math.round((10 + Math.random() * 90) * 10) / 10,
    ttl: timeout ? null : Math.floor(55 + Math.random() * 10),
    status: timeout ? 'timeout' : 'success',
  }
}

export function mockPingStats(entries: PingEntry[]): PingStats {
  const successful = entries.filter(e => e.status === 'success')
  const rtts = successful.map(e => e.rtt_ms!).filter(Boolean)
  return {
    target: '8.8.8.8',
    sent: entries.length,
    received: successful.length,
    lost: entries.length - successful.length,
    lost_pct: entries.length > 0 ? ((entries.length - successful.length) / entries.length) * 100 : 0,
    min_ms: rtts.length > 0 ? Math.min(...rtts) : null,
    max_ms: rtts.length > 0 ? Math.max(...rtts) : null,
    avg_ms: rtts.length > 0 ? Math.round(rtts.reduce((a, b) => a + b, 0) / rtts.length * 10) / 10 : null,
  }
}

export function mockDNSResult(hostname: string): DNSResult {
  return {
    hostname,
    a: ['142.250.80.46', '142.250.80.78', '142.250.80.110'],
    aaaa: ['2607:f8b0:4004:c08::8a', '2607:f8b0:4004:c08::8b'],
    mx: ['10 aspmx.l.google.com', '20 alt1.aspmx.l.google.com', '30 alt2.aspmx.l.google.com', '40 alt3.aspmx.l.google.com'],
    ns: ['ns1.google.com', 'ns2.google.com', 'ns3.google.com', 'ns4.google.com'],
    cname: '',
    txt: ['v=spf1 include:_spf.google.com ~all', 'docusign=05958488-4752-4ef2-95eb-aa7ba8a3bd0e'],
  }
}

export function mockPortScan(_target: string, ports: number[]): PortResult[] {
  return ports.map(port => ({
    port,
    open: Math.random() > 0.6,
    service: commonServices[port] || 'unknown',
  }))
}

export function mockTraceResult(target: string): TraceResult {
  const ips = ['192.168.1.1', '10.0.0.1', '72.14.237.300', '216.239.43.300', '142.250.80.46']
  const hosts = ['router.local', 'isp-gw.example.com', 'core1.nyc.google.com', 'edge1.lga.google.com', target]
  const hops: TraceHop[] = ips.map((ip, i) => {
    const timed = i === 2 && Math.random() < 0.3
    return {
      number: i + 1,
      host: hosts[i],
      ip,
      rtts_ms: timed ? [] : [10 + Math.random() * 40, 10 + Math.random() * 40, 10 + Math.random() * 40].map(v => Math.round(v)),
      timed,
      avg_rtt: timed ? null : Math.round((15 + Math.random() * 30) * 10) / 10,
    }
  })
  return { target, hops }
}

export function mockConnections(): ConnectionInfo[] {
  const procs = ['chrome.exe', 'firefox.exe', 'msedge.exe', 'Slack.exe', 'Discord.exe', 'Spotify.exe', 'Teams.exe', 'outlook.exe']
  const remoteAddrs = ['142.250.80.46', '151.101.1.140', '52.84.120.10', '13.107.42.16', '204.79.197.200', '20.53.120.50', '52.112.120.10', '40.97.120.10']
  return Array.from({ length: 24 }, (_, i) => {
    const proc = procs[i % procs.length]
    const state = states[i % states.length]
    const port = 40000 + Math.floor(Math.random() * 20000)
    const remotePort = [80, 443, 8080, 8443][Math.floor(Math.random() * 4)]
    return {
      proto: Math.random() > 0.15 ? 'TCP' : 'UDP',
      local_addr: '192.168.1.100',
      local_port: port,
      remote_addr: remoteAddrs[i % remoteAddrs.length],
      remote_port: remotePort,
      state,
      pid: 4000 + Math.floor(Math.random() * 15000),
      process_name: proc,
    }
  })
}

export function mockInterfaces(): InterfaceInfo[] {
  return [
    {
      name: 'Ethernet0',
      mac: '00:1A:2B:3C:4D:5E',
      ips: ['192.168.1.100/24', 'fe80::21a:2bff:fe3c:4d5e'],
      is_up: true,
      speed: '1 Gbps',
      mtu: 1500,
      flags: 'UP,BROADCAST,RUNNING,MULTICAST',
      rx_bytes: 1250000000,
      tx_bytes: 850000000,
      rx_rate_bps: 2.5 * 1024 * 1024,
      tx_rate_bps: 1.2 * 1024 * 1024,
      rx_history: generateMockHistory(20, 0.5, 8),
      tx_history: generateMockHistory(20, 0.2, 4),
    },
    {
      name: 'WiFi',
      mac: '00:1A:2B:3C:4D:5F',
      ips: ['192.168.1.101/24', 'fe80::21a:2bff:fe3c:4d5f'],
      is_up: true,
      speed: '866 Mbps',
      mtu: 1500,
      flags: 'UP,BROADCAST,RUNNING,MULTICAST',
      rx_bytes: 450000000,
      tx_bytes: 320000000,
      rx_rate_bps: 1.1 * 1024 * 1024,
      tx_rate_bps: 0.6 * 1024 * 1024,
      rx_history: generateMockHistory(20, 0.1, 3),
      tx_history: generateMockHistory(20, 0.1, 2),
    },
    {
      name: 'Loopback',
      mac: '00:00:00:00:00:00',
      ips: ['127.0.0.1/8', '::1/128'],
      is_up: true,
      speed: 'N/A',
      mtu: 65536,
      flags: 'UP,LOOPBACK,RUNNING',
      rx_bytes: 12000,
      tx_bytes: 12000,
      rx_rate_bps: 0,
      tx_rate_bps: 0,
      rx_history: generateMockHistory(20, 0, 0.1),
      tx_history: generateMockHistory(20, 0, 0.1),
    },
    {
      name: 'Bluetooth',
      mac: '00:1A:2B:3C:4D:60',
      ips: [],
      is_up: false,
      speed: '3 Mbps',
      mtu: 1024,
      flags: 'DOWN',
      rx_bytes: 0,
      tx_bytes: 0,
      rx_rate_bps: 0,
      tx_rate_bps: 0,
      rx_history: generateMockHistory(20, 0, 0),
      tx_history: generateMockHistory(20, 0, 0),
    },
  ]
}

// ── SecOps Mock Data ──

export function mockFirewallRules(): FirewallRule[] {
  return [
    { name: 'Core Networking - DNS (UDP Out)', direction: 'Outbound', action: 'Allow', protocol: 'UDP', local_port: '*', remote_port: '53', remote_ip: '*', profile: 'Domain,Private,Public', enabled: true },
    { name: 'Core Networking - DNS (TCP Out)', direction: 'Outbound', action: 'Allow', protocol: 'TCP', local_port: '*', remote_port: '53', remote_ip: '*', profile: 'Domain,Private,Public', enabled: true },
    { name: 'World Wide Web (HTTP Out)', direction: 'Outbound', action: 'Allow', protocol: 'TCP', local_port: '*', remote_port: '80', remote_ip: '*', profile: 'Domain,Private,Public', enabled: true },
    { name: 'World Wide Web (HTTPS Out)', direction: 'Outbound', action: 'Allow', protocol: 'TCP', local_port: '*', remote_port: '443', remote_ip: '*', profile: 'Domain,Private,Public', enabled: true },
    { name: 'File and Printer Sharing (SMB In)', direction: 'Inbound', action: 'Block', protocol: 'TCP', local_port: '445', remote_port: '*', remote_ip: '*', profile: 'Public', enabled: true },
    { name: 'Remote Desktop (RDP In)', direction: 'Inbound', action: 'Block', protocol: 'TCP', local_port: '3389', remote_port: '*', remote_ip: '*', profile: 'Public', enabled: true },
    { name: 'SSH (Port 22 In)', direction: 'Inbound', action: 'Block', protocol: 'TCP', local_port: '22', remote_port: '*', remote_ip: '*', profile: 'Public', enabled: true },
    { name: 'Ping (ICMP In)', direction: 'Inbound', action: 'Allow', protocol: 'ICMPv4', local_port: '*', remote_port: '*', remote_ip: '*', profile: 'Domain,Private', enabled: true },
    { name: 'Windows Update Out', direction: 'Outbound', action: 'Allow', protocol: 'TCP', local_port: '*', remote_port: '443', remote_ip: '*', profile: 'Domain,Private,Public', enabled: true },
    { name: 'DHCP (UDP Out)', direction: 'Outbound', action: 'Allow', protocol: 'UDP', local_port: '*', remote_port: '67,68', remote_ip: '*', profile: 'Domain,Private,Public', enabled: true },
    { name: 'NetBIOS (UDP In)', direction: 'Inbound', action: 'Block', protocol: 'UDP', local_port: '137,138', remote_port: '*', remote_ip: '*', profile: 'Public', enabled: false },
    { name: 'SQL Server (TCP In)', direction: 'Inbound', action: 'Allow', protocol: 'TCP', local_port: '1433', remote_port: '*', remote_ip: '10.0.0.0/8', profile: 'Domain,Private', enabled: true },
    { name: 'Custom App Port 9000', direction: 'Inbound', action: 'Allow', protocol: 'TCP', local_port: '9000', remote_port: '*', remote_ip: '192.168.1.0/24', profile: 'Private', enabled: true },
    { name: 'Legacy FTP (Port 21 In)', direction: 'Inbound', action: 'Block', protocol: 'TCP', local_port: '21', remote_port: '*', remote_ip: '*', profile: 'Public', enabled: true },
    { name: 'Docker Bridge Traffic', direction: 'Inbound', action: 'Allow', protocol: 'TCP', local_port: '*', remote_port: '*', remote_ip: '172.17.0.0/16', profile: 'Private', enabled: true },
  ]
}

export function mockUsers(): UserInfo[] {
  return [
    { username: 'Administrator', full_name: 'Built-in Admin Account', sid: 'S-1-5-21-...-500', group: 'Administrators', is_admin: true, is_enabled: true },
    { username: 'shahr', full_name: 'Shah Rukh', sid: 'S-1-5-21-...-1001', group: 'Administrators', is_admin: true, is_enabled: true },
    { username: 'jdoe', full_name: 'John Doe', sid: 'S-1-5-21-...-1002', group: 'Users', is_admin: false, is_enabled: true },
    { username: 'asmith', full_name: 'Alice Smith', sid: 'S-1-5-21-...-1003', group: 'Users', is_admin: false, is_enabled: true },
    { username: 'bob', full_name: 'Bob Wilson', sid: 'S-1-5-21-...-1004', group: 'Users', is_admin: false, is_enabled: true },
    { username: 'svc_backup', full_name: 'Backup Service Account', sid: 'S-1-5-21-...-2001', group: 'Backup Operators', is_admin: false, is_enabled: true },
    { username: 'svc_web', full_name: 'Web Service Account', sid: 'S-1-5-21-...-2002', group: 'IIS_IUSRS', is_admin: false, is_enabled: true },
    { username: 'guest', full_name: 'Guest Account', sid: 'S-1-5-21-...-501', group: 'Guests', is_admin: false, is_enabled: false },
  ]
}

export function mockListeningPorts(): ListeningPort[] {
  return [
    { port: 135, protocol: 'TCP', process_name: 'svchost.exe', pid: 1216, state: 'LISTENING' },
    { port: 139, protocol: 'TCP', process_name: 'System', pid: 4, state: 'LISTENING' },
    { port: 445, protocol: 'TCP', process_name: 'System', pid: 4, state: 'LISTENING' },
    { port: 3389, protocol: 'TCP', process_name: 'svchost.exe', pid: 1480, state: 'LISTENING' },
    { port: 5040, protocol: 'TCP', process_name: 'svchost.exe', pid: 1216, state: 'LISTENING' },
    { port: 7680, protocol: 'TCP', process_name: 'wuauserv', pid: 2104, state: 'LISTENING' },
    { port: 9229, protocol: 'TCP', process_name: 'chrome.exe', pid: 8234, state: 'LISTENING' },
    { port: 5000, protocol: 'TCP', process_name: 'python.exe', pid: 6712, state: 'LISTENING' },
    { port: 5432, protocol: 'TCP', process_name: 'postgres.exe', pid: 3401, state: 'LISTENING' },
    { port: 6379, protocol: 'TCP', process_name: 'redis-server.exe', pid: 2904, state: 'LISTENING' },
    { port: 8080, protocol: 'TCP', process_name: 'node.exe', pid: 4512, state: 'LISTENING' },
    { port: 5353, protocol: 'UDP', process_name: 'svchost.exe', pid: 2032, state: 'LISTENING' },
    { port: 5355, protocol: 'UDP', process_name: 'svchost.exe', pid: 2032, state: 'LISTENING' },
    { port: 1900, protocol: 'UDP', process_name: 'svchost.exe', pid: 1480, state: 'LISTENING' },
  ]
}

export function mockDefenderStatus(): DefenderStatus {
  return {
    enabled: true,
    up_to_date: true,
    signature_age: '1 day ago',
    last_scan: '2026-07-06 03:00 AM',
    real_time_protection: true,
    cloud_protection: true,
    am_service_enabled: true,
    antispyware_enabled: true,
    nis_enabled: true,
    quick_scan_age: 0,
    full_scan_age: 5,
  }
}

export function mockScheduledTasks(): ScheduledTask[] {
  return [
    { name: 'GoogleUpdateTaskMachineCore', status: 'Ready', next_run: 'N/A', last_run: '7/7/2026 2:30 AM', author: 'GOOGLE', trigger: 'Daily' },
    { name: 'GoogleUpdateTaskMachineUA', status: 'Ready', next_run: '7/7/2026 3:00 PM', last_run: '7/7/2026 2:30 AM', author: 'GOOGLE', trigger: 'Daily' },
    { name: 'MicrosoftEdgeUpdateTaskMachineCore', status: 'Ready', next_run: 'N/A', last_run: '7/7/2026 2:00 AM', author: 'MICROSOFT', trigger: 'Daily' },
    { name: 'OneDrive Standalone Update Task', status: 'Ready', next_run: '7/8/2026 2:00 AM', last_run: '7/7/2026 2:00 AM', author: 'MICROSOFT', trigger: 'Daily' },
    { name: 'Adobe Acrobat Update Task', status: 'Disabled', next_run: 'N/A', last_run: '6/15/2026 12:00 PM', author: 'ADOBE', trigger: 'Weekly' },
    { name: 'Windows Defender Scheduled Scan', status: 'Ready', next_run: '7/8/2026 3:00 AM', last_run: '7/7/2026 3:00 AM', author: 'MICROSOFT', trigger: 'Daily' },
    { name: 'Windows Defender Verification', status: 'Running', next_run: 'N/A', last_run: '7/7/2026 2:45 AM', author: 'MICROSOFT', trigger: 'OnEvent' },
    { name: 'System Restore Scheduled', status: 'Ready', next_run: '7/14/2026 12:00 AM', last_run: '6/30/2026 12:00 AM', author: 'MICROSOFT', trigger: 'Weekly' },
    { name: 'Disk Cleanup', status: 'Disabled', next_run: 'N/A', last_run: '5/1/2026 12:00 PM', author: 'MICROSOFT', trigger: 'Monthly' },
    { name: 'VS Code Update', status: 'Ready', next_run: '7/8/2026 12:00 PM', last_run: '7/7/2026 10:00 AM', author: 'MICROSOFT', trigger: 'Daily' },
    { name: 'npm cache cleanup', status: 'Disabled', next_run: 'N/A', last_run: '6/1/2026 8:00 AM', author: 'SYSTEM', trigger: 'Weekly' },
    { name: 'Chrome Cleanup Tool', status: 'Ready', next_run: '7/14/2026 4:00 AM', last_run: '7/7/2026 4:00 AM', author: 'GOOGLE', trigger: 'Weekly' },
  ]
}

export function mockSecurityEvents(): SecurityEvent[] {
  const events: SecurityEvent[] = [
    { id: 4624, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 10:30:15 AM', message: 'An account was successfully logged on. Subject: SYSTEM, Logon ID: 0x3E7', important: false },
    { id: 4634, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 10:25:00 AM', message: 'An account was logged off. Subject: shahr, Logon ID: 0xABCD', important: false },
    { id: 4672, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 10:30:15 AM', message: 'Special privileges assigned to new logon. Privileges: SeSecurityPrivilege, SeBackupPrivilege', important: false },
    { id: 4688, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 10:28:00 AM', message: 'A new process has been created. Creator: SYSTEM, New Process: svchost.exe', important: false },
    { id: 5156, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 10:27:30 AM', message: 'The Windows Filtering Platform has permitted a connection. Local: 192.168.1.100:5432, Remote: 10.0.0.5:40012', important: false },
    { id: 5379, level: 'Warning', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 9:15:00 AM', message: 'Credential Manager credentials were read. Target: WindowsLiveID, User: shahr', important: true },
    { id: 4625, level: 'Warning', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 8:45:22 AM', message: 'An account failed to log on. Subject: UNKNOWN, Failure Reason: Unknown user name or bad password', important: true },
    { id: 4648, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/7/2026 8:30:00 AM', message: 'A logon was attempted using explicit credentials. Target: LOCAL SERVICE', important: false },
    { id: 1102, level: 'Warning', provider: 'Microsoft-Windows-EventLog', time: '7/7/2026 7:00:00 AM', message: 'The audit log was cleared. Subject: SYSTEM', important: true },
    { id: 4720, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/6/2026 3:00:00 PM', message: 'A user account was created. New Account: svc_test', important: true },
    { id: 4732, level: 'Info', provider: 'Microsoft-Windows-Security-Auditing', time: '7/6/2026 3:00:05 PM', message: 'A member was added to a security-enabled local group. Member: svc_test, Group: Administrators', important: true },
    { id: 5038, level: 'Error', provider: 'Microsoft-Windows-Security-Auditing', time: '7/6/2026 1:00:00 PM', message: 'Code integrity determined that the image hash of a file is not valid. File: driver.sys', important: true },
    { id: 6281, level: 'Error', provider: 'Microsoft-Windows-Security-Auditing', time: '7/6/2026 12:30:00 PM', message: 'Code Integrity determined that a page hashes do not match. File: unknown', important: true },
  ]
  return events
}

export const opsLayers = [
  { id: 'sysops', icon: 'Monitor', title: 'System', description: 'CPU, memory, disk, processes', color: '#38bdf8' },
  { id: 'netops', icon: 'Network', title: 'Network', description: 'Ping, DNS, ports, traceroute', color: '#4ade80' },
  { id: 'secops', icon: 'Shield', title: 'Security', description: 'Firewall, users, defender', color: '#f87171' },
  { id: 'devops', icon: 'Terminal', title: 'DevOps', description: 'Commands, logs, files, services', color: '#fbbf24' },
  { id: 'aiops', icon: 'Brain', title: 'AI Ops', description: 'Ollama, chat, anomalies', color: '#a78bfa' },
] as const

// ── Log Entry Types ──

export interface LogEntry {
  id: string
  timestamp: string
  level: 'INFO' | 'WARN' | 'ERROR' | 'DEBUG'
  source: 'system' | 'network' | 'security' | 'devops' | 'ai'
  message: string
  module: string
  details?: string
  stackTrace?: string
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
  model?: string
}

export interface AnomalyPoint {
  time: string
  value: number
  anomaly: boolean
  severity?: 'low' | 'medium' | 'high' | 'critical'
  metric: string
}

export interface CommandHistoryItem {
  id: string
  command: string
  timestamp: string
  workingDir: string
  exitCode: number
}

// ── DevOps Mock Data ──

const logMessages: Record<string, string[]> = {
  system: [
    'Service started successfully',
    'Configuration loaded from /etc/hawkward/config.yml',
    'Health check passed for all services',
    'Database connection established',
    'Cache warmed up successfully',
    'Scheduled task triggered: log rotation',
    'Service watchdog fired, restarting service',
    'Resource limits updated for container',
    'Filesystem check completed, 0 errors',
    'System time synchronized with NTP server',
  ],
  network: [
    'Connection established to 192.168.1.1:443',
    'DNS resolution successful for api.example.com',
    'TLS handshake completed for grafana.example.com',
    'Packet loss detected on interface eth0',
    'BGP session with upstream 203.0.113.1 established',
    'HTTP request completed with status 200',
    'Interface eth0 link state changed to UP',
    'DHCP lease renewed for 192.168.1.100',
    'ARP resolution failed for 192.168.1.200',
    'Netflow export batch sent to collector',
  ],
  security: [
    'Login attempt from 203.0.113.42 failed: invalid credentials',
    'Firewall rule updated: block port 22 from public',
    'Certificate for *.example.com expires in 30 days',
    'User admin changed password successfully',
    'SSH key fingerprint verified for host git.example.com',
    'Audit log: configuration change by user shahr',
    'Failed SSH login from 10.0.0.99: max attempts exceeded',
    'SELinux policy reloaded successfully',
    'File integrity check: /etc/hosts unchanged',
    'OAuth token refreshed for service account',
  ],
  devops: [
    'Deployment v2.3.1 promoted to production',
    'CI pipeline #4287 passed in 3m 42s',
    'Docker image built: hawkward:v2.3.1',
    'Kubernetes rollout of deployment/api-gateway completed',
    'npm audit: 0 vulnerabilities found',
    'Terraform plan: 5 resources to add, 0 to destroy',
    'Helm release upgraded: prometheus-community/kube-prometheus-stack',
    'ArgoCD sync for app-hawkward completed successfully',
    'Grafana dashboard provisioned: system-overview-v2',
    'Git push to main: 12 commits, 4 files changed',
  ],
  ai: [
    'Model llama3.2:3b loaded into memory (2.1s)',
    'Inference request completed: 1250ms, 420 tokens',
    'Embedding generated for query: 1536 dimensions',
    'Model response cached for prompt hash a3f8c2',
    'Prompt template "system-health" compiled',
    'Token usage warning: 85% of context window used',
    'RAG query returned 5 relevant chunks from vector store',
    'Batch inference: 50 requests processed in 12.4s',
    'Model quantized to 4-bit: size reduced by 75%',
    'Inference fallback triggered: primary model unavailable',
  ],
}

const modules: Record<string, string[]> = {
  system: ['kernel', 'systemd', 'cron', 'fstab', 'ntpd', 'syslog'],
  network: ['eth0', 'wlan0', 'netfilter', 'dhcpcd', 'resolved', 'nginx'],
  security: ['pam', 'sshd', 'auditd', 'selinux', 'ufw', 'certbot'],
  devops: ['docker', 'kubelet', 'helm', 'argo', 'terraform', 'grafana'],
  ai: ['ollama', 'llama.cpp', 'chromadb', 'langchain', 'sentence-transformers'],
}

const levelPool: Array<'INFO' | 'WARN' | 'ERROR' | 'DEBUG'> = ['INFO', 'INFO', 'INFO', 'WARN', 'WARN', 'ERROR', 'DEBUG', 'DEBUG', 'DEBUG', 'DEBUG']

const sourcePool: Array<'system' | 'network' | 'security' | 'devops' | 'ai'> = ['system', 'network', 'security', 'devops', 'ai']

export function mockLogEntries(count: number = 50): LogEntry[] {
  return Array.from({ length: count }, (_, i) => {
    const source = sourcePool[Math.floor(Math.random() * sourcePool.length)]
    const level = levelPool[Math.floor(Math.random() * levelPool.length)]
    const msgs = logMessages[source]
    const mods = modules[source]
    const hasStackTrace = level === 'ERROR' && Math.random() > 0.4
    return {
      id: `log-${Date.now()}-${i}`,
      timestamp: new Date(Date.now() - (count - i) * 60000 * (0.5 + Math.random())).toISOString(),
      level,
      source,
      message: msgs[Math.floor(Math.random() * msgs.length)],
      module: mods[Math.floor(Math.random() * mods.length)],
      details: level === 'ERROR' || level === 'WARN' ? `Detail: additional context for this ${level} event from ${source}` : undefined,
      stackTrace: hasStackTrace
        ? `  at runtime.main (runtime/proc.go:267)\n  at main.run (cmd/hawkward/main.go:45)\n  at main.handleRequest (cmd/hawkward/handler.go:112)\n  at pkg/service.Process (pkg/service/service.go:78)\n  at pkg/worker.Execute (pkg/worker/worker.go:34)`
        : undefined,
    }
  })
}

export function mockChatMessages(): ChatMessage[] {
  const now = Date.now()
  return [
    {
      id: 'chat-1',
      role: 'user',
      content: 'Can you analyze the current system health and identify any issues?',
      timestamp: new Date(now - 300000).toISOString(),
    },
    {
      id: 'chat-2',
      role: 'assistant',
      content: '# System Health Analysis\n\n## Overview\nThe system is currently **healthy** with no critical issues detected.\n\n## Key Metrics\n| Metric | Value | Status |\n|--------|-------|--------|\n| CPU Usage | 42% | ✅ Normal |\n| Memory | 58% | ⚠️ Elevated |\n| Disk | 65% | ✅ Normal |\n| Network | 2.3 MB/s | ✅ Normal |\n\n## Recommendations\n1. Monitor memory usage - currently above 50%\n2. No immediate action required\n3. Schedule routine maintenance for next weekend',
      timestamp: new Date(now - 285000).toISOString(),
      model: 'llama3.2:3b',
    },
    {
      id: 'chat-3',
      role: 'user',
      content: 'Show me recent security events from the last 24 hours',
      timestamp: new Date(now - 240000).toISOString(),
    },
    {
      id: 'chat-4',
      role: 'assistant',
      content: '# Security Events (Last 24h)\n\n| Time | Event | Severity |\n|------|-------|----------|\n| 10:30 AM | Logon success - SYSTEM | Info |\n| 09:15 AM | Failed login attempt from 203.0.113.42 | **Warning** |\n| 08:45 AM | Credential Manager access detected | **Warning** |\n| 07:00 AM | Audit log cleared | **Critical** |\n| 03:00 AM | New user account created: svc_test | Info |\n\n## Flags\n🔴 **1 critical event**: Audit log was cleared by SYSTEM at 7:00 AM\n🟡 **2 warnings**: Failed logins and credential access\n\n> Recommend investigating the audit log clearing event immediately.',
      timestamp: new Date(now - 220000).toISOString(),
      model: 'llama3.2:3b',
    },
    {
      id: 'chat-5',
      role: 'user',
      content: 'Generate a network audit report for the last hour',
      timestamp: new Date(now - 120000).toISOString(),
    },
    {
      id: 'chat-6',
      role: 'assistant',
      content: '# Network Audit Report\n\n**Time Range:** ${new Date(now - 3600000).toLocaleTimeString()} - ${new Date(now).toLocaleTimeString()}\n**Generated:** ${new Date(now).toLocaleString()}\n\n## Interfaces\n- **Ethernet0**: 1 Gbps, UP - RX: 1.25 GB, TX: 850 MB\n- **WiFi**: 866 Mbps, UP - RX: 450 MB, TX: 320 MB\n- **Loopback**: UP - Minimal traffic\n\n## Active Connections\n- **Total**: 24 connections\n- **Established**: 18\n- **Time Wait**: 4\n- **Close Wait**: 2\n\n## DNS\n- **Queries**: 156 resolved, 0 failures\n- **Avg resolution time**: 23ms\n\n## Port Scan\n- **Open ports**: 22, 80, 443, 3389, 8080, 5432, 6379\n- **Filtered**: 8 ports\n\n✅ **Network health: Good** - No anomalies detected.',
      timestamp: new Date(now - 100000).toISOString(),
      model: 'llama3.2:3b',
    },
  ]
}

export function mockAnomalyData(points: number = 50): AnomalyPoint[] {
  const now = Date.now()
  const metrics = ['CPU', 'Memory', 'Network RX', 'Network TX', 'Disk I/O']
  return Array.from({ length: points }, (_, i) => {
    const isAnomaly = Math.random() < 0.08
    const baseValue = 30 + Math.random() * 40
    const metric = metrics[Math.floor(Math.random() * metrics.length)]
    const anomalyValue = isAnomaly ? baseValue + 50 + Math.random() * 80 : 0
    const severities: Array<'low' | 'medium' | 'high' | 'critical'> = ['low', 'medium', 'high', 'critical']
    return {
      time: new Date(now - (points - i) * 60000 * 3).toLocaleTimeString(),
      value: isAnomaly ? anomalyValue : baseValue,
      anomaly: isAnomaly,
      severity: isAnomaly ? severities[Math.floor(Math.random() * severities.length)] : undefined,
      metric,
    }
  })
}

export function mockCommandHistory(): CommandHistoryItem[] {
  const commands = [
    'systemctl status docker',
    'kubectl get pods -n hawkward',
    'docker ps --format "table {{.Names}}\\t{{.Status}}"',
    'curl -s -o /dev/null -w "%{http_code}" https://api.example.com/health',
    'helm list -n monitoring',
    'terraform plan -out=plan.tfplan',
    'npm run build -- --prod',
    'git log --oneline -n 10',
    'journalctl -u sshd --since "1 hour ago"',
    'df -h | grep -E "^/dev/"',
    'netstat -tulpn | grep LISTEN',
    'ping -c 4 8.8.8.8',
    'traceroute -n google.com',
    'ps aux --sort=-%mem | head -10',
    'free -h',
  ]
  return Array.from({ length: 10 }, (_, i) => ({
    id: `cmd-${i}`,
    command: commands[Math.floor(Math.random() * commands.length)],
    timestamp: new Date(Date.now() - i * 300000).toISOString(),
    workingDir: i % 2 === 0 ? '/home/hawkward' : '/var/log',
    exitCode: Math.random() > 0.2 ? 0 : 1,
  }))
}

const reportTemplates = {
  health: `# System Health Report

**Generated:** {date}
**Status:** ✅ All systems nominal

## Summary
| Component | Status | Value |
|-----------|--------|-------|
| CPU | ✅ OK | 42% |
| Memory | ⚠️ Warning | 78% |
| Disk (C:) | ✅ OK | 55% |
| Disk (D:) | ✅ OK | 40% |
| Network | ✅ OK | 2.1 MB/s |

## Top Processes by Memory
1. chrome.exe - 420 MB
2. code.exe - 380 MB
3. discord.exe - 210 MB
4. node.exe - 180 MB
5. python.exe - 150 MB

## Recommendations
- Memory usage is elevated. Consider closing unused applications.
- No critical issues detected.
`,
  network: `# Network Audit Report

**Generated:** {date}

## Interfaces
| Interface | Status | Speed | RX | TX |
|-----------|--------|-------|----|----|
| Ethernet0 | ✅ UP | 1 Gbps | 1.25 GB | 850 MB |
| WiFi | ✅ UP | 866 Mbps | 450 MB | 320 MB |
| Loopback | ✅ UP | N/A | 12 KB | 12 KB |

## Active Connections: 24
- Established: 18
- Time Wait: 4
- Close Wait: 2

## DNS Resolution
- Queries: 156
- Failures: 0
- Avg time: 23ms

## Open Ports
22 (SSH), 80 (HTTP), 443 (HTTPS), 3389 (RDP), 5432 (PostgreSQL), 6379 (Redis)

## Audit Result: ✅ PASS
`,
  security: `# Security Audit Report

**Generated:** {date}

## Events (Last 24h)
| Time | Event | Severity |
|------|-------|----------|
| 10:30 AM | Logon success - SYSTEM | Info |
| 09:15 AM | Failed login from 203.0.113.42 | Warning |
| 08:45 AM | Credential Manager access | Warning |
| 07:00 AM | Audit log cleared | 🔴 Critical |

## Firewall Rules
- Total rules: 15
- Inbound allow: 4
- Inbound block: 5
- Outbound allow: 6

## Windows Defender
- Status: ✅ Active
- Real-time: Enabled
- Signatures: Up to date

## Findings
🔴 **Critical**: Audit log was cleared by SYSTEM
🟡 **Warning**: Multiple failed login attempts detected
`,
  combined: `# Combined Operations Report

**Generated:** {date}

## System Health: ✅ PASS
- CPU: 42% | Memory: 78% | Disk: 55% | Network: 2.1 MB/s

## Network Audit: ✅ PASS
- 4 interfaces, 24 connections, 156 DNS queries

## Security Audit: ⚠️ WARNINGS
- 1 critical event (audit log cleared)
- 2 warnings (failed logins, credential access)

## Overall Assessment
| Category | Status |
|----------|--------|
| System | ✅ Healthy |
| Network | ✅ Healthy |
| Security | ⚠️ Needs Review |
| DevOps | ✅ Operational |
| AI Ops | ✅ Operational |
`,
}

export function mockReportContent(type: 'health' | 'network' | 'security' | 'combined'): string {
  const template = reportTemplates[type]
  return template.replace('{date}', new Date().toLocaleString())
}
