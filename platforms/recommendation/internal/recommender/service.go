package recommender

import (
	"context"
	"sort"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/collaborative"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/content"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/personalization"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/reranker"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/trending"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/types"
)

type Service interface {
	GetRecommendations(ctx context.Context, recCtx RecommendationContext) ([]ProductRecommendation, error)
}

type service struct {
	repo            Repository
	collabSvc       collaborative.Service
	contentSvc      content.Service
	trendingSvc     trending.Service
	personalSvc     personalization.Service
	rerankerSvc     reranker.Service
}

func NewService(
	repo Repository,
	collabSvc collaborative.Service,
	contentSvc content.Service,
	trendingSvc trending.Service,
	personalSvc personalization.Service,
	rerankerSvc reranker.Service,
) Service {
	return &service{
		repo:        repo,
		collabSvc:   collabSvc,
		contentSvc:  contentSvc,
		trendingSvc: trendingSvc,
		personalSvc: personalSvc,
		rerankerSvc: rerankerSvc,
	}
}

func (s *service) GetRecommendations(ctx context.Context, recCtx RecommendationContext) ([]ProductRecommendation, error) {
	if recCtx.Limit <= 0 {
		recCtx.Limit = 20
	}
	if recCtx.Limit > 100 {
		recCtx.Limit = 100
	}

	var recs []types.ProductRecommendation

	switch recCtx.Type {
	case types.RecTypeTrending:
		return s.getTrending(ctx, recCtx.Limit)
	case types.RecTypePersonalized:
		if recCtx.UserID == "" {
			return nil, ErrEmptyInput
		}
		recs = s.getPersonalized(ctx, recCtx)
	case types.RecTypeRelated:
		if recCtx.ProductID == "" {
			return nil, ErrEmptyInput
		}
		recs = s.getRelated(ctx, recCtx)
	default:
		recs = s.getHybrid(ctx, recCtx)
	}

	recs = s.deduplicate(recs)
	recs = s.applyRerank(ctx, recs)
	recs = s.truncate(recs, recCtx.Limit)

	if len(recs) == 0 {
		return nil, ErrNoRecommendations
	}

	s.repo.StoreRecommendations(ctx, recs)
	return recs, nil
}

func (s *service) getRelated(ctx context.Context, recCtx RecommendationContext) []ProductRecommendation {
	candidates := make(map[string]*scoredRec)

	similar, err := s.collabSvc.ItemBasedSimilar(ctx, recCtx.ProductID, 50)
	if err == nil {
		for _, sim := range similar {
			if _, ok := candidates[sim.ItemID]; !ok {
				candidates[sim.ItemID] = &scoredRec{
					productID: sim.ItemID,
					reason:    string(types.ReasonBoughtAlsoBought),
				}
			}
			candidates[sim.ItemID].collab = sim.Similarity * 0.4
		}
	}

	contentIDs, err := s.contentSvc.SimilarByContent(ctx, recCtx.ProductID, 50)
	if err == nil {
		for _, id := range contentIDs {
			if _, ok := candidates[id]; !ok {
				candidates[id] = &scoredRec{
					productID: id,
					reason:    string(types.ReasonSimilarProducts),
				}
			}
			candidates[id].content = 0.25
		}
	}

	return s.scoreAndSort(candidates, 1, 1, 0, 0)
}

func (s *service) getTrending(ctx context.Context, limit int) ([]ProductRecommendation, error) {
	trendingScores, err := s.trendingSvc.GetTrending(ctx, limit)
	if err != nil {
		return nil, err
	}

	recs := make([]ProductRecommendation, len(trendingScores))
	for i, ts := range trendingScores {
		recs[i] = ProductRecommendation{
			ProductID: ts.ProductID,
			Score:     ts.Score,
			Type:      types.RecTypeTrending,
			Reason:    string(types.ReasonTrendingNow),
		}
	}
	return recs, nil
}

