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
