package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ollama/ollama/api"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/shahriarhaqueabir/UniversalOps/internal/aiops"
	"github.com/shahriarhaqueabir/UniversalOps/internal/aiops/mcp"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// titleCaser replaces the deprecated strings.Title (removed guidance since
// Go 1.18); see CROSS-1.
var titleCaser = cases.Title(language.English)

// systemContextCache holds a recently-built system context snapshot to avoid
// redundant DB reads on consecutive chat messages.
type systemContextCache struct {
	mu        sync.Mutex
	snapshot  string
	timestamp time.Time
}

// AIOps exposes AI operations bindings to the frontend.
type AIOps struct {
	ctx          context.Context
	pipeline     *common.DataPipeline
	knowledge    *KnowledgeAPI
	capabilities *common.CapabilityRegistry
	pipelineAPI  *PipelineAPI
	sysOps       *SysOps
	netOps       *NetOps
	devOps       *DevOps
	dataDir      string

	// Configurable models per capability
	AnalystModel   string
	OptimizerModel string

	// MCP Server: exposes 24 on-demand system/network/security/storage tools
	// to the AI model via <function name="...">...</function> tags.
	mcpServer *mcp.Server

	contextCache *systemContextCache
}

// NewAIOps creates a new AIOps facade.
func NewAIOps(ctx context.Context, pipeline *common.DataPipeline, knowledge *KnowledgeAPI, capabilities *common.CapabilityRegistry, pipelineAPI *PipelineAPI, sysOps *SysOps, netOps *NetOps, devOps *DevOps, dataDir string) *AIOps {
	return &AIOps{
		ctx:          ctx,
		pipeline:     pipeline,
		knowledge:    knowledge,
		capabilities: capabilities,
		pipelineAPI:  pipelineAPI,
		sysOps:       sysOps,
		netOps:       netOps,
		devOps:       devOps,
		dataDir:      dataDir,
		mcpServer:    mcp.NewServer(pipeline, netOps, devOps),
		contextCache: &systemContextCache{},
	}
}

// requestContext returns a.ctx when non-nil, otherwise context.Background().
// This avoids repeating the nil-check + fallback pattern throughout the facade.
func (a *AIOps) requestContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func (a *AIOps) SetDataDir(dir string) {
	defer common.RecoverPanic()
	a.dataDir = dir
}

// SetCompanionName persists the AI assistant name.
func (a *AIOps) SetCompanionName(name string) {
	defer common.RecoverPanic()
	s := common.GetStorage()
	if s != nil {
		s.UpsertSetting("companionName", name)
	}
}

// LoadModels retrieves persisted model overrides.
func (a *AIOps) LoadModels() {
	defer common.RecoverPanic()
	s := common.GetStorage()
	if s == nil {
		return
	}
	a.AnalystModel, _ = s.GetSetting("model_analyst")
	a.OptimizerModel, _ = s.GetSetting("model_optimizer")
}

// SetModelForCapability updates and persists a model override for a specific capability.
func (a *AIOps) SetModelForCapability(capability, model string) {
	defer common.RecoverPanic()
	s := common.GetStorage()
	if s == nil {
		return
	}
	key := "model_" + capability
	s.UpsertSetting(key, model)

	if capability == "analyst" {
		a.AnalystModel = model
	} else if capability == "optimizer" {
		a.OptimizerModel = model
	}
}

// GetCompanionName retrieves the persisted AI assistant name.
func (a *AIOps) GetCompanionName() string {
	defer common.RecoverPanic()
	s := common.GetStorage()
	if s == nil {
		return "Hawk"
	}
	val, err := s.GetSetting("companionName")
	if err != nil || val == "" {
		return "Hawk"
	}
	return val
}

