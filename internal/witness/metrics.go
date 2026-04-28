package witness

import (
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	KnownLogs          prometheus.Gauge
	LogSize            *prometheus.GaugeVec
	AddCheckpointCount *prometheus.CounterVec

	ReqCount    *prometheus.CounterVec
	ReqInFlight *prometheus.GaugeVec
	ReqDuration *prometheus.SummaryVec
}

func initMetrics() metrics {
	return metrics{
		KnownLogs: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "known_logs",
			Help: "Number of logs known to the witness.",
		}),
		LogSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "log_tree_size",
				Help: "Latest cosigned tree size for each log, by origin.",
				// n.b. yes "tree size" is the term for number of entries, per language in c2sp.org/tlog-checkpoint .
			},
			[]string{"origin"},
		),
		AddCheckpointCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "add_checkpoint_requests_total",
				Help: "Total add-checkpoint requests processed, by log origin.",
			},
			// error is empty on success;
			// origin is empty for requests that fail before the origin is parsed (e.g. malformed input);
			// progress is true if the new checkpoint is for a larger tree, false if same, or empty if earlier errors.
			[]string{"error", "origin", "progress"},
		),

		ReqInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_in_flight_requests",
				Help: "HTTP requests currently being served, by endpoint.",
			},
			[]string{"endpoint"},
		),
		ReqCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "HTTP requests served, by endpoint and response code.",
			},
			[]string{"endpoint", "code"},
		),
		ReqDuration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "http_request_duration_seconds",
				Help:       "HTTP request latency in seconds, by endpoint.",
				Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.025, 0.9: 0.01, 0.99: 0.001},
				MaxAge:     1 * time.Minute,
				AgeBuckets: 6,
			},
			[]string{"endpoint"},
		),
	}
}

// Metrics returns all Prometheus collectors owned by this Witness.
// The caller should register them to a [prometheus.Registry].
func (w *Witness) Metrics() []prometheus.Collector {
	var collectors []prometheus.Collector
	for i := 0; i < reflect.ValueOf(w.metrics).NumField(); i++ {
		collectors = append(collectors, reflect.ValueOf(w.metrics).Field(i).Interface().(prometheus.Collector))
	}
	return collectors
}
