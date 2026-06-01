package metrics
import ("github.com/prometheus/client_golang/prometheus"; "github.com/prometheus/client_golang/prometheus/promauto")
var (
	RecRequestsTotal  = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_rec_requests_total", Help: "Total recommendation requests"})
	RecLatency        = promauto.NewHistogram(prometheus.HistogramOpts{Name: "tiki_rec_latency_seconds", Help: "Recommendation latency", Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1}})
	RecErrors         = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_rec_errors_total", Help: "Total recommendation errors"})
	RecCTR            = promauto.NewGauge(prometheus.GaugeOpts{Name: "tiki_rec_ctr", Help: "Click-through rate"})
	RecCoverage       = promauto.NewGauge(prometheus.GaugeOpts{Name: "tiki_rec_coverage", Help: "Product coverage"})
	RecDiversity      = promauto.NewGauge(prometheus.GaugeOpts{Name: "tiki_rec_diversity", Help: "Recommendation diversity"})
	EventsTracked     = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_rec_events_tracked_total", Help: "Total events tracked"})
)
