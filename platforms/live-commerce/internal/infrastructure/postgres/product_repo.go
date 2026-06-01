package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type ProductRepository struct {
	pool *Pool
}

func NewProductRepository(pool *Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) FindByLivestream(ctx context.Context, livestreamID string) ([]*domain.PinnedProduct, error) {
	query := `SELECT id, livestream_id, product_id, product_name, price, image_url, is_active, pinned_at 
		FROM pinned_products WHERE livestream_id = $1 ORDER BY pinned_at DESC`
	rows, err := r.pool.Query(ctx, query, livestreamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.PinnedProduct
	for rows.Next() {
		pp := &domain.PinnedProduct{}
		if err := rows.Scan(&pp.ID, &pp.LivestreamID, &pp.ProductID, &pp.ProductName,
			&pp.Price, &pp.ImageURL, &pp.IsActive, &pp.PinnedAt); err != nil {
			return nil, err
		}
		result = append(result, pp)
	}
	return result, nil
}

func (r *ProductRepository) FindActive(ctx context.Context, livestreamID string) (*domain.PinnedProduct, error) {
	query := `SELECT id, livestream_id, product_id, product_name, price, image_url, is_active, pinned_at 
		FROM pinned_products WHERE livestream_id = $1 AND is_active = true LIMIT 1`
	pp := &domain.PinnedProduct{}
	err := r.pool.QueryRow(ctx, query, livestreamID).Scan(
		&pp.ID, &pp.LivestreamID, &pp.ProductID, &pp.ProductName,
		&pp.Price, &pp.ImageURL, &pp.IsActive, &pp.PinnedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return pp, err
}

func (r *ProductRepository) Create(ctx context.Context, pp *domain.PinnedProduct) error {
	query := `INSERT INTO pinned_products (id, livestream_id, product_id, product_name, price, image_url, is_active, pinned_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query, pp.ID, pp.LivestreamID, pp.ProductID, pp.ProductName,
		pp.Price, pp.ImageURL, pp.IsActive, pp.PinnedAt)
	return err
}

func (r *ProductRepository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE pinned_products SET is_active = false WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

type ModerationRepository struct {
	pool *Pool
}

func NewModerationRepository(pool *Pool) *ModerationRepository {
	return &ModerationRepository{pool: pool}
}

func (r *ModerationRepository) CreateAction(ctx context.Context, action *domain.ModerationAction) error {
	query := `INSERT INTO moderation_actions (id, room_id, user_id, action, reason, moderated_by, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, query, action.ID, action.RoomID, action.UserID,
		action.Action, action.Reason, action.ModeratedBy, action.CreatedAt)
	return err
}

func (r *ModerationRepository) IsMuted(ctx context.Context, roomID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM moderation_actions 
		WHERE room_id = $1 AND user_id = $2 AND action = 'mute' 
		AND created_at > $3)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, roomID, userID, time.Now().Add(-30*time.Minute)).Scan(&exists)
	return exists, err
}

func (r *ModerationRepository) FindActionsByUser(ctx context.Context, roomID, userID string, limit int) ([]*domain.ModerationAction, error) {
	query := `SELECT id, room_id, user_id, action, reason, moderated_by, created_at 
		FROM moderation_actions WHERE room_id = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, query, roomID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.ModerationAction
	for rows.Next() {
		a := &domain.ModerationAction{}
		if err := rows.Scan(&a.ID, &a.RoomID, &a.UserID, &a.Action, &a.Reason, &a.ModeratedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

type RoomRepository struct {
	pool *Pool
}

func NewRoomRepository(pool *Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) FindByLivestreamID(ctx context.Context, livestreamID string) (*domain.Room, error) {
	query := `SELECT id, livestream_id, status, viewer_count, created_at FROM rooms WHERE livestream_id = $1`
	room := &domain.Room{}
	err := r.pool.QueryRow(ctx, query, livestreamID).Scan(&room.ID, &room.LivestreamID, &room.Status, &room.ViewerCount, &room.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return room, err
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (id, livestream_id, status, viewer_count, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, room.ID, room.LivestreamID, room.Status, room.ViewerCount, room.CreatedAt)
	return err
}

func (r *RoomRepository) Update(ctx context.Context, room *domain.Room) error {
	query := `UPDATE rooms SET status = $1, viewer_count = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, room.Status, room.ViewerCount, room.ID)
	return err
}
