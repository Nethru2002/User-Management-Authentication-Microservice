package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	ActiveRequests      prometheus.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of response latency for HTTP requests",
				Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"method", "path"},
		),
		ActiveRequests: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_requests",
				Help: "Number of currently active HTTP requests",
			},
		),
	}
}

func (m *Metrics) TrackRequest(method, path string, status int, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}