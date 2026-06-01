package tests

import (
	"testing"

	"github.com/tikiclone/tiki/platforms/api-gateway/internal/transform"
)

func TestTransformSetHeader(t *testing.T) {
	repo := transform.NewInMemoryRepository()
	tf := transform.NewTransformer(repo)

	tf.CreateRule(&transform.Rule{
		ID: "rule-1",
		Actions: []transform.Action{
			{Type: "set_header", Name: "X-Custom", Value: "test-value"},
		},
	})

	resp, err := tf.Apply("rule-1", &transform.TransformRequest{
		Path:    "/api/test",
		Method:  "GET",
		Headers: map[string]string{},
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if resp.Headers["X-Custom"] != "test-value" {
		t.Errorf("expected X-Custom header, got %v", resp.Headers)
	}
}

func TestTransformRemoveHeader(t *testing.T) {
	repo := transform.NewInMemoryRepository()
	tf := transform.NewTransformer(repo)

	tf.CreateRule(&transform.Rule{
		ID: "rule-1",
		Actions: []transform.Action{
			{Type: "remove_header", Name: "Authorization"},
		},
	})

	resp, err := tf.Apply("rule-1", &transform.TransformRequest{
		Path:   "/api/test",
		Method: "GET",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"X-Other":      "value",
		},
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if _, exists := resp.Headers["Authorization"]; exists {
		t.Error("Authorization header should have been removed")
	}
	if resp.Headers["X-Other"] != "value" {
		t.Error("other headers should remain")
	}
}

func TestTransformSetQuery(t *testing.T) {
	repo := transform.NewInMemoryRepository()
	tf := transform.NewTransformer(repo)

	tf.CreateRule(&transform.Rule{
		ID: "rule-1",
		Actions: []transform.Action{
			{Type: "set_query", Name: "locale", Value: "en-US"},
		},
	})

	resp, err := tf.Apply("rule-1", &transform.TransformRequest{
		Path:   "/api/test",
		Method: "GET",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if resp.Query["locale"] != "en-US" {
		t.Errorf("expected locale=en-US, got %v", resp.Query)
	}
}

func TestTransformRewritePath(t *testing.T) {
	repo := transform.NewInMemoryRepository()
	tf := transform.NewTransformer(repo)

	tf.CreateRule(&transform.Rule{
		ID: "rule-1",
		Actions: []transform.Action{
			{Type: "rewrite_path", Value: "/v2$path"},
		},
	})

	resp, err := tf.Apply("rule-1", &transform.TransformRequest{
		Path:   "/api/v1/users",
		Method: "GET",
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if resp.Path != "/v2/api/v1/users" {
		t.Errorf("expected /v2/api/v1/users, got %s", resp.Path)
	}
}

func TestTransformBody(t *testing.T) {
	repo := transform.NewInMemoryRepository()
	tf := transform.NewTransformer(repo)

	tf.CreateRule(&transform.Rule{
		ID: "rule-1",
		Actions: []transform.Action{
			{Type: "transform_body", Value: `{"transformed": true}`},
		},
	})

	resp, err := tf.Apply("rule-1", &transform.TransformRequest{
		Path:   "/api/test",
		Method: "POST",
		Body:   `{"original": true}`,
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if resp.Body != `{"transformed": true}` {
		t.Errorf("expected transformed body, got %s", resp.Body)
	}
}

func TestTransformMultipleActions(t *testing.T) {
	repo := transform.NewInMemoryRepository()
	tf := transform.NewTransformer(repo)

	tf.CreateRule(&transform.Rule{
		ID: "rule-1",
		Actions: []transform.Action{
			{Type: "set_header", Name: "X-First", Value: "1"},
			{Type: "set_header", Name: "X-Second", Value: "2"},
			{Type: "remove_header", Name: "X-Remove"},
		},
	})

	resp, err := tf.Apply("rule-1", &transform.TransformRequest{
		Path:   "/api/test",
		Method: "GET",
		Headers: map[string]string{
			"X-Remove": "bye",
		},
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if resp.Headers["X-First"] != "1" || resp.Headers["X-Second"] != "2" {
		t.Error("expected both headers to be set")
	}
	if _, exists := resp.Headers["X-Remove"]; exists {
		t.Error("X-Remove should have been removed")
	}
}

func TestTransformRuleNotFound(t *testing.T) {
	tf := transform.NewTransformer(transform.NewInMemoryRepository())

	_, err := tf.Apply("nonexistent", &transform.TransformRequest{
		Path: "/test", Method: "GET",
	})
	if err == nil {
		t.Error("expected error for unknown rule")
	}
}

func TestResponseTransformEnvelope(t *testing.T) {
	rule := transform.NewDefaultResponseTransform()
	rule.WrapInEnvelope = true

	result := transform.ApplyResponseTransform(rule, "hello")
	envelope, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}
	if envelope["data"] != "hello" {
		t.Errorf("expected data=hello, got %v", envelope["data"])
	}
}

func TestResponseTransformCORS(t *testing.T) {
	rule := transform.NewDefaultResponseTransform()

	result := transform.ApplyResponseTransform(rule, "test")
	if result != "test" {
		t.Error("default transform should pass through")
	}
}

func TestComposerMultipleTransforms(t *testing.T) {
	repo := transform.NewInMemoryRepository()
	tf := transform.NewTransformer(repo)

	tf.CreateRule(&transform.Rule{
		ID: "rule-1",
		Actions: []transform.Action{
			{Type: "set_header", Name: "X-Stage", Value: "1"},
		},
	})
	tf.CreateRule(&transform.Rule{
		ID: "rule-2",
		Actions: []transform.Action{
			{Type: "set_header", Name: "X-Stage", Value: "2"},
		},
	})

	composer := transform.NewComposer([]*transform.Transformer{tf})

	resp, err := composer.Apply([]string{"rule-1", "rule-2"}, &transform.TransformRequest{
		Path:    "/test",
		Method:  "GET",
		Headers: map[string]string{},
	})
	if err != nil {
		t.Fatalf("compose failed: %v", err)
	}
	if resp.Headers["X-Stage"] != "2" {
		t.Errorf("expected final value 2, got %s", resp.Headers["X-Stage"])
	}
}
