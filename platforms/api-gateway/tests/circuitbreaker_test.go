package tests

import (
	"testing"

	"github.com/tikiclone/tiki/platforms/api-gateway/internal/circuitbreaker"
)

func TestCircuitBreakerClosedInitialState(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, err := svc.Create("payment-cb", "payment-service", 5, 60, 3)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if cb.State != circuitbreaker.StateClosed {
		t.Errorf("expected closed state, got %s", cb.State)
	}
}

func TestCircuitBreakerClosedToOpen(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, _ := svc.Create("test-cb", "test-svc", 3, 60, 3)

	for i := 0; i < 3; i++ {
		svc.RecordFailure(cb.ID)
	}

	state, _ := svc.GetState(cb.ID)
	if state != circuitbreaker.StateOpen {
		t.Errorf("expected open state, got %s", state)
	}
}

func TestCircuitBreakerOpenToHalfOpen(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, _ := svc.Create("test-cb", "test-svc", 1, 0, 3)
	svc.RecordFailure(cb.ID)

	state, _ := svc.GetState(cb.ID)
	if state != circuitbreaker.StateOpen {
		t.Errorf("expected open state, got %s", state)
	}

	// recovery_timeout is 0, so CanPass should transition to half-open immediately
	canPass, _ := svc.CanPass(cb.ID)
	if !canPass {
		t.Error("should be allowed to pass after recovery timeout")
	}

	state, _ = svc.GetState(cb.ID)
	if state != circuitbreaker.StateHalfOpen {
		t.Errorf("expected half_open state, got %s", state)
	}
}

func TestCircuitBreakerHalfOpenToClosed(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, _ := svc.Create("test-cb", "test-svc", 1, 0, 2)

	svc.RecordFailure(cb.ID)
	svc.CanPass(cb.ID)

	svc.RecordSuccess(cb.ID)
	svc.RecordSuccess(cb.ID)

	state, _ := svc.GetState(cb.ID)
	if state != circuitbreaker.StateClosed {
		t.Errorf("expected closed state after successes, got %s", state)
	}
}

func TestCircuitBreakerHalfOpenToOpen(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, _ := svc.Create("test-cb", "test-svc", 1, 0, 3)

	svc.RecordFailure(cb.ID)
	svc.CanPass(cb.ID)

	svc.RecordFailure(cb.ID)

	state, _ := svc.GetState(cb.ID)
	if state != circuitbreaker.StateOpen {
		t.Errorf("expected open state after half-open failure, got %s", state)
	}
}

func TestCircuitBreakerCanPassClosed(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, _ := svc.Create("test-cb", "test-svc", 5, 60, 3)

	canPass, _ := svc.CanPass(cb.ID)
	if !canPass {
		t.Error("should be allowed to pass when closed")
	}
}

func TestCircuitBreakerCanPassOpen(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, _ := svc.Create("test-cb", "test-svc", 1, 3600, 3)
	svc.RecordFailure(cb.ID)

	canPass, _ := svc.CanPass(cb.ID)
	if canPass {
		t.Error("should NOT be allowed to pass when open without recovery")
	}
}

func TestCircuitBreakerNotFound(t *testing.T) {
	svc := circuitbreaker.NewService(circuitbreaker.NewInMemoryRepository())

	canPass, _ := svc.CanPass("nonexistent")
	if !canPass {
		t.Error("should allow pass when breaker not found")
	}
}

func TestCircuitBreakerList(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	svc.Create("cb-1", "svc-1", 5, 60, 3)
	svc.Create("cb-2", "svc-2", 3, 30, 5)

	list, err := svc.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 breakers, got %d", len(list))
	}
}

func TestCircuitBreakerClosedSuccessResetsFailureCount(t *testing.T) {
	repo := circuitbreaker.NewInMemoryRepository()
	svc := circuitbreaker.NewService(repo)

	cb, _ := svc.Create("test-cb", "test-svc", 3, 60, 3)

	svc.RecordFailure(cb.ID)
	svc.RecordFailure(cb.ID)
	svc.RecordSuccess(cb.ID)
	svc.RecordFailure(cb.ID)

	state, _ := svc.GetState(cb.ID)
	if state != circuitbreaker.StateClosed {
		t.Errorf("expected closed (success reset failure count), got %s", state)
	}
}
