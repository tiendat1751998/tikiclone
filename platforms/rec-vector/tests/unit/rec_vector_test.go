package unit

import (
	"context"
	"math"
	"testing"

	"github.com/tikiclone/tiki/platforms/rec-vector/internal/collabvector"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/itemembedding"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/realtime"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/similarity"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/userembedding"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/vectorstore"
)

func TestVectorStoreInsertAndGet(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	rec := &vectorstore.VectorRecord{
		ID:        "v1",
		Vector:    []float64{1, 0, 0, 0},
		Metadata:  map[string]interface{}{"name": "test"},
		Namespace: "ns1",
	}
	if err := store.Insert(ctx, rec); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	got, err := store.Get(ctx, "v1", "ns1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != "v1" {
		t.Errorf("expected v1, got %s", got.ID)
	}
}

func TestVectorStoreBatchInsert(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	records := []*vectorstore.VectorRecord{
		{ID: "a", Vector: []float64{1, 0, 0}, Namespace: "ns"},
		{ID: "b", Vector: []float64{0, 1, 0}, Namespace: "ns"},
		{ID: "c", Vector: []float64{0, 0, 1}, Namespace: "ns"},
	}
	if err := store.BatchInsert(ctx, records); err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	for _, id := range []string{"a", "b", "c"} {
		if _, err := store.Get(ctx, id, "ns"); err != nil {
			t.Errorf("expected to find %s: %v", id, err)
		}
	}
}

func TestVectorStoreDelete(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "del", Vector: []float64{1, 0, 0}, Namespace: "ns"})
	if err := store.Delete(ctx, "del", "ns"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := store.Get(ctx, "del", "ns"); err != vectorstore.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestVectorStoreSearchKNN(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "a", Vector: []float64{1, 0, 0}, Namespace: "ns"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "b", Vector: []float64{0.9, 0.1, 0}, Namespace: "ns"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "c", Vector: []float64{0, 1, 0}, Namespace: "ns"})

	results, err := store.Search(ctx, []float64{1, 0, 0}, "ns", 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if results[0].ID != "a" {
		t.Errorf("expected 'a' as top result, got %s", results[0].ID)
	}
	if results[0].Score < results[1].Score {
		t.Error("results must be sorted descending by score")
	}
}

func TestVectorStoreNamespaceIsolation(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "x", Vector: []float64{1, 0, 0}, Namespace: "a"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "y", Vector: []float64{1, 0, 0}, Namespace: "b"})

	results, err := store.Search(ctx, []float64{1, 0, 0}, "a", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "x" {
		t.Errorf("expected only 'x' in namespace a, got %+v", results)
	}
}

func TestVectorStoreGetNotFound(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	_, err := store.Get(context.Background(), "nonexistent", "ns")
	if err != vectorstore.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestVectorStoreEmptyVector(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	err := store.Insert(context.Background(), &vectorstore.VectorRecord{Vector: []float64{}, Namespace: "ns"})
	if err != vectorstore.ErrEmptyVector {
		t.Errorf("expected ErrEmptyVector, got %v", err)
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	c := []float64{0.707, 0.707, 0}

	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "a", Vector: a, Namespace: "t"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "b", Vector: b, Namespace: "t"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "c", Vector: c, Namespace: "t"})

	results, _ := store.Search(ctx, a, "t", 3)
	if len(results) > 0 && results[0].ID != "a" {
		t.Errorf("expected 'a' as most similar to itself, got %s", results[0].ID)
	}
}

func TestUserEmbeddingGenerate(t *testing.T) {
	repo := userembedding.NewInMemoryRepository()
	svc := userembedding.NewService(repo)

	emb, err := svc.GenerateUserEmbedding(context.Background(), "user-1", "v1")
	if err != nil {
		t.Fatalf("GenerateUserEmbedding failed: %v", err)
	}
	if emb.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", emb.UserID)
	}
	if len(emb.Embedding) != 128 {
		t.Errorf("expected 128-dim, got %d", len(emb.Embedding))
	}
	if emb.ModelVersion != "v1" {
		t.Errorf("expected v1, got %s", emb.ModelVersion)
	}
}

func TestUserEmbeddingDeterminism(t *testing.T) {
	repo := userembedding.NewInMemoryRepository()
	svc := userembedding.NewService(repo)

	e1, _ := svc.GenerateUserEmbedding(context.Background(), "uid", "v1")
	e2, _ := svc.GenerateUserEmbedding(context.Background(), "uid", "v1")

	for i := range e1.Embedding {
		if e1.Embedding[i] != e2.Embedding[i] {
			t.Fatalf("vectors differ at index %d", i)
		}
	}
}

