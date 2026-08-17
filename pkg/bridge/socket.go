package bridge

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

const SocketPath = "/tmp/uild.sock"

type SafetyEvent struct {
	InterlockID string             `json:"interlock_id"`
	Breach      bool               `json:"breach"`
	Timestamp   int64              `json:"timestamp"`
	Metrics     map[string]float64 `json:"metrics"`
}

type BridgeEmitter struct {
	listener net.Listener
	conns    map[net.Conn]bool
	mu       sync.Mutex
}

func NewBridgeEmitter() (*BridgeEmitter, error) {
	_ = os.Remove(SocketPath)

	l, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to bind socket at %s: %w", SocketPath, err)
	}

	_ = os.Chmod(SocketPath, 0666)

	emitter := &BridgeEmitter{
		listener: l,
		conns:    make(map[net.Conn]bool),
	}

	// Accept incoming subscriber connections in background goroutine
	go emitter.acceptLoop()

	return emitter, nil
}

func (b *BridgeEmitter) acceptLoop() {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			return
		}

		b.mu.Lock()
		b.conns[conn] = true
		b.mu.Unlock()
	}
}

func (b *BridgeEmitter) BroadcastBreach(event SafetyEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	payload = append(payload, '\n')

	b.mu.Lock()
	defer b.mu.Unlock()

	for conn := range b.conns {
		_, err := conn.Write(payload)
		if err != nil {
			conn.Close()
			delete(b.conns, conn)
		}
	}
}

func (b *BridgeEmitter) Close() {
	if b.listener != nil {
		b.listener.Close()
		_ = os.Remove(SocketPath)
	}
	b.mu.Lock()
	for conn := range b.conns {
		conn.Close()
	}
	b.mu.Unlock()
}
