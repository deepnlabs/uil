// v0.9 config loader (stub)
package core

type Config struct {
    NodeID string `json:"node_id"`
}

func LoadConfig() Config {
    // TODO: load config from file
    return Config{}
}
