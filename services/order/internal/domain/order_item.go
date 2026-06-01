package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID         string `db:"id" json:"id"`
	OrderID    string `db:"order_id" json:"order_id"`
	ProductID  string `db:"product_id" json:"product_id"`
	SkuID      string `db:"sku_id" json:"sku_id"`
	ShopID     string `db:"shop_id" json:"shop_id"`
	Name       string `db:"-" json:"name,omitempty"`
	ImageURL   string `db:"-" json:"image_url,omitempty"`
	Price      int64  `db:"-" json:"price"`
	Quantity   int    `db:"quantity" json:"quantity"`
	UnitPrice  int64  `db:"unit_price" json:"unit_price"`
	TotalPrice int64  `db:"total_price" json:"total_price"`
	Total      int64  `db:"-" json:"total"`
	Snapshot   []byte `db:"snapshot" json:"snapshot,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

func NewOrderItem(orderID, productID, skuID, shopID string, quantity int, unitPrice int64, snapshot []byte) *OrderItem {
	return &OrderItem{
		ID:         uuid.New().String(),
		OrderID:    orderID,
		ProductID:  productID,
		SkuID:      skuID,
		ShopID:     shopID,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		TotalPrice: int64(quantity) * unitPrice,
		Snapshot:   snapshot,
		CreatedAt:  time.Now().UTC(),
	}
}
