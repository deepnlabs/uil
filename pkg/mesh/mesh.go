package mesh

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "sync"
    "time"

    "github.com/deepnlabs/uil/pkg/uil"
)

type PeerInfo struct {
    NodeID   string    `json:"node_id"`
    Address  string    `json:"address"`
    Port     int       `json:"port"`
    LastSeen time.Time `json:"last_seen"`
    CPUTemp  float64   `json:"cpu_temp"`
}

type NodeMesh struct {
    nodeID     string
    port       int
    conn       *net.UDPConn
    peers      map[string]*PeerInfo
    mu         sync.RWMutex
    onEnvelope func(env uil.UILEnvelope)

    configuredPeers []string // static peer list from config
}

type MeshRuntime struct {
    Mesh *NodeMesh
}

func StartMesh(port int, peers []string, handler func(env uil.UILEnvelope)) (*MeshRuntime, error) {
    nm, err := NewNodeMesh(port, peers, handler)
    if err != nil {
        return nil, err
    }

    ctx := context.Background()
    nm.Start(ctx)

    log.Printf("mesh: started with node ID %s on UDP port %d", nm.nodeID, port)
    return &MeshRuntime{Mesh: nm}, nil
}

func NewNodeMesh(port int, peers []string, handler func(env uil.UILEnvelope)) (*NodeMesh, error) {
    nodeID, err := LoadOrCreateIdentity()
    if err != nil {
        return nil, fmt.Errorf("failed to load node identity: %w", err)
    }

    if port <= 0 {
        port = 9091
    }

    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
    if err != nil {
        return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        return nil, fmt.Errorf("failed to listen on UDP port %d: %w", port, err)
    }

    mesh := &NodeMesh{
        nodeID:         nodeID,
        port:           port,
        conn:           conn,
        peers:          make(map[string]*PeerInfo),
        onEnvelope:     handler,
        configuredPeers: peers,
    }

    return mesh, nil
}

func (m *NodeMesh) Start(ctx context.Context) {
    go m.listenLoop(ctx)
    go m.heartbeatLoop(ctx)
    go m.dialPeers(ctx)
}

func (m *NodeMesh) listenLoop(ctx context.Context) {
    buf := make([]byte, 4096)

    for {
        select {
        case <-ctx.Done():
            return
        default:
            m.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
            n, remoteAddr, err := m.conn.ReadFromUDP(buf)
            if err != nil {
                continue
            }

            var env uil.UILEnvelope
            if err := json.Unmarshal(buf[:n], &env); err != nil {
                continue
            }

            if env.SourceNode == m.nodeID {
                continue
            }

            m.updatePeer(env, remoteAddr)
            if m.onEnvelope != nil {
                m.onEnvelope(env)
            }
        }
    }
}

func (m *NodeMesh) updatePeer(env uil.UILEnvelope, addr *net.UDPAddr) {
    m.mu.Lock()
    defer m.mu.Unlock()

    temp, _ := env.Payload["cpu_temp"].(float64)
    port, _ := env.Payload["mesh_port"].(float64)

    m.peers[env.SourceNode] = &PeerInfo{
        NodeID:   env.SourceNode,
        Address:  addr.IP.String(),
        Port:     int(port),
        LastSeen: time.Now(),
        CPUTemp:  temp,
    }
}

func (m *NodeMesh) dialPeers(ctx context.Context) {
    for _, addr := range m.configuredPeers {
        go func(addr string) {
            for {
                conn, err := net.Dial("udp", addr)
                if err == nil {
                    m.sendHello(conn)
                    m.sendHeartbeat(conn)
                }
                time.Sleep(5 * time.Second)
            }
        }(addr)
    }
}

func (m *NodeMesh) sendHello(conn net.Conn) {
    env := uil.NewEnvelope(m.nodeID, "hello", uil.SubstrateAuto, map[string]interface{}{
        "mesh_port": m.port,
    })
    data, _ := json.Marshal(env)
    conn.Write(data)
}

func (m *NodeMesh) sendHeartbeat(conn net.Conn) {
    env := uil.NewEnvelope(m.nodeID, "heartbeat", uil.SubstrateAuto, map[string]interface{}{
        "mesh_port": m.port,
    })
    data, _ := json.Marshal(env)
    conn.Write(data)
}

func (m *NodeMesh) BroadcastGossip(env *uil.UILEnvelope) error {
    if env == nil {
        return nil
    }

    data, err := json.Marshal(env)
    if err != nil {
        return err
    }

    // Broadcast to LAN
    dst, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", m.port))
    if err != nil {
        return err
    }

    _, err = m.conn.WriteToUDP(data, dst)
    return err
}

func (m *NodeMesh) heartbeatLoop(ctx context.Context) {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.mu.Lock()
            now := time.Now()
            for id, p := range m.peers {
                if now.Sub(p.LastSeen) > 10*time.Second {
                    delete(m.peers, id)
                }
            }
            m.mu.Unlock()
        }
    }
}

func (m *NodeMesh) GetActivePeers() []PeerInfo {
    m.mu.RLock()
    defer m.mu.RUnlock()

    active := make([]PeerInfo, 0, len(m.peers))
    for _, p := range m.peers {
        active = append(active, *p)
    }
    return active
}

func (m *NodeMesh) Close() {
    if m.conn != nil {
        m.conn.Close()
    }
}
