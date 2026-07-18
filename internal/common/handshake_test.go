package common

import (
	"testing"
)

func TestHandshakeRegistry(t *testing.T) {
	reg := GetHandshakeRegistry()

	params := map[string]interface{}{"pid": 1234}
	id := reg.Register("kill_process", params)

	if len(id) != 32 {
		t.Errorf("expected 32-char hex ID, got %d", len(id))
	}

	// Test Consume
	pending, err := reg.Consume(id)
	if err != nil {
		t.Fatalf("failed to consume handshake: %v", err)
	}
	if pending.Action != "kill_process" {
		t.Errorf("expected kill_process, got %s", pending.Action)
	}
	if pending.Params["pid"] != 1234 {
		t.Errorf("expected pid 1234, got %v", pending.Params["pid"])
	}

	// Test Double Consume (should fail)
	_, err = reg.Consume(id)
	if err == nil {
		t.Error("expected error on double consume, got nil")
	}
}

func TestHandshakeCreatePreview(t *testing.T) {
	reg := GetHandshakeRegistry()
	params := map[string]interface{}{"ip": "1.1.1.1"}
	preview := reg.CreatePreview("block_ip", params)

	if preview.Action != "block_ip" {
		t.Errorf("expected block_ip, got %s", preview.Action)
	}
	if len(preview.Risks) == 0 {
		t.Error("expected risks to be populated")
	}
	if preview.HandshakeID == "" {
		t.Error("expected HandshakeID to be populated")
	}
}
