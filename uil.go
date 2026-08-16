package uil

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// ProtocolVersion defines the current UIL specification version
	ProtocolVersion = "UIP-1.0-CDID"

	// Substrate Targets
	SubstrateAuto        = "SUBSTRATE_AUTO"
	SubstrateRadeonVulkan = "RADEON_840M_VULKAN"
	SubstrateNvidiaCUDA   = "NVIDIA_RTX_CUDA"
	SubstrateHailoNPU    = "HAILO8_NPU"
	SubstrateMemryXNPU   = "MEMRYX_NPU"
	SubstrateXDNA2NPU    = "AMD_XDNA2_NPU"
)

// UILEnvelope defines the standardized message container across the UIL Mesh.
type UILEnvelope struct {
	ProtocolVersion string                 `json:"protocol_version"`
	EnvelopeID      string                 `json:"envelope_id"`
	SourceNode      string                 `json:"source_node"`
	TargetNode      string                 `json:"target_node,omitempty"`
	Timestamp       string                 `json:"timestamp"`
	ImportanceScore float64                `json:"importance_score"`
	SubstrateTarget string                 `json:"substrate_target"`
	ProofCommitment string                 `json:"proof_commitment,omitempty"`
	Payload         map[string]interface{} `json:"payload"`
}

// NewEnvelope constructs a compliant UIL Envelope with auto-generated ID and RFC3339 timestamp.
func NewEnvelope(sourceNode, targetNode, substrate string, payload map[string]interface{}) *UILEnvelope {
	if substrate == "" {
		substrate = SubstrateAuto
	}
	return &UILEnvelope{
		ProtocolVersion: ProtocolVersion,
		EnvelopeID:      fmt.Sprintf("uil-env-%d", time.Now().UnixNano()),
		SourceNode:      sourceNode,
		TargetNode:      targetNode,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		ImportanceScore: 1.0,
		SubstrateTarget: substrate,
		Payload:         payload,
	}
}

// Validate checks structural integrity and protocol compliance of the envelope.
func (e *UILEnvelope) Validate() error {
	if e.ProtocolVersion != ProtocolVersion {
		return fmt.Errorf("unsupported protocol version: %s (expected %s)", e.ProtocolVersion, ProtocolVersion)
	}
	if e.EnvelopeID == "" {
		return errors.New("envelope_id cannot be empty")
	}
	if e.SourceNode == "" {
		return errors.New("source_node cannot be empty")
	}
	if e.Timestamp == "" {
		return errors.New("timestamp cannot be empty")
	}
	if e.Payload == nil {
		return errors.New("payload cannot be nil")
	}
	return nil
}

// ComputePayloadHash calculates a deterministic SHA-256 hash of the payload for proof verification.
func (e *UILEnvelope) ComputePayloadHash() (string, error) {
	bytesData, err := json.Marshal(e.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload for hashing: %w", err)
	}
	hash := sha256.Sum256(bytesData)
	return hex.EncodeToString(hash[:]), nil
}
