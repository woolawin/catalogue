package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
	config   internal.Config
	system   internal.System
}

func NewHTTPServer(registry reg.Registry, config internal.Config, system internal.System) *HTTPServer {
	return &HTTPServer{registry: registry, config: config, system: system}
}

func (server *HTTPServer) start() error {

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/repositories/{repo}/dists/{distro}/Release", server.Release)
	router.Get("/repositories/{repo}/dists/{distro}/InRelease", server.InRelease)
	router.Get("/repositories/{repo}/pool/{file}", server.Pool)
	router.Get("/repositories/{repo}/dist/{distro}/packages/binary-{arch}/{file}", server.Packages)

	server.server = &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", server.config.Port),
		Handler: router,
	}

	go func() {
		err := server.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "error", err)
		}
		slog.Info("stopping http server")
	}()

	slog.Info("started http server", "port", server.config.Port)

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

func (server *HTTPServer) Packages(writer http.ResponseWriter, request *http.Request) {
	file := strings.TrimSpace(chi.URLParam(request, "file"))

	if len(file) == 0 {
		slog.Error("bad URL to packages file", "file", file)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	compression := ""
	if strings.HasSuffix(file, "/Packages") {
		compression = "plain"
	} else if strings.HasSuffix(file, "/Packages.xz") {
		compression = "xz"
	}

	if compression == "" {
		slog.Error("packages file compression not supported", "file", file)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	contents, found, err := server.registry.ReadReleaseCache(compression)
	if err != nil {
		slog.Error("failed to read cached release file", "file", file, "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		slog.Warn("did not find cached release file", "file", file)
		server.Release(writer, request)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(contents))
}

func (server *HTTPServer) InRelease(writer http.ResponseWriter, request *http.Request) {
	plain, err := server.packagesFile()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	plainBytes := []byte(plain)

	checksum := func(in []byte) string {
		hash := sha256.Sum256(in)
		slice := hash[:]
		return hex.EncodeToString(slice)
	}

	plainHash := checksum(plainBytes)

	var xzHash string
	xzBytes, err := internal.XZ(plainBytes)
	if err != nil {
		slog.Warn("failed to compress package files using xz", "error", err)
	} else {
		xzHash = checksum(xzBytes)
	}

	err = server.registry.CacheRelease("plain", plainBytes)
	if err != nil {
		slog.Warn("failed to save plain cache release", "error", err)
	}

	err = server.registry.CacheRelease("xz", xzBytes)
	if err != nil {
		slog.Warn("failed to save xz compressed cache release", "error", err)
	}

	arch := server.system.Architecture

	sha256 := []string{
		"",
		fmt.Sprintf("%s %d repositories/catalogue/dist/stable/packages/binary-%s/Packages", plainHash, len(plainBytes), arch),
		fmt.Sprintf("%s %d repositories/catalogue/dist/stable/packages/binary-%s/Packages.xz", xzHash, len(xzBytes), arch),
	}

	message := internal.SerializeDebFile([]map[string]string{
		{
			"Hash": "SHA512",
		},
		{
			"Origin":        "Catalogue",
			"Label":         "Catalogue",
			"Suite":         "stable",
			"Codename":      "stable",
			"Version":       server.system.APTDistroVersion,
			"Date":          time.Now().UTC().Truncate(time.Second).Format(time.RFC1123),
			"Architectures": string(server.system.Architecture),
			"Components":    "packages",
			"SHA256":        internal.DebMultiLine(sha256),
		},
	})

	messageHash := sha512.Sum512([]byte(message))

	signature, err := internal.PGPSign(server.config.PrivateAPTKey, messageHash[:])
	if err != nil {
		slog.Error("failed to create signature of message", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	output := strings.Builder{}
	output.WriteString("-----BEGIN PGP SIGNED MESSAGE-----\n")
	output.WriteString(message)
	output.WriteString(string(signature))
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(output.String()))
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
		paragraph["Filename"] = fmt.Sprintf("repositories/catalogue/pool/%s.deb", record.Name)
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
	log := internal.NewLog(internal.NewStdoutLogger(1))

	buffer := bytes.NewBuffer([]byte{})
	ok := assemble.Assemble(buffer, record, log, server.system, api, server.registry)
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
