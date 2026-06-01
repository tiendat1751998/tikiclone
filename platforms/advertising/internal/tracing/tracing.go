package tracing

import (
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/advertising/internal/config"
)

func Init(cfg config.OTELConfig) (func(), error) {
	return observability.InitTracer(cfg.ServiceName, cfg.Endpoint)
}
