package logging

import (
	"go.uber.org/zap"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
)

var Logger *zap.Logger

func Init(appName, logLevel string) {
	Logger = observability.InitLogger(appName, logLevel)
}

func WithFulfillmentID(id string) zap.Field {
	return zap.String("fulfillment_id", id)
}

func WithOrderID(id string) zap.Field {
	return zap.String("order_id", id)
}

func WithReturnID(id string) zap.Field {
	return zap.String("return_id", id)
}
