// Example mesh plugin (stub)
package main

import "fmt"

func Init() error {
    fmt.Println("mesh_basic plugin initialized")
    return nil
}

func Tick() error {
    return nil
}

func Shutdown() error {
    return nil
}
