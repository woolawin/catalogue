package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/woolawin/catalogue/internal/daemon"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

func main() {

	registry := reg.NewRegistry()
	api := ext.NewAPI("/")
	system, err := api.Host.GetSystem()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	server := daemon.NewServer(system, api, registry)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	_, release := context.WithTimeout(context.Background(), 10*time.Second)
	defer release()

	server.Shutdown()
}
