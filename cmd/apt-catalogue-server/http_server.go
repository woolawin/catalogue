package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/registry"
	"github.com/woolawin/catalogue/internal/update"
)

type HTTPServer struct {
	server *http.Server
	config internal.Config
	system internal.System
	api    *ext.API
}

func NewHTTPServer(config internal.Config, system internal.System, api *ext.API) *HTTPServer {
	return &HTTPServer{config: config, system: system, api: api}
}

func (server *HTTPServer) start() error {

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/repositories/{repo}/dists/{distro}/Release", server.Release)
	router.Get("/repositories/{repo}/dists/{distro}/InRelease", server.InRelease)
	router.Get("/repositories/{repo}/pool/{package}/{version}/{commit}/install.deb", server.Pool)
	router.Get("/repositories/{repo}/dists/{distro}/packages/{arch}/{file}", server.Packages)

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
	if file == "Packages" {
		compression = "plain"
	} else if file == "Packages.xz" {
		compression = "xz"
	}

	if compression == "" {
		slog.Error("packages file compression not supported", "file", file)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	contents, found, err := registry.ReadReleaseCache(compression)
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

	err = registry.CacheRelease("plain", plainBytes)
	if err != nil {
		slog.Warn("failed to save plain cache release", "error", err)
	}

	arch := server.system.Architecture

	sha256 := []string{
		"",
		fmt.Sprintf("%s %d packages/binary-%s/Packages", plainHash, len(plainBytes), arch),
	}

	message := internal.SerializeDebFile([]map[string]string{
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

	signature, err := internal.PGPSign(server.config.PrivateAPTKey, []byte(message))
	if err != nil {
		slog.Error("failed to create signature of message", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Write([]byte(signature))
}

func (server *HTTPServer) packagesFile() (string, error) {
	packages, err := registry.ListPackages()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		return "", err
	}

	group := sync.WaitGroup{}
	mutex := sync.Mutex{}
	var paragraphs []map[string]string

	for _, pkg := range packages {
		group.Go(func() {

			record, found, err := registry.GetPackageRecord(pkg)
			if err != nil {
				slog.Error("failed to get package config", "package", pkg, "error", err)
				return
			}

			if !found {
				slog.Error("no record for package", "package", pkg, "error", err)
				return
			}

			log := internal.NewLog(internal.NewStdoutLogger(5))
			updated, build, ok := update.Update(record, log, server.system, server.api)
			if ok {
				record = updated
			}

			paragraph := make(map[string]string)
			paragraph["Package"] = record.Name
			paragraph["Version"] = record.LatestPin.VersionName
			paragraph["Filename"] = packageFilename(record)
			paragraph["Depends"] = record.Metadata.Dependencies
			paragraph["Section"] = record.Metadata.Category
			paragraph["Homepage"] = record.Metadata.Homepage
			paragraph["Maintainer"] = record.Metadata.Maintainer
			paragraph["Description"] = record.Metadata.Description
			paragraph["Architecture"] = record.Metadata.Architecture
			paragraph["SHA256"] = build.SHA245
			paragraph["Size"] = strconv.FormatInt(build.Size, 10)

			mutex.Lock()
			defer mutex.Unlock()
			paragraphs = append(paragraphs, paragraph)
		})
	}

	group.Wait()

	return internal.SerializeDebFile(paragraphs), nil
}

func packageFilename(record config.Record) string {
	filename := strings.Builder{}
	filename.WriteString("pool/")
	filename.WriteString(record.Name)
	filename.WriteString("/")
	filename.WriteString(record.LatestPin.VersionName)
	filename.WriteString("/")
	filename.WriteString(record.LatestPin.CommitHash)
	filename.WriteString("/install.deb")
	return filename.String()
}

func (server *HTTPServer) Pool(writer http.ResponseWriter, request *http.Request) {
	pkg := strings.TrimSpace(chi.URLParam(request, "package"))
	if len(pkg) == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	version := strings.TrimSpace(chi.URLParam(request, "version"))
	commit := strings.TrimSpace(chi.URLParam(request, "commit"))

	record, found, err := registry.GetPackageRecord(pkg)
	if err != nil {
		slog.Error("could not get record file for package", "package", pkg, "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		slog.Error("could not find package", "package", pkg)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if len(version) != 0 && len(commit) != 0 {
		record.LatestPin = config.Pin{VersionName: version, CommitHash: commit}
	}

	var wanted *config.BuildFile
	for _, build := range record.Builds {
		if build.Version != version || build.CommitHash != commit {
			continue
		}
		wanted = &build
		break
	}

	if wanted == nil {
		slog.Error("unable to find build", "package", pkg, "version", version, "commit", commit)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	file, err := os.Open(wanted.Path)
	if err != nil {
		slog.Error("could not open file", "path", wanted.Path, "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(writer, file)
	if err != nil {
		slog.Error("could not write file to response", "package", pkg, "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
