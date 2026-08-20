package substrate

import "context"

// Versioned driver interface for safety‑critical operation.
const SubstrateAPIVersion = 1

type SubstrateDriver interface {
    // Metadata
    ID() string
    Name() string

    // Version compliance
    APIVersion() int

    // Lifecycle
    Initialize(ctx context.Context, config map[string]any) error
    EvaluateInterlock(metrics map[string]float64) (bool, error)
    Shutdown(ctx context.Context) error
}
