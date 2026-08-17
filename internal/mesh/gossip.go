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
