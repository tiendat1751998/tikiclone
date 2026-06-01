package mysql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/cart/internal/domain"
)

type CartRepository struct {
	db *sqlx.DB
}

func NewCartRepository(db *sqlx.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) FindByID(ctx context.Context, id string) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.GetContext(ctx, &cart, 	"SELECT id, user_id, session_id, status, currency, item_count, subtotal, version, expires_at, created_at, updated_at FROM carts WHERE id = ? AND deleted_at IS NULL", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) FindByUserID(ctx context.Context, userID string) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.GetContext(ctx, &cart, 	"SELECT id, user_id, session_id, status, currency, item_count, subtotal, version, expires_at, created_at, updated_at FROM carts WHERE user_id = ? AND status = 'active' AND deleted_at IS NULL", userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) FindBySessionID(ctx context.Context, sessionID string) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.GetContext(ctx, &cart, 	"SELECT id, user_id, session_id, status, currency, item_count, subtotal, version, expires_at, created_at, updated_at FROM carts WHERE session_id = ? AND status = 'active' AND deleted_at IS NULL", sessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) Create(ctx context.Context, cart *domain.Cart) error {
	query := `INSERT INTO carts (id, user_id, session_id, status, currency, item_count, subtotal, version, expires_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, cart.ID, cart.UserID, cart.SessionID, cart.Status, cart.Currency, cart.ItemCount, cart.Subtotal, cart.Version, cart.ExpiresAt, cart.CreatedAt, cart.UpdatedAt)
	return err
}

func (r *CartRepository) Update(ctx context.Context, cart *domain.Cart) error {
	result, err := r.db.ExecContext(ctx, "UPDATE carts SET status = ?, item_count = ?, subtotal = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ? AND deleted_at IS NULL", cart.Status, cart.ItemCount, cart.Subtotal, cart.UpdatedAt, cart.ID, cart.Version)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrCartNotFound
	}
	cart.Version++
	return nil
}

func (r *CartRepository) UpdateInTx(ctx context.Context, tx *sql.Tx, cart *domain.Cart) error {
	result, err := tx.ExecContext(ctx, "UPDATE carts SET status = ?, item_count = ?, subtotal = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ? AND deleted_at IS NULL", cart.Status, cart.ItemCount, cart.Subtotal, cart.UpdatedAt, cart.ID, cart.Version)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrCartNotFound
	}
	cart.Version++
	return nil
}

func (r *CartRepository) FindByIDForUpdate(ctx context.Context, tx *sql.Tx, id string) (*domain.Cart, error) {
	var cart domain.Cart
	err := tx.QueryRowContext(ctx, "SELECT id, user_id, session_id, status, currency, item_count, subtotal, version, expires_at, created_at, updated_at FROM carts WHERE id = ? AND deleted_at IS NULL FOR UPDATE", id).Scan(
		&cart.ID, &cart.UserID, &cart.SessionID, &cart.Status, &cart.Currency,
		&cart.ItemCount, &cart.Subtotal, &cart.Version, &cart.ExpiresAt, &cart.CreatedAt, &cart.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
}

func (r *CartRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE carts SET deleted_at = NOW() WHERE id = ?", id)
	return err
}

func (r *CartRepository) FindExpired(ctx context.Context, before string, limit int) ([]*domain.Cart, error) {
	var carts []*domain.Cart
	err := r.db.SelectContext(ctx, &carts, 	"SELECT id, user_id, session_id, status, currency, item_count, subtotal, version, expires_at, created_at, updated_at FROM carts WHERE status = 'active' AND expires_at < ? AND deleted_at IS NULL LIMIT ?", before, limit)
	return carts, err
}

type CartItemRepository struct {
	db *sqlx.DB
}

func NewCartItemRepository(db *sqlx.DB) *CartItemRepository {
	return &CartItemRepository{db: db}
}

func (r *CartItemRepository) FindByID(ctx context.Context, id string) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.GetContext(ctx, &item, 	"SELECT id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, attributes, is_selected, is_available, added_at, updated_at FROM cart_items WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartItemRepository) FindByCartID(ctx context.Context, cartID string) ([]*domain.CartItem, error) {
	var items []*domain.CartItem
	err := r.db.SelectContext(ctx, &items, 	"SELECT id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, attributes, is_selected, is_available, added_at, updated_at FROM cart_items WHERE cart_id = ? ORDER BY added_at DESC LIMIT 500", cartID)
	return items, err
}

func (r *CartItemRepository) FindByCartAndSKU(ctx context.Context, cartID, sku string) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.GetContext(ctx, &item, 	"SELECT id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, attributes, is_selected, is_available, added_at, updated_at FROM cart_items WHERE cart_id = ? AND sku = ?", cartID, sku)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartItemRepository) FindByCartIDInTx(ctx context.Context, tx *sql.Tx, cartID string) ([]*domain.CartItem, error) {
	var items []*domain.CartItem
	rows, err := tx.QueryContext(ctx, "SELECT id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, attributes, is_selected, is_available, added_at, updated_at FROM cart_items WHERE cart_id = ? ORDER BY added_at ASC FOR UPDATE", cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(&item.ID, &item.CartID, &item.SKU, &item.ProductName, &item.ShopID, &item.ShopName, &item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.ImageURL, &item.Attributes, &item.IsSelected, &item.IsAvailable, &item.AddedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *CartItemRepository) CreateInTx(ctx context.Context, tx *sql.Tx, item *domain.CartItem) error {
	query := `INSERT INTO cart_items (id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, attributes, is_selected, is_available, added_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, item.ID, item.CartID, item.SKU, item.ProductName, item.ShopID, item.ShopName, item.Quantity, item.UnitPrice, item.TotalPrice, item.ImageURL, item.Attributes, item.IsSelected, item.IsAvailable, item.AddedAt, item.UpdatedAt)
	return err
}

