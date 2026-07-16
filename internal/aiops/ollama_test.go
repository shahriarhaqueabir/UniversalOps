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

func TestCheckOllamaBinary(t *testing.T) {
	// This function should exist in ollama.go
	exists := CheckOllamaBinary()
	// We don't necessarily know if it exists on the host, but the test
	// serves to verify the function signature and availability.
	t.Logf("Ollama binary exists: %v", exists)
}

func TestPullModel_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/pull" {
			resp := api.ProgressResponse{Status: "success"}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	os.Setenv("OLLAMA_HOST", server.URL)
	defer os.Unsetenv("OLLAMA_HOST")

	var progressCalled bool
	err := PullModel("test-model", func(p api.ProgressResponse) error {
		progressCalled = true
		return nil
	})

	if err != nil {
		t.Fatalf("PullModel failed: %v", err)
	}
	if !progressCalled {
		t.Error("Expected progress callback to be called")
	}
}

func TestCreateModel_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/create" {
			resp := api.ProgressResponse{Status: "success"}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	os.Setenv("OLLAMA_HOST", server.URL)
	defer os.Unsetenv("OLLAMA_HOST")

	err := CreateModel("test-persona", "llama3", "You are a test", map[string]any{"temperature": 0.1})
	if err != nil {
		t.Fatalf("CreateModel failed: %v", err)
	}
}
