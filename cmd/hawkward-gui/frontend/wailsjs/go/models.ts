export namespace app {
	
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
	    uptime: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.uptime = source["uptime"];
	    }
	}
	export class CPUInfo {
	    percent: number;
	    per_cpu: number[];
	    model_name: string;
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
	        this.state = source["state"];
	        this.process_name = source["process_name"];
	        this.pid = source["pid"];
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
	
	export class FileEntry {
	    name: string;
	    path: string;
	    size: string;
	    is_dir: boolean;
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
	        this.is_dir = source["is_dir"];
	        this.mode = source["mode"];
	        this.mod_time = source["mod_time"];
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
	export class ListeningPort {
	    port: number;
	    protocol: string;
	    process_name: string;
	    pid: number;
	    state: string;
	
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
	export class MemoryInfo {
	    total_bytes: number;
	    available_bytes: number;
	    used_bytes: number;
	    used_percent: number;
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
	
	export class OllamaStatus {
	    available: boolean;
	    model: string;
	    version: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new OllamaStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.model = source["model"];
	        this.version = source["version"];
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
	        this.name = source["name"];
	        this.cpu = source["cpu"];
	        this.memory = source["memory"];
	        this.mem_pct = source["mem_pct"];
	        this.status = source["status"];
	        this.num_fds = source["num_fds"];
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
	    }
	}

}

