package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/search-indexing/internal/synonyms"
)

func TestCreateSynonymSet(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	set, err := svc.CreateSet(ctx, []string{"happy", "glad", "joyful"}, "en")
	if err != nil {
		t.Fatalf("CreateSet failed: %v", err)
	}
	if set.ID == "" {
		t.Error("expected set ID to be set")
	}
	if !set.IsActive {
		t.Error("expected set to be active")
	}
	if len(set.Words) != 3 {
		t.Errorf("expected 3 words, got %d", len(set.Words))
	}
}

func TestCreateEmptySynonymSet(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	_, err := svc.CreateSet(ctx, []string{}, "en")
	if err != synonyms.ErrEmptyWords {
		t.Errorf("expected ErrEmptyWords, got %v", err)
	}
}

func TestExpandQuerySingleWord(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	svc.CreateSet(ctx, []string{"happy", "glad", "joyful"}, "en")

	expanded, err := svc.ExpandQuery(ctx, "happy")
	if err != nil {
		t.Fatalf("ExpandQuery failed: %v", err)
	}

	hasHappy := false
	hasGlad := false
	hasJoyful := false
	for _, w := range expanded {
		switch w {
		case "happy":
			hasHappy = true
		case "glad":
			hasGlad = true
		case "joyful":
			hasJoyful = true
		}
	}
	if !hasHappy || !hasGlad || !hasJoyful {
		t.Errorf("expected all synonyms in expanded set, got %v", expanded)
	}
}

func TestExpandQueryMultipleWords(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	svc.CreateSet(ctx, []string{"happy", "glad"}, "en")
	svc.CreateSet(ctx, []string{"sad", "unhappy"}, "en")

	expanded, err := svc.ExpandQuery(ctx, "happy sad")
	if err != nil {
		t.Fatalf("ExpandQuery failed: %v", err)
	}

	expected := map[string]bool{"happy": true, "glad": true, "sad": true, "unhappy": true}
	for _, w := range expanded {
		delete(expected, w)
	}
	if len(expected) > 0 {
		t.Errorf("missing expected words: %v, got %v", expected, expanded)
	}
}

func TestExpandQueryNoSynonyms(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	expanded, err := svc.ExpandQuery(ctx, "unique_word")
	if err != nil {
		t.Fatalf("ExpandQuery failed: %v", err)
	}
	if len(expanded) != 1 || expanded[0] != "unique_word" {
		t.Errorf("expected only original word, got %v", expanded)
	}
}

func TestGetSynonyms(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	svc.CreateSet(ctx, []string{"fast", "quick", "swift"}, "en")

	syns, err := svc.GetSynonyms(ctx, "fast")
	if err != nil {
		t.Fatalf("GetSynonyms failed: %v", err)
	}

	hasQuick := false
	hasSwift := false
	for _, s := range syns {
		if s == "quick" {
			hasQuick = true
		}
		if s == "swift" {
			hasSwift = true
		}
	}
	if !hasQuick || !hasSwift {
		t.Errorf("expected quick and swift as synonyms, got %v", syns)
	}
}

func TestGetSynonymsNonexistent(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	syns, err := svc.GetSynonyms(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetSynonyms failed: %v", err)
	}
	if len(syns) != 0 {
		t.Errorf("expected empty slice, got %v", syns)
	}
}

func TestRemoveSynonymSet(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	set, _ := svc.CreateSet(ctx, []string{"big", "large", "huge"}, "en")
	if err := svc.RemoveSet(ctx, set.ID); err != nil {
		t.Fatalf("RemoveSet failed: %v", err)
	}

	_, err := svc.GetSet(ctx, set.ID)
	if err != synonyms.ErrSynonymSetNotFound {
		t.Errorf("expected ErrSynonymSetNotFound after remove, got %v", err)
	}
}

func TestListSynonymSets(t *testing.T) {
	repo := synonyms.NewInMemoryRepository()
	svc := synonyms.NewService(repo)
	ctx := context.Background()

	svc.CreateSet(ctx, []string{"a", "b"}, "en")
	svc.CreateSet(ctx, []string{"c", "d"}, "fr")

	sets, err := svc.ListSets(ctx)
	if err != nil {
		t.Fatalf("ListSets failed: %v", err)
	}
	if len(sets) != 2 {
		t.Errorf("expected 2 sets, got %d", len(sets))
	}
}
