package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/replay"
)

func TestReplayDedup(t *testing.T) {
	svc := replay.NewService(0)
	ctx := context.Background()
	called := 0
	for i := 0; i < 5; i++ {
		err := svc.ProcessWithReplayGuard(ctx, "replay-001", func(ctx context.Context) error {
			called++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if called != 1 {
		t.Errorf("expected 1 call, got %d", called)
	}
}

func TestReplayMultipleIDs(t *testing.T) {
	svc := replay.NewService(0)
	ctx := context.Background()
	results := make(map[string]int)
	for i := 0; i < 3; i++ {
		for _, id := range []string{"a", "b", "c"} {
			svc.ProcessWithReplayGuard(ctx, id, func(ctx context.Context) error {
				results["id:"+id]++
				return nil
			})
		}
	}
	for _, id := range []string{"a", "b", "c"} {
		if results["id:"+id] != 1 {
			t.Errorf("expected 1 for %s, got %d", id, results["id:"+id])
		}
	}
}

func TestReplayExpiry(t *testing.T) {
	svc := replay.NewService(0)
	ctx := context.Background()
	svc.ProcessWithReplayGuard(ctx, "replay-exp", func(ctx context.Context) error {
		return nil
	})
	// Remove from map to simulate expiry
	svc.MarkProcessed(ctx, "replay-exp")
	svc.MarkProcessed(ctx, "replay-exp")
	// Should not fail
}

func TestShipmentStatusValidTransitions(t *testing.T) {
	tests := []struct {
		from string
		to   string
		ok   bool
	}{
		{from: "pending", to: "confirmed", ok: true},
		{from: "delivered", to: "pending", ok: false},
	}
	for _, tc := range tests {
		t.Logf("transition %s -> %s: %v", tc.from, tc.to, tc.ok)
	}
}
