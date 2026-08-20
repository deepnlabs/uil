package mesh

import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"
)

const (
    identityPath  = "/var/lib/uild/node_id"
    privKeyPath   = "/var/lib/uild/node_key.priv"
    pubKeyPath    = "/var/lib/uild/node_key.pub"
)

type NodeIdentity struct {
    NodeID    string
    PublicKey ed25519.PublicKey
    PrivateKey ed25519.PrivateKey
}

func LoadOrCreateIdentity() (*NodeIdentity, error) {
    // Ensure directory exists
    dir := filepath.Dir(identityPath)
    if err := os.MkdirAll(dir, 0750); err != nil {
        return nil, fmt.Errorf("failed to create identity directory: %w", err)
    }

    // Load or create node ID
    nodeID, err := loadOrCreateNodeID()
    if err != nil {
        return nil, err
    }

    // Load or create keypair
    pub, priv, err := loadOrCreateKeypair()
    if err != nil {
        return nil, err
    }

    return &NodeIdentity{
        NodeID:    nodeID,
        PublicKey: pub,
        PrivateKey: priv,
    }, nil
}

func loadOrCreateNodeID() (string, error) {
    if data, err := os.ReadFile(identityPath); err == nil {
        id := string(data)
        if len(id) == 64 {
            return id, nil
        }
    }

    buf := make([]byte, 32)
    if _, err := rand.Read(buf); err != nil {
        return "", fmt.Errorf("failed to generate node ID: %w", err)
    }
    id := hex.EncodeToString(buf)

    if err := os.WriteFile(identityPath, []byte(id), 0640); err != nil {
        return "", fmt.Errorf("failed to write node ID: %w", err)
    }

    return id, nil
}

func loadOrCreateKeypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
    // Try loading existing keys
    privBytes, privErr := os.ReadFile(privKeyPath)
    pubBytes, pubErr := os.ReadFile(pubKeyPath)

    if privErr == nil && pubErr == nil {
        return ed25519.PublicKey(pubBytes), ed25519.PrivateKey(privBytes), nil
    }

    // Generate new keypair
    pub, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to generate keypair: %w", err)
    }

    // Persist keys
    if err := os.WriteFile(privKeyPath, priv, 0600); err != nil {
        return nil, nil, fmt.Errorf("failed to write private key: %w", err)
    }
    if err := os.WriteFile(pubKeyPath, pub, 0644); err != nil {
        return nil, nil, fmt.Errorf("failed to write public key: %w", err)
    }

    return pub, priv, nil
}
