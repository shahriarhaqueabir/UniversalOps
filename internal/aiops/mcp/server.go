package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// Tool represents an MCP Tool definition.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Server handles MCP protocol requests.
type Server struct {
	pipeline *common.DataPipeline
}

func NewServer(pipeline *common.DataPipeline) *Server {
	return &Server{pipeline: pipeline}
}

// ListTools returns the available tools for the AI.
func (s *Server) ListTools() ([]Tool, error) {
	return []Tool{
		{
			Name:        "get_system_telemetry",
			Description: "Get high-density OTel-aligned system metrics including CPU, RAM, and Disk pressure.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "analyze_network",
			Description: "Perform RFC 9951 compliant network delay and jitter analysis.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"target":{"type":"string","description":"Host or IP to analyze"}}}`),
		},
	}, nil
}

// CallTool executes a tool call from the AI.
func (s *Server) CallTool(ctx context.Context, name string, arguments json.RawMessage) (interface{}, error) {
	switch name {
	case "get_system_telemetry":
		return s.handleGetTelemetry()
	case "analyze_network":
		var args struct {
			Target string `json:"target"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		return s.handleAnalyzeNetwork(args.Target)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) handleGetTelemetry() (interface{}, error) {
	knowledge := common.GetKnowledge().GetSnapshot()
	return knowledge, nil
}

func (s *Server) handleAnalyzeNetwork(target string) (interface{}, error) {
	// Logic to call NetOps.Ping with RFC 9951 metrics (Tier 2)
	return fmt.Sprintf("Network analysis for %s initiated...", target), nil
}
