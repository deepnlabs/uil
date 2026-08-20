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
