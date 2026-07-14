// ── Collector Types ──

export interface CollectorStatus {
  id: string
  name: string
  description: string
  enabled: boolean
  interval_ms: number
  default_interval_ms: number
  last_run: string | null
}

// ── Go type mirrors for frontend use ──

export interface GaugeMetric {
  value: number
  unit: string
  history: number[]
  forecast: number[]
  trend: 'rising' | 'falling' | 'stable'
}

export interface NetworkMetric {
  rx_rate: number
  tx_rate: number
  unit: string
}

export interface DashboardData {
  cpu: GaugeMetric
  memory: GaugeMetric
  disk: GaugeMetric
  gpu: GPUData
  battery: BatteryData
  network: NetworkMetric
  processes: number
  connections: number
  alerts: number
  uptime: string
}

export interface GPUData {
  name: string
  vendor: string
  memory_gb: number
  driver: string
  detected: boolean
}

export interface BatteryData {
  percent: number
  charging: boolean
  time_left_sec: number
  status: string
  detected: boolean
}

export interface CPUInfo {
  percent: number
  per_cpu: number[]
  model_name: string
  logical_cores: number
  physical_cores: number
  core_count: number
  load_avg_1: number
  load_avg_5: number
  load_avg_15: number
}

export interface MemoryInfo {
  total_bytes: number
  available_bytes: number
  used_bytes: number
  used_percent: number
  cached_bytes: number
  total_gb: number
  used_gb: number
  swap_total: number
  swap_used: number
  swap_percent: number
}

export interface DiskPartition {
  mountpoint: string
  total_bytes: number
  free_bytes: number
  used_bytes: number
  used_percent: number
  fs_type: string
  device: string
}

export interface DiskInfo {
  partitions: DiskPartition[]
}

export interface ProcessInfo {
  pid: number
  ppid: number
  name: string
  cpu: number
  memory: number
  mem_pct: number
  status: string
  num_fds: number
  read_bytes: number
  write_bytes: number
}

export interface SystemInfo {
  hostname: string
  os: string
  platform: string
  platform_version: string
  kernel_version: string
  kernel_arch: string
  uptime: string
  process_count: number
  virtualization: string
}

export interface AlertInfo {
  id: string
  level: string
  metric: string
  message: string
  value: number
  threshold: number
  timestamp: string
  resolved: boolean
}

export interface AlertRuleInfo {
  metric: string
  condition: string
  threshold: number
  flap_count: number
  severity: string
  message: string
}

export type HealthStatus = 'healthy' | 'warning' | 'critical'
export type TrendDirection = 'up' | 'down' | 'stable'

export interface TimeSeriesPoint {
  time: string
  value: number
}

export interface MetricDataPoint {
  Time: string
  Value: number
}

// ── NetOps Types ──

export interface PingEntry {
  seq: number
  ip: string
  rtt_ms: number | null
  jitter_ms: number | null
  ttl: number | null
  status: 'success' | 'timeout'
}

export interface PingStats {
  target: string
  sent: number
  received: number
  lost: number
  lost_pct: number
  min_ms: number | null
  max_ms: number | null
  avg_ms: number | null
}

export interface DNSResult {
  hostname: string
  a: string[]
  aaaa: string[]
  mx: string[]
  ns: string[]
  cname: string
  txt: string[]
  error?: string
}

export interface PortResult {
  port: number
  open: boolean
  service: string
}

export interface TraceHop {
  number: number
  host: string
  ip: string
  rtts_ms: number[]
  timed: boolean
  avg_rtt: number | null
}

export interface TraceResult {
  target: string
  hops: TraceHop[]
  error?: string
}

export interface ConnectionInfo {
  local_addr: string
  local_port: number
  remote_addr: string
  remote_port: number
  protocol: string
  state: string
  pid: number
  process_name: string
}

