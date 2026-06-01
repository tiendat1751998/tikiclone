package mysql
import ("context"; "database/sql"; "github.com/jmoiron/sqlx"; "github.com/tikiclone/tiki/services/product-catalog/internal/domain")

type ProductRepository struct{ db *sqlx.DB }
func NewProductRepository(db *sqlx.DB) *ProductRepository { return &ProductRepository{db: db} }
func (r *ProductRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product; err := r.db.GetContext(ctx, &p, "SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE id = ? AND deleted_at IS NULL", id)
	if err == sql.ErrNoRows { return nil, nil }; if err != nil { return nil, err }; return &p, nil
}
func (r *ProductRepository) FindByShopID(ctx context.Context, shopID string, offset, limit int) ([]*domain.Product, int64, error) {
	var total int64; r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM products WHERE shop_id = ? AND deleted_at IS NULL", shopID)
	var products []*domain.Product; err := r.db.SelectContext(ctx, &products, "SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE shop_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?", shopID, limit, offset)
	return products, total, err
}
func (r *ProductRepository) FindByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*domain.Product, int64, error) {
	var total int64; r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM products WHERE category_id = ? AND status = 'active' AND deleted_at IS NULL", categoryID)
	var products []*domain.Product; err := r.db.SelectContext(ctx, &products, "SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE category_id = ? AND status = 'active' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?", categoryID, limit, offset)
	return products, total, err
}
func (r *ProductRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Product, error) {
	var p domain.Product; err := r.db.GetContext(ctx, &p, "SELECT id, shop_id, name, description, category_id, brand, idempotency_key, status, condition, weight, dimensions, metadata, currency, version, created_at, updated_at, deleted_at FROM products WHERE idempotency_key = ? AND deleted_at IS NULL", key)
	if err == sql.ErrNoRows { return nil, nil }; if err != nil { return nil, err }; return &p, nil
}
func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `INSERT INTO products (id, shop_id, name, description, category_id, status, currency, idempotency_key, version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, p.ID, p.ShopID, p.Name, p.Description, p.CategoryID, p.Status, p.Currency, p.IdempotencyKey, p.Version, p.CreatedAt, p.UpdatedAt); return err
}
func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) error {
	query := `UPDATE products SET name = ?, description = ?, category_id = ?, status = ?, version = version + 1, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, p.Name, p.Description, p.CategoryID, p.Status, p.UpdatedAt, p.ID); return err
}
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE products SET deleted_at = NOW() WHERE id = ?", id); return err
}

type SKURepository struct{ db *sqlx.DB }
func NewSKURepository(db *sqlx.DB) *SKURepository { return &SKURepository{db: db} }
func (r *SKURepository) FindByID(ctx context.Context, id string) (*domain.SKU, error) {
	var s domain.SKU; err := r.db.GetContext(ctx, &s, "SELECT id, product_id, sku_code, name, price, compare_price, currency, stock, reserved_stock, weight, dimensions, status, attributes, metadata, version, created_at, updated_at FROM skus WHERE id = ?", id)
	if err == sql.ErrNoRows { return nil, nil }; return &s, err
}
func (r *SKURepository) FindByProductID(ctx context.Context, productID string) ([]*domain.SKU, error) {
	var skus []*domain.SKU; err := r.db.SelectContext(ctx, &skus, "SELECT id, product_id, sku_code, name, price, compare_price, currency, stock, reserved_stock, weight, dimensions, status, attributes, metadata, version, created_at, updated_at FROM skus WHERE product_id = ? ORDER BY sort_order LIMIT 500", productID)
	return skus, err
}
func (r *SKURepository) Create(ctx context.Context, sku *domain.SKU) error {
	query := `INSERT INTO skus (id, product_id, sku_code, attributes, price, compare_price, stock, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, sku.ID, sku.ProductID, sku.SkuCode, sku.Attributes, sku.Price, sku.ComparePrice, sku.Stock, sku.Status, sku.CreatedAt, sku.UpdatedAt); return err
}
func (r *SKURepository) Update(ctx context.Context, sku *domain.SKU) error {
	query := `UPDATE skus SET attributes = ?, price = ?, compare_price = ?, stock = ?, status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, sku.Attributes, sku.Price, sku.ComparePrice, sku.Stock, sku.Status, sku.UpdatedAt, sku.ID); return err
}
func (r *SKURepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM skus WHERE id = ?", id); return err
}