func TestUserEmbeddingUnitNorm(t *testing.T) {
	repo := userembedding.NewInMemoryRepository()
	svc := userembedding.NewService(repo)

	emb, _ := svc.GenerateUserEmbedding(context.Background(), "uid", "v1")
	norm := 0.0
	for _, v := range emb.Embedding {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 0.001 {
		t.Errorf("expected unit norm, got %f", norm)
	}
}

func TestUserEmbeddingUpdate(t *testing.T) {
	repo := userembedding.NewInMemoryRepository()
	svc := userembedding.NewService(repo)

	svc.GenerateUserEmbedding(context.Background(), "u1", "v1")
	newVec := make([]float64, 128)
	newVec[0] = 1.0
	emb, err := svc.UpdateEmbedding(context.Background(), "u1", newVec, "v2")
	if err != nil {
		t.Fatalf("UpdateEmbedding failed: %v", err)
	}
	if emb.Embedding[0] != 1.0 {
		t.Errorf("expected updated vector, got %f", emb.Embedding[0])
	}
	if emb.ModelVersion != "v2" {
		t.Errorf("expected v2, got %s", emb.ModelVersion)
	}
}

func TestUserEmbeddingGet(t *testing.T) {
	repo := userembedding.NewInMemoryRepository()
	svc := userembedding.NewService(repo)

	svc.GenerateUserEmbedding(context.Background(), "u1", "v1")
	emb, err := svc.GetEmbedding(context.Background(), "u1")
	if err != nil {
		t.Fatalf("GetEmbedding failed: %v", err)
	}
	if emb.UserID != "u1" {
		t.Errorf("expected u1, got %s", emb.UserID)
	}
}

func TestUserEmbeddingNotFound(t *testing.T) {
	repo := userembedding.NewInMemoryRepository()
	svc := userembedding.NewService(repo)

	_, err := svc.GetEmbedding(context.Background(), "nonexistent")
	if err != userembedding.ErrUserEmbeddingNotFound {
		t.Errorf("expected ErrUserEmbeddingNotFound, got %v", err)
	}
}

func TestUserEmbeddingBatchGet(t *testing.T) {
	repo := userembedding.NewInMemoryRepository()
	svc := userembedding.NewService(repo)

	svc.GenerateUserEmbedding(context.Background(), "u1", "v1")
	svc.GenerateUserEmbedding(context.Background(), "u2", "v1")

	results, err := svc.BatchGetEmbeddings(context.Background(), []string{"u1", "u2", "u3"})
	if err != nil {
		t.Fatalf("BatchGetEmbeddings failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 found embeddings, got %d", len(results))
	}
}

func TestItemEmbeddingGenerate(t *testing.T) {
	repo := itemembedding.NewInMemoryRepository()
	svc := itemembedding.NewService(repo)

	emb, err := svc.GenerateItemEmbedding(context.Background(), "item-1", "electronics", []string{"phone", "gadget"}, "v1")
	if err != nil {
		t.Fatalf("GenerateItemEmbedding failed: %v", err)
	}
	if emb.ItemID != "item-1" {
		t.Errorf("expected item-1, got %s", emb.ItemID)
	}
	if emb.Category != "electronics" {
		t.Errorf("expected electronics, got %s", emb.Category)
	}
	if len(emb.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(emb.Tags))
	}
}

func TestItemEmbeddingDeterminism(t *testing.T) {
	repo := itemembedding.NewInMemoryRepository()
	svc := itemembedding.NewService(repo)

	e1, _ := svc.GenerateItemEmbedding(context.Background(), "i1", "cat", nil, "v1")
	e2, _ := svc.GenerateItemEmbedding(context.Background(), "i1", "cat", nil, "v1")

	for i := range e1.Embedding {
		if e1.Embedding[i] != e2.Embedding[i] {
			t.Fatalf("vectors differ at index %d", i)
		}
	}
}

func TestItemEmbeddingUpdate(t *testing.T) {
	repo := itemembedding.NewInMemoryRepository()
	svc := itemembedding.NewService(repo)

	svc.GenerateItemEmbedding(context.Background(), "i1", "cat", nil, "v1")
	newVec := make([]float64, 128)
	newVec[0] = 1.0
	emb, err := svc.UpdateEmbedding(context.Background(), "i1", "cat2", []string{"tag"}, newVec, "v2")
	if err != nil {
		t.Fatalf("UpdateEmbedding failed: %v", err)
	}
	if emb.ModelVersion != "v2" {
		t.Errorf("expected v2, got %s", emb.ModelVersion)
	}
}

