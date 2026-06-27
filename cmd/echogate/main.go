package main

import (
	"fmt"
	"log"

	utils "github.com/Clint-Mathews/EchoGate/internal/config"
	proxy "github.com/Clint-Mathews/EchoGate/internal/proxy"
)

func main() {
	fmt.Println("Echo Gate!")
	// Load ENV
	utils.LoadEnv()

	// Proxy Server
	if err := proxy.ProxyServer(); err != nil {
		log.Fatal("Failed to run server")
	}
}
