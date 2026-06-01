package logging

import (
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

func Init(serviceName, level string) *zap.Logger {
	return observability.InitLogger(serviceName, level)
}

func WithSKU(logger *zap.Logger, sku string) *zap.Logger {
	return logger.With(zap.String("sku", sku))
}

func WithWarehouse(logger *zap.Logger, warehouseID string) *zap.Logger {
	return logger.With(zap.String("warehouse_id", warehouseID))
}

func WithReservation(logger *zap.Logger, reservationID string) *zap.Logger {
	return logger.With(zap.String("reservation_id", reservationID))
}