// ChatResponse contains the AI message and any proposed actions.
type ChatResponse struct {
	Content string                 `json:"content"`
	Actions []common.ActionPreview `json:"actions,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Chat sends a message to the Ollama chat API and returns the response.
func (a *AIOps) Chat(sessionID, message string) ChatResponse {
	defer common.RecoverPanic()
	// 1. Prepare Context
	knowledge := a.knowledge.GetSnapshot()

	storage := common.GetStorage()
	var historyContext string
	var chatHistory []aiops.ChatMessage

	// Load model override from settings for this specific request to avoid global leakage
	activeModel := a.AnalystModel
	if storage != nil {
		if m, err := storage.GetSetting("model_analyst"); err == nil && m != "" {
			activeModel = m
		}
		// Load historical metrics for prompt context - Expanded to 100 points (~5 mins at 3s)
		cpuHistory, _ := storage.GetMetricHistory(common.MetricCPU, 100)
		memHistory, _ := storage.GetMetricHistory(common.MetricMem, 100)
		diskHistory, _ := storage.GetMetricHistory(common.MetricDisk, 100)
		netRxHistory, _ := storage.GetMetricHistory(common.MetricNetRX, 100)
		netTxHistory, _ := storage.GetMetricHistory(common.MetricNetTX, 100)
		load1History, _ := storage.GetMetricHistory(common.MetricLoad1, 100)
		load5History, _ := storage.GetMetricHistory(common.MetricLoad5, 100)
		load15History, _ := storage.GetMetricHistory(common.MetricLoad15, 100)
		swapHistory, _ := storage.GetMetricHistory(common.MetricSwap, 100)
		diskIOReadHistory, _ := storage.GetMetricHistory(common.MetricDiskIORead, 100)
		diskIOWriteHistory, _ := storage.GetMetricHistory(common.MetricDiskIOWrite, 100)

		// Load recent system events (last 50)
		events, _ := storage.QueryEvents("", "", 50, 0)
		var eventStrings []string
		for _, e := range events {
			eventStrings = append(eventStrings, fmt.Sprintf("[%s] %s: %s", e.Timestamp.Format("15:04:05"), e.Title, e.Detail))
		}

		// Load recent anomalies (last 20)
		anoms := a.DetectAnomalies()
		var anomStrings []string
		for i, anom := range anoms {
			if i >= 20 {
				break
			}
			anomStrings = append(anomStrings, fmt.Sprintf("%s: %s (Value: %.1f)", anom.Timestamp, anom.Metric, anom.Value))
		}

		historyContext = fmt.Sprintf("\n--- SYSTEM HISTORY (Snapshot: %s) ---\n"+
			"CPU Summary: %s\n"+
			"RAM Summary: %s\n"+
			"Disk Summary: %s\n"+
			"Net RX Summary: %s\n"+
			"Net TX Summary: %s\n"+
			"Load 1m Summary: %s\n"+
			"Load 5m Summary: %s\n"+
			"Load 15m Summary: %s\n"+
			"Swap Summary: %s\n"+
			"Disk IO Read Summary: %s\n"+
			"Disk IO Write Summary: %s\n\n"+
			"Recent System Events:\n%s\n\n"+
			"Recent Anomalies:\n%s\n"+
			"------------------------------------\n",
			time.Now().Format("15:04:05.000"),
			a.summarizeMetrics("CPU", cpuHistory),
			a.summarizeMetrics("RAM", memHistory),
			a.summarizeMetrics("Disk", diskHistory),
			a.summarizeMetrics("NetRX", netRxHistory),
			a.summarizeMetrics("NetTX", netTxHistory),
			a.summarizeMetrics("Load1m", load1History),
			a.summarizeMetrics("Load5m", load5History),
			a.summarizeMetrics("Load15m", load15History),
			a.summarizeMetrics("Swap", swapHistory),
			a.summarizeMetrics("DiskIORead", diskIOReadHistory),
			a.summarizeMetrics("DiskIOWrite", diskIOWriteHistory),
			strings.Join(eventStrings, "\n"),
			strings.Join(anomStrings, "\n"))

		// Load conversation history if sessionID is provided
		if sessionID != "" {
			msgs, _ := storage.GetMessages(sessionID)
			for _, m := range msgs {
				chatHistory = append(chatHistory, aiops.ChatMessage{
					Role:    m.Role,
					Content: m.Content,
				})
			}
			// ENSURE: Context window truncation to prevent Ollama crashes on long sessions
			chatHistory = aiops.TruncateHistory(chatHistory, 20000)
		}
	}

	systemPrompt := aiops.BuildAnalystPrompt(knowledge, historyContext)
	wrappedMessage := fmt.Sprintf("<user_query>%s</user_query>", aiops.SanitizeInput(message))

	// Construct full message list: [System, ...History, CurrentUser]
	messages := []aiops.ChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, chatHistory...)
	messages = append(messages, aiops.ChatMessage{Role: "user", Content: wrappedMessage})

	// 2. Execute Chat with per-request timeout
	chatParent := a.requestContext()
	chatCtx, cancel := context.WithTimeout(chatParent, 120*time.Second)
	defer cancel()

	chatStart := time.Now()

	// CRITICAL: We pass the locally resolved activeModel to ensure session isolation
	// and prevent background diagnostics from changing the user's preferred model.
	rawResponse, err := aiops.ChatWithModelAndContext(chatCtx, messages, activeModel)

	chatElapsed := time.Since(chatStart)

	if err != nil {
		common.LogWarn("AI Chat failed after %v: %v", chatElapsed, err)
		return ChatResponse{Content: "AI assistant unavailable. Please verify Ollama is running and the model is downloaded."}
	}

	common.LogInfo("AI Chat completed in %v (session=%s)", chatElapsed, sessionID)

	// 3. Parse and auto-execute MCP function calls
	content, actions := a.resolveMCPCalls(chatCtx, rawResponse, systemPrompt, chatHistory, wrappedMessage, sessionID, activeModel)

	return ChatResponse{
		Content: content,
		Actions: actions,
	}
}

// resolveMCPCalls extracts MCP <function> tags from the raw AI response,
// executes them via the MCP server, feeds results back to the AI for
// synthesis, and returns the final parsed content and HITL action requests.
// If no function calls are present, it falls through to ParseActions.
func (a *AIOps) resolveMCPCalls(ctx context.Context, rawResponse, systemPrompt string, chatHistory []aiops.ChatMessage, wrappedMessage, sessionID, activeModel string) (string, []common.ActionPreview) {
	mcpCalls, cleanResponse := aiops.ExtractMCPFunctionCalls(rawResponse)

	// No MCP calls — parse for HITL action_requests only
	if len(mcpCalls) == 0 || a.mcpServer == nil {
		return aiops.ParseActions(sessionID, rawResponse)
	}

	// Execute all MCP functions and collect results
	common.LogInfo("MCP: executing %d function call(s) from session %s", len(mcpCalls), sessionID)

	var resultsBuilder strings.Builder
	resultsBuilder.WriteString("SYSTEM: Function call results:\n\n")

	for _, call := range mcpCalls {
		var argsJSON json.RawMessage
		if call.Arguments != "" {
			argsJSON = json.RawMessage(call.Arguments)
		}

		result, err := a.mcpServer.CallTool(ctx, call.Name, argsJSON)
		if err != nil {
			fmt.Fprintf(&resultsBuilder, "[%s] Error: %v\n\n", call.Name, err)
			common.LogWarn("MCP: tool %q failed: %v", call.Name, err)
			continue
		}
		resultBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(&resultsBuilder, "[%s] Result (marshal error: %v)\n\n", call.Name, err)
			continue
		}
		// Truncate very large responses (e.g. full process lists) to avoid
		// blowing the LLM context window.
		if len(resultBytes) > 4000 {
			fmt.Fprintf(&resultsBuilder, "[%s] Result (%d bytes, truncated):\n%s\n\n", call.Name, len(resultBytes), string(resultBytes[:4000]))
		} else {
			fmt.Fprintf(&resultsBuilder, "[%s] Result:\n%s\n\n", call.Name, string(resultBytes))
		}
	}

	// Build continuation messages: original context + AI's first response + tool results
	contMessages := []aiops.ChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	contMessages = append(contMessages, chatHistory...)
	contMessages = append(contMessages, aiops.ChatMessage{Role: "user", Content: wrappedMessage})

	// Only include the non-function parts of the AI's first response
	if trimmed := strings.TrimSpace(cleanResponse); trimmed != "" {
		contMessages = append(contMessages, aiops.ChatMessage{Role: "assistant", Content: trimmed})
	}

	// Feed tool results as a user message
	contMessages = append(contMessages, aiops.ChatMessage{Role: "user", Content: resultsBuilder.String()})

	// Get final synthesized response from the AI
	common.LogInfo("MCP: synthesizing %d tool result(s) with LLM", len(mcpCalls))
	finalResponse, err := aiops.ChatWithModelAndContext(ctx, contMessages, activeModel)
	if err != nil {
		common.LogWarn("MCP: synthesis failed: %v — falling back to clean response", err)
		content, actions := aiops.ParseActions(sessionID, cleanResponse)
		return content, actions
	}

	// Parse the synthesis for any remaining <action_request> tags (HITL)
	return aiops.ParseActions(sessionID, finalResponse)
}

// ── Context Cache Helpers ──────────────────────────────────────────────

const contextCacheTTL = 5 * time.Second

// isTrivialMessage returns true when the message is short and contains no
// system-related keywords — likely a greeting or small talk. For these messages
// we skip heavy context loading to reduce first-token latency.
func isTrivialMessage(msg string) bool {
	if len(msg) > 20 {
		return false
	}
	lower := strings.ToLower(msg)
	keywords := []string{
		"cpu", "mem", "disk", "network", "anomaly", "error", "alert",
		"health", "service", "process", "system", "status", "log",
		"event", "threat", "scan", "port", "latency", "utilization",
		"performance", "bottleneck", "memory", "storage",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return false
		}
	}
	return true
}

// getCachedSnapshot returns the cached system context snapshot if it's still
// within the TTL and forceRefresh is false. Returns ("", false) if no valid
// cache entry is available.
func (a *AIOps) getCachedSnapshot(forceRefresh bool) (string, bool) {
	cache := a.contextCache
	if cache == nil {
		return "", false
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if !forceRefresh && cache.snapshot != "" && time.Since(cache.timestamp) < contextCacheTTL {
		return cache.snapshot, true
	}
	return "", false
}

// setCachedSnapshot atomically stores a system context snapshot in the cache.
func (a *AIOps) setCachedSnapshot(snapshot string) {
	cache := a.contextCache
	if cache == nil {
		return
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.snapshot = snapshot
	cache.timestamp = time.Now()
}

// buildSystemContextSnapshot loads metrics, events and anomalies in parallel
// and returns a formatted context string. Results are cached with a 5-second TTL.
func (a *AIOps) buildSystemContextSnapshot(forceRefresh bool) string {
	if snap, ok := a.getCachedSnapshot(forceRefresh); ok {
		return snap
	}

	storage := common.GetStorage()
	if storage == nil {
		return ""
	}

	var (
		cpuHistory  []float64
		memHistory  []float64
		diskHistory []float64
		eventStrs   []string
		anomStrs    []string
	)

	// Load metrics, events, and anomalies concurrently
	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		cpuHistory, _ = storage.GetMetricHistory(common.MetricCPU, 100)
	}()
	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		memHistory, _ = storage.GetMetricHistory(common.MetricMem, 100)
	}()
	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		diskHistory, _ = storage.GetMetricHistory(common.MetricDisk, 100)
	}()
	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		events, _ := storage.QueryEvents("", "", 50, 0)
		for _, e := range events {
			eventStrs = append(eventStrs, fmt.Sprintf("[%s] %s: %s", e.Timestamp.Format("15:04:05"), e.Title, e.Detail))
		}
	}()
	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		anoms := a.DetectAnomalies()
		for i, anom := range anoms {
			if i >= 20 {
				break
			}
			anomStrs = append(anomStrs, fmt.Sprintf("%s: %s (Value: %.1f)", anom.Timestamp, anom.Metric, anom.Value))
		}
	}()

	wg.Wait()

	snapshot := fmt.Sprintf("\n--- SYSTEM HISTORY (Snapshot: %s) ---\n"+
		"CPU Summary: %s\n"+
		"RAM Summary: %s\n"+
		"Disk Summary: %s\n\n"+
		"Recent System Events:\n%s\n\n"+
		"Recent Anomalies:\n%s\n"+
		"------------------------------------\n",
		time.Now().Format("15:04:05.000"),
		a.summarizeMetrics("CPU", cpuHistory),
		a.summarizeMetrics("RAM", memHistory),
		a.summarizeMetrics("Disk", diskHistory),
		strings.Join(eventStrs, "\n"),
		strings.Join(anomStrs, "\n"))

	a.setCachedSnapshot(snapshot)

	return snapshot
}

// ── Streaming Chat ─────────────────────────────────────────────────────

// ChatStream sends a message to the Ollama chat API and streams each
// response token as a Wails "chat:token" event. Emits "chat:done" when
// complete, or "chat:error" on failure. Returns the parsed ChatResponse.
func (a *AIOps) ChatStream(sessionID, message string) ChatResponse {
	defer common.RecoverPanic()
	if sessionID == "" || message == "" {
		return ChatResponse{Content: "Error: sessionID and message are required"}
	}

	knowledge := a.knowledge.GetSnapshot()
	storage := common.GetStorage()

	// Resolve model once (no need to re-read from DB in goroutine)
	activeModel := a.AnalystModel
	if storage != nil {
		if m, err := storage.GetSetting("model_analyst"); err == nil && m != "" {
			activeModel = m
		}
	}

	// ── Workflow: Loading system context ──
	a.emitWorkflowEvent(sessionID, "context", "running", "Loading system metrics, events, and anomaly history...")

	// Decide whether to load heavy context
	// NOTE: isTrivialMessage returns true for short/greeting messages.
	// When trivial, we want to use the cached context (forceRefresh=false)
	// to reduce first-token latency. When non-trivial, force a fresh snapshot.
	trivial := isTrivialMessage(message)
	historyContext := a.buildSystemContextSnapshot(!trivial)

	a.emitWorkflowEvent(sessionID, "context", "completed", "System context loaded — CPU, RAM, Disk history + events + anomalies")

	// Load chat history (must happen on this goroutine; storage is not
	// guaranteed goroutine-safe for sequential access patterns).
	var chatHistory []aiops.ChatMessage
	if storage != nil && sessionID != "" {
		msgs, _ := storage.GetMessages(sessionID)
		for _, m := range msgs {
			chatHistory = append(chatHistory, aiops.ChatMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
		chatHistory = aiops.TruncateHistory(chatHistory, 20000)
	}

	systemPrompt := aiops.BuildAnalystPrompt(knowledge, historyContext)
	wrappedMessage := fmt.Sprintf("<user_query>%s</user_query>", aiops.SanitizeInput(message))

	messages := []aiops.ChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, chatHistory...)
	messages = append(messages, aiops.ChatMessage{Role: "user", Content: wrappedMessage})

	// Emit a start event so the frontend can correlate
	wailsruntime.EventsEmit(a.ctx, "chat:start", map[string]string{
		"sessionId": sessionID,
	})

	// Apply a per-request timeout so a hung Ollama process doesn't
	// block the facade indefinitely. The underlying HTTP client has its
	// own 60s timeout, but this gives us an extra layer of protection.
	chatParent := a.requestContext()
	chatCtx, cancel := context.WithTimeout(chatParent, 120*time.Second)
	defer cancel()

	chatStart := time.Now()

	// ── Workflow: Querying AI model ──
	modelLabel := activeModel
	if modelLabel == "" {
		modelLabel = "default"
	}
	a.emitWorkflowEvent(sessionID, "inference", "running", "Querying Ollama model: "+modelLabel)

	// Execute streaming chat — the onToken callback emits events inline
	// while the Ollama API delivers chunks. Events reach the frontend
	// immediately because Wails processes them asynchronously.
	rawResponse, err := aiops.ChatStreamWithModelAndContext(chatCtx, messages, activeModel, func(token string) {
		wailsruntime.EventsEmit(a.ctx, "chat:token", map[string]string{
			"sessionId": sessionID,
			"token":     token,
		})
	})

	chatElapsed := time.Since(chatStart)

	if err != nil {
		common.LogWarn("AI ChatStream failed after %v: %v", chatElapsed, err)
		a.emitWorkflowEvent(sessionID, "inference", "error", "AI stream generation failed")
		wailsruntime.EventsEmit(a.ctx, "chat:error", map[string]string{
			"sessionId": sessionID,
			"error":     "AI stream generation failed",
		})
		return ChatResponse{Content: "AI assistant unavailable. Please verify Ollama is running and try again."}
	}

	a.emitWorkflowEvent(sessionID, "inference", "completed",
		fmt.Sprintf("Model responded in %.1fs (~%d tokens)", chatElapsed.Seconds(), len(strings.Fields(rawResponse))))
	common.LogInfo("AI ChatStream completed in %v (session=%s, tokens=%d)", chatElapsed, sessionID, len(strings.Fields(rawResponse)))

	// ── MCP Tool Resolution (Auto-Execute & Re-Synthesize) ──────────
	mcpCalls, cleanResponse := aiops.ExtractMCPFunctionCalls(rawResponse)

	var finalContent string
	var actions []common.ActionPreview

	if len(mcpCalls) > 0 && a.mcpServer != nil {
		common.LogInfo("MCP: executing %d function call(s) from streaming session %s", len(mcpCalls), sessionID)

		// Execute all MCP functions
		var resultsBuilder strings.Builder
		resultsBuilder.WriteString("SYSTEM: Function call results:\n\n")

		for _, call := range mcpCalls {
			a.emitWorkflowEvent(sessionID, "mcp", "running", "Executing MCP tool: "+call.Name)

			var argsJSON json.RawMessage
			if call.Arguments != "" {
				argsJSON = json.RawMessage(call.Arguments)
			}

			toolStart := time.Now()
			result, err := a.mcpServer.CallTool(chatCtx, call.Name, argsJSON)
			toolElapsed := time.Since(toolStart)

			if err != nil {
				fmt.Fprintf(&resultsBuilder, "[%s] Error: %v\n\n", call.Name, err)
				a.emitWorkflowEvent(sessionID, "mcp", "error", "MCP tool "+call.Name+" failed: "+err.Error())
				common.LogWarn("MCP: tool %q failed: %v", call.Name, err)
				continue
			}
			a.emitWorkflowEvent(sessionID, "mcp", "completed",
				fmt.Sprintf("MCP tool %s completed in %.1fs", call.Name, toolElapsed.Seconds()))
			resultBytes, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Fprintf(&resultsBuilder, "[%s] Result (marshal error: %v)\n\n", call.Name, err)
				continue
			}
			if len(resultBytes) > 4000 {
				fmt.Fprintf(&resultsBuilder, "[%s] Result (%d bytes, truncated):\n%s\n\n", call.Name, len(resultBytes), string(resultBytes[:4000]))
			} else {
				fmt.Fprintf(&resultsBuilder, "[%s] Result:\n%s\n\n", call.Name, string(resultBytes))
			}
		}

		// Build continuation messages
		contMessages := []aiops.ChatMessage{
			{Role: "system", Content: systemPrompt},
		}
		contMessages = append(contMessages, chatHistory...)
		contMessages = append(contMessages, aiops.ChatMessage{Role: "user", Content: wrappedMessage})
		if trimmed := strings.TrimSpace(cleanResponse); trimmed != "" {
			contMessages = append(contMessages, aiops.ChatMessage{Role: "assistant", Content: trimmed})
		}
		contMessages = append(contMessages, aiops.ChatMessage{Role: "user", Content: resultsBuilder.String()})

		// Emit a brief status token so the frontend indicates work is happening
		wailsruntime.EventsEmit(a.ctx, "chat:token", map[string]string{
			"sessionId": sessionID,
			"token":     "\n\n*Analyzing tool results...*\n\n",
		})

		// ── Workflow: Synthesizing results ──
		a.emitWorkflowEvent(sessionID, "synthesis", "running", "Synthesizing MCP tool results into final response...")

		// Stream the synthesis (short, typically <200 tokens)
		synthesis, err := aiops.ChatStreamWithModelAndContext(chatCtx, contMessages, activeModel, func(token string) {
			wailsruntime.EventsEmit(a.ctx, "chat:token", map[string]string{
				"sessionId": sessionID,
				"token":     token,
			})
		})
		if err != nil {
			common.LogWarn("MCP: streaming synthesis failed: %v — falling back to clean response", err)
			a.emitWorkflowEvent(sessionID, "synthesis", "error", "Synthesis failed, using clean response")
			finalContent, actions = aiops.ParseActions(sessionID, cleanResponse)
		} else {
			a.emitWorkflowEvent(sessionID, "synthesis", "completed", "Tool results synthesized into natural language response")
			finalContent, actions = aiops.ParseActions(sessionID, synthesis)
		}
	} else {
		// No MCP calls — standard ParseActions only
		finalContent, actions = aiops.ParseActions(sessionID, rawResponse)
	}

	// Persist messages (save the clean, final content)
	if storage != nil && sessionID != "" {
		if err := storage.InsertMessage(sessionID, "user", message); err != nil {
			common.LogError("AIOps: failed to persist user message: %v", err)
		}
		if err := storage.InsertMessage(sessionID, "assistant", finalContent); err != nil {
			common.LogError("AIOps: failed to persist assistant message: %v", err)
		}
	}

	wailsruntime.EventsEmit(a.ctx, "chat:done", map[string]string{
		"sessionId":  sessionID,
		"content":    finalContent,
		"latency_ms": fmt.Sprintf("%.0f", chatElapsed.Seconds()*1000),
	})

	a.emitWorkflowEvent(sessionID, "complete", "completed",
		fmt.Sprintf("Analysis complete — total latency %.1fs", chatElapsed.Seconds()))

	// Also save actions as JSON payload
	if len(actions) > 0 {
		actionJSON, _ := json.Marshal(actions)
		common.LogInfo("ChatStream actions for session %s: %s", sessionID, string(actionJSON))
	}

	return ChatResponse{
		Content: finalContent,
		Actions: actions,
	}
}

// GenerateReport creates a formatted text report from the given sections.
func (a *AIOps) GenerateReport(sections []string) string {
	defer common.RecoverPanic()
	var reportSections []aiops.ReportSection
	for _, title := range sections {
		title = aiops.SanitizeInput(title)
		prompt := "Generate a brief operations report section for: " + title +
			". Include key metrics and observations based on recent system data."
		genParent := a.requestContext()
		genCtx, cancel := context.WithTimeout(genParent, 120*time.Second)
		defer cancel()
		genStart := time.Now()
		resp, err := aiops.ChatWithContext(genCtx, []aiops.ChatMessage{
			{Role: "system", Content: "You are a system operations analyst. Be concise and factual."},
			{Role: "user", Content: prompt},
		})
		genElapsed := time.Since(genStart)
		if err != nil {
			common.LogWarn("GenerateReport section %q failed after %v: %v", title, genElapsed, err)
		} else {
			common.LogInfo("GenerateReport section %q completed in %v", title, genElapsed)
		}
		content := ""
		if err == nil {
			content = resp
		} else {
			content = "Content generation unavailable: " + err.Error()
		}
		reportSections = append(reportSections, aiops.ReportSection{
			Title:   title,
			Content: content,
		})
	}
	return aiops.GenerateReport(reportSections)
}

// ListMCPTools returns the full list of available MCP tool definitions for
// the frontend to display or introspect.
func (a *AIOps) ListMCPTools() []mcp.Tool {
	defer common.RecoverPanic()
	if a.mcpServer == nil {
		return nil
	}
	tools, err := a.mcpServer.ListTools()
	if err != nil {
		common.LogWarn("AIOps: ListMCPTools failed: %v", err)
		return nil
	}
	return tools
}

// ── Live Context: Workflow & Data Stream Events ──────────────────────────

// emitWorkflowEvent pushes a real-time AI workflow event to the frontend
// Live Context view via the "ai:workflow" Wails event channel.
func (a *AIOps) emitWorkflowEvent(sessionID, stage, status, detail string) {
	if a.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(a.ctx, "ai:workflow", AIWorkflowEvent{
		SessionID: sessionID,
		Stage:     stage,
		Status:    status,
		Detail:    detail,
		Timestamp: time.Now().Format("15:04:05.000"),
	})
}

// GetDataStreamSnapshot returns a snapshot of every tracked metric in the
// pipeline ring buffer for the frontend Live Context data stream view.
func (a *AIOps) GetDataStreamSnapshot() []DataStreamMetric {
	defer common.RecoverPanic()
	if a.pipeline == nil {
		return nil
	}
	series := a.pipeline.AllSeries()
	if len(series) == 0 {
		return nil
	}

	result := make([]DataStreamMetric, 0, len(series))
	now := time.Now().Format("15:04:05.000")

	for name, ts := range series {
		if ts.Count() == 0 {
			continue
		}
		trend := "stable"
		trendInfo := a.pipeline.GetTrend(name)
		// Use correlation as a confidence proxy (|R| > 0.5 = moderate+ trend)
		strong := trendInfo.Correlation > 0.5 || trendInfo.Correlation < -0.5
		if trendInfo.Direction == common.TrendRising && strong {
			trend = "rising"
		} else if trendInfo.Direction == common.TrendFalling && strong {
			trend = "falling"
		}
		result = append(result, DataStreamMetric{
			Name:      name,
			Unit:      ts.Unit,
			LastValue: ts.Last(),
			Samples:   ts.Count(),
			Trend:     trend,
			UpdatedAt: now,
		})
	}

	// Sort alphabetically for deterministic display
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// GetOllamaStatus returns the current Ollama service status.
func (a *AIOps) GetOllamaStatus() OllamaStatus {
	defer common.RecoverPanic()
	status, err := aiops.CheckOllamaWithContext(a.ctx)

	// Use the centralized CapabilityRegistry for binary detection
	binaryExists := false
	if a.capabilities != nil {
		binaryExists = a.capabilities.IsAvailable(common.CapOllama)
	}

	if err != nil {
		return OllamaStatus{
			Available:    false,
			BinaryExists: binaryExists,
			Error:        err.Error(),
		}
	}
	return OllamaStatus{
		Available:       status.Available,
		BinaryExists:    binaryExists,
		Model:           status.Model,
		Version:         status.Version,
		AvailableModels: status.AvailableModels,
	}
}

// GetModelfile reads the local universalops.modelfile and returns its content.
func (a *AIOps) GetModelfile() (string, error) {
	defer common.RecoverPanic()
	path := filepath.Join(a.dataDir, "universalops.modelfile")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read universalops.modelfile: %w", err)
	}
	return string(content), nil
}

// SaveModelfile writes the given content to the local universalops.modelfile.
func (a *AIOps) SaveModelfile(content string) error {
	defer common.RecoverPanic()
	common.LogInfo("AIOps: Saving updated universalops.modelfile")
	path := filepath.Join(a.dataDir, "universalops.modelfile")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to save universalops.modelfile: %w", err)
	}
	return nil
}

// SetOllamaModel updates the active Ollama model.
func (a *AIOps) SetOllamaModel(modelName string) {
	defer common.RecoverPanic()
	common.LogInfo("AIOps: Switching active model to %q", modelName)
	aiops.SetModel(modelName)
}

// emitCreateProgress emits an ollama:progress event for model creation/pull.
// It is safe to call from the create/pull progress callback (runs on a worker
// goroutine) because wailsruntime.EventsEmit is thread-safe.
func (a *AIOps) emitCreateProgress(resp api.ProgressResponse) error {
	if a.ctx == nil {
		return nil
	}
	percent := 0.0
	if resp.Total > 0 {
		percent = float64(resp.Completed) / float64(resp.Total) * 100
	}
	wailsruntime.EventsEmit(a.ctx, "ollama:progress", OllamaProgress{
		Status:    resp.Status,
		Percent:   percent,
		Total:     resp.Total,
		Completed: resp.Completed,
	})
	return nil
}

// PullModel initiates a model pull and emits progress events to the frontend.
func (a *AIOps) PullModel(modelName string) error {
	defer common.RecoverPanic()
	common.LogInfo("AIOps: Pulling model %q", modelName)
	// Use background context so pull survives window lifecycle changes
	pullCtx := context.Background()
	err := aiops.PullModelWithContext(pullCtx, modelName, a.emitCreateProgress)

	if err != nil {
		common.LogWarn("AIOps: PullModel failed: %v", err)
		return err
	}
	return nil
}

// DeleteModel removes a local model from Ollama.
func (a *AIOps) DeleteModel(modelName string) error {
	defer common.RecoverPanic()
	common.LogInfo("AIOps: Deleting model %q", modelName)
	return aiops.DeleteModelWithContext(a.ctx, modelName)
}

// ── AI Model Constants ──────────────────────────────────────────────────────

const (
	// Qwythos-9B is the primary default — a custom GGUF from HuggingFace.
	// Based on Qwen3.5-9B with 1M context window, Q6_K quant = ~6.85 GiB.
	qwythosModel = "hf.co/empero-ai/Qwythos-9B-Claude-Mythos-5-1M-GGUF:Q6_K"
	qwythosLabel = "Qwythos-9B"

	// Qwen3 fallbacks use Ollama-native tags (no HF GGUF URL needed).
	// All verified via ollama.com/library/qwen3 (33M+ downloads).
	// See: https://ollama.com/library/qwen3
	qwen3_8B_Model = "qwen3:8b"
	qwen3_8B_Label = "Qwen3-8B"

	qwen3_4B_Model = "qwen3:4b"
	qwen3_4B_Label = "Qwen3-4B"

	// qwen3:1.7b is the nearest to the original "1.5B" target;
	// there is no qwen3:1.5b on Ollama.
	qwen3_1_7B_Model = "qwen3:1.7b"
	qwen3_1_7B_Label = "Qwen3-1.7B"

	qwen3_0_6B_Model = "qwen3:0.6b"
	qwen3_0_6B_Label = "Qwen3-0.6B"
)

// analystSystemPrompt is the shared system prompt for the UniversalOps
// System Analyst persona. Both SetupOllamaPersona and CreateOpsPersona
// reference this constant so the prompt text cannot drift between code paths.
const analystSystemPrompt = `You are the UniversalOps System Analyst, a high-density technical co-pilot.
Objective: Synthesize complex telemetry into factual technical briefings.

### Operational Protocol
1. Use your <thought> block to correlate telemetry history and identify root causes.
2. Provide a concise technical justification for any proposed action.
3. Emit action requests using: <action_request name="ACTION_NAME" param1="VALUE" />

### Available Actions
- kill_process (pid): Stop a specific process.
- restart_service (name): Restart a system service.
- disk_cleanup: Initiate temporary file removal.
- defrag: Optimize primary drive storage.

Anchor all findings in the provided System History and live Metrics.`

// qwen3FallbackForRAM returns the best QWEN 3.x alternative for the available RAM.
// Thresholds account for OS + app overhead leaving enough headroom for the model.
func qwen3FallbackForRAM(ramGB float64) (modelName, label string) {
	switch {
	case ramGB >= 8:
		return qwen3_4B_Model, qwen3_4B_Label + " (recommended for " + fmt.Sprintf("%.0fGB RAM", ramGB) + ")"
	case ramGB >= 4:
		return qwen3_1_7B_Model, qwen3_1_7B_Label + " (recommended for " + fmt.Sprintf("%.0fGB RAM", ramGB) + ")"
	default:
		return qwen3_0_6B_Model, qwen3_0_6B_Label + " (recommended for " + fmt.Sprintf("%.0fGB RAM", ramGB) + ")"
	}
}

// GetAISetupRecommendation evaluates the user's system (CPU cores, available RAM)
// and recommends the best AI model. Qwythos-9B is always preferred when the system
// can support it (≥12 GB RAM + ≥4 CPU cores). Otherwise a QWEN 3.x fallback is chosen
// based on available memory. The timestamp is recorded for present-day awareness.
func (a *AIOps) GetAISetupRecommendation() AISetupRecommendation {
	defer common.RecoverPanic()

	// 1. Snapshot system resources
	v, err := mem.VirtualMemory()
	totalMemGB := float64(0)
	if err == nil && v != nil {
		totalMemGB = float64(v.Total) / 1024 / 1024 / 1024
	}
	cpuThreads := goruntime.NumCPU()

	rec := AISetupRecommendation{
		SystemRAMGB:      totalMemGB,
		SystemCPUThreads: cpuThreads,
		Timestamp:        time.Now().Format("2006-01-02 15:04:05 MST"),
	}

	// 2. Can the system host Qwythos-9B (~8 GB weights + ~2 GB context + OS overhead)?
	canRunQwythos := totalMemGB >= 12 && cpuThreads >= 4
	rec.CanRunQwythos = canRunQwythos

	// 3. Discover what models Ollama already has
	ollamaModels := []string{}
	if status, err := aiops.CheckOllamaWithContext(a.ctx); err == nil && status.Available {
		ollamaModels = status.AvailableModels
	}

	// 4. Check if the recommendation already exists locally
	modelExists := func(name string) bool {
		for _, m := range ollamaModels {
			if m == name || strings.HasPrefix(m, name+":") {
				return true
			}
		}
		return false
	}

	// 5. Build recommendation
	if canRunQwythos {
		rec.RecommendedModel = qwythosModel
		rec.RecommendedLabel = qwythosLabel
		rec.QwythosExists = modelExists(qwythosModel)
		rec.PullRequired = !rec.QwythosExists
	} else {
		rec.RecommendedModel, rec.RecommendedLabel = qwen3FallbackForRAM(totalMemGB)
		rec.PullRequired = !modelExists(rec.RecommendedModel)
	}

	// 6. Always populate the full QWEN 3.x fallback list (Ollama-native tags)
	rec.FallbackModels = []ModelOption{
		{Name: qwen3_8B_Model, Label: qwen3_8B_Label, SizeGB: 5.2},
		{Name: qwen3_4B_Model, Label: qwen3_4B_Label, SizeGB: 2.5},
		{Name: qwen3_1_7B_Model, Label: qwen3_1_7B_Label, SizeGB: 1.4},
		{Name: qwen3_0_6B_Model, Label: qwen3_0_6B_Label, SizeGB: 0.5},
	}

	common.LogInfo("AIOps: SetupRecommendation → canRunQwythos=%t, pullRequired=%t, model=%s",
		rec.CanRunQwythos, rec.PullRequired, rec.RecommendedModel)
	return rec
}

// SetupOllamaPersona drives the full AI setup flow for the onboarding wizard.
// It accepts a base model (from GetAISetupRecommendation), pulls it if missing,
// creates the "universalops" persona in Ollama, persists the model setting,
// and runs a quick connection test.
func (a *AIOps) SetupOllamaPersona(baseModel string) error {
	defer common.RecoverPanic()
	common.LogInfo("AIOps: SetupOllamaPersona starting (base: %s)", baseModel)

	// 1. Single CheckOllamaWithContext call — checks for universalops persona + base model
	common.LogInfo("AIOps: Checking Ollama status for persona setup (base: %s)", baseModel)
	universalopsExists := false
	baseModelFound := false

	if status, err := aiops.CheckOllamaWithContext(a.ctx); err == nil && status.Available {
		for _, m := range status.AvailableModels {
			switch {
			case m == "universalops" || strings.HasPrefix(m, "universalops:"):
				universalopsExists = true
			case m == baseModel || strings.HasPrefix(m, baseModel+":"):
				baseModelFound = true
			}
		}
	}

	// If universalops already exists, activate and test
	if universalopsExists {
		common.LogInfo("AIOps: universalops already exists — activating")
		aiops.SetModel("universalops")
		a.SetModelForCapability("analyst", "universalops")
		return a.testPersonaConnection()
	}

	// 2. Ensure base model is pulled locally
	if !baseModelFound {
		common.LogInfo("AIOps: Base model %s not found — pulling from registry", baseModel)
		if err := a.PullModel(baseModel); err != nil {
			return fmt.Errorf("failed to pull base model %s: %w", baseModel, err)
		}
	}

	// 3. Write a correct modelfile to data/ for transparency and later editing
	modelfilePath := filepath.Join(a.dataDir, "universalops.modelfile")
	modelfileContent := fmt.Sprintf(`FROM %s

# System message for UniversalOps System Analyst
SYSTEM """
%s
"""

# Parameters for technical precision and consistency
PARAMETER temperature 0.1
PARAMETER repeat_penalty 1.1
PARAMETER top_k 40
PARAMETER num_ctx 32768
PARAMETER stop "</action_request>"
PARAMETER stop "──"
`, baseModel, analystSystemPrompt)
	if err := os.WriteFile(modelfilePath, []byte(modelfileContent), 0644); err != nil {
		common.LogWarn("AIOps: failed to write modelfile: %v", err)
	}

	// 4. Create the universalops persona via the Ollama SDK
	common.LogInfo("AIOps: Creating universalops model from %s", baseModel)

	params := map[string]any{
		"temperature": 0.1,
		"num_ctx":     32768,
		"stop":        []string{"</action_request>", "──"},
	}

	// Use a lifecycle-independent context (matching PullModel) so a window/app
	// shutdown mid-create cannot abort a multi-GiB base-model download. The
	// operation is still bounded by the long-running client's 45-minute timeout.
	if err := aiops.CreateModelWithContext(context.Background(), "universalops", baseModel, analystSystemPrompt, params, a.emitCreateProgress); err != nil {
		return fmt.Errorf("persona creation failed for base %s: %w", baseModel, err)
	}

	// 5. Persist the model setting so future sessions remember it
	a.SetModelForCapability("analyst", "universalops")

	// 6. Activate
	aiops.SetModel("universalops")

	// 7. Quick sanity check — can the model respond?
	if err := a.testPersonaConnection(); err != nil {
		// Model was created but inference is broken — surface the error with context
		return fmt.Errorf("model created but connection verification failed: %w", err)
	}

	common.LogInfo("AIOps: SetupOllamaPersona completed successfully (base: %s)", baseModel)
	return nil
}

// testPersonaConnection sends a minimal chat to verify the active model can respond.
func (a *AIOps) testPersonaConnection() error {
	resp, err := aiops.ChatWithModelAndContext(a.ctx, []aiops.ChatMessage{
		{Role: "user", Content: "Respond with exactly: OK"},
	}, "universalops")
	if err != nil {
		return fmt.Errorf("persona connection test failed — Ollama may be overloaded or the model incompatible: %w", err)
	}
	if resp == "" {
		return fmt.Errorf("persona returned empty response — the model may not be fully loaded")
	}
	common.LogInfo("AIOps: Persona connection test passed (response: %.60s)", resp)
	return nil
}

// CreateOpsPersona creates or activates the specialized 'universalops' model from the local modelfile.
// If the model already exists in Ollama, it simply activates it without re-creating.
func (a *AIOps) CreateOpsPersona() error {
	defer common.RecoverPanic()
	common.LogInfo("AIOps: Creating/activating universalops persona")

	// 0. Check if the model already exists in Ollama — just activate it
	if status, err := aiops.CheckOllamaWithContext(a.ctx); err == nil && status.Available {
		for _, m := range status.AvailableModels {
			if strings.HasPrefix(m, "universalops") || strings.HasPrefix(m, "universalops:") {
				common.LogInfo("AIOps: universalops model already exists (%s). Activating it.", m)
				aiops.SetModel(m)
				return nil
			}
		}
	}

	modelfilePath := filepath.Join(a.dataDir, "universalops.modelfile")

	// 1. Check for legacy root Modelfile and migrate if necessary
	legacyPaths := []string{"Modelfile", filepath.Join("data", "Modelfile")}
	for _, lp := range legacyPaths {
		if _, err := os.Stat(lp); err == nil {
			if _, errData := os.Stat(modelfilePath); os.IsNotExist(errData) {
				common.LogInfo("AIOps: Migrating legacy %s to %s", lp, modelfilePath)
				_ = os.Rename(lp, modelfilePath)
			}
		}
	}

	// 2. Ensure universalops.modelfile exists in data/
	if _, err := os.Stat(modelfilePath); os.IsNotExist(err) {
		common.LogInfo("AIOps: universalops.modelfile missing in data/. Creating default (Qwythos-9B).")
		content := fmt.Sprintf(`FROM %s

# System message for UniversalOps System Analyst
SYSTEM """
%s
"""

# Parameters for technical precision and consistency
PARAMETER temperature 0.1
PARAMETER repeat_penalty 1.1
PARAMETER top_k 40
PARAMETER num_ctx 32768
PARAMETER stop "</action_request>"
PARAMETER stop "──"
`, qwythosModel, analystSystemPrompt)
		if err := os.WriteFile(modelfilePath, []byte(content), 0644); err != nil {
			common.LogWarn("AIOps: Failed to create modelfile: %v", err)
		}
	}

	// 3. Parse Modelfile
	config, err := aiops.ParseModelfile(modelfilePath)
	if err != nil {
		common.LogWarn("AIOps: Failed to parse Modelfile: %v. Using hardcoded defaults (Qwythos-9B).", err)
		config = &aiops.ModelfileConfig{
			From:   qwythosModel,
			System: analystSystemPrompt,
			Parameters: map[string]any{
				"temperature": 0.1,
				"num_ctx":     32768,
				"top_p":       0.95,
				"stop":        []string{"──", "</action_request>"},
			},
		}
	} else {
		common.LogInfo("AIOps: Successfully parsed Modelfile from %s (Base: %s)", modelfilePath, config.From)
	}

	// 4. Create Model via SDK
	common.LogInfo("AIOps: Requesting model creation for 'universalops' (Base: %s)", config.From)

	// Note: Ollama's Create API handles pulling the base model automatically
	// in recent versions if it's missing from the library.
	// Use a lifecycle-independent context (matching PullModel) so a window/app
	// shutdown mid-create cannot abort a multi-GiB base-model download. The
	// operation is still bounded by the long-running client's 45-minute timeout.
	err = aiops.CreateModelWithContext(context.Background(), "universalops", config.From, config.System, config.Parameters, a.emitCreateProgress)
	if err != nil {
		// If creation fails because base is missing, attempt an explicit pull
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "pull") {
			common.LogInfo("AIOps: Base model missing. Attempting explicit pull of %s", config.From)
			if pErr := a.PullModel(config.From); pErr != nil {
				// The pull is a hard prerequisite for creation — surface its
				// error directly rather than retrying a create that will fail
				// again against a still-missing base model.
				return fmt.Errorf("failed to pull base model %s: %w", config.From, pErr)
			}
			// Retry creation after successful pull
			err = aiops.CreateModelWithContext(context.Background(), "universalops", config.From, config.System, config.Parameters, a.emitCreateProgress)
		}
	}

	if err != nil {
		common.LogWarn("AIOps: CreateModel failed: %v", err)
		return fmt.Errorf("persona creation failed after pulling base model: %w", err)
	}

	// 5. Activate the newly-created model
	aiops.SetModel("universalops")
	common.LogInfo("AIOps: Successfully created and activated universalops persona")
	return nil
}

// DetectAnomalies performs anomaly detection on pipeline metrics.
func (a *AIOps) DetectAnomalies() []AnomalyInfo {
	defer common.RecoverPanic()
	anoms := aiops.DetectPipelineAnomalies(a.pipeline)
	var out []AnomalyInfo
	for _, anom := range anoms {
		out = append(out, AnomalyInfo{
			Metric:    anom.Metric,
			Value:     anom.Value,
			Expected:  anom.Expected,
			Deviation: anom.Deviation,
			Severity:  anom.Severity,
			Timestamp: anom.Timestamp,
		})
	}
	return out
}

// ── AI Methods for Timeline Integration ──────────────────────────────────────

// AskAI sends a prompt to the AI with the given context and returns the response.
func (a *AIOps) AskAI(ctx context.Context, prompt string) (string, error) {
	defer common.RecoverPanic()
	messages := []aiops.ChatMessage{
		{Role: "system", Content: "You are the UniversalOps AI assistant, an expert operations analyst. Be concise and specific."},
		{Role: "user", Content: prompt},
	}
	return aiops.ChatWithContext(ctx, messages)
}

// QuerySystemState answers a natural-language system-state question using live metrics.
func (a *AIOps) QuerySystemState(query string) string {
	defer common.RecoverPanic()
	stats, _ := a.sysOps.collector.CollectAllStats()

	return aiops.AnswerSystemStateQuery(query, stats, nil, nil)
}

// GetAIInsights synthesizes anomaly, trend, and alert data into actionable insights.
func (a *AIOps) GetAIInsights() []AIInsight {
	defer common.RecoverPanic()
	var insights []AIInsight
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")

	// 1. Pull anomalies from existing detection
	anomalies := a.DetectAnomalies()
	for _, anom := range anomalies {
		insights = append(insights, AIInsight{
			Category:   a.metricCategory(anom.Metric),
			Severity:   anom.Severity,
			Title:      fmt.Sprintf("%s anomaly detected", titleCaser.String(anom.Metric)),
			Message:    fmt.Sprintf("%s is at %.1f (expected ~%.1f, %.1fσ deviation)", anom.Metric, anom.Value, anom.Expected, anom.Deviation),
			Action:     fmt.Sprintf("Investigate %s usage for runaway processes or resource leaks.", anom.Metric),
			ActionPage: "aiops",
			Timestamp:  now,
		})
	}

	// 2. Check metric trends
	metricNames := []string{common.MetricCPU, common.MetricMem, common.MetricDisk}
	for _, name := range metricNames {
		mf := a.pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 10 {
			continue
		}
		if mf.Trend.Direction == common.TrendRising && mf.LastValue > 75 {
			insights = append(insights, AIInsight{
				Category:  a.metricCategory(name),
				Severity:  "warning",
				Title:     fmt.Sprintf("%s trending upward", strings.ToUpper(name)),
				Message:   fmt.Sprintf("%s is at %.1f%% and rising. May reach critical threshold.", name, mf.LastValue),
				Action:    "Monitor closely and consider preemptive action.",
				Timestamp: now,
			})
		} else if mf.Trend.Direction == common.TrendFalling && mf.LastValue > 90 {
			insights = append(insights, AIInsight{
				Category:  a.metricCategory(name),
				Severity:  "info",
				Title:     fmt.Sprintf("%s recovering", strings.ToUpper(name)),
				Message:   fmt.Sprintf("%s was high but is now trending down from %.1f%%.", name, mf.LastValue),
				Action:    "No action needed — conditions improving.",
				Timestamp: now,
			})
		}
	}

	// 3. Check recent alerts
	storage := common.GetStorage()
	if storage != nil {
		history, _ := storage.QueryAlertHistory(20)
		var activeCritical, activeWarning int
		for _, alert := range history {
			if alert.Resolved {
				continue
			}
			if alert.Level == "CRITICAL" {
				activeCritical++
			} else if alert.Level == "WARNING" {
				activeWarning++
			}
		}
		if activeCritical > 0 {
			insights = append(insights, AIInsight{
				Category:   "alerts",
				Severity:   "critical",
				Title:      fmt.Sprintf("%d active critical alert(s)", activeCritical),
				Message:    "There are unresolved critical alerts that require immediate attention.",
				Action:     "Review and resolve critical alerts in the Alerts dashboard.",
				ActionPage: "alerts",
				Timestamp:  now,
			})
		}
		if activeWarning > 0 {
			insights = append(insights, AIInsight{
				Category:   "alerts",
				Severity:   "warning",
				Title:      fmt.Sprintf("%d active warning alert(s)", activeWarning),
				Message:    "There are unresolved warnings that should be reviewed.",
				Action:     "Check warning-level alerts for emerging issues.",
				ActionPage: "alerts",
				Timestamp:  now,
			})
		}
	}

	if len(insights) == 0 {
		insights = append(insights, AIInsight{
			Category:  "general",
			Severity:  "info",
			Title:     "System operating normally",
			Message:   "No anomalies, concerning trends, or active alerts detected.",
			Action:    "Continue monitoring.",
			Timestamp: now,
		})
	}

	// Sort by severity: critical > warning > info
	severityRank := map[string]int{"critical": 0, "warning": 1, "info": 2}
	sort.Slice(insights, func(i, j int) bool {
		return severityRank[insights[i].Severity] < severityRank[insights[j].Severity]
	})

	return insights
}

// GetConfidenceScore computes an overall system confidence score (0-100).
func (a *AIOps) GetConfidenceScore() AIConfidence {
	defer common.RecoverPanic()
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	factors := make(map[string]float64)

	// Factor 1: Data freshness (30%) — how recent is the latest metric?
	storage := common.GetStorage()
	freshnessScore := 50.0 // default if no storage
	if storage != nil {
		recent, _ := storage.GetMetricHistory(common.MetricCPU, 1)
		if len(recent) > 0 {
			freshnessScore = 100.0 // we have recent data
		}
	}
	factors["data_freshness"] = freshnessScore

	// Factor 2: Metric stability (25%) — low variance across metrics
	stabilityScore := 100.0
	metricNames := []string{common.MetricCPU, common.MetricMem, common.MetricDisk}
	for _, name := range metricNames {
		mf := a.pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 5 {
			continue
		}
		stddev := (mf.Stats.Max - mf.Stats.Min) / 2
		if stddev > 20 {
			stabilityScore -= 15
		} else if stddev > 10 {
			stabilityScore -= 5
		}
		// Penalize if value is near critical threshold
		if mf.LastValue > 90 {
			stabilityScore -= 10
		} else if mf.LastValue > 80 {
			stabilityScore -= 3
		}
	}
	if stabilityScore < 0 {
		stabilityScore = 0
	}
	factors["metric_stability"] = stabilityScore

	// Factor 3: Anomaly count (25%)
	anomalies := a.DetectAnomalies()
	anomalyScore := 100.0
	for _, anom := range anomalies {
		if anom.Severity == "critical" {
			anomalyScore -= 25
		} else {
			anomalyScore -= 10
		}
	}
	if anomalyScore < 0 {
		anomalyScore = 0
	}
	factors["anomaly_count"] = anomalyScore

	// Factor 4: Alert history (20%)
	alertScore := 100.0
	if storage != nil {
		history, _ := storage.QueryAlertHistory(20)
		for _, alert := range history {
			if alert.Resolved {
				continue
			}
			if alert.Level == "CRITICAL" {
				alertScore -= 15
			} else if alert.Level == "WARNING" {
				alertScore -= 5
			}
		}
	}
	if alertScore < 0 {
		alertScore = 0
	}
	factors["alert_health"] = alertScore

	overall := (freshnessScore*0.30 + stabilityScore*0.25 + anomalyScore*0.25 + alertScore*0.20)

	return AIConfidence{
		Overall:   math.Round(overall*10) / 10,
		Factors:   factors,
		UpdatedAt: now,
	}
}

// ── Conversation Persistence ───────────────────────────────────────────────

// LearnedBaseline represents the statistical baseline for a metric.
type LearnedBaseline struct {
	Metric string  `json:"metric"`
	Mean   float64 `json:"mean"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	StdDev float64 `json:"stdDev"`
	Count  int     `json:"count"` // number of samples
}

