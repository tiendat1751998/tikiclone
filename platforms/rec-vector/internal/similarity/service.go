package similarity

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/tikiclone/tiki/platforms/rec-vector/internal/vectorstore"
)

type Service struct {
	store vectorstore.VectorStore
}

func NewService(store vectorstore.VectorStore) *Service {
	return &Service{store: store}
}

func (s *Service) Search(ctx context.Context, req *SimilarityRequest) ([]SimilarityResult, error) {
	if len(req.QueryEmbedding) == 0 {
		return nil, ErrEmptyQueryVector
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	results, err := s.store.Search(ctx, req.QueryEmbedding, req.Namespace, 0)
	if err != nil {
		return nil, err
	}

	var filtered []vectorstore.SearchResult
	for _, r := range results {
		if r.Score < req.MinScore {
			continue
		}
		if !matchFilter(r.Metadata, req.Filter) {
			continue
		}
		filtered = append(filtered, r)
	}

	if topK > len(filtered) {
		topK = len(filtered)
	}
	filtered = filtered[:topK]

	out := make([]SimilarityResult, len(filtered))
	for i, r := range filtered {
		out[i] = SimilarityResult{
			ID:       r.ID,
			Score:    r.Score,
			Metadata: r.Metadata,
			Rank:     i + 1,
		}
	}
	return out, nil
}

func (s *Service) HybridSearch(ctx context.Context, req *SimilarityRequest) ([]SimilarityResult, error) {
	if len(req.QueryEmbedding) == 0 {
		return nil, ErrEmptyQueryVector
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	results, err := s.store.Search(ctx, req.QueryEmbedding, req.Namespace, 0)
	if err != nil {
		return nil, err
	}

	alpha := 0.7
	type hybrid struct {
		result   vectorstore.SearchResult
		combined float64
	}
	var candidates []hybrid

	for _, r := range results {
		if r.Score < req.MinScore {
			continue
		}
		if !matchFilter(r.Metadata, req.Filter) {
			continue
		}
		keywordScore := computeKeywordScore(r, req.Keyword)
		combined := alpha*r.Score + (1-alpha)*keywordScore
		candidates = append(candidates, hybrid{result: r, combined: combined})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].combined > candidates[j].combined
	})

	if topK > len(candidates) {
		topK = len(candidates)
	}
	candidates = candidates[:topK]

	out := make([]SimilarityResult, len(candidates))
	for i, c := range candidates {
		out[i] = SimilarityResult{
			ID:       c.result.ID,
			Score:    math.Round(c.combined*10000) / 10000,
			Metadata: c.result.Metadata,
			Rank:     i + 1,
		}
	}
	return out, nil
}

func (s *Service) FilteredSearch(ctx context.Context, req *SimilarityRequest) ([]SimilarityResult, error) {
	return s.Search(ctx, req)
}

func matchFilter(metadata map[string]interface{}, filter map[string]interface{}) bool {
	if len(filter) == 0 {
		return true
	}
	for k, v := range filter {
		mv, ok := metadata[k]
		if !ok {
			return false
		}
		if mv != v {
			return false
		}
	}
	return true
}

func computeKeywordScore(result vectorstore.SearchResult, keyword string) float64 {
	if keyword == "" {
		return 0
	}
	kw := strings.ToLower(keyword)
	score := 0.0
	if id, ok := result.Metadata["title"]; ok {
		if title, ok := id.(string); ok {
			score += float64(strings.Count(strings.ToLower(title), kw)) * 0.5
		}
	}
	if tags, ok := result.Metadata["tags"]; ok {
		if tagList, ok := tags.(string); ok {
			score += float64(strings.Count(strings.ToLower(tagList), kw)) * 0.3
		}
	}
	if description, ok := result.Metadata["description"]; ok {
		if desc, ok := description.(string); ok {
			score += float64(strings.Count(strings.ToLower(desc), kw)) * 0.2
		}
	}
	if score > 1 {
		score = 1
	}
	return score
}
