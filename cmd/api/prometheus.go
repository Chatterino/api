package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/Chatterino/api/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
        servedRoutes = prometheus.NewCounterVec( prometheus.CounterOpts{
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

func PrometheusMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
                path := strings.Split(r.URL.Path, string(os.PathSeparator))[1]
                servedRoutes.WithLabelValues(path).Inc()
                next.ServeHTTP(w, r)
        })
}

func init() {
    prometheus.MustRegister(servedRoutes)
}

