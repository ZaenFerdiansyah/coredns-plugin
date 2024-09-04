package main

import (
    "github.com/coredns/coredns/coremain"
    _ "github.com/ZaenFerdiansyah/coredns-plugin/domainforwarder" // Import other necessary plugins here
)

func main() {
    coremain.Run()
}
