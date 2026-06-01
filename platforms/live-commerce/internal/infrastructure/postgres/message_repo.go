package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type MessageRepo struct {
	pool *Pool
}

func NewMessageRepo(pool *Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

func (r *MessageRepo) Save(ctx context.Context, msg *domain.ChatMessage) error {
	query := `INSERT INTO chat_messages (id, room_id, user_id, username, content, type, is_moderated, sequence, timestamp)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.pool.Exec(ctx, query,
		msg.ID, msg.RoomID, msg.UserID, msg.Username, msg.Content, msg.Type, msg.IsModerated, msg.Sequence, msg.Timestamp)
	if err != nil {
		return fmt.Errorf("save message: %w", err)
	}
	return nil
}

func (r *MessageRepo) GetByRoom(ctx context.Context, roomID string, offset, limit int) ([]*domain.ChatMessage, int64, error) {
	countQuery := `SELECT COUNT(*) FROM chat_messages WHERE room_id=$1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, roomID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}
	query := `SELECT id, room_id, user_id, username, content, type, is_moderated, sequence, timestamp
		FROM chat_messages WHERE room_id=$1 ORDER BY sequence DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()
	var result []*domain.ChatMessage
	for rows.Next() {
		msg := &domain.ChatMessage{}
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.Content, &msg.Type, &msg.IsModerated, &msg.Sequence, &msg.Timestamp); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		result = append(result, msg)
	}
	return result, total, nil
}

func (r *MessageRepo) MarkModerated(ctx context.Context, messageID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE chat_messages SET is_moderated=true WHERE id=$1`, messageID)
	if err != nil {
		return fmt.Errorf("mark moderated: %w", err)
	}
	return nil
}

func (r *MessageRepo) Delete(ctx context.Context, messageID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM chat_messages WHERE id=$1`, messageID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	return nil
}

func (r *MessageRepo) GetLastSequence(ctx context.Context, roomID string) (int64, error) {
	var seq int64
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(sequence), 0) FROM chat_messages WHERE room_id=$1`, roomID).Scan(&seq)
	if err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("get last sequence: %w", err)
	}
	return seq, nil
}
