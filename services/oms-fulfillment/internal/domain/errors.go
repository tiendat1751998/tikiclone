package domain

import "errors"

var (
	ErrFulfillmentNotFound     = errors.New("fulfillment not found")
	ErrFulfillmentNotModifiable = errors.New("fulfillment cannot be modified in current state")
	ErrInvalidStateTransition  = errors.New("invalid state transition")
	ErrWarehouseNotFound       = errors.New("warehouse not found")
	ErrReturnNotFound          = errors.New("return request not found")
	ErrReturnNotModifiable     = errors.New("return cannot be modified in current state")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrItemNotFound            = errors.New("fulfillment item not found")
	ErrDuplicateRequest        = errors.New("duplicate request")
	ErrConcurrentModification  = errors.New("concurrent modification detected")
)
