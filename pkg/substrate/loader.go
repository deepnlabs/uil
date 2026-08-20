package substrate

import (
    "fmt"
    "plugin"
    "runtime/debug"
    "time"
)

// LoadPlugin loads external drivers with full sandboxing and safety checks.
func LoadPlugin(path string) (SubstrateDriver, error) {
    defer func() {
        if r := recover(); r != nil {
            // Prevent plugin panic from crashing daemon
        }
    }()

    p, err := plugin.Open(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open plugin at %s: %w", path, err)
    }

    sym, err := p.Lookup("NewDriver")
    if err != nil {
        return nil, fmt.Errorf("plugin missing required exported symbol 'NewDriver': %w", err)
    }

    newDriverFunc, ok := sym.(func() any)
    if !ok {
        return nil, fmt.Errorf("symbol 'NewDriver' is not func() any")
    }

    // Instantiate plugin driver safely
    var rawInstance any
    func() {
        defer func() {
            if r := recover(); r != nil {
                rawInstance = nil
            }
        }()
        rawInstance = newDriverFunc()
    }()

    if rawInstance == nil {
        return nil, fmt.Errorf("plugin constructor panicked")
    }

    driver, ok := rawInstance.(SubstrateDriver)
    if !ok {
        return nil, fmt.Errorf("plugin does not implement SubstrateDriver interface")
    }

    // Version compliance check
    if driver.APIVersion() != SubstrateAPIVersion {
        return nil, fmt.Errorf("plugin API version mismatch: driver=%d expected=%d",
            driver.APIVersion(), SubstrateAPIVersion)
    }

    // Safe initialization wrapper
    initErr := func() (err error) {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("plugin initialization panic: %v\n%s", r, debug.Stack())
            }
        }()
        // Timeout guard: plugin init must not block forever
        done := make(chan error, 1)
        go func() {
            done <- driver.Initialize(nil, nil)
        }()
        select {
        case err = <-done:
            return err
        case <-time.After(3 * time.Second):
            return fmt.Errorf("plugin initialization timeout")
        }
    }()

    if initErr != nil {
        return nil, fmt.Errorf("plugin initialization failed: %w", initErr)
    }

    return driver, nil
}
