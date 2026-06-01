package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

const categoriesCollection = "categories"

type CategoryRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewCategoryRepository(client *mongo.Client, dbName string) *CategoryRepository {
	db := client.Database(dbName)
	return &CategoryRepository{
		db:         db,
		collection: db.Collection(categoriesCollection),
	}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "repository.category.create")
	defer span.End()

	if category.CategoryID == "" {
		category.CategoryID = uuid.New().String()
	}

	span.SetAttributes(attribute.String("category_id", category.CategoryID))

	_, err := r.collection.InsertOne(ctx, category)
	if err != nil {
		observability.LogWithTrace(ctx).Error("failed to insert category",
			zap.Error(err),
			zap.String("category_id", category.CategoryID),
		)
		return err
	}

	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "repository.category.get_by_id")
	defer span.End()

	span.SetAttributes(attribute.String("category_id", categoryID))

	var category domain.Category
	err := r.collection.FindOne(ctx, bson.M{"category_id": categoryID}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

func (r *CategoryRepository) List(ctx context.Context, parentID string, level int32) ([]domain.Category, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "repository.category.list")
	defer span.End()

	query := bson.M{}
	if parentID != "" {
		query["parent_id"] = parentID
	}
	if level > 0 {
		query["level"] = level
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []domain.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if categories == nil {
		categories = []domain.Category{}
	}

	if parentID == "" && level <= 0 {
		categories = r.buildTree(categories)
	}

	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "repository.category.update")
	defer span.End()

	update := bson.M{
		"$set": bson.M{
			"name":       category.Name,
			"parent_id":  category.ParentID,
			"level":      category.Level,
			"sort_order": category.SortOrder,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"category_id": category.CategoryID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *CategoryRepository) buildTree(categories []domain.Category) []domain.Category {
	categoryMap := make(map[string]*domain.Category, len(categories))
	for i := range categories {
		categoryMap[categories[i].CategoryID] = &categories[i]
	}

	for i := range categories {
		c := &categories[i]
		if c.ParentID != "" {
			if parent, ok := categoryMap[c.ParentID]; ok {
				parent.Children = append(parent.Children, *c)
			}
		}
	}

	var roots []domain.Category
	for _, c := range categories {
		if c.ParentID == "" {
			roots = append(roots, c)
		}
	}

	return roots
}

// GetBySlug retrieves a category by slug directly - avoids loading all categories for slug lookup
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "repository.category.get_by_slug")
	defer span.End()

	span.SetAttributes(attribute.String("slug", slug))

	var category domain.Category
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}
