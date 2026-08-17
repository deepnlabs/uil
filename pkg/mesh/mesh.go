package mesh

import (
    "context"
    "log"

    "github.com/deepnlabs/uil/pkg/uil"
)

type MeshRuntime struct {
    Mesh *NodeMesh
}

func StartMesh(port int, handler func(env uil.UILEnvelope)) (*MeshRuntime, error) {
    nm, err := NewNodeMesh(port, handler)
    if err != nil {
        return nil, err
    }

    ctx := context.Background()
    nm.Start(ctx)

    log.Printf("mesh: started with node ID %s on UDP port %d", nm.nodeID, port)

    return &MeshRuntime{Mesh: nm}, nil
}
