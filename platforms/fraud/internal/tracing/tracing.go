package tracing

import (
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
)

func Init(serviceName, endpoint string) (func(), error) {
	return observability.InitTracer(serviceName, endpoint)
}
