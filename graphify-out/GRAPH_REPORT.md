# Graph Report - .  (2026-07-10)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1675 nodes · 3025 edges · 126 communities (88 shown, 38 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 492 edges (avg confidence: 0.8)
- Token cost: 6,592 input · 3,445 output

## Graph Freshness
- Built from commit: `b861467e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Hawk Management
- Storage & Timeline
- Network Lookup & Scan
- Anomaly Detection
- Dev Report
- Active Job Sandbox
- CPU Info Collector
- Alerts & System Status
- App Information
- Storage & Timeline
- Dependencies & CSS Classes
- Network Design & Connections
- Log Management & DevOps
- Theme Configuration
- Configuration & Compiler Options
- Data Pipeline Configurations
- Time Series Data
- Service Status & Listeners
- Log Management & Event Bus
- Alerts & Conditions
- Dev Dependencies & ESLint Plugins
- AIOps.tsx
- Platform Checks & Admin Privileges
- AIOps Dashboard & Tabs
- Metrics Definition & Pipeline API
- Events Handling & Dashboard
- Configuration & Compiler Options
- New Data Pipeline & Tests
- New Data Pipeline & Tests
- Tasks Management & Parsing
- Wails Configuration & Build Tools
- AIOps
- SecOps Firewall Rules & Status
- Package Information & Repository Details
- Connections Management & Process Mapping
- Dashboard Briefing Generation
- Dashboard Data Retrieval
- SysOps
- App.tsx
- NetOps Dashboard & Tabs
- Forecast Engine & Trend Prediction
- Defender Configuration & JSON Parsing
- Alert API & Alerts Management
- Settings Configuration & Mocks
- Bandwidth Counter & Network Metrics
- Logs.tsx
- Main Content & System Queries
- System Query Sandbox & Network Metrics
- ErrorBoundary.tsx
- Confirm Dialog & Tabs
- Empty State & Security Tabs
- NewApp
- Backend Configuration & Anomaly Tab
- DashboardData
- Metric History Management
- Firewall Rules
- SecOps Tests & Firewall Rules
- Event Stream
- Metrics Exporter Initialization
- Page Components
- Workflows & Security Events
- Sandbox Tests
- LogInfo
- runtime.d.ts
- setup.ts
- DiskInfo
- TraceHop
- scripts
- Connection Line Visualization
- events.go
- package.json
- tsconfig.json
- AlertInfo
- AnomalyInfo
- AppInfo
- BriefingSection
- CommandResult
- ConnectionInfo
- CPUInfo
- DefenderStatus
- DiagnosticResult
- DNSResult
- FileEntry
- FirewallRule
- InterfaceInfo
- ListeningPort
- LogEntry
- MemoryInfo
- MetricDef
- OllamaStatus
- PingResult
- PortResult
- ProcessInfo
- ScheduledTask
- SecurityEvent
- ServiceEntry
- SystemInfo
- UserInfo
- EventsOn
- MetricDef
- build_target
- constants.ts
- build.sh
- release-gh.sh
- github.com/shahriarhaqueabir/AllOpsFull
- Network Metrics & Bandwidth Counters
- TimelineEvent
- TestGetDashboardData

## God Nodes (most connected - your core abstractions)
1. `Write-HawkSubHeader()` - 84 edges
2. `Write-HawkError()` - 72 edges
3. `App` - 43 edges
4. `LogWarn()` - 41 edges
5. `SandboxedCommandWithConfig()` - 39 edges
6. `Write-HawkResult()` - 37 edges
7. `NewDataPipeline()` - 34 edges
8. `SystemQuerySandbox()` - 33 edges
9. `NewApp()` - 28 edges
10. `Storage` - 22 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewApp()`  [INFERRED]
  main.go → internal/app/App.go
- `ChatTab()` --calls--> `useBackend()`  [INFERRED]
  cmd/hawkward-gui/frontend/src/pages/AIOps.tsx → cmd/hawkward-gui/frontend/src/hooks/useBackend.ts
