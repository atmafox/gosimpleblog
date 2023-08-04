package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In

	Handlers []Handler `group:"handlers"`
}

type RouterResult struct {
	fx.Out

	Router *chi.Mux
}

func NewRouter(p RouterParams) (RouterResult, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	for _, handler := range p.Handlers {
		if handler.Type() == "GET" {
			r.Get(handler.Pattern(), handler.ServeHTTP)
		}
		if handler.Type() == "POST" {
			r.Post(handler.Pattern(), handler.ServeHTTP)
		}
	}

	return RouterResult{Router: r}, nil
}
