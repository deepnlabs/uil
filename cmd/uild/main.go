package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    "time"

    "github.com/deepnlabs/uil/pkg/bridge"
    "github.com/deepnlabs/uil/pkg/crypto"
    "github.com/deepnlabs/uil/pkg/mesh"
    "github.com/deepnlabs/uil/pkg/substrate"
    "github.com/deepnlabs/uil/pkg/uil"
)

type MeshConfig struct {
    Enabled          bool     `json:"enabled"`
    Port             int      `json:"port"`
    BroadcastAddress string   `json:"broadcast_address"`
    APIPort          int      `json:"api_port"`
    Peers            []string `json:"peers"`
}

type Config struct {
    NodeID              string     `json:"node_id"`
    LogLevel            string     `json:"log_level"`
    ThermalLimitCelsius float64    `json:"thermal_limit_celsius"`
    Mesh                MeshConfig `json:"mesh"`
    PluginsDir          string     `json:"plugins_dir"`
    PluginsEnabled      bool       `json:"plugins_enabled"`
}

func loadConfig() Config {
    const configPath = "/etc/uil/config.json"

    f, err := os.Open(configPath)
    if err != nil {
        fmt.Printf("⚠️  Config file not found at %s, using defaults\n", configPath)
        return Config{
            NodeID:              "uil-node",
            ThermalLimitCelsius: 80.0,
            Mesh: MeshConfig{
                Enabled:          false,
                Port:             9091,
                BroadcastAddress: "255.255.255.255:9091",
                APIPort:          9410,
                Peers:            []string{},
            },
            PluginsDir:     "/var/lib/uild/plugins",
            PluginsEnabled: false,
        }
    }
    defer f.Close()

    var cfg Config
    if err := json.NewDecoder(f).Decode(&cfg); err != nil {
        fmt.Printf("⚠️  Config decode failed: %v — using defaults\n", err)
        return Config{
            NodeID:              "uil-node",
            ThermalLimitCelsius: 80.0,
            Mesh: MeshConfig{
                Enabled:          false,
                Port:             9091,
                BroadcastAddress: "255.255.255.255:9091",
                APIPort:          9410,
                Peers:            []string{},
            },
            PluginsDir:     "/var/lib/uild/plugins",
            PluginsEnabled: false,
        }
    }

    if cfg.Mesh.Port == 0 {
        cfg.Mesh.Port = 9091
    }
    if cfg.Mesh.APIPort == 0 {
        cfg.Mesh.APIPort = 9410
    }
    if cfg.PluginsDir == "" {
        cfg.PluginsDir = "/var/lib/uild/plugins"
    }

    return cfg
}

