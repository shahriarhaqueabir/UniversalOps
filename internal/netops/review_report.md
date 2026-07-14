# NetOps Code Review

```json
{
  "module": "netops",
  "files_reviewed": [
    "ping.go", "ping_multi.go", "dns.go", "dns_advanced.go",
    "portscan.go", "traceroute.go", "connections.go", "interfaces.go",
    "arp.go", "routing.go", "wifi.go", "gateway.go", "firewall.go",
    "vpn.go", "discovery.go", "monitoring.go", "health.go", "oui.go",
    "actions.go", "workflows.go"
  ],
  "summary": {
    "total_issues": 27,
    "critical": 1,
    "high": 10,
    "medium": 9,
    "low": 7
  },
  "findings": [
    {
      "file": "discovery.go",
      "line": 72,
      "severity": "CRITICAL",
      "category": "correctness",
      "title": "generateSubnetHosts produces garbage IPs for CIDR subnet input",
      "detail": "The function checks only the last char for '/' (e.g. '192.168.1.0/24' last char is '4'). CIDR never stripped, base stays '192.168.1.0/24' and generated hosts become '192.168.1.0/24.1' through '192.168.1.0/24.254'. Even the working path emits a trailing zero octet. Function does not actually parse CIDR at all.",
      "recommendation": "Use net.ParseCIDR(subnet) to extract network IP and mask, then iterate over valid host addresses with proper IP increment logic."
    },
    {
      "file": "connections.go",
      "line": 410,
      "severity": "HIGH",
      "category": "correctness",
      "title": "Manual CSV parsing of PowerShell output breaks on process names with commas or quotes",
      "detail": "getPidMapViaPowerShell splits on ',\"' to parse CSV. Process names containing commas (e.g. 'Microsoft Office, 64-bit') or quotes produce wrong fields. Fallback getPidMapViaWmic correctly uses csv.NewReader.",
      "recommendation": "Use encoding/csv reader for PowerShell output, matching the wmic code path."
    },
    {
      "file": "routing.go",
      "line": 107,
      "severity": "HIGH",
      "category": "security",
      "title": "ManageStaticRoutes passes user-supplied parameters directly to OS commands",
      "detail": "dest, mask, and gateway go verbatim into exec.Command args for netsh and sudo ip route. No validation: empty, malformed input reaches OS commands. On Linux runs via sudo.",
      "recommendation": "Validate all params via net.ParseIP or net.ParseCIDR before passing to exec. Return typed errors for invalid input."
    },
    {
      "file": "arp.go",
      "line": 31,
      "severity": "HIGH",
      "category": "security",
      "title": "ARP table functions use raw exec.Command instead of sandboxed wrapper",
      "detail": "getARPTableWindows/Linux use os/exec.Command directly. Circumvents common.SandboxedCommandWithConfig sandbox used everywhere else.",
      "recommendation": "Replace with common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), ...) with DenyNetworkAccess=false."
    },
    {
      "file": "routing.go",
      "line": 33,
      "severity": "HIGH",
      "category": "security",
      "title": "Routing table functions use raw exec.Command instead of sandboxed wrapper",
      "detail": "getRoutingTableWindows/Linux use exec.Command directly for netstat -rn and ip route show. Same sandbox bypass as ARP.",
      "recommendation": "Replace with sandboxed command wrappers matching ping.go and traceroute.go pattern."
    },
    {
      "file": "wifi.go",
      "line": 34,
      "severity": "HIGH",
      "category": "security",
      "title": "All WiFi commands use raw exec.Command instead of sandboxed wrapper",
      "detail": "ScanWiFiNetworks and GetWiFiInfo call exec.Command for netsh wlan and nmcli. Not sandboxed.",
      "recommendation": "Replace with common.SandboxedCommandWithConfig with DenyNetworkAccess=false."
    },
    {
      "file": "interfaces.go",
      "line": 91,
      "severity": "HIGH",
      "category": "go-idioms",
      "title": "Interface name modification in data layer breaks downstream lookups",
      "detail": "GetInterfaces mutates Info.Name by prepending '[Loopback]', '[WiFi]', or '[Wired]'. Callers get wrong system name, breaking any name-keyed lookup. UI concern leaking into data layer.",
      "recommendation": "Remove name mutation. Add a separate DisplayLabel field or let the frontend apply labels from flags."
    },
    {
      "file": "dns_advanced.go",
      "line": 84,
      "severity": "HIGH",
      "category": "error-handling",
      "title": "Silently discarded json.Marshal error in DoH query",
      "detail": "body, _ = json.Marshal(...) silently discards error. If marshal fails, http.Post receives nil body with confusing failure.",
      "recommendation": "Check json.Marshal error and return it."
    },
    {
      "file": "connections.go",
      "line": 100,
      "severity": "HIGH",
      "category": "performance",
      "title": "buildProcInodeMap walks entire /proc/*/fd three times per GetConnections call",
      "detail": "getConnectionsLinux calls parseProcNet 3x (tcp, tcp6, udp). Each call invokes buildProcInodeMap scanning ALL /proc/<pid>/fd dirs. 500 procs x 100 fds x 3 = 150K readlink calls per invocation.",
      "recommendation": "Move buildProcInodeMap out into getConnectionsLinux, pass as parameter. Cache with TTL."
    },
    {
      "file": "dns_advanced.go",
      "line": 29,
      "severity": "HIGH",
      "category": "portability",
      "title": "FlushDNSCache uses sudo and will hang a GUI app with no TTY",
      "detail": "On Linux, runs sudo resolvectl/systemd-resolve. Wails GUI app has no TTY for sudo password prompt, causing indefinite hang. No timeout or context support.",
      "recommendation": "Use pkexec or D-Bus instead of sudo. Add context parameter with timeout. Document elevated privilege requirement."
    },
    {
      "file": "ping_multi.go",
      "line": 62,
      "severity": "MEDIUM",
      "category": "correctness",
      "title": "Individual RTTs reconstructed from average produces StdDev of 0",
      "detail": "pingOneTarget fills IndividualRTTs with avg value repeated (rtts[i]=result.AvgMs). StdDev always 0. Callers get misleading stddev=0 and fake individual samples.",
      "recommendation": "Make Ping return individual RTTs, or remove IndividualRTTs and StdDevMs from PingResultMulti."
    },
    {
      "file": "dns.go",
      "line": 104,
      "severity": "MEDIUM",
      "category": "correctness",
      "title": "Early return only checks A/AAAA/MX records, ignores NS/CNAME/TXT-only domains",
      "detail": "Success check only looks at A/AAAA/MX. Domains serving only NS, CNAME, or TXT never trigger early return, causing iteration through all DNS servers.",
      "recommendation": "Include all queried record types in the success condition."
    },
    {
      "file": "gateway.go",
      "line": 22,
      "severity": "MEDIUM",
      "category": "error-handling",
      "title": "GetDefaultGateway returns zero-value on failure, indistinguishable from no-gateway",
      "detail": "No error return. Command failure, missing /proc file, and no-default-route all return empty GatewayInfo{} with no way to distinguish.",
      "recommendation": "Change signature to (GatewayInfo, error) and wrap errors on failure."
    },
    {
      "file": "dns_advanced.go",
      "line": 57,
      "severity": "MEDIUM",
      "category": "network-specific",
      "title": "ReverseLookup hardcodes 8.8.8.8:53 as only DNS server, rejects IPv6",
      "detail": "Always queries Google DNS. Fails on networks blocking 8.8.8.8 or with internal reverse zones. Also rejects all IPv6 addresses.",
      "recommendation": "Accept optional DNS server param. Try net.LookupAddr as fallback. Support IPv6 via ip6.arpa."
    },
    {
      "file": "discovery.go",
      "line": 52,
      "severity": "MEDIUM",
      "category": "network-specific",
      "title": "Ping sweep sends 1 packet per host, producing flaky results",
      "detail": "Ping(ip, 1) = single ICMP per host. On lossy networks this gives many false negatives. 254 goroutines on Windows each spawn a process via pingExec.",
      "recommendation": "Use 2-3 pings per host. Add throttling/batching to avoid process-spawn storm on Windows."
    },
    {
      "file": "monitoring.go",
      "line": 61,
      "severity": "MEDIUM",
      "category": "correctness",
      "title": "Bandwidth rate uses configured interval not actual elapsed time",
      "detail": "Rate divides by intervalSec rather than actual wall-clock delta. Late ticker firings (GC, load) produce incorrect rates.",
      "recommendation": "Record actual timestamps and compute delta-seconds between samples."
    },
    {
      "file": "health.go",
      "line": 33,
      "severity": "MEDIUM",
      "category": "performance",
      "title": "Duplicate ping of 8.8.8.8 in same health check run",
      "detail": "checkPingHealth('Internet Reachability', 8.8.8.8, 3) and checkPingHealth('Internet Latency', 8.8.8.8, 5) run sequentially. 8 pings sent when 5 suffice.",
      "recommendation": "Consolidate into one ping with 5 packets, compute reachability and latency from same result."
    },
    {
      "file": "vpn.go",
      "line": 49,
      "severity": "MEDIUM",
      "category": "portability",
      "title": "Windows VPN detection depends on English locale string",
      "detail": "strings.Contains(out, 'Connected') fails on non-English Windows where rasdial output is localized. Silently misses VPNs.",
      "recommendation": "Check rasdial exit code or use PowerShell Get-VpnConnection for locale-independent detection."
    },
    {
      "file": "ping.go",
      "line": 243,
      "severity": "LOW",
      "category": "go-idioms",
      "title": "resolveTarget returns hostname as IP with nil error when no addresses found",
      "detail": "If net.LookupHost returns empty slice with nil error, falls through to return target,nil — caller gets a hostname string as if it were a resolved IP.",
      "recommendation": "Return error when len(ips)==0."
    },
    {
      "file": "ping.go",
      "line": 102,
      "severity": "LOW",
      "category": "performance",
      "title": "Reply buffer allocated on every ping iteration",
      "detail": "make([]byte, 1500) inside the loop body. Unnecessary allocation churn.",
      "recommendation": "Move allocation outside loop, reuse buffer."
    },
    {
      "file": "connections.go",
      "line": 415,
      "severity": "LOW",
      "category": "go-idioms",
      "title": "Inconsistent CSV parsing: primary path manual, fallback uses csv.NewReader",
      "detail": "getPidMapViaPowerShell (primary) splits on ',', getPidMapViaWmic (fallback) uses csv.NewReader. The robust approach should be the primary one.",
      "recommendation": "Replace manual splitting with csv.NewReader in PowerShell path."
    },
    {
      "file": "dns.go",
      "line": 38,
      "severity": "LOW",
      "category": "performance",
      "title": "New dns.Client created on every LookupDNSWithContext call",
      "detail": "dns.Client is stateless (except timeout). Creating a new one per call is unnecessary.",
      "recommendation": "Use a shared package-level dns.Client with timeout set at init."
    },
    {
      "file": "oui.go",
      "line": 5,
      "severity": "LOW",
      "category": "completeness",
      "title": "OUI database contains only 45 entries",
      "detail": "Hardcoded map covers only major vendors. Returns empty string for vast majority of real-world MAC addresses with no indication lookup was attempted.",
      "recommendation": "Either document as stub, or load full IEEE OUI database (~30K entries) from bundled file."
    },
    {
      "file": "vpn.go",
      "line": 37,
      "severity": "LOW",
      "category": "correctness",
      "title": "VPN status returns on first detected VPN, ignoring multiple active VPNs",
      "detail": "Returns immediately after first matching VPN interface. Split-tunnel or multi-VPN setups only get the first one reported.",
      "recommendation": "Collect all VPN interfaces into a slice, or document single-VPN limitation."
    },
    {
      "file": "actions.go",
      "line": 60,
      "severity": "LOW",
      "category": "error-handling",
      "title": "Error from sleep/timeout command silently ignored in resetInterface",
      "detail": "cmd.Run() return value discarded for the sleep/timeout wait. Function proceeds regardless, which is acceptable but should be explicit.",
      "recommendation": "Log the error if sleep fails: if err := cmd.Run(); err != nil { common.LogInfo(...) }"
    },
    {
      "file": "portscan.go",
      "line": 58,
      "severity": "LOW",
      "category": "network-specific",
      "title": "ScanPorts has no context support for cancellation",
      "detail": "No context.Context parameter. If user cancels, goroutines continue until all TCP dials complete.",
      "recommendation": "Accept optional context.Context and use net.Dialer.DialContext."
    }
  ],
  "strengths": [
    "Consistent goroutine+WaitGroup pattern for concurrent operations across ping_multi, portscan, discovery, workflows — no goroutine leaks visible",
    "Good use of common.RecoverPanic in goroutines preventing single panic from crashing the app",
    "Well-structured fallback chains: Ping ICMP->exec, DNS explicit->system resolver, Gateway /proc/net->ip route, process name PS->wmic",
    "Context-aware variants exist for key functions (LookupDNSWithContext, TraceRouteWithContext)",
    "Proper ICMP message handling with type switching for EchoReply",
    "Monitoring goroutine has clean shutdown path via done channel pattern",
    "Comprehensive cross-platform coverage — nearly every function has Windows and Linux implementations",
    "Bandwidth history bounded at 360 samples with mutex-protected concurrent access"
  ],
  "overall_assessment": "The netops module is a well-structured, feature-rich network diagnostics library with strong cross-platform awareness and sensible fallback chains. The critical finding is generateSubnetHosts in discovery.go producing garbage IPs from CIDR input — needs a complete rewrite using net.ParseCIDR. The highest-impact systemic issue is inconsistent sandbox use: ~40% of OS command invocations (arp, routing, wifi, actions) bypass the project's common.SandboxedCommandWithConfig sandbox via raw exec.Command, creating both security and consistency risks. The most impactful user-facing data bug is the always-zero StdDev in multi-ping results from reconstructing individual RTTs from aggregate averages. The most actionable pattern fix is replacing manual CSV splitting in getPidMapViaPowerShell with encoding/csv, matching the fallback path that already does it correctly."
}
```
