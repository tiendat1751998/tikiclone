package postgres

import (
	"context"
	"fmt"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type ReactionRepo struct {
	pool *Pool
}

func NewReactionRepo(pool *Pool) *ReactionRepo {
	return &ReactionRepo{pool: pool}
}

func (r *ReactionRepo) Save(ctx context.Context, reaction *domain.Reaction) error {
	query := `INSERT INTO reactions (id, room_id, user_id, type, timestamp) VALUES ($1,$2,$3,$4,$5)`
	_, err := r.pool.Exec(ctx, query, reaction.ID, reaction.RoomID, reaction.UserID, reaction.Type, reaction.Timestamp)
	if err != nil {
		return fmt.Errorf("save reaction: %w", err)
	}
	return nil
}

func (r *ReactionRepo) GetCountByRoom(ctx context.Context, roomID string, reactionType string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reactions WHERE room_id=$1 AND type=$2`, roomID, reactionType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count reactions: %w", err)
	}
	return count, nil
}

func (r *ReactionRepo) GetSummaryByRoom(ctx context.Context, roomID string) (map[string]int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT type, COUNT(*) as count FROM reactions WHERE room_id=$1 GROUP BY type`, roomID)
	if err != nil {
		return nil, fmt.Errorf("summary: %w", err)
	}
	defer rows.Close()
	result := make(map[string]int64)
	for rows.Next() {
		var t string
		var c int64
		if err := rows.Scan(&t, &c); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result[t] = c
	}
	return result, nil
}
