package defaultresolver

import "github.com/prometheus/client_golang/prometheus"

var (
	resolverHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resolver_hits_total",
			Help: "Number of DB cache hits",
		},
		[]string{"resolver_id"},
	)
)

func init() {
	prometheus.MustRegister(resolverHits)
}
