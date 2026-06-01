package unit

import (
	"context"
	"testing"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/application"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type mockPublisher struct{}

func (m *mockPublisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	return nil
}

type mockLivestreamRepo struct {
	store map[string]*domain.Livestream
}

func newMockLivestreamRepo() *mockLivestreamRepo {
	return &mockLivestreamRepo{store: make(map[string]*domain.Livestream)}
}

func (r *mockLivestreamRepo) Create(ctx context.Context, ls *domain.Livestream) error {
	r.store[ls.ID] = ls
	return nil
}

func (r *mockLivestreamRepo) GetByID(ctx context.Context, id string) (*domain.Livestream, error) {
	ls, ok := r.store[id]
	if !ok {
		return nil, domain.ErrLivestreamNotFound
	}
	return ls, nil
}

func (r *mockLivestreamRepo) Update(ctx context.Context, ls *domain.Livestream) error {
	r.store[ls.ID] = ls
	return nil
}

func (r *mockLivestreamRepo) ListBySeller(ctx context.Context, sellerID string, offset, limit int) ([]*domain.Livestream, int64, error) {
	return nil, 0, nil
}

func (r *mockLivestreamRepo) ListActive(ctx context.Context, offset, limit int) ([]*domain.Livestream, int64, error) {
	return nil, 0, nil
}

type mockMessageRepo struct{}

func (r *mockMessageRepo) Save(ctx context.Context, msg *domain.ChatMessage) error { return nil }
func (r *mockMessageRepo) GetByRoom(ctx context.Context, roomID string, offset, limit int) ([]*domain.ChatMessage, int64, error) { return nil, 0, nil }
func (r *mockMessageRepo) GetLastSequence(ctx context.Context, roomID string) (int64, error) { return 0, nil }
func (r *mockMessageRepo) MarkModerated(ctx context.Context, messageID string) error { return nil }
func (r *mockMessageRepo) Delete(ctx context.Context, messageID string) error { return nil }

type mockReactionRepo struct{}

func (r *mockReactionRepo) Save(ctx context.Context, reaction *domain.Reaction) error { return nil }
func (r *mockReactionRepo) GetCountByRoom(ctx context.Context, roomID string, reactionType string) (int64, error) { return 0, nil }
func (r *mockReactionRepo) GetSummaryByRoom(ctx context.Context, roomID string) (map[string]int64, error) { return nil, nil }

type mockGiftRepo struct{}
func (r *mockGiftRepo) Save(ctx context.Context, gift *domain.Gift) error { return nil }
func (r *mockGiftRepo) GetLeaderboardByRoom(ctx context.Context, roomID string, limit int) ([]*domain.GiftLeaderboardEntry, error) { return nil, nil }

type mockPinnedRepo struct{}
func (r *mockPinnedRepo) Pin(ctx context.Context, pp *domain.PinnedProduct) error { return nil }
func (r *mockPinnedRepo) Unpin(ctx context.Context, livestreamID, productID string) error { return nil }
func (r *mockPinnedRepo) GetActiveByLivestream(ctx context.Context, livestreamID string) ([]*domain.PinnedProduct, error) { return nil, nil }

type mockModRepo struct{}
func (r *mockModRepo) SaveAction(ctx context.Context, action *domain.ModerationAction) error { return nil }
func (r *mockModRepo) IsUserMuted(ctx context.Context, roomID, userID string) (bool, error) { return false, nil }
func (r *mockModRepo) GetMuteDuration(ctx context.Context, roomID, userID string) (int64, error) { return 0, nil }

func TestCreateLivestream(t *testing.T) {
	svc := application.NewLiveCommerceService(
		newMockLivestreamRepo(),
		&mockMessageRepo{},
		&mockReactionRepo{},
		&mockGiftRepo{},
		&mockPinnedRepo{},
		&mockModRepo{},
		&mockPublisher{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	ls, err := svc.CreateLivestream(context.Background(), "seller1", "Test Stream", "", "", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ls.SellerID != "seller1" {
		t.Errorf("expected seller1, got %s", ls.SellerID)
	}
	if ls.Title != "Test Stream" {
		t.Errorf("expected Test Stream, got %s", ls.Title)
	}
}

func TestLivestreamLifecycle(t *testing.T) {
	repo := newMockLivestreamRepo()
	svc := application.NewLiveCommerceService(
		repo, &mockMessageRepo{}, &mockReactionRepo{}, &mockGiftRepo{},
		&mockPinnedRepo{}, &mockModRepo{}, &mockPublisher{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	ls, _ := svc.CreateLivestream(context.Background(), "seller1", "Test", "", "", "", nil, nil)
	if err := svc.StartLivestream(context.Background(), ls.ID); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	updated, _ := repo.GetByID(context.Background(), ls.ID)
	if !updated.IsLive() {
		t.Error("expected live after start")
	}
	if err := svc.EndLivestream(context.Background(), ls.ID); err != nil {
		t.Fatalf("end failed: %v", err)
	}
	ended, _ := repo.GetByID(context.Background(), ls.ID)
	if ended.Status != domain.LiveStatusEnded {
		t.Errorf("expected ended, got %s", ended.Status)
	}
}
