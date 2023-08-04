package main

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			zap.NewExample,
			NewHTTPServer,
			NewRouter,
			NewHomeHandler,
			NewEchoHandler,
		),
		fx.Invoke(
			func(*http.Server) {},
		),
	).Run()

}
