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
