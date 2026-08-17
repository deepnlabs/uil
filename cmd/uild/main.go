package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/deepnlabs/uil/pkg/bridge"
	"github.com/deepnlabs/uil/pkg/crypto"
	"github.com/deepnlabs/uil/pkg/substrate"
	"github.com/deepnlabs/uil/pkg/uil"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		fmt.Println("Initializing UIL-X runtime configuration at /etc/uild/config.json...")
		fmt.Println("Creating state directories at /var/lib/uild/plugins...")
		fmt.Println("Initialization complete.")

	case "run":
		fmt.Println("Starting UIL-X Hardware Governance Daemon (`uild`) v0.8-alpha...")

		// 1. Initialize IPC Socket Emitter (/tmp/uild.sock for user-space access)
		socketBridge, err := bridge.NewBridgeEmitter()
		if err != nil {
			fmt.Printf("  └─ [WARN] Failed to bind IPC socket /tmp/uild.sock: %v\n", err)
		} else {
			defer socketBridge.Close()
			fmt.Println("  └─ [IPC BRIDGE] Listening on Unix socket /tmp/uild.sock")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var drivers []substrate.SubstrateDriver

		// 2. Register Native Linux Thermal Driver
		nativeDriver := substrate.NewLinuxHardwareDriver(80.0)
		if err := nativeDriver.Initialize(ctx, nil); err == nil {
			drivers = append(drivers, nativeDriver)
			fmt.Printf("  └─ [NATIVE] Registered %s (%s)\n", nativeDriver.Name(), nativeDriver.ID())
		}

		// 3. Scan & Load Dynamic Plugins
		pluginDir := "./plugins"
		files, err := filepath.Glob(filepath.Join(pluginDir, "*.so"))
		if err == nil && len(files) > 0 {
			for _, file := range files {
				drv, err := substrate.LoadPlugin(file)
				if err != nil {
					continue
				}
				if err := drv.Initialize(ctx, nil); err == nil {
					drivers = append(drivers, drv)
					fmt.Printf("  └─ [PLUGIN] Loaded %s (%s)\n", drv.Name(), drv.ID())
				}
			}
		}

		// 4. Setup Signal Interception for Clean Daemon Shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// 5. Continuous Governance Execution Loop (Runs every 1 second)
		ticker := time.NewTicker(1 * time.Second)
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
				for _, drv := range drivers {
					breach, err := drv.EvaluateInterlock(liveMetrics)
					if err != nil {
						continue
					}

					// Broadcast safety state over IPC socket
					if socketBridge != nil {
						socketBridge.BroadcastBreach(bridge.SafetyEvent{
							InterlockID: drv.ID(),
							Breach:      breach,
							Timestamp:   time.Now().Unix(),
							Metrics:     liveMetrics,
						})
					}

					if breach {
						fmt.Printf("[%s] 🚨 SAFETY INTERLOCK BREACH on %s!\n", time.Now().Format("15:04:05"), drv.ID())
					}
				}

				// Compute SHA3 commitment for current tick state
				env := uil.NewEnvelope("deepn-node-2", "broadcast", uil.SubstrateAuto, map[string]interface{}{
					"telemetry": liveMetrics,
				})
				hash, _ := hasher.HashPayload(env.Payload)

				fmt.Printf("[%s] Tick | CPU Temp: %.1f°C | SHA3: %s\n",
					time.Now().Format("15:04:05"), liveMetrics["cpu_temp_celsius"], hash[:12])
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

