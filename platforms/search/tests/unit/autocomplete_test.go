package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/search/internal/autocomplete"
)

func TestAutocompletePrefixMatching(t *testing.T) {
	repo := autocomplete.NewInMemoryRepository()
	svc := autocomplete.NewService(repo)
	ctx := context.Background()

	suggestions := []string{"iphone 15", "iphone 14", "ipad air", "macbook pro", "iphone se"}
	for i, s := range suggestions {
		repo.Store(ctx, s, autocomplete.Suggestion{
			Text:  s,
			Score: float64(len(suggestions) - i),
			Type:  "product",
		})
	}

	result, err := svc.Suggest(ctx, "iphone", 10)
	if err != nil {
		t.Fatalf("Suggest failed: %v", err)
	}
	if len(result.Suggestions) == 0 {
		t.Fatal("Expected suggestions, got none")
	}

	hasIPhone := false
	for _, s := range result.Suggestions {
		if s.Text == "iphone 15" || s.Text == "iphone 14" || s.Text == "iphone se" {
			hasIPhone = true
			break
		}
	}
	if !hasIPhone {
		t.Errorf("Expected iPhone-related suggestions, got %+v", result.Suggestions)
	}

	result, err = svc.Suggest(ctx, "ipad", 10)
	if err != nil {
		t.Fatalf("Suggest failed: %v", err)
	}
	if len(result.Suggestions) == 0 {
		t.Fatal("Expected suggestions for 'ipad', got none")
	}
	if result.Suggestions[0].Text != "ipad air" {
		t.Errorf("Expected 'ipad air', got '%s'", result.Suggestions[0].Text)
	}
}

func TestAutocompleteNoMatch(t *testing.T) {
	repo := autocomplete.NewInMemoryRepository()
	svc := autocomplete.NewService(repo)
	ctx := context.Background()

	repo.Store(ctx, "iphone", autocomplete.Suggestion{Text: "iphone", Score: 1.0, Type: "product"})

	result, err := svc.Suggest(ctx, "zzzzz", 10)
	if err != nil {
		t.Fatalf("Suggest failed: %v", err)
	}
	if len(result.Suggestions) != 0 {
		t.Errorf("Expected 0 suggestions for non-matching prefix, got %d", len(result.Suggestions))
	}
}

func TestAutocompleteEmptyPrefix(t *testing.T) {
	repo := autocomplete.NewInMemoryRepository()
	svc := autocomplete.NewService(repo)
	ctx := context.Background()

	_, err := svc.Suggest(ctx, "", 10)
	if err == nil {
		t.Error("Expected error for empty prefix, got nil")
	}
}

func TestAutocompleteLimit(t *testing.T) {
	repo := autocomplete.NewInMemoryRepository()
	svc := autocomplete.NewService(repo)
	ctx := context.Background()

	for i := 0; i < 20; i++ {
		text := "product"
		repo.Store(ctx, text+string(rune('a'+i)), autocomplete.Suggestion{
			Text: text + string(rune('a'+i)), Score: float64(20 - i), Type: "product",
		})
	}

	result, err := svc.Suggest(ctx, "product", 5)
	if err != nil {
		t.Fatalf("Suggest failed: %v", err)
	}
	if len(result.Suggestions) > 5 {
		t.Errorf("Expected at most 5 suggestions, got %d", len(result.Suggestions))
	}
}

func TestTrendingQueries(t *testing.T) {
	repo := autocomplete.NewInMemoryRepository()
	svc := autocomplete.NewService(repo)
	ctx := context.Background()

	repo.StoreTrending(ctx, "iphone 15", 100)
	repo.StoreTrending(ctx, "samsung tv", 80)
	repo.StoreTrending(ctx, "nike shoes", 60)

	trending, err := svc.GetTrending(ctx, 10)
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}
	if len(trending) != 3 {
		t.Errorf("Expected 3 trending queries, got %d", len(trending))
	}
	if trending[0].Query != "iphone 15" || trending[0].Score != 100 {
		t.Errorf("Expected 'iphone 15' with score 100 as top trending, got %+v", trending[0])
	}
}

func TestRecordSearch(t *testing.T) {
	repo := autocomplete.NewInMemoryRepository()
	svc := autocomplete.NewService(repo)
	ctx := context.Background()

	svc.RecordSearch(ctx, "wireless headphones")
	svc.RecordSearch(ctx, "wireless headphones")
	svc.RecordSearch(ctx, "wireless headphones")
	svc.RecordSearch(ctx, "gaming mouse")

	trending, err := svc.GetTrending(ctx, 10)
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}

	if len(trending) != 2 {
		t.Errorf("Expected 2 trending queries, got %d", len(trending))
	}

	if trending[0].Query != "wireless headphones" {
		t.Errorf("Expected 'wireless headphones' as top trending, got '%s'", trending[0].Query)
	}
}
