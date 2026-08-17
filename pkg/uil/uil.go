package uil

import (
	"errors"
	"fmt"
	"time"
)

const (
	ProtocolVersion = "UIP-0.5-PQ"

	SubstrateAuto         = "SUBSTRATE_AUTO"
	SubstrateRadeonVulkan = "RADEON_840M_VULKAN"
	SubstrateNvidiaCUDA   = "NVIDIA_RTX_CUDA"
	SubstrateHailoNPU     = "HAILO8_NPU"
	SubstrateMemryXNPU    = "MEMRYX_NPU"
	SubstrateXDNA2NPU     = "AMD_XDNA2_NPU"
)

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
	PQSignature     string                 `json:"pq_signature,omitempty"`
	Payload         map[string]interface{} `json:"payload"`
}

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
		HashAlg:         "SHA3-256",
		Payload:         payload,
	}
}

func (e *UILEnvelope) Validate() error {
	if e.ProtocolVersion != ProtocolVersion {
		return fmt.Errorf("unsupported protocol version: %s (expected %s)", e.ProtocolVersion, ProtocolVersion)
	}
	if e.EnvelopeID == "" || e.SourceNode == "" || e.Timestamp == "" || e.Payload == nil {
		return errors.New("invalid envelope parameters")
	}
	return nil
}
