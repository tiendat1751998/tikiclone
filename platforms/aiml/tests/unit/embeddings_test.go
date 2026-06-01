package unit

import (
	"context"
	"math"
	"testing"

	"github.com/tikiclone/tiki/platforms/aiml/internal/embeddings"
)

func TestEmbeddingGeneration(t *testing.T) {
	gen := embeddings.NewEmbeddingGenerator()
	emb := gen.Generate("user-1", "user", "bert", "v1")

	if emb.EntityID != "user-1" {
		t.Errorf("expected user-1, got %s", emb.EntityID)
	}
	if len(emb.Vector) != 128 {
		t.Errorf("expected 128-dim vector, got %d", len(emb.Vector))
	}

	norm := 0.0
	for _, v := range emb.Vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 0.001 {
		t.Errorf("expected unit vector, got norm=%f", norm)
	}
}

func TestEmbeddingDeterminism(t *testing.T) {
	gen := embeddings.NewEmbeddingGenerator()
	e1 := gen.Generate("same-id", "user", "m", "v1")
	e2 := gen.Generate("same-id", "user", "m", "v1")

	for i := range e1.Vector {
		if e1.Vector[i] != e2.Vector[i] {
			t.Errorf("vectors differ at index %d", i)
			break
		}
	}
}

func TestEmbeddingStoreAndRetrieve(t *testing.T) {
	store := embeddings.NewInMemoryVectorStore()
	gen := embeddings.NewEmbeddingGenerator()

	emb := gen.Generate("prod-1", "product", "m", "v1")
	err := store.Store(context.Background(), emb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := store.Get(context.Background(), "prod-1", "product")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.EntityID != "prod-1" {
		t.Errorf("expected prod-1, got %s", got.EntityID)
	}
}

func TestEmbeddingNotFound(t *testing.T) {
	store := embeddings.NewInMemoryVectorStore()
	_, err := store.Get(context.Background(), "nonexistent", "user")
	if err != embeddings.ErrEmbeddingNotFound {
		t.Errorf("expected ErrEmbeddingNotFound, got %v", err)
	}
}

func TestSimilaritySearch(t *testing.T) {
	store := embeddings.NewInMemoryVectorStore()
	gen := embeddings.NewEmbeddingGenerator()

	for i := 0; i < 10; i++ {
		emb := gen.Generate(string(rune('A'+i)), "product", "m", "v1")
		store.Store(context.Background(), emb)
	}

	queryEmb := gen.Generate("query", "product", "m", "v1")
	results, err := store.FindSimilar(context.Background(), queryEmb.Vector, "product", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(results))
	}
	if len(results) > 1 && results[0].Score < results[1].Score {
		t.Error("results should be sorted by score descending")
	}
}

func TestCosineSimilarity(t *testing.T) {
	v1 := []float64{1, 0, 0}
	v2 := []float64{0, 1, 0}
	v3 := []float64{0.707, 0.707, 0}

	store := embeddings.NewInMemoryVectorStore()
	store.Store(context.Background(), &embeddings.Embedding{EntityID: "e1", EntityType: "t", Vector: v1})
	store.Store(context.Background(), &embeddings.Embedding{EntityID: "e2", EntityType: "t", Vector: v2})
	store.Store(context.Background(), &embeddings.Embedding{EntityID: "e3", EntityType: "t", Vector: v3})

	results, _ := store.FindSimilar(context.Background(), v1, "t", 3)
	if len(results) > 0 && results[0].EntityID != "e1" {
		t.Errorf("expected e1 as most similar, got %s", results[0].EntityID)
	}
}

func TestBatchGenerate(t *testing.T) {
	gen := embeddings.NewEmbeddingGenerator()
	store := embeddings.NewInMemoryVectorStore()
	svc := embeddings.NewService(gen, store)

	ids := []string{"u1", "u2", "u3"}
	embeddingsList, err := svc.BatchGenerate(context.Background(), ids, "user", "m", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(embeddingsList) != 3 {
		t.Errorf("expected 3 embeddings, got %d", len(embeddingsList))
	}
}
