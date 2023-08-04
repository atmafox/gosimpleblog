package main

import (
	"context"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type HTTPServerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Router    *chi.Mux
	Log       *zap.Logger
}

type HTTPServerResult struct {
	fx.Out

	Server *http.Server
}

func NewHTTPServer(p HTTPServerParams) (HTTPServerResult, error) {
	srv := &http.Server{Addr: ":3000", Handler: p.Router}

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