func (s *service) getPersonalized(ctx context.Context, recCtx RecommendationContext) []ProductRecommendation {
	candidates := make(map[string]*scoredRec)

	profile, err := s.personalSvc.GetProfile(ctx, recCtx.UserID)
	if err == nil && profile != nil {
		allFeatures, _ := content.NewInMemoryRepository().GetAllProductFeatures(ctx)
		for _, f := range allFeatures {
			if _, ok := candidates[f.ProductID]; !ok {
				candidates[f.ProductID] = &scoredRec{
					productID: f.ProductID,
					reason:    string(types.ReasonPersonalized),
				}
			}
			candidates[f.ProductID].personalized = profile.CategoryWeights[f.Category]
		}
	}

	recommended, err := s.collabSvc.UserBasedRecommend(ctx, recCtx.UserID, 30)
	if err == nil {
		for _, rec := range recommended {
			if _, ok := candidates[rec.ItemID]; !ok {
				candidates[rec.ItemID] = &scoredRec{
					productID: rec.ItemID,
					reason:    string(types.ReasonBoughtAlsoBought),
				}
			}
			candidates[rec.ItemID].collab = rec.Similarity * 0.4
		}
	}

	trendingScores, err := s.trendingSvc.GetTrending(ctx, 20)
	if err == nil {
		for _, ts := range trendingScores {
			if _, ok := candidates[ts.ProductID]; !ok {
				candidates[ts.ProductID] = &scoredRec{
					productID: ts.ProductID,
					reason:    string(types.ReasonTrendingNow),
				}
			}
			candidates[ts.ProductID].trending = ts.Score * 0.2
		}
	}

	return s.scoreAndSort(candidates, 0.4, 0, 0.2, 0.4)
}

func (s *service) getHybrid(ctx context.Context, recCtx RecommendationContext) []ProductRecommendation {
	candidates := make(map[string]*scoredRec)

	if recCtx.UserID != "" {
		recommended, err := s.collabSvc.UserBasedRecommend(ctx, recCtx.UserID, 30)
		if err == nil {
			for _, rec := range recommended {
				if _, ok := candidates[rec.ItemID]; !ok {
					candidates[rec.ItemID] = &scoredRec{
						productID: rec.ItemID,
						reason:    string(types.ReasonBoughtAlsoBought),
					}
				}
				candidates[rec.ItemID].collab = rec.Similarity * 0.4
			}
		}

		profile, _ := s.personalSvc.GetProfile(ctx, recCtx.UserID)
		if profile != nil {
			allFeatures, _ := content.NewInMemoryRepository().GetAllProductFeatures(ctx)
			for _, f := range allFeatures {
				if _, ok := candidates[f.ProductID]; !ok {
					candidates[f.ProductID] = &scoredRec{
						productID: f.ProductID,
						reason:    string(types.ReasonPersonalized),
					}
				}
				candidates[f.ProductID].personalized = profile.CategoryWeights[f.Category] * 0.15
			}
		}
	}

	trendingScores, err := s.trendingSvc.GetTrending(ctx, 20)
	if err == nil {
		for _, ts := range trendingScores {
			if _, ok := candidates[ts.ProductID]; !ok {
				candidates[ts.ProductID] = &scoredRec{
					productID: ts.ProductID,
					reason:    string(types.ReasonTrendingNow),
				}
			}
			candidates[ts.ProductID].trending = ts.Score * 0.2
		}
	}

	return s.scoreAndSort(candidates, 0.4, 0.25, 0.2, 0.15)
}

type scoredRec struct {
	productID    string
	collab       float64
	content      float64
	trending     float64
	personalized float64
	reason       string
}

func (s *service) scoreAndSort(candidates map[string]*scoredRec, wCollab, wContent, wTrending, wPersonal float64) []ProductRecommendation {
	results := make([]ProductRecommendation, 0, len(candidates))
	for id, cr := range candidates {
		total := cr.collab*wCollab + cr.content*wContent + cr.trending*wTrending + cr.personalized*wPersonal
		results = append(results, ProductRecommendation{
			ProductID: id,
			Score:     total,
			Type:      types.RecTypeRelated,
			Reason:    cr.reason,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func (s *service) deduplicate(recs []ProductRecommendation) []ProductRecommendation {
	seen := make(map[string]bool)
	result := make([]ProductRecommendation, 0, len(recs))
	for _, r := range recs {
		if !seen[r.ProductID] {
			seen[r.ProductID] = true
			result = append(result, r)
		}
	}
	return result
}

func (s *service) applyRerank(ctx context.Context, recs []ProductRecommendation) []ProductRecommendation {
	reranked, err := s.rerankerSvc.ReRank(ctx, recs)
	if err != nil {
		return recs
	}
	return reranked
}

func (s *service) truncate(recs []ProductRecommendation, limit int) []ProductRecommendation {
	if len(recs) <= limit {
		return recs
	}
	return recs[:limit]
}
