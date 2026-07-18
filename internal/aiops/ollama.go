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

// effectiveModel is resolved by CheckOllama and read by Chat. Both can run
// concurrently (e.g. a status poll while a chat request is in flight), so
// access is guarded by effectiveModelMu instead of a bare package var (H5).
var (
	effectiveModelMu sync.RWMutex
	effectiveModel   string
)

func setEffectiveModel(model string) {
	effectiveModelMu.Lock()
	effectiveModel = model
	effectiveModelMu.Unlock()
}

func getEffectiveModel() string {
	effectiveModelMu.RLock()
	defer effectiveModelMu.RUnlock()
	return effectiveModel
}

// SetModel updates the effective model to be used for future Chat calls.
func SetModel(model string) {
	setEffectiveModel(model)
}

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultOllamaModel = "opsforall"
	defaultHttpTimeout = 60 * time.Second // Increased from 30s for general metadata queries
)

func getOllamaURL() string {
	if url := os.Getenv("OLLAMA_HOST"); url != "" {
		return url
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

func newClientWithTimeout(timeout time.Duration) *api.Client {
	baseURL := getOllamaURL()
	u, err := url.Parse(baseURL)
	if err != nil {
		u = &url.URL{Scheme: "http", Host: "localhost:11434"}
	}
	return api.NewClient(u, &http.Client{Timeout: timeout})
}

// CheckOllamaBinary returns true if the ollama binary is found in the system PATH.
func CheckOllamaBinary() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

// PullModel downloads a model from the Ollama library.
func PullModel(name string, onProgress func(api.ProgressResponse) error) error {
	// Pulling models can take a long time; use no timeout on the client (0)
	client := newClientWithTimeout(0)
	req := &api.PullRequest{
		Model: name,
	}

	err := client.Pull(context.Background(), req, onProgress)
	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	return nil
}

// CreateModel creates a new model from structured parameters.
// Note: v0.31.2 SDK uses structured fields instead of a raw Modelfile string.
func CreateModel(name string, from string, system string, parameters map[string]any) error {
	// Creating models (especially if it involves downloading layers) can take time
	client := newClientWithTimeout(0)
	req := &api.CreateRequest{
		Model:      name,
		From:       from,
		System:     system,
		Parameters: parameters,
	}

	err := client.Create(context.Background(), req, func(resp api.ProgressResponse) error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	return nil
}

// CheckOllama checks if the Ollama service is available using the typed SDK.
func CheckOllama() (*OllamaStatus, error) {
	client := newClient()

	listResp, err := client.List(context.Background())
	if err != nil {
		// Ollama not running or unreachable
		return &OllamaStatus{Available: false}, nil
	}

	modelName := getOllamaModel()
	availableModels := make([]string, 0, len(listResp.Models))
	if len(listResp.Models) > 0 {
		found := false
		for _, m := range listResp.Models {
			availableModels = append(availableModels, m.Name)
			// Fuzzy match to handle :latest or other tags
			if m.Name == modelName || strings.HasPrefix(m.Name, modelName+":") {
				found = true
				modelName = m.Name // Use the fully qualified name
			}
		}
		if !found {
			modelName = listResp.Models[0].Name
		}
	}

	// Store the effective model so Chat() uses the same resolved model
	setEffectiveModel(modelName)

	version := "detected"
	if len(listResp.Models) > 0 {
		if v := listResp.Models[0].Details.Family; v != "" {
			version = v
		}
	}

	return &OllamaStatus{
		Available:       true,
		Model:           modelName,
		Version:         version,
		AvailableModels: availableModels,
	}, nil
}

// Chat sends a chat request to the Ollama API using the typed SDK and returns the response text.
func Chat(messages []ChatMessage) (string, error) {
	// Use a longer timeout for chat generation, as reasoning can take time
	client := newClientWithTimeout(5 * time.Minute)

	// Convert our ChatMessage type to the SDK's Message type
	apiMessages := make([]api.Message, len(messages))
	for i, m := range messages {
		apiMessages[i] = api.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	// Use the model resolved by CheckOllama, falling back to env/default
	model := getEffectiveModel()
	if model == "" {
		model = getOllamaModel()
	}

	stream := false
	req := &api.ChatRequest{
		Model:    model,
		Messages: apiMessages,
		Stream:   &stream,
	}

	var responseText string
	err := client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
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
