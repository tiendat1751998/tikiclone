package tests

import (
	"testing"

	"github.com/tikiclone/tiki/platforms/api-gateway/internal/routes"
)

func TestRouteExactMatch(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/api/v1/users",
		Methods:     []string{"GET"},
		ServiceName: "user-service",
		TargetURL:   "http://users:8080",
	})

	route, err := svc.Match("/api/v1/users", "GET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route == nil {
		t.Fatal("expected route match")
	}
	if route.ServiceName != "user-service" {
		t.Errorf("expected user-service, got %s", route.ServiceName)
	}
}

func TestRoutePrefixMatch(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/api/v1/*",
		Methods:     []string{"GET"},
		ServiceName: "api-service",
		TargetURL:   "http://api:8080",
	})

	route, err := svc.Match("/api/v1/products/123", "GET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route == nil {
		t.Fatal("expected route match")
	}
}

func TestRouteNoMatch(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/api/v1/users",
		Methods:     []string{"GET"},
		ServiceName: "user-service",
		TargetURL:   "http://users:8080",
	})

	route, err := svc.Match("/api/v1/orders", "GET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route != nil {
		t.Fatal("expected no match")
	}
}

func TestRouteMethodMismatch(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/api/v1/users",
		Methods:     []string{"POST"},
		ServiceName: "user-service",
		TargetURL:   "http://users:8080",
	})

	route, err := svc.Match("/api/v1/users", "GET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route != nil {
		t.Fatal("expected no match for different method")
	}
}

func TestRouteLongestPrefixMatch(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/api/*",
		Methods:     []string{"GET"},
		ServiceName: "api-service",
		TargetURL:   "http://api:8080",
	})
	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/api/v1/*",
		Methods:     []string{"GET"},
		ServiceName: "v1-service",
		TargetURL:   "http://v1:8080",
	})

	route, err := svc.Match("/api/v1/products", "GET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route == nil {
		t.Fatal("expected a match")
	}
	if route.ServiceName != "v1-service" {
		t.Errorf("expected v1-service (longest prefix), got %s", route.ServiceName)
	}
}

func TestRouteRegisterAndDeregister(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	route, err := svc.Register(&routes.RegisterRouteRequest{
		Path:        "/test",
		Methods:     []string{"GET"},
		ServiceName: "test-svc",
		TargetURL:   "http://test:8080",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = svc.Deregister(route.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r, _ := svc.Get(route.ID)
	if r != nil {
		t.Fatal("route should be deleted")
	}
}

func TestRouteList(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/a", Methods: []string{"GET"}, ServiceName: "s1", TargetURL: "http://a:8080",
	})
	svc.Register(&routes.RegisterRouteRequest{
		Path:        "/b", Methods: []string{"POST"}, ServiceName: "s2", TargetURL: "http://b:8080",
	})

	list, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 routes, got %d", len(list))
	}
}

func TestRouteInactiveNotMatched(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	route, _ := svc.Register(&routes.RegisterRouteRequest{
		Path: "/api/v1/old", Methods: []string{"GET"}, ServiceName: "old", TargetURL: "http://old:8080",
	})
	r, _ := svc.Get(route.ID)
	r.IsActive = false
	repo.Store(r)

	matched, _ := svc.Match("/api/v1/old", "GET")
	if matched != nil {
		t.Fatal("inactive route should not match")
	}
}

func TestRouteRegisterValidation(t *testing.T) {
	svc := routes.NewService(routes.NewInMemoryRepository())

	_, err := svc.Register(&routes.RegisterRouteRequest{
		Path: "", Methods: []string{"GET"}, ServiceName: "s", TargetURL: "http://t",
	})
	if err == nil {
		t.Error("expected validation error for empty path")
	}
}

func TestRouteDefaultValues(t *testing.T) {
	repo := routes.NewInMemoryRepository()
	svc := routes.NewService(repo)

	route, err := svc.Register(&routes.RegisterRouteRequest{
		Path: "/test", Methods: []string{"GET"}, ServiceName: "s", TargetURL: "http://t",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route.TimeoutMs != 30000 {
		t.Errorf("expected default 30000ms timeout, got %d", route.TimeoutMs)
	}
	if !route.IsActive {
		t.Error("route should be active by default")
	}
	if route.Version != "1.0.0" {
		t.Errorf("expected default version 1.0.0, got %s", route.Version)
	}
}
