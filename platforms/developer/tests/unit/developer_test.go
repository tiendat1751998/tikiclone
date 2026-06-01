package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/developer/internal/apikeys"
	"github.com/tikiclone/tiki/platforms/developer/internal/cicd"
	"github.com/tikiclone/tiki/platforms/developer/internal/docs"
	"github.com/tikiclone/tiki/platforms/developer/internal/onboarding"
	"github.com/tikiclone/tiki/platforms/developer/internal/sdk"
	"github.com/tikiclone/tiki/platforms/developer/internal/webhooks"
)

func TestGenerateAndValidateAPIKey(t *testing.T) {
	repo := apikeys.NewInMemoryRepository()
	svc := apikeys.NewService(repo)

	key, rawKey, err := svc.Generate(context.Background(), "test-key", []string{"read", "write"}, "test-service", time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
	if rawKey == "" {
		t.Fatal("expected non-empty raw key")
	}

	validated, valid := svc.Validate(context.Background(), rawKey)
	if !valid {
		t.Fatal("expected valid key")
	}
	if validated == nil {
		t.Fatal("expected non-nil validated key")
	}
	if validated.Name != "test-key" {
		t.Errorf("expected name test-key, got %s", validated.Name)
	}
}

func TestRevokeAPIKey(t *testing.T) {
	repo := apikeys.NewInMemoryRepository()
	svc := apikeys.NewService(repo)

	key, rawKey, _ := svc.Generate(context.Background(), "revoke-key", []string{"read"}, "svc", time.Now().Add(24*time.Hour))

	svc.Revoke(context.Background(), key.ID)

	_, valid := svc.Validate(context.Background(), rawKey)
	if valid {
		t.Fatal("expected revoked key to be invalid")
	}
}

func TestListAPIKeys(t *testing.T) {
	repo := apikeys.NewInMemoryRepository()
	svc := apikeys.NewService(repo)

	svc.Generate(context.Background(), "key1", nil, "svc1", time.Now().Add(24*time.Hour))
	svc.Generate(context.Background(), "key2", nil, "svc2", time.Now().Add(24*time.Hour))

	keys, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestRotateAPIKey(t *testing.T) {
	repo := apikeys.NewInMemoryRepository()
	svc := apikeys.NewService(repo)

	key, rawKey, _ := svc.Generate(context.Background(), "rotate-key", []string{"read"}, "svc", time.Now().Add(24*time.Hour))

	_, newRawKey, err := svc.Rotate(context.Background(), key.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, oldValid := svc.Validate(context.Background(), rawKey)
	if oldValid {
		t.Fatal("expected old key to be invalid after rotation")
	}

	_, newValid := svc.Validate(context.Background(), newRawKey)
	if !newValid {
		t.Fatal("expected new key to be valid")
	}
}

func TestInvalidAPIKey(t *testing.T) {
	svc := apikeys.NewService(apikeys.NewInMemoryRepository())

	_, valid := svc.Validate(context.Background(), "nonexistent-key")
	if valid {
		t.Fatal("expected invalid key")
	}
}

func TestExpiredAPIKey(t *testing.T) {
	repo := apikeys.NewInMemoryRepository()
	svc := apikeys.NewService(repo)

	_, rawKey, _ := svc.Generate(context.Background(), "expired-key", nil, "svc", time.Now().Add(-1*time.Hour))

	_, valid := svc.Validate(context.Background(), rawKey)
	if valid {
		t.Fatal("expected expired key to be invalid")
	}
}

func TestCreateDocPage(t *testing.T) {
	repo := docs.NewInMemoryRepository()
	svc := docs.NewService(repo)

	doc, err := svc.Create(context.Background(), "Test Doc", "# Content", "svc1", "guide", []string{"tag1"}, "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil doc")
	}
	if doc.Title != "Test Doc" {
		t.Errorf("expected 'Test Doc', got '%s'", doc.Title)
	}
}

func TestGetDocPage(t *testing.T) {
	repo := docs.NewInMemoryRepository()
	svc := docs.NewService(repo)

	created, _ := svc.Create(context.Background(), "Get Test", "content", "svc1", "guide", nil, "1.0")
	doc, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil doc")
	}
}

func TestUpdateDocPage(t *testing.T) {
	repo := docs.NewInMemoryRepository()
	svc := docs.NewService(repo)

	created, _ := svc.Create(context.Background(), "Original", "content", "svc1", "guide", nil, "1.0")
	updated, err := svc.Update(context.Background(), created.ID, "Updated Title", "", "", "", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil {
		t.Fatal("expected non-nil updated doc")
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected 'Updated Title', got '%s'", updated.Title)
	}
}

func TestSearchDocsByTitle(t *testing.T) {
	repo := docs.NewInMemoryRepository()
	svc := docs.NewService(repo)

	svc.Create(context.Background(), "Getting Started Guide", "content", "svc1", "guide", nil, "1.0")
	svc.Create(context.Background(), "API Reference", "content", "svc1", "ref", nil, "1.0")

	results, err := svc.Search(context.Background(), "Getting")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
}

func TestSearchDocsByContent(t *testing.T) {
	repo := docs.NewInMemoryRepository()
	svc := docs.NewService(repo)

	svc.Create(context.Background(), "Doc1", "this is about authentication", "svc1", "guide", nil, "1.0")
	svc.Create(context.Background(), "Doc2", "something else", "svc1", "guide", nil, "1.0")

	results, _ := svc.Search(context.Background(), "authentication")
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestListDocsByService(t *testing.T) {
	repo := docs.NewInMemoryRepository()
	svc := docs.NewService(repo)

	svc.Create(context.Background(), "Doc1", "c1", "svc-a", "guide", nil, "1.0")
	svc.Create(context.Background(), "Doc2", "c2", "svc-b", "guide", nil, "1.0")
	svc.Create(context.Background(), "Doc3", "c3", "svc-a", "guide", nil, "1.0")

	results, _ := svc.List(context.Background(), "svc-a", "")
	if len(results) != 2 {
		t.Errorf("expected 2 docs for svc-a, got %d", len(results))
	}
}

func TestListDocsByCategory(t *testing.T) {
	repo := docs.NewInMemoryRepository()
	svc := docs.NewService(repo)

	svc.Create(context.Background(), "Doc1", "c1", "svc1", "tutorial", nil, "1.0")
	svc.Create(context.Background(), "Doc2", "c2", "svc1", "reference", nil, "1.0")

	results, _ := svc.List(context.Background(), "", "tutorial")
	if len(results) != 1 {
		t.Errorf("expected 1 doc in tutorial, got %d", len(results))
	}
}

func TestNonExistentDocPage(t *testing.T) {
	svc := docs.NewService(docs.NewInMemoryRepository())
	doc, err := svc.GetByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc != nil {
		t.Error("expected nil for non-existent doc")
	}
}

func TestDocSearchNoResults(t *testing.T) {
	svc := docs.NewService(docs.NewInMemoryRepository())
	svc.Create(context.Background(), "Title", "content", "svc", "cat", nil, "1.0")

	results, _ := svc.Search(context.Background(), "zzzzzzz")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestRegisterSDK(t *testing.T) {
	repo := sdk.NewInMemoryRepository()
	svc := sdk.NewService(repo)

	s, err := svc.Register(context.Background(), "Go SDK", "go", "1.0.0", "https://github.com/test", "https://docs.test", "go 1.22+")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil SDK")
	}
	if s.Name != "Go SDK" {
		t.Errorf("expected 'Go SDK', got '%s'", s.Name)
	}
}

func TestListSDKsByLanguage(t *testing.T) {
	repo := sdk.NewInMemoryRepository()
	svc := sdk.NewService(repo)

	svc.Register(context.Background(), "Go SDK", "go", "1.0.0", "", "", "go 1.22+")
	svc.Register(context.Background(), "Node SDK", "node", "2.0.0", "", "", "node 18+")
	svc.Register(context.Background(), "Go SDK v2", "go", "2.0.0", "", "", "go 1.22+")

	sdks, _ := svc.List(context.Background(), "go")
	if len(sdks) != 2 {
		t.Errorf("expected 2 Go SDKs, got %d", len(sdks))
	}
}

func TestMarkSDKAsLatest(t *testing.T) {
	repo := sdk.NewInMemoryRepository()
	svc := sdk.NewService(repo)

	s1, _ := svc.Register(context.Background(), "Go SDK v1", "go", "1.0.0", "", "", "go 1.22+")
	s2, _ := svc.Register(context.Background(), "Go SDK v2", "go", "2.0.0", "", "", "go 1.22+")

	updated, _ := svc.MarkLatest(context.Background(), s2.ID)
	if !updated.IsLatest {
		t.Fatal("expected s2 to be latest")
	}

	s1ref, _ := repo.GetByID(context.Background(), s1.ID)
	if s1ref.IsLatest {
		t.Fatal("expected s1 to no longer be latest")
	}
}

func TestRegisterWebhook(t *testing.T) {
	wr := webhooks.NewInMemoryWebhookRepository()
	dr := webhooks.NewInMemoryDeliveryRepository()
	svc := webhooks.NewService(wr, dr)

	w, err := svc.Register(context.Background(), "order-webhook", "https://example.com/hook", "secret123", []string{"order.created", "order.updated"}, 3, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil webhook")
	}
	if w.Name != "order-webhook" {
		t.Errorf("expected 'order-webhook', got '%s'", w.Name)
	}
}

func TestUpdateWebhook(t *testing.T) {
	wr := webhooks.NewInMemoryWebhookRepository()
	dr := webhooks.NewInMemoryDeliveryRepository()
	svc := webhooks.NewService(wr, dr)

	w, _ := svc.Register(context.Background(), "hook", "https://example.com", "secret", []string{"event1"}, 3, 30)

	active := false
	updated, err := svc.Update(context.Background(), w.ID, "updated-hook", "", "", nil, &active, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil {
		t.Fatal("expected non-nil webhook")
	}
	if updated.Name != "updated-hook" {
		t.Errorf("expected 'updated-hook', got '%s'", updated.Name)
	}
	if updated.IsActive {
		t.Fatal("expected webhook to be inactive")
	}
}

func TestTriggerWebhookEvent(t *testing.T) {
	wr := webhooks.NewInMemoryWebhookRepository()
	dr := webhooks.NewInMemoryDeliveryRepository()
	svc := webhooks.NewService(wr, dr)

	svc.Register(context.Background(), "hook1", "https://example.com/1", "secret", []string{"order.created"}, 3, 30)
	svc.Register(context.Background(), "hook2", "https://example.com/2", "secret", []string{"order.created", "order.updated"}, 3, 30)

	deliveries, err := svc.TriggerEvent(context.Background(), "order.created", map[string]string{"order_id": "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deliveries) != 2 {
		t.Errorf("expected 2 deliveries, got %d", len(deliveries))
	}
}

func TestListDeliveries(t *testing.T) {
	wr := webhooks.NewInMemoryWebhookRepository()
	dr := webhooks.NewInMemoryDeliveryRepository()
	svc := webhooks.NewService(wr, dr)

	w, _ := svc.Register(context.Background(), "hook", "https://example.com", "secret", []string{"event1"}, 3, 30)
	svc.TriggerEvent(context.Background(), "event1", nil)

	deliveries, _ := svc.ListDeliveries(context.Background(), w.ID)
	if len(deliveries) == 0 {
		t.Error("expected at least 1 delivery")
	}
}

func TestCreatePipeline(t *testing.T) {
	repo := cicd.NewInMemoryRepository()
	svc := cicd.NewService(repo)

	p, err := svc.Create(context.Background(), "CI Pipeline", "my-service", cicd.TriggerPush, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}
	if p.Status != cicd.StatusPending {
		t.Errorf("expected pending status, got %s", p.Status)
	}
	if len(p.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(p.Stages))
	}
}

func TestTriggerPipeline(t *testing.T) {
	repo := cicd.NewInMemoryRepository()
	svc := cicd.NewService(repo)

	p, _ := svc.Create(context.Background(), "CI Pipeline", "my-service", cicd.TriggerPR, "def456")
	triggered, err := svc.Trigger(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if triggered.Status != cicd.StatusSuccess {
		t.Errorf("expected success status, got %s", triggered.Status)
	}
	if triggered.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
}

func TestGetPipelineStatus(t *testing.T) {
	repo := cicd.NewInMemoryRepository()
	svc := cicd.NewService(repo)

	p, _ := svc.Create(context.Background(), "CI Pipeline", "my-service", cicd.TriggerSchedule, "ghi789")
	svc.Trigger(context.Background(), p.ID)

	status, err := svc.GetStatus(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil pipeline")
	}
	if status.Status != cicd.StatusSuccess {
		t.Errorf("expected success, got %s", status.Status)
	}
}

func TestListPipelines(t *testing.T) {
	repo := cicd.NewInMemoryRepository()
	svc := cicd.NewService(repo)

	svc.Create(context.Background(), "Pipeline 1", "svc1", cicd.TriggerPush, "abc")
	svc.Create(context.Background(), "Pipeline 2", "svc2", cicd.TriggerPR, "def")

	pipelines, _ := svc.List(context.Background())
	if len(pipelines) != 2 {
		t.Errorf("expected 2 pipelines, got %d", len(pipelines))
	}
}

func TestGetOnboardingTemplate(t *testing.T) {
	repo := onboarding.NewInMemoryRepository()
	svc := onboarding.NewService(repo)

	tmpl, err := svc.GetTemplate(context.Background(), "microservice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected non-nil template")
	}
	if len(tmpl.Tasks) != 5 {
		t.Errorf("expected 5 tasks, got %d", len(tmpl.Tasks))
	}
}

func TestCompleteTask(t *testing.T) {
	repo := onboarding.NewInMemoryRepository()
	svc := onboarding.NewService(repo)

	err := svc.CompleteTask(context.Background(), "ms-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	progress, _ := svc.GetProgress(context.Background())
	if progress.CompletedTasks != 1 {
		t.Errorf("expected 1 completed task, got %d", progress.CompletedTasks)
	}
}

func TestGetProgress(t *testing.T) {
	repo := onboarding.NewInMemoryRepository()
	svc := onboarding.NewService(repo)

	progress, err := svc.GetProgress(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progress.TotalTasks != 9 {
		t.Errorf("expected 9 total tasks, got %d", progress.TotalTasks)
	}
	if progress.CompletedTasks != 0 {
		t.Errorf("expected 0 completed, got %d", progress.CompletedTasks)
	}
	if progress.Percentage != 0 {
		t.Errorf("expected 0%% progress, got %f%%", progress.Percentage)
	}
}

func TestListTemplates(t *testing.T) {
	repo := onboarding.NewInMemoryRepository()
	svc := onboarding.NewService(repo)

	templates, err := svc.ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestPipelineStages(t *testing.T) {
	repo := cicd.NewInMemoryRepository()
	svc := cicd.NewService(repo)

	p, _ := svc.Create(context.Background(), "Test Stages", "svc", cicd.TriggerPush, "abc")
	if p.Stages[0].Name != "build" {
		t.Errorf("expected first stage 'build', got '%s'", p.Stages[0].Name)
	}
	if p.Stages[1].Name != "test" {
		t.Errorf("expected second stage 'test', got '%s'", p.Stages[1].Name)
	}
	if p.Stages[2].Name != "deploy" {
		t.Errorf("expected third stage 'deploy', got '%s'", p.Stages[2].Name)
	}
}

func TestOnboardingMultipleTasks(t *testing.T) {
	repo := onboarding.NewInMemoryRepository()
	svc := onboarding.NewService(repo)

	svc.CompleteTask(context.Background(), "ms-1")
	svc.CompleteTask(context.Background(), "ms-2")
	svc.CompleteTask(context.Background(), "ms-3")

	progress, _ := svc.GetProgress(context.Background())
	if progress.CompletedTasks != 3 {
		t.Errorf("expected 3 completed tasks, got %d", progress.CompletedTasks)
	}
	expectedPct := float64(3) / float64(9) * 100
	if progress.Percentage != expectedPct {
		t.Errorf("expected %.2f%% progress, got %f%%", expectedPct, progress.Percentage)
	}
}

func TestWebhookDeliveryStatus(t *testing.T) {
	wr := webhooks.NewInMemoryWebhookRepository()
	dr := webhooks.NewInMemoryDeliveryRepository()
	svc := webhooks.NewService(wr, dr)

	svc.Register(context.Background(), "hook", "https://example.com", "secret", []string{"test.event"}, 3, 30)
	deliveries, _ := svc.TriggerEvent(context.Background(), "test.event", nil)

	if len(deliveries) == 0 {
		t.Fatal("expected at least one delivery")
	}
	for _, d := range deliveries {
		if d.Status == webhooks.DeliveryPending {
			t.Errorf("expected delivery to be delivered or failed, got pending")
		}
		if d.ID == "" {
			t.Error("expected delivery to have an ID")
		}
	}
}

func TestSDKCRUD(t *testing.T) {
	repo := sdk.NewInMemoryRepository()
	svc := sdk.NewService(repo)

	s, _ := svc.Register(context.Background(), "Test SDK", "python", "3.0.0", "https://repo", "https://docs", "python 3.9+")

	stored, err := repo.GetByID(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stored == nil {
		t.Fatal("expected stored SDK")
	}
	if stored.Version != "3.0.0" {
		t.Errorf("expected version 3.0.0, got %s", stored.Version)
	}
}
