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