- `ReportsTab()` --calls--> `useBackend()`  [INFERRED]
  cmd/hawkward-gui/frontend/src/pages/AIOps.tsx → cmd/hawkward-gui/frontend/src/hooks/useBackend.ts
- `ChatBubble()` --calls--> `cn()`  [INFERRED]
  cmd/hawkward-gui/frontend/src/pages/AIOps.tsx → cmd/hawkward-gui/frontend/src/lib/utils.ts
- `StatusBadge()` --calls--> `cn()`  [INFERRED]
  cmd/hawkward-gui/frontend/src/pages/AIOps.tsx → cmd/hawkward-gui/frontend/src/lib/utils.ts

## Import Cycles
- None detected.

## Communities (126 total, 38 thin omitted)

### Community 0 - "Hawk Management"
Cohesion: 0.08
Nodes (98): Get-HawkActiveConnections(), Get-HawkAdapterBinding(), Get-HawkAdminGroup(), Get-HawkApplicationEvents(), Get-HawkARPTable(), Get-HawkAutoStartServices(), Get-HawkBandwidthUsage(), Get-HawkBiosInfo() (+90 more)

### Community 1 - "Storage & Timeline"
Cohesion: 0.08
Nodes (30): Timeline, LogEntryData, MetricWrite, Storage, DB, T, TestGetLogs(), buildTimelineAnalysisPrompt() (+22 more)

### Community 3 - "Network Lookup & Scan"
Cohesion: 0.06
Nodes (50): DNSResult, PortResult, Client, Context, LookupDNS(), LookupDNSWithContext(), queryCNAME(), queryDNS() (+42 more)

### Community 4 - "Anomaly Detection"
Cohesion: 0.08
Nodes (40): Anomaly, ChatMessage, EnhancedReport, OllamaStatus, ReportSection, SystemStats, Chat(), CheckOllama() (+32 more)

### Community 5 - "Dev Report"
Cohesion: 0.07
Nodes (44): DevReport, FileEntry, LogEntry, ProcessEntry, ShellResult, FileInfo, FileMode, formatSize() (+36 more)

### Community 6 - "Active Job Sandbox"
Cohesion: 0.08
Nodes (40): activeJob, SandboxConfig, tokenPrivileges, Handle, applySandbox(), DefaultSandbox(), Cmd, SandboxedCmd (+32 more)

### Community 7 - "CPU Info Collector"
Cohesion: 0.07
Nodes (34): CPUInfo, buildSystemStats(), CollectAllStats(), SystemInfo, primaryDiskUsage(), GetCPUStats(), T, TestGetCPUStats() (+26 more)

### Community 8 - "Alerts & System Status"
Cohesion: 0.05
Nodes (43): AlertInfo, AlertRuleInfo, AnomalyInfo, AppInfo, BriefingSection, ChatMessage, CommandResult, ConnectionInfo (+35 more)

### Community 9 - "App Information"
Cohesion: 0.11
Nodes (26): AppInfo, PlatformInfo, T, TestConfigDir(), TestDetectPlatform(), TestFixPowerShellDashes(), TestFormatBytes(), TestFormatPercent() (+18 more)

### Community 10 - "Storage & Timeline"
Cohesion: 0.05
Nodes (38): AnomalyInfo, ChatMessage, CommandResult, ConnectionInfo, ConnectionType, CPUInfo, DashboardData, DefenderStatus (+30 more)

### Community 11 - "Dependencies & CSS Classes"
Cohesion: 0.06
Nodes (31): dependencies, class-variance-authority, clsx, date-fns, lucide-react, motion, nanoid, @radix-ui/react-avatar (+23 more)

### Community 12 - "Network Design & Connections"
Cohesion: 0.08
Nodes (11): CanvasMode, deviceTypes, genConnId(), genId(), initialConnections(), initialDevices(), NetworkDesign(), OpenFileDialog() (+3 more)

### Community 13 - "Log Management & DevOps"
Cohesion: 0.14
Nodes (8): DevOps, CommandResult, FileEntry, ProcessInfo, ServiceEntry, NewDevOps(), sanitizeError(), LogWarn()

