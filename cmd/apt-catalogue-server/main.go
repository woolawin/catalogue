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
	server, err := NewServer()
	if err != nil {
		fmt.Println(err.Error())
	}
	server.start()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, release := context.WithTimeout(context.Background(), 10*time.Second)
	defer release()

	server.shutdown(ctx)
}

type Server struct {
	socket *daemon.Server
	http   *HTTPServer
}

func NewServer() (Server, error) {
	registry := reg.NewRegistry()
	api := ext.NewAPI("/")
	system, err := api.Host.GetSystem()
	if err != nil {
		return Server{}, err
	}

	server := Server{
		socket: daemon.NewServer(system, api, registry),
		http:   NewHTTPServer(registry),
	}

	return server, nil
}

func (server *Server) start() error {
	err := server.socket.Start()
	if err != nil {
		return err
	}

	err = server.http.start()
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) shutdown(ctx context.Context) {
	server.socket.Shutdown()
	server.http.Shutdown(ctx)
}
