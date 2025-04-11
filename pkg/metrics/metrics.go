package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

var (
	PVZCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of created PVZ",
		},
	)

	ReceptionsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Total number of created receptions",
		},
	)

	ProductsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total number of added products",
		},
	)
)
