package postgres

import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type LivestreamRepo struct {
	pool *Pool
}

func NewLivestreamRepo(pool *Pool) *LivestreamRepo {
	return &LivestreamRepo{pool: pool}
}

func (r *LivestreamRepo) Create(ctx context.Context, ls *domain.Livestream) error {
	query := `INSERT INTO livestreams (id, seller_id, title, description, cover_url, status, category, tags, viewer_count, peak_viewers, total_likes, total_gifts, total_shares, scheduled_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`
	_, err := r.pool.Exec(ctx, query,
		ls.ID, ls.SellerID, ls.Title, ls.Description, ls.CoverURL, ls.Status, ls.Category, ls.Tags,
		ls.ViewerCount, ls.PeakViewers, ls.TotalLikes, ls.TotalGifts, ls.TotalShares,
		ls.ScheduledAt, ls.CreatedAt, ls.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create livestream: %w", err)
	}
	return nil
}

func (r *LivestreamRepo) GetByID(ctx context.Context, id string) (*domain.Livestream, error) {
	query := `SELECT id, seller_id, title, description, cover_url, status, category, tags, viewer_count, peak_viewers, total_likes, total_gifts, total_shares, started_at, ended_at, scheduled_at, created_at, updated_at
		FROM livestreams WHERE id=$1`
	ls := &domain.Livestream{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ls.ID, &ls.SellerID, &ls.Title, &ls.Description, &ls.CoverURL, &ls.Status, &ls.Category, &ls.Tags,
		&ls.ViewerCount, &ls.PeakViewers, &ls.TotalLikes, &ls.TotalGifts, &ls.TotalShares,
		&ls.StartedAt, &ls.EndedAt, &ls.ScheduledAt, &ls.CreatedAt, &ls.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrLivestreamNotFound
		}
		return nil, fmt.Errorf("get livestream: %w", err)
	}
	return ls, nil
}

func (r *LivestreamRepo) Update(ctx context.Context, ls *domain.Livestream) error {
	ls.UpdatedAt = time.Now()
	query := `UPDATE livestreams SET title=$2, description=$3, cover_url=$4, status=$5, category=$6, tags=$7,
		viewer_count=$8, peak_viewers=$9, total_likes=$10, total_gifts=$11, total_shares=$12,
		started_at=$13, ended_at=$14, updated_at=$15 WHERE id=$1`
	_, err := r.pool.Exec(ctx, query,
		ls.ID, ls.Title, ls.Description, ls.CoverURL, ls.Status, ls.Category, ls.Tags,
		ls.ViewerCount, ls.PeakViewers, ls.TotalLikes, ls.TotalGifts, ls.TotalShares,
		ls.StartedAt, ls.EndedAt, ls.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update livestream: %w", err)
	}
	return nil
}

func (r *LivestreamRepo) ListBySeller(ctx context.Context, sellerID string, offset, limit int) ([]*domain.Livestream, int64, error) {
	countQuery := `SELECT COUNT(*) FROM livestreams WHERE seller_id=$1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, sellerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}
	query := `SELECT id, seller_id, title, description, cover_url, status, category, tags, viewer_count, peak_viewers, total_likes, total_gifts, total_shares, started_at, ended_at, scheduled_at, created_at, updated_at
		FROM livestreams WHERE seller_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, sellerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()
	var result []*domain.Livestream
	for rows.Next() {
		ls := &domain.Livestream{}
		if err := rows.Scan(&ls.ID, &ls.SellerID, &ls.Title, &ls.Description, &ls.CoverURL, &ls.Status, &ls.Category, &ls.Tags,
			&ls.ViewerCount, &ls.PeakViewers, &ls.TotalLikes, &ls.TotalGifts, &ls.TotalShares,
			&ls.StartedAt, &ls.EndedAt, &ls.ScheduledAt, &ls.CreatedAt, &ls.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		result = append(result, ls)
	}
	return result, total, nil
}

func (r *LivestreamRepo) ListActive(ctx context.Context, offset, limit int) ([]*domain.Livestream, int64, error) {
	countQuery := `SELECT COUNT(*) FROM livestreams WHERE status=$1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, domain.LiveStatusLive).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}
	query := `SELECT id, seller_id, title, description, cover_url, status, category, tags, viewer_count, peak_viewers, total_likes, total_gifts, total_shares, started_at, ended_at, scheduled_at, created_at, updated_at
		FROM livestreams WHERE status=$1 ORDER BY viewer_count DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, domain.LiveStatusLive, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list active: %w", err)
	}
	defer rows.Close()
	var result []*domain.Livestream
	for rows.Next() {
		ls := &domain.Livestream{}
		if err := rows.Scan(&ls.ID, &ls.SellerID, &ls.Title, &ls.Description, &ls.CoverURL, &ls.Status, &ls.Category, &ls.Tags,
			&ls.ViewerCount, &ls.PeakViewers, &ls.TotalLikes, &ls.TotalGifts, &ls.TotalShares,
			&ls.StartedAt, &ls.EndedAt, &ls.ScheduledAt, &ls.CreatedAt, &ls.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		result = append(result, ls)
	}
	return result, total, nil
}
