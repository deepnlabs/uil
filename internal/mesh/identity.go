package mesh

import (
    "crypto/rand"
    "encoding/hex"
    "os"
)

type NodeIdentity struct {
    ID   string
    Addr string
}

func LoadOrCreateIdentity() (*NodeIdentity, error) {
    path := "/var/lib/uild/node_id"

    if data, err := os.ReadFile(path); err == nil {
        return &NodeIdentity{ID: string(data)}, nil
    }

    // Generate random 32-byte ID
    buf := make([]byte, 32)
    rand.Read(buf)
    id := hex.EncodeToString(buf)

    os.WriteFile(path, []byte(id), 0644)
    return &NodeIdentity{ID: id}, nil
}