export interface InterfaceInfo {
  name: string
  mac: string
  ips: string[]
  is_up: boolean
  speed: string
  mtu: number
  flags: string
  rx_bytes: number
  tx_bytes: number
  rx_rate_bps: number
  tx_rate_bps: number
  rx_history: number[]
  tx_history: number[]
}

export type NetworkChangeType = 'up' | 'down' | 'ip_added' | 'ip_removed' | 'appeared' | 'disappeared'

export interface NetworkChange {
  type: NetworkChangeType
  interface: string
  detail: string
  timestamp: string
}

// ── SecOps Types ──

export interface FirewallRule {
  name: string
  direction: string
  action: string
  protocol: string
  local_port: string
  remote_port: string
  remote_ip: string
  profile: string
  enabled: boolean
  is_high_risk: boolean
}

export interface UserInfo {
  username: string
  full_name: string
  sid: string
  group: string
  is_admin: boolean
  is_enabled: boolean
  password_never_expires: boolean
  last_logon: string
}

export interface ListeningPort {
  port: number
  protocol: string
  process_name: string
  pid: number
  state: string
  is_external: boolean
}

export interface DefenderStatus {
  enabled: boolean
  up_to_date: boolean
  signature_age: string
  last_scan: string
  real_time_protection: boolean
  cloud_protection: boolean
  am_service_enabled: boolean
  antispyware_enabled: boolean
  nis_enabled: boolean
  quick_scan_age: number
  full_scan_age: number
  threats_detected: number
}

export interface ScheduledTask {
  name: string
  status: string
  next_run: string
  last_run: string
  author: string
  trigger: string
}

export interface SecurityEvent {
  id: number
  level: string
  provider: string
  time: string
  message: string
  important: boolean
}

export interface SecurityScore {
  score: number
  grade: string
  breakdown: Record<string, number>
  recommendations: string[]
}

export interface FirewallProfile {
  name: string
  enabled: boolean
}

export interface FirewallStatus {
  enabled: boolean
  profiles: FirewallProfile[]
}

export interface RiskInfo {
  category: string
  severity: string
  title: string
  description: string
  recommendation: string
}

export interface SecuritySummary {
  score: number
  summary: string
  risks: string[]
  recommendations: string[]
  analyzedAt: string
}

// ── DevOps Types ─────────────────────────────────────────────────────────────
export interface DockerStatus {
  installed: boolean
  running: boolean
  version: string
  containers: ContainerSummary
}

export interface KubernetesStatus {
  installed: boolean
  connected: boolean
  cluster: string
  nodes: number
  pods: number
}

export interface ServiceInfo {
  name: string
  status: string
  port: number
}

export interface ServiceCategory {
  category: string
  services: ServiceInfo[]
}

export interface ServiceGroupSummary {
  databases: number
  messageQueues: number
  webServers: number
  containers: number
  other: number
  running: number
  stopped: number
}

export interface ContainerInfo {
  id: string
  name: string
  image: string
  state: string
  status: string
  ports: string
}

export interface ContainerSummary {
  running: number
  stopped: number
  failed: number
  total: number
  containers: ContainerInfo[]
}

export interface GitRepoInfo {
  path: string
  branch: string
  modified_files: number
  untracked_files: number
  ahead: number
  behind: number
  clean: boolean
}

export interface GitSummary {
  repositories: GitRepoInfo[]
  total_repos: number
}

export interface LocalServer {
  port: number
  protocol: string
  process: string
  pid: number
  framework: string
  health: string
}

export interface EnvVarInfo {
  name: string
  value: string
}

export interface ToolVersion {
  name: string
  version: string
}

export interface EnvironmentInfo {
  path_dirs: string[]
  key_vars: EnvVarInfo[]
  sdks: ToolVersion[]
  package_managers: ToolVersion[]
}

export interface ToolInfo {
  name: string
  version: string
  path: string
  status: 'installed' | 'not-found' | 'error'
}

