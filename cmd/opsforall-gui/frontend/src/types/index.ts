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
  health_score: number
  health_trend: HealthScorePoint[]
}

export interface HealthScorePoint {
  day: string
  score: number
}

export interface GPUData {
  name: string
  vendor: string
  memory_gb: number
  driver: string
  detected: boolean
  temperature: number
  utilization: number
  fan_speed: number
}

export interface GPUInfo {
  name: string
  vendor: string
  memory_gb: number
  driver: string
  detected: boolean
  temperature: number
  utilization: number
  fan_speed: number
}

export interface BatteryData {
  percent: number
  charging: boolean
  time_left_sec: number
  status: string
  detected: boolean
}

export interface BatteryInfo {
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
  is_signed: boolean
  publisher: string
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
  // No lost_pct field is sent by the backend (PingResult in Types.go has no
  // matching field) — derive it from lost/sent at the call site instead of
  // relying on this phantom field (H12).
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
  service_name: string
  risk_level: 'high' | 'medium' | 'low'
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

// ── SecOps Phase 2 Types ──

export interface PasswordPolicy {
  max_age: number
  min_length: number
  complexity: boolean
  lockout_threshold: number
  lockout_duration: number
}

export interface FailedLogin {
  time: string
  username: string
  source_ip: string
  count: number
}

export interface LockedAccount {
  username: string
  locked_since: string
}

export interface DiskEncryption {
  volume: string
  encrypted: boolean
  method: string
  status: string
}

export interface SecureBoot {
  enabled: boolean
  state: string
}

export interface SystemService {
  name: string
  display_name: string
  status: string
  startup_type: string
}

export interface TLSCertificate {
  subject: string
  issuer: string
  not_after: string
  key_size: number
  is_expiring: boolean
  days_left: number
}

export interface SSHConfig {
  permit_root_login: string
  password_authentication: string
  pubkey_authentication: string
  x11_forwarding: string
  max_auth_tries: string
}

export interface HardeningCheck {
  category: string
  check: string
  passed: boolean
  severity: string
  remediation: string
}

export interface AuditCheckItem {
  category: string
  check: string
  passed: boolean
  description: string
  remediation: string
}

export interface SecurityAuditResult {
  score: number
  total: number
  passed: number
  failed: number
  items: AuditCheckItem[]
  timestamp: string
}

export interface PrivilegeEvent {
  time: string
  username: string
  privilege: string
  process: string
}

export interface SecTimelineEvent {
  time: string
  type: string
  detail: string
  severity: string
}

export interface PublicExposure {
  port: number
  protocol: string
  process_name: string
  severity: string
}

export interface SecActionResult {
  success: boolean
  message: string
  error?: string
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

// ── DevOps Extended Types ──

export interface GitBranchInfo {
  name: string
  current: boolean
  upstream: string
  ahead: number
  behind: number
  last_commit: string
}

export interface GitTagInfo {
  name: string
  commit: string
  date: string
  msg: string
}

export interface GitStashEntry {
  index: number
  message: string
}

export interface GitRemoteInfo {
  name: string
  url: string
  type: string
}

export interface DockerStatsEntry {
  container_id: string
  name: string
  cpu_percent: string
  memory_usage: string
  memory_limit: string
  memory_percent: string
  net_io: string
  block_io: string
  pid_count: string
}

export interface DockerComposeService {
  name: string
  state: string
  ports: string
}

export interface DockerComposeProject {
  project: string
  status: string
  work_dir: string
  services: DockerComposeService[]
}

export interface DockerNetworkInfo {
  id: string
  name: string
  driver: string
  scope: string
  subnet: string
  gateway: string
  containers: number
}

export interface DockerVolumeInfo {
  driver: string
  name: string
  mountpoint: string
  size: string
}

export interface K8sResourceItem {
  name: string
  namespace: string
  status: string
  age: string
  details: string
}

export interface K8sRolloutStatus {
  name: string
  kind: string
  ready: boolean
  replicas: string
  updated: string
  available: string
}

export interface K8sEvent {
  last_seen: string
  type: string
  reason: string
  object: string
  message: string
}

export interface K8sNamespaceInfo {
  name: string
  status: string
  age: string
}

export interface K8sScalingResult {
  current: number
  desired: number
  success: boolean
  output: string
}

export interface BuildSystemInfo {
  name: string
  version: string
  found: boolean
  path: string
}

export interface BuildTargetInfo {
  name: string
  type: string
  path: string
  has_build: boolean
  has_test: boolean
  has_lint: boolean
  has_package: boolean
  dep_count: number
}

export interface CICDConfig {
  platform: string
  config_files: string[]
  detected: boolean
}

export interface CICDPipelineInfo {
  name: string
  status: string
  branch: string
  commit: string
  duration: string
  updated_at: string
  url: string
}

export interface CICDStatus {
  platform: string
  enabled: boolean
  config_found: boolean
  pipelines: CICDPipelineInfo[]
  configs: CICDConfig[]
}

export interface ReleaseInfo {
  version: string
  date: string
  branch: string
  tag: string
  commit: string
  status: string
  notes: string
}

export interface ReleaseHistory {
  releases: ReleaseInfo[]
  total_count: number
  last_release: string
}

export interface DeploymentRecord {
  id: string
  version: string
  environment: string
  status: string
  timestamp: string
  duration: string
  commit: string
  trigger: string
}

export interface DORAMetrics {
  deployment_frequency: string
  lead_time_for_changes: string
  change_failure_rate: string
  mttr: string
  period: string
  deploy_count: number
  incident_count: number
  lead_time_avg_hours: number
  mttr_avg_minutes: number
  failure_pct: number
}

export interface DevOpsDiagCheck {
  name: string
  status: string
  message: string
  value: string
}

export interface DevOpsDiagResult {
  checks: DevOpsDiagCheck[]
  score: number
  timestamp: string
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
  total: number
  errors: number
  warnings: number
  info: number
}

export interface LogSummary {
  summaryText: string
  topSource: string
  topMessage: string
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

export interface ActionPreview {
  handshake_id: string
  action: string
  command?: string
  description: string
  risks: string[]
  rollback: string
}

export interface ChatMessage {
  role: string
  content: string
  actions?: ActionPreview[]
}

export interface OllamaStatus {
  available: boolean
  binary_exists: boolean
  model: string
  version: string
  available_models?: string[]
  error?: string
}

export interface OllamaProgress {
  status: string
  percent: number
  total: number
  completed: number
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

export type DeviceType = 'router' | 'switch' | 'server' | 'workstation' | 'firewall' | 'cloud' | 'gateway' | 'printer' | 'iot' | 'unknown'
export type TopologyStatus = 'healthy' | 'warning' | 'critical'
export type ConnectionType = 'ethernet' | 'fiber' | 'wireless' | 'vpn' | 'direct'

export interface TopologyDevice {
  id: string
  type: DeviceType
  label: string
  x: number
  y: number
  ip?: string
  subnet?: string
  mac?: string
  vendor?: string
  hostname?: string
  status: TopologyStatus
  online?: boolean
  notes?: string
}

export interface TopologyConnection {
  id: string
  sourceId: string
  targetId: string
  label?: string
  type: ConnectionType
  metric?: number
}

// ── Backend Mirror Types (from internal/app/Types.go) ──

export interface TopologyDeviceData {
  id: string
  type: string
  label: string
  ip?: string
  mac?: string
  subnet?: string
  vendor?: string
  hostname?: string
  status: string
  x: number
  y: number
  online: boolean
  notes?: string
}

export interface TopologyConnectionData {
  id: string
  source_id: string
  target_id: string
  type: string
  label?: string
  metric?: number
}

export interface NetworkTopologyData {
  devices: TopologyDeviceData[]
  connections: TopologyConnectionData[]
  generated_at: string
  subnet: string
}

export interface DiscoveryTemplateData {
  id: string
  name: string
  description: string
  run_ping: boolean
  run_dns: boolean
  run_trace: boolean
  run_arp: boolean
  run_routing: boolean
  run_port_scan: boolean
  ping_count: number
}

export interface AIInsight {
  category: string
  severity: 'info' | 'warning' | 'critical'
  title: string
  message: string
  action: string
  actionPage?: string
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

// ── Hardware Telemetry Types ──

export interface SensorData {
  name: string
  type: string
  value: number
  unit: string
  category: string
}

export interface BaseboardInfo {
  manufacturer: string
  product: string
  version: string
  serial_number: string
}

export interface HardwareInfo {
  cpu: CPUExtendedInfo
  gpu: GPUInfo
  battery: BatteryInfo
  sensors: SensorData[]
  baseboard: BaseboardInfo
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

export interface SystemRecommendation {
  category: string
  severity: string
  message: string
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

export interface ReportRecord {
  id: string
  timestamp: string
  type: string
  score: number
  data_json: string
}

export interface PrebuiltTemplate {
  id: string
  category: string
  preset_name: string
  description: string
  metric: string
  condition: string
  threshold: number
  report_type: string
  schedule: string
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

export interface PingResult {
  target: string
  ip: string
  sent: number
  received: number
  lost: number
  min_ms: number
  max_ms: number
  avg_ms: number
  jitter_ms: number
  ttl: number
  error?: string
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

// ── Control Plane Types ──

export interface CapabilityInfo {
  id: string
  available: boolean
  path: string
}

// ── LHM (LibreHardwareMonitor) Types ──

export interface LHMStatusResult {
  available: boolean
  running: boolean
  needsAdmin: boolean
  version: string
  error?: string
}

export interface LHMAuthorization {
  reason: string
  capabilities: string[]
  risks: string[]
  binaryName: string
  publisher: string
}

// ── AIOps Live Context Types ──

export interface AIWorkflowEvent {
  sessionId: string
  stage: string
  status: 'running' | 'completed' | 'error'
  detail: string
  timestamp: string
}

export interface DataStreamMetric {
  name: string
  unit: string
  lastValue: number
  samples: number
  trend: 'rising' | 'falling' | 'stable'
  updatedAt: string
}

/* ── SLO/SLI Types ── */

export interface SLODefinition {
  id: string
  name: string
  metric: string
  comparison: string
  threshold: number
  targetPct: number
  windowDays: number
  enabled: boolean
  description: string
}

export interface SLIResult {
  sloId: string
  sloName: string
  compliantPct: number
  targetPct: number
  met: boolean
  samples: number
  evaluatedAt: string
}

export interface SLOSummary {
  totalSLOs: number
  metCount: number
  missCount: number
  overallPct: number
  results: SLIResult[]
}