### Community 14 - "Theme Configuration"
Cohesion: 0.21
Nodes (22): Palette, ThemeName, T, TestAllThemePalettesAreValid(), TestParseThemeName(), CurrentPalette(), CurrentTheme(), LoadTheme() (+14 more)

### Community 15 - "Configuration & Compiler Options"
Cohesion: 0.09
Nodes (22): compilerOptions, allowImportingTsExtensions, baseUrl, isolatedModules, jsx, lib, module, moduleDetection (+14 more)

### Community 16 - "Data Pipeline Configurations"
Cohesion: 0.15
Nodes (10): CollectionConfig, DataPipeline, MetricForecast, WindowStats, NewForecastEngine(), Duration, RWMutex, Time (+2 more)

### Community 17 - "Time Series Data"
Cohesion: 0.15
Nodes (7): DataPoint, TimeSeries, TimeSeriesStore, Time, NewTimeSeries(), NewTimeSeriesStore(), percentile()

### Community 18 - "Service Status & Listeners"
Cohesion: 0.23
Nodes (19): ServiceEntry, listLinuxServices(), ListServices(), listWindowsServices(), parseSCQuery(), parseSystemctlServices(), parseWindowsServicesJSON(), serviceStatusFromNumber() (+11 more)

### Community 19 - "Log Management & Event Bus"
Cohesion: 0.21
Nodes (12): EventBus, EventCategory, EventHandler, EventLevel, TimelineEvent, PingResult, RWMutex, Time (+4 more)

### Community 20 - "Alerts & Conditions"
Cohesion: 0.16
Nodes (7): Alert, AlertCondition, AlertEngine, AlertLevel, AlertRule, RWMutex, Time

### Community 21 - "Dev Dependencies & ESLint Plugins"
Cohesion: 0.11
Nodes (19): devDependencies, eslint, @eslint/js, eslint-plugin-react-hooks, eslint-plugin-react-refresh, globals, jiti, jsdom (+11 more)

### Community 22 - "AIOps.tsx"
Cohesion: 0.22
Nodes (8): AIOps(), ChatBubble(), ChatTab(), ReportsTab(), StatusBadge(), TabId, OllamaState, useOllamaStore

### Community 23 - "Platform Checks & Admin Privileges"
Cohesion: 0.21
Nodes (19): TestPlatformChecks(), IsAdminRequired(), IsLinux(), IsWindows(), ControlService(), isValidServiceName(), SetFirewallRuleState(), collectUserDetails() (+11 more)

### Community 24 - "AIOps Dashboard & Tabs"
Cohesion: 0.12
Nodes (14): AIOps, App, Logs, Dashboard, DevOps, Duration, Time, WaitGroup (+6 more)

### Community 25 - "Metrics Definition & Pipeline API"
Cohesion: 0.14
Nodes (9): MetricDef, PipelineAPI, convertStatsInfo(), convertTrendInfo(), StatsInfo, TrendInfo, NewPipelineAPI(), MetricDef (+1 more)

### Community 26 - "Events Handling & Dashboard"
Cohesion: 0.15
Nodes (11): EventPayload, Runtime, useEvents(), BriefingSection, clamp(), computeRedFlags(), Dashboard(), DiagnosticResult (+3 more)

### Community 27 - "Configuration & Compiler Options"
Cohesion: 0.11
Nodes (17): compilerOptions, allowImportingTsExtensions, isolatedModules, lib, module, moduleDetection, moduleResolution, noEmit (+9 more)

### Community 28 - "New Data Pipeline & Tests"
Cohesion: 0.29
Nodes (17): NewAlertEngine(), T, TestAlertAutoResolveOnRecovery(), TestAlertLessThanCondition(), TestAlertLevelStrings(), TestAlertRuleFiresCorrectly(), TestAlertRuleNoFireBelowThreshold(), TestAllAlerts() (+9 more)

### Community 29 - "New Data Pipeline & Tests"
Cohesion: 0.30
Nodes (17): DefaultCollectionConfig(), NewDataPipeline(), T, TestNewDataPipeline(), TestPipelineAllSeries(), TestPipelineClear(), TestPipelineConcurrency(), TestPipelineConfigDefaults() (+9 more)

