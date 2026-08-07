package app

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ollamaReachable reports whether an Ollama server responds on localhost:11434.
func ollamaReachable() bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:11434", 300*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ── qwen3FallbackForRAM (pure function) ────────────────────────────────────

func TestQwen3FallbackForRAM(t *testing.T) {
	tests := []struct {
		name      string
		ramGB     float64
		wantModel string
		wantLabel string
	}{
		{name: "high RAM ≥8GB → 4B", ramGB: 16, wantModel: qwen3_4B_Model, wantLabel: qwen3_4B_Label + " (recommended for 16GB RAM)"},
		{name: "exactly 8GB → 4B", ramGB: 8, wantModel: qwen3_4B_Model, wantLabel: qwen3_4B_Label + " (recommended for 8GB RAM)"},
		{name: "6GB → 1.7B", ramGB: 6, wantModel: qwen3_1_7B_Model, wantLabel: qwen3_1_7B_Label + " (recommended for 6GB RAM)"},
		{name: "exactly 4GB → 1.7B", ramGB: 4, wantModel: qwen3_1_7B_Model, wantLabel: qwen3_1_7B_Label + " (recommended for 4GB RAM)"},
		{name: "2GB → 0.6B", ramGB: 2, wantModel: qwen3_0_6B_Model, wantLabel: qwen3_0_6B_Label + " (recommended for 2GB RAM)"},
		{name: "0GB → 0.6B", ramGB: 0, wantModel: qwen3_0_6B_Model, wantLabel: qwen3_0_6B_Label + " (recommended for 0GB RAM)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModel, gotLabel := qwen3FallbackForRAM(tt.ramGB)
			if gotModel != tt.wantModel {
				t.Errorf("qwen3FallbackForRAM(%v) model = %q, want %q", tt.ramGB, gotModel, tt.wantModel)
			}
			if gotLabel != tt.wantLabel {
				t.Errorf("qwen3FallbackForRAM(%v) label = %q, want %q", tt.ramGB, gotLabel, tt.wantLabel)
			}
		})
	}
}

// ── GetAISetupRecommendation (structure + internal consistency) ────────────

func TestAIOps_GetAISetupRecommendation_Structure(t *testing.T) {
	// Use a minimal AIOps — no App construction needed; ctx nil is fine
	// because CheckOllamaWithContext falls back gracefully when Ollama is absent.
	ai := &AIOps{ctx: context.Background(), contextCache: &systemContextCache{}}
	rec := ai.GetAISetupRecommendation()

	if rec.SystemRAMGB <= 0 {
		t.Errorf("SystemRAMGB = %.2f, want > 0", rec.SystemRAMGB)
	}
	if rec.SystemCPUThreads <= 0 {
		t.Errorf("SystemCPUThreads = %d, want > 0", rec.SystemCPUThreads)
	}
	if rec.Timestamp == "" {
		t.Error("Timestamp is empty, want formatted timestamp")
	}
	if len(rec.FallbackModels) != 4 {
		t.Fatalf("FallbackModels has %d entries, want 4", len(rec.FallbackModels))
	}

	// Fallback list must be the full QWEN 3.x ladder, in size-descending order
	wantNames := []string{qwen3_8B_Model, qwen3_4B_Model, qwen3_1_7B_Model, qwen3_0_6B_Model}
	for i, m := range rec.FallbackModels {
		if m.Name != wantNames[i] {
			t.Errorf("FallbackModels[%d].Name = %q, want %q", i, m.Name, wantNames[i])
		}
		if m.Label == "" {
			t.Errorf("FallbackModels[%d].Label is empty", i)
		}
		if m.SizeGB <= 0 {
			t.Errorf("FallbackModels[%d].SizeGB = %.1f, want > 0", i, m.SizeGB)
		}
	}
}

func TestAIOps_GetAISetupRecommendation_Logic(t *testing.T) {
	ai := &AIOps{ctx: context.Background(), contextCache: &systemContextCache{}}
	rec := ai.GetAISetupRecommendation()

	// CanRunQwythos must be consistent with the reported system specs
	expectQwythos := rec.SystemRAMGB >= 12 && rec.SystemCPUThreads >= 4
	if rec.CanRunQwythos != expectQwythos {
		t.Errorf("CanRunQwythos = %t, want %t (RAM=%.1fGB, threads=%d)",
			rec.CanRunQwythos, expectQwythos, rec.SystemRAMGB, rec.SystemCPUThreads)
	}

	// RecommendedModel must match the selection logic
	if rec.CanRunQwythos {
		if rec.RecommendedModel != qwythosModel {
			t.Errorf("RecommendedModel = %q, want qwythos %q", rec.RecommendedModel, qwythosModel)
		}
		if rec.RecommendedLabel != qwythosLabel {
			t.Errorf("RecommendedLabel = %q, want %q", rec.RecommendedLabel, qwythosLabel)
		}
		if !strings.Contains(rec.RecommendedModel, "Qwythos") {
			t.Errorf("RecommendedModel %q does not reference Qwythos", rec.RecommendedModel)
		}
	} else {
		wantModel, wantLabel := qwen3FallbackForRAM(rec.SystemRAMGB)
		if rec.RecommendedModel != wantModel {
			t.Errorf("RecommendedModel = %q, want fallback %q", rec.RecommendedModel, wantModel)
		}
		if rec.RecommendedLabel != wantLabel {
			t.Errorf("RecommendedLabel = %q, want %q", rec.RecommendedLabel, wantLabel)
		}
	}

	// PullRequired must be the inverse of "recommended model present":
	// in the Qwythos branch the code sets PullRequired = !QwythosExists,
	// so a present model must never require a pull (and vice versa).
	// (Previously the assertion was inverted and flagged the correct state
	// whenever a Qwythos model existed in Ollama.)
	if rec.CanRunQwythos && rec.PullRequired == rec.QwythosExists {
		t.Errorf("QwythosExists = %t but PullRequired = %t — want inverse (model present ⇒ no pull needed)",
			rec.QwythosExists, rec.PullRequired)
	}
}

