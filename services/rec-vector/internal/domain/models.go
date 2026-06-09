package domain

import (
	"time"

	"github.com/google/uuid"
)

type CollectionStatus string

const (
	CollectionActive     CollectionStatus = "active"
	CollectionBuilding   CollectionStatus = "building"
	CollectionReindexing CollectionStatus = "reindexing"
	CollectionError      CollectionStatus = "error"
	CollectionDeprecated CollectionStatus = "deprecated"
)

type EntityType string

const (
	EntityProduct  EntityType = "product"
	EntityUser     EntityType = "user"
	EntitySeller   EntityType = "seller"
	EntityCategory EntityType = "category"
	EntityQuery    EntityType = "query"
	EntityImage    EntityType = "image"
)

type DistanceMetric string

const (
	DistanceCosine    DistanceMetric = "cosine"
	DistanceEuclidean DistanceMetric = "euclidean"
	DistanceDot       DistanceMetric = "dot"
	DistanceManhattan DistanceMetric = "manhattan"
)

type RecommendAlgorithm string

const (
	AlgoCollaborative RecommendAlgorithm = "collaborative_filtering"
	AlgoContentBased  RecommendAlgorithm = "content_based"
	AlgoHybrid        RecommendAlgorithm = "hybrid"
	AlgoTrending      RecommendAlgorithm = "trending"
	AlgoPersonalized  RecommendAlgorithm = "personalized"
	AlgoSimilarItems  RecommendAlgorithm = "similar_items"
	AlgoFrequentlyBought RecommendAlgorithm = "frequently_bought"
)

type VectorCollection struct {
	ID              string           `db:"id" json:"id"`
	Name            string           `db:"name" json:"name"`
	Description     *string          `db:"description" json:"description,omitempty"`
	EntityType      EntityType       `db:"entity_type" json:"entity_type"`
	VectorDimension int              `db:"vector_dimension" json:"vector_dimension"`
	DistanceMetric  DistanceMetric   `db:"distance_metric" json:"distance_metric"`
	IndexType       string           `db:"index_type" json:"index_type"`
	TotalVectors    int64            `db:"total_vectors" json:"total_vectors"`
	Status          CollectionStatus `db:"status" json:"status"`
	CreatedAt       time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at" json:"updated_at"`
}

func NewCollection(name string, entityType EntityType, dimension int) *VectorCollection {
	return &VectorCollection{
		ID:              uuid.New().String(),
		Name:            name,
		EntityType:      entityType,
		VectorDimension: dimension,
		DistanceMetric:  DistanceCosine,
		IndexType:       "hnsw",
		Status:          CollectionActive,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

type VectorMetadata struct {
	ID           string    `db:"id" json:"id"`
	CollectionID string    `db:"collection_id" json:"collection_id"`
	EntityID     string    `db:"entity_id" json:"entity_id"`
	VectorID     string    `db:"vector_id" json:"vector_id"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type EmbeddingModel struct {
	ID                string    `db:"id" json:"id"`
	Name              string    `db:"name" json:"name"`
	Version           string    `db:"version" json:"version"`
	ModelType         string    `db:"model_type" json:"model_type"`
	Framework         string    `db:"framework" json:"framework"`
	VectorDimension   int       `db:"vector_dimension" json:"vector_dimension"`
	Status            string    `db:"status" json:"status"`
	IsDefault         bool      `db:"is_default" json:"is_default"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

type RecommendationResult struct {
	ID              string             `db:"id" json:"id"`
	RequestID       string             `db:"request_id" json:"request_id"`
	UserID          *string            `db:"user_id" json:"user_id,omitempty"`
	CollectionID    string             `db:"collection_id" json:"collection_id"`
	Algorithm       RecommendAlgorithm `db:"algorithm" json:"algorithm"`
	Recommendations string             `db:"recommendations" json:"-"`
	LatencyMs       int                `db:"latency_ms" json:"latency_ms"`
	CreatedAt       time.Time          `db:"created_at" json:"created_at"`
}

type IndexJob struct {
	ID               string     `db:"id" json:"id"`
	CollectionID     string     `db:"collection_id" json:"collection_id"`
	JobType          string     `db:"job_type" json:"job_type"`
	Status           string     `db:"status" json:"status"`
	VectorsTotal     int64      `db:"vectors_total" json:"vectors_total"`
	VectorsProcessed int64      `db:"vectors_processed" json:"vectors_processed"`
	ErrorMessage     *string    `db:"error_message" json:"error_message,omitempty"`
	StartedAt        *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt      *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}
