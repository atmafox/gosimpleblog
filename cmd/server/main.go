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

	Handlers []Handler `group:"handlers"`
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

type Handler interface {
	http.Handler

	Pattern() string
	Type() string
}

type HandlerParams struct {
	fx.In

	Log *zap.Logger
}

type HandlerResult struct {
	fx.Out

	Handler Handler `group:"handlers"`
}

type HomeHandler struct {
	log *zap.Logger
}

func (h *HomeHandler) Pattern() string {
	return "/"
}

func (h *HomeHandler) Type() string {
	return "GET"
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := fmt.Fprintln(w, "Hello world!")
	if err != nil {
		h.log.Warn("failed to write hello world", zap.Error(err))
	}
}

func NewHomeHandler(p HandlerParams) (HandlerResult, error) {
	handler := &HomeHandler{
		log: p.Log,
	}

	return HandlerResult{Handler: handler}, nil
}

type EchoHandler struct {
	log *zap.Logger
}

func (h *EchoHandler) Pattern() string {
	return "/"
}

func (h *EchoHandler) Type() string {
	return "POST"
}

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := io.Copy(w, r.Body); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to handle request:", err)
	}
}

func NewEchoHandler(p HandlerParams) (HandlerResult, error) {
	handler := &EchoHandler{
		log: p.Log,
	}

	return HandlerResult{Handler: handler}, nil
}

func main() {
	fx.New(
		fx.Provide(
			zap.NewExample,
			NewHTTPServer,
			ProvideRouter,
			NewHomeHandler,
			NewEchoHandler,
		),
		fx.Invoke(
			func(*http.Server) {},
		),
	).Run()

}