// GetLearnedBaselines returns the statistical baseline for each core metric.
func (a *AIOps) GetLearnedBaselines() []LearnedBaseline {
	defer common.RecoverPanic()
	metricNames := []string{common.MetricCPU, common.MetricMem, common.MetricDisk, common.MetricNetRX, common.MetricNetTX}
	var baselines []LearnedBaseline

	for _, name := range metricNames {
		mf := a.pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 5 {
			continue
		}
		stddev := (mf.Stats.Max - mf.Stats.Min) / 2
		baselines = append(baselines, LearnedBaseline{
			Metric: name,
			Mean:   math.Round(mf.Stats.Avg*10) / 10,
			Min:    math.Round(mf.Stats.Min*10) / 10,
			Max:    math.Round(mf.Stats.Max*10) / 10,
			StdDev: math.Round(stddev*10) / 10,
			Count:  len(mf.Values),
		})
	}

	if baselines == nil {
		baselines = []LearnedBaseline{}
	}
	return baselines
}

// SaveMessage persists a chat message. If sessionID is empty, generates a new one.
func (a *AIOps) SaveMessage(sessionID, role, content string) string {
	defer common.RecoverPanic()
	storage := common.GetStorage()
	if storage == nil {
		return sessionID
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess-%d", time.Now().UnixMilli())
	}
	if err := storage.InsertMessage(sessionID, role, content); err != nil {
		common.LogWarn("SaveMessage: %v", err)
	}
	return sessionID
}

