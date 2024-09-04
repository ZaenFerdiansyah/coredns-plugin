package main

import (
    "github.com/coredns/coredns/coremain"
    _ "github.com/coredns/coredns/plugin" // Import other necessary plugins here
)

func main() {
    coremain.Run()
}
