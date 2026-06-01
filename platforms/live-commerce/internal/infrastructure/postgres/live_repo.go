package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type LiveRepository struct {
	pool *pgxpool.Pool
}

func NewLiveRepository(pool *pgxpool.Pool) *LiveRepository {
	return &LiveRepository{pool: pool}
}

func (r *LiveRepository) FindByID(ctx context.Context, id string) (*domain.Livestream, error) {
	query := `SELECT id, seller_id, title, COALESCE(description,'') as description, status, 
		viewer_count, peak_viewers, total_likes, total_gifts, started_at, ended_at, 
		scheduled_at, created_at, updated_at FROM livestreams WHERE id = $1`
	
	row := r.pool.QueryRow(ctx, query, id)
	ls := &domain.Livestream{}
	err := row.Scan(&ls.ID, &ls.SellerID, &ls.Title, &ls.Description, &ls.Status,
		&ls.ViewerCount, &ls.PeakViewers, &ls.TotalLikes, &ls.TotalGifts,
		&ls.StartedAt, &ls.EndedAt, &ls.ScheduledAt, &ls.CreatedAt, &ls.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return ls, err
}

func (r *LiveRepository) FindBySellerID(ctx context.Context, sellerID string, status string, limit, offset int) ([]*domain.Livestream, error) {
	query := `SELECT id, seller_id, title, COALESCE(description,''), status, 
		viewer_count, peak_viewers, total_likes, total_gifts, started_at, ended_at, 
		scheduled_at, created_at, updated_at FROM livestreams 
		WHERE seller_id = $1 AND ($2 = '' OR status = $2)
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	
	rows, err := r.pool.Query(ctx, query, sellerID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLivestreams(rows)
}

func (r *LiveRepository) FindActive(ctx context.Context, limit, offset int) ([]*domain.Livestream, error) {
	query := `SELECT id, seller_id, title, COALESCE(description,''), status, 
		viewer_count, peak_viewers, total_likes, total_gifts, started_at, ended_at, 
		scheduled_at, created_at, updated_at FROM livestreams 
		WHERE status = 'live' ORDER BY viewer_count DESC LIMIT $1 OFFSET $2`
	
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLivestreams(rows)
}

func (r *LiveRepository) Create(ctx context.Context, ls *domain.Livestream) error {
	query := `INSERT INTO livestreams (id, seller_id, title, description, status, 
		viewer_count, peak_viewers, total_likes, total_gifts, started_at, ended_at, 
		scheduled_at, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	_, err := r.pool.Exec(ctx, query, ls.ID, ls.SellerID, ls.Title, ls.Description, ls.Status,
		ls.ViewerCount, ls.PeakViewers, ls.TotalLikes, ls.TotalGifts,
		ls.StartedAt, ls.EndedAt, ls.ScheduledAt, ls.CreatedAt, ls.UpdatedAt)
	return err
}

func (r *LiveRepository) Update(ctx context.Context, ls *domain.Livestream) error {
	query := `UPDATE livestreams SET status=$1, viewer_count=$2, peak_viewers=$3, 
		total_likes=$4, total_gifts=$5, started_at=$6, ended_at=$7, updated_at=$8 
		WHERE id=$9`
	
	_, err := r.pool.Exec(ctx, query, ls.Status, ls.ViewerCount, ls.PeakViewers,
		ls.TotalLikes, ls.TotalGifts, ls.StartedAt, ls.EndedAt, ls.UpdatedAt, ls.ID)
	return err
}

func scanLivestreams(rows pgx.Rows) ([]*domain.Livestream, error) {
	var result []*domain.Livestream
	for rows.Next() {
		ls := &domain.Livestream{}
		if err := rows.Scan(&ls.ID, &ls.SellerID, &ls.Title, &ls.Description, &ls.Status,
			&ls.ViewerCount, &ls.PeakViewers, &ls.TotalLikes, &ls.TotalGifts,
			&ls.StartedAt, &ls.EndedAt, &ls.ScheduledAt, &ls.CreatedAt, &ls.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan livestream: %w", err)
		}
		result = append(result, ls)
	}
	return result, nil
}