type CategoryRepository struct{ db *sqlx.DB }
func NewCategoryRepository(db *sqlx.DB) *CategoryRepository { return &CategoryRepository{db: db} }
func (r *CategoryRepository) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	var c domain.Category; err := r.db.GetContext(ctx, &c, "SELECT id, parent_id, name, slug, description, image_url, sort_order, is_active, depth, path, version, created_at, updated_at FROM categories WHERE id = ?", id)
	if err == sql.ErrNoRows { return nil, nil }; return &c, err
}
func (r *CategoryRepository) FindByParentID(ctx context.Context, parentID string) ([]*domain.Category, error) {
	var cats []*domain.Category; err := r.db.SelectContext(ctx, &cats, "SELECT id, parent_id, name, slug, description, image_url, sort_order, is_active, depth, path, version, created_at, updated_at FROM categories WHERE parent_id = ? AND is_active = true ORDER BY sort_order", parentID)
	return cats, err
}
func (r *CategoryRepository) GetTree(ctx context.Context) ([]*domain.Category, error) {
	var cats []*domain.Category; err := r.db.SelectContext(ctx, &cats, "SELECT id, parent_id, name, slug, description, image_url, sort_order, is_active, depth, path, version, created_at, updated_at FROM categories WHERE is_active = true ORDER BY depth, sort_order LIMIT 1000")
	return cats, err
}
func (r *CategoryRepository) Create(ctx context.Context, c *domain.Category) error {
	query := `INSERT INTO categories (id, parent_id, name, slug, description, image_url, depth, sort_order, is_active, path, version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.ParentID, c.Name, c.Slug, c.Description, c.ImageURL, c.Depth, c.SortOrder, c.IsActive, c.Path, c.Version, c.CreatedAt, c.UpdatedAt); return err
}
func (r *CategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	query := `UPDATE categories SET name = ?, slug = ?, sort_order = ?, is_active = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, c.Name, c.Slug, c.SortOrder, c.IsActive, c.UpdatedAt, c.ID); return err
}
func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM categories WHERE id = ?", id); return err
}

type AttributeRepository struct{ db *sqlx.DB }
func NewAttributeRepository(db *sqlx.DB) *AttributeRepository { return &AttributeRepository{db: db} }
func (r *AttributeRepository) FindByCategoryID(ctx context.Context, categoryID string) ([]*domain.Attribute, error) {
	var attrs []*domain.Attribute; err := r.db.SelectContext(ctx, &attrs, "SELECT id, category_id, name, slug, type, is_required, is_filterable, is_searchable, options, sort_order, created_at, updated_at FROM attributes WHERE category_id = ? AND is_active = true ORDER BY sort_order LIMIT 200", categoryID)
	return attrs, err
}
func (r *AttributeRepository) Create(ctx context.Context, a *domain.Attribute) error {
	query := `INSERT INTO attributes (id, category_id, name, slug, type, is_required, is_filterable, is_searchable, options, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, a.ID, a.CategoryID, a.Name, a.Slug, a.Type, a.IsRequired, a.IsFilterable, a.IsSearchable, a.Options, a.SortOrder, a.CreatedAt, a.UpdatedAt); return err
}
func (r *AttributeRepository) Update(ctx context.Context, a *domain.Attribute) error {
	query := `UPDATE attributes SET name = ?, slug = ?, type = ?, is_required = ?, is_filterable = ?, is_searchable = ?, options = ?, sort_order = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, a.Name, a.Slug, a.Type, a.IsRequired, a.IsFilterable, a.IsSearchable, a.Options, a.SortOrder, a.UpdatedAt, a.ID); return err
}

type ProductMediaRepository struct{ db *sqlx.DB }
func NewProductMediaRepository(db *sqlx.DB) *ProductMediaRepository { return &ProductMediaRepository{db: db} }
func (r *ProductMediaRepository) FindByProductID(ctx context.Context, productID string) ([]*domain.ProductMedia, error) {
	var media []*domain.ProductMedia; err := r.db.SelectContext(ctx, &media, "SELECT id, product_id, media_type, url, thumbnail, sort_order, is_primary, created_at FROM product_media WHERE product_id = ? ORDER BY sort_order LIMIT 100", productID)
	return media, err
}
func (r *ProductMediaRepository) Create(ctx context.Context, m *domain.ProductMedia) error {
	query := `INSERT INTO product_media (id, product_id, media_type, url, thumbnail, sort_order, is_primary, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.ProductID, m.MediaType, m.URL, m.Thumbnail, m.SortOrder, m.IsPrimary, m.CreatedAt); return err
}
func (r *ProductMediaRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM product_media WHERE id = ?", id); return err
}
