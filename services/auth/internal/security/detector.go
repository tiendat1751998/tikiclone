package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/auth/internal/config"
)

type SuspiciousDetector struct {
	rdb *redis.Client
	cfg config.SecurityConfig
}

func NewSuspiciousDetector(rdb *redis.Client, cfg config.SecurityConfig) *SuspiciousDetector {
	return &SuspiciousDetector{rdb: rdb, cfg: cfg}
}

func (d *SuspiciousDetector) IsSuspicious(ctx context.Context, userID, ip string) bool {
	if d.rdb == nil {
		return false
	}

	key := fmt.Sprintf("known_ips:%s", userID)
	isKnown, err := d.rdb.SIsMember(ctx, key, ip).Result()
	if err == nil && isKnown {
		d.rdb.Expire(ctx, key, 90*24*time.Hour)
		return false
	}

	knownCount, _ := d.rdb.SCard(ctx, key).Result()
	if knownCount < 5 {
		d.rdb.SAdd(ctx, key, ip)
		d.rdb.Expire(ctx, key, 90*24*time.Hour)
	}

	suspiciousKey := fmt.Sprintf("suspicious:%s", ip)
	suspiciousCount, _ := d.rdb.Incr(ctx, suspiciousKey).Result()
	d.rdb.Expire(ctx, suspiciousKey, d.cfg.SuspiciousIPTTL)

	return suspiciousCount > int64(d.cfg.SuspiciousLoginCount)
}

func (d *SuspiciousDetector) IsIPBlocked(ctx context.Context, ip string) bool {
	if d.rdb == nil {
		return false
	}

	key := fmt.Sprintf("blocked_ip:%s", ip)
	exists, _ := d.rdb.Exists(ctx, key).Result()
	return exists > 0
}

func (d *SuspiciousDetector) BlockIP(ctx context.Context, ip string, duration time.Duration) error {
	if d.rdb == nil {
		return nil
	}
	return d.rdb.Set(ctx, fmt.Sprintf("blocked_ip:%s", ip), "1", duration).Err()
}

func (d *SuspiciousDetector) RecordDeviceFingerprint(ctx context.Context, userID, fingerprint string) error {
	if d.rdb == nil || !d.cfg.DeviceFingerprinting {
		return nil
	}

	key := fmt.Sprintf("device_fingerprints:%s", userID)
	pipe := d.rdb.Pipeline()
	pipe.SAdd(ctx, key, fingerprint)
	pipe.Expire(ctx, key, 90*24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (d *SuspiciousDetector) IsDeviceKnown(ctx context.Context, userID, fingerprint string) bool {
	if d.rdb == nil || !d.cfg.DeviceFingerprinting {
		return true
	}

	key := fmt.Sprintf("device_fingerprints:%s", userID)
	exists, err := d.rdb.SIsMember(ctx, key, fingerprint).Result()
	return err == nil && exists
}
