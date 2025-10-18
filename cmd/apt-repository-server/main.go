package main

import (
	"fmt"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Hello")
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}
