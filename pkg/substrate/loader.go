package substrate

import (
	"fmt"
	"plugin"
)

// LoadPlugin dynamically loads external binary plugins (.so files) at runtime.
func LoadPlugin(path string) (SubstrateDriver, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin at %s: %w", path, err)
	}

	sym, err := p.Lookup("NewDriver")
	if err != nil {
		// Fallback check for DriverSymbol if needed
		return nil, fmt.Errorf("plugin missing required exported symbol 'NewDriver': %w", err)
	}

	newDriverFunc, ok := sym.(func() any)
	if !ok {
		return nil, fmt.Errorf("symbol 'NewDriver' is not a func() any")
	}

	rawInstance := newDriverFunc()
	driver, ok := rawInstance.(SubstrateDriver)
	if !ok {
		return nil, fmt.Errorf("instantiated driver does not fulfill substrate.SubstrateDriver interface")
	}

	return driver, nil
}
