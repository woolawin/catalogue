package main

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

type HTTPServer struct {
	registry reg.Registry
}

func NewHTTPServer(registry reg.Registry) *HTTPServer {
	return &HTTPServer{}
}

func (server *HTTPServer) start() error {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/dists/{distro}/Release", server.Release)
	router.Get("/pool/{file}", server.Pool)
	return nil
}

func (server *HTTPServer) Release(writer http.ResponseWriter, request *http.Request) {
	packages, err := server.registry.ListPackages()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	output := strings.Builder{}

	for _, pkg := range packages {
		config, found, err := server.registry.GetPackageConfig(pkg)
		if err != nil {
			slog.Error("failed to get package config", "package", pkg, "error", err)
			continue
		}

		if !found {
			slog.Error("no config for package", "package", pkg, "error", err)
			continue
		}

		// TODO add architecture
		output.WriteString("Package: ")
		output.WriteString(config.Name)
		output.WriteString("\nFilename: pool/")
		output.WriteString(config.Name)
		output.WriteString(".deb")

		output.WriteString("\n")
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(output.String()))

	/*
	   Package: myapp
	   Version: 2.3.1-1
	   Architecture: amd64
	   Maintainer: Alice Example <alice@example.com>
	   Filename: pool/main/m/myapp_2.3.1-1_amd64.deb
	   Size: 53212
	   SHA256: 7c1e...
	*/
}

func (server *HTTPServer) Pool(writer http.ResponseWriter, request *http.Request) {
	file := strings.TrimSpace(chi.URLParam(request, "file"))
	if len(file) == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	dot := strings.Index(file, ".")
	if dot == -1 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	packageName := file[:dot]
	config, found, err := server.registry.GetPackageConfig(packageName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	api := ext.NewAPI("/")
	system, err := api.Host().GetSystem()
	err = build.Build("", config, system, api)
}
