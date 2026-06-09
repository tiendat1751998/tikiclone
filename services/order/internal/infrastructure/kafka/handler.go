package kafka

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func NewOrderEventHandler() EventHandler {
	return func(ctx context.Context, eventType string, payload []byte) error {
		zap.L().Info("processing kafka event",
			zap.String("event_type", eventType),
			zap.String("payload", string(payload)),
		)
		if eventType == "" {
			return fmt.Errorf("empty event type")
		}
		return nil
	}
}