func main() {
    if len(os.Args) < 2 {
        printHelp()
        os.Exit(1)
    }

    command := os.Args[1]

    switch command {
    case "init":
        fmt.Println("Initializing UIL-X runtime configuration at /etc/uil/config.json...")
        fmt.Println("Creating state directories at /var/lib/uild/plugins...")
        fmt.Println("Initialization complete.")

    case "run":
        cfg := loadConfig()

        // Load persistent node identity (ID + keypair)
        id, err := mesh.LoadOrCreateIdentity()
        if err != nil {
            fmt.Printf("Failed to load node identity: %v\n", err)
            os.Exit(1)
        }
        nodeID := id.NodeID
        fmt.Printf("Starting UIL-X Hardware Governance Daemon (`uild`) v0.8-alpha on [%s]...\n", nodeID)

        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        // 1. Local Unix Socket Bridge (secure)
        socketPath := "/run/uild/uild.sock"
        var socketBridge *bridge.BridgeEmitter

        if err := os.MkdirAll("/run/uild", 0750); err != nil {
            fmt.Printf("  └─ [IPC BRIDGE] Failed to create /run/uild: %v\n", err)
        } else {
            if _, err := os.Stat(socketPath); err == nil {
                _ = os.Remove(socketPath)
            }

            socketBridge, err = bridge.NewBridgeEmitter(socketPath)
            if err != nil {
                fmt.Printf("  └─ [IPC BRIDGE] Failed to start: %v\n", err)
                socketBridge = nil
            } else {
                defer socketBridge.Close()
                _ = os.Chmod(socketPath, 0600)
                fmt.Printf("  └─ [IPC BRIDGE] Listening on Unix socket %s\n", socketPath)
            }
        }

        // 2. Mesh Network UDP Gossip
        var meshNetwork *mesh.NodeMesh
        if cfg.Mesh.Enabled {
            nm, err := mesh.NewNodeMesh(cfg.Mesh.Port, cfg.Mesh.Peers, func(remoteEnv uil.UILEnvelope) {
                if remoteEnv.ImportanceScore >= 0.9 {
                    proof := remoteEnv.ProofCommitment
                    if len(proof) < 12 {
                        fmt.Printf("\n🚨 [REMOTE MESH ALERT] Malformed proof from [%s] — dropping packet\n",
                            remoteEnv.SourceNode)
                        return
                    }
                    fmt.Printf("\n🚨 [REMOTE MESH ALERT] High-priority interlock breach from [%s]! SHA3: %s\n",
                        remoteEnv.SourceNode, proof[:12])
                }
            })
            if err != nil {
                fmt.Printf("  └─ [MESH NETWORK] FAILED to start: %v\n", err)
                meshNetwork = nil
            } else {
                meshNetwork = nm
                meshNetwork.Start(ctx)
                fmt.Printf("  └─ [MESH NETWORK] UDP Peer Discovery Active on Port %d\n", cfg.Mesh.Port)

                go func() {
                    http.HandleFunc("/mesh/peers", func(w http.ResponseWriter, r *http.Request) {
                        peers := meshNetwork.GetActivePeers()
                        w.Header().Set("Content-Type", "application/json")
                        _ = json.NewEncoder(w).Encode(peers)
                    })
                    addr := fmt.Sprintf("127.0.0.1:%d", cfg.Mesh.APIPort)
                    _ = http.ListenAndServe(addr, nil)
                }()
            }
        }

        // 3. Register Hardware Substrates
        var drivers []substrate.SubstrateDriver
        nativeDriver := substrate.NewLinuxHardwareDriver(cfg.ThermalLimitCelsius)
        if err := nativeDriver.Initialize(ctx, nil); err == nil {
            drivers = append(drivers, nativeDriver)
            fmt.Printf("  └─ [NATIVE] Registered %s (%s)\n", nativeDriver.Name(), nativeDriver.ID())
        } else {
            fmt.Printf("  └─ [NATIVE] Failed to initialize native driver: %v\n", err)
        }

        // Load Dynamic Plugins (opt-in)
        if cfg.PluginsEnabled {
            files, _ := filepath.Glob(cfg.PluginsDir + "/*.so")
            for _, file := range files {
                drv, err := substrate.LoadPlugin(file)
                if err == nil && drv.Initialize(ctx, nil) == nil {
                    drivers = append(drivers, drv)
                    fmt.Printf("  └─ [PLUGIN] Loaded %s (%s)\n", drv.Name(), drv.ID())
                } else if err != nil {
                    fmt.Printf("  └─ [PLUGIN] Failed to load plugin %s: %v\n", file, err)
                }
            }
        } else {
            fmt.Println("  └─ [PLUGIN] Dynamic plugins disabled by config")
        }

        // 4. Signal Interception
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

        ticker := time.NewTicker(2 * time.Second)
        defer ticker.Stop()

        hasher := crypto.NewSHA3Hasher("SHA3-256")
        fmt.Println("Governance loop active. Press Ctrl+C to stop.")

        for {
            select {
            case <-sigChan:
                fmt.Println("\nShutdown signal received. Cleaning up UIL-X daemon...")
                return

            case <-ticker.C:
                liveMetrics := make(map[string]float64)
                anyBreach := false

                for _, drv := range drivers {
                    breach, _ := drv.EvaluateInterlock(liveMetrics)
                    if breach {
                        anyBreach = true
                        fmt.Printf("[%s] 🚨 SAFETY BREACH on local substrate %s!\n",
                            time.Now().Format("15:04:05"), drv.ID())
                    }
                }

                if socketBridge != nil {
                    socketBridge.BroadcastBreach(bridge.SafetyEvent{
                        InterlockID: "local-aggregate",
                        Breach:      anyBreach,
                        Timestamp:   time.Now().Unix(),
                        Metrics:     liveMetrics,
                    })
                }

                env := uil.NewEnvelope(nodeID, "broadcast", uil.SubstrateAuto, map[string]interface{}{
                    "cpu_temp":  liveMetrics["cpu_temp_celsius"],
                    "breach":    anyBreach,
                    "mesh_port": cfg.Mesh.Port,
                })
                hash, _ := hasher.HashPayload(env.Payload)
                env.ProofCommitment = hash

                if anyBreach {
                    env.ImportanceScore = 1.0
                } else {
                    env.ImportanceScore = 0.1
                }

                shortHash := hash
                if len(shortHash) >= 12 {
                    shortHash = shortHash[:12]
                }

                if meshNetwork != nil {
                    _ = meshNetwork.BroadcastGossip(env)
                    peers := meshNetwork.GetActivePeers()
                    fmt.Printf("[%s] Tick | Temp: %.1f°C | Mesh Peers: %d | SHA3: %s\n",
                        time.Now().Format("15:04:05"), liveMetrics["cpu_temp_celsius"], len(peers), shortHash)
                } else {
                    fmt.Printf("[%s] Tick | Temp: %.1f°C | Mesh Peers: N/A | SHA3: %s\n",
                        time.Now().Format("15:04:05"), liveMetrics["cpu_temp_celsius"], shortHash)
                }
            }
        }

    case "status":
        fmt.Println("UIL-X Daemon Status: ACTIVE")

    default:
        printHelp()
    }
}

func printHelp() {
    fmt.Println("UIL-X Hardware Governance Daemon (uild)")
    fmt.Println("Usage: uild [init|run|status]")
}
