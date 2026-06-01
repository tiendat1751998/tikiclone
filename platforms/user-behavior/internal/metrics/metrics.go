package metrics
import ("github.com/prometheus/client_golang/prometheus"; "github.com/prometheus/client_golang/prometheus/promauto")
var (
	EventsIngested  = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_ub_events_ingested_total", Help: "Total events ingested"})
	BatchIngested   = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_ub_batch_ingested_total", Help: "Total batch events ingested"})
	IngestionErrors = promauto.NewCounter(prometheus.CounterOpts{Name: "tiki_ub_ingestion_errors_total", Help: "Total ingestion errors"})
)
