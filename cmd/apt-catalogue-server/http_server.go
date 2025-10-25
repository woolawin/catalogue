package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/woolawin/catalogue/internal"
	assemble "github.com/woolawin/catalogue/internal/assmeble"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

type HTTPServer struct {
	registry reg.Registry
	server   *http.Server
}

func NewHTTPServer(registry reg.Registry) *HTTPServer {
	return &HTTPServer{}
}

func (server *HTTPServer) start() error {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/dists/{distro}/Release", server.Release)
	router.Get("/dists/{distro}/InRelease", server.InRelease)
	router.Get("/pool/{file}", server.Pool)

	server.server = &http.Server{
		Addr:    "localhost:3465",
		Handler: router,
	}

	go func() {
		err := server.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "error", err)
		}
		slog.Info("stopping http server")
	}()

	return nil
}

func (server *HTTPServer) Shutdown(ctx context.Context) {
	if server.server == nil {
		return
	}
	err := server.server.Shutdown(ctx)
	if err != nil {
		slog.Error("error shutting down server", "error", err)
	}
}

func (server *HTTPServer) Release(writer http.ResponseWriter, request *http.Request) {
	content, err := server.packagesFile()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(content))
}

func (server *HTTPServer) InRelease(writer http.ResponseWriter, request *http.Request) {
	content, err := server.packagesFile()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(content))
}

func (server *HTTPServer) packagesFile() (string, error) {
	packages, err := server.registry.ListPackages()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		return "", err
	}

	var paragraphs []map[string]string

	for _, pkg := range packages {
		record, found, err := server.registry.GetPackageRecord(pkg)
		if err != nil {
			slog.Error("failed to get package config", "package", pkg, "error", err)
			continue
		}

		if !found {
			slog.Error("no record for package", "package", pkg, "error", err)
			continue
		}

		paragraph := make(map[string]string)
		paragraph["Package"] = record.Name
		paragraph["Version"] = record.LatestPin.VersionName
		paragraph["Filename"] = record.Name + ".deb"
		paragraph["Depends"] = record.Metadata.Dependencies
		paragraph["Section"] = record.Metadata.Category
		paragraph["Homepage"] = record.Metadata.Homepage
		paragraph["Maintainer"] = record.Metadata.Maintainer
		paragraph["Description"] = record.Metadata.Description
		paragraph["Architecture"] = record.Metadata.Architecture

		paragraphs = append(paragraphs, paragraph)
	}

	return internal.SerializeDebFile(paragraphs), nil
}

func (server *HTTPServer) Pool(writer http.ResponseWriter, request *http.Request) {
	file := strings.TrimSpace(chi.URLParam(request, "file"))
	if len(file) == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	dot := strings.Index(file, ".")
	if dot == -1 {
		slog.Error("faile did not contain dot", "file", file)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	packageName := file[:dot]
	record, found, err := server.registry.GetPackageRecord(packageName)
	if err != nil {
		slog.Error("could not get record file for package", "package", packageName, "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		slog.Error("could not find package", "package", packageName)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	api := ext.NewAPI("/")
	system, err := api.Host.GetSystem()
	log := internal.NewLog(internal.NewStdoutLogger(1))

	buffer := bytes.NewBuffer([]byte{})
	ok := assemble.Assemble(buffer, record, log, system, api, server.registry)
	if !ok {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(writer, buffer)
	if err != nil {
		slog.Error("could not write file to response", "package", packageName, "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
