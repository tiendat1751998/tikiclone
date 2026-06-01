package redis
import ("context"; "fmt"; "time"; "github.com/redis/go-redis/v9"; "github.com/tikiclone/tiki/platforms/notification/internal/config")
type Store struct { rdb *redis.Client; cfg config.RedisConfig }
func NewStore(rdb *redis.Client, cfg config.RedisConfig) *Store { return &Store{rdb: rdb, cfg: cfg} }
func (s *Store) Ping(ctx context.Context) error { return s.rdb.Ping(ctx).Err() }
func (s *Store) Close() error { return s.rdb.Close() }
func (s *Store) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.rdb.Get(ctx, fmt.Sprintf("unread:%s", userID)).Int64()
}
func (s *Store) IncrementUnread(ctx context.Context, userID string) error {
	return s.rdb.Incr(ctx, fmt.Sprintf("unread:%s", userID)).Err()
}
func (s *Store) DecrementUnread(ctx context.Context, userID string) error {
	return s.rdb.Decr(ctx, fmt.Sprintf("unread:%s", userID)).Err()
}
func (s *Store) SetUnreadCount(ctx context.Context, userID string, count int64) error {
	return s.rdb.Set(ctx, fmt.Sprintf("unread:%s", userID), count, 24*time.Hour).Err()
}
func (s *Store) CheckRateLimit(ctx context.Context, key string, maxPerMinute int) (bool, error) {
	return s.rdb.Eval(ctx, `
		local current = redis.call("INCR", KEYS[1])
		if current == 1 then redis.call("EXPIRE", KEYS[1], 60) end
		return current <= tonumber(ARGV[1])
	`, []string{"ratelimit:" + key}, maxPerMinute).Bool()
}