### Community 30 - "Tasks Management & Parsing"
Cohesion: 0.29
Nodes (17): findCSVColumn(), formatTaskTime(), GetScheduledTasks(), getTasksLinux(), parseCronFile(), parseCronTab(), parseSystemdTimers(), parseTasksJSON() (+9 more)

### Community 31 - "Wails Configuration & Build Tools"
Cohesion: 0.12
Nodes (16): author, email, name, frontend:build, frontend:dev:serverUrl, frontend:dev:watcher, frontend:dir, frontend:install (+8 more)

### Community 32 - "AIOps"
Cohesion: 0.19
Nodes (7): AnomalyInfo, AIOps, CancelFunc, Context, Duration, NewAIOps(), OllamaStatus

### Community 33 - "SecOps Firewall Rules & Status"
Cohesion: 0.13
Nodes (8): SecOps, DefenderStatus, FirewallRule, ListeningPort, ScheduledTask, SecurityEvent, UserInfo, NewSecOps()

### Community 34 - "Package Information & Repository Details"
Cohesion: 0.12
Nodes (15): author, bugs, url, description, homepage, keywords, license, main (+7 more)

### Community 35 - "Connections Management & Process Mapping"
Cohesion: 0.28
Nodes (15): buildProcInodeMap(), GetConnections(), getConnectionsLinux(), getConnectionsWindows(), getPidMapViaPowerShell(), getPidMapViaWmic(), hexToUint8(), parseHexAddr() (+7 more)

### Community 36 - "Dashboard Briefing Generation"
Cohesion: 0.30
Nodes (9): Dashboard, BriefingSection, DiagnosticResult, cpuDiagMsg(), diagStatus(), diskDiagMsg(), memDiagMsg(), NewDashboard() (+1 more)

### Community 37 - "Dashboard Data Retrieval"
Cohesion: 0.21
Nodes (12): DashboardData, TrendDirection, lastValue(), safeLastN(), T, TestConvertAlert(), TestGetAppInfo(), TestLastValue() (+4 more)

### Community 38 - "SysOps"
Cohesion: 0.20
Nodes (6): SysOps, sysopsFacade, DiskInfo, ProcessInfo, SystemInfo, MemoryInfo

### Community 39 - "App.tsx"
Cohesion: 0.23
Nodes (10): App(), WailsRuntime, MainContent(), pageLabels, TopBar(), TopBarProps, queryClient, useAlertStore (+2 more)

### Community 40 - "NetOps Dashboard & Tabs"
Cohesion: 0.12
Nodes (10): NetOps(), NetOpsTab, StatusBadge(), tabs, SysOps(), AlertState, SettingsState, Theme (+2 more)

### Community 41 - "Forecast Engine & Trend Prediction"
Cohesion: 0.19
Nodes (4): ForecastEngine, TrendDirection, TrendInfo, TrendDirection

### Community 42 - "Defender Configuration & JSON Parsing"
Cohesion: 0.35
Nodes (13): defenderWMICFallback(), formatAge(), formatTimeStr(), GetDefenderStatus(), getJSONBool(), getJSONInt(), getJSONString(), parseDefenderJSON() (+5 more)

### Community 43 - "Alert API & Alerts Management"
Cohesion: 0.22
Nodes (5): AlertAPI, AlertInfo, NewAlertAPI(), convertAlert(), AlertInfo

### Community 44 - "Settings Configuration & Mocks"
Cohesion: 0.15
Nodes (10): AppInfo, DEFAULT_APP_INFO, DEFAULT_SETTINGS, intervalOptions, Settings(), mockSetDnsTimeout, mockSetPingCount, mockSetRefreshInterval (+2 more)

### Community 45 - "Bandwidth Counter & Network Metrics"
Cohesion: 0.20
Nodes (9): BandwidthCounter, NetOps, netOpsModel, BandwidthCounter, ConnectionInfo, InterfaceInfo, Time, NewNetOps() (+1 more)

