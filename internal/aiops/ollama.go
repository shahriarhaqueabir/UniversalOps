package aiops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
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

// ollamaChatRequest is the request body for /api/chat.
type ollamaChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ollamaChatResponse is the response body from /api/chat.
type ollamaChatResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

// ollamaTagsResponse is the response from /api/tags.
type ollamaTagsResponse struct {
	Models []ollamaModel `json:"models"`
}

type ollamaModel struct {
	Name string `json:"name"`
}

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultOllamaModel = "llama3.2"
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

// CheckOllama checks if the Ollama service is available.
func CheckOllama() (*OllamaStatus, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(getOllamaURL() + "/api/tags")
	if err != nil {
		return &OllamaStatus{Available: false}, nil
	}
	defer resp.Body.Close()

	var tagsResp ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return &OllamaStatus{Available: true, Model: "unknown"}, nil
	}

	modelName := getOllamaModel()
	availableModels := make([]string, 0, len(tagsResp.Models))
	if len(tagsResp.Models) > 0 {
		// Verify if the requested model exists, otherwise fallback to the first available
		found := false
		for _, m := range tagsResp.Models {
			availableModels = append(availableModels, m.Name)
			if m.Name == modelName {
				found = true
			}
		}
		if !found {
			modelName = tagsResp.Models[0].Name
		}
	}

	return &OllamaStatus{
		Available:       true,
		Model:           modelName,
		Version:         "detected",
		AvailableModels: availableModels,
	}, nil
}

// Chat sends a chat request to the Ollama API and returns the response text.
func Chat(messages []ChatMessage) (string, error) {
	client := &http.Client{Timeout: httpTimeout}

	reqBody := ollamaChatRequest{
		Model:    getOllamaModel(),
		Messages: messages,
		Stream:   false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := client.Post(
		getOllamaURL()+"/api/chat",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var chatResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return chatResp.Message.Content, nil
}
