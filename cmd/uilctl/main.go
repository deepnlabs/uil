package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: uilctl <command>")
        os.Exit(1)
    }

    switch os.Args[1] {

    case "mesh":
        if len(os.Args) < 3 {
            fmt.Println("Usage: uilctl mesh peers")
            os.Exit(1)
        }
        switch os.Args[2] {
        case "peers":
            meshPeers()
        default:
            fmt.Println("Unknown mesh command:", os.Args[2])
            os.Exit(1)
        }

    case "update":
        // update logic goes here

    default:
        fmt.Println("Unknown command:", os.Args[1])
        os.Exit(1)
    }
}