export interface DevOpsSuggestion {
  category: 'docker' | 'git' | 'node' | 'general'
  severity: 'info' | 'warning' | 'critical'
  message: string
  action: string
}

export interface CommandResult {
  command: string
  output: string
  exit_code: number
  duration_ms: number
  error?: string
}

export interface ServiceEntry {
  name: string
  display_name: string
  status: string
  start_type: string
}

export interface FileEntry {
  name: string
  path: string
  size: string
  raw_size: number
  is_dir: boolean
  is_binary: boolean
  mode: string
  mod_time: string
}

// ── Log Types ──

export interface LogEntry {
  timestamp: string
  level: string
  module: string
  message: string
  line: string
  source: string
}

export interface LogSourceCount {
  source: string
  count: number
}

export interface TrendingError {
  message: string
  count: number
  lastSeen: string
}

export interface LogStats {
  totalToday: number
  totalThisHour: number
  totalLastMin: number
  errorCount: number
  warningCount: number
  infoCount: number
  debugCount: number
  topSources: LogSourceCount[]
  trendingErrors: TrendingError[]
}

export interface LogTimelinePoint {
  timestamp: string
  errors: number
  warnings: number
  info: number
}

export interface LogSummary {
  summaryText: string
  topSource: string
  trend: string
}

// ── NetOps Overview Types ──

export interface GatewayInfo {
  ip: string
  interface: string
  reachable: boolean
}

export interface NetworkSummary {
  summaryText: string
  topInterface: string
  issues: string[]
}

// ── AIOps Types ──

export interface ChatMessage {
  role: string
  content: string
}

export interface OllamaStatus {
  available: boolean
  model: string
  version: string
  available_models?: string[]
  error?: string
}

export interface AnomalyInfo {
  metric: string
  value: number
  expected: number
  deviation: number
  severity: string
  timestamp: string
}

// ── Topology Types ──

export type DeviceType = 'router' | 'switch' | 'server' | 'workstation' | 'firewall' | 'cloud'
export type TopologyStatus = 'healthy' | 'warning' | 'critical'
export type ConnectionType = 'ethernet' | 'fiber' | 'wireless'

export interface TopologyDevice {
  id: string
  type: DeviceType
  label: string
  x: number
  y: number
  ip?: string
  subnet?: string
  mac?: string
  status: TopologyStatus
  notes?: string
}

export interface TopologyConnection {
  id: string
  sourceId: string
  targetId: string
  label?: string
  type: ConnectionType
}

export interface AIInsight {
  category: string
  severity: 'info' | 'warning' | 'critical'
  title: string
  message: string
  action: string
  timestamp: string
}

export interface AIConfidence {
  overall: number
  factors: Record<string, number>
  updatedAt: string
}

export interface ConversationMessage {
  id: number
  session_id: string
  role: string
  content: string
  timestamp: string
}

export interface LearnedBaseline {
  metric: string
  mean: number
  min: number
  max: number
  stdDev: number
  count: number
}

// ── NetworkDesign Analysis Types ──

export interface TopologyNode {
  id: string
  type: string
  label: string
  ip: string
  mac: string
  vendor: string
  notes: string
  vlan: string
  online: boolean
  props: Record<string, string>
}

export interface TopologyEdge {
  id: string
  source: string
  target: string
  label: string
  type: string
  bandwidth: string
  status: string
}

export interface DuplicateIPEntry {
  ip: string
  nodes: string[]
}

export interface TopologyHealth {
  totalNodes: number
  totalEdges: number
  brokenLinks: number
  missingLabels: number
  orphanNodes: string[]
  duplicateIPs: DuplicateIPEntry[]
  subnetErrors: string[]
  suggestions: string[]
}

export interface DeviceInventoryGroup {
  type: string
  count: number
  devices: TopologyDevice[]
}

export interface ChatSession {
  session_id: string
  last_active: string
  msg_count: number
}

