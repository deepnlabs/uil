#!/usr/bin/env bash
set -e

echo "=> Generating UIL Mesh scaffolding..."

BASE_DIR="internal/mesh"
mkdir -p "$BASE_DIR"

########################################
# identity.go
########################################
cat > "$BASE_DIR/identity.go" << 'EOF'
package mesh

import (
    "crypto/rand"
    "encoding/hex"
    "os"
)

type NodeIdentity struct {
    ID   string
    Addr string
}

func LoadOrCreateIdentity() (*NodeIdentity, error) {
    path := "/var/lib/uild/node_id"

    if data, err := os.ReadFile(path); err == nil {
        return &NodeIdentity{ID: string(data)}, nil
    }

    // Generate random 32-byte ID
    buf := make([]byte, 32)
    rand.Read(buf)
    id := hex.EncodeToString(buf)

    os.WriteFile(path, []byte(id), 0644)
    return &NodeIdentity{ID: id}, nil
}
EOF

########################################
# gossip.go
########################################
cat > "$BASE_DIR/gossip.go" << 'EOF'
package mesh

import (
    "log"
    "net"
)

func ListenUDP(port int, handler func([]byte, *net.UDPAddr)) {
    addr := net.UDPAddr{Port: port, IP: net.IPv4zero}
    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        log.Fatalf("mesh: failed to listen on UDP %d: %v", port, err)
    }

    log.Printf("mesh: listening on UDP :%d", port)

    go func() {
        buf := make([]byte, 2048)
        for {
            n, remote, err := conn.ReadFromUDP(buf)
            if err != nil {
                continue
            }
            handler(buf[:n], remote)
        }
    }()
}
EOF

########################################
# heartbeat.go
########################################
cat > "$BASE_DIR/heartbeat.go" << 'EOF'
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
EOF

########################################
# membership.go
########################################
cat > "$BASE_DIR/membership.go" << 'EOF'
package mesh

import (
    "encoding/json"
    "log"
    "net"
    "sync"
    "time"
)

type Peer struct {
    ID       string
    Addr     string
    LastSeen time.Time
    Thermal  float64
    Load     float64
}

type PeerTable struct {
    mu    sync.Mutex
    peers map[string]*Peer
}

func NewPeerTable() *PeerTable {
    return &PeerTable{
        peers: make(map[string]*Peer),
    }
}

func (pt *PeerTable) UpdateFromHeartbeat(hb Heartbeat, addr *net.UDPAddr) {
    pt.mu.Lock()
    defer pt.mu.Unlock()

    p, ok := pt.peers[hb.NodeID]
    if !ok {
        p = &Peer{ID: hb.NodeID, Addr: addr.String()}
        pt.peers[hb.NodeID] = p
    }

    p.LastSeen = time.Now()
    p.Thermal = hb.Thermal
    p.Load = hb.Load
}

func (pt *PeerTable) Broadcast(data []byte, port int) {
    pt.mu.Lock()
    defer pt.mu.Unlock()

    for _, peer := range pt.peers {
        udpAddr, err := net.ResolveUDPAddr("udp", peer.Addr)
        if err != nil {
            continue
        }
        conn, err := net.DialUDP("udp", nil, udpAddr)
        if err != nil {
            continue
        }
        conn.Write(data)
        conn.Close()
    }
}

func (pt *PeerTable) HandleIncoming(data []byte, addr *net.UDPAddr) {
    var hb Heartbeat
    if err := json.Unmarshal(data, &hb); err != nil {
        log.Printf("mesh: invalid heartbeat: %v", err)
        return
    }
    pt.UpdateFromHeartbeat(hb, addr)
}
EOF

########################################
# mesh.go
########################################
cat > "$BASE_DIR/mesh.go" << 'EOF'
package mesh

import (
    "log"
    "net"
)

type Mesh struct {
    Identity *NodeIdentity
    Peers    *PeerTable
    Port     int
}

func NewMesh(port int) (*Mesh, error) {
    id, err := LoadOrCreateIdentity()
    if err != nil {
        return nil, err
    }

    return &Mesh{
        Identity: id,
        Peers:    NewPeerTable(),
        Port:     port,
    }, nil
}

func (m *Mesh) Start() {
    ListenUDP(m.Port, func(data []byte, addr *net.UDPAddr) {
        m.Peers.HandleIncoming(data, addr)
    })

    StartHeartbeat(m.Identity, m.Peers, m.Port)

    log.Printf("mesh: started with node ID %s", m.Identity.ID)
}
EOF

echo "=> Mesh scaffolding generated in internal/mesh/"
