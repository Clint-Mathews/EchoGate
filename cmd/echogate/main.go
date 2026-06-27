package main

import (
	"fmt"

	utils "github.com/Clint-Mathews/EchoGate/internal/config"
	proxy "github.com/Clint-Mathews/EchoGate/internal/proxy"
)

func main() {
	fmt.Println("Echo Gate!")
	// Load ENV
	utils.LoadEnv()

	// Proxy Server
	proxy.ProxyServer()
}
