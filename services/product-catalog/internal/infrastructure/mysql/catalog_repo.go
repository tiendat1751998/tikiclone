package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tikiclone/tiki/services/product-catalog/internal/domain"
)

type CatalogRepository struct {
	db *sqlx.DB
}

func NewCatalogRepository(db *sqlx.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

// ─── Product ────────────────────────────────────────────────────────

func (r *CatalogRepository) CreateProduct(ctx context.Context, p *domain.Product) error {
	query := `INSERT INTO products (id, shop_id, name, description, category_id, brand, status, ` +
		`condition, weight, dimensions, metadata, version, created_at, updated_at) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.ShopID, p.Name, p.Description, p.CategoryID, p.Brand,
		p.Status, p.Condition, p.Weight, p.Dimensions, p.Metadata,
		p.Version, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}
	return nil
}

func (r *CatalogRepository) FindProductByID(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	query := `SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &p, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}
	return &p, nil
}

func (r *CatalogRepository) FindProductsByShopID(ctx context.Context, shopID, status string, limit, offset int) ([]*domain.Product, error) {
	var products []*domain.Product
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE shop_id = ? AND status = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{shopID, status, limit, offset}
	} else {
		query = `SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE shop_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{shopID, limit, offset}
	}

	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	return products, nil
}

func (r *CatalogRepository) CountProductsByShopID(ctx context.Context, shopID, status string) (int, error) {
	var count int
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT COUNT(*) FROM products WHERE shop_id = ? AND status = ? AND deleted_at IS NULL`
		args = []interface{}{shopID, status}
	} else {
		query = `SELECT COUNT(*) FROM products WHERE shop_id = ? AND deleted_at IS NULL`
		args = []interface{}{shopID}
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
}

func (r *CatalogRepository) UpdateProduct(ctx context.Context, p *domain.Product) error {
	query := `UPDATE products SET name = ?, description = ?, category_id = ?, brand = ?, ` +
		`status = ?, condition = ?, weight = ?, dimensions = ?, metadata = ?, ` +
		`version = version + 1, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query,
		p.Name, p.Description, p.CategoryID, p.Brand, p.Status, p.Condition,
		p.Weight, p.Dimensions, p.Metadata, time.Now().UTC(), p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrConcurrentModification
	}
	return nil
}

func (r *CatalogRepository) FindProductByIdempotencyKey(ctx context.Context, key string) (*domain.Product, error) {
	var p domain.Product
	if err := r.db.GetContext(ctx, &p, "SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE idempotency_key = ?", key); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// ─── SKU ────────────────────────────────────────────────────────────

func (r *CatalogRepository) CreateSKU(ctx context.Context, sku *domain.SKU) error {
	query := `INSERT INTO skus (id, product_id, sku_code, name, price, compare_price, currency, ` +
		`stock, reserved_stock, weight, dimensions, status, attributes, metadata, version, created_at, updated_at) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		sku.ID, sku.ProductID, sku.SkuCode, sku.Name, sku.Price, sku.ComparePrice,
		sku.Currency, sku.Stock, sku.ReservedStock, sku.Weight, sku.Dimensions,
		sku.Status, sku.Attributes, sku.Metadata, sku.Version, sku.CreatedAt, sku.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) FindSKUsByProductID(ctx context.Context, productID string) ([]domain.SKU, error) {
	var skus []domain.SKU
	err := r.db.SelectContext(ctx, &skus, "SELECT id, product_id, sku_code, name, price, compare_price, currency, stock, reserved_stock, weight, dimensions, status, attributes, metadata, version, created_at, updated_at FROM skus WHERE product_id = ? ORDER BY sort_order ASC LIMIT 500", productID)
	return skus, err
}

// ─── Category ───────────────────────────────────────────────────────

func (r *CatalogRepository) CreateCategory(ctx context.Context, c *domain.Category) error {
	query := `INSERT INTO categories (id, parent_id, name, slug, description, image_url, sort_order, ` +
		`is_active, depth, path, version, created_at, updated_at) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.ParentID, c.Name, c.Slug, c.Description, c.ImageURL,
		c.SortOrder, c.IsActive, c.Depth, c.Path, c.Version, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) FindCategoryByID(ctx context.Context, id string) (*domain.Category, error) {
	var c domain.Category
	if err := r.db.GetContext(ctx, &c, "SELECT id, parent_id, name, slug, description, image_url, sort_order, is_active, depth, path, version, created_at, updated_at FROM categories WHERE id = ?", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CatalogRepository) FindCategoryTree(ctx context.Context, rootID string) ([]*domain.Category, error) {
	var categories []*domain.Category
	var query string
	var args []interface{}

	if rootID != "" {
		// Get all descendants of the root
		query = `SELECT id, parent_id, name, slug, description, image_url, sort_order, is_active, depth, path, version, created_at, updated_at FROM categories WHERE path LIKE ? AND is_active = true ORDER BY depth ASC, sort_order ASC LIMIT 1000`
		args = []interface{}{rootID + "%"}
	} else {
		query = `SELECT id, parent_id, name, slug, description, image_url, sort_order, is_active, depth, path, version, created_at, updated_at FROM categories WHERE is_active = true ORDER BY depth ASC, sort_order ASC LIMIT 1000`
		args = []interface{}{}
	}

	err := r.db.SelectContext(ctx, &categories, query, args...)
	return categories, err
}

// ─── Media ──────────────────────────────────────────────────────────

func (r *CatalogRepository) FindMediaByProductID(ctx context.Context, productID string) ([]domain.Media, error) {
	var media []domain.Media
	err := r.db.SelectContext(ctx, &media,
		"SELECT id, product_id, sku_id, type, url, thumbnail_url, alt_text, sort_order, status, metadata, version, created_at, updated_at FROM media WHERE product_id = ? AND status != ? ORDER BY sort_order ASC LIMIT 100",
		productID, domain.MediaStatusDeleted,
	)
	return media, err
}

// ─── Outbox ─────────────────────────────────────────────────────────

func (r *CatalogRepository) SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.AggregateType, event.AggregateID,
		event.EventType, event.Payload, event.CreatedAt, event.Processed,
	)
	return err
}

func (r *CatalogRepository) GetUnprocessedOutboxEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent
	err := r.db.SelectContext(ctx, &events,
		"SELECT event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed FROM outbox_events WHERE processed = FALSE ORDER BY created_at ASC LIMIT ?",
		limit,
	)
	return events, err
}

func (r *CatalogRepository) MarkOutboxEventProcessed(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET processed = TRUE WHERE event_id = ?", eventID)
	return err
}

// ─── Idempotency ────────────────────────────────────────────────────

func (r *CatalogRepository) SaveIdempotencyKey(ctx context.Context, rec *domain.IdempotencyRecord) error {
	query := `INSERT INTO idempotency_keys (` + "`key`" + `, product_id, expires_at, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, rec.Key, rec.ProductID, rec.ExpiresAt, rec.CreatedAt)
	return err
}

func (r *CatalogRepository) GetIdempotencyKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var rec domain.IdempotencyRecord
	if err := r.db.GetContext(ctx, &rec, "SELECT `key`, product_id, expires_at, created_at FROM idempotency_keys WHERE `key` = ?", key); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}
