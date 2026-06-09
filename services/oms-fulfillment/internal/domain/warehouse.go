package domain

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Code          string    `db:"code" json:"code"`
	AddressJSON   string    `db:"address" json:"-"`
	ContactPhone  *string   `db:"contact_phone" json:"contact_phone,omitempty"`
	ContactEmail  *string   `db:"contact_email" json:"contact_email,omitempty"`
	CapacityTotal int       `db:"capacity_total" json:"capacity_total"`
	CapacityUsed  int       `db:"capacity_used" json:"capacity_used"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func NewWarehouse(name, code string, capacity int) *Warehouse {
	return &Warehouse{
		ID:            uuid.New().String(),
		Name:          name,
		Code:          code,
		CapacityTotal: capacity,
		IsActive:      true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

type WarehouseInventory struct {
	ID                string     `db:"id" json:"id"`
	WarehouseID       string     `db:"warehouse_id" json:"warehouse_id"`
	SKUID             string     `db:"sku_id" json:"sku_id"`
	QuantityAvailable int        `db:"quantity_available" json:"quantity_available"`
	QuantityReserved  int        `db:"quantity_reserved" json:"quantity_reserved"`
	QuantityDamaged   int        `db:"quantity_damaged" json:"quantity_damaged"`
	ReorderLevel      int        `db:"reorder_level" json:"reorder_level"`
	ReorderQuantity   int        `db:"reorder_quantity" json:"reorder_quantity"`
	LocationCode      *string    `db:"location_code" json:"location_code,omitempty"`
	LastCountedAt     *time.Time `db:"last_counted_at" json:"last_counted_at,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}