// GetMessages returns all messages for a given session.
func (a *AIOps) GetMessages(sessionID string) []ConversationMessage {
	defer common.RecoverPanic()
	storage := common.GetStorage()
	if storage == nil {
		return []ConversationMessage{}
	}
	msgs, err := storage.GetMessages(sessionID)
	if err != nil {
		return []ConversationMessage{}
	}
	result := make([]ConversationMessage, len(msgs))
	for i, m := range msgs {
		result[i] = ConversationMessage{
			ID:        m.ID,
			SessionID: m.SessionID,
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp.Format(time.RFC3339),
		}
	}
	return result
}

// ListSessions returns all chat sessions with metadata.
func (a *AIOps) ListSessions() []map[string]interface{} {
	defer common.RecoverPanic()
	storage := common.GetStorage()
	if storage == nil {
		return []map[string]interface{}{}
	}
	sessions, err := storage.ListSessions()
	if err != nil {
		return []map[string]interface{}{}
	}
	return sessions
}

// DeleteSession removes a chat session and all its messages.
func (a *AIOps) DeleteSession(sessionID string) {
	defer common.RecoverPanic()
	storage := common.GetStorage()
	if storage == nil {
		return
	}
	if err := storage.DeleteSession(sessionID); err != nil {
		common.LogWarn("DeleteSession: %v", err)
	}
}

