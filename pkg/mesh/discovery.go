package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/deepnlabs/uil/pkg/uil"
)

const (
	DefaultMeshPort = 9090
	BroadcastAddr   = "255.255.255.255:9090"
)

type PeerInfo struct {
	NodeID   string    `json:"node_id"`
	Address  string    `json:"address"`
	LastSeen time.Time `json:"last_seen"`
	CPUTemp  float64   `json:"cpu_temp"`
}

type NodeMesh struct {
	nodeID    string
	port      int
	conn      *net.UDPConn
	peers     map[string]*PeerInfo
	mu        sync.RWMutex
	onEnvelope func(env uil.UILEnvelope)
}

// NewNodeMesh initializes the UDP socket for local subnet broadcast/multicast
func NewNodeMesh(nodeID string, port int, handler func(env uil.UILEnvelope)) (*NodeMesh, error) {
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

	mesh := &NodeMesh{
		nodeID:     nodeID,
		port:       port,
		conn:       conn,
		peers:      make(map[string]*PeerInfo),
		onEnvelope: handler,
	}

	return mesh, nil
}

// Start Listening and Heartbeat Loops
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

			// Ignore self-broadcasts
			if env.SourceNode == m.nodeID {
				continue
			}

			// Update Peer Registry
			m.mu.Lock()
			temp, _ := env.Payload["cpu_temp"].(float64)
			m.peers[env.SourceNode] = &PeerInfo{
				NodeID:   env.SourceNode,
				Address:  remoteAddr.IP.String(),
				LastSeen: time.Now(),
				CPUTemp:  temp,
			}
			m.mu.Unlock()

			// Trigger high-priority envelope callback
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
			// Prune stale peers (>10s inactive)
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

// BroadcastGossip sends a UILEnvelope pointer to all subnet peers
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

// GetActivePeers returns a snapshot of discovered nodes
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
