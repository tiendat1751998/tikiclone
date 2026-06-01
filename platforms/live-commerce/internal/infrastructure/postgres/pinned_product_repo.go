package postgres

import (
	"context"
	"fmt"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type PinnedProductRepo struct {
	pool *Pool
}

func NewPinnedProductRepo(pool *Pool) *PinnedProductRepo {
	return &PinnedProductRepo{pool: pool}
}

func (r *PinnedProductRepo) Pin(ctx context.Context, pp *domain.PinnedProduct) error {
	query := `INSERT INTO pinned_products (id, livestream_id, product_id, product_name, price, image_url, is_active, pinned_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, query, pp.ID, pp.LivestreamID, pp.ProductID, pp.ProductName, pp.Price, pp.ImageURL, pp.IsActive, pp.PinnedAt)
	if err != nil {
		return fmt.Errorf("pin product: %w", err)
	}
	return nil
}

func (r *PinnedProductRepo) Unpin(ctx context.Context, livestreamID, productID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE pinned_products SET is_active=false WHERE livestream_id=$1 AND product_id=$2`, livestreamID, productID)
	if err != nil {
		return fmt.Errorf("unpin: %w", err)
	}
	return nil
}

func (r *PinnedProductRepo) GetActiveByLivestream(ctx context.Context, livestreamID string) ([]*domain.PinnedProduct, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, livestream_id, product_id, product_name, price, image_url, is_active, pinned_at
		FROM pinned_products WHERE livestream_id=$1 AND is_active=true ORDER BY pinned_at`, livestreamID)
	if err != nil {
		return nil, fmt.Errorf("get pinned: %w", err)
	}
	defer rows.Close()
	var result []*domain.PinnedProduct
	for rows.Next() {
		pp := &domain.PinnedProduct{}
		if err := rows.Scan(&pp.ID, &pp.LivestreamID, &pp.ProductID, &pp.ProductName, &pp.Price, &pp.ImageURL, &pp.IsActive, &pp.PinnedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result = append(result, pp)
	}
	return result, nil
}
