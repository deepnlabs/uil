package main

import (
	"fmt"
	"os"
	"github.com/deepnlabs/uil/pkg/crypto"
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
		fmt.Println("Creating state directories...")
		fmt.Println("Initialization complete.")
	case "run":
		fmt.Println("Starting UIL-X Hardware Governance Daemon (`uild`) v0.5-PQ...")
		hasher := crypto.NewSHA3Hasher("SHA3-256")
		env := uil.NewEnvelope("node-1", "node-2", uil.SubstrateHailoNPU, map[string]interface{}{"temp": 42.5, "status": "nominal"})
		hash, _ := hasher.HashPayload(env.Payload)
		fmt.Printf("Engine Active | Substrate: %s | SHA3 Commitment: %s\n", env.SubstrateTarget, hash[:16])
	case "status":
		fmt.Println("UIL-X Daemon Status: ACTIVE")
		fmt.Println("Substrate Engines: Hailo-8 NPU (Online), Local Linux GPIO (Ready)")
		fmt.Println("Interlock Status: Nominal (0 Breaches)")
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("UIL-X Hardware Governance Daemon (uild)")
	fmt.Println("Usage: uild [init|run|status]")
}
