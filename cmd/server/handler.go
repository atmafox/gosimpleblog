package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

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
