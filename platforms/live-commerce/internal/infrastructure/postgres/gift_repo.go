package postgres

import (
	"context"
	"fmt"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type GiftRepo struct {
	pool *Pool
}

func NewGiftRepo(pool *Pool) *GiftRepo {
	return &GiftRepo{pool: pool}
}

func (r *GiftRepo) Save(ctx context.Context, gift *domain.Gift) error {
	query := `INSERT INTO gifts (id, room_id, user_id, username, gift_type, amount, currency, timestamp)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, query, gift.ID, gift.RoomID, gift.UserID, gift.Username, gift.GiftType, gift.Amount, gift.Currency, gift.Timestamp)
	if err != nil {
		return fmt.Errorf("save gift: %w", err)
	}
	return nil
}

func (r *GiftRepo) GetLeaderboardByRoom(ctx context.Context, roomID string, limit int) ([]*domain.GiftLeaderboardEntry, error) {
	query := `SELECT user_id, username, SUM(amount) as total FROM gifts
		WHERE room_id=$1 GROUP BY user_id, username ORDER BY total DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("leaderboard: %w", err)
	}
	defer rows.Close()
	var result []*domain.GiftLeaderboardEntry
	rank := 1
	for rows.Next() {
		entry := &domain.GiftLeaderboardEntry{Rank: rank}
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.Total); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result = append(result, entry)
		rank++
	}
	return result, nil
}
