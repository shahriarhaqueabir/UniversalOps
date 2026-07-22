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
	defaultOllamaModel = "opsforall"
	defaultHttpTimeout = 60 * time.Second
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
	baseURL := getOllamaURL()
	u, err := url.Parse(baseURL)
	if err != nil {
		u = &url.URL{Scheme: "http", Host: "127.0.0.1:11434"}
	}
	return api.NewClient(u, &http.Client{Timeout: defaultHttpTimeout})
}

// ── Global Compatibility Wrappers ───────────────────────────────────────────

var (
	defaultClientMu   sync.Mutex
	defaultClient     *OllamaClient
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
	return CreateModelWithContext(context.Background(), name, from, system, parameters)
}

// CreateModelWithContext creates a model using the default client and context.
func CreateModelWithContext(ctx context.Context, name, from, system string, parameters map[string]any) error {
	return getDefaultClient().CreateModel(ctx, name, from, system, parameters)
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
func (c *OllamaClient) PullModel(ctx context.Context, name string, onProgress func(api.ProgressResponse) error) error {
	req := &api.PullRequest{Model: name}
	return c.api.Pull(ctx, req, onProgress)
}

// DeleteModel removes a local model.
func (c *OllamaClient) DeleteModel(ctx context.Context, name string) error {
	req := &api.DeleteRequest{Model: name}
	return c.api.Delete(ctx, req)
}

// CreateModel creates a new model from structured parameters.
func (c *OllamaClient) CreateModel(ctx context.Context, name string, from string, system string, parameters map[string]any) error {
	req := &api.CreateRequest{
		Model:      name,
		From:       from,
		System:     system,
		Parameters: parameters,
	}

	return c.api.Create(ctx, req, func(resp api.ProgressResponse) error {
		return nil
	})
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

	if !found && len(listResp.Models) > 0 {
		currentModel = listResp.Models[0].Name
	}

	c.SetModel(currentModel)

	version := "detected"
	if len(listResp.Models) > 0 {
		if v := listResp.Models[0].Details.Family; v != "" {
			version = v
		}
	}

	return &OllamaStatus{
		Available:       true,
		Model:           currentModel,
		Version:         version,
		AvailableModels: availableModels,
	}, nil
}

// Chat sends a chat request to the Ollama API.
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
