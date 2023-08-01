package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HTTPServerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Mux       *chi.Mux
	Log       *zap.Logger
}

type HTTPServerResult struct {
	fx.Out

	Server *http.Server
}

func NewHTTPServer(p HTTPServerParams) (HTTPServerResult, error) {
	srv := &http.Server{Addr: ":3000", Handler: p.Mux}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			p.Log.Info("starting HTTP server", zap.String("addr", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return HTTPServerResult{Server: srv}, nil
}

type RouterParams struct {
	fx.In

	Routes []Route `group:"routes"`
}

type RouterResult struct {
	fx.Out

	Router *chi.Mux
}

func ProvideRouter(p RouterParams) (RouterResult, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	for _, route := range p.Routes {
		if route.Type() == "GET" {
			r.Get(route.Pattern(), route.ServeHTTP)
		}
		if route.Type() == "POST" {
			r.Post(route.Pattern(), route.ServeHTTP)
		}
	}

	return RouterResult{Router: r}, nil
}

type Route interface {
	http.Handler

	Pattern() string
	Type() string
}

type RouteParams struct {
	fx.In

	Log *zap.Logger
}

type RouteResult struct {
	fx.Out

	Route Route `group:"routes"`
}

type HomeRoute struct {
	log *zap.Logger
}

func (h *HomeRoute) Pattern() string {
	return "/"
}

func (h *HomeRoute) Type() string {
	return "GET"
}

func (h *HomeRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := fmt.Fprintln(w, "Hello world!")
	if err != nil {
		h.log.Warn("failed to write hello world", zap.Error(err))
	}
}

func NewHomeRoute(p RouteParams) (RouteResult, error) {
	route := &HomeRoute{
		log: p.Log,
	}

	return RouteResult{Route: route}, nil
}

type EchoRoute struct {
	log *zap.Logger
}

func (h *EchoRoute) Pattern() string {
	return "/"
}

func (h *EchoRoute) Type() string {
	return "POST"
}

func (h *EchoRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := io.Copy(w, r.Body); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to handle request:", err)
	}
}

func NewEchoRoute(p RouteParams) (RouteResult, error) {
	route := &EchoRoute{
		log: p.Log,
	}

	return RouteResult{Route: route}, nil
}

func main() {
	fx.New(
		fx.Provide(
			zap.NewExample,
			NewHTTPServer,
			ProvideRouter,
			NewHomeRoute,
			NewEchoRoute,
		),
		fx.Invoke(
			func(*http.Server) {},
		),
	).Run()

}
