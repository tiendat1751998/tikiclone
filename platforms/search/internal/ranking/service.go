package ranking

import (
	"context"
	"math"
	"sort"
	"time"
	"strings"

	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

type Service interface {
	Score(ctx context.Context, doc *search.ProductDocument, query string) (RankScore, []RankingFactor)
	Rank(ctx context.Context, docs []search.ProductDocument, query string) []search.ProductDocument
	GetConfig(ctx context.Context) (*RankingConfig, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Score(ctx context.Context, doc *search.ProductDocument, query string) (RankScore, []RankingFactor) {
	cfg, err := s.repo.GetConfig(ctx)
	if err != nil {
		cfg = &RankingConfig{}
	}

	var factors []RankingFactor

	titleScore := 0.0
	if query != "" {
		qLower := strings.ToLower(query)
		titleLower := strings.ToLower(doc.Title)
		if strings.Contains(titleLower, qLower) {
			titleScore = 1.0
			if titleLower == qLower {
				titleScore = 1.5
			}
		} else {
			for _, term := range strings.Fields(qLower) {
				if strings.Contains(titleLower, term) {
					titleScore += 0.5
				}
			}
			if titleScore > 1.5 {
				titleScore = 1.5
			}
		}
	}
	factors = append(factors, RankingFactor{
		Name: "title_match", Weight: cfg.TitleBoost, Value: titleScore,
	})

	categoryScore := 0.0
	if query != "" && doc.Category != "" {
		qLower := strings.ToLower(query)
		catLower := strings.ToLower(doc.Category)
		if strings.Contains(catLower, qLower) || strings.Contains(qLower, catLower) {
			categoryScore = 1.0
		}
	}
	factors = append(factors, RankingFactor{
		Name: "category_match", Weight: cfg.CategoryBoost, Value: categoryScore,
	})

	ratingScore := doc.Rating / 5.0
	factors = append(factors, RankingFactor{
		Name: "rating", Weight: cfg.RatingBoost, Value: ratingScore,
	})

	hoursSinceCreation := time.Since(doc.CreatedAt).Hours()
	recencyScore := math.Exp(-hoursSinceCreation / 720)
	factors = append(factors, RankingFactor{
		Name: "recency", Weight: cfg.RecencyBoost, Value: recencyScore,
	})

	popularityScore := math.Min(1.0, float64(doc.Stock)/1000.0)
	factors = append(factors, RankingFactor{
		Name: "popularity", Weight: cfg.PopularityBoost, Value: popularityScore,
	})

	signals, _ := s.repo.GetClickSignals(ctx, doc.ID)
	clickScore := 0.0
	if signals != nil {
		clickScore = signals.CTR
	}
	factors = append(factors, RankingFactor{
		Name: "clicks", Weight: cfg.ClickBoost, Value: clickScore,
	})

	totalScore := RankScore(0)
	for _, f := range factors {
		totalScore += RankScore(f.Weight * f.Value)
	}

	return totalScore, factors
}

func (s *service) Rank(ctx context.Context, docs []search.ProductDocument, query string) []search.ProductDocument {
	type scoredDoc struct {
		doc   search.ProductDocument
		score RankScore
	}

	scored := make([]scoredDoc, len(docs))
	for i, doc := range docs {
		score, _ := s.Score(ctx, &doc, query)
		scored[i] = scoredDoc{doc: doc, score: score}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	result := make([]search.ProductDocument, len(scored))
	for i, sd := range scored {
		result[i] = sd.doc
	}

	return result
}

func (s *service) GetConfig(ctx context.Context) (*RankingConfig, error) {
	return s.repo.GetConfig(ctx)
}
