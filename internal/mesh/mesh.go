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
