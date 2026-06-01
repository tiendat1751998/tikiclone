package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/search-indexing/internal/pipeline"
)

func TestCreatePipeline(t *testing.T) {
	repo := pipeline.NewInMemoryRepository()
	svc := pipeline.NewService(repo)
	ctx := context.Background()

	p, err := svc.CreatePipeline(ctx, "search-pipeline", "products", []pipeline.PipelineStage{
		{Name: "tokenize", Processor: pipeline.ProcessorTokenize},
		{Name: "analyze", Processor: pipeline.ProcessorAnalyze},
	})
	if err != nil {
		t.Fatalf("CreatePipeline failed: %v", err)
	}
	if p.ID == "" {
		t.Error("expected pipeline ID to be set")
	}
	if !p.IsActive {
		t.Error("expected pipeline to be active")
	}
	if len(p.Stages) != 2 {
		t.Errorf("expected 2 stages, got %d", len(p.Stages))
	}
}

func TestCreatePipelineInvalidProcessor(t *testing.T) {
	repo := pipeline.NewInMemoryRepository()
	svc := pipeline.NewService(repo)
	ctx := context.Background()

	_, err := svc.CreatePipeline(ctx, "bad-pipeline", "products", []pipeline.PipelineStage{
		{Name: "unknown", Processor: "invalid"},
	})
	if err != pipeline.ErrUnknownProcessor {
		t.Errorf("expected ErrUnknownProcessor, got %v", err)
	}
}

func TestProcessDocumentSingleStage(t *testing.T) {
	repo := pipeline.NewInMemoryRepository()
	svc := pipeline.NewService(repo)
	ctx := context.Background()

	p, _ := svc.CreatePipeline(ctx, "tokenizer", "products", []pipeline.PipelineStage{
		{Name: "tokenize", Processor: pipeline.ProcessorTokenize},
	})

	doc := &pipeline.Document{
		ID: "doc-1",
		Fields: map[string]interface{}{
			"content": "Hello World Test",
		},
	}

	result, err := svc.ProcessDocument(ctx, p.ID, doc)
	if err != nil {
		t.Fatalf("ProcessDocument failed: %v", err)
	}

	tokens, ok := result.Fields["content_tokens"].([]string)
	if !ok {
		t.Fatal("expected content_tokens to be []string")
	}
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0] != "hello" || tokens[1] != "world" || tokens[2] != "test" {
		t.Errorf("unexpected tokens: %v", tokens)
	}

	if result.Fields["_index"] != "products" {
		t.Errorf("expected _index products, got %v", result.Fields["_index"])
	}
}

func TestProcessDocumentMultipleStages(t *testing.T) {
	repo := pipeline.NewInMemoryRepository()
	svc := pipeline.NewService(repo)
	ctx := context.Background()

	p, _ := svc.CreatePipeline(ctx, "full-pipeline", "products", []pipeline.PipelineStage{
		{Name: "tokenize", Processor: pipeline.ProcessorTokenize},
		{Name: "analyze", Processor: pipeline.ProcessorAnalyze},
		{Name: "transform", Processor: pipeline.ProcessorTransform},
		{Name: "enrich", Processor: pipeline.ProcessorEnrich, Config: map[string]interface{}{"source": "internal"}},
	})

	doc := &pipeline.Document{
		ID: "doc-2",
		Fields: map[string]interface{}{
			"content": "The quick brown fox jumps over the lazy dog",
		},
	}

	result, err := svc.ProcessDocument(ctx, p.ID, doc)
	if err != nil {
		t.Fatalf("ProcessDocument failed: %v", err)
	}

	if result.Fields["_index"] != "products" {
		t.Errorf("expected _index products, got %v", result.Fields["_index"])
	}

	analyzed, ok := result.Fields["analyzed_tokens"].([]string)
	if !ok {
		t.Fatal("expected analyzed_tokens to be []string")
	}

	stopwords := map[string]bool{"the": true, "a": true, "an": true, "is": true}
	for _, tok := range analyzed {
		if stopwords[tok] {
			t.Errorf("stopword '%s' should have been filtered", tok)
		}
	}

	if result.Fields["enriched_source"] != "internal" {
		t.Errorf("expected enriched_source internal, got %v", result.Fields["enriched_source"])
	}
}

func TestProcessDocumentNotFound(t *testing.T) {
	repo := pipeline.NewInMemoryRepository()
	svc := pipeline.NewService(repo)
	ctx := context.Background()

	_, err := svc.ProcessDocument(ctx, "non-existent", &pipeline.Document{
		ID:     "doc-1",
		Fields: map[string]interface{}{"content": "test"},
	})
	if err != pipeline.ErrPipelineNotFound {
		t.Errorf("expected ErrPipelineNotFound, got %v", err)
	}
}

func TestGetPipeline(t *testing.T) {
	repo := pipeline.NewInMemoryRepository()
	svc := pipeline.NewService(repo)
	ctx := context.Background()

	created, _ := svc.CreatePipeline(ctx, "test-pipeline", "test-index", []pipeline.PipelineStage{
		{Name: "tokenize", Processor: pipeline.ProcessorTokenize},
	})

	retrieved, err := svc.GetPipeline(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetPipeline failed: %v", err)
	}
	if retrieved.Name != "test-pipeline" {
		t.Errorf("expected name test-pipeline, got %s", retrieved.Name)
	}
}

func TestListPipelines(t *testing.T) {
	repo := pipeline.NewInMemoryRepository()
	svc := pipeline.NewService(repo)
	ctx := context.Background()

	svc.CreatePipeline(ctx, "p1", "i1", []pipeline.PipelineStage{{Name: "t", Processor: pipeline.ProcessorTokenize}})
	svc.CreatePipeline(ctx, "p2", "i2", []pipeline.PipelineStage{{Name: "a", Processor: pipeline.ProcessorAnalyze}})

	pipelines, err := svc.ListPipelines(ctx)
	if err != nil {
		t.Fatalf("ListPipelines failed: %v", err)
	}
	if len(pipelines) != 2 {
		t.Errorf("expected 2 pipelines, got %d", len(pipelines))
	}
}
