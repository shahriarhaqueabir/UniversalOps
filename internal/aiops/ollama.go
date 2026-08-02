package aiops

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
)

// ChatMessage represents a message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // message content
}

// OllamaStatus represents the availability and configuration of Ollama.
type OllamaStatus struct {
	Available       bool
	Model           string
	Version         string
	AvailableModels []string
}

const (
	defaultOllamaURL   = "http://127.0.0.1:11434"
	defaultOllamaModel = "universalops"
	defaultHttpTimeout = 120 * time.Second

	// longRunningHttpTimeout is used for model create/pull operations, which
	// can take many minutes when Ollama must first download a large base model
	// (e.g. Qwythos-9B ~6.85 GiB). The 120s chat timeout is far too short for
	// these streaming requests and caused "context deadline exceeded" failures
	// during persona deployment.
	//
	// A generous-but-finite bound is used (rather than 0 = no timeout) so that
	// a stalled download or unresponsive daemon cannot hang the Wails binding
	// call forever. 45 minutes comfortably covers even a slow multi-GiB pull
	// while still guaranteeing the operation eventually returns.
	longRunningHttpTimeout = 45 * time.Minute
)

func getOllamaURL() string {
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		return host
	}
	return defaultOllamaURL
}

func getOllamaModel() string {
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		return model
	}
	return defaultOllamaModel
}

func newClient() *api.Client {
	return newClientWithTimeout(defaultHttpTimeout)
}

// newLongRunningClient returns an Ollama client with a generous (45-minute)
// timeout, intended for long-running operations such as model create/pull that
// may need to download large base models. The operation is still bounded by the
// caller's context deadline, whichever comes first.
func newLongRunningClient() *api.Client {
	return newClientWithTimeout(longRunningHttpTimeout)
}

func newClientWithTimeout(timeout time.Duration) *api.Client {
	baseURL := getOllamaURL()
	u, err := url.Parse(baseURL)
	if err != nil {
		u = &url.URL{Scheme: "http", Host: "127.0.0.1:11434"}
	}
	return api.NewClient(u, &http.Client{Timeout: timeout})
}

// ── Global Compatibility Wrappers ───────────────────────────────────────────

var (
	defaultClientMu sync.Mutex
	defaultClient   *OllamaClient
)

func getDefaultClient() *OllamaClient {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()

	if defaultClient == nil {
		defaultClient = NewOllamaClient()
	}
	return defaultClient
}

// ResetDefaultClient forces the global client to be re-initialized from environment.
// Primarily used for testing.
func ResetDefaultClient() {
	defaultClientMu.Lock()
	defer defaultClientMu.Unlock()
	defaultClient = nil
}

// Chat sends a chat request using the default client and context.Background().
func Chat(messages []ChatMessage) (string, error) {
	return ChatWithContext(context.Background(), messages)
}

// ChatWithContext sends a chat request using the default client and the provided context.
func ChatWithContext(ctx context.Context, messages []ChatMessage) (string, error) {
	return getDefaultClient().Chat(ctx, messages, "")
}

// ChatWithModel sends a chat request using a specific model override.
func ChatWithModel(messages []ChatMessage, modelOverride string) (string, error) {
	return ChatWithModelAndContext(context.Background(), messages, modelOverride)
}

// ChatWithModelAndContext sends a chat request using a specific model override and context.
func ChatWithModelAndContext(ctx context.Context, messages []ChatMessage, modelOverride string) (string, error) {
	return getDefaultClient().Chat(ctx, messages, modelOverride)
}

// ChatStreamWithModelAndContext sends a streaming chat request using a specific model
// override and context. Each token is delivered via onToken.
func ChatStreamWithModelAndContext(ctx context.Context, messages []ChatMessage, modelOverride string, onToken func(string)) (string, error) {
	return getDefaultClient().ChatStream(ctx, messages, modelOverride, onToken)
}

// SetModel updates the model for the default client.
func SetModel(model string) {
	getDefaultClient().SetModel(model)
}

// CheckOllama checks the status using the default client.
func CheckOllama() (*OllamaStatus, error) {
	return CheckOllamaWithContext(context.Background())
}

// CheckOllamaWithContext checks the status using the default client and context.
func CheckOllamaWithContext(ctx context.Context) (*OllamaStatus, error) {
	return getDefaultClient().CheckStatus(ctx)
}

// PullModel downloads a model using the default client.
func PullModel(name string, onProgress func(api.ProgressResponse) error) error {
	return PullModelWithContext(context.Background(), name, onProgress)
}

// PullModelWithContext downloads a model using the default client and context.
func PullModelWithContext(ctx context.Context, name string, onProgress func(api.ProgressResponse) error) error {
	return getDefaultClient().PullModel(ctx, name, onProgress)
}

// CreateModel creates a model using the default client.
func CreateModel(name, from, system string, parameters map[string]any) error {
	return CreateModelWithContext(context.Background(), name, from, system, parameters, nil)
}

// CreateModelWithContext creates a model using the default client and context.
// onProgress, if non-nil, is invoked with each progress update from Ollama
// (useful for surfacing base-model pull progress during persona creation).
func CreateModelWithContext(ctx context.Context, name, from, system string, parameters map[string]any, onProgress func(api.ProgressResponse) error) error {
	return getDefaultClient().CreateModel(ctx, name, from, system, parameters, onProgress)
}

// DeleteModel removes a model using the default client.
func DeleteModel(name string) error {
	return DeleteModelWithContext(context.Background(), name)
}

// DeleteModelWithContext removes a model using the default client and context.
func DeleteModelWithContext(ctx context.Context, name string) error {
	return getDefaultClient().DeleteModel(ctx, name)
}

// ── OllamaClient Implementation ─────────────────────────────────────────────

