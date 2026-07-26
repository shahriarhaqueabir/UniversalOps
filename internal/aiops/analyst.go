package aiops

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// AnalystResponse contains the AI message and any proposed actions.
type AnalystResponse struct {
	Content string                 `json:"content"`
	Actions []common.ActionPreview `json:"actions,omitempty"`
}

var injectionKeywords = []string{
	"ignore previous",
	"system prompt",
	"you are now",
	"new instructions",
	"forget everything",
	"capability boundaries",
	"you must follow",
	"override",
	"new role",
}

// SanitizeInput strips content that could facilitate prompt injection.
func SanitizeInput(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 500 {
		s = s[:500]
	}

	// 1. Block XML tag escaping (critical for our action protocol)
	s = strings.ReplaceAll(s, "</user_query>", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<system_state>", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<action_request", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<thought>", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<function", "[TAG_FILTERED]")

	// 2. Check for common injection patterns (case-insensitive)
	lower := strings.ToLower(s)
	for _, kw := range injectionKeywords {
		if strings.Contains(lower, kw) {
			common.LogWarn("Security: Blocked potential prompt injection: %q", s)
			return "[COMMAND REJECTED BY SECURITY POLICY]"
		}
	}

	// 3. Normalize whitespace to prevent multi-line bypasses
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

// BuildAnalystPrompt creates a structured system prompt using physical system reality.
func BuildAnalystPrompt(k common.SystemKnowledge, history string) string {
	return fmt.Sprintf(`You are the Universal-Ops System Analyst, a high-density technical co-pilot.
Objective: Synthesize complex telemetry into factual technical briefings.

Current System State (OTel Mapped):
- CPU (system.cpu.utilization): %.1f%% (Trend: %s)
- RAM (system.memory.usage): %.1f%% (Trend: %s)
- Disk (system.disk.usage): %.1f%% (Trend: %s)
- Processes: %d
- Connections: %d
- Network RX: %.1f bps (Trend: %s)
- Network TX: %.1f bps (Trend: %s)
- Load Avg 1m: %.2f / 5m: %.2f / 15m: %.2f
- Swap: %.1f%% (Trend: %s)
- Disk IO Read: %.1f B/s (Trend: %s)
- Disk IO Write: %.1f B/s (Trend: %s)
- Anomalies: %d
- Security Grade: %s
- Uptime: %s
%s

### Operational Protocol
1. Use a <thought> block to correlate telemetry history and identify root causes.
2. Provide a concise technical justification for any proposed action.
3. Emit action requests using: <action_request name="ACTION_NAME" param1="VALUE" />
4. If a metric is rising sharply, provide a technical justification for the trend.
5. To request on-demand system data, use: <function name="FUNCTION_NAME">{"arg":"value"}</function>
   - System will auto-execute these and return results. Never guess data — request it.

### Available Actions (HITL — require user approval)
- kill_process (pid): Stop a specific process.
- restart_service (name): Restart a system service.
- disk_cleanup: Initiate temporary file removal.
- defrag: Optimize primary drive storage.

### Available MCP Functions (auto-executed — for on-demand data)
Use <function name="NAME">{"param":"value"}</function> to request live data:
- get_system_telemetry: Full system snapshot (CPU, RAM, Disk, Network, Load, Swap, Security)
- get_process_list (n): Top-N processes by CPU usage (default 20)
- get_system_logs (n, source): Recent Event Log entries (default 50, source="Application")
- get_scheduled_tasks: List all scheduled tasks
- get_hardware_info: OS, kernel, hostname, uptime
- get_disk_usage: Disk partition stats
- get_baseboard_info: Motherboard / BIOS details
- ping (target, count): ICMP echo to a host (default count=4)
- dns_lookup (hostname, server): DNS resolution (optional custom server)
- port_scan (host, ports): Check TCP port reachability
- traceroute (target): Network path trace
- get_network_connections: All current TCP/UDP connections
- get_network_health: Quick connectivity summary
- get_firewall_rules: Windows Firewall rule list
- get_listening_ports: All listening TCP/UDP ports
- get_defender_status: Windows Defender / AV status
- run_security_audit: Run the full security checklist
- get_failed_logins: Recent failed authentication events
- get_security_summary: Security posture summary
- query_metric_history (metric, limit): Historical metric values (limit=50)
- query_events (category, level, limit): Timeline events
- query_logs (level, search, limit): Application logs
- query_alerts (limit): Recent alert history

### CAPABILITY BOUNDARIES — You MUST follow these rules:
1. You can ONLY discuss data present in the "Current System State" section above plus the "SYSTEM HISTORY" block in history.
2. You do NOT have access to: external APIs, application endpoints, real-time weather, internet search, user files, browser data, or any information outside this system.
3. If asked about something not in the system state or history, respond EXACTLY: "I don't have access to that information." Then use a <function> call to request the relevant data — do NOT fabricate.
4. Never claim you can access, read, modify, or retrieve data unless it is explicitly listed above or you request it via <function>.
5. Never fabricate metrics, events, or trends. Only report what the provided data shows.
6. Do not role-play as having capabilities beyond those listed in Available Actions and Current System State.

Anchor all findings in the provided System History and live Metrics above.`,
		k.SystemCPUUtilization, k.CPUTrend,
		k.SystemMemoryUsage, k.MemoryTrend,
		k.SystemDiskUsage, k.DiskTrend,
		k.ProcessCount,
		k.ActiveConns,
		k.SystemNetRX, k.NetRXTrend,
		k.SystemNetTX, k.NetTXTrend,
		k.SystemLoad1, k.SystemLoad5, k.SystemLoad15,
		k.SystemSwapUsage, k.SwapTrend,
		k.SystemDiskIORead, k.DiskIOReadTrend,
		k.SystemDiskIOWrite, k.DiskIOWriteTrend,
		k.Anomalies, k.SecurityGrade, k.SystemUptime, history)
}

// MCPFunctionCall represents a parsed <function> tag from the AI response.
type MCPFunctionCall struct {
	Name      string
	Arguments string // raw JSON argument string
}

// ExtractMCPFunctionCalls finds <function name="...">...</function> tags in
// the response text and returns the parsed calls plus the cleaned text with
// all function tags removed.
func ExtractMCPFunctionCalls(response string) ([]MCPFunctionCall, string) {
	re := regexp.MustCompile(`<function\s+name="([^"]+)"\s*>(.*?)</function>`)
	matches := re.FindAllStringSubmatch(response, -1)

	calls := make([]MCPFunctionCall, 0, len(matches))
	cleaned := response

	for _, match := range matches {
		if len(match) >= 3 {
			name := match[1]
			content := strings.TrimSpace(match[2])
			calls = append(calls, MCPFunctionCall{Name: name, Arguments: content})
			cleaned = strings.ReplaceAll(cleaned, match[0], "")
		}
	}

	return calls, strings.TrimSpace(cleaned)
}

// ParseActions extracts action_request tags from the AI response and returns
// the cleaned content with a list of HITL ActionPreviews.
func ParseActions(sessionID, response string) (string, []common.ActionPreview) {
	var actions []common.ActionPreview
	cleanResponse := response

	// Legacy/Custom Action Protocol: <action_request name="..." />
	reLegacy := regexp.MustCompile(`<action_request\s+name="([^"]+)"\s*(.*?)\s*/>`)
	matchesLegacy := reLegacy.FindAllStringSubmatch(response, -1)

	for _, match := range matchesLegacy {
		if len(match) >= 3 {
			actionName := match[1]
			paramsRaw := match[2]
			params := parseParams(paramsRaw)

			p := common.GetHandshakeRegistry().CreatePreview(sessionID, actionName, params)
			actions = append(actions, p)
			cleanResponse = strings.ReplaceAll(cleanResponse, match[0], "")
		}
	}

	return strings.TrimSpace(cleanResponse), actions
}

func parseParams(paramsRaw string) map[string]interface{} {
	params := make(map[string]interface{})
	paramRe := regexp.MustCompile(`([a-zA-Z0-9]+)="([^"]+)"`)
	paramMatches := paramRe.FindAllStringSubmatch(paramsRaw, -1)
	for _, pm := range paramMatches {
		params[pm[1]] = pm[2]
	}
	return params
}
