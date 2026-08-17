package uil

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/sha3"
)

const (
	// ProtocolVersion defines the current UIL specification version for v0.5
	ProtocolVersion = "UIP-0.5-PQ"

	// Substrate Targets
	SubstrateAuto         = "SUBSTRATE_AUTO"
	SubstrateRadeonVulkan = "RADEON_840M_VULKAN"
	SubstrateNvidiaCUDA   = "NVIDIA_RTX_CUDA"
	SubstrateHailoNPU     = "HAILO8_NPU"
	SubstrateMemryXNPU    = "MEMRYX_NPU"
	SubstrateXDNA2NPU     = "AMD_XDNA2_NPU"

	// Cryptographic Hash Algorithms
	HashAlgSHA3_256 = "SHA3-256"
	HashAlgSHA3_512 = "SHA3-512"
	HashAlgBLAKE3   = "BLAKE3"
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
	HashAlg         string                 `json:"hash_alg"`
	ProofCommitment string                 `json:"proof_commitment,omitempty"`
	PQSignature     string                 `json:"pq_signature,omitempty"` // Base64 Dilithium/SPHINCS+ sig
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
		Timestamp:       time.Now().UTC().Format(time.RFC3339Nano),
		ImportanceScore: 1.0,
		SubstrateTarget: substrate,
		HashAlg:         HashAlgSHA3_256, // Default to SHA3 for PQ resilience
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

// ComputePayloadHash calculates a deterministic hash using the configured algorithm (SHA3-256 / SHA3-512).
func (e *UILEnvelope) ComputePayloadHash() (string, error) {
	bytesData, err := json.Marshal(e.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload for hashing: %w", err)
	}

	switch e.HashAlg {
	case HashAlgSHA3_256, "":
		hash := sha3.Sum256(bytesData)
		return hex.EncodeToString(hash[:]), nil
	case HashAlgSHA3_512:
		hash := sha3.Sum512(bytesData)
		return hex.EncodeToString(hash[:]), nil
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", e.HashAlg)
	}
}
