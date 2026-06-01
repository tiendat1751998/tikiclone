package recommendations

import (
	"context"
	"sort"
	"sync"
	"time"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type TrendingScore struct {
	LivestreamID string  `json:"livestream_id"`
	Score        float64 `json:"score"`
	ViewerCount  int64   `json:"viewer_count"`
	Engagement   int64   `json:"engagement"`
}

type Engine struct {
	mu         sync.RWMutex
	trending   []*TrendingScore
	updatedAt  time.Time
}

func NewEngine() *Engine {
	return &Engine{
		trending: make([]*TrendingScore, 0),
	}
}

func (e *Engine) UpdateTrending(ctx context.Context, streams []*domain.Livestream) {
	e.mu.Lock()
	defer e.mu.Unlock()
	scores := make([]*TrendingScore, 0, len(streams))
	for _, s := range streams {
		engagement := s.TotalLikes + s.TotalGifts*10 + s.TotalShares*5
		score := float64(s.ViewerCount)*0.4 + float64(engagement)*0.6
		scores = append(scores, &TrendingScore{
			LivestreamID: s.ID,
			Score:        score,
			ViewerCount:  s.ViewerCount,
			Engagement:   engagement,
		})
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	if len(scores) > 50 {
		scores = scores[:50]
	}
	e.trending = scores
	e.updatedAt = time.Now()
}

func (e *Engine) GetTrending(ctx context.Context, limit int) []*TrendingScore {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if limit > len(e.trending) {
		limit = len(e.trending)
	}
	result := make([]*TrendingScore, limit)
	copy(result, e.trending[:limit])
	return result
}

func (e *Engine) GetRecommended(ctx context.Context, userID string, limit int) []*TrendingScore {
	return e.GetTrending(ctx, limit)
}
