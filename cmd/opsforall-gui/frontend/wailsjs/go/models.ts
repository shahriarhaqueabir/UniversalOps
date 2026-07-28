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
	    actionPage: string;
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
	        this.actionPage = source["actionPage"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class AIOpsSummary {
	    ollamaAvailable: boolean;
	    ollamaModel: string;
	    anomalyCount: number;
	    criticalAnomalies: number;
	    recentInsights: AIInsight[];
	
	    static createFrom(source: any = {}) {
	        return new AIOpsSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ollamaAvailable = source["ollamaAvailable"];
	        this.ollamaModel = source["ollamaModel"];
	        this.anomalyCount = source["anomalyCount"];
	        this.criticalAnomalies = source["criticalAnomalies"];
	        this.recentInsights = this.convertValues(source["recentInsights"], AIInsight);
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
	export class AuditCheckItem {
	    category: string;
	    check: string;
	    passed: boolean;
	    description: string;
	    remediation: string;
	
	    static createFrom(source: any = {}) {
	        return new AuditCheckItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.check = source["check"];
	        this.passed = source["passed"];
	        this.description = source["description"];
	        this.remediation = source["remediation"];
	    }
	}
	export class BaseboardInfo {
	    manufacturer: string;
	    product: string;
	    version: string;
	    serial_number: string;
	
	    static createFrom(source: any = {}) {
	        return new BaseboardInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.manufacturer = source["manufacturer"];
	        this.product = source["product"];
	        this.version = source["version"];
	        this.serial_number = source["serial_number"];
	    }
	}
	export class BaselineSnapshot {
	    // Go type: time
	    timestamp: any;
	    hardware: any;
	    software: any;
	    network: any;
	    security: any;
	    health: any;
	
	    static createFrom(source: any = {}) {
	        return new BaselineSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.hardware = source["hardware"];
	        this.software = source["software"];
	        this.network = source["network"];
	        this.security = source["security"];
	        this.health = source["health"];
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
	export class PerCPUInfoData {
	    core: number;
	    frequency_mhz: number;
	    usage_percent: number;
	
	    static createFrom(source: any = {}) {
	        return new PerCPUInfoData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.core = source["core"];
	        this.frequency_mhz = source["frequency_mhz"];
	        this.usage_percent = source["usage_percent"];
	    }
	}
	export class CPUExtendedInfo {
	    model_name: string;
	    frequency_mhz: number;
	    cache_size_kb: number;
	    temperature: number;
	    per_cpu_info: PerCPUInfoData[];
	
	    static createFrom(source: any = {}) {
	        return new CPUExtendedInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model_name = source["model_name"];
	        this.frequency_mhz = source["frequency_mhz"];
	        this.cache_size_kb = source["cache_size_kb"];
	        this.temperature = source["temperature"];
	        this.per_cpu_info = this.convertValues(source["per_cpu_info"], PerCPUInfoData);
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
	export class CPUTimesData {
	    user: number;
	    system: number;
	    idle: number;
	    iowait: number;
	    steal: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new CPUTimesData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user = source["user"];
	        this.system = source["system"];
	        this.idle = source["idle"];
	        this.iowait = source["iowait"];
	        this.steal = source["steal"];
	        this.total = source["total"];
	    }
	}
	export class ChatResponse {
	    content: string;
	    actions?: common.ActionPreview[];
	    payload?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ChatResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.actions = this.convertValues(source["actions"], common.ActionPreview);
	        this.payload = source["payload"];
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
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversationMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.session_id = source["session_id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
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
	export class HealthScorePoint {
	    day: string;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new HealthScorePoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.score = source["score"];
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
	export class GPUInfo {
	    name: string;
	    vendor: string;
	    memory_gb: number;
	    driver: string;
	    detected: boolean;
	    temperature: number;
	    utilization: number;
	    fan_speed: number;
	
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
	        this.temperature = source["temperature"];
	        this.utilization = source["utilization"];
	        this.fan_speed = source["fan_speed"];
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
	    gpu: GPUInfo;
	    battery: BatteryInfo;
	    network: NetworkMetric;
	    processes: number;
	    connections: number;
	    alerts: number;
	    uptime: string;
	    health_score: number;
	    health_trend: HealthScorePoint[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu = this.convertValues(source["cpu"], GaugeMetric);
	        this.memory = this.convertValues(source["memory"], GaugeMetric);
	        this.disk = this.convertValues(source["disk"], GaugeMetric);
	        this.gpu = this.convertValues(source["gpu"], GPUInfo);
	        this.battery = this.convertValues(source["battery"], BatteryInfo);
	        this.network = this.convertValues(source["network"], NetworkMetric);
	        this.processes = source["processes"];
	        this.connections = source["connections"];
	        this.alerts = source["alerts"];
	        this.uptime = source["uptime"];
	        this.health_score = source["health_score"];
	        this.health_trend = this.convertValues(source["health_trend"], HealthScorePoint);
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
	export class DataPoint {
	    time: string;
	    value: number;
	
	    static createFrom(source: any = {}) {
	        return new DataPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.value = source["value"];
	    }
	}
	export class DataStreamMetric {
	    name: string;
	    unit: string;
	    lastValue: number;
	    samples: number;
	    trend: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DataStreamMetric(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.lastValue = source["lastValue"];
	        this.samples = source["samples"];
	        this.trend = source["trend"];
	        this.updatedAt = source["updatedAt"];
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
	export class DevOpsDiagCheck {
	    name: string;
	    status: string;
	    message: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new DevOpsDiagCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.value = source["value"];
	    }
	}
	export class DevOpsDiagResult {
	    checks: DevOpsDiagCheck[];
	    score: number;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new DevOpsDiagResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.checks = this.convertValues(source["checks"], DevOpsDiagCheck);
	        this.score = source["score"];
	        this.timestamp = source["timestamp"];
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
	export class DevOpsSummary {
	    serviceCount: number;
	    runningCount: number;
	    dockerInstalled: boolean;
	    dockerRunning: boolean;
	    containerCount: number;
	    k8sInstalled: boolean;
	    k8sConnected: boolean;
	    k8sPods: number;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new DevOpsSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceCount = source["serviceCount"];
	        this.runningCount = source["runningCount"];
	        this.dockerInstalled = source["dockerInstalled"];
	        this.dockerRunning = source["dockerRunning"];
	        this.containerCount = source["containerCount"];
	        this.k8sInstalled = source["k8sInstalled"];
	        this.k8sConnected = source["k8sConnected"];
	        this.k8sPods = source["k8sPods"];
	        this.summary = source["summary"];
	    }
	}
	export class DiagnosticCheckData {
	    name: string;
	    status: string;
	    message: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticCheckData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.value = source["value"];
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
	export class DiscoveryTemplateData {
	    id: string;
	    name: string;
	    description: string;
	    run_ping: boolean;
	    run_dns: boolean;
	    run_trace: boolean;
	    run_arp: boolean;
	    run_routing: boolean;
	    run_port_scan: boolean;
	    ping_count: number;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveryTemplateData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.run_ping = source["run_ping"];
	        this.run_dns = source["run_dns"];
	        this.run_trace = source["run_trace"];
	        this.run_arp = source["run_arp"];
	        this.run_routing = source["run_routing"];
	        this.run_port_scan = source["run_port_scan"];
	        this.ping_count = source["ping_count"];
	    }
	}
	export class DiskEncryption {
	    volume: string;
	    encrypted: boolean;
	    method: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskEncryption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.volume = source["volume"];
	        this.encrypted = source["encrypted"];
	        this.method = source["method"];
	        this.status = source["status"];
	    }
	}
	export class DiskIOEntry {
	    name: string;
	    read_bytes: number;
	    write_bytes: number;
	    read_count: number;
	    write_count: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskIOEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.read_bytes = source["read_bytes"];
	        this.write_bytes = source["write_bytes"];
	        this.read_count = source["read_count"];
	        this.write_count = source["write_count"];
	    }
	}
	export class DiskIOData {
	    disks: DiskIOEntry[];
	    total_read_bytes: number;
	    total_write_bytes: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskIOData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.disks = this.convertValues(source["disks"], DiskIOEntry);
	        this.total_read_bytes = source["total_read_bytes"];
	        this.total_write_bytes = source["total_write_bytes"];
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
	
	export class DoHResultData {
	    server: string;
	    latency_ms: number;
	    success: boolean;
	    resolved_ip: string;
	
	    static createFrom(source: any = {}) {
	        return new DoHResultData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server = source["server"];
	        this.latency_ms = source["latency_ms"];
	        this.success = source["success"];
	        this.resolved_ip = source["resolved_ip"];
	    }
	}
	export class DockerActionResult {
	    action: string;
	    message: string;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DockerActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.message = source["message"];
	        this.success = source["success"];
	    }
	}
	export class DockerComposeService {
	    name: string;
	    state: string;
	    ports: string;
	
	    static createFrom(source: any = {}) {
	        return new DockerComposeService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.state = source["state"];
	        this.ports = source["ports"];
	    }
	}
	export class DockerComposeProject {
	    project: string;
	    status: string;
	    work_dir: string;
	    services: DockerComposeService[];
	
	    static createFrom(source: any = {}) {
	        return new DockerComposeProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = source["project"];
	        this.status = source["status"];
	        this.work_dir = source["work_dir"];
	        this.services = this.convertValues(source["services"], DockerComposeService);
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
	
	export class DockerNetworkInfo {
	    id: string;
	    name: string;
	    driver: string;
	    scope: string;
	    subnet: string;
	    gateway: string;
	    containers: number;
	
	    static createFrom(source: any = {}) {
	        return new DockerNetworkInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.driver = source["driver"];
	        this.scope = source["scope"];
	        this.subnet = source["subnet"];
	        this.gateway = source["gateway"];
	        this.containers = source["containers"];
	    }
	}
	export class DockerStatsEntry {
	    container_id: string;
	    name: string;
	    cpu_percent: string;
	    memory_usage: string;
	    memory_limit: string;
	    memory_percent: string;
	    net_io: string;
	    block_io: string;
	    pid_count: string;
	
	    static createFrom(source: any = {}) {
	        return new DockerStatsEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.container_id = source["container_id"];
	        this.name = source["name"];
	        this.cpu_percent = source["cpu_percent"];
	        this.memory_usage = source["memory_usage"];
	        this.memory_limit = source["memory_limit"];
	        this.memory_percent = source["memory_percent"];
	        this.net_io = source["net_io"];
	        this.block_io = source["block_io"];
	        this.pid_count = source["pid_count"];
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
	export class DockerVolumeInfo {
	    driver: string;
	    name: string;
	    mountpoint: string;
	    size: string;
	
	    static createFrom(source: any = {}) {
	        return new DockerVolumeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.driver = source["driver"];
	        this.name = source["name"];
	        this.mountpoint = source["mountpoint"];
	        this.size = source["size"];
	    }
	}
	export class EnvReport {
	    hostname: string;
	    cpu: string;
	    cores: number;
	    memory: string;
	    os: string;
	    arch: string;
	    interfaces: string[];
	    package_mgrs: string[];
	    shells: string[];
	
	    static createFrom(source: any = {}) {
	        return new EnvReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.cpu = source["cpu"];
	        this.cores = source["cores"];
	        this.memory = source["memory"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.interfaces = source["interfaces"];
	        this.package_mgrs = source["package_mgrs"];
	        this.shells = source["shells"];
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
	export class ExtendedDiagnosticResult {
	    id: string;
	    checks: DiagnosticCheckData[];
	    score: number;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new ExtendedDiagnosticResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.checks = this.convertValues(source["checks"], DiagnosticCheckData);
	        this.score = source["score"];
	        this.timestamp = source["timestamp"];
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
	export class FailedLogin {
	    time: string;
	    username: string;
	    source_ip: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new FailedLogin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.username = source["username"];
	        this.source_ip = source["source_ip"];
	        this.count = source["count"];
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
	export class ForensicDiff {
	    new_processes: string[];
	    gone_processes: string[];
	    new_connections: string[];
	    snapshot_a: string;
	    snapshot_b: string;
	
	    static createFrom(source: any = {}) {
	        return new ForensicDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.new_processes = source["new_processes"];
	        this.gone_processes = source["gone_processes"];
	        this.new_connections = source["new_connections"];
	        this.snapshot_a = source["snapshot_a"];
	        this.snapshot_b = source["snapshot_b"];
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
	
	export class HardeningCheck {
	    category: string;
	    check: string;
	    passed: boolean;
	    severity: string;
	    remediation: string;
	
	    static createFrom(source: any = {}) {
	        return new HardeningCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.check = source["check"];
	        this.passed = source["passed"];
	        this.severity = source["severity"];
	        this.remediation = source["remediation"];
	    }
	}
	export class SensorData {
	    name: string;
	    type: string;
	    value: number;
	    unit: string;
	    category: string;
	
	    static createFrom(source: any = {}) {
	        return new SensorData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.category = source["category"];
	    }
	}
	export class HardwareInfo {
	    cpu: CPUExtendedInfo;
	    gpu: GPUInfo;
	    battery: BatteryInfo;
	    sensors: SensorData[];
	    baseboard: BaseboardInfo;
	
	    static createFrom(source: any = {}) {
	        return new HardwareInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu = this.convertValues(source["cpu"], CPUExtendedInfo);
	        this.gpu = this.convertValues(source["gpu"], GPUInfo);
	        this.battery = this.convertValues(source["battery"], BatteryInfo);
	        this.sensors = this.convertValues(source["sensors"], SensorData);
	        this.baseboard = this.convertValues(source["baseboard"], BaseboardInfo);
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
	export class K8sActionResult {
	    action: string;
	    message: string;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new K8sActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.message = source["message"];
	        this.success = source["success"];
	    }
	}
	export class K8sEvent {
	    last_seen: string;
	    type: string;
	    reason: string;
	    object: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new K8sEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.last_seen = source["last_seen"];
	        this.type = source["type"];
	        this.reason = source["reason"];
	        this.object = source["object"];
	        this.message = source["message"];
	    }
	}
	export class K8sNamespaceInfo {
	    name: string;
	    status: string;
	    age: string;
	
	    static createFrom(source: any = {}) {
	        return new K8sNamespaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.age = source["age"];
	    }
	}
	export class K8sResourceItem {
	    name: string;
	    namespace: string;
	    status: string;
	    age: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new K8sResourceItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.namespace = source["namespace"];
	        this.status = source["status"];
	        this.age = source["age"];
	        this.details = source["details"];
	    }
	}
	export class K8sRolloutStatus {
	    name: string;
	    kind: string;
	    ready: boolean;
	    replicas: string;
	    updated: string;
	    available: string;
	
	    static createFrom(source: any = {}) {
	        return new K8sRolloutStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.ready = source["ready"];
	        this.replicas = source["replicas"];
	        this.updated = source["updated"];
	        this.available = source["available"];
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
	export class LHMAuthorization {
	    reason: string;
	    capabilities: string[];
	    risks: string[];
	    binaryName: string;
	    publisher: string;
	
	    static createFrom(source: any = {}) {
	        return new LHMAuthorization(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reason = source["reason"];
	        this.capabilities = source["capabilities"];
	        this.risks = source["risks"];
	        this.binaryName = source["binaryName"];
	        this.publisher = source["publisher"];
	    }
	}
	export class LHMStatusResult {
	    available: boolean;
	    running: boolean;
	    needsAdmin: boolean;
	    version: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new LHMStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.running = source["running"];
	        this.needsAdmin = source["needsAdmin"];
	        this.version = source["version"];
	        this.error = source["error"];
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
	    service_name: string;
	    risk_level: string;
	
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
	        this.service_name = source["service_name"];
	        this.risk_level = source["risk_level"];
	    }
	}
	export class LoadAverageData {
	    load_1: number;
	    load_5: number;
	    load_15: number;
	
	    static createFrom(source: any = {}) {
	        return new LoadAverageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.load_1 = source["load_1"];
	        this.load_5 = source["load_5"];
	        this.load_15 = source["load_15"];
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
	export class LockedAccount {
	    username: string;
	    locked_since: string;
	
	    static createFrom(source: any = {}) {
	        return new LockedAccount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.locked_since = source["locked_since"];
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
	    trend: string;
	    summaryText: string;
	
	    static createFrom(source: any = {}) {
	        return new LogSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.topSource = source["topSource"];
	        this.topMessage = source["topMessage"];
	        this.trend = source["trend"];
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
	export class LoggedInUserData {
	    user: string;
	    terminal: string;
	    host: string;
	    started: string;
	
	    static createFrom(source: any = {}) {
	        return new LoggedInUserData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user = source["user"];
	        this.terminal = source["terminal"];
	        this.host = source["host"];
	        this.started = source["started"];
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
	export class NetOpsFirewallRuleData {
	    name: string;
	    direction: string;
	    action: string;
	    protocol: string;
	    ports: string;
	    enabled: boolean;
	    source: string;
	    destination: string;
	
	    static createFrom(source: any = {}) {
	        return new NetOpsFirewallRuleData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.direction = source["direction"];
	        this.action = source["action"];
	        this.protocol = source["protocol"];
	        this.ports = source["ports"];
	        this.enabled = source["enabled"];
	        this.source = source["source"];
	        this.destination = source["destination"];
	    }
	}
	export class NetworkActionResult {
	    action: string;
	    message: string;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NetworkActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.message = source["message"];
	        this.success = source["success"];
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
	export class NetworkHealthCheck {
	    name: string;
	    status: string;
	    detail: string;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new NetworkHealthCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.detail = source["detail"];
	        this.score = source["score"];
	    }
	}
	export class NetworkHealthReport {
	    score: number;
	    checks: NetworkHealthCheck[];
	    summary: string;
	    duration: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkHealthReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.checks = this.convertValues(source["checks"], NetworkHealthCheck);
	        this.summary = source["summary"];
	        this.duration = source["duration"];
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
	export class TopologyConnectionData {
	    id: string;
	    source_id: string;
	    target_id: string;
	    type: string;
	    label?: string;
	    metric?: number;
	
	    static createFrom(source: any = {}) {
	        return new TopologyConnectionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source_id = source["source_id"];
	        this.target_id = source["target_id"];
	        this.type = source["type"];
	        this.label = source["label"];
	        this.metric = source["metric"];
	    }
	}
	export class TopologyDeviceData {
	    id: string;
	    type: string;
	    label: string;
	    ip?: string;
	    mac?: string;
	    subnet?: string;
	    vendor?: string;
	    hostname?: string;
	    status: string;
	    x: number;
	    y: number;
	    online: boolean;
	    notes?: string;
	
	    static createFrom(source: any = {}) {
	        return new TopologyDeviceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.label = source["label"];
	        this.ip = source["ip"];
	        this.mac = source["mac"];
	        this.subnet = source["subnet"];
	        this.vendor = source["vendor"];
	        this.hostname = source["hostname"];
	        this.status = source["status"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.online = source["online"];
	        this.notes = source["notes"];
	    }
	}
	export class NetworkTopologyData {
	    devices: TopologyDeviceData[];
	    connections: TopologyConnectionData[];
	    generated_at: string;
	    subnet: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkTopologyData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.devices = this.convertValues(source["devices"], TopologyDeviceData);
	        this.connections = this.convertValues(source["connections"], TopologyConnectionData);
	        this.generated_at = source["generated_at"];
	        this.subnet = source["subnet"];
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
	    binary_exists: boolean;
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
	        this.binary_exists = source["binary_exists"];
	        this.model = source["model"];
	        this.version = source["version"];
	        this.available_models = source["available_models"];
	        this.error = source["error"];
	    }
	}
	export class PackageInfo {
	    name: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new PackageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	    }
	}
	export class PackageManagerData {
	    name: string;
	    found: boolean;
	    packages: PackageInfo[];
	
	    static createFrom(source: any = {}) {
	        return new PackageManagerData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.found = source["found"];
	        this.packages = this.convertValues(source["packages"], PackageInfo);
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
	export class PasswordPolicy {
	    max_age: number;
	    min_length: number;
	    complexity: boolean;
	    lockout_threshold: number;
	    lockout_duration: number;
	
	    static createFrom(source: any = {}) {
	        return new PasswordPolicy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.max_age = source["max_age"];
	        this.min_length = source["min_length"];
	        this.complexity = source["complexity"];
	        this.lockout_threshold = source["lockout_threshold"];
	        this.lockout_duration = source["lockout_duration"];
	    }
	}
	
	export class PerformanceData {
	    cpu_times: CPUTimesData;
	    load_average: LoadAverageData;
	    io_wait: number;
	
	    static createFrom(source: any = {}) {
	        return new PerformanceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu_times = this.convertValues(source["cpu_times"], CPUTimesData);
	        this.load_average = this.convertValues(source["load_average"], LoadAverageData);
	        this.io_wait = source["io_wait"];
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
	export class PerformanceProfile {
	    category: string;
	    parallelism: number;
	    scan_intensity: string;
	    cpu_threads: number;
	    memory_gb: number;
	
	    static createFrom(source: any = {}) {
	        return new PerformanceProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.parallelism = source["parallelism"];
	        this.scan_intensity = source["scan_intensity"];
	        this.cpu_threads = source["cpu_threads"];
	        this.memory_gb = source["memory_gb"];
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
	export class PingResultMultiData {
	    target: string;
	    min_ms: number;
	    avg_ms: number;
	    max_ms: number;
	    stddev_ms: number;
	    packet_loss: number;
	    jitter_ms: number;
	    individual_rtts: number[];
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new PingResultMultiData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.min_ms = source["min_ms"];
	        this.avg_ms = source["avg_ms"];
	        this.max_ms = source["max_ms"];
	        this.stddev_ms = source["stddev_ms"];
	        this.packet_loss = source["packet_loss"];
	        this.jitter_ms = source["jitter_ms"];
	        this.individual_rtts = source["individual_rtts"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class PingStatsData {
	    avg_latency: number;
	    max_latency: number;
	    total_loss: number;
	    worst_target: string;
	
	    static createFrom(source: any = {}) {
	        return new PingStatsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.avg_latency = source["avg_latency"];
	        this.max_latency = source["max_latency"];
	        this.total_loss = source["total_loss"];
	        this.worst_target = source["worst_target"];
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
	export class PrebuiltReportTemplate {
	    id: string;
	    category: string;
	    preset_name: string;
	    description: string;
	    metric: string;
	    condition: string;
	    threshold: number;
	    report_type: string;
	    schedule: string;
	
	    static createFrom(source: any = {}) {
	        return new PrebuiltReportTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.preset_name = source["preset_name"];
	        this.description = source["description"];
	        this.metric = source["metric"];
	        this.condition = source["condition"];
	        this.threshold = source["threshold"];
	        this.report_type = source["report_type"];
	        this.schedule = source["schedule"];
	    }
	}
	export class PrivilegeEvent {
	    time: string;
	    username: string;
	    privilege: string;
	    process: string;
	
	    static createFrom(source: any = {}) {
	        return new PrivilegeEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.username = source["username"];
	        this.privilege = source["privilege"];
	        this.process = source["process"];
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
	    is_signed: boolean;
	    publisher: string;
	
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
	        this.is_signed = source["is_signed"];
	        this.publisher = source["publisher"];
	    }
	}
	export class ProcessNode {
	    pid: number;
	    ppid: number;
	    name: string;
	    cpu: number;
	    memory: number;
	    mem_pct: number;
	    status: string;
	    num_fds: number;
	    is_signed: boolean;
	    publisher: string;
	    children?: ProcessNode[];
	
	    static createFrom(source: any = {}) {
	        return new ProcessNode(source);
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
	        this.is_signed = source["is_signed"];
	        this.publisher = source["publisher"];
	        this.children = this.convertValues(source["children"], ProcessNode);
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
	export class PublicExposure {
	    port: number;
	    protocol: string;
	    process_name: string;
	    severity: string;
	
	    static createFrom(source: any = {}) {
	        return new PublicExposure(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.process_name = source["process_name"];
	        this.severity = source["severity"];
	    }
	}
	export class ReportGenerationResult {
	    report_id: string;
	    type: string;
	    timestamp: string;
	    score: number;
	    summary?: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportGenerationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.report_id = source["report_id"];
	        this.type = source["type"];
	        this.timestamp = source["timestamp"];
	        this.score = source["score"];
	        this.summary = source["summary"];
	    }
	}
	export class ReportTypeMeta {
	    type: string;
	    label: string;
	    description: string;
	    available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ReportTypeMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.available = source["available"];
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
	export class SSHConfig {
	    permit_root_login: string;
	    password_authentication: string;
	    pubkey_authentication: string;
	    x11_forwarding: string;
	    max_auth_tries: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.permit_root_login = source["permit_root_login"];
	        this.password_authentication = source["password_authentication"];
	        this.pubkey_authentication = source["pubkey_authentication"];
	        this.x11_forwarding = source["x11_forwarding"];
	        this.max_auth_tries = source["max_auth_tries"];
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
	export class ScheduledTaskData {
	    name: string;
	    schedule: string;
	    command: string;
	    enabled: boolean;
	    next_run: string;
	
	    static createFrom(source: any = {}) {
	        return new ScheduledTaskData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.schedule = source["schedule"];
	        this.command = source["command"];
	        this.enabled = source["enabled"];
	        this.next_run = source["next_run"];
	    }
	}
	export class SecTimelineEvent {
	    time: string;
	    type: string;
	    detail: string;
	    severity: string;
	
	    static createFrom(source: any = {}) {
	        return new SecTimelineEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.type = source["type"];
	        this.detail = source["detail"];
	        this.severity = source["severity"];
	    }
	}
	export class SecureBoot {
	    enabled: boolean;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new SecureBoot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.state = source["state"];
	    }
	}
	export class SecurityAuditResult {
	    id: string;
	    score: number;
	    total: number;
	    passed: number;
	    failed: number;
	    items: AuditCheckItem[];
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new SecurityAuditResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.score = source["score"];
	        this.total = source["total"];
	        this.passed = source["passed"];
	        this.failed = source["failed"];
	        this.items = this.convertValues(source["items"], AuditCheckItem);
	        this.timestamp = source["timestamp"];
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
	export class SystemLogEntry {
	    timestamp: string;
	    level: string;
	    source: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.source = source["source"];
	        this.message = source["message"];
	    }
	}
	export class SystemLogsResultData {
	    entries: SystemLogEntry[];
	    source: string;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemLogsResultData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entries = this.convertValues(source["entries"], SystemLogEntry);
	        this.source = source["source"];
	        this.total = source["total"];
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
	export class SystemService {
	    name: string;
	    display_name: string;
	    status: string;
	    startup_type: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display_name = source["display_name"];
	        this.status = source["status"];
	        this.startup_type = source["startup_type"];
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
	export class SystemSnapshot {
	    metrics: DashboardData;
	    timeline: TimelineEvent[];
	    alerts: AlertInfo[];
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metrics = this.convertValues(source["metrics"], DashboardData);
	        this.timeline = this.convertValues(source["timeline"], TimelineEvent);
	        this.alerts = this.convertValues(source["alerts"], AlertInfo);
	        this.timestamp = source["timestamp"];
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
	export class TLSCertificate {
	    subject: string;
	    issuer: string;
	    not_after: string;
	    key_size: number;
	    is_expiring: boolean;
	    days_left: number;
	
	    static createFrom(source: any = {}) {
	        return new TLSCertificate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subject = source["subject"];
	        this.issuer = source["issuer"];
	        this.not_after = source["not_after"];
	        this.key_size = source["key_size"];
	        this.is_expiring = source["is_expiring"];
	        this.days_left = source["days_left"];
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
	export class VPNStatusData {
	    active: boolean;
	    type: string;
	    interface: string;
	    remote_ip: string;
	    local_ip: string;
	    protocol: string;
	
	    static createFrom(source: any = {}) {
	        return new VPNStatusData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.active = source["active"];
	        this.type = source["type"];
	        this.interface = source["interface"];
	        this.remote_ip = source["remote_ip"];
	        this.local_ip = source["local_ip"];
	        this.protocol = source["protocol"];
	    }
	}

}

export namespace common {
	
	export class WorkflowStep {
	    id: string;
	    type: string;
	    label: string;
	    description: string;
	    command: string;
	    expected_outcome: string;
	    result?: any;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.command = source["command"];
	        this.expected_outcome = source["expected_outcome"];
	        this.result = source["result"];
	        this.error = source["error"];
	    }
	}
	export class ActionPreview {
	    handshake_id: string;
	    action: string;
	    command?: string;
	    description: string;
	    risks: string[];
	    rollback: string;
	    typical_values: string;
	    workflow_id?: string;
	    steps?: WorkflowStep[];
	
	    static createFrom(source: any = {}) {
	        return new ActionPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.handshake_id = source["handshake_id"];
	        this.action = source["action"];
	        this.command = source["command"];
	        this.description = source["description"];
	        this.risks = source["risks"];
	        this.rollback = source["rollback"];
	        this.typical_values = source["typical_values"];
	        this.workflow_id = source["workflow_id"];
	        this.steps = this.convertValues(source["steps"], WorkflowStep);
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
	export class AutoReportRule {
	    id: string;
	    name: string;
	    description: string;
	    metric: string;
	    condition: string;
	    threshold: number;
	    report_type: string;
	    schedule: string;
	    enabled: boolean;
	    created_at: string;
	    last_triggered_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new AutoReportRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.metric = source["metric"];
	        this.condition = source["condition"];
	        this.threshold = source["threshold"];
	        this.report_type = source["report_type"];
	        this.schedule = source["schedule"];
	        this.enabled = source["enabled"];
	        this.created_at = source["created_at"];
	        this.last_triggered_at = source["last_triggered_at"];
	    }
	}
	export class CapabilityInfo {
	    id: string;
	    available: boolean;
	    path: string;
	    version: string;
	    is_custom: boolean;
	    is_supported: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CapabilityInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.available = source["available"];
	        this.path = source["path"];
	        this.version = source["version"];
	        this.is_custom = source["is_custom"];
	        this.is_supported = source["is_supported"];
	    }
	}
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
	export class ForensicRecord {
	    id: string;
	    timestamp: string;
	    type: string;
	    title: string;
	    data_json: string;
	    metadata: string;
	
	    static createFrom(source: any = {}) {
	        return new ForensicRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.data_json = source["data_json"];
	        this.metadata = source["metadata"];
	    }
	}
	export class ReportRecord {
	    id: string;
	    timestamp: string;
	    type: string;
	    score: number;
	    data_json: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.type = source["type"];
	        this.score = source["score"];
	        this.data_json = source["data_json"];
	    }
	}
	export class SLIResult {
	    sloId: string;
	    sloName: string;
	    compliantPct: number;
	    targetPct: number;
	    met: boolean;
	    samples: number;
	    evaluatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SLIResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sloId = source["sloId"];
	        this.sloName = source["sloName"];
	        this.compliantPct = source["compliantPct"];
	        this.targetPct = source["targetPct"];
	        this.met = source["met"];
	        this.samples = source["samples"];
	        this.evaluatedAt = source["evaluatedAt"];
	    }
	}
	export class SLODefinition {
	    id: string;
	    name: string;
	    metric: string;
	    comparison: string;
	    threshold: number;
	    targetPct: number;
	    windowDays: number;
	    enabled: boolean;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new SLODefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.metric = source["metric"];
	        this.comparison = source["comparison"];
	        this.threshold = source["threshold"];
	        this.targetPct = source["targetPct"];
	        this.windowDays = source["windowDays"];
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	    }
	}
	export class SLOEngine {
	
	
	    static createFrom(source: any = {}) {
	        return new SLOEngine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SLOSummary {
	    totalSLOs: number;
	    metCount: number;
	    missCount: number;
	    overallPct: number;
	    results: SLIResult[];
	
	    static createFrom(source: any = {}) {
	        return new SLOSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalSLOs = source["totalSLOs"];
	        this.metCount = source["metCount"];
	        this.missCount = source["missCount"];
	        this.overallPct = source["overallPct"];
	        this.results = this.convertValues(source["results"], SLIResult);
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
	export class SecActionResult {
	    success: boolean;
	    message: string;
	    detail?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SecActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.detail = source["detail"];
	        this.error = source["error"];
	    }
	}
	export class SystemKnowledge {
	    "system.cpu.utilization": number;
	    cpu_trend: string;
	    "system.memory.usage": number;
	    memory_trend: string;
	    "system.disk.usage": number;
	    disk_trend: string;
	    active_conns: number;
	    anomalies: number;
	    "system.uptime": string;
	    security_grade: string;
	    "system.network.rx": number;
	    net_rx_trend: string;
	    "system.network.tx": number;
	    net_tx_trend: string;
	    "system.load.1m": number;
	    "system.load.5m": number;
	    "system.load.15m": number;
	    "system.swap.usage": number;
	    swap_trend: string;
	    "system.disk.io.read": number;
	    disk_io_read_trend: string;
	    "system.disk.io.write": number;
	    disk_io_write_trend: string;
	    "process.count": number;
	    "connection.count": number;
	
	    static createFrom(source: any = {}) {
	        return new SystemKnowledge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this["system.cpu.utilization"] = source["system.cpu.utilization"];
	        this.cpu_trend = source["cpu_trend"];
	        this["system.memory.usage"] = source["system.memory.usage"];
	        this.memory_trend = source["memory_trend"];
	        this["system.disk.usage"] = source["system.disk.usage"];
	        this.disk_trend = source["disk_trend"];
	        this.active_conns = source["active_conns"];
	        this.anomalies = source["anomalies"];
	        this["system.uptime"] = source["system.uptime"];
	        this.security_grade = source["security_grade"];
	        this["system.network.rx"] = source["system.network.rx"];
	        this.net_rx_trend = source["net_rx_trend"];
	        this["system.network.tx"] = source["system.network.tx"];
	        this.net_tx_trend = source["net_tx_trend"];
	        this["system.load.1m"] = source["system.load.1m"];
	        this["system.load.5m"] = source["system.load.5m"];
	        this["system.load.15m"] = source["system.load.15m"];
	        this["system.swap.usage"] = source["system.swap.usage"];
	        this.swap_trend = source["swap_trend"];
	        this["system.disk.io.read"] = source["system.disk.io.read"];
	        this.disk_io_read_trend = source["disk_io_read_trend"];
	        this["system.disk.io.write"] = source["system.disk.io.write"];
	        this.disk_io_write_trend = source["disk_io_write_trend"];
	        this["process.count"] = source["process.count"];
	        this["connection.count"] = source["connection.count"];
	    }
	}
	export class WorkflowDefinition {
	    id: string;
	    name: string;
	    description: string;
	    category: string;
	    why: string;
	    risks: string[];
	    typical_values: string;
	    steps: WorkflowStep[];
	
	    static createFrom(source: any = {}) {
	        return new WorkflowDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.why = source["why"];
	        this.risks = source["risks"];
	        this.typical_values = source["typical_values"];
	        this.steps = this.convertValues(source["steps"], WorkflowStep);
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

export namespace mcp {
	
	export class Tool {
	    name: string;
	    description: string;
	    inputSchema: number[];
	
	    static createFrom(source: any = {}) {
	        return new Tool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.inputSchema = source["inputSchema"];
	    }
	}

}

export namespace netops {
	
	export class ARPEntry {
	    ip: string;
	    mac: string;
	    vendor: string;
	    interface: string;
	
	    static createFrom(source: any = {}) {
	        return new ARPEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.mac = source["mac"];
	        this.vendor = source["vendor"];
	        this.interface = source["interface"];
	    }
	}
	export class DiscoveredDevice {
	    ip: string;
	    mac: string;
	    vendor: string;
	    hostname: string;
	    response_time_ms: number;
	    is_gateway: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.mac = source["mac"];
	        this.vendor = source["vendor"];
	        this.hostname = source["hostname"];
	        this.response_time_ms = source["response_time_ms"];
	        this.is_gateway = source["is_gateway"];
	    }
	}
	export class DiscoveryResult {
	    devices: DiscoveredDevice[];
	    subnet: string;
	    scan_time_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.devices = this.convertValues(source["devices"], DiscoveredDevice);
	        this.subnet = source["subnet"];
	        this.scan_time_ms = source["scan_time_ms"];
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
	export class RouteEntry {
	    destination: string;
	    mask: string;
	    gateway: string;
	    interface: string;
	    metric: number;
	    is_default: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RouteEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.destination = source["destination"];
	        this.mask = source["mask"];
	        this.gateway = source["gateway"];
	        this.interface = source["interface"];
	        this.metric = source["metric"];
	        this.is_default = source["is_default"];
	    }
	}
	export class WiFiInfo {
	    interface: string;
	    ssid: string;
	    signal: number;
	    speed: string;
	    channel: number;
	
	    static createFrom(source: any = {}) {
	        return new WiFiInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interface = source["interface"];
	        this.ssid = source["ssid"];
	        this.signal = source["signal"];
	        this.speed = source["speed"];
	        this.channel = source["channel"];
	    }
	}
	export class WiFiNetwork {
	    ssid: string;
	    signal: number;
	    channel: number;
	    security: string;
	    bssid: string;
	    frequency: string;
	
	    static createFrom(source: any = {}) {
	        return new WiFiNetwork(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ssid = source["ssid"];
	        this.signal = source["signal"];
	        this.channel = source["channel"];
	        this.security = source["security"];
	        this.bssid = source["bssid"];
	        this.frequency = source["frequency"];
	    }
	}

}

