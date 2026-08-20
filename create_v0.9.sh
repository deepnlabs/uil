#!/usr/bin/env bash
set -e

echo "Creating UIL v0.9 scaffolding..."

# --- Directory structure ---
mkdir -p v0.9/{cmd/uild,cmd/uilctl,internal/{core,plugin},pkg/{api,crypto,utils},plugins/examples,docs,scripts}

###############################################
# CMD binaries
###############################################

cat > v0.9/cmd/uild/main.go << 'EOF'
// v0.9 uild daemon entrypoint (stub)
package main

import "github.com/deepnlabs/uil/v0.9/internal/core"

func main() {
    core.Run()
}
EOF

cat > v0.9/cmd/uilctl/main.go << 'EOF'
// v0.9 uilctl CLI entrypoint (stub)
package main

func main() {
    // TODO: implement CLI commands
}
EOF

###############################################
# INTERNAL CORE
###############################################

cat > v0.9/internal/core/runtime.go << 'EOF'
// Core runtime loop (stub)
package core

func Run() {
    // TODO: scheduler + plugin dispatch
}
EOF

cat > v0.9/internal/core/config.go << 'EOF'
// v0.9 config loader (stub)
package core

type Config struct {
    NodeID string `json:"node_id"`
}

func LoadConfig() Config {
    // TODO: load config from file
    return Config{}
}
EOF

cat > v0.9/internal/core/identity.go << 'EOF'
// v0.9 identity loader (copied from v0.8.x)
package core

import (
    "crypto/rand"
    "encoding/hex"
    "os"
)

const identityPath = "/var/lib/uild/node_id"

func LoadOrCreateIdentity() (string, error) {
    if data, err := os.ReadFile(identityPath); err == nil {
        return string(data), nil
    }

    buf := make([]byte, 32)
    rand.Read(buf)
    id := hex.EncodeToString(buf)

    os.WriteFile(identityPath, []byte(id), 0644)
    return id, nil
}
EOF

###############################################
# INTERNAL PLUGIN SYSTEM
###############################################

cat > v0.9/internal/plugin/abi.go << 'EOF'
// v0.9 plugin ABI (stub)
package plugin

type Plugin interface {
    Init() error
    Tick() error
    Shutdown() error
}
EOF

cat > v0.9/internal/plugin/loader.go << 'EOF'
// v0.9 plugin loader (stub)
package plugin

func LoadPlugins(dir string) ([]Plugin, error) {
    // TODO: load .so plugins
    return nil, nil
}
EOF

###############################################
# PKG API (Envelope copied from v0.8.x)
###############################################

cat > v0.9/pkg/api/envelope.go << 'EOF'
// v0.9 UILEnvelope (copied from v0.8.x)
package api

import (
    "errors"
    "fmt"
    "time"
)

const ProtocolVersion = "UIP-0.5-PQ"

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
        substrate = "SUBSTRATE_AUTO"
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

###############################################
# PKG CRYPTO (copy from v0.8.x)
###############################################

cp pkg/crypto/hash.go v0.9/pkg/crypto/sha3.go

###############################################
# PKG UTILS
###############################################

cat > v0.9/pkg/utils/errors.go << 'EOF'
// v0.9 error helpers (stub)
package utils
EOF

###############################################
# PLUGIN EXAMPLES
###############################################

cat > v0.9/plugins/examples/mesh_basic/main.go << 'EOF'
// Example mesh plugin (stub)
package main

import "fmt"

func Init() error {
    fmt.Println("mesh_basic plugin initialized")
    return nil
}

func Tick() error {
    return nil
}

func Shutdown() error {
    return nil
}
EOF

echo "v0.9 scaffolding created successfully."
