package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
}

type Server struct {
}

func (server *Server) start() error {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/dists/{distro}/Release", server.Release)
	router.Get("/pool/{file}", server.Pool)
	return nil
}

func (server *Server) Release(writer http.ResponseWriter, request *http.Request) {
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

func (server *Server) Pool(writer http.ResponseWriter, request *http.Request) {

}