func TestItemEmbeddingBatchGet(t *testing.T) {
	repo := itemembedding.NewInMemoryRepository()
	svc := itemembedding.NewService(repo)

	svc.GenerateItemEmbedding(context.Background(), "i1", "cat", nil, "v1")
	svc.GenerateItemEmbedding(context.Background(), "i2", "cat", nil, "v1")

	results, err := svc.BatchGetEmbeddings(context.Background(), []string{"i1", "i2", "i3"})
	if err != nil {
		t.Fatalf("BatchGetEmbeddings failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2, got %d", len(results))
	}
}

func TestSimilaritySearch(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "a", Vector: []float64{1, 0, 0}, Namespace: "ns", Metadata: map[string]interface{}{"type": "test"}})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "b", Vector: []float64{0, 1, 0}, Namespace: "ns", Metadata: map[string]interface{}{"type": "other"}})

	svc := similarity.NewService(store)
	results, err := svc.Search(ctx, &similarity.SimilarityRequest{
		QueryEmbedding: []float64{1, 0, 0},
		Namespace:      "ns",
		TopK:           5,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result (only positive scores), got %d", len(results))
	}
}

func TestSimilarityFilteredSearch(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "a", Vector: []float64{1, 0, 0}, Namespace: "ns", Metadata: map[string]interface{}{"brand": "nike"}})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "b", Vector: []float64{1, 0, 0}, Namespace: "ns", Metadata: map[string]interface{}{"brand": "adidas"}})

	svc := similarity.NewService(store)
	results, err := svc.Search(ctx, &similarity.SimilarityRequest{
		QueryEmbedding: []float64{1, 0, 0},
		Namespace:      "ns",
		TopK:           5,
		Filter:         map[string]interface{}{"brand": "nike"},
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "a" {
		t.Errorf("expected only 'a' with brand=nike, got %+v", results)
	}
}

func TestSimilarityMinScore(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "a", Vector: []float64{1, 0, 0}, Namespace: "ns"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "b", Vector: []float64{0.5, 0.5, 0}, Namespace: "ns"})

	svc := similarity.NewService(store)
	results, err := svc.Search(ctx, &similarity.SimilarityRequest{
		QueryEmbedding: []float64{1, 0, 0},
		Namespace:      "ns",
		TopK:           5,
		MinScore:       0.9,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) > 1 {
		t.Errorf("expected at most 1 result with min_score=0.9, got %d", len(results))
	}
}