func (r *CartItemRepository) UpdateInTx(ctx context.Context, tx *sql.Tx, item *domain.CartItem) error {
	query := `UPDATE cart_items SET quantity = ?, total_price = ?, is_selected = ?, is_available = ?, updated_at = ? WHERE id = ?`
	_, err := tx.ExecContext(ctx, query, item.Quantity, item.TotalPrice, item.IsSelected, item.IsAvailable, item.UpdatedAt, item.ID)
	return err
}

func (r *CartItemRepository) DeleteInTx(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM cart_items WHERE id = ?", id)
	return err
}

func (r *CartItemRepository) DeleteByCartIDInTx(ctx context.Context, tx *sql.Tx, cartID string) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM cart_items WHERE cart_id = ?", cartID)
	return err
}

func (r *CartItemRepository) CountByCartIDInTx(ctx context.Context, tx *sql.Tx, cartID string) (int, error) {
	var count int
	err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM cart_items WHERE cart_id = ? FOR UPDATE", cartID).Scan(&count)
	return count, err
}

func (r *CartItemRepository) FindByCartAndSKUInTx(ctx context.Context, tx *sql.Tx, cartID, sku string) (*domain.CartItem, error) {
	var item domain.CartItem
	err := tx.QueryRowContext(ctx, "SELECT id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, attributes, is_selected, is_available, added_at, updated_at FROM cart_items WHERE cart_id = ? AND sku = ? FOR UPDATE", cartID, sku).Scan(
		&item.ID, &item.CartID, &item.SKU, &item.ProductName, &item.ShopID, &item.ShopName,
		&item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.ImageURL, &item.Attributes,
		&item.IsSelected, &item.IsAvailable, &item.AddedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartItemRepository) Create(ctx context.Context, item *domain.CartItem) error {
	query := `INSERT INTO cart_items (id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, attributes, is_selected, is_available, added_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, item.ID, item.CartID, item.SKU, item.ProductName, item.ShopID, item.ShopName, item.Quantity, item.UnitPrice, item.TotalPrice, item.ImageURL, item.Attributes, item.IsSelected, item.IsAvailable, item.AddedAt, item.UpdatedAt)
	return err
}

func (r *CartItemRepository) Update(ctx context.Context, item *domain.CartItem) error {
	query := `UPDATE cart_items SET quantity = ?, total_price = ?, is_selected = ?, is_available = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, item.Quantity, item.TotalPrice, item.IsSelected, item.IsAvailable, item.UpdatedAt, item.ID)
	return err
}

