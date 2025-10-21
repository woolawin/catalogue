package main

import (
	reg "github.com/woolawin/catalogue/internal/registry"
)

func main() {
	registry := reg.NewDiskRegistry()
	server := NewHTTPServer(registry)
	server.start()
}
