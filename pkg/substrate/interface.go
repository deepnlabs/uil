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
