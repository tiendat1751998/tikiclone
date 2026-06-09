package kafka

import (
	"context"
	"go.uber.org/zap"
)

type FulfillmentEventHandler struct{}

func NewFulfillmentEventHandler() *FulfillmentEventHandler {
	return &FulfillmentEventHandler{}
}

func (h *FulfillmentEventHandler) HandleEvent(ctx context.Context, eventType string, payload []byte) error {
	zap.L().Info("received fulfillment event",
		zap.String("event_type", eventType),
		zap.ByteString("payload", payload),
	)
	return nil
}
