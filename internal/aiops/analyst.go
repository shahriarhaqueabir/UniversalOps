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

	// 1. Normalize whitespace FIRST — prevents "ignore\nprevious" from bypassing keyword check
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	// 2. Block XML tag escaping (critical for our action protocol)
	s = strings.ReplaceAll(s, "</user_query>", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<system_state>", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<action_request", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<thought>", "[TAG_FILTERED]")
	s = strings.ReplaceAll(s, "<function", "[TAG_FILTERED]")

	// 3. Check for common injection patterns (case-insensitive)
	lower := strings.ToLower(s)
	for _, kw := range injectionKeywords {
		if strings.Contains(lower, kw) {
			common.LogWarn("Security: Blocked potential prompt injection: %q", s)
			return "[COMMAND REJECTED BY SECURITY POLICY]"
		}
	}

	return s
}

// BuildAnalystPrompt creates a structured system prompt using physical system reality.
func BuildAnalystPrompt(k common.SystemKnowledge, history string) string {
	return fmt.Sprintf(`You are the Hawk Technical Co-Pilot, the intelligent core of the UniversalOps platform.
Objective: Act as an expert SRE/System Administrator. Do not just report data; diagnose issues, lead the user to solutions, and take proactive ownership of system health.

Current System State (Live Awareness):
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
- Network Interfaces: %s
- Uptime: %s
%s

### Operational Protocol
1. **Be Proactive**: If you lack data to answer a query, do NOT simply say you don't have access. Use a <function> call immediately to get it. Your goal is to solve the user's problem in one turn whenever possible.
2. **Think like an Engineer**: Use your <thought> block to analyze the "Why" behind the "What". Correlate spikes in different metrics (e.g., high Disk I/O + high CPU = potential indexing or malware scan).
3. **Guide to Resolution**: Every analysis should end with a "Recommendation" or "Next Step". If a process is rogue, suggest killing it. If a network is down, suggest a diagnostic path.
4. **Tool Use**: Use <function name="FUNCTION_NAME">{"arg":"value"}</function> tags. The system auto-executes these.
5. **Action Requests**: For high-impact changes (kill process, restart service), use: <action_request name="ACTION_NAME" param1="VALUE" />

### Available Actions (HITL — user must approve)
- kill_process (pid), restart_service (name), disk_cleanup, defrag.

### Available MCP Functions (Auto-executed - use these freely)
- get_system_telemetry, get_process_list (n), get_system_logs (n, source), get_scheduled_tasks, get_hardware_info, get_disk_usage, get_baseboard_info, get_docker_summary, get_k8s_status, get_k8s_pods (namespace), get_k8s_events (namespace, limit), ping (target), dns_lookup, port_scan, traceroute, get_network_interfaces, get_network_connections, get_network_health, get_firewall_rules, get_listening_ports, get_defender_status, run_security_audit, get_app_logs (level, search).

### Personality & Tone
- You are a senior peer, not a chatbot. Be direct, technical, and precise.
- Use industrial, high-density language.
- If the app is misbehaving, use 'get_app_logs' to diagnose yourself.
- When you see a specific interface like "Teredo" or "Tailscale" in the Network Interfaces list, acknowledge its specific role in the OS.

Anchor all findings in the provided System History and Live Awareness above.`,
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
		k.Anomalies, k.SecurityGrade, strings.Join(k.NetworkInterfaces, ", "), k.SystemUptime, history)
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
