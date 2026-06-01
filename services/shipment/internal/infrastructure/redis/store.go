package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/shipment/internal/config"
)

type Store struct {
	client *redis.Client
	cfg    config.RedisConfig
}

func NewStore(client *redis.Client, cfg config.RedisConfig) *Store { return &Store{client: client, cfg: cfg} }
func (s *Store) IsAvailable() bool { return s.client != nil }

func (s *Store) CheckIdempotencyKey(ctx context.Context, key string) (string, error) {
	if s.client == nil { return "", nil }
	return s.client.Get(ctx, fmt.Sprintf("shipment:idempotency:%s", key)).Result()
}

func (s *Store) StoreIdempotencyKey(ctx context.Context, key, shipmentID string, ttl time.Duration) error {
	if s.client == nil { return nil }
	return s.client.Set(ctx, fmt.Sprintf("shipment:idempotency:%s", key), shipmentID, ttl).Err()
}

func (s *Store) CheckWebhookReplay(ctx context.Context, key string) (bool, error) {
	if s.client == nil { return false, nil }
	exists, err := s.client.Exists(ctx, fmt.Sprintf("shipment:webhook:%s", key)).Result()
	return exists > 0, err
}

func (s *Store) MarkWebhookProcessed(ctx context.Context, key string, ttl time.Duration) error {
	if s.client == nil { return nil }
	return s.client.Set(ctx, fmt.Sprintf("shipment:webhook:%s", key), "1", ttl).Err()
}

func (s *Store) AcquireShipmentLock(ctx context.Context, shipmentID string, ttl time.Duration) (bool, error) {
	if s.client == nil { return true, nil }
	return s.client.SetNX(ctx, fmt.Sprintf("lock:shipment:%s", shipmentID), "1", ttl).Result()
}

func (s *Store) ReleaseShipmentLock(ctx context.Context, shipmentID string) error {
	if s.client == nil { return nil }
	return s.client.Del(ctx, fmt.Sprintf("lock:shipment:%s", shipmentID)).Err()
}

// --- QR Code Redis methods ---

func (s *Store) StoreQRCodeState(ctx context.Context, code string, qrID string, ttl time.Duration) error {
	if s.client == nil { return nil }
	key := fmt.Sprintf("shipment:qr:%s", code)
	return s.client.Set(ctx, key, qrID, ttl).Err()
}

func (s *Store) GetQRCodeState(ctx context.Context, code string) (string, error) {
	if s.client == nil { return "", nil }
	return s.client.Get(ctx, fmt.Sprintf("shipment:qr:%s", code)).Result()
}

func (s *Store) InvalidateQRCodeState(ctx context.Context, code string) error {
	if s.client == nil { return nil }
	return s.client.Del(ctx, fmt.Sprintf("shipment:qr:%s", code)).Err()
}

func (s *Store) AcquireScanLock(ctx context.Context, code string, ttl time.Duration) (bool, error) {
	if s.client == nil { return true, nil }
	key := fmt.Sprintf("shipment:qr:scan_lock:%s", code)
	return s.client.SetNX(ctx, key, "1", ttl).Result()
}

func (s *Store) ReleaseScanLock(ctx context.Context, code string) error {
	if s.client == nil { return nil }
	key := fmt.Sprintf("shipment:qr:scan_lock:%s", code)
	return s.client.Del(ctx, key).Err()
}

func (s *Store) IncrementScanCounter(ctx context.Context, shipmentID string) (int64, error) {
	if s.client == nil { return 0, nil }
	key := fmt.Sprintf("shipment:scan_count:%s", shipmentID)
	return s.client.Incr(ctx, key).Result()
}
