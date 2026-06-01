package postgres

import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type ModerationRepo struct {
	pool *Pool
}

func NewModerationRepo(pool *Pool) *ModerationRepo {
	return &ModerationRepo{pool: pool}
}

func (r *ModerationRepo) SaveAction(ctx context.Context, action *domain.ModerationAction) error {
	query := `INSERT INTO moderation_actions (id, room_id, user_id, action, reason, moderated_by, duration_sec, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, query, action.ID, action.RoomID, action.UserID, action.Action, action.Reason, action.ModeratedBy, action.DurationSec, action.CreatedAt)
	if err != nil {
		return fmt.Errorf("save moderation: %w", err)
	}
	return nil
}

func (r *ModerationRepo) IsUserMuted(ctx context.Context, roomID, userID string) (bool, error) {
	query := `SELECT COUNT(*) FROM moderation_actions
		WHERE room_id=$1 AND user_id=$2 AND action=$3 AND created_at > $4`
	var count int64
	threshold := time.Now().Add(-5 * time.Minute)
	err := r.pool.QueryRow(ctx, query, roomID, userID, domain.ModActionMute, threshold).Scan(&count)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("check mute: %w", err)
	}
	return count > 0, nil
}

func (r *ModerationRepo) GetMuteDuration(ctx context.Context, roomID, userID string) (int64, error) {
	query := `SELECT duration_sec FROM moderation_actions
		WHERE room_id=$1 AND user_id=$2 AND action=$3 AND created_at > $4
		ORDER BY created_at DESC LIMIT 1`
	var dur int64
	threshold := time.Now().Add(-30 * time.Minute)
	err := r.pool.QueryRow(ctx, query, roomID, userID, domain.ModActionMute, threshold).Scan(&dur)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("get mute: %w", err)
	}
	return dur, nil
}
