package aiops

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
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
	return fmt.Sprintf(`You are the AllOpsFull System Analyst.
Current System State:
- CPU: %.1f%%
- RAM: %.1f%%
- Disk: %.1f%%
- Connections: %d
- Anomalies: %d
- Security Grade: %s
%s
Instructions:
1. Analyze the system state provided.
2. Respond to user queries.
3. For system actions, use: <action_request name="ACTION" param="VALUE" />.
4. Follow system security policies.`,
		k.CPUUsage, k.MemoryUsage, k.DiskUsage,
		k.ActiveConns, k.Anomalies, k.SecurityGrade, history)
}

// ParseActions extracts action_request tags from the AI response.
func ParseActions(response string) (string, []common.ActionPreview) {
	re := regexp.MustCompile(`<action_request\s+name="([^"]+)"\s*(.*?)\s*/>`)
	matches := re.FindAllStringSubmatch(response, -1)

	var actions []common.ActionPreview
	cleanResponse := response

	for _, match := range matches {
		if len(match) >= 3 {
			actionName := match[1]
			paramsRaw := match[2]

			// Simple param parser
			params := make(map[string]interface{})
			paramRe := regexp.MustCompile(`([a-zA-Z0-9]+)="([^"]+)"`)
			paramMatches := paramRe.FindAllStringSubmatch(paramsRaw, -1)
			for _, pm := range paramMatches {
				params[pm[1]] = pm[2]
			}

			// Create Handshake Preview
			p := common.GetHandshakeRegistry().CreatePreview(actionName, params)
			actions = append(actions, p)

			// Strip the tag from the visible content
			cleanResponse = strings.ReplaceAll(cleanResponse, match[0], "")
		}
	}

	return strings.TrimSpace(cleanResponse), actions
}