// OllamaClient wraps the official Ollama API client.
type OllamaClient struct {
	api            *api.Client
	modelMu        sync.RWMutex
	effectiveModel string
}

func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		api:            newClient(),
		effectiveModel: getOllamaModel(),
	}
}

func (c *OllamaClient) SetModel(model string) {
	c.modelMu.Lock()
	defer c.modelMu.Unlock()
	c.effectiveModel = model
}

func (c *OllamaClient) GetModel() string {
	c.modelMu.RLock()
	defer c.modelMu.RUnlock()
	return c.effectiveModel
}

// CheckOllamaBinary returns true if the ollama binary is found in the system PATH.
func CheckOllamaBinary() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

// PullModel downloads a model from the Ollama library.
// Uses a long-running client because pulling a large base model can take
// minutes and must not be cut off by the 120s chat timeout.
func (c *OllamaClient) PullModel(ctx context.Context, name string, onProgress func(api.ProgressResponse) error) error {
	req := &api.PullRequest{Model: name}
	return newLongRunningClient().Pull(ctx, req, onProgress)
}

// DeleteModel removes a local model.
func (c *OllamaClient) DeleteModel(ctx context.Context, name string) error {
	req := &api.DeleteRequest{Model: name}
	return c.api.Delete(ctx, req)
}

// CreateModel creates a new model from structured parameters.
// Uses a long-running client because Ollama may need to pull the base model
// first (e.g. Qwythos-9B ~6.85 GiB), which exceeds the 120s chat timeout.
// onProgress, when non-nil, receives each progress update from the create
// stream so the frontend can show live download/build progress.
func (c *OllamaClient) CreateModel(ctx context.Context, name string, from string, system string, parameters map[string]any, onProgress func(api.ProgressResponse) error) error {
	req := &api.CreateRequest{
		Model:      name,
		From:       from,
		System:     system,
		Parameters: parameters,
	}

	if onProgress == nil {
		onProgress = func(resp api.ProgressResponse) error { return nil }
	}
	return newLongRunningClient().Create(ctx, req, onProgress)
}

// CheckStatus checks if the Ollama service is available.
func (c *OllamaClient) CheckStatus(ctx context.Context) (*OllamaStatus, error) {
	listResp, err := c.api.List(ctx)
	if err != nil {
		return &OllamaStatus{Available: false}, fmt.Errorf("ollama list failed: %w", err)
	}

	currentModel := c.GetModel()
	availableModels := make([]string, 0, len(listResp.Models))
	found := false
	for _, m := range listResp.Models {
		availableModels = append(availableModels, m.Name)
		if m.Name == currentModel || strings.HasPrefix(m.Name, currentModel+":") {
			found = true
			currentModel = m.Name
		}
	}

	// When the configured model isn't found in Ollama, do NOT silently
	// overwrite effectiveModel with the first available model (e.g., qwen2.5:7b).
	// Keeping the configured value lets the frontend detect the persona is
	// missing and show the "Initialize Persona" action button.
	if found {
		c.SetModel(currentModel)
	}

	version := "detected"
	if v, err := c.api.Version(ctx); err == nil && v != "" {
		version = v
	}

	return &OllamaStatus{
		Available:       true,
		Model:           currentModel,
		Version:         version,
		AvailableModels: availableModels,
	}, nil
}

// Chat sends a chat request to the Ollama API (non-streaming).
func (c *OllamaClient) Chat(ctx context.Context, messages []ChatMessage, modelOverride string) (string, error) {
	apiMessages := make([]api.Message, len(messages))
	for i, m := range messages {
		apiMessages[i] = api.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	model := c.GetModel()
	if modelOverride != "" {
		model = modelOverride
	}

	stream := false
	req := &api.ChatRequest{
		Model:    model,
		Messages: apiMessages,
		Stream:   &stream,
	}

	var responseText string
	err := c.api.Chat(ctx, req, func(resp api.ChatResponse) error {
		responseText += resp.Message.Content
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("ollama chat failed: %w", err)
	}

	if responseText == "" {
		return "", fmt.Errorf("ollama returned empty response")
	}

	return responseText, nil
}

// ChatStream sends a streaming chat request to the Ollama API. Each content
// token is delivered via the onToken callback. The final accumulated response
// is returned along with any error.
func (c *OllamaClient) ChatStream(ctx context.Context, messages []ChatMessage, modelOverride string, onToken func(string)) (string, error) {
	apiMessages := make([]api.Message, len(messages))
	for i, m := range messages {
		apiMessages[i] = api.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	model := c.GetModel()
	if modelOverride != "" {
		model = modelOverride
	}

	stream := true
	req := &api.ChatRequest{
		Model:    model,
		Messages: apiMessages,
		Stream:   &stream,
	}

	var responseText string
	err := c.api.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.Message.Content != "" {
			responseText += resp.Message.Content
			onToken(resp.Message.Content)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("ollama chat stream failed: %w", err)
	}

	if responseText == "" {
		return "", fmt.Errorf("ollama returned empty response")
	}

	return responseText, nil
}

// TruncateHistory ensures the chat history doesn't exceed a safe context window.
// Uses a simple word-count heuristic (approx 0.75 words per token).
// Limit is roughly 24k words for a 32k context.
func TruncateHistory(messages []ChatMessage, wordLimit int) []ChatMessage {
	if wordLimit <= 0 {
		wordLimit = 24000
	}

	totalWords := 0
	// Count from newest to oldest
	var keptMessages []ChatMessage
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		words := len(strings.Fields(msg.Content))
		if totalWords+words > wordLimit {
			break
		}
		totalWords += words
		keptMessages = append([]ChatMessage{msg}, keptMessages...)
	}

	return keptMessages
}
