package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/woolawin/catalogue/internal/ext"
)

//go:embed version.txt
var Version string

func main() {
	slog.Info("booting catalogue-apt-server", "version", Version)
	api := ext.NewAPI("/")
	config, err := api.Host.GetConfig()
	if err != nil {
		slog.Error("failed to get config", "error", err)
		os.Exit(1)
	}

	system, err := api.Host.GetSystem()
	if err != nil {
		slog.Error("failed to get system", "error", err)
		os.Exit(1)
	}

	server := NewHTTPServer(config, system, api)

	err = server.start()
	if err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, release := context.WithTimeout(context.Background(), 10*time.Second)
	defer release()

	server.Shutdown(ctx)
}