### Community 46 - "Logs.tsx"
Cohesion: 0.25
Nodes (6): LEVELS, levelStyle, LogBadge(), Logs(), localStorageMock, mockLogs

### Community 47 - "Main Content & System Queries"
Cohesion: 0.17
Nodes (10): AIOps, Dashboard, DevOps, Logs, NetOps, NetworkDesign, pageVariants, SecOps (+2 more)

### Community 48 - "System Query Sandbox & Network Metrics"
Cohesion: 0.41
Nodes (11): SystemQuerySandbox(), extractPort(), extractPortFromSS(), GetListeningPorts(), getListeningPortsLinux(), getProcessNameByPID(), parseListeningPorts(), parseSSOutput() (+3 more)

### Community 49 - "ErrorBoundary.tsx"
Cohesion: 0.20
Nodes (4): ErrorBoundary, Props, State, renderWithProviders()

### Community 50 - "Confirm Dialog & Tabs"
Cohesion: 0.20
Nodes (7): ConfirmDialog(), ConfirmDialogProps, Bar(), pctColor(), SysOpsTab, TabDef, tabs

### Community 51 - "Empty State & Security Tabs"
Cohesion: 0.15
Nodes (4): EmptyStateProps, BackendCall, SecOpsTab, tabs

### Community 53 - "NewApp"
Cohesion: 0.42
Nodes (9): NewApp(), NewSysOps(), T, TestSysOps_GetCPUInfo(), TestSysOps_GetDiskInfo(), TestSysOps_GetMemoryInfo(), TestSysOps_GetSystemInfo(), TestSysOps_GetTopProcesses() (+1 more)

### Community 54 - "Backend Configuration & Anomaly Tab"
Cohesion: 0.18
Nodes (11): Args, useBackend(), AnomaliesTab(), FileBrowserTab(), PowerShellProTab(), ServicesTab(), StatusBadge(), stripAnsi() (+3 more)

### Community 55 - "DashboardData"
Cohesion: 0.20
Nodes (3): DashboardData, GaugeMetric, NetworkMetric

### Community 56 - "Metric History Management"
Cohesion: 0.20
Nodes (3): MetricHistory, StatsInfo, TrendInfo

### Community 57 - "Firewall Rules"
Cohesion: 0.64
Nodes (7): GetFirewallRules(), getFirewallRulesLinux(), getFirewallRulesPowerShell(), parseFirewallRules(), parseIPTablesSave(), parseNFTList(), FirewallRule

### Community 58 - "SecOps Tests & Firewall Rules"
Cohesion: 0.36
Nodes (9): T, TestGetDefenderStatus(), TestGetFirewallRules(), TestGetJSONIntEdgeCases(), TestGetJSONStringWithValueObject(), TestGetListeningPorts(), TestGetUsers(), TestParseFirewallRulesEmpty() (+1 more)

### Community 59 - "Event Stream"
Cohesion: 0.22
Nodes (8): AlertEvent, LogEvent, MetricsEvent, PipelineEvent, TimelineEvent, AlertInfo, GaugeMetric, NetworkMetric

### Community 60 - "Metrics Exporter Initialization"
Cohesion: 0.31
Nodes (6): IncPipelineTick(), SetAlertCountMetric(), SetCPUMetric(), SetDiskMetric(), SetMemoryMetric(), SetProcessCountMetric()

### Community 61 - "Page Components"
Cohesion: 0.31
Nodes (7): Page, MainContentProps, NavItem, opsItems, Sidebar(), SidebarProps, toolsItems

### Community 62 - "Workflows & Security Events"
Cohesion: 0.22
Nodes (9): boolStr(), DefenderStatus, FirewallRule, ListeningPort, ScheduledTask, SecurityEvent, UserInfo, RunSecurityAudit() (+1 more)

### Community 64 - "Sandbox Tests"
Cohesion: 0.47
Nodes (8): SandboxedCommandWithConfig(), T, TestWindowsSandboxDefaultConfig(), TestWindowsSandboxHideWindow(), TestWindowsSandboxJobObjectCreated(), TestWindowsSandboxRestrictedToken(), TestWindowsSandboxSysProcAttrPreserved(), TestWindowsSandboxSystemQuery()

