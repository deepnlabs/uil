// v0.9 plugin ABI (stub)
package plugin

type Plugin interface {
    Init() error
    Tick() error
    Shutdown() error
}
