package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

func main() {
	registry := reg.NewRegistry()
	host := ext.NewHost()
	server := NewHTTPServer(host, registry)

	err := server.start()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, release := context.WithTimeout(context.Background(), 10*time.Second)
	defer release()

	server.Shutdown(ctx)
}
