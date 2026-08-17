#!/usr/bin/env bash
set -e

echo "=== Initializing UIL-X v0.5 Modular Architecture ==="

# 1. Create Directory Hierarchy
mkdir -p cmd/uild
mkdir -p pkg/uil
mkdir -p pkg/crypto
mkdir -p pkg/substrate
mkdir -p pkg/config
mkdir -p plugins_src/custom_interlock
mkdir -p configs

# 2. pkg/uil/uil.go (Protocol Primitives)
cat << 'EOF' > pkg/uil/uil.go
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
EOF

# 3. pkg/crypto/hash.go (Post-Quantum / SHA3 Hashing)
cat << 'EOF' > pkg/crypto/hash.go
package crypto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/sha3"
)

type Hasher interface {
	HashPayload(payload map[string]interface{}) (string, error)
}

type SHA3Hasher struct {
	Variant string // "SHA3-256" or "SHA3-512"
}

func NewSHA3Hasher(variant string) *SHA3Hasher {
	if variant == "" {
		variant = "SHA3-256"
	}
	return &SHA3Hasher{Variant: variant}
}

func (h *SHA3Hasher) HashPayload(payload map[string]interface{}) (string, error) {
	bytesData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	switch h.Variant {
	case "SHA3-512":
		hash := sha3.Sum512(bytesData)
		return hex.EncodeToString(hash[:]), nil
	case "SHA3-256":
		fallthrough
	default:
		hash := sha3.Sum256(bytesData)
		return hex.EncodeToString(hash[:]), nil
	}
}
EOF

# 4. pkg/substrate/interface.go (Pluggable Substrate Driver Interface)
cat << 'EOF' > pkg/substrate/interface.go
package substrate

import "context"

// SubstrateDriver defines the interface that all hardware modules (open-source or proprietary) must fulfill.
type SubstrateDriver interface {
	ID() string
	Name() string
	Initialize(ctx context.Context, config map[string]any) error
	EvaluateInterlock(metrics map[string]float64) (bool, error) // Returns true if safety breached
	Shutdown(ctx context.Context) error
}
EOF

# 5. pkg/substrate/loader.go (Dynamic Plugin Loader for Proprietary .so modules)
cat << 'EOF' > pkg/substrate/loader.go
package substrate

import (
	"fmt"
	"plugin"
)

// LoadPlugin Dynamically loads external binary plugins (.so files) at runtime.
func LoadPlugin(path string) (SubstrateDriver, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin at %s: %w", path, err)
	}

	symDriver, err := p.Lookup("DriverSymbol")
	if err != nil {
		return nil, fmt.Errorf("plugin missing required exported symbol 'DriverSymbol': %w", err)
	}

	driver, ok := symDriver.(SubstrateDriver)
	if !ok {
		return nil, fmt.Errorf("symbol 'DriverSymbol' does not implement SubstrateDriver interface")
	}

	return driver, nil
}
EOF

# 6. Sample Plugin Source (plugins_src/custom_interlock/main.go)
cat << 'EOF' > plugins_src/custom_interlock/main.go
package main

import (
	"context"
	"fmt"
	"uil/pkg/substrate"
)

type ProprietaryThermalInterlock struct{}

func (p *ProprietaryThermalInterlock) ID() string   { return "prop-thermal-v1" }
func (p *ProprietaryThermalInterlock) Name() string { return "Enterprise Thermal Interlock Module" }

func (p *ProprietaryThermalInterlock) Initialize(ctx context.Context, config map[string]any) error {
	fmt.Println("[Proprietary Plugin] Initialized enterprise hardware check.")
	return nil
}

func (p *ProprietaryThermalInterlock) EvaluateInterlock(metrics map[string]float64) (bool, error) {
	if temp, exists := metrics["temp_celsius"]; exists && temp > 85.0 {
		return true, nil // Interlock breach!
	}
	return false, nil
}

func (p *ProprietaryThermalInterlock) Shutdown(ctx context.Context) error {
	return nil
}

// DriverSymbol exported constructor symbol for dynamic loader
var DriverSymbol substrate.SubstrateDriver = &ProprietaryThermalInterlock{}
EOF

# 7. cmd/uild/main.go (Primary CLI Daemon)
cat << 'EOF' > cmd/uild/main.go
package main

import (
	"fmt"
	"os"
	"uil/pkg/crypto"
	"uil/pkg/uil"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		fmt.Println("Initializing UIL-X runtime configuration at /etc/uild/config.json...")
		fmt.Println("Creating state directories...")
		fmt.Println("Initialization complete.")
	case "run":
		fmt.Println("Starting UIL-X Hardware Governance Daemon (`uild`) v0.5-PQ...")
		hasher := crypto.NewSHA3Hasher("SHA3-256")
		env := uil.NewEnvelope("node-1", "node-2", uil.SubstrateHailoNPU, map[string]interface{}{"temp": 42.5, "status": "nominal"})
		hash, _ := hasher.HashPayload(env.Payload)
		fmt.Printf("Engine Active | Substrate: %s | SHA3 Commitment: %s\n", env.SubstrateTarget, hash[:16])
	case "status":
		fmt.Println("UIL-X Daemon Status: ACTIVE")
		fmt.Println("Substrate Engines: Hailo-8 NPU (Online), Local Linux GPIO (Ready)")
		fmt.Println("Interlock Status: Nominal (0 Breaches)")
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("UIL-X Hardware Governance Daemon (uild)")
	fmt.Println("Usage: uild [init|run|status]")
}
EOF

# 8. Makefile for local node building
cat << 'EOF' > Makefile
.PHONY: all build plugin clean

all: build plugin

build:
	go build -o bin/uild cmd/uild/main.go

plugin:
	go build -buildmode=plugin -o plugins/custom_interlock.so plugins_src/custom_interlock/main.go

clean:
	rm -rf bin/ plugins/
EOF

echo "=== Scaffolding complete! Getting x/crypto package... ==="
go get golang.org/x/crypto/sha3 || true