// ── isTrivialMessage (pure function) ───────────────────────────────────────

func TestIsTrivialMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "greeting", msg: "hi", want: true},
		{name: "hello", msg: "hello there", want: true},
		{name: "thanks", msg: "thanks!", want: true},
		{name: "empty", msg: "", want: true},
		{name: "short greeting with exclamation", msg: "hey you", want: true},
		{name: "cpu keyword", msg: "how is my cpu", want: false},
		{name: "mem keyword", msg: "mem usage", want: false},
		{name: "disk keyword", msg: "disk full?", want: false},
		{name: "network keyword", msg: "network down", want: false},
		{name: "status keyword", msg: "status?", want: false},
		{name: "long question", msg: "can you tell me what the current utilization of the system is right now please", want: false},
		{name: "case insensitive", msg: "CPU", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTrivialMessage(tt.msg); got != tt.want {
				t.Errorf("isTrivialMessage(%q) = %t, want %t", tt.msg, got, tt.want)
			}
		})
	}
}

// ── Context cache (getCachedSnapshot / setCachedSnapshot) ──────────────────

func TestAIOps_ContextCache_FreshHit(t *testing.T) {
	ai := &AIOps{ctx: context.Background(), contextCache: &systemContextCache{}}

	ai.setCachedSnapshot("snapshot-v1")
	snap, ok := ai.getCachedSnapshot(false)
	if !ok {
		t.Fatal("getCachedSnapshot(false) = not ok, want fresh hit")
	}
	if snap != "snapshot-v1" {
		t.Errorf("getCachedSnapshot returned %q, want %q", snap, "snapshot-v1")
	}
}

func TestAIOps_ContextCache_ForceRefresh(t *testing.T) {
	ai := &AIOps{ctx: context.Background(), contextCache: &systemContextCache{}}

	ai.setCachedSnapshot("stale")
	if _, ok := ai.getCachedSnapshot(true); ok {
		t.Error("getCachedSnapshot(true) = ok, want miss when forceRefresh=true")
	}
}

func TestAIOps_ContextCache_Expired(t *testing.T) {
	ai := &AIOps{ctx: context.Background(), contextCache: &systemContextCache{}}

	ai.setCachedSnapshot("old")
	// Rewind the timestamp past the TTL
	ai.contextCache.mu.Lock()
	ai.contextCache.timestamp = time.Now().Add(-(contextCacheTTL + time.Second))
	ai.contextCache.mu.Unlock()

	if _, ok := ai.getCachedSnapshot(false); ok {
		t.Error("getCachedSnapshot(false) = ok, want miss after TTL expiry")
	}
}

func TestAIOps_ContextCache_NilCache(t *testing.T) {
	ai := &AIOps{ctx: context.Background()}

	// setCachedSnapshot must be a no-op with nil cache (no panic)
	ai.setCachedSnapshot("ignored")
	if _, ok := ai.getCachedSnapshot(false); ok {
		t.Error("getCachedSnapshot with nil cache = ok, want miss")
	}
}

// ── GetModelfile (error path) ──────────────────────────────────────────────

func TestAIOps_GetModelfile_Missing(t *testing.T) {
	dir := t.TempDir()
	ai := &AIOps{ctx: context.Background(), dataDir: dir}

	content, err := ai.GetModelfile()
	if err == nil {
		t.Fatal("GetModelfile returned nil error, want error for missing file")
	}
	if content != "" {
		t.Errorf("GetModelfile returned content %q on error, want empty", content)
	}
	if !strings.Contains(err.Error(), "universalops.modelfile") {
		t.Errorf("error %q does not mention universalops.modelfile", err.Error())
	}
}

func TestAIOps_GetModelfile_Present(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "universalops.modelfile")
	content := "FROM qwen3:8b\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test modelfile: %v", err)
	}
	ai := &AIOps{ctx: context.Background(), dataDir: dir}

	got, err := ai.GetModelfile()
	if err != nil {
		t.Fatalf("GetModelfile returned error: %v", err)
	}
	if got != content {
		t.Errorf("GetModelfile = %q, want %q", got, content)
	}
}

// ── SetupOllamaPersona (error path: Ollama unavailable) ────────────────────

func TestAIOps_SetupOllamaPersona_NoOllama(t *testing.T) {
	// This test exercises the "Ollama is down" path: CheckOllamaWithContext
	// errors, the persona flow must surface a contextualized error instead of
	// hanging or writing files. It only runs when Ollama is NOT reachable —
	// with a live server the pull path would emit Wails progress events that
	// require a real lifecycle context (panic risk) and perform real network
	// side effects, so we skip there.
	if ollamaReachable() {
		t.Skip("Ollama is running — this test covers the unavailable path only")
	}

	ai := &AIOps{ctx: context.Background(), dataDir: t.TempDir(), contextCache: &systemContextCache{}}
	err := ai.SetupOllamaPersona(qwen3_0_6B_Model)
	if err == nil {
		t.Fatal("SetupOllamaPersona returned nil error while Ollama is unavailable, want error")
	}
	// The error must be contextualized (mention the pull), not a raw panic
	if !strings.Contains(err.Error(), "failed to pull base model") {
		t.Errorf("unexpected error: %v", err)
	}
}
