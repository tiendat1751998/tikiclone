package tracing
import ("github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "github.com/tikiclone/tiki/services/product-catalog/internal/config")
func Init(cfg config.OTELConfig) (func(), error) { return observability.InitTracer(cfg.ServiceName, cfg.Endpoint) }
