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
    case "update":
        // call update logic
    default:
        fmt.Println("Unknown command:", os.Args[1])
        os.Exit(1)
    }
}