// RequestOptimization asks Hawk to analyze system load and settings to propose improvements.
func (a *AIOps) RequestOptimization() ChatResponse {
	defer common.RecoverPanic()
	// 1. Gather Context
	knowledge := a.knowledge.GetSnapshot()
	settings := a.pipelineAPI.GetCurrentSettings()

	prompt := fmt.Sprintf(`Analyze the current workstation load and engine settings.
Metrics: CPU (system.cpu.utilization)=%.1f%%, RAM (system.memory.usage)=%.1f%%, Disk (system.disk.usage)=%.1f%%, Anomalies=%d.
Current Settings: refreshInterval=%vms, pingCount=%v, dnsTimeout=%vms.

Propose optimizations to balance telemetry density and system performance.
Your response MUST be a valid JSON object with this exact structure:
{
  "content": "Brief reasoning in your industrial co-pilot tone.",
  "payload": {
    "refreshInterval": 1000,
    "pingCount": 4,
    "dnsTimeout": 2000
  }
}
If no changes are needed, return the current settings in the payload.`,
		knowledge.SystemCPUUtilization, knowledge.SystemMemoryUsage, knowledge.SystemDiskUsage, knowledge.Anomalies,
		settings["refreshInterval"], settings["pingCount"], settings["dnsTimeout"])

	// 2. Call AI with Optimizer model override (with per-request timeout)
	optParent := a.requestContext()
	optCtx, cancel := context.WithTimeout(optParent, 120*time.Second)
	defer cancel()

	optStart := time.Now()
	resp, err := aiops.ChatWithModelAndContext(optCtx, []aiops.ChatMessage{
		{Role: "system", Content: "You are the UniversalOps Engine Optimizer. Output JSON only."},
		{Role: "user", Content: prompt},
	}, a.OptimizerModel)
	optElapsed := time.Since(optStart)

	if err != nil {
		common.LogWarn("RequestOptimization failed after %v: %v", optElapsed, err)
		return ChatResponse{Content: "Optimization analysis unavailable. Please check that Ollama is running."}
	}

	common.LogInfo("RequestOptimization completed in %v", optElapsed)

	// 3. Robust JSON Extraction
	// (AI might wrap JSON in markdown blocks or output Thought blocks first)
	jsonStr := ""
	startIdx := strings.Index(resp, "{")
	endIdx := strings.LastIndex(resp, "}")

	if startIdx >= 0 && endIdx > startIdx {
		jsonStr = resp[startIdx : endIdx+1]
	} else {
		common.LogWarn("AIOps: Could not find JSON block in AI response. Raw: %q", resp)
		return ChatResponse{Content: "Hawk generated an invalid response format. Please try again."}
	}

	var result ChatResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		common.LogWarn("AIOps: Failed to parse optimization JSON: %v. Raw: %q", err, jsonStr)
		return ChatResponse{Content: "Hawk generated an invalid optimization payload. Please try again."}
	}

	return result
}

