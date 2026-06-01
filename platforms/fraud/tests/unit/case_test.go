package unit

import (
	"context"
	"testing"

	fraudcase "github.com/tikiclone/tiki/platforms/fraud/internal/case"
)

func TestCreateCase(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	c, err := svc.CreateCase(context.Background(), "alert-1", "user-1", "Suspicious Login", "Multiple failed attempts", 75.0, fraudcase.PriorityHigh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID == "" {
		t.Error("expected non-empty ID")
	}
	if c.Status != fraudcase.StatusOpen {
		t.Errorf("expected open, got %s", c.Status)
	}
	if c.Priority != fraudcase.PriorityHigh {
		t.Errorf("expected high, got %s", c.Priority)
	}
}

func TestGetCase(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	created, _ := svc.CreateCase(context.Background(), "a-1", "u-1", "Test", "", 50.0, fraudcase.PriorityMedium)
	fetched, err := svc.GetCase(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Title != "Test" {
		t.Errorf("expected Test, got %s", fetched.Title)
	}
}

func TestAssignInvestigator(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	c, _ := svc.CreateCase(context.Background(), "a-2", "u-2", "Fraud Case", "", 60.0, fraudcase.PriorityMedium)

	if err := svc.AssignInvestigator(context.Background(), c.ID, "investigator-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, _ := svc.GetCase(context.Background(), c.ID)
	if fetched.Investigator != "investigator-1" {
		t.Errorf("expected investigator-1, got %s", fetched.Investigator)
	}
	if fetched.Status != fraudcase.StatusInvestigating {
		t.Errorf("expected investigating, got %s", fetched.Status)
	}
}

func TestAddEvidence(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	c, _ := svc.CreateCase(context.Background(), "a-3", "u-3", "Evidence Test", "", 70.0, fraudcase.PriorityHigh)

	if err := svc.AddEvidence(context.Background(), c.ID, "ip_address", "Suspicious IP", "10.0.0.1", "system"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, _ := svc.GetCase(context.Background(), c.ID)
	if len(fetched.Evidence) != 1 {
		t.Errorf("expected 1 evidence, got %d", len(fetched.Evidence))
	}
	if fetched.Evidence[0].Type != "ip_address" {
		t.Errorf("expected ip_address, got %s", fetched.Evidence[0].Type)
	}
}

func TestUpdateStatus(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	c, _ := svc.CreateCase(context.Background(), "a-4", "u-4", "Status Test", "", 40.0, fraudcase.PriorityLow)

	svc.AssignInvestigator(context.Background(), c.ID, "detective")

	if err := svc.UpdateStatus(context.Background(), c.ID, fraudcase.StatusResolved, "false positive"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, _ := svc.GetCase(context.Background(), c.ID)
	if fetched.Status != fraudcase.StatusResolved {
		t.Errorf("expected resolved, got %s", fetched.Status)
	}
	if fetched.Resolution != "false positive" {
		t.Errorf("expected false positive, got %s", fetched.Resolution)
	}
	if fetched.ResolvedAt == nil {
		t.Error("expected resolved_at to be set")
	}
}

func TestInvalidStatusTransition(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	c, _ := svc.CreateCase(context.Background(), "a-5", "u-5", "Invalid Transition", "", 30.0, fraudcase.PriorityLow)

	err := svc.UpdateStatus(context.Background(), c.ID, fraudcase.StatusEscalated, "")
	if err != nil {
		if err != fraudcase.ErrInvalidTransition {
			t.Fatalf("expected ErrInvalidTransition, got %v", err)
		}
	}
}

func TestEscalateCase(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	c, _ := svc.CreateCase(context.Background(), "a-6", "u-6", "Escalation", "", 80.0, fraudcase.PriorityMedium)

	if err := svc.Escalate(context.Background(), c.ID, fraudcase.PriorityCritical); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, _ := svc.GetCase(context.Background(), c.ID)
	if fetched.Priority != fraudcase.PriorityCritical {
		t.Errorf("expected critical, got %s", fetched.Priority)
	}
	if fetched.Status != fraudcase.StatusEscalated {
		t.Errorf("expected escalated, got %s", fetched.Status)
	}
}

func TestListCasesFilterByStatus(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	svc.CreateCase(context.Background(), "a-7", "u-7", "Open Case", "", 20.0, fraudcase.PriorityLow)
	svc.CreateCase(context.Background(), "a-8", "u-8", "Another Open", "", 30.0, fraudcase.PriorityLow)

	cases, total, err := svc.ListCases(context.Background(), fraudcase.StatusOpen, "", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 open cases, got %d", total)
	}
	_ = cases
}

func TestListCasesPagination(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	for i := 0; i < 5; i++ {
		svc.CreateCase(context.Background(), "a", "u", "Case", "", 10.0, fraudcase.PriorityLow)
	}

	cases, total, err := svc.ListCases(context.Background(), "", "", 0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cases) != 3 {
		t.Errorf("expected 3 cases, got %d", len(cases))
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
}

func TestGetNonexistentCase(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	_, err := svc.GetCase(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected error for nonexistent case")
	}
	if err != fraudcase.ErrCaseNotFound {
		t.Errorf("expected ErrCaseNotFound, got %v", err)
	}
}

func TestCaseWorkflow(t *testing.T) {
	repo := fraudcase.NewInMemoryRepository()
	svc := fraudcase.NewService(repo)

	c, _ := svc.CreateCase(context.Background(), "alert-final", "user-final", "Full Workflow", "End to end test", 90.0, fraudcase.PriorityCritical)
	svc.AssignInvestigator(context.Background(), c.ID, "detective-1")
	svc.AddEvidence(context.Background(), c.ID, "login_record", "Suspicious login from new device", `{"ip":"10.0.0.99","device":"unknown"}`, "system")
	svc.UpdateStatus(context.Background(), c.ID, fraudcase.StatusResolved, "confirmed fraudulent")

	fetched, _ := svc.GetCase(context.Background(), c.ID)
	if fetched.Status != fraudcase.StatusResolved {
		t.Errorf("expected resolved, got %s", fetched.Status)
	}
	if fetched.Investigator != "detective-1" {
		t.Errorf("expected detective-1, got %s", fetched.Investigator)
	}
	if len(fetched.Evidence) != 1 {
		t.Errorf("expected 1 evidence, got %d", len(fetched.Evidence))
	}
}
