package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderSnapshot struct {
	ID           string    `db:"id" json:"id"`
	OrderID      string    `db:"order_id" json:"order_id"`
	SnapshotData []byte    `db:"snapshot_data" json:"snapshot_data"`
	Checksum     string    `db:"checksum" json:"checksum"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type CartSnapshot struct {
	Items       []SnapshotItem `json:"items"`
	TotalAmount int64          `json:"total_amount"`
	Currency    string         `json:"currency"`
	Coupons     []string       `json:"coupons,omitempty"`
}

type SnapshotItem struct {
	ProductID string `json:"product_id"`
	SkuID     string `json:"sku_id"`
	ShopID    string `json:"shop_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
	ImageURL  string `json:"image_url,omitempty"`
}

// UnmarshalJSON handles both "unit_price" and "price" field names
func (s *SnapshotItem) UnmarshalJSON(data []byte) error {
	type Alias SnapshotItem
	aux := &struct {
		Price int64 `json:"price"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	// If unit_price is 0 but price is set, use price
	if s.UnitPrice == 0 && aux.Price != 0 {
		s.UnitPrice = aux.Price
	}
	return nil
}

func NewOrderSnapshot(orderID string, cart *CartSnapshot) (*OrderSnapshot, error) {
	data, err := json.Marshal(cart)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	return &OrderSnapshot{
		ID:           uuid.New().String(),
		OrderID:      orderID,
		SnapshotData: data,
		Checksum:     hex.EncodeToString(hash[:]),
		CreatedAt:    time.Now().UTC(),
	}, nil
}

func (s *OrderSnapshot) VerifyChecksum() bool {
	hash := sha256.Sum256(s.SnapshotData)
	return hex.EncodeToString(hash[:]) == s.Checksum
}

func (s *OrderSnapshot) CartSnapshot() (*CartSnapshot, error) {
	var cart CartSnapshot
	if err := json.Unmarshal(s.SnapshotData, &cart); err != nil {
		return nil, err
	}
	return &cart, nil
}