// NotifyActionResult informs the AI of an action outcome and returns its response.
// handshakeID is used for DORA compliance logging on rejected actions.
func (a *AIOps) NotifyActionResult(sessionID, actionName, status, details, handshakeID string) ChatResponse {
	defer common.RecoverPanic()
	// DORA compliance: Log rejected (aborted) actions to decisions_audit
	if status == "ABORTED" && handshakeID != "" {
		if reg := common.GetHandshakeRegistry(); reg != nil {
			if pending, ok := reg.Peek(handshakeID); ok {
				if storage := common.GetStorage(); storage != nil {
					argsJSON, _ := json.Marshal(pending.Params)
					ctxSnap := ""
					if a.knowledge != nil {
						snap, _ := json.Marshal(a.knowledge.GetSnapshot())
						ctxSnap = string(snap)
					}
					_ = storage.LogDecision(sessionID, actionName, string(argsJSON), ctxSnap, false)
				}
			}
		}
	}

	message := fmt.Sprintf("SYSTEM_NOTIFICATION: Action %q finished with status %q. Details: %s", actionName, status, details)
	a.SaveMessage(sessionID, "user", message)
	resp := a.Chat(sessionID, message)
	if resp.Content != "" {
		a.SaveMessage(sessionID, "assistant", resp.Content)
	}
	return resp
}

