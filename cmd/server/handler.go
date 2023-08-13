package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
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

	Log             *zap.Logger
	RequestsCounter *RequestsCounter
	RequestsTimer   *RequestsTimer
}

type HandlerResult struct {
	fx.Out

	Handler Handler `group:"handlers"`
}

type HomeHandler struct {
	log     *zap.Logger
	counter *RequestsCounter
	timer   *RequestsTimer
}

func (h *HomeHandler) Pattern() string {
	return "/"
}

func (h *HomeHandler) Type() string {
	return "GET"
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := prometheus.NewTimer(h.timer.WithLabelValues("/"))
	defer t.ObserveDuration()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := fmt.Fprintln(w, "Hello world!")
	if err != nil {
		h.log.Warn("failed to write hello world", zap.Error(err))
	}

	c, err := h.counter.GetMetricWith(prometheus.Labels{"handler": "/", "method": "GET", "status": "200"})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to increment counter:", err)
	}
	c.Inc()
}

func NewHomeHandler(p HandlerParams) (HandlerResult, error) {
	handler := &HomeHandler{
		log:     p.Log,
		counter: p.RequestsCounter,
		timer:   p.RequestsTimer,
	}

	return HandlerResult{Handler: handler}, nil
}

type EchoHandler struct {
	log     *zap.Logger
	counter *RequestsCounter
	timer   *RequestsTimer
}

func (h *EchoHandler) Pattern() string {
	return "/"
}

func (h *EchoHandler) Type() string {
	return "POST"
}

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := prometheus.NewTimer(h.timer.WithLabelValues("/"))
	defer t.ObserveDuration()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := io.Copy(w, r.Body); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to handle request:", err)
	}

	c, err := h.counter.GetMetricWith(prometheus.Labels{"handler": "/", "method": "POST", "status": "200"})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to increment counter:", err)
	}
	c.Inc()
}

func NewEchoHandler(p HandlerParams) (HandlerResult, error) {
	handler := &EchoHandler{
		log:     p.Log,
		counter: p.RequestsCounter,
		timer:   p.RequestsTimer,
	}

	return HandlerResult{Handler: handler}, nil
}
