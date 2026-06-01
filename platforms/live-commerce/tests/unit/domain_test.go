package unit

import (
	"testing"
	"time"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

func TestNewLivestream(t *testing.T) {
	now := time.Now()
	ls := domain.NewLivestream("seller1", "Test Stream", "A test", "cover.jpg", "gaming", []string{"game", "fun"}, &now)
	if ls.Status != domain.LiveStatusScheduled {
		t.Errorf("expected scheduled, got %s", ls.Status)
	}
	if ls.SellerID != "seller1" {
		t.Errorf("expected seller1, got %s", ls.SellerID)
	}
	if ls.Title != "Test Stream" {
		t.Errorf("expected Test Stream, got %s", ls.Title)
	}
}

func TestLivestreamStart(t *testing.T) {
	now := time.Now()
	ls := domain.NewLivestream("seller1", "Test", "", "", "", nil, &now)
	if err := ls.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ls.IsLive() {
		t.Error("expected live status")
	}
	if err := ls.Start(); err == nil {
		t.Error("expected error starting already live stream")
	}
}

func TestLivestreamEnd(t *testing.T) {
	ls := domain.NewLivestream("seller1", "Test", "", "", "", nil, nil)
	ls.Start()
	if err := ls.End(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ls.Status != domain.LiveStatusEnded {
		t.Errorf("expected ended, got %s", ls.Status)
	}
}

func TestLivestreamCancel(t *testing.T) {
	ls := domain.NewLivestream("seller1", "Test", "", "", "", nil, nil)
	if err := ls.Cancel(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ls.Status != domain.LiveStatusCancelled {
		t.Errorf("expected cancelled, got %s", ls.Status)
	}
	if err := ls.Start(); err == nil {
		t.Error("expected error: cannot start cancelled stream")
	}
}

func TestLivestreamViewers(t *testing.T) {
	ls := domain.NewLivestream("seller1", "Test", "", "", "", nil, nil)
	ls.UpdateViewers(100)
	if ls.ViewerCount != 100 {
		t.Errorf("expected 100 viewers, got %d", ls.ViewerCount)
	}
	if ls.PeakViewers != 100 {
		t.Errorf("expected peak 100, got %d", ls.PeakViewers)
	}
	ls.UpdateViewers(50)
	if ls.PeakViewers != 100 {
		t.Errorf("expected peak still 100, got %d", ls.PeakViewers)
	}
	ls.UpdateViewers(200)
	if ls.PeakViewers != 200 {
		t.Errorf("expected peak 200, got %d", ls.PeakViewers)
	}
}

func TestNewChatMessage(t *testing.T) {
	msg := domain.NewChatMessage("room1", "user1", "testuser", "hello", domain.MsgTypeText)
	if msg.RoomID != "room1" || msg.UserID != "user1" || msg.Content != "hello" {
		t.Error("chat message fields mismatch")
	}
	if msg.Type != domain.MsgTypeText {
		t.Errorf("expected text type, got %s", msg.Type)
	}
}

func TestNewReaction(t *testing.T) {
	r := domain.NewReaction("room1", "user1", domain.ReactionLike)
	if r.Type != domain.ReactionLike {
		t.Errorf("expected like, got %s", r.Type)
	}
	if r.RoomID != "room1" {
		t.Errorf("expected room1, got %s", r.RoomID)
	}
}

func TestErrorTypes(t *testing.T) {
	if domain.ErrLivestreamNotFound.Error() != "live: livestream_not_found" {
		t.Errorf("unexpected error message: %s", domain.ErrLivestreamNotFound.Error())
	}
	if domain.ErrInvalidLiveState.Error() != "live: invalid_live_state" {
		t.Errorf("unexpected error message: %s", domain.ErrInvalidLiveState.Error())
	}
}

func TestNewGift(t *testing.T) {
	g := &domain.Gift{
		ID:       "gift1",
		RoomID:   "room1",
		UserID:   "user1",
		GiftType: "diamond",
		Amount:   100,
		Currency: "VND",
	}
	if g.Amount != 100 || g.GiftType != "diamond" {
		t.Error("gift fields mismatch")
	}
}

func TestPinnedProduct(t *testing.T) {
	pp := &domain.PinnedProduct{
		ID:        "pin1",
		ProductID: "prod1",
		ProductName: "Test Product",
		Price:     50000,
		IsActive:  true,
	}
	if !pp.IsActive {
		t.Error("expected active")
	}
}

func TestModerationAction(t *testing.T) {
	action := &domain.ModerationAction{
		Action:      domain.ModActionMute,
		Reason:      "spam",
		DurationSec: 300,
	}
	if action.Action != domain.ModActionMute {
		t.Errorf("expected mute, got %s", action.Action)
	}
	if action.DurationSec != 300 {
		t.Errorf("expected 300s, got %d", action.DurationSec)
	}
}

func TestStateTransitions(t *testing.T) {
	transitionTests := []struct {
		from string
		to   string
		fn   func(*domain.Livestream) error
	}{
		{domain.LiveStatusScheduled, domain.LiveStatusLive, func(ls *domain.Livestream) error { return ls.Start() }},
		{domain.LiveStatusScheduled, domain.LiveStatusCancelled, func(ls *domain.Livestream) error { return ls.Cancel() }},
		{domain.LiveStatusLive, domain.LiveStatusEnded, func(ls *domain.Livestream) error { return ls.End() }},
	}
	for _, tt := range transitionTests {
		ls := &domain.Livestream{Status: tt.from}
		if err := tt.fn(ls); err != nil {
			t.Errorf("transition %s->%s failed: %v", tt.from, tt.to, err)
		}
	}
	invalidTransitions := []struct {
		from string
		to   string
		fn   func(*domain.Livestream) error
	}{
		{domain.LiveStatusEnded, domain.LiveStatusLive, func(ls *domain.Livestream) error { return ls.Start() }},
		{domain.LiveStatusCancelled, domain.LiveStatusLive, func(ls *domain.Livestream) error { return ls.Start() }},
		{domain.LiveStatusLive, domain.LiveStatusCancelled, func(ls *domain.Livestream) error { return ls.Cancel() }},
	}
	for _, tt := range invalidTransitions {
		ls := &domain.Livestream{Status: tt.from}
		if err := tt.fn(ls); err == nil {
			t.Errorf("expected error for transition %s->%s", tt.from, tt.to)
		}
	}
}

func TestMaxMessageLength(t *testing.T) {
	short := "hello"
	if len(short) > domain.MaxMsgLength {
		t.Error("unexpected max length")
	}
	long := make([]byte, domain.MaxMsgLength+1)
	if len(long) <= domain.MaxMsgLength {
		t.Error("expected to exceed max length")
	}
}
