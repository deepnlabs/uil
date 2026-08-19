package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/deepnlabs/uil/pkg/mesh"
)

func meshPeers() {
    // Query the UIL daemon's mesh endpoint
    resp, err := http.Get("http://127.0.0.1:9410/mesh/peers")
    if err != nil {
        fmt.Println("Error: unable to reach UIL daemon:", err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        fmt.Printf("Error: daemon returned status %d\n", resp.StatusCode)
        os.Exit(1)
    }

    var peers []mesh.PeerInfo
    if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
        fmt.Println("Error decoding mesh peer list:", err)
        os.Exit(1)
    }

    if len(peers) == 0 {
        fmt.Println("No active mesh peers detected.")
        return
    }

    fmt.Println("Active Mesh Peers:")
    fmt.Println("--------------------------------------------------------------")
    fmt.Printf("%-16s %-16s %-8s %-24s\n", "NODE ID", "ADDRESS", "PORT", "LAST SEEN")
    fmt.Println("--------------------------------------------------------------")

    for _, p := range peers {
        fmt.Printf(
            "%-16s %-16s %-8d %-24s\n",
            p.NodeID,
            p.Address,
            p.Port,
            p.LastSeen.Format(time.RFC3339),
        )
    }
}
