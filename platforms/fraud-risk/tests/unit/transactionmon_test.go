package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/devicefp"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/transactionmon"
)

func TestRecordTransactionNewUser(t *testing.T) {
	repo := transactionmon.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	deviceSvc := devicefp.NewService(deviceRepo)
	mon := transactionmon.NewMonitor(repo, deviceSvc)

	rec := &transactionmon.TransactionRecord{
		UserID: "user1",
		Amount: 100.50,
	}

	result, err := mon.RecordTransaction(context.Background(), rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DailyCount != 1 {
		t.Errorf("expected daily count 1, got %d", result.DailyCount)
	}
	if result.DailyVolume != 100.50 {
		t.Errorf("expected daily volume 100.50, got %f", result.DailyVolume)
	}
}

func TestRecordTransactionMultiple(t *testing.T) {
	repo := transactionmon.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	deviceSvc := devicefp.NewService(deviceRepo)
	mon := transactionmon.NewMonitor(repo, deviceSvc)

	mon.RecordTransaction(context.Background(), &transactionmon.TransactionRecord{UserID: "u1", Amount: 50})
	mon.RecordTransaction(context.Background(), &transactionmon.TransactionRecord{UserID: "u1", Amount: 150})
	mon.RecordTransaction(context.Background(), &transactionmon.TransactionRecord{UserID: "u1", Amount: 200})

	result, err := mon.GetPattern(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DailyCount != 3 {
		t.Errorf("expected daily count 3, got %d", result.DailyCount)
	}
	if result.DailyVolume != 400 {
		t.Errorf("expected daily volume 400, got %f", result.DailyVolume)
	}
}

func TestRecordTransactionAvgTicket(t *testing.T) {
	repo := transactionmon.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	deviceSvc := devicefp.NewService(deviceRepo)
	mon := transactionmon.NewMonitor(repo, deviceSvc)

	mon.RecordTransaction(context.Background(), &transactionmon.TransactionRecord{UserID: "u2", Amount: 100})
	mon.RecordTransaction(context.Background(), &transactionmon.TransactionRecord{UserID: "u2", Amount: 200})

	result, _ := mon.GetPattern(context.Background(), "u2")
	if result.AvgTicket != 150 {
		t.Errorf("expected avg ticket 150, got %f", result.AvgTicket)
	}
}

func TestDetectAnomalyNoAnomaly(t *testing.T) {
	repo := transactionmon.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	deviceSvc := devicefp.NewService(deviceRepo)
	mon := transactionmon.NewMonitor(repo, deviceSvc)

	rec := &transactionmon.TransactionRecord{UserID: "user3", Amount: 50}
	mon.RecordTransaction(context.Background(), rec)

	result, err := mon.DetectAnomaly(context.Background(), rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasAnomaly {
		t.Error("expected no anomaly for single normal transaction")
	}
}

func TestDetectAnomalyAmountAnomaly(t *testing.T) {
	repo := transactionmon.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	deviceSvc := devicefp.NewService(deviceRepo)
	mon := transactionmon.NewMonitor(repo, deviceSvc)

	rec1 := &transactionmon.TransactionRecord{UserID: "u4", Amount: 100}
	mon.RecordTransaction(context.Background(), rec1)

	rec2 := &transactionmon.TransactionRecord{UserID: "u4", Amount: 100}
	mon.RecordTransaction(context.Background(), rec2)

	anomalyRec := &transactionmon.TransactionRecord{UserID: "u4", Amount: 1000}
	result, err := mon.DetectAnomaly(context.Background(), anomalyRec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasAnomaly {
		t.Error("expected anomaly for amount > 3x avg")
	}
}

func TestDetectAnomalyVelocity(t *testing.T) {
	repo := transactionmon.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	deviceSvc := devicefp.NewService(deviceRepo)
	mon := transactionmon.NewMonitor(repo, deviceSvc)

	for i := 0; i < 15; i++ {
		mon.RecordTransaction(context.Background(), &transactionmon.TransactionRecord{
			UserID: "u5", Amount: 10,
		})
	}

	result, err := mon.DetectAnomaly(context.Background(), &transactionmon.TransactionRecord{
		UserID: "u5", Amount: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasAnomaly {
		t.Error("expected anomaly for high velocity")
	}
}

func TestGetPatternNotFound(t *testing.T) {
	repo := transactionmon.NewInMemoryRepository()
	deviceRepo := devicefp.NewInMemoryRepository()
	deviceSvc := devicefp.NewService(deviceRepo)
	mon := transactionmon.NewMonitor(repo, deviceSvc)

	_, err := mon.GetPattern(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}
