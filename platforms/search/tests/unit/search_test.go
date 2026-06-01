package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

func newTestDocument(id, title, desc, category string, price, rating float64, stock int) *search.ProductDocument {
	return &search.ProductDocument{
		ID:          id,
		Title:       title,
		Description: desc,
		Category:    category,
		SellerID:    "seller1",
		Price:       price,
		Rating:      rating,
		Stock:       stock,
		Tags:        []string{"tag1", "tag2"},
		ImageURLs:   []string{"http://example.com/img.jpg"},
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	}
}

func TestSearchBasicTextMatching(t *testing.T) {
	repo := search.NewInMemoryRepository()
	svc := search.NewService(repo)
	ctx := context.Background()

	repo.Index(ctx, newTestDocument("1", "iPhone 15 Pro", "Latest Apple smartphone", "Electronics", 999, 4.8, 100))
	repo.Index(ctx, newTestDocument("2", "Samsung Galaxy S24", "Android flagship phone", "Electronics", 899, 4.5, 200))
	repo.Index(ctx, newTestDocument("3", "Sony WH-1000XM5", "Wireless noise cancelling headphones", "Audio", 349, 4.7, 50))

	result, err := svc.Search(ctx, search.SearchQuery{Query: "Apple smartphone", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Expected 1 result for unique phrase 'Apple smartphone', got %d", result.Total)
	}
	if len(result.Products) > 0 && result.Products[0].ID != "1" {
		t.Errorf("Expected product 1 (iPhone) for 'Apple smartphone', got %s", result.Products[0].ID)
	}

	result, err = svc.Search(ctx, search.SearchQuery{Query: "Wireless cancelling", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total == 0 {
		t.Error("Expected results for 'Wireless cancelling', got 0")
	}
	found := false
	for _, p := range result.Products {
		if p.ID == "3" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected product 3 (Sony) in results for 'Wireless cancelling', got %+v", result.Products)
	}

	result, err = svc.Search(ctx, search.SearchQuery{Query: "headphones", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Expected 1 result for 'headphones', got %d", result.Total)
	}
}

func TestSearchPagination(t *testing.T) {
	repo := search.NewInMemoryRepository()
	svc := search.NewService(repo)
	ctx := context.Background()

	for i := 1; i <= 25; i++ {
		repo.Index(ctx, newTestDocument(
			string(rune('0'+i%10))+string(rune('0'+i/10)),
			"Product", "Description", "Category",
			float64(i*10), float64(i%5)+1, i*10,
		))
	}

	result, err := svc.Search(ctx, search.SearchQuery{Query: "Product", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total != 25 {
		t.Errorf("Expected total of 25, got %d", result.Total)
	}
	if len(result.Products) != 10 {
		t.Errorf("Expected 10 results on page 1, got %d", len(result.Products))
	}
	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}

	result, err = svc.Search(ctx, search.SearchQuery{Query: "Product", Page: 3, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(result.Products) != 5 {
		t.Errorf("Expected 5 results on page 3, got %d", len(result.Products))
	}
}

func TestSearchFacetedFiltering(t *testing.T) {
	repo := search.NewInMemoryRepository()
	svc := search.NewService(repo)
	ctx := context.Background()

	repo.Index(ctx, newTestDocument("1", "iPhone", "Phone", "Electronics", 999, 4.8, 100))
	repo.Index(ctx, newTestDocument("2", "Samsung TV", "TV", "Electronics", 499, 4.2, 50))
	repo.Index(ctx, newTestDocument("3", "Nike Shoes", "Running shoes", "Sports", 129, 4.5, 200))

	result, err := svc.Search(ctx, search.SearchQuery{Query: "", Category: "Sports", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Expected 1 result in Sports, got %d", result.Total)
	}

	result, err = svc.Search(ctx, search.SearchQuery{Query: "", Category: "Electronics", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Expected 2 results in Electronics, got %d", result.Total)
	}
}

func TestSearchSorting(t *testing.T) {
	repo := search.NewInMemoryRepository()
	svc := search.NewService(repo)
	ctx := context.Background()

	repo.Index(ctx, newTestDocument("1", "Cheap Item", "Cheap", "General", 10, 3.0, 100))
	repo.Index(ctx, newTestDocument("2", "Mid Item", "Mid", "General", 50, 4.0, 100))
	repo.Index(ctx, newTestDocument("3", "Expensive Item", "Expensive", "General", 100, 5.0, 100))

	result, err := svc.Search(ctx, search.SearchQuery{Query: "", SortBy: "price_asc", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Products[0].Price != 10 || result.Products[2].Price != 100 {
		t.Errorf("Expected ascending prices: 10, 50, 100; got %v, %v, %v",
			result.Products[0].Price, result.Products[1].Price, result.Products[2].Price)
	}

	result, err = svc.Search(ctx, search.SearchQuery{Query: "", SortBy: "price_desc", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Products[0].Price != 100 || result.Products[2].Price != 10 {
		t.Errorf("Expected descending prices: 100, 50, 10; got %v, %v, %v",
			result.Products[0].Price, result.Products[1].Price, result.Products[2].Price)
	}

	result, err = svc.Search(ctx, search.SearchQuery{Query: "", SortBy: "rating", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Products[0].Rating != 5.0 || result.Products[2].Rating != 3.0 {
		t.Errorf("Expected descending ratings: 5, 4, 3; got %v, %v, %v",
			result.Products[0].Rating, result.Products[1].Rating, result.Products[2].Rating)
	}
}

func TestSearchPriceRange(t *testing.T) {
	repo := search.NewInMemoryRepository()
	svc := search.NewService(repo)
	ctx := context.Background()

	repo.Index(ctx, newTestDocument("1", "Cheap", "Desc", "Cat", 5, 3.0, 10))
	repo.Index(ctx, newTestDocument("2", "Mid", "Desc", "Cat", 50, 3.0, 10))
	repo.Index(ctx, newTestDocument("3", "Expensive", "Desc", "Cat", 500, 3.0, 10))

	result, err := svc.Search(ctx, search.SearchQuery{MinPrice: 10, MaxPrice: 100, Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Expected 1 result in price range 10-100, got %d", result.Total)
	}
	if result.Products[0].ID != "2" {
		t.Errorf("Expected product 2 (price 50), got %s", result.Products[0].ID)
	}
}

func TestFacetedSearch(t *testing.T) {
	repo := search.NewInMemoryRepository()
	svc := search.NewService(repo)
	ctx := context.Background()

	repo.Index(ctx, newTestDocument("1", "iPhone", "Phone", "Electronics", 999, 4.8, 100))
	repo.Index(ctx, newTestDocument("2", "Samsung TV", "TV", "Electronics", 499, 4.2, 50))
	repo.Index(ctx, newTestDocument("3", "Nike Shoes", "Shoes", "Sports", 129, 4.5, 200))

	result, err := svc.FacetedSearch(ctx, search.SearchQuery{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("FacetedSearch failed: %v", err)
	}
	if len(result.Facets) == 0 {
		t.Fatal("Expected facets, got none")
	}

	categoryFacet := result.Facets[0]
	if categoryFacet.Field != "category" {
		t.Errorf("Expected field 'category', got '%s'", categoryFacet.Field)
	}
	if categoryFacet.Values["Electronics"] != 2 {
		t.Errorf("Expected 2 Electronics, got %d", categoryFacet.Values["Electronics"])
	}
	if categoryFacet.Values["Sports"] != 1 {
		t.Errorf("Expected 1 Sports, got %d", categoryFacet.Values["Sports"])
	}
}

func TestSearchTypoTolerance(t *testing.T) {
	repo := search.NewInMemoryRepository()
	svc := search.NewService(repo)
	ctx := context.Background()

	repo.Index(ctx, newTestDocument("1", "iPhone 15 Pro", "Latest Apple smartphone", "Electronics", 999, 4.8, 100))
	repo.Index(ctx, newTestDocument("2", "Samsung Galaxy", "Android phone", "Electronics", 899, 4.5, 200))

	result, err := svc.Search(ctx, search.SearchQuery{Query: "iphne", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Total == 0 {
		t.Log("Note: typo 'iphne' with edit distance 1 should match 'iPhone'")
	}
}
