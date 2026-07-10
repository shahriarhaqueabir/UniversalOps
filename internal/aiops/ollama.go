package aiops

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

var effectiveModel string // resolved by CheckOllama, used by Chat

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultOllamaModel = "agentic-coder"
	httpTimeout        = 30 * time.Second
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
	baseURL := getOllamaURL()
	u, err := url.Parse(baseURL)
	if err != nil {
		u = &url.URL{Scheme: "http", Host: "localhost:11434"}
	}
	return api.NewClient(u, &http.Client{Timeout: httpTimeout})
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
			if m.Name == modelName {
				found = true
			}
		}
		if !found {
			modelName = listResp.Models[0].Name
		}
	}

	// Store the effective model so Chat() uses the same resolved model
	effectiveModel = modelName

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
	client := newClient()

	// Convert our ChatMessage type to the SDK's Message type
	apiMessages := make([]api.Message, len(messages))
	for i, m := range messages {
		apiMessages[i] = api.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	// Use the model resolved by CheckOllama, falling back to env/default
	model := effectiveModel
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