func (r *CartItemRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM cart_items WHERE id = ?", id)
	return err
}

func (r *CartItemRepository) DeleteByCartID(ctx context.Context, cartID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM cart_items WHERE cart_id = ?", cartID)
	return err
}

func (r *CartItemRepository) CountByCartID(ctx context.Context, cartID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM cart_items WHERE cart_id = ?", cartID)
	return count, err
}

type CartSnapshotRepository struct {
	db *sqlx.DB
}

func NewCartSnapshotRepository(db *sqlx.DB) *CartSnapshotRepository {
	return &CartSnapshotRepository{db: db}
}

func (r *CartSnapshotRepository) FindByID(ctx context.Context, id string) (*domain.CartSnapshot, error) {
	var snap domain.CartSnapshot
	err := r.db.GetContext(ctx, &snap, 	"SELECT id, cart_id, user_id, items, seller_groups, subtotal, item_count, currency, idempotency_key, expires_at, created_at FROM cart_snapshots WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *CartSnapshotRepository) FindByCartID(ctx context.Context, cartID string) (*domain.CartSnapshot, error) {
	var snap domain.CartSnapshot
	err := r.db.GetContext(ctx, &snap, 	"SELECT id, cart_id, user_id, items, seller_groups, subtotal, item_count, currency, idempotency_key, expires_at, created_at FROM cart_snapshots WHERE cart_id = ? ORDER BY created_at DESC LIMIT 1", cartID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *CartSnapshotRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.CartSnapshot, error) {
	var snap domain.CartSnapshot
	err := r.db.GetContext(ctx, &snap, 	"SELECT id, cart_id, user_id, items, seller_groups, subtotal, item_count, currency, idempotency_key, expires_at, created_at FROM cart_snapshots WHERE idempotency_key = ?", key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *CartSnapshotRepository) Create(ctx context.Context, snapshot *domain.CartSnapshot) error {
	query := `INSERT INTO cart_snapshots (id, cart_id, user_id, items, seller_groups, subtotal, item_count, currency, idempotency_key, expires_at, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, snapshot.ID, snapshot.CartID, snapshot.UserID, snapshot.Items, snapshot.SellerGroups, snapshot.Subtotal, snapshot.ItemCount, snapshot.Currency, snapshot.IdempotencyKey, snapshot.ExpiresAt, snapshot.CreatedAt)
	return err
}

func (r *CartSnapshotRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM cart_snapshots WHERE id = ?", id)
	return err
}

type CartMergeHistoryRepository struct {
	db *sqlx.DB
}

func NewCartMergeHistoryRepository(db *sqlx.DB) *CartMergeHistoryRepository {
	return &CartMergeHistoryRepository{db: db}
}

func (r *CartMergeHistoryRepository) Create(ctx context.Context, history *domain.CartMergeHistory) error {
	query := `INSERT INTO cart_merge_history (id, source_cart_id, target_cart_id, user_id, merge_type, items_merged, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, history.ID, history.SourceCartID, history.TargetCartID, history.UserID, history.MergeType, history.ItemsMerged, history.CreatedAt)
	return err
}

func (r *CartMergeHistoryRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]*domain.CartMergeHistory, error) {
	var history []*domain.CartMergeHistory
	err := r.db.SelectContext(ctx, &history, "SELECT id, source_cart_id, target_cart_id, user_id, merge_type, items_merged, created_at FROM cart_merge_history WHERE user_id = ? ORDER BY created_at DESC LIMIT ?", userID, limit)
	return history, err
}
