package uil

import (
	"testing"
)

func TestNewEnvelope(t *testing.T) {
	payload := map[string]interface{}{"prompt": "Test UIL validation"}
	env := NewEnvelope("node-2", "node-1", SubstrateHailoNPU, payload)

	if err := env.Validate(); err != nil {
		t.Fatalf("Expected valid envelope, got error: %v", err)
	}

	hash, err := env.ComputePayloadHash()
	if err != nil || hash == "" {
		t.Fatalf("Failed to compute payload hash: %v", err)
	}
}
