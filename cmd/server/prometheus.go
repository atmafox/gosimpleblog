package main

import (
	"go.uber.org/fx"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Registry struct {
	*prometheus.Registry
}

type RegistryParams struct {
	fx.In
	Lifecycle fx.Lifecycle
}

type RegistryResult struct {
	fx.Out

	Registry Registry
}

func NewRegistry(p RegistryParams) (RegistryResult, error) {
	r := prometheus.NewRegistry()

	// Get our default Golang and process info collectors
	r.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollections(collectors.GoRuntimeMetricsCollection),
	))

	r.MustRegister(collectors.NewProcessCollector(
		collectors.ProcessCollectorOpts{},
	))

	return RegistryResult{Registry: Registry{r}}, nil
}

type RequestsCounter struct {
	*prometheus.CounterVec
}

type RequestsCounterParams struct {
	fx.In

	Registry Registry
}

type RequestsCounterResult struct {
	fx.Out

	RequestsCounter *RequestsCounter
}

func NewRequestsCounter(p RequestsCounterParams) (RequestsCounterResult, error) {
	v := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "How many HTTP requests processed, partitioned by handler, HTTP method, and status code.",
		},
		[]string{"handler", "method", "status"},
	)

	if err := p.Registry.Register(v); err != nil {
		return RequestsCounterResult{}, err
	}

	return RequestsCounterResult{RequestsCounter: &RequestsCounter{v}}, nil
}

type RequestsTimer struct {
	*prometheus.HistogramVec
}

type RequestsTimerParams struct {
	fx.In

	Registry Registry
}

type RequestsTimerResult struct {
	fx.Out

	RequestsTimer *RequestsTimer
}

func NewRequestsTimer(p RequestsTimerParams) (RequestsTimerResult, error) {
	v := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requests_duration_seconds",
			Help:    "Histogram for the handler duration, partitioned by handler.",
			Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
		},
		[]string{"handler"},
	)

	if err := p.Registry.Register(v); err != nil {
		return RequestsTimerResult{}, err
	}

	return RequestsTimerResult{RequestsTimer: &RequestsTimer{v}}, nil
}
