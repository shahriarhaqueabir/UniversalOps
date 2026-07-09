package aiops

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/ollama/ollama/api"
)

func TestCheckOllama_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			resp := api.ListResponse{
				Models: []api.ListModelResponse{
					{Name: "llama3.2"},
					{Name: "mistral"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	os.Setenv("OLLAMA_HOST", server.URL)
	defer os.Unsetenv("OLLAMA_HOST")

	status, err := CheckOllama()
	if err != nil {
		t.Fatalf("CheckOllama failed: %v", err)
	}

	if !status.Available {
		t.Error("Expected Ollama to be available")
	}

	if len(status.AvailableModels) != 2 {
		t.Errorf("Expected 2 models, got %d", len(status.AvailableModels))
	}
}

func TestChat_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			resp := api.ChatResponse{
				Message: api.Message{Role: "assistant", Content: "Mock response"},
				Done:    true,
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	os.Setenv("OLLAMA_HOST", server.URL)
	defer os.Unsetenv("OLLAMA_HOST")

	resp, err := Chat([]ChatMessage{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp != "Mock response" {
		t.Errorf("Expected 'Mock response', got %q", resp)
	}
}

func TestNewClient(t *testing.T) {
	u, _ := url.Parse("http://localhost:11434")
	client := api.NewClient(u, http.DefaultClient)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}