### Community 65 - "LogInfo"
Cohesion: 0.33
Nodes (6): Context, CloseLogger(), InitLogger(), LogError(), LogInfo(), InitMetricsExporter()

### Community 69 - "runtime.d.ts"
Cohesion: 0.25
Nodes (7): EnvironmentInfo, NotificationAction, NotificationCategory, NotificationOptions, Position, Screen, Size

### Community 74 - "scripts"
Cohesion: 0.33
Nodes (6): scripts, build, dev, lint, preview, test

### Community 75 - "Connection Line Visualization"
Cohesion: 0.16
Nodes (11): connectionColors, ConnectionLine(), ConnectionLineProps, deviceColors, deviceIcons, DeviceNode(), DeviceNodeProps, statusDotColors (+3 more)

### Community 76 - "events.go"
Cohesion: 0.73
Nodes (5): formatEventTime(), GetSecurityEvents(), isImportantSecurityEvent(), parseSecurityEventsJSON(), SecurityEvent

### Community 77 - "package.json"
Cohesion: 0.40
Nodes (4): name, private, type, version

### Community 108 - "EventsOn"
Cohesion: 0.67
Nodes (3): EventsOn(), EventsOnce(), EventsOnMultiple()

### Community 122 - "Network Metrics & Bandwidth Counters"
Cohesion: 0.36
Nodes (10): calculateBandwidthRates(), counterDelta(), GetBandwidthCounters(), GetInterfaces(), BandwidthCounter, Duration, BandwidthCounter, bandwidthRate (+2 more)

## Knowledge Gaps
- **274 isolated node(s):** `name`, `private`, `version`, `type`, `dev` (+269 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **38 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `LogWarn()` connect `Log Management & DevOps` to `AIOps`, `LogInfo`, `SecOps Firewall Rules & Status`, `Network Lookup & Scan`, `Storage & Timeline`, `SysOps`, `CPU Info Collector`, `Active Job Sandbox`, `Bandwidth Counter & Network Metrics`, `Log Management & Event Bus`, `AIOps Dashboard & Tabs`, `Metrics Exporter Initialization`?**
  _High betweenness centrality (0.067) - this node is a cross-community bridge._
- **Why does `App` connect `AIOps Dashboard & Tabs` to `AIOps`, `LogInfo`, `Storage & Timeline`, `SecOps Firewall Rules & Status`, `Dashboard Briefing Generation`, `Dashboard Data Retrieval`, `SysOps`, `App Information`, `Alert API & Alerts Management`, `Log Management & DevOps`, `Bandwidth Counter & Network Metrics`, `Data Pipeline Configurations`, `Log Management & Event Bus`, `Alerts & Conditions`, `NewApp`, `Metrics Definition & Pipeline API`, `Metrics Exporter Initialization`?**
  _High betweenness centrality (0.061) - this node is a cross-community bridge._
- **Why does `LogInfo()` connect `LogInfo` to `SecOps Firewall Rules & Status`, `Storage & Timeline`, `Connections Management & Process Mapping`, `Dev Report`, `events.go`, `Log Management & DevOps`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Are the 40 inferred relationships involving `LogWarn()` (e.g. with `.Chat()` and `.collectAndEmit()`) actually correct?**
  _`LogWarn()` has 40 INFERRED edges - model-reasoned connections that need verification._
- **Are the 35 inferred relationships involving `SandboxedCommandWithConfig()` (e.g. with `TestSandboxedCommandWithConfig()` and `TestSandboxedCommandWithConfigDefaults()`) actually correct?**
  _`SandboxedCommandWithConfig()` has 35 INFERRED edges - model-reasoned connections that need verification._
- **What connects `name`, `private`, `version` to the rest of the system?**
  _274 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Hawk Management` be split into smaller, more focused modules?**
  _Cohesion score 0.07919191919191919 - nodes in this community are weakly interconnected._