package realtime

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/itemembedding"
	"github.com/tikiclone/tiki/platforms/rec-vector/internal/vectorstore"
)

type Service struct {
	repo        Repository
	itemEmbSvc  *itemembedding.Service
	vectorStore vectorstore.VectorStore
	banditArms  []string
}

func NewService(repo Repository, itemEmbSvc *itemembedding.Service, vectorStore vectorstore.VectorStore) *Service {
	return &Service{
		repo:        repo,
		itemEmbSvc:  itemEmbSvc,
		vectorStore: vectorStore,
		banditArms:  []string{"popular", "collaborative", "content_based", "trending"},
	}
}

func (s *Service) TrackEvent(ctx context.Context, userID, sessionID, eventType, itemID, query string) (*UserSession, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		session = &UserSession{
			UserID:    userID,
			SessionID: sessionID,
			Events:    make([]SessionEvent, 0),
			StartedAt: time.Now(),
		}
	}

	event := SessionEvent{
		EventType: eventType,
		ItemID:    itemID,
		Query:     query,
		Timestamp: time.Now(),
	}
	session.Events = append(session.Events, event)
	session.LastActiveAt = time.Now()

	if err := s.repo.StoreSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) GetSessionEmbedding(ctx context.Context, sessionID string) ([]float64, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var itemVectors [][]float64
	for _, event := range session.Events {
		if event.ItemID == "" {
			continue
		}
		emb, err := s.itemEmbSvc.GetEmbedding(ctx, event.ItemID)
		if err != nil {
			emb, err = s.itemEmbSvc.GenerateItemEmbedding(ctx, event.ItemID, "", nil, "v1")
			if err != nil {
				continue
			}
		}
		itemVectors = append(itemVectors, emb.Embedding)
	}

	if len(itemVectors) == 0 {
		return make([]float64, itemembedding.DefaultDimension), nil
	}

	dim := len(itemVectors[0])
	mean := make([]float64, dim)
	for _, vec := range itemVectors {
		for i, v := range vec {
			mean[i] += v
		}
	}
	for i := range mean {
		mean[i] /= float64(len(itemVectors))
	}

	norm := 0.0
	for _, v := range mean {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range mean {
			mean[i] /= norm
		}
	}

	if session.SessionID == "" {
		session.SessionID = sessionID
	}
	if session.UserID == "" {
		session.UserID = "anon"
	}

	return mean, nil
}

func (s *Service) RecommendWithContext(ctx context.Context, sessionID, namespace string, topK int) ([]vectorstore.SearchResult, error) {
	embedding, err := s.GetSessionEmbedding(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	armID := s.selectArm(ctx)
	results, err := s.vectorStore.Search(ctx, embedding, namespace, topK)
	if err != nil {
		return nil, err
	}

	s.recordReward(ctx, armID, float64(len(results)))
	return results, nil
}

func (s *Service) selectArm(ctx context.Context) string {
	stats, err := s.repo.GetArmStats(ctx)
	if err != nil || len(stats) == 0 {
		stats = make(map[string]*ArmStat)
		for _, arm := range s.banditArms {
			stats[arm] = &ArmStat{ArmID: arm}
		}
	}

	if rand.Float64() < 0.1 {
		return s.banditArms[rand.Intn(len(s.banditArms))]
	}

	var bestArm string
	var bestMean float64
	for _, arm := range s.banditArms {
		stat, ok := stats[arm]
		if !ok {
			return arm
		}
		if stat.Plays == 0 {
			return arm
		}
		if bestArm == "" || stat.MeanReward > bestMean {
			bestArm = arm
			bestMean = stat.MeanReward
		}
	}
	if bestArm == "" {
		return s.banditArms[0]
	}
	return bestArm
}

func (s *Service) recordReward(ctx context.Context, armID string, reward float64) {
	stats, err := s.repo.GetArmStats(ctx)
	if err != nil || stats == nil {
		stats = make(map[string]*ArmStat)
	}
	stat, ok := stats[armID]
	if !ok {
		stat = &ArmStat{ArmID: armID}
		stats[armID] = stat
	}
	stat.Plays++
	stat.Rewards += reward
	stat.MeanReward = stat.Rewards / float64(stat.Plays)
	s.repo.StoreArmStats(ctx, stats)
}

func (s *Service) CreateSession(ctx context.Context, userID string) *UserSession {
	return &UserSession{
		UserID:    userID,
		SessionID: uuid.New().String(),
		Events:    make([]SessionEvent, 0),
		StartedAt: time.Now(),
	}
}
