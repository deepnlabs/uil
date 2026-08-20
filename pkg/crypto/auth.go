package crypto

import (
    "crypto/ed25519"
    "encoding/base64"
    "fmt"

    "github.com/deepnlabs/uil/pkg/uil"
)

// SignEnvelope signs the envelope's ProofCommitment using the node's private key.
// The signature is stored in PQSignature (base64-encoded).
func SignEnvelope(env *uil.UILEnvelope, priv ed25519.PrivateKey) error {
    if env == nil {
        return fmt.Errorf("nil envelope")
    }
    if env.ProofCommitment == "" {
        return fmt.Errorf("cannot sign envelope: missing ProofCommitment")
    }

    sig := ed25519.Sign(priv, []byte(env.ProofCommitment))
    env.PQSignature = base64.StdEncoding.EncodeToString(sig)
    return nil
}

// VerifyEnvelope verifies PQSignature against ProofCommitment using the sender's public key.
func VerifyEnvelope(env *uil.UILEnvelope, pub ed25519.PublicKey) bool {
    if env == nil {
        return false
    }
    if env.ProofCommitment == "" || env.PQSignature == "" {
        return false
    }

    sigBytes, err := base64.StdEncoding.DecodeString(env.PQSignature)
    if err != nil {
        return false
    }

    return ed25519.Verify(pub, []byte(env.ProofCommitment), sigBytes)
}