// VerifyRemediation asks Hawk to verify if a workflow improved system metrics.
func (a *AIOps) VerifyRemediation(sessionID, workflowID string, beforeMetrics map[string]float64) ChatResponse {
	defer common.RecoverPanic()
	knowledge := a.knowledge.GetSnapshot()

	prompt := fmt.Sprintf(`Verify the outcome of workflow '%s'.
Baseline before execution: %v
Current metrics: CPU (system.cpu.utilization)=%.1f%%, RAM (system.memory.usage)=%.1f%%, Disk (system.disk.usage)=%.1f%%.

Determine if the remediation was successful and provide a brief technical assessment.`,
		workflowID, beforeMetrics, knowledge.SystemCPUUtilization, knowledge.SystemMemoryUsage, knowledge.SystemDiskUsage)

	a.SaveMessage(sessionID, "user", prompt)
	resp := a.Chat(sessionID, prompt)
	if resp.Content != "" {
		a.SaveMessage(sessionID, "assistant", resp.Content)
	}
	return resp
}

// PlanWorkflow asks Hawk to design a custom operational workflow.
func (a *AIOps) PlanWorkflow(sessionID, objective string) ChatResponse {
	defer common.RecoverPanic()
	prompt := fmt.Sprintf(`Plan a multi-step operational workflow to achieve this objective: "%s".
Your response MUST be a valid JSON object matching this structure:
{
  "content": "Reasoning for the plan.",
  "payload": {
    "id": "wf-slug",
    "name": "Human Name",
    "description": "Short description",
    "why": "Context",
    "steps": [
      {"id": "step-1", "label": "Label", "description": "Action details", "command": "Optional PowerShell command"}
    ]
  }
}
Keep it technical and focused on system/network/security ops.`, objective)

	a.SaveMessage(sessionID, "user", prompt)

	// Use OptimizerModel for structured planning if available
	resp, err := aiops.ChatWithModelAndContext(a.ctx, []aiops.ChatMessage{
		{Role: "system", Content: "You are the UniversalOps Workflow Architect. Output JSON only."},
		{Role: "user", Content: prompt},
	}, a.OptimizerModel)

	if err != nil {
		common.LogWarn("RequestWorkflowPlan failed: %v", err)
		return ChatResponse{Content: "Workflow planning unavailable. Please check that Ollama is running."}
	}

	// Simple JSON Extraction
	jsonStr := resp
	if strings.Contains(resp, "```json") {
		parts := strings.Split(resp, "```json")
		if len(parts) > 1 {
			jsonStr = strings.Split(parts[1], "```")[0]
		}
	}

	var result ChatResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		common.LogWarn("AIOps: Failed to parse workflow JSON: %v", err)
		return ChatResponse{Content: "Hawk generated an invalid workflow plan. Please try again."}
	}

	if result.Content != "" {
		a.SaveMessage(sessionID, "assistant", result.Content)
	}
	return result
}

