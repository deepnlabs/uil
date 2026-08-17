package mesh

import (
    "encoding/json"
    "log"
    "net"
    "time"
)

type Heartbeat struct {
    NodeID    string  `json:"node_id"`
    Thermal   float64 `json:"thermal"`
    Load      float64 `json:"load"`
    Timestamp int64   `json:"ts"`
}

func StartHeartbeat(identity *NodeIdentity, peers *PeerTable, port int) {
    go func() {
        for {
            hb := Heartbeat{
                NodeID:    identity.ID,
                Thermal:   ReadThermal(),
                Load:      ReadLoad(),
                Timestamp: time.Now().Unix(),
            }

            data, _ := json.Marshal(hb)
            peers.Broadcast(data, port)

            time.Sleep(1 * time.Second)
        }
    }()
}

// Stub telemetry
func ReadThermal() float64 { return 42.0 }
func ReadLoad() float64    { return 0.15 }
