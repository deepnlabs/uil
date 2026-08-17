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
    "github.com/deepnlabs/uil/pkg/mesh"
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
    	// Load persistent node identity FIRST
    	nodeID, _ := mesh.LoadOrCreateIdentity()
    	fmt.Printf("Starting UIL-X Hardware Governance Daemon (`uild`) v0.8-alpha on [%s]...\n", nodeID)

    	ctx, cancel := context.WithCancel(context.Background())
    	defer cancel()

	    // 1. Local Unix Socket Bridge
	    socketBridge, _ := bridge.NewBridgeEmitter()
	    if socketBridge != nil {
	        defer socketBridge.Close()
	        fmt.Println("  └─ [IPC BRIDGE] Listening on Unix socket /tmp/uild.sock")
	    }

	    // 2. Mesh Network UDP Gossip
	meshNetwork, err := mesh.NewNodeMesh(9090, func(remoteEnv uil.UILEnvelope) {
	    if remoteEnv.ImportanceScore >= 0.9 {
	        fmt.Printf("\n🚨 [REMOTE MESH ALERT] High-priority interlock breach from [%s]! SHA3: %s\n",
	            remoteEnv.SourceNode, remoteEnv.ProofCommitment[:12])
	    }
	})

	if err != nil {
	    fmt.Printf("  └─ [MESH NETWORK] FAILED to start: %v\n", err)
	    meshNetwork = nil
	} else {
	    defer meshNetwork.Close()
	    meshNetwork.Start(ctx)
	    fmt.Println("  └─ [MESH NETWORK] UDP Peer Discovery Active on Port 9090")
	}


	// 3. Register Hardware Substrates
    	var drivers []substrate.SubstrateDriver
    	nativeDriver := substrate.NewLinuxHardwareDriver(80.0)
    	if err := nativeDriver.Initialize(ctx, nil); err == nil {
    	    drivers = append(drivers, nativeDriver)
    	    fmt.Printf("  └─ [NATIVE] Registered %s (%s)\n", nativeDriver.Name(), nativeDriver.ID())
    	}

    	// Load Dynamic Plugins
    	files, _ := filepath.Glob("./plugins/*.so")
    	for _, file := range files {
    	    drv, err := substrate.LoadPlugin(file)
    	    if err == nil && drv.Initialize(ctx, nil) == nil {
    	        drivers = append(drivers, drv)
    	        fmt.Printf("  └─ [PLUGIN] Loaded %s (%s)\n", drv.Name(), drv.ID())
    	    }
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

        	    // Local IPC Broadcast
        	    if socketBridge != nil {
        	        socketBridge.BroadcastBreach(bridge.SafetyEvent{
        	            InterlockID: "local-aggregate",
        	            Breach:      anyBreach,
        	            Timestamp:   time.Now().Unix(),
        	            Metrics:     liveMetrics,
        	        })
        	    }

        	    // Compute SHA3 State Commitment
        	    env := uil.NewEnvelope(nodeID, "broadcast", uil.SubstrateAuto, map[string]interface{}{
        	        "cpu_temp": liveMetrics["cpu_temp_celsius"],
        	        "breach":   anyBreach,
        	    })
        	    hash, _ := hasher.HashPayload(env.Payload)
        	    env.ProofCommitment = hash

        	    if anyBreach {
        	        env.ImportanceScore = 1.0
        	    } else {
        	        env.ImportanceScore = 0.1
        	    }

		    // Broadcast state over P2P mesh
		    if meshNetwork != nil {
		        _ = meshNetwork.BroadcastGossip(env)
		        peers := meshNetwork.GetActivePeers()
    			fmt.Printf("[%s] Tick | Temp: %.1f°C | Mesh Peers: %d | SHA3: %s\n",
        			time.Now().Format("15:04:05"), liveMetrics["cpu_temp_celsius"], len(peers), hash[:12])
		    } else {
    			fmt.Printf("[%s] Tick | Temp: %.1f°C | Mesh Peers: N/A | SHA3: %s\n",
        		time.Now().Format("15:04:05"), liveMetrics["cpu_temp_celsius"], hash[:12])
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
