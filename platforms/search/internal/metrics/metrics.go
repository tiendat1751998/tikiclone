package metrics
import ("github.com/prometheus/client_golang/prometheus"; "github.com/prometheus/client_golang/prometheus/promauto")
var (
	SearchQueriesTotal     = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_search_queries_total", Help: "Total search queries"})
	AutocompleteRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_search_autocomplete_total", Help: "Total autocomplete requests"})
	DocumentsIndexed       = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_search_documents_indexed_total", Help: "Total documents indexed"})
	SearchErrors           = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_search_errors_total", Help: "Total search errors"})
	SearchLatency          = promauto.NewHistogram(prometheus.HistogramOpts{Name: "tiki_search_latency_seconds", Help: "Search latency", Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1}})
)
