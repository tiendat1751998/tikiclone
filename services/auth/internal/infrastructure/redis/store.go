package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/auth/internal/config"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Store struct {
	rdb  *redis.Client
	cfg  config.RedisConfig
}

func NewStore(rdb *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{rdb: rdb, cfg: cfg}
}

func (s *Store) StoreSession(ctx context.Context, session *domain.Session) error {
	ctx, span := otel.Tracer("tiki-auth").Start(ctx, "redis.store_session")
	defer span.End()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("session marshal: %w", err)
	}

	key := fmt.Sprintf("session:%s", session.ID)
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.SAdd(ctx, fmt.Sprintf("user_sessions:%s", session.UserID), session.ID)
	pipe.Expire(ctx, fmt.Sprintf("user_sessions:%s", session.UserID), ttl)
	_, err = pipe.Exec(ctx)
	if err != nil {
		span.SetAttributes(attribute.Bool("error", true))
		return fmt.Errorf("redis store session: %w", err)
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Store) DeleteSession(ctx context.Context, sessionID, userID string) error {
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("session:%s", sessionID))
	pipe.SRem(ctx, fmt.Sprintf("user_sessions:%s", userID), sessionID)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) RevokeAllUserSessions(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user_sessions:%s", userID)
	sessionIDs, err := s.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}

	pipe := s.rdb.Pipeline()
	for _, sid := range sessionIDs {
		pipe.Del(ctx, fmt.Sprintf("session:%s", sid))
	}
	pipe.Del(ctx, key)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) CountUserSessions(ctx context.Context, userID string) (int64, error) {
	return s.rdb.SCard(ctx, fmt.Sprintf("user_sessions:%s", userID)).Result()
}

func (s *Store) BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:access:%s", hashToken(token))
	return s.rdb.Set(ctx, key, "1", ttl).Err()
}

func (s *Store) BlacklistRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:refresh:%s", tokenID)
	return s.rdb.Set(ctx, key, "1", ttl).Err()
}

func (s *Store) IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:access:%s", hashToken(token))
	exists, err := s.rdb.Exists(ctx, key).Result()
	return exists > 0, err
}

func (s *Store) IsRefreshTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:refresh:%s", tokenID)
	exists, err := s.rdb.Exists(ctx, key).Result()
	return exists > 0, err
}

func (s *Store) MarkRefreshTokenUsed(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("used_refresh:%s", tokenID)
	return s.rdb.Set(ctx, key, "1", ttl).Err()
}

func (s *Store) IsRefreshTokenReused(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("used_refresh:%s", tokenID)
	exists, err := s.rdb.Exists(ctx, key).Result()
	return exists > 0, err
}

func (s *Store) RecordLoginAttempt(ctx context.Context, email, ip string, success bool) error {
	key := fmt.Sprintf("login_attempts:%s", email)
	pipe := s.rdb.Pipeline()
	pipe.LPush(ctx, key, fmt.Sprintf("%s:%s:%t", time.Now().Format(time.RFC3339), ip, success))
	pipe.LTrim(ctx, key, 0, 99)
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) GetRecentLoginAttempts(ctx context.Context, email string, window time.Duration) (int, error) {
	key := fmt.Sprintf("login_attempts:%s", email)
	attempts, err := s.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-window)
	count := 0
	for _, a := range attempts {
		t, err := time.Parse(time.RFC3339, a[:25])
		if err == nil && t.After(cutoff) {
			count++
		}
	}
	return count, nil
}

func (s *Store) RecordSuspiciousIP(ctx context.Context, ip string, ttl time.Duration) error {
	key := fmt.Sprintf("suspicious_ip:%s", ip)
	return s.rdb.Incr(ctx, key).Err()
}

func (s *Store) GetLoginCountByIP(ctx context.Context, ip string, window time.Duration) (int, error) {
	key := fmt.Sprintf("login_ip:%s", ip)
	val, err := s.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (s *Store) Ping(ctx context.Context) error {
	return s.rdb.Ping(ctx).Err()
}

func (s *Store) Close() error {
	return s.rdb.Close()
}

func (s *Store) SetResetToken(ctx context.Context, userID, tokenHash string, ttl time.Duration) error {
	key := fmt.Sprintf("password_reset:%s:%s", userID, tokenHash)
	return s.rdb.Set(ctx, key, userID, ttl).Err()
}

func (s *Store) ValidateAndConsumeResetToken(ctx context.Context, tokenHash string) (string, error) {
	keys, err := s.rdb.Keys(ctx, "password_reset:*:"+tokenHash).Result()
	if err != nil || len(keys) == 0 {
		return "", fmt.Errorf("reset token not found or expired")
	}
	pipe := s.rdb.Pipeline()
	userID := pipe.Get(ctx, keys[0])
	pipe.Del(ctx, keys[0])
	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", err
	}
	return userID.Val(), nil
}

func (s *Store) SetVerifyToken(ctx context.Context, userID, tokenHash string, ttl time.Duration) error {
	key := fmt.Sprintf("email_verify:%s:%s", userID, tokenHash)
	return s.rdb.Set(ctx, key, userID, ttl).Err()
}

func (s *Store) ValidateAndConsumeVerifyToken(ctx context.Context, tokenHash string) (string, error) {
	keys, err := s.rdb.Keys(ctx, "email_verify:*:"+tokenHash).Result()
	if err != nil || len(keys) == 0 {
		return "", fmt.Errorf("verify token not found or expired")
	}
	pipe := s.rdb.Pipeline()
	userID := pipe.Get(ctx, keys[0])
	pipe.Del(ctx, keys[0])
	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", err
	}
	return userID.Val(), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
