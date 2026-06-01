package recommender

import "github.com/tikiclone/tiki/platforms/recommendation/internal/types"

type RecommendationType = types.RecommendationType
type ProductRecommendation = types.ProductRecommendation
type RecommendationContext = types.RecommendationContext
type Score = types.Score
type Reason = types.Reason

const (
	RecTypeRelated      RecommendationType = types.RecTypeRelated
	RecTypeTrending     RecommendationType = types.RecTypeTrending
	RecTypePersonalized RecommendationType = types.RecTypePersonalized
)

const (
	ReasonBoughtAlsoBought Reason = types.ReasonBoughtAlsoBought
	ReasonSimilarProducts  Reason = types.ReasonSimilarProducts
	ReasonTrendingNow      Reason = types.ReasonTrendingNow
	ReasonPersonalized     Reason = types.ReasonPersonalized
)
