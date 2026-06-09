package application

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/rec-vector/internal/config"
	"github.com/tikiclone/tiki/services/rec-vector/internal/domain"
	redisinfra "github.com/tikiclone/tiki/services/rec-vector/internal/infrastructure/redis"
)

type Repository interface {
	CreateCollection(ctx context.Context, c *domain.VectorCollection) error
	GetCollection(ctx context.Context, id string) (*domain.VectorCollection, error)
	ListCollections(ctx context.Context) ([]domain.VectorCollection, error)

	CreateMetadata(ctx context.Context, m *domain.VectorMetadata) error
	GetMetadata(ctx context.Context, collectionID, entityID string) (*domain.VectorMetadata, error)
	ListMetadata(ctx context.Context, collectionID string) ([]domain.VectorMetadata, error)

	ListModels(ctx context.Context) ([]domain.EmbeddingModel, error)
	GetModel(ctx context.Context, id string) (*domain.EmbeddingModel, error)

	SaveRecommendation(ctx context.Context, rec *domain.RecommendationResult) error
	GetRecommendations(ctx context.Context, userID, collectionID string, algorithm domain.RecommendAlgorithm, limit int) ([]domain.RecommendationResult, error)

	ListJobs(ctx context.Context) ([]domain.IndexJob, error)
	CreateJob(ctx context.Context, job *domain.IndexJob) error
}

type RecService struct {
	repo       Repository
	redisStore *redisinfra.Store
	cfg        *config.RecConfig
}

func NewRecService(repo Repository, redisStore *redisinfra.Store, cfg *config.RecConfig) *RecService {
	return &RecService{repo: repo, redisStore: redisStore, cfg: cfg}
}

func (s *RecService) GetRecommendations(ctx context.Context, userID, collectionID string, algorithm domain.RecommendAlgorithm, limit int) ([]domain.RecommendationResult, error) {
	if limit <= 0 || limit > s.cfg.MaxLimit {
		limit = s.cfg.DefaultLimit
	}
	if algorithm == "" {
		algorithm = domain.AlgoHybrid
	}

	cacheKey := fmt.Sprintf("rec:%s:%s:%s", userID, collectionID, algorithm)
	var cached []domain.RecommendationResult
	if err := s.redisStore.GetCachedRecommendations(ctx, cacheKey, &cached); err == nil && len(cached) > 0 {
		return cached, nil
	}

	recs, err := s.repo.GetRecommendations(ctx, userID, collectionID, algorithm, limit)
	if err != nil {
		return nil, err
	}

	if len(recs) > 0 {
		_ = s.redisStore.CacheRecommendations(ctx, cacheKey, recs, s.cfg.CacheTTL)
		return recs, nil
	}

	rec := s.generateFallbackRecommendations(ctx, userID, collectionID, algorithm, limit)
	_ = s.repo.SaveRecommendation(ctx, rec)

	result := []domain.RecommendationResult{*rec}
	_ = s.redisStore.CacheRecommendations(ctx, cacheKey, result, s.cfg.CacheTTL)
	return result, nil
}

func (s *RecService) generateFallbackRecommendations(ctx context.Context, userID, collectionID string, algorithm domain.RecommendAlgorithm, limit int) *domain.RecommendationResult {
	now := time.Now().UTC()
	items := make([]map[string]interface{}, 0, limit)

	var latency int

	switch algorithm {
	case domain.AlgoTrending:
		for i := 0; i < limit; i++ {
			items = append(items, map[string]interface{}{
				"product_id":    uuid.New().String(),
				"score":         fmt.Sprintf("%.4f", 1.0-float64(i)*0.05),
				"reason":        "trending",
				"trending_rank": i + 1,
			})
		}
		latency = rand.Intn(50) + 5
	case domain.AlgoSimilarItems:
		for i := 0; i < limit; i++ {
			items = append(items, map[string]interface{}{
				"product_id": uuid.New().String(),
				"similarity":  fmt.Sprintf("%.4f", 0.95-float64(i)*0.03),
				"reason":      "similar_items",
			})
		}
		latency = rand.Intn(100) + 20
	default:
		for i := 0; i < limit; i++ {
			items = append(items, map[string]interface{}{
				"product_id": uuid.New().String(),
				"score":      fmt.Sprintf("%.4f", 0.9-float64(i)*0.04),
				"reason":     "recommended_for_you",
			})
		}
		latency = rand.Intn(80) + 15
	}

	recommendations, _ := json.Marshal(items)

	var uid *string
	if userID != "" {
		uid = &userID
	}

	return &domain.RecommendationResult{
		ID:              uuid.New().String(),
		RequestID:       uuid.New().String(),
		UserID:          uid,
		CollectionID:    collectionID,
		Algorithm:       algorithm,
		Recommendations: string(recommendations),
		LatencyMs:       latency,
		CreatedAt:       now,
	}
}

func (s *RecService) ListCollections(ctx context.Context) ([]domain.VectorCollection, error) {
	return s.repo.ListCollections(ctx)
}

func (s *RecService) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*domain.VectorCollection, error) {
	c := domain.NewCollection(req.Name, req.EntityType, req.VectorDimension)
	if req.Description != "" {
		c.Description = &req.Description
	}
	if err := s.repo.CreateCollection(ctx, c); err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}
	return c, nil
}

func (s *RecService) GetCollection(ctx context.Context, id string) (*domain.VectorCollection, error) {
	return s.repo.GetCollection(ctx, id)
}

func (s *RecService) AddMetadata(ctx context.Context, req *AddMetadataRequest) (*domain.VectorMetadata, error) {
	m := &domain.VectorMetadata{
		ID:           uuid.New().String(),
		CollectionID: req.CollectionID,
		EntityID:     req.EntityID,
		VectorID:     uuid.New().String(),
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := s.repo.CreateMetadata(ctx, m); err != nil {
		return nil, fmt.Errorf("create metadata: %w", err)
	}
	return m, nil
}

func (s *RecService) ListMetadata(ctx context.Context, collectionID string) ([]domain.VectorMetadata, error) {
	return s.repo.ListMetadata(ctx, collectionID)
}

func (s *RecService) ListModels(ctx context.Context) ([]domain.EmbeddingModel, error) {
	return s.repo.ListModels(ctx)
}

func (s *RecService) GetModel(ctx context.Context, id string) (*domain.EmbeddingModel, error) {
	return s.repo.GetModel(ctx, id)
}

func (s *RecService) ListJobs(ctx context.Context) ([]domain.IndexJob, error) {
	return s.repo.ListJobs(ctx)
}

func (s *RecService) CreateJob(ctx context.Context, collectionID, jobType string) (*domain.IndexJob, error) {
	now := time.Now().UTC()
	job := &domain.IndexJob{
		ID:           uuid.New().String(),
		CollectionID: collectionID,
		JobType:      jobType,
		Status:       "queued",
		CreatedAt:    now,
	}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}
	return job, nil
}

type CreateCollectionRequest struct {
	Name            string           `json:"name" binding:"required"`
	Description     string           `json:"description"`
	EntityType      domain.EntityType `json:"entity_type" binding:"required"`
	VectorDimension int              `json:"vector_dimension" binding:"required"`
}

type AddMetadataRequest struct {
	CollectionID string `json:"-"`
	EntityID     string `json:"entity_id" binding:"required"`
}