// ExportReportAsPDF converts a markdown report to a formatted file.
func (a *AIOps) ExportReportAsPDF(title, markdown string) (string, error) {
	defer common.RecoverPanic()
	// For portability without heavy dependencies (like wkhtmltopdf/gotenberg),
	// we will generate a nicely formatted .txt/md "Professional Report"
	// until a light PDF lib is integrated.

	filename := fmt.Sprintf("Report_%d.md", time.Now().Unix())
	path := filepath.Join(a.dataDir, "reports", filename)

	_ = os.MkdirAll(filepath.Join(a.dataDir, "reports"), 0755)

	report := fmt.Sprintf("# %s\n\n_Generated by %s AI Analyst_\n_Date: %s_\n\n%s",
		title, a.GetCompanionName(), time.Now().Format(time.RFC1123), markdown)

	err := os.WriteFile(path, []byte(report), 0644)
	return path, err
}

// summarizeMetrics produces a technical string summary of a value slice.
func (a *AIOps) summarizeMetrics(name string, values []float64) string {
	if len(values) == 0 {
		return "No data"
	}

	var sum, min, max float64
	min = values[0]
	max = values[0]

	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	avg := sum / float64(len(values))

	// Calculate StdDev
	var sumSqDiff float64
	for _, v := range values {
		diff := v - avg
		sumSqDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSqDiff / float64(len(values)))

	return fmt.Sprintf("Avg=%.1f%%, Min=%.1f%%, Max=%.1f%%, StdDev=%.1f (n=%d)",
		avg, min, max, stdDev, len(values))
}

// metricCategory maps a metric name to an insight category.
func (a *AIOps) metricCategory(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "cpu"):
		return "performance"
	case strings.Contains(name, "mem"):
		return "performance"
	case strings.Contains(name, "disk"):
		return "storage"
	case strings.Contains(name, "net"):
		return "network"
	default:
		return "general"
	}
}
