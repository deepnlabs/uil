package mesh

import (
    "context"
    "encoding/json"
    "fmt"
    "net"
    "sync"
    "syscall"
    "time"

    "github.com/deepnlabs/uil/pkg/uil"
)

const (
    DefaultMeshPort = 9091
    BroadcastAddr   = "255.255.255.255:9090"
)

type PeerInfo struct {
    NodeID   string    `json:"node_id"`
    Address  string    `json:"address"`
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
}

func NewNodeMesh(port int, handler func(env uil.UILEnvelope)) (*NodeMesh, error) {
    nodeID, err := LoadOrCreateIdentity()
    if err != nil {
        return nil, fmt.Errorf("failed to load node identity: %w", err)
    }

    if port <= 0 {
        port = DefaultMeshPort
    }

    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
    if err != nil {
        return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        return nil, fmt.Errorf("failed to listen on UDP port %d: %w", port, err)
    }

    // Enable UDP broadcast (critical fix)
    rawConn, err := conn.SyscallConn()
    if err == nil {
        rawConn.Control(func(fd uintptr) {
            syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
        })
    }

    mesh := &NodeMesh{
        nodeID:     nodeID,
        port:       port,
        conn:       conn,
        peers:      make(map[string]*PeerInfo),
        onEnvelope: handler,
    }

    return mesh, nil
}

func (m *NodeMesh) Start(ctx context.Context) {
    go m.listenLoop(ctx)
    go m.heartbeatLoop(ctx)
}

func (m *NodeMesh) listenLoop(ctx context.Context) {
    buf := make([]byte, 4096)

    for {
        select {
        case <-ctx.Done():
            return
        default:
            _ = m.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
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

            m.mu.Lock()
            temp, _ := env.Payload["cpu_temp"].(float64)
            m.peers[env.SourceNode] = &PeerInfo{
                NodeID:   env.SourceNode,
                Address:  remoteAddr.IP.String(),
                LastSeen: time.Now(),
                CPUTemp:  temp,
            }
            m.mu.Unlock()

            if m.onEnvelope != nil {
                m.onEnvelope(env)
            }
        }
    }
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

func (m *NodeMesh) BroadcastGossip(env *uil.UILEnvelope) error {
    if env == nil {
        return nil
    }

    data, err := json.Marshal(env)
    if err != nil {
        return err
    }

    dst, err := net.ResolveUDPAddr("udp", BroadcastAddr)
    if err != nil {
        return err
    }

    _, err = m.conn.WriteToUDP(data, dst)
    return err
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