func TestSimilarityHybridSearch(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "a", Vector: []float64{1, 0, 0}, Namespace: "ns", Metadata: map[string]interface{}{"title": "red shoes"}})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "b", Vector: []float64{0, 1, 0}, Namespace: "ns", Metadata: map[string]interface{}{"title": "blue shirt"}})

	svc := similarity.NewService(store)
	results, err := svc.HybridSearch(ctx, &similarity.SimilarityRequest{
		QueryEmbedding: []float64{1, 0, 0},
		Namespace:      "ns",
		TopK:           5,
		Keyword:        "blue",
	})
	if err != nil {
		t.Fatalf("HybridSearch failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}

func TestSimilarityEmptyQuery(t *testing.T) {
	svc := similarity.NewService(vectorstore.NewInMemoryStore())
	_, err := svc.Search(context.Background(), &similarity.SimilarityRequest{
		QueryEmbedding: []float64{},
		Namespace:      "ns",
	})
	if err != similarity.ErrEmptyQueryVector {
		t.Errorf("expected ErrEmptyQueryVector, got %v", err)
	}
}

func TestCollabRecordInteraction(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	if err := svc.RecordInteraction(context.Background(), "u1", "i1", "purchase"); err != nil {
		t.Fatalf("RecordInteraction failed: %v", err)
	}
	interactions, err := repo.GetInteractions(context.Background())
	if err != nil {
		t.Fatalf("GetInteractions failed: %v", err)
	}
	if len(interactions) != 1 {
		t.Errorf("expected 1 interaction, got %d", len(interactions))
	}
	if interactions[0].Weight != 1.0 {
		t.Errorf("expected purchase weight=1.0, got %f", interactions[0].Weight)
	}
}

func TestCollabInteractionWeights(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	svc.RecordInteraction(context.Background(), "u1", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u1", "i2", "click")
	svc.RecordInteraction(context.Background(), "u1", "i3", "view")

	matrix, err := repo.GetMatrix(context.Background())
	if err != nil {
		t.Fatalf("GetMatrix failed: %v", err)
	}
	if matrix.UserItemRatings["u1"]["i1"] != 1.0 {
		t.Errorf("expected purchase=1.0, got %f", matrix.UserItemRatings["u1"]["i1"])
	}
	if matrix.UserItemRatings["u1"]["i2"] != 0.5 {
		t.Errorf("expected click=0.5, got %f", matrix.UserItemRatings["u1"]["i2"])
	}
	if matrix.UserItemRatings["u1"]["i3"] != 0.1 {
		t.Errorf("expected view=0.1, got %f", matrix.UserItemRatings["u1"]["i3"])
	}
}

func TestCollabInteractionMatrix(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	svc.RecordInteraction(context.Background(), "u1", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u1", "i2", "click")
	svc.RecordInteraction(context.Background(), "u2", "i1", "purchase")

	matrix, err := repo.GetMatrix(context.Background())
	if err != nil {
		t.Fatalf("GetMatrix failed: %v", err)
	}
	if len(matrix.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(matrix.Users))
	}
	if len(matrix.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(matrix.Items))
	}
}

func TestCollabTrainFactorization(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	svc.RecordInteraction(context.Background(), "u1", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u1", "i2", "click")
	svc.RecordInteraction(context.Background(), "u2", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u2", "i3", "purchase")

	if err := svc.TrainFactorization(context.Background(), 5, 10, 0.01); err != nil {
		t.Fatalf("TrainFactorization failed: %v", err)
	}

	factors, err := repo.GetFactors(context.Background())
	if err != nil {
		t.Fatalf("GetFactors failed: %v", err)
	}
	if len(factors.UserFactors) != 2 {
		t.Errorf("expected 2 user factor vectors, got %d", len(factors.UserFactors))
	}
	if len(factors.ItemFactors) != 3 {
		t.Errorf("expected 3 item factor vectors, got %d", len(factors.ItemFactors))
	}
}

func TestCollabRecommendByFactorization(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	svc.RecordInteraction(context.Background(), "u1", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u1", "i2", "click")
	svc.RecordInteraction(context.Background(), "u2", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u2", "i3", "purchase")
	svc.TrainFactorization(context.Background(), 5, 10, 0.01)

	recs, err := svc.RecommendByFactorization(context.Background(), "u1", 5)
	if err != nil {
		t.Fatalf("RecommendByFactorization failed: %v", err)
	}
	if len(recs) == 0 {
		t.Error("expected at least one recommendation")
	}
	for _, r := range recs {
		if r.ItemID == "i1" || r.ItemID == "i2" {
			t.Errorf("should not recommend already interacted item: %s", r.ItemID)
		}
	}
}

func TestCollabNotEnoughInteractions(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	err := svc.TrainFactorization(context.Background(), 5, 10, 0.01)
	if err != collabvector.ErrNotEnoughInteractions {
		t.Errorf("expected ErrNotEnoughInteractions, got %v", err)
	}
}

func TestCollabUserNotFound(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	svc.RecordInteraction(context.Background(), "u1", "i1", "purchase")
	svc.TrainFactorization(context.Background(), 5, 10, 0.01)

	_, err := svc.RecommendByFactorization(context.Background(), "nonexistent", 5)
	if err != collabvector.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRealtimeTrackEvent(t *testing.T) {
	itemRepo := itemembedding.NewInMemoryRepository()
	itemSvc := itemembedding.NewService(itemRepo)
	store := vectorstore.NewInMemoryStore()
	rtRepo := realtime.NewInMemoryRepository()
	rtSvc := realtime.NewService(rtRepo, itemSvc, store)

	session, err := rtSvc.TrackEvent(context.Background(), "u1", "sess1", "view", "item1", "")
	if err != nil {
		t.Fatalf("TrackEvent failed: %v", err)
	}
	if session.UserID != "u1" {
		t.Errorf("expected u1, got %s", session.UserID)
	}
	if len(session.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(session.Events))
	}
}

func TestRealtimeSessionEmbedding(t *testing.T) {
	itemRepo := itemembedding.NewInMemoryRepository()
	itemSvc := itemembedding.NewService(itemRepo)
	store := vectorstore.NewInMemoryStore()
	rtRepo := realtime.NewInMemoryRepository()
	rtSvc := realtime.NewService(rtRepo, itemSvc, store)

	itemSvc.GenerateItemEmbedding(context.Background(), "item1", "cat", nil, "v1")
	rtSvc.TrackEvent(context.Background(), "u1", "sess1", "view", "item1", "")

	emb, err := rtSvc.GetSessionEmbedding(context.Background(), "sess1")
	if err != nil {
		t.Fatalf("GetSessionEmbedding failed: %v", err)
	}
	if len(emb) == 0 {
		t.Error("expected non-empty embedding")
	}
}

func TestRealtimeSessionEmbeddingEmpty(t *testing.T) {
	itemRepo := itemembedding.NewInMemoryRepository()
	itemSvc := itemembedding.NewService(itemRepo)
	store := vectorstore.NewInMemoryStore()
	rtRepo := realtime.NewInMemoryRepository()
	rtSvc := realtime.NewService(rtRepo, itemSvc, store)

	session := rtSvc.CreateSession(context.Background(), "u1")
	rtRepo.StoreSession(context.Background(), session)

	emb, err := rtSvc.GetSessionEmbedding(context.Background(), session.SessionID)
	if err != nil {
		t.Fatalf("GetSessionEmbedding failed: %v", err)
	}
	if len(emb) != 128 {
		t.Errorf("expected 128-dim zero embedding, got %d", len(emb))
	}
}

func TestRealtimeContextBanditExploration(t *testing.T) {
	itemRepo := itemembedding.NewInMemoryRepository()
	itemSvc := itemembedding.NewService(itemRepo)
	store := vectorstore.NewInMemoryStore()
	rtRepo := realtime.NewInMemoryRepository()
	rtSvc := realtime.NewService(rtRepo, itemSvc, store)

	itemSvc.GenerateItemEmbedding(context.Background(), "i1", "cat", nil, "v1")
	rtSvc.TrackEvent(context.Background(), "u1", "s1", "view", "i1", "")
	store.Insert(context.Background(), &vectorstore.VectorRecord{ID: "i1", Vector: []float64{1, 0, 0}, Namespace: "ns"})

	for i := 0; i < 100; i++ {
		results, err := rtSvc.RecommendWithContext(context.Background(), "s1", "ns", 5)
		if err != nil {
			t.Fatalf("RecommendWithContext failed: %v", err)
		}
		_ = results
	}

	stats, _ := rtRepo.GetArmStats(context.Background())
	if len(stats) == 0 {
		t.Error("expected arm stats to be recorded")
	}
}

func TestRealtimeContextBanditExploitBest(t *testing.T) {
	itemRepo := itemembedding.NewInMemoryRepository()
	itemSvc := itemembedding.NewService(itemRepo)
	store := vectorstore.NewInMemoryStore()
	rtRepo := realtime.NewInMemoryRepository()
	rtSvc := realtime.NewService(rtRepo, itemSvc, store)

	itemSvc.GenerateItemEmbedding(context.Background(), "i1", "cat", nil, "v1")
	rtSvc.TrackEvent(context.Background(), "u1", "s1", "view", "i1", "")
	store.Insert(context.Background(), &vectorstore.VectorRecord{ID: "i1", Vector: []float64{1, 0, 0}, Namespace: "ns"})

	for i := 0; i < 50; i++ {
		rtSvc.RecommendWithContext(context.Background(), "s1", "ns", 5)
	}

	stats, _ := rtRepo.GetArmStats(context.Background())
	for arm, stat := range stats {
		if stat.Plays > 0 {
			t.Logf("arm %s: plays=%d mean_reward=%.2f", arm, stat.Plays, stat.MeanReward)
		}
	}
}

func TestRealtimeRecommendWithContext(t *testing.T) {
	itemRepo := itemembedding.NewInMemoryRepository()
	itemSvc := itemembedding.NewService(itemRepo)
	store := vectorstore.NewInMemoryStore()
	rtRepo := realtime.NewInMemoryRepository()
	rtSvc := realtime.NewService(rtRepo, itemSvc, store)

	itemSvc.GenerateItemEmbedding(context.Background(), "i1", "cat", nil, "v1")
	rtSvc.TrackEvent(context.Background(), "u1", "s1", "view", "i1", "")
	store.Insert(context.Background(), &vectorstore.VectorRecord{ID: "i1", Vector: []float64{1, 0, 0}, Namespace: "ns"})

	results, err := rtSvc.RecommendWithContext(context.Background(), "s1", "ns", 5)
	if err != nil {
		t.Fatalf("RecommendWithContext failed: %v", err)
	}
	if len(results) == 0 {
		t.Log("no results (expected in empty namespace after bandit)")
	}
}

func TestSessionEventOrderPreserved(t *testing.T) {
	itemRepo := itemembedding.NewInMemoryRepository()
	itemSvc := itemembedding.NewService(itemRepo)
	store := vectorstore.NewInMemoryStore()
	rtRepo := realtime.NewInMemoryRepository()
	rtSvc := realtime.NewService(rtRepo, itemSvc, store)

	rtSvc.TrackEvent(context.Background(), "u1", "s1", "view", "i1", "")
	rtSvc.TrackEvent(context.Background(), "u1", "s1", "click", "i2", "")
	rtSvc.TrackEvent(context.Background(), "u1", "s1", "purchase", "i3", "")

	session, _ := rtRepo.GetSession(context.Background(), "s1")
	if len(session.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(session.Events))
	}
	if session.Events[0].EventType != "view" || session.Events[2].EventType != "purchase" {
		t.Error("events order not preserved")
	}
}

func TestVectorSearchScoreRanking(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "d1", Vector: []float64{0.1, 0.9, 0}, Namespace: "ns"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "d2", Vector: []float64{0.95, 0.05, 0}, Namespace: "ns"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "d3", Vector: []float64{0.5, 0.5, 0}, Namespace: "ns"})

	results, _ := store.Search(ctx, []float64{1, 0, 0}, "ns", 3)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("ranking violated at %d: %f > %f", i, results[i].Score, results[i-1].Score)
		}
	}
	if results[0].Rank != 1 || results[1].Rank != 2 || results[2].Rank != 3 {
		t.Errorf("ranks not sequential: got %d,%d,%d", results[0].Rank, results[1].Rank, results[2].Rank)
	}
}

func TestVectorSearchEmptyNamespace(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	store.Insert(ctx, &vectorstore.VectorRecord{ID: "a", Vector: []float64{1, 0, 0}, Namespace: "ns1"})
	store.Insert(ctx, &vectorstore.VectorRecord{ID: "b", Vector: []float64{1, 0, 0}, Namespace: "ns2"})

	results, _ := store.Search(ctx, []float64{1, 0, 0}, "nonexistent", 10)
	if len(results) != 0 {
		t.Errorf("expected 0 results for unknown namespace, got %d", len(results))
	}
}

func TestGenerateLatentFactors(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	svc.RecordInteraction(context.Background(), "u1", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u1", "i2", "click")
	svc.RecordInteraction(context.Background(), "u2", "i1", "purchase")

	factors, err := svc.GenerateLatentFactors(context.Background(), 3)
	if err != nil {
		t.Fatalf("GenerateLatentFactors failed: %v", err)
	}
	if factors == nil {
		t.Fatal("expected non-nil factors")
	}
	if len(factors.UserFactors) != 2 {
		t.Errorf("expected user factors for 2 users, got %d", len(factors.UserFactors))
	}
	if len(factors.ItemFactors) != 2 {
		t.Errorf("expected item factors for 2 items, got %d", len(factors.ItemFactors))
	}
}

func TestCollabRecommendByFactorizationScoreOrder(t *testing.T) {
	repo := collabvector.NewInMemoryRepository()
	svc := collabvector.NewService(repo)

	svc.RecordInteraction(context.Background(), "u1", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u1", "i2", "click")
	svc.RecordInteraction(context.Background(), "u2", "i1", "purchase")
	svc.RecordInteraction(context.Background(), "u2", "i3", "purchase")
	svc.TrainFactorization(context.Background(), 5, 20, 0.01)

	recs, _ := svc.RecommendByFactorization(context.Background(), "u1", 10)
	for i := 1; i < len(recs); i++ {
		if recs[i].Score > recs[i-1].Score {
			t.Errorf("scores not sorted descending at %d: %f > %f", i, recs[i].Score, recs[i-1].Score)
		}
	}
}

func TestVectorStoreAutoGenerateID(t *testing.T) {
	store := vectorstore.NewInMemoryStore()
	ctx := context.Background()

	rec := &vectorstore.VectorRecord{
		Vector:    []float64{1, 0, 0},
		Namespace: "ns",
	}
	if err := store.Insert(ctx, rec); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	if rec.ID == "" {
		t.Error("expected auto-generated ID")
	}
}
