package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/cart/internal/config"
	"github.com/tikiclone/tiki/services/cart/internal/metrics"
)

type Store struct {
	rdb *redis.Client
	cfg config.RedisConfig
}

func NewStore(rdb *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{rdb: rdb, cfg: cfg}
}

func (s *Store) isAvailable() bool {
	return s.rdb != nil
}

func (s *Store) Ping(ctx context.Context) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Ping(ctx).Err()
}

func (s *Store) Close() error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Close()
}

// === Cart Cache ===

type CartCache struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Items     []CartItemCache `json:"items"`
	ItemCount int             `json:"item_count"`
	Subtotal  int64           `json:"subtotal"`
	Currency  string          `json:"currency"`
	Version   int64           `json:"version"`
}

type CartItemCache struct {
	ID          string `json:"id"`
	SKU         string `json:"sku"`
	ProductName string `json:"product_name"`
	ShopID      string `json:"shop_id"`
	ShopName    string `json:"shop_name"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	TotalPrice  int64  `json:"total_price"`
	ImageURL    string `json:"image_url"`
	IsSelected  bool   `json:"is_selected"`
	IsAvailable bool   `json:"is_available"`
}

func (s *Store) GetCart(ctx context.Context, cartID string) (*CartCache, error) {
	if !s.isAvailable() {
		return nil, nil
	}
	key := cartKey(cartID)
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		metrics.CacheMissesTotal.WithLabelValues("cart", "redis").Inc()
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cc CartCache
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, err
	}
	metrics.CacheHitsTotal.WithLabelValues("cart", "redis").Inc()
	return &cc, nil
}

func (s *Store) SetCart(ctx context.Context, cartID string, cc *CartCache, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	key := cartKey(cartID)
	data, err := json.Marshal(cc)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, data, ttl).Err()
}

func (s *Store) DeleteCart(ctx context.Context, cartID string) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Del(ctx, cartKey(cartID)).Err()
}

// === User Cart Index ===

func (s *Store) SetUserCart(ctx context.Context, userID, cartID string, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	key := userCartKey(userID)
	return s.rdb.Set(ctx, key, cartID, ttl).Err()
}

func (s *Store) GetUserCart(ctx context.Context, userID string) (string, error) {
	if !s.isAvailable() {
		return "", nil
	}
	key := userCartKey(userID)
	return s.rdb.Get(ctx, key).Result()
}

func (s *Store) DeleteUserCart(ctx context.Context, userID string) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Del(ctx, userCartKey(userID)).Err()
}

// === Session Cart ===

func (s *Store) SetSessionCart(ctx context.Context, sessionID, cartID string, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	key := sessionCartKey(sessionID)
	return s.rdb.Set(ctx, key, cartID, ttl).Err()
}

func (s *Store) GetSessionCart(ctx context.Context, sessionID string) (string, error) {
	if !s.isAvailable() {
		return "", nil
	}
	key := sessionCartKey(sessionID)
	return s.rdb.Get(ctx, key).Result()
}

// === Checkout Preview Cache ===

func (s *Store) SetCheckoutPreview(ctx context.Context, previewID string, data []byte, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	key := checkoutPreviewKey(previewID)
	return s.rdb.Set(ctx, key, data, ttl).Err()
}

func (s *Store) GetCheckoutPreview(ctx context.Context, previewID string) ([]byte, error) {
	if !s.isAvailable() {
		return nil, nil
	}
	key := checkoutPreviewKey(previewID)
	return s.rdb.Get(ctx, key).Bytes()
}

// === Idempotency ===

func (s *Store) CheckIdempotency(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if !s.isAvailable() {
		return false, nil
	}
	fullKey := "idempotency:" + key
	return s.rdb.SetNX(ctx, fullKey, "1", ttl).Result()
}

func (s *Store) SetIdempotencyResult(ctx context.Context, key string, result []byte, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	fullKey := "idempotency_result:" + key
	return s.rdb.Set(ctx, fullKey, result, ttl).Err()
}

func (s *Store) GetIdempotencyResult(ctx context.Context, key string) ([]byte, error) {
	if !s.isAvailable() {
		return nil, nil
	}
	fullKey := "idempotency_result:" + key
	return s.rdb.Get(ctx, fullKey).Bytes()
}

// === Key helpers ===

func cartKey(cartID string) string {
	return fmt.Sprintf("cart:%s", cartID)
}

func userCartKey(userID string) string {
	return fmt.Sprintf("user_cart:%s", userID)
}

func sessionCartKey(sessionID string) string {
	return fmt.Sprintf("session_cart:%s", sessionID)
}

func checkoutPreviewKey(previewID string) string {
	return fmt.Sprintf("checkout_preview:%s", previewID)
}
