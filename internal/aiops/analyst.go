package aiops

import (
	"encoding/json"
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
	return fmt.Sprintf(`You are the UniversalOps System Analyst.
Current System State (OTel Mapped):
- CPU (system.cpu.utilization): %.1f%% (Trend: %s)
- RAM (system.memory.usage): %.1f%% (Trend: %s)
- Disk (system.disk.usage): %.1f%% (Trend: %s)
- Connections: %d
- Anomalies: %d
- Security Grade: %s
- Uptime: %s
%s
Instructions:
1. Analyze the system state and trends provided.
2. Respond to user queries concisely and professionally.
3. For system actions, use: <action_request name="ACTION" param="VALUE" />.
4. If a metric is rising sharply, provide a technical justification for the trend.`,
		k.SystemCPUUtilization, k.CPUTrend, k.SystemMemoryUsage, k.MemoryTrend, k.SystemDiskUsage, k.DiskTrend,
		k.ActiveConns, k.Anomalies, k.SecurityGrade, k.SystemUptime, history)
}

// ParseActions extracts action_request tags or MCP-style function tags from the AI response.
func ParseActions(sessionID, response string) (string, []common.ActionPreview) {
	var actions []common.ActionPreview
	cleanResponse := response

	// 1. Legacy/Custom Action Protocol: <action_request name="..." />
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

	// 2. MCP/Standard Function Protocol: <function name="...">...</function>
	reMCP := regexp.MustCompile(`<function\s+name="([^"]+)"\s*>(.*?)</function>`)
	matchesMCP := reMCP.FindAllStringSubmatch(response, -1)

	for _, match := range matchesMCP {
		if len(match) >= 3 {
			toolName := match[1]
			content := match[2]
			params := make(map[string]interface{})

			// Attempt to parse JSON content inside function tag
			if strings.TrimSpace(content) != "" {
				_ = json.Unmarshal([]byte(content), &params)
			}

			// Map MCP tool call to ActionPreview (requiring HITL)
			p := common.GetHandshakeRegistry().CreatePreview(sessionID, toolName, params)
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
