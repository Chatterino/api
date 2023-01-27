package main

import (
	"net/http"

	"github.com/Chatterino/api/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	servedRoutes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total number of requests",
	},
		[]string{"path"},
	)
)

func listenPrometheus(cfg config.APIConfig) {
	router := chi.NewRouter()

	srv := &http.Server{
		Handler: router,
		Addr:    cfg.PrometheusBindAddress,
	}

	router.Handle("/metrics", promhttp.Handler())

	go srv.ListenAndServe()
}
