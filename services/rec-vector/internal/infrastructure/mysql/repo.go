package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tikiclone/tiki/services/rec-vector/internal/config"
	"github.com/tikiclone/tiki/services/rec-vector/internal/domain"
)

type Repository struct {
	db *sqlx.DB
}

func NewDB(cfg config.MySQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to mysql: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

var collectionCols = "id, name, description, entity_type, vector_dimension, distance_metric, index_type, total_vectors, status, created_at, updated_at"

func (r *Repository) CreateCollection(ctx context.Context, c *domain.VectorCollection) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO vector_collections
		(id, name, description, entity_type, vector_dimension, distance_metric, index_type, 
		 total_vectors, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		c.ID, c.Name, c.Description, c.EntityType, c.VectorDimension, c.DistanceMetric,
		c.IndexType, c.TotalVectors, c.Status, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *Repository) GetCollection(ctx context.Context, id string) (*domain.VectorCollection, error) {
	c := &domain.VectorCollection{}
	err := r.db.GetContext(ctx, c, "SELECT "+collectionCols+" FROM vector_collections WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return c, nil
}

func (r *Repository) ListCollections(ctx context.Context) ([]domain.VectorCollection, error) {
	var cols []domain.VectorCollection
	err := r.db.SelectContext(ctx, &cols, "SELECT "+collectionCols+" FROM vector_collections ORDER BY name")
	return cols, err
}

var metadataCols = "id, collection_id, entity_id, vector_id, is_active, created_at, updated_at"

func (r *Repository) CreateMetadata(ctx context.Context, m *domain.VectorMetadata) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO vector_metadata
		(id, collection_id, entity_id, vector_id, is_active, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?)`,
		m.ID, m.CollectionID, m.EntityID, m.VectorID, m.IsActive, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *Repository) GetMetadata(ctx context.Context, collectionID, entityID string) (*domain.VectorMetadata, error) {
	m := &domain.VectorMetadata{}
	err := r.db.GetContext(ctx, m, "SELECT "+metadataCols+" FROM vector_metadata WHERE collection_id = ? AND entity_id = ?", collectionID, entityID)
	if err != nil {
		return nil, mapError(err)
	}
	return m, nil
}

func (r *Repository) ListMetadata(ctx context.Context, collectionID string) ([]domain.VectorMetadata, error) {
	var items []domain.VectorMetadata
	err := r.db.SelectContext(ctx, &items, "SELECT "+metadataCols+" FROM vector_metadata WHERE collection_id = ? AND is_active = 1", collectionID)
	return items, err
}

var modelCols = "id, name, version, model_type, framework, vector_dimension, status, is_default, created_at, updated_at"

func (r *Repository) ListModels(ctx context.Context) ([]domain.EmbeddingModel, error) {
	var models []domain.EmbeddingModel
	err := r.db.SelectContext(ctx, &models, "SELECT "+modelCols+" FROM embedding_models ORDER BY name, version DESC")
	return models, err
}

func (r *Repository) GetModel(ctx context.Context, id string) (*domain.EmbeddingModel, error) {
	m := &domain.EmbeddingModel{}
	err := r.db.GetContext(ctx, m, "SELECT "+modelCols+" FROM embedding_models WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return m, nil
}

var recCols = "id, request_id, user_id, collection_id, algorithm, recommendations, latency_ms, created_at"

func (r *Repository) SaveRecommendation(ctx context.Context, rec *domain.RecommendationResult) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO recommendation_results
		(id, request_id, user_id, collection_id, algorithm, recommendations, latency_ms, created_at)
		VALUES (?,?,?,?,?,?,?,?)`,
		rec.ID, rec.RequestID, rec.UserID, rec.CollectionID, rec.Algorithm, rec.Recommendations, rec.LatencyMs, rec.CreatedAt)
	return err
}

func (r *Repository) GetRecommendations(ctx context.Context, userID, collectionID string, algorithm domain.RecommendAlgorithm, limit int) ([]domain.RecommendationResult, error) {
	var recs []domain.RecommendationResult
	err := r.db.SelectContext(ctx, &recs,
		"SELECT "+recCols+" FROM recommendation_results "+
			"WHERE (user_id = ? OR ? = '') AND (collection_id = ? OR ? = '') AND algorithm = ? "+
			"ORDER BY created_at DESC LIMIT ?",
		userID, userID, collectionID, collectionID, algorithm, limit)
	return recs, err
}

var jobCols = "id, collection_id, job_type, status, vectors_total, vectors_processed, error_message, started_at, completed_at, created_at"

func (r *Repository) ListJobs(ctx context.Context) ([]domain.IndexJob, error) {
	var jobs []domain.IndexJob
	err := r.db.SelectContext(ctx, &jobs, "SELECT "+jobCols+" FROM vector_index_jobs ORDER BY created_at DESC LIMIT 50")
	return jobs, err
}

func (r *Repository) CreateJob(ctx context.Context, job *domain.IndexJob) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO vector_index_jobs
		(id, collection_id, job_type, status, priority, vectors_total, vectors_processed, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		job.ID, job.CollectionID, job.JobType, job.Status, 5, job.VectorsTotal, job.VectorsProcessed, job.CreatedAt, job.CreatedAt)
	return err
}

func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