// ── Extended SysOps Types ──

export interface PerCPUInfo {
  core: number
  frequency_mhz: number
  usage_percent: number
}

export interface CPUExtendedInfo {
  model_name: string
  frequency_mhz: number
  cache_size_kb: number
  temperature: number
  per_cpu_info: PerCPUInfo[]
}

export interface DiskIOEntry {
  name: string
  read_bytes: number
  write_bytes: number
  read_count: number
  write_count: number
}

export interface DiskIOData {
  disks: DiskIOEntry[]
  total_read_bytes: number
  total_write_bytes: number
}

export interface LoggedInUserData {
  user: string
  terminal: string
  host: string
  started: string
}

export interface CPUTimesData {
  user: number
  system: number
  idle: number
  iowait: number
  steal: number
  total: number
}

export interface LoadAverageData {
  load_1: number
  load_5: number
  load_15: number
}

export interface PerformanceData {
  cpu_times: CPUTimesData
  load_average: LoadAverageData
  io_wait: number
}

export interface ActionResult {
  action: string
  success: boolean
  message: string
  output: string
}

export interface SystemLogEntry {
  timestamp: string
  level: string
  source: string
  message: string
}

export interface SystemLogsResult {
  entries: SystemLogEntry[]
  source: string
  total: number
}

export interface PackageData {
  name: string
  version: string
}

export interface PackageManagerData {
  name: string
  found: boolean
  packages: PackageData[]
}

export interface ScheduledTaskData {
  name: string
  schedule: string
  command: string
  enabled: boolean
  next_run: string
}

export interface DiagnosticCheckData {
  name: string
  status: 'pass' | 'warn' | 'fail'
  message: string
  value: string
}

export interface ExtendedDiagnosticResult {
  checks: DiagnosticCheckData[]
  score: number
  timestamp: string
}

// ── NetOps Extended Types ──

export interface ARPEntryData {
  ip: string
  mac: string
  vendor: string
  interface: string
}

export interface RouteEntryData {
  destination: string
  mask: string
  gateway: string
  interface: string
  metric: number
  is_default: boolean
}

export interface WiFiNetworkData {
  ssid: string
  signal: number
  channel: number
  security: string
  bssid: string
  frequency: string
}

export interface WiFiInfoData {
  interface: string
  ssid: string
  signal: number
  speed: string
  channel: number
}

export interface DoHResultData {
  server: string
  latency_ms: number
  success: boolean
  resolved_ip: string
}

export interface PingResultMultiData {
  target: string
  min_ms: number
  avg_ms: number
  max_ms: number
  stddev_ms: number
  packet_loss: number
  jitter_ms: number
  individual_rtts: number[]
  success: boolean
  error?: string
}

export interface PingStatsData {
  avg_latency: number
  max_latency: number
  total_loss: number
  worst_target: string
}

export interface HealthCheckData {
  name: string
  status: 'pass' | 'warn' | 'fail'
  detail: string
  score: number
}

export interface HealthReportData {
  score: number
  checks: HealthCheckData[]
  summary: string
  duration: string
}

export interface VPNStatusData {
  active: boolean
  type: string
  interface: string
  remote_ip: string
  local_ip: string
  protocol: string
}

export interface NetOpsFirewallRuleData {
  name: string
  direction: string
  action: string
  protocol: string
  ports: string
  enabled: boolean
  source: string
  destination: string
}

export interface DiscoveredDeviceData {
  ip: string
  mac: string
  vendor: string
  hostname: string
  open_ports: number[]
  response_time_ms: number
}

export interface DiscoveryResultData {
  devices: DiscoveredDeviceData[]
  subnet: string
  scan_time_ms: number
}

export interface BandwidthSampleData {
  timestamp: string
  rx_bytes_per_sec: number
  tx_bytes_per_sec: number
  interface: string
}

export interface NetworkActionResult {
  action: string
  message: string
  success: boolean
}
