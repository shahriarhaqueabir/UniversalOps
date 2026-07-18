package aiops

import (
	"os"
	"testing"
)

func TestParseModelfile(t *testing.T) {
	content := `FROM base-model
SYSTEM """
You are a test assistant.
Multi-line instructions.
"""
PARAMETER temperature 0.7
PARAMETER top_p 0.9
PARAMETER stop "END"
PARAMETER stop "STOP"
`
	tmpfile, err := os.CreateTemp("", "Modelfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	config, err := ParseModelfile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseModelfile failed: %v", err)
	}

	if config.From != "base-model" {
		t.Errorf("Expected From 'base-model', got %q", config.From)
	}

	expectedSystem := "You are a test assistant.\nMulti-line instructions."
	if config.System != expectedSystem {
		t.Errorf("Expected System %q, got %q", expectedSystem, config.System)
	}

	if config.Parameters["temperature"] != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", config.Parameters["temperature"])
	}

	stops, ok := config.Parameters["stop"].([]string)
	if !ok {
		t.Fatalf("Expected stop parameter to be []string, got %T", config.Parameters["stop"])
	}
	if len(stops) != 2 || stops[0] != "END" || stops[1] != "STOP" {
		t.Errorf("Expected stops [END STOP], got %v", stops)
	}
}
