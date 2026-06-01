package redis
import ("context"; "fmt"; "time"; "github.com/redis/go-redis/v9"; "github.com/tikiclone/tiki/services/product-catalog/internal/config")
type Store struct { rdb *redis.Client; cfg config.RedisConfig }
func NewStore(rdb *redis.Client, cfg config.RedisConfig) *Store { return &Store{rdb: rdb, cfg: cfg} }
func (s *Store) Ping(ctx context.Context) error { return s.rdb.Ping(ctx).Err() }
func (s *Store) Close() error { return s.rdb.Close() }
func (s *Store) GetProduct(ctx context.Context, id string) ([]byte, error) {
	return s.rdb.Get(ctx, fmt.Sprintf("product:%s", id)).Bytes()
}
func (s *Store) SetProduct(ctx context.Context, id string, data []byte, ttl time.Duration) error {
	return s.rdb.Set(ctx, fmt.Sprintf("product:%s", id), data, ttl).Err()
}
func (s *Store) DeleteProduct(ctx context.Context, id string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("product:%s", id)).Err()
}
func (s *Store) GetCategoryTree(ctx context.Context) ([]byte, error) {
	return s.rdb.Get(ctx, "category:tree").Bytes()
}
func (s *Store) SetCategoryTree(ctx context.Context, data []byte, ttl time.Duration) error {
	return s.rdb.Set(ctx, "category:tree", data, ttl).Err()
}
func (s *Store) InvalidateCategory(ctx context.Context, id string) error {
	return s.rdb.Del(ctx, fmt.Sprintf("category:%s", id)).Err()
}
func (s *Store) CheckIdempotency(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.rdb.SetNX(ctx, "idempotency:"+key, "1", ttl).Result()
}
