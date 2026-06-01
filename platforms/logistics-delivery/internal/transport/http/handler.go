package http

import (
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/shipments"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/tracking"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/routing"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/dispatch"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/couriers"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/fulfillment"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/pickups"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/estimations"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/replay"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/synchronization"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/idempotency"
)

type Handler struct {
	shipments      *shipments.Service
	tracking       *tracking.Service
	routing        *routing.Service
	dispatch       *dispatch.Service
	couriers       *couriers.Service
	fulfillment    *fulfillment.Service
	pickups        *pickups.Service
	estimations    *estimations.Service
	replay         *replay.Service
	syncService    *synchronization.Service
	idempotency    *idempotency.Store
}

func NewHandler(
	shipments *shipments.Service,
	tracking *tracking.Service,
	routing *routing.Service,
	dispatch *dispatch.Service,
	couriers *couriers.Service,
	fulfillment *fulfillment.Service,
	pickups *pickups.Service,
	estimations *estimations.Service,
	replay *replay.Service,
	syncService *synchronization.Service,
	idempotency *idempotency.Store,
) *Handler {
	return &Handler{
		shipments:    shipments,
		tracking:     tracking,
		routing:      routing,
		dispatch:     dispatch,
		couriers:     couriers,
		fulfillment:  fulfillment,
		pickups:      pickups,
		estimations:  estimations,
		replay:       replay,
		syncService:  syncService,
		idempotency:  idempotency,
	}
}
