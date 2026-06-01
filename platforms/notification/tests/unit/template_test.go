package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/template"
)

func TestCreateTemplate(t *testing.T) {
	repo := template.NewInMemoryRepository()
	svc := template.NewService(repo)
	ctx := context.Background()

	tmpl, err := svc.CreateTemplate(ctx, &template.CreateTemplateRequest{
		Name:      "welcome_email",
		Subject:   "Welcome {{.UserName}}!",
		Body:      "<h1>Hello {{.UserName}}</h1><p>Your order {{.OrderID}} is confirmed</p>",
		Variables: []string{"UserName", "OrderID"},
	})
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}
	if tmpl.ID == "" {
		t.Error("expected template ID")
	}
	if tmpl.Version != 1 {
		t.Errorf("expected version 1, got %d", tmpl.Version)
	}
}

func TestRenderTemplate(t *testing.T) {
	repo := template.NewInMemoryRepository()
	svc := template.NewService(repo)
	ctx := context.Background()

	tmpl, _ := svc.CreateTemplate(ctx, &template.CreateTemplateRequest{
		Name:    "order_confirmation",
		Subject: "Order {{.OrderID}} Confirmed",
		Body:    "Dear {{.UserName}}, your order of ${{.Amount}} is confirmed!",
		Variables: []string{"UserName", "OrderID", "Amount"},
	})

	subject, body, err := svc.RenderTemplate(ctx, tmpl.ID, map[string]interface{}{
		"UserName": "John",
		"OrderID":  "ORD-123",
		"Amount":   "99.99",
	})
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}
	if subject != "Order ORD-123 Confirmed" {
		t.Errorf("expected subject 'Order ORD-123 Confirmed', got '%s'", subject)
	}
	if body != "Dear John, your order of $99.99 is confirmed!" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestListTemplates(t *testing.T) {
	repo := template.NewInMemoryRepository()
	svc := template.NewService(repo)
	ctx := context.Background()

	svc.CreateTemplate(ctx, &template.CreateTemplateRequest{Name: "t1", Subject: "S1", Body: "B1"})
	svc.CreateTemplate(ctx, &template.CreateTemplateRequest{Name: "t2", Subject: "S2", Body: "B2"})

	templates, err := svc.ListTemplates(ctx)
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestUpdateTemplate(t *testing.T) {
	repo := template.NewInMemoryRepository()
	svc := template.NewService(repo)
	ctx := context.Background()

	tmpl, _ := svc.CreateTemplate(ctx, &template.CreateTemplateRequest{
		Name:    "test",
		Subject: "Old Subject",
		Body:    "Old Body",
	})

	newSubject := "New Subject"
	updated, err := svc.UpdateTemplate(ctx, tmpl.ID, &template.UpdateTemplateRequest{
		Subject: &newSubject,
	})
	if err != nil {
		t.Fatalf("UpdateTemplate failed: %v", err)
	}
	if updated.Subject != "New Subject" {
		t.Errorf("expected 'New Subject', got '%s'", updated.Subject)
	}
	if updated.Version != 2 {
		t.Errorf("expected version 2 after update, got %d", updated.Version)
	}
}

func TestTemplateVersions(t *testing.T) {
	repo := template.NewInMemoryRepository()
	svc := template.NewService(repo)
	ctx := context.Background()

	tmpl, _ := svc.CreateTemplate(ctx, &template.CreateTemplateRequest{
		Name:    "versioned",
		Subject: "V1",
		Body:    "Version 1",
	})

	s2 := "V2"
	svc.UpdateTemplate(ctx, tmpl.ID, &template.UpdateTemplateRequest{Subject: &s2})

	versions, err := svc.ListVersions(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}
