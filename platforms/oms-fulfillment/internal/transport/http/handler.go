package http

import (
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/inventory"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/ordermanagement"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/pickpack"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/returns"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/warehouse"
)

type Handler struct {
	orders    *ordermanagement.Service
	inventory *inventory.Service
	pickpack  *pickpack.Service
	returns   *returns.Service
	warehouse *warehouse.Service
}

func NewHandler(
	orders *ordermanagement.Service,
	inventory *inventory.Service,
	pickpack *pickpack.Service,
	returns *returns.Service,
	warehouse *warehouse.Service,
) *Handler {
	return &Handler{
		orders:    orders,
		inventory: inventory,
		pickpack:  pickpack,
		returns:   returns,
		warehouse: warehouse,
	}
}
