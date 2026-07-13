export namespace app {
	
	export class AIConfidence {
	    overall: number;
	    factors: Record<string, number>;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfidence(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall = source["overall"];
	        this.factors = source["factors"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class AIInsight {
	    category: string;
	    severity: string;
	    title: string;
	    message: string;
	    action: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new AIInsight(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.title = source["title"];
	        this.message = source["message"];
	        this.action = source["action"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class AlertInfo {
	    id: string;
	    level: string;
	    metric: string;
	    message: string;
	    value: number;
	    threshold: number;
	    timestamp: string;
	    resolved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AlertInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.metric = source["metric"];
	        this.message = source["message"];
	        this.value = source["value"];
	        this.threshold = source["threshold"];
	        this.timestamp = source["timestamp"];
	        this.resolved = source["resolved"];
	    }
	}
	export class AlertRuleInfo {
	    metric: string;
	    condition: string;
	    threshold: number;
	    flap_count: number;
	    severity: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new AlertRuleInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metric = source["metric"];
	        this.condition = source["condition"];
	        this.threshold = source["threshold"];
	        this.flap_count = source["flap_count"];
	        this.severity = source["severity"];
	        this.message = source["message"];
	    }
	}
	export class AnomalyInfo {
	    metric: string;
	    value: number;
	    expected: number;
	    deviation: number;
	    severity: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new AnomalyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metric = source["metric"];
	        this.value = source["value"];
	        this.expected = source["expected"];
	        this.deviation = source["deviation"];
	        this.severity = source["severity"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class AppInfo {
	    name: string;
	    version: string;
	    go_version: string;
	    uptime: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.go_version = source["go_version"];
	        this.uptime = source["uptime"];
	    }
	}
	export class BatteryData {
	    percent: number;
	    charging: boolean;
	    time_left_sec: number;
	    status: string;
	    detected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BatteryData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.percent = source["percent"];
	        this.charging = source["charging"];
	        this.time_left_sec = source["time_left_sec"];
	        this.status = source["status"];
	        this.detected = source["detected"];
	    }
	}
	export class BatteryInfo {
	    percent: number;
	    charging: boolean;
	    time_left_sec: number;
	    status: string;
	    detected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BatteryInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.percent = source["percent"];
	        this.charging = source["charging"];
	        this.time_left_sec = source["time_left_sec"];
	        this.status = source["status"];
	        this.detected = source["detected"];
	    }
	}
	export class BriefingSection {
	    title: string;
	    content: string;
	    level: string;
	
	    static createFrom(source: any = {}) {
	        return new BriefingSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.content = source["content"];
	        this.level = source["level"];
	    }
	}
	export class CPUInfo {
	    percent: number;
	    per_cpu: number[];
	    model_name: string;
	    logical_cores: number;
	    physical_cores: number;
	    core_count: number;
	    load_avg_1: number;
	    load_avg_5: number;
	    load_avg_15: number;
	
	    static createFrom(source: any = {}) {
	        return new CPUInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.percent = source["percent"];
	        this.per_cpu = source["per_cpu"];
	        this.model_name = source["model_name"];
	        this.logical_cores = source["logical_cores"];
	        this.physical_cores = source["physical_cores"];
	        this.core_count = source["core_count"];
	        this.load_avg_1 = source["load_avg_1"];
	        this.load_avg_5 = source["load_avg_5"];
	        this.load_avg_15 = source["load_avg_15"];
	    }
	}
	export class CommandResult {
	    command: string;
	    output: string;
	    exit_code: number;
	    duration_ms: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.output = source["output"];
	        this.exit_code = source["exit_code"];
	        this.duration_ms = source["duration_ms"];
	        this.error = source["error"];
	    }
	}
	export class ConnectionInfo {
	    local_addr: string;
	    remote_addr: string;
	    local_port: number;
	    remote_port: number;
	    protocol: string;
	    state: string;
	    process_name: string;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.local_addr = source["local_addr"];
	        this.remote_addr = source["remote_addr"];
	        this.local_port = source["local_port"];
	        this.remote_port = source["remote_port"];
	        this.protocol = source["protocol"];
	        this.state = source["state"];
	        this.process_name = source["process_name"];
	        this.pid = source["pid"];
	    }
	}
	export class ContainerInfo {
	    id: string;
	    name: string;
	    image: string;
	    state: string;
	    status: string;
	    ports: string;
	
	    static createFrom(source: any = {}) {
	        return new ContainerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.image = source["image"];
	        this.state = source["state"];
	        this.status = source["status"];
	        this.ports = source["ports"];
	    }
	}
	export class ContainerSummary {
	    running: number;
	    stopped: number;
	    failed: number;
	    total: number;
	    containers: ContainerInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ContainerSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.stopped = source["stopped"];
	        this.failed = source["failed"];
	        this.total = source["total"];
	        this.containers = this.convertValues(source["containers"], ContainerInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConversationMessage {
	    id: number;
	    session_id: string;
	    role: string;
	    content: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new ConversationMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.session_id = source["session_id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DNSResult {
	    hostname: string;
	    a: string[];
	    aaaa: string[];
	    mx: string[];
	    ns: string[];
	    cname: string;
	    txt: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DNSResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.a = source["a"];
	        this.aaaa = source["aaaa"];
	        this.mx = source["mx"];
	        this.ns = source["ns"];
	        this.cname = source["cname"];
	        this.txt = source["txt"];
	        this.error = source["error"];
	    }
	}
	export class NetworkMetric {
	    rx_rate: number;
	    tx_rate: number;
	    unit: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkMetric(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rx_rate = source["rx_rate"];
	        this.tx_rate = source["tx_rate"];
	        this.unit = source["unit"];
	    }
	}
	export class GPUData {
	    name: string;
	    vendor: string;
	    memory_gb: number;
	    driver: string;
	    detected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GPUData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.vendor = source["vendor"];
	        this.memory_gb = source["memory_gb"];
	        this.driver = source["driver"];
	        this.detected = source["detected"];
	    }
	}
	export class GaugeMetric {
	    value: number;
	    unit: string;
	    history: number[];
	    forecast: number[];
	    trend: string;
	
	    static createFrom(source: any = {}) {
	        return new GaugeMetric(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.history = source["history"];
	        this.forecast = source["forecast"];
	        this.trend = source["trend"];
	    }
	}
	export class DashboardData {
	    cpu: GaugeMetric;
	    memory: GaugeMetric;
	    disk: GaugeMetric;
	    gpu: GPUData;
	    battery: BatteryData;
	    network: NetworkMetric;
	    processes: number;
	    connections: number;
	    alerts: number;
	    uptime: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu = this.convertValues(source["cpu"], GaugeMetric);
	        this.memory = this.convertValues(source["memory"], GaugeMetric);
	        this.disk = this.convertValues(source["disk"], GaugeMetric);
	        this.gpu = this.convertValues(source["gpu"], GPUData);
	        this.battery = this.convertValues(source["battery"], BatteryData);
	        this.network = this.convertValues(source["network"], NetworkMetric);
	        this.processes = source["processes"];
	        this.connections = source["connections"];
	        this.alerts = source["alerts"];
	        this.uptime = source["uptime"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DefenderStatus {
	    enabled: boolean;
	    up_to_date: boolean;
	    signature_age: string;
	    last_scan: string;
	    real_time_protection: boolean;
	    cloud_protection: boolean;
	    am_service_enabled: boolean;
	    antispyware_enabled: boolean;
	    nis_enabled: boolean;
	    quick_scan_age: number;
	    full_scan_age: number;
	    threats_detected: number;
	
	    static createFrom(source: any = {}) {
	        return new DefenderStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.up_to_date = source["up_to_date"];
	        this.signature_age = source["signature_age"];
	        this.last_scan = source["last_scan"];
	        this.real_time_protection = source["real_time_protection"];
	        this.cloud_protection = source["cloud_protection"];
	        this.am_service_enabled = source["am_service_enabled"];
	        this.antispyware_enabled = source["antispyware_enabled"];
	        this.nis_enabled = source["nis_enabled"];
	        this.quick_scan_age = source["quick_scan_age"];
	        this.full_scan_age = source["full_scan_age"];
	        this.threats_detected = source["threats_detected"];
	    }
	}
	export class DevOpsSuggestion {
	    category: string;
	    severity: string;
	    message: string;
	    action: string;
	
	    static createFrom(source: any = {}) {
	        return new DevOpsSuggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.message = source["message"];
	        this.action = source["action"];
	    }
	}
	export class DiagnosticResult {
	    category: string;
	    status: string;
	    message: string;
	    value: number;
	    unit: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	    }
	}
	export class DiskPartition {
	    mountpoint: string;
	    total_bytes: number;
	    free_bytes: number;
	    used_bytes: number;
	    used_percent: number;
	    fs_type: string;
	    device: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskPartition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mountpoint = source["mountpoint"];
	        this.total_bytes = source["total_bytes"];
	        this.free_bytes = source["free_bytes"];
	        this.used_bytes = source["used_bytes"];
	        this.used_percent = source["used_percent"];
	        this.fs_type = source["fs_type"];
	        this.device = source["device"];
	    }
	}
	export class DiskInfo {
	    partitions: DiskPartition[];
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.partitions = this.convertValues(source["partitions"], DiskPartition);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class DockerStatus {
	    installed: boolean;
	    running: boolean;
	    version: string;
	    containers: ContainerSummary;
	
	    static createFrom(source: any = {}) {
	        return new DockerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.running = source["running"];
	        this.version = source["version"];
	        this.containers = this.convertValues(source["containers"], ContainerSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EnvVarInfo {
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvVarInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class ToolVersion {
	    name: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	    }
	}
	export class EnvironmentInfo {
	    path_dirs: string[];
	    key_vars: EnvVarInfo[];
	    sdks: ToolVersion[];
	    package_managers: ToolVersion[];
	
	    static createFrom(source: any = {}) {
	        return new EnvironmentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path_dirs = source["path_dirs"];
	        this.key_vars = this.convertValues(source["key_vars"], EnvVarInfo);
	        this.sdks = this.convertValues(source["sdks"], ToolVersion);
	        this.package_managers = this.convertValues(source["package_managers"], ToolVersion);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileEntry {
	    name: string;
	    path: string;
	    size: string;
	    raw_size: number;
	    is_dir: boolean;
	    is_binary: boolean;
	    mode: string;
	    mod_time: string;
	
	    static createFrom(source: any = {}) {
	        return new FileEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.raw_size = source["raw_size"];
	        this.is_dir = source["is_dir"];
	        this.is_binary = source["is_binary"];
	        this.mode = source["mode"];
	        this.mod_time = source["mod_time"];
	    }
	}
	export class FirewallProfile {
	    name: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FirewallProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	    }
	}
	export class FirewallRule {
	    name: string;
	    direction: string;
	    action: string;
	    protocol: string;
	    local_port: string;
	    remote_port: string;
	    remote_ip: string;
	    profile: string;
	    enabled: boolean;
	    is_high_risk: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FirewallRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.direction = source["direction"];
	        this.action = source["action"];
	        this.protocol = source["protocol"];
	        this.local_port = source["local_port"];
	        this.remote_port = source["remote_port"];
	        this.remote_ip = source["remote_ip"];
	        this.profile = source["profile"];
	        this.enabled = source["enabled"];
	        this.is_high_risk = source["is_high_risk"];
	    }
	}
	export class FirewallStatus {
	    enabled: boolean;
	    profiles: FirewallProfile[];
	
	    static createFrom(source: any = {}) {
	        return new FirewallStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.profiles = this.convertValues(source["profiles"], FirewallProfile);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class GPUInfo {
	    name: string;
	    vendor: string;
	    memory_gb: number;
	    driver: string;
	    detected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GPUInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.vendor = source["vendor"];
	        this.memory_gb = source["memory_gb"];
	        this.driver = source["driver"];
	        this.detected = source["detected"];
	    }
	}
	export class GatewayInfo {
	    ip: string;
	    interface: string;
	    reachable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GatewayInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.interface = source["interface"];
	        this.reachable = source["reachable"];
	    }
	}
	
	export class GitRepoInfo {
	    path: string;
	    branch: string;
	    modified_files: number;
	    untracked_files: number;
	    ahead: number;
	    behind: number;
	    clean: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GitRepoInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.branch = source["branch"];
	        this.modified_files = source["modified_files"];
	        this.untracked_files = source["untracked_files"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.clean = source["clean"];
	    }
	}
	export class GitSummary {
	    repositories: GitRepoInfo[];
	    total_repos: number;
	
	    static createFrom(source: any = {}) {
	        return new GitSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repositories = this.convertValues(source["repositories"], GitRepoInfo);
	        this.total_repos = source["total_repos"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InterfaceInfo {
	    name: string;
	    mac: string;
	    ips: string[];
	    is_up: boolean;
	    speed: string;
	    mtu: number;
	    flags: string;
	    rx_bytes: number;
	    tx_bytes: number;
	    rx_rate_bps: number;
	    tx_rate_bps: number;
	    rx_history: number[];
	    tx_history: number[];
	
	    static createFrom(source: any = {}) {
	        return new InterfaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.mac = source["mac"];
	        this.ips = source["ips"];
	        this.is_up = source["is_up"];
	        this.speed = source["speed"];
	        this.mtu = source["mtu"];
	        this.flags = source["flags"];
	        this.rx_bytes = source["rx_bytes"];
	        this.tx_bytes = source["tx_bytes"];
	        this.rx_rate_bps = source["rx_rate_bps"];
	        this.tx_rate_bps = source["tx_rate_bps"];
	        this.rx_history = source["rx_history"];
	        this.tx_history = source["tx_history"];
	    }
	}
	export class KubernetesStatus {
	    installed: boolean;
	    connected: boolean;
	    cluster: string;
	    nodes: number;
	    pods: number;
	
	    static createFrom(source: any = {}) {
	        return new KubernetesStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.connected = source["connected"];
	        this.cluster = source["cluster"];
	        this.nodes = source["nodes"];
	        this.pods = source["pods"];
	    }
	}
	export class LearnedBaseline {
	    metric: string;
	    mean: number;
	    min: number;
	    max: number;
	    stdDev: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new LearnedBaseline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metric = source["metric"];
	        this.mean = source["mean"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.stdDev = source["stdDev"];
	        this.count = source["count"];
	    }
	}
	export class ListeningPort {
	    port: number;
	    protocol: string;
	    process_name: string;
	    pid: number;
	    state: string;
	    is_external: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ListeningPort(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.process_name = source["process_name"];
	        this.pid = source["pid"];
	        this.state = source["state"];
	        this.is_external = source["is_external"];
	    }
	}
	export class LocalServer {
	    port: number;
	    protocol: string;
	    process: string;
	    pid: number;
	    framework: string;
	    health: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.process = source["process"];
	        this.pid = source["pid"];
	        this.framework = source["framework"];
	        this.health = source["health"];
	    }
	}
	export class LogEntry {
	    timestamp: string;
	    level: string;
	    module: string;
	    message: string;
	    line: string;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.module = source["module"];
	        this.message = source["message"];
	        this.line = source["line"];
	        this.source = source["source"];
	    }
	}
	export class LogSourceCount {
	    source: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new LogSourceCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.count = source["count"];
	    }
	}
	export class TrendingError {
	    message: string;
	    count: number;
	    lastSeen: string;
	
	    static createFrom(source: any = {}) {
	        return new TrendingError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.count = source["count"];
	        this.lastSeen = source["lastSeen"];
	    }
	}
	export class LogStats {
	    totalToday: number;
	    totalThisHour: number;
	    totalLastMin: number;
	    errorCount: number;
	    warningCount: number;
	    infoCount: number;
	    debugCount: number;
	    topSources: LogSourceCount[];
	    trendingErrors: TrendingError[];
	
	    static createFrom(source: any = {}) {
	        return new LogStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalToday = source["totalToday"];
	        this.totalThisHour = source["totalThisHour"];
	        this.totalLastMin = source["totalLastMin"];
	        this.errorCount = source["errorCount"];
	        this.warningCount = source["warningCount"];
	        this.infoCount = source["infoCount"];
	        this.debugCount = source["debugCount"];
	        this.topSources = this.convertValues(source["topSources"], LogSourceCount);
	        this.trendingErrors = this.convertValues(source["trendingErrors"], TrendingError);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LogSummary {
	    topSource: string;
	    topMessage: string;
	    errorTrend: string;
	    summaryText: string;
	
	    static createFrom(source: any = {}) {
	        return new LogSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.topSource = source["topSource"];
	        this.topMessage = source["topMessage"];
	        this.errorTrend = source["errorTrend"];
	        this.summaryText = source["summaryText"];
	    }
	}
	export class LogTimelinePoint {
	    timestamp: string;
	    total: number;
	    errors: number;
	    warnings: number;
	    info: number;
	
	    static createFrom(source: any = {}) {
	        return new LogTimelinePoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.total = source["total"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	        this.info = source["info"];
	    }
	}
	export class MemoryInfo {
	    total_bytes: number;
	    available_bytes: number;
	    used_bytes: number;
	    used_percent: number;
	    cached_bytes: number;
	    total_gb: number;
	    used_gb: number;
	    swap_total: number;
	    swap_used: number;
	    swap_percent: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoryInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_bytes = source["total_bytes"];
	        this.available_bytes = source["available_bytes"];
	        this.used_bytes = source["used_bytes"];
	        this.used_percent = source["used_percent"];
	        this.cached_bytes = source["cached_bytes"];
	        this.total_gb = source["total_gb"];
	        this.used_gb = source["used_gb"];
	        this.swap_total = source["swap_total"];
	        this.swap_used = source["swap_used"];
	        this.swap_percent = source["swap_percent"];
	    }
	}
	export class MetricDef {
	    name: string;
	    unit: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new MetricDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.label = source["label"];
	    }
	}
	export class StatsInfo {
	    min: number;
	    max: number;
	    avg: number;
	    p50: number;
	    p95: number;
	    p99: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new StatsInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.min = source["min"];
	        this.max = source["max"];
	        this.avg = source["avg"];
	        this.p50 = source["p50"];
	        this.p95 = source["p95"];
	        this.p99 = source["p99"];
	        this.count = source["count"];
	    }
	}
	export class TrendInfo {
	    direction: string;
	    change_pct: number;
	    slope: number;
	    correlation: number;
	
	    static createFrom(source: any = {}) {
	        return new TrendInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.direction = source["direction"];
	        this.change_pct = source["change_pct"];
	        this.slope = source["slope"];
	        this.correlation = source["correlation"];
	    }
	}
	export class MetricHistory {
	    name: string;
	    unit: string;
	    values: number[];
	    forecast: number[];
	    trend: TrendInfo;
	    stats: StatsInfo;
	    last_value: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.values = source["values"];
	        this.forecast = source["forecast"];
	        this.trend = this.convertValues(source["trend"], TrendInfo);
	        this.stats = this.convertValues(source["stats"], StatsInfo);
	        this.last_value = source["last_value"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NetworkChange {
	    type: string;
	    interface: string;
	    detail: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.interface = source["interface"];
	        this.detail = source["detail"];
	        this.timestamp = source["timestamp"];
	    }
	}
	
	export class NetworkSummary {
	    summaryText: string;
	    topInterface: string;
	    issues: string[];
	
	    static createFrom(source: any = {}) {
	        return new NetworkSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summaryText = source["summaryText"];
	        this.topInterface = source["topInterface"];
	        this.issues = source["issues"];
	    }
	}
	export class OllamaStatus {
	    available: boolean;
	    model: string;
	    version: string;
	    available_models: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new OllamaStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.model = source["model"];
	        this.version = source["version"];
	        this.available_models = source["available_models"];
	        this.error = source["error"];
	    }
	}
	export class PingResult {
	    target: string;
	    ip: string;
	    sent: number;
	    received: number;
	    lost: number;
	    min_ms: number;
	    max_ms: number;
	    avg_ms: number;
	    jitter_ms: number;
	    ttl: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new PingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.ip = source["ip"];
	        this.sent = source["sent"];
	        this.received = source["received"];
	        this.lost = source["lost"];
	        this.min_ms = source["min_ms"];
	        this.max_ms = source["max_ms"];
	        this.avg_ms = source["avg_ms"];
	        this.jitter_ms = source["jitter_ms"];
	        this.ttl = source["ttl"];
	        this.error = source["error"];
	    }
	}
	export class PortResult {
	    port: number;
	    open: boolean;
	    service: string;
	
	    static createFrom(source: any = {}) {
	        return new PortResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.open = source["open"];
	        this.service = source["service"];
	    }
	}
	export class ProcessInfo {
	    pid: number;
	    ppid: number;
	    name: string;
	    cpu: number;
	    memory: number;
	    mem_pct: number;
	    status: string;
	    num_fds: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.ppid = source["ppid"];
	        this.name = source["name"];
	        this.cpu = source["cpu"];
	        this.memory = source["memory"];
	        this.mem_pct = source["mem_pct"];
	        this.status = source["status"];
	        this.num_fds = source["num_fds"];
	    }
	}
	export class RiskInfo {
	    category: string;
	    severity: string;
	    title: string;
	    description: string;
	    recommendation: string;
	
	    static createFrom(source: any = {}) {
	        return new RiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.recommendation = source["recommendation"];
	    }
	}
	export class ScheduledTask {
	    name: string;
	    status: string;
	    next_run: string;
	    last_run: string;
	    author: string;
	    trigger: string;
	
	    static createFrom(source: any = {}) {
	        return new ScheduledTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.next_run = source["next_run"];
	        this.last_run = source["last_run"];
	        this.author = source["author"];
	        this.trigger = source["trigger"];
	    }
	}
	export class SecurityEvent {
	    id: number;
	    level: string;
	    provider: string;
	    time: string;
	    message: string;
	    important: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SecurityEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.provider = source["provider"];
	        this.time = source["time"];
	        this.message = source["message"];
	        this.important = source["important"];
	    }
	}
	export class SecurityScore {
	    score: number;
	    grade: string;
	    breakdown: Record<string, number>;
	    recommendations: string[];
	
	    static createFrom(source: any = {}) {
	        return new SecurityScore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.grade = source["grade"];
	        this.breakdown = source["breakdown"];
	        this.recommendations = source["recommendations"];
	    }
	}
	export class SecuritySummary {
	    score: number;
	    summary: string;
	    risks: string[];
	    recommendations: string[];
	    analyzedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SecuritySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.summary = source["summary"];
	        this.risks = source["risks"];
	        this.recommendations = source["recommendations"];
	        this.analyzedAt = source["analyzedAt"];
	    }
	}
	export class ServiceInfo {
	    name: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	    }
	}
	export class ServiceCategory {
	    category: string;
	    services: ServiceInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ServiceCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.services = this.convertValues(source["services"], ServiceInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ServiceEntry {
	    name: string;
	    display_name: string;
	    status: string;
	    start_type: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display_name = source["display_name"];
	        this.status = source["status"];
	        this.start_type = source["start_type"];
	    }
	}
	export class ServiceGroupSummary {
	    databases: number;
	    messageQueues: number;
	    webServers: number;
	    containers: number;
	    other: number;
	    running: number;
	    stopped: number;
	
	    static createFrom(source: any = {}) {
	        return new ServiceGroupSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.databases = source["databases"];
	        this.messageQueues = source["messageQueues"];
	        this.webServers = source["webServers"];
	        this.containers = source["containers"];
	        this.other = source["other"];
	        this.running = source["running"];
	        this.stopped = source["stopped"];
	    }
	}
	
	
	export class SystemInfo {
	    hostname: string;
	    os: string;
	    platform: string;
	    platform_version: string;
	    kernel_version: string;
	    kernel_arch: string;
	    uptime: string;
	    process_count: number;
	    virtualization: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.os = source["os"];
	        this.platform = source["platform"];
	        this.platform_version = source["platform_version"];
	        this.kernel_version = source["kernel_version"];
	        this.kernel_arch = source["kernel_arch"];
	        this.uptime = source["uptime"];
	        this.process_count = source["process_count"];
	        this.virtualization = source["virtualization"];
	    }
	}
	export class SystemRecommendation {
	    category: string;
	    severity: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemRecommendation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.severity = source["severity"];
	        this.message = source["message"];
	    }
	}
	export class TimelineEvent {
	    id: string;
	    timestamp: string;
	    category: string;
	    level: string;
	    title: string;
	    detail: string;
	    module: string;
	    related?: string[];
	    metadata?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new TimelineEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.category = source["category"];
	        this.level = source["level"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.module = source["module"];
	        this.related = source["related"];
	        this.metadata = source["metadata"];
	    }
	}
	export class ToolInfo {
	    name: string;
	    version: string;
	    path: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.status = source["status"];
	    }
	}
	
	export class TraceHop {
	    number: number;
	    host: string;
	    ip: string;
	    rtts_ms: number[];
	    timed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TraceHop(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.number = source["number"];
	        this.host = source["host"];
	        this.ip = source["ip"];
	        this.rtts_ms = source["rtts_ms"];
	        this.timed = source["timed"];
	    }
	}
	export class TraceResult {
	    target: string;
	    hops: TraceHop[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TraceResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.hops = this.convertValues(source["hops"], TraceHop);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class UserInfo {
	    username: string;
	    full_name: string;
	    sid: string;
	    group: string;
	    is_admin: boolean;
	    is_enabled: boolean;
	    password_never_expires: boolean;
	    last_logon: string;
	
	    static createFrom(source: any = {}) {
	        return new UserInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.full_name = source["full_name"];
	        this.sid = source["sid"];
	        this.group = source["group"];
	        this.is_admin = source["is_admin"];
	        this.is_enabled = source["is_enabled"];
	        this.password_never_expires = source["password_never_expires"];
	        this.last_logon = source["last_logon"];
	    }
	}

}

export namespace common {
	
	export class CollectorStatus {
	    id: string;
	    name: string;
	    description: string;
	    enabled: boolean;
	    interval_ms: number;
	    default_interval_ms: number;
	    last_run: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectorStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	        this.interval_ms = source["interval_ms"];
	        this.default_interval_ms = source["default_interval_ms"];
	        this.last_run = source["last_run"];
	    }
	}
	export class DataPoint {
	    // Go type: time
	    Time: any;
	    Value: number;
	
	    static createFrom(source: any = {}) {
	        return new DataPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Value = source["Value"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace networkdesign {
	
	export class DuplicateIP {
	    ip: string;
	    nodes: string[];
	
	    static createFrom(source: any = {}) {
	        return new DuplicateIP(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.nodes = source["nodes"];
	    }
	}
	export class TopologyHealth {
	    totalNodes: number;
	    totalEdges: number;
	    brokenLinks: number;
	    missingLabels: number;
	    orphanNodes: string[];
	    duplicateIPs: DuplicateIP[];
	    subnetErrors: string[];
	    suggestions: string[];
	
	    static createFrom(source: any = {}) {
	        return new TopologyHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalNodes = source["totalNodes"];
	        this.totalEdges = source["totalEdges"];
	        this.brokenLinks = source["brokenLinks"];
	        this.missingLabels = source["missingLabels"];
	        this.orphanNodes = source["orphanNodes"];
	        this.duplicateIPs = this.convertValues(source["duplicateIPs"], DuplicateIP);
	        this.subnetErrors = source["subnetErrors"];
	        this.suggestions = source["suggestions"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TopologyNode {
	    id: string;
	    type: string;
	    label: string;
	    ip: string;
	    subnet: string;
	    mac: string;
	    notes: string;
	    vendor: string;
	    vlan: string;
	    online: boolean;
	    props: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new TopologyNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.label = source["label"];
	        this.ip = source["ip"];
	        this.subnet = source["subnet"];
	        this.mac = source["mac"];
	        this.notes = source["notes"];
	        this.vendor = source["vendor"];
	        this.vlan = source["vlan"];
	        this.online = source["online"];
	        this.props = source["props"];
	    }
	}

}

