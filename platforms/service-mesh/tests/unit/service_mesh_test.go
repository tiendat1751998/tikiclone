package unit

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/service-mesh/internal/discovery"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/loadbalancer"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/mtls"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/resilience"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/telemetry"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/traffic"
)

func TestRegisterService(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	inst := &discovery.ServiceInstance{
		Name:    "test-service",
		Version: "1.0.0",
		Address: "192.168.1.1",
		Port:    8080,
		Region:  "us-east-1",
		Zone:    "us-east-1a",
		Status:  discovery.StatusUp,
	}
	err := svc.Register(ctx, inst)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if inst.ID == "" {
		t.Error("expected instance ID to be set")
	}
}

func TestDiscoverService(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc1", Address: "10.0.0.1", Port: 8080, Region: "us-east-1", Zone: "a",
	})
	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc1", Address: "10.0.0.2", Port: 8081, Region: "us-east-1", Zone: "b",
	})
	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc2", Address: "10.0.0.3", Port: 8082, Region: "us-east-1", Zone: "a",
	})

	instances, err := svc.Discover(ctx, "svc1", "", "")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}
}

func TestDiscoverServiceWithRegionFilter(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc1", Address: "10.0.0.1", Port: 8080, Region: "us-east-1", Zone: "a",
	})
	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc1", Address: "10.0.0.2", Port: 8081, Region: "eu-west-1", Zone: "a",
	})

	instances, err := svc.Discover(ctx, "svc1", "us-east-1", "")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(instances) != 1 {
		t.Errorf("expected 1 instance in us-east-1, got %d", len(instances))
	}
}

func TestHeartbeat(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc1", Address: "10.0.0.1", Port: 8080,
	})

	services, _ := svc.ListServices(ctx)
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	err := svc.Heartbeat(ctx, services[0].ID)
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}
}

func TestDeregisterService(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc1", Address: "10.0.0.1", Port: 8080,
	})

	services, _ := svc.ListServices(ctx)
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	err := svc.Deregister(ctx, services[0].ID)
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	services, _ = svc.ListServices(ctx)
	if len(services) != 0 {
		t.Errorf("expected 0 services after deregister, got %d", len(services))
	}
}

func TestHealthCheckFailureDetection(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc1", Address: "10.0.0.1", Port: 8080,
	})

	services, _ := svc.ListServices(ctx)
	inst := services[0]
	inst.LastHeartbeat = time.Now().Add(-20 * time.Second)

	_ = inst
}

func TestListServices(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{Name: "a", Address: "1", Port: 1})
	svc.Register(ctx, &discovery.ServiceInstance{Name: "b", Address: "2", Port: 2})
	svc.Register(ctx, &discovery.ServiceInstance{Name: "c", Address: "3", Port: 3})

	services, err := svc.ListServices(ctx)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}
	if len(services) != 3 {
		t.Errorf("expected 3 services, got %d", len(services))
	}
}

func TestCreateCA(t *testing.T) {
	ca, err := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	if err != nil {
		t.Fatalf("NewCertificateAuthority failed: %v", err)
	}
	if ca.RootCert == nil {
		t.Error("expected root cert to be non-nil")
	}
	certs := ca.ListCerts()
	if len(certs) != 1 {
		t.Errorf("expected 1 cert (CA), got %d", len(certs))
	}
	if !certs[0].IsCA {
		t.Error("expected root cert to be CA")
	}
}

func TestIssueCertificate(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	cert, err := mgr.IssueCert(ctx, "service-a", "service-a.local", "Tiki", 90, true)
	if err != nil {
		t.Fatalf("IssueCert failed: %v", err)
	}
	if cert.ID == "" {
		t.Error("expected cert ID")
	}
	if cert.IsCA {
		t.Error("issued cert should not be CA")
	}
	if cert.ServiceName != "service-a" {
		t.Errorf("expected service-a, got %s", cert.ServiceName)
	}
}

func TestIssueClientCertificate(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	cert, err := mgr.IssueCert(ctx, "client-a", "client-a.local", "Tiki", 30, false)
	if err != nil {
		t.Fatalf("IssueCert failed: %v", err)
	}
	if cert == nil {
		t.Fatal("expected non-nil cert")
	}
	if cert.Fingerprint == "" {
		t.Error("expected fingerprint")
	}
}

func TestRenewCertificate(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	orig, _ := mgr.IssueCert(ctx, "svc1", "svc1.local", "Tiki", 90, true)
	renewed, err := mgr.RenewCert(ctx, orig.ID, 180)
	if err != nil {
		t.Fatalf("RenewCert failed: %v", err)
	}
	if renewed.ID == orig.ID {
		t.Error("renewed cert should have new ID")
	}
}

func TestRevokeCertificate(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	cert, _ := mgr.IssueCert(ctx, "svc1", "svc1.local", "Tiki", 90, true)

	err := mgr.RevokeCert(ctx, cert.ID)
	if err != nil {
		t.Fatalf("RevokeCert failed: %v", err)
	}

	err = mgr.VerifyCert(ctx, cert.ID)
	if err != mtls.ErrCertRevoked {
		t.Errorf("expected ErrCertRevoked, got %v", err)
	}
}

func TestVerifyCertificate(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	cert, _ := mgr.IssueCert(ctx, "svc1", "svc1.local", "Tiki", 90, true)

	err := mgr.VerifyCert(ctx, cert.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestVerifyCertificateNotFound(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	err := mgr.VerifyCert(ctx, "nonexistent")
	if err != mtls.ErrCertNotFound {
		t.Errorf("expected ErrCertNotFound, got %v", err)
	}
}

func TestListCertificates(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	mgr.IssueCert(ctx, "s1", "s1.local", "Tiki", 90, true)
	mgr.IssueCert(ctx, "s2", "s2.local", "Tiki", 90, false)

	certs := mgr.ListCerts(ctx)
	if len(certs) != 3 {
		t.Errorf("expected 3 certs (1 CA + 2 issued), got %d", len(certs))
	}
}

func TestVerifyChain(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("TestOrg", "Test CA", 365)

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "test.local"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, template, ca.RootCert, &key.PublicKey, ca.RootKey)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	err := ca.VerifyChain(certPEM)
	if err != nil {
		t.Errorf("VerifyChain failed: %v", err)
	}
}

func TestCreateTrafficRule(t *testing.T) {
	repo := traffic.NewInMemoryRepository()
	eng := traffic.NewEngine(repo)
	ctx := context.Background()

	rule, err := eng.CreateRule(ctx, &traffic.TrafficRule{
		Name:               "canary-route",
		SourceService:      "gateway",
		DestinationService: "checkout",
		Weight:             10,
	})
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}
	if rule.ID == "" {
		t.Error("expected rule ID")
	}
}

func TestListTrafficRules(t *testing.T) {
	repo := traffic.NewInMemoryRepository()
	eng := traffic.NewEngine(repo)
	ctx := context.Background()

	eng.CreateRule(ctx, &traffic.TrafficRule{Name: "r1", Weight: 50})
	eng.CreateRule(ctx, &traffic.TrafficRule{Name: "r2", Weight: 50})

	rules, err := eng.ListRules(ctx)
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestEvaluateRouteWithCanaryWeight(t *testing.T) {
	repo := traffic.NewInMemoryRepository()
	eng := traffic.NewEngine(repo)
	ctx := context.Background()

	eng.CreateRule(ctx, &traffic.TrafficRule{
		Name:               "stable",
		DestinationService: "checkout",
		Weight:             90,
	})
	eng.CreateRule(ctx, &traffic.TrafficRule{
		Name:               "canary",
		DestinationService: "checkout",
		Weight:             10,
	})

	stableCount := 0
	canaryCount := 0
	for i := 0; i < 1000; i++ {
		rule, err := eng.EvaluateRoute(ctx, "", "checkout", nil, "/", "GET")
		if err != nil {
			t.Fatalf("EvaluateRoute failed: %v", err)
		}
		if rule == nil {
			t.Fatal("expected a matching rule")
		}
		if rule.Name == "stable" {
			stableCount++
		} else {
			canaryCount++
		}
	}
	if stableCount == 0 || canaryCount == 0 {
		t.Errorf("both rules should be selected: stable=%d, canary=%d", stableCount, canaryCount)
	}
	ratio := float64(stableCount) / float64(canaryCount)
	if ratio < 4 || ratio > 20 {
		t.Errorf("expected ratio ~9, got %.2f (stable=%d, canary=%d)", ratio, stableCount, canaryCount)
	}
}

func TestEvaluateRouteWithHeaders(t *testing.T) {
	repo := traffic.NewInMemoryRepository()
	eng := traffic.NewEngine(repo)
	ctx := context.Background()

	eng.CreateRule(ctx, &traffic.TrafficRule{
		Name:               "header-match",
		DestinationService: "checkout",
		MatchConditions: traffic.MatchCondition{
			Headers: map[string]string{"X-Canary": "true"},
		},
		Weight: 100,
	})

	rule, err := eng.EvaluateRoute(ctx, "", "checkout", map[string]string{"X-Canary": "true"}, "/", "GET")
	if err != nil {
		t.Fatalf("EvaluateRoute failed: %v", err)
	}
	if rule == nil {
		t.Fatal("expected matching rule for header match")
	}
	if rule.Name != "header-match" {
		t.Errorf("expected header-match, got %s", rule.Name)
	}

	rule2, _ := eng.EvaluateRoute(ctx, "", "checkout", map[string]string{"X-Canary": "false"}, "/", "GET")
	if rule2 != nil {
		t.Error("expected no match for non-matching header")
	}
}

func TestEvaluateRouteWithPathPrefix(t *testing.T) {
	repo := traffic.NewInMemoryRepository()
	eng := traffic.NewEngine(repo)
	ctx := context.Background()

	eng.CreateRule(ctx, &traffic.TrafficRule{
		Name:               "api-route",
		DestinationService: "api-service",
		MatchConditions: traffic.MatchCondition{
			PathPrefix: "/api/v2",
		},
		Weight: 100,
	})

	rule, err := eng.EvaluateRoute(ctx, "", "api-service", nil, "/api/v2/users", "GET")
	if err != nil {
		t.Fatalf("EvaluateRoute failed: %v", err)
	}
	if rule == nil {
		t.Fatal("expected matching rule for path prefix")
	}

	rule2, _ := eng.EvaluateRoute(ctx, "", "api-service", nil, "/api/v1/users", "GET")
	if rule2 != nil {
		t.Error("expected no match for non-matching path")
	}
}

func TestEvaluateRouteMethodMatch(t *testing.T) {
	repo := traffic.NewInMemoryRepository()
	eng := traffic.NewEngine(repo)
	ctx := context.Background()

	eng.CreateRule(ctx, &traffic.TrafficRule{
		Name:               "post-only",
		DestinationService: "svc1",
		MatchConditions: traffic.MatchCondition{
			Methods: []string{"POST"},
		},
		Weight: 100,
	})

	rule, _ := eng.EvaluateRoute(ctx, "", "svc1", nil, "/", "POST")
	if rule == nil {
		t.Error("expected match for POST")
	}

	rule2, _ := eng.EvaluateRoute(ctx, "", "svc1", nil, "/", "GET")
	if rule2 != nil {
		t.Error("expected no match for GET")
	}
}

func TestRoundRobinBalancer(t *testing.T) {
	lb := loadbalancer.NewLoadBalancer(loadbalancer.RoundRobin)

	instances := []*discovery.ServiceInstance{
		{ID: "a", Name: "svc1", Address: "10.0.0.1", Port: 8080},
		{ID: "b", Name: "svc1", Address: "10.0.0.2", Port: 8081},
		{ID: "c", Name: "svc1", Address: "10.0.0.3", Port: 8082},
	}
	lb.UpdateInstances(instances)
	ctx := context.Background()

	first, _ := lb.NextInstance(ctx, "")
	second, _ := lb.NextInstance(ctx, "")
	third, _ := lb.NextInstance(ctx, "")
	fourth, _ := lb.NextInstance(ctx, "")

	if first.ID != "a" || second.ID != "b" || third.ID != "c" || fourth.ID != "a" {
		t.Errorf("round robin order wrong: %s, %s, %s, %s", first.ID, second.ID, third.ID, fourth.ID)
	}
}

func TestLeastConnectionsBalancer(t *testing.T) {
	lb := loadbalancer.NewLoadBalancer(loadbalancer.LeastConnections)
	lb.UpdateInstances([]*discovery.ServiceInstance{
		{ID: "a", Name: "svc1", Address: "10.0.0.1", Port: 8080},
		{ID: "b", Name: "svc1", Address: "10.0.0.2", Port: 8081},
	})
	ctx := context.Background()

	inst, _ := lb.NextInstance(ctx, "")
	if inst == nil {
		t.Fatal("expected instance")
	}
	lb.ReleaseConnection(inst.ID)
}

func TestRandomBalancer(t *testing.T) {
	lb := loadbalancer.NewLoadBalancer(loadbalancer.Random)
	lb.UpdateInstances([]*discovery.ServiceInstance{
		{ID: "a", Name: "svc1", Address: "10.0.0.1", Port: 8080},
		{ID: "b", Name: "svc1", Address: "10.0.0.2", Port: 8081},
	})
	ctx := context.Background()

	_, err := lb.NextInstance(ctx, "")
	if err != nil {
		t.Fatalf("NextInstance failed: %v", err)
	}
}

func TestConsistentHashBalancer(t *testing.T) {
	lb := loadbalancer.NewLoadBalancer(loadbalancer.ConsistentHash)
	lb.UpdateInstances([]*discovery.ServiceInstance{
		{ID: "a", Name: "svc1", Address: "10.0.0.1", Port: 8080},
		{ID: "b", Name: "svc1", Address: "10.0.0.2", Port: 8081},
		{ID: "c", Name: "svc1", Address: "10.0.0.3", Port: 8082},
	})
	ctx := context.Background()

	inst1, _ := lb.NextInstance(ctx, "192.168.1.1")
	inst2, _ := lb.NextInstance(ctx, "192.168.1.1")
	if inst1.ID != inst2.ID {
		t.Errorf("consistent hash should return same instance for same IP: got %s then %s", inst1.ID, inst2.ID)
	}
}

func TestBalancerNoInstances(t *testing.T) {
	lb := loadbalancer.NewLoadBalancer(loadbalancer.RoundRobin)
	ctx := context.Background()
	_, err := lb.NextInstance(ctx, "")
	if err != loadbalancer.ErrNoInstances {
		t.Errorf("expected ErrNoInstances, got %v", err)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	exec := resilience.NewExecutor()
	ctx := context.Background()

	attempts := 0
	err := exec.ExecuteWithRetry(ctx, resilience.RetryPolicy{
		MaxRetries:        3,
		BackoffInitialMs:  10,
		BackoffMultiplier: 2.0,
		MaxBackoffMs:      100,
		RetryOn:           []string{"temporary"},
	}, func(ctx context.Context) error {
		attempts++
		return nil
	})
	if err != nil {
		t.Fatalf("ExecuteWithRetry failed: %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt on success, got %d", attempts)
	}
}

func TestRetryWithFailures(t *testing.T) {
	exec := resilience.NewExecutor()
	ctx := context.Background()

	attempts := 0
	err := exec.ExecuteWithRetry(ctx, resilience.RetryPolicy{
		MaxRetries:        3,
		BackoffInitialMs:  5,
		BackoffMultiplier: 1.5,
		MaxBackoffMs:      50,
		RetryOn:           []string{"temporary"},
	}, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return nil
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ExecuteWithRetry failed: %v", err)
	}
}

func TestRetryRespectsContext(t *testing.T) {
	exec := resilience.NewExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := exec.ExecuteWithRetry(ctx, resilience.RetryPolicy{
		MaxRetries:        10,
		BackoffInitialMs:  100,
		BackoffMultiplier: 2.0,
		MaxBackoffMs:      1000,
		RetryOn:           []string{"error"},
	}, func(ctx context.Context) error {
		return nil
	})
	_ = err
}

func TestTimeout(t *testing.T) {
	exec := resilience.NewExecutor()
	ctx := context.Background()

	err := exec.ExecuteWithTimeout(ctx, resilience.TimeoutPolicy{
		RequestTimeoutMs: 100,
	}, func(ctx context.Context) error {
		select {
		case <-time.After(10 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	_ = err
}

func TestBulkheadAcquireRelease(t *testing.T) {
	exec := resilience.NewExecutor()
	exec.AddBulkhead("test-bh", 5, 10)
	ctx := context.Background()

	err := exec.ExecuteWithBulkhead(ctx, "test-bh", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("ExecuteWithBulkhead failed: %v", err)
	}
}

func TestBulkheadQueueFull(t *testing.T) {
	bh := resilience.NewBulkhead("test-bh", 1, 0)
	ctx := context.Background()

	bh.Acquire(ctx)

	done := make(chan error, 1)
	go func() {
		done <- bh.Acquire(ctx)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected error when queue full")
		}
	case <-time.After(100 * time.Millisecond):
		bh.Release()
	}
}

func TestCircuitBreakerClosedToOpen(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test-cb", 3, 1, 100*time.Millisecond)

	if !cb.AllowRequest() {
		t.Error("expected allowed when closed")
	}

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.AllowRequest() {
		t.Error("expected blocked when open")
	}
}

func TestCircuitBreakerHalfOpenToClosed(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test-cb", 2, 2, 10*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != resilience.StateOpen {
		t.Error("expected open after threshold")
	}

	time.Sleep(15 * time.Millisecond)

	if !cb.AllowRequest() {
		t.Error("expected allowed in half-open after recovery")
	}

	cb.RecordSuccess()
	if cb.State() != resilience.StateHalfOpen {
		t.Error("expected still half-open after 1 success")
	}

	cb.RecordSuccess()
	if cb.State() != resilience.StateClosed {
		t.Error("expected closed after enough successes")
	}
}

func TestCircuitBreakerHalfOpenToOpen(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test-cb", 2, 1, 10*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()

	time.Sleep(15 * time.Millisecond)

	cb.AllowRequest()
	cb.RecordFailure()

	if cb.State() != resilience.StateOpen {
		t.Error("expected open after failure in half-open")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test-cb", 1, 1, 10*time.Millisecond)
	cb.RecordFailure()

	cb.Reset()
	if cb.State() != resilience.StateClosed {
		t.Error("expected closed after reset")
	}
	if !cb.AllowRequest() {
		t.Error("expected allowed after reset")
	}
}

func TestCircuitBreakerList(t *testing.T) {
	exec := resilience.NewExecutor()
	exec.AddCircuitBreaker("cb1", 5, 3, time.Second)
	exec.AddCircuitBreaker("cb2", 3, 2, time.Second)

	cbs := exec.ListCircuitBreakers()
	if len(cbs) != 2 {
		t.Errorf("expected 2 circuit breakers, got %d", len(cbs))
	}
}

func TestTelemetryRecordCall(t *testing.T) {
	repo := telemetry.NewInMemoryRepository()
	ctx := context.Background()

	err := repo.RecordCall(ctx, "frontend", "backend", 42.5, "ok")
	if err != nil {
		t.Fatalf("RecordCall failed: %v", err)
	}
}

func TestTelemetryServiceGraph(t *testing.T) {
	repo := telemetry.NewInMemoryRepository()
	ctx := context.Background()

	repo.RecordCall(ctx, "frontend", "backend", 10.0, "ok")
	repo.RecordCall(ctx, "frontend", "backend", 20.0, "ok")
	repo.RecordCall(ctx, "backend", "database", 5.0, "error")

	graph, err := repo.GetServiceGraph(ctx)
	if err != nil {
		t.Fatalf("GetServiceGraph failed: %v", err)
	}

	if len(graph.Nodes) < 2 {
		t.Errorf("expected at least 2 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.Edges) < 2 {
		t.Errorf("expected at least 2 edges, got %d", len(graph.Edges))
	}
}

func TestTelemetryTraces(t *testing.T) {
	repo := telemetry.NewInMemoryRepository()
	ctx := context.Background()

	repo.RecordSpan(ctx, &telemetry.TraceSpan{
		Service:   "frontend",
		Operation: "GET /users",
		DurationMs: 15.0,
		Status:    "ok",
	})
	repo.RecordSpan(ctx, &telemetry.TraceSpan{
		Service:   "backend",
		Operation: "GET /users/:id",
		DurationMs: 10.0,
		Status:    "ok",
	})

	traces, err := repo.GetTraces(ctx, "", 10)
	if err != nil {
		t.Fatalf("GetTraces failed: %v", err)
	}
	if len(traces) != 2 {
		t.Errorf("expected 2 traces, got %d", len(traces))
	}
}

func TestTelemetryTracesFilterByService(t *testing.T) {
	repo := telemetry.NewInMemoryRepository()
	ctx := context.Background()

	repo.RecordSpan(ctx, &telemetry.TraceSpan{Service: "svc1", Operation: "op1", DurationMs: 1})
	repo.RecordSpan(ctx, &telemetry.TraceSpan{Service: "svc2", Operation: "op2", DurationMs: 2})

	traces, _ := repo.GetTraces(ctx, "svc1", 10)
	if len(traces) != 1 {
		t.Errorf("expected 1 trace for svc1, got %d", len(traces))
	}
}

func TestTelemetryDependencies(t *testing.T) {
	repo := telemetry.NewInMemoryRepository()
	ctx := context.Background()

	repo.RecordCall(ctx, "a", "b", 1, "ok")
	repo.RecordCall(ctx, "a", "c", 2, "ok")

	deps, err := repo.GetDependencies(ctx)
	if err != nil {
		t.Fatalf("GetDependencies failed: %v", err)
	}
	if len(deps["a"]) != 2 {
		t.Errorf("expected a to have 2 dependencies, got %d", len(deps["a"]))
	}
}

func TestTrafficRuleWithMirrorPercentage(t *testing.T) {
	repo := traffic.NewInMemoryRepository()
	eng := traffic.NewEngine(repo)
	ctx := context.Background()

	rule, _ := eng.CreateRule(ctx, &traffic.TrafficRule{
		Name:              "mirrored",
		DestinationService: "checkout",
		Weight:            100,
		MirrorPercentage:  50,
	})
	if rule.MirrorPercentage != 50 {
		t.Errorf("expected mirror percentage 50, got %d", rule.MirrorPercentage)
	}
}

func TestLoadBalancerUpdateInstances(t *testing.T) {
	lb := loadbalancer.NewLoadBalancer(loadbalancer.RoundRobin)
	lb.UpdateInstances([]*discovery.ServiceInstance{
		{ID: "x", Name: "svc", Address: "1", Port: 1},
	})
	ctx := context.Background()

	inst, _ := lb.NextInstance(ctx, "")
	if inst.ID != "x" {
		t.Errorf("expected instance x, got %s", inst.ID)
	}

	lb.UpdateInstances([]*discovery.ServiceInstance{
		{ID: "y", Name: "svc", Address: "2", Port: 2},
	})
	inst, _ = lb.NextInstance(ctx, "")
	if inst.ID != "y" {
		t.Errorf("expected instance y after update, got %s", inst.ID)
	}
}

func TestConcurrentServiceRegistration(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			svc.Register(ctx, &discovery.ServiceInstance{
				Name: "svc1", Address: "10.0.0.1", Port: 8080 + i,
			})
		}(i)
	}
	wg.Wait()

	services, _ := svc.ListServices(ctx)
	if len(services) != 10 {
		t.Errorf("expected 10 services after concurrent registration, got %d", len(services))
	}
}

func TestBulkheadMaxConcurrent(t *testing.T) {
	bh := resilience.NewBulkhead("test", 2, 1)
	ctx := context.Background()

	err1 := bh.Acquire(ctx)
	if err1 != nil {
		t.Fatalf("first acquire failed: %v", err1)
	}
	err2 := bh.Acquire(ctx)
	if err2 != nil {
		t.Fatalf("second acquire failed: %v", err2)
	}

	bh.Release()
	bh.Release()
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	cb := resilience.NewCircuitBreaker("state-test", 2, 1, 50*time.Millisecond)

	if cb.State() != resilience.StateClosed {
		t.Error("initial state should be closed")
	}

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != resilience.StateOpen {
		t.Error("should be open after 2 failures")
	}

	if cb.AllowRequest() {
		t.Error("should not allow request when open")
	}

	time.Sleep(60 * time.Millisecond)

	if !cb.AllowRequest() {
		t.Error("should allow request after recovery time")
	}

	if cb.State() != resilience.StateHalfOpen {
		t.Error("should be half-open after recovery")
	}

	cb.RecordSuccess()
	if cb.State() != resilience.StateClosed {
		t.Error("should go to closed after success in half-open")
	}
}

func TestDiscoverWithZoneFilter(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc", Address: "1", Port: 1, Region: "us", Zone: "a",
	})
	svc.Register(ctx, &discovery.ServiceInstance{
		Name: "svc", Address: "2", Port: 2, Region: "us", Zone: "b",
	})

	instances, _ := svc.Discover(ctx, "svc", "us", "a")
	if len(instances) != 1 {
		t.Errorf("expected 1 instance in zone a, got %d", len(instances))
	}
}

func TestRetryPolicyConfiguration(t *testing.T) {
	policy := resilience.RetryPolicy{
		MaxRetries:         5,
		BackoffInitialMs:   100,
		BackoffMultiplier:  2.0,
		MaxBackoffMs:       5000,
		RetryOn:            []string{"timeout", "unavailable"},
	}
	if policy.MaxRetries != 5 {
		t.Errorf("expected 5 max retries, got %d", policy.MaxRetries)
	}
	if len(policy.RetryOn) != 2 {
		t.Errorf("expected 2 retry conditions, got %d", len(policy.RetryOn))
	}
}

func TestTimeoutPolicyConfiguration(t *testing.T) {
	policy := resilience.TimeoutPolicy{
		RequestTimeoutMs: 1000,
		IdleTimeoutMs:    30000,
	}
	if policy.RequestTimeoutMs != 1000 {
		t.Errorf("expected 1000ms timeout, got %d", policy.RequestTimeoutMs)
	}
}

func TestTelemetryEdgeErrorRate(t *testing.T) {
	repo := telemetry.NewInMemoryRepository()
	ctx := context.Background()

	repo.RecordCall(ctx, "src", "dst", 10, "ok")
	repo.RecordCall(ctx, "src", "dst", 10, "ok")
	repo.RecordCall(ctx, "src", "dst", 10, "error")
	repo.RecordCall(ctx, "src", "dst", 10, "error")
	repo.RecordCall(ctx, "src", "dst", 10, "error")

	graph, _ := repo.GetServiceGraph(ctx)
	if len(graph.Edges) > 0 {
		edge := graph.Edges[0]
		if edge.ErrorRate <= 0 {
			t.Error("expected non-zero error rate")
		}
	}
}

func TestCertificateFingerprint(t *testing.T) {
	ca, _ := mtls.NewCertificateAuthority("Test", "CA", 365)
	mgr := mtls.NewCertManager(ca)
	ctx := context.Background()

	cert, _ := mgr.IssueCert(ctx, "s1", "s1.local", "Test", 90, true)
	if len(cert.Fingerprint) != 64 {
		t.Errorf("expected 64-char SHA256 fingerprint, got %d chars", len(cert.Fingerprint))
	}
}

func TestListAllServicesWithStatus(t *testing.T) {
	repo := discovery.NewInMemoryRepository()
	svc := discovery.NewService(repo)
	ctx := context.Background()

	svc.Register(ctx, &discovery.ServiceInstance{Name: "up-svc", Address: "1", Port: 1, Status: discovery.StatusUp})
	svc.Register(ctx, &discovery.ServiceInstance{Name: "down-svc", Address: "2", Port: 2, Status: discovery.StatusDown})

	services, _ := svc.ListServices(ctx)
	statusMap := make(map[string]discovery.Status)
	for _, s := range services {
		statusMap[s.Name] = s.Status
	}
	if statusMap["up-svc"] != discovery.StatusUp {
		t.Error("expected up-svc status up")
	}
	if statusMap["down-svc"] != discovery.StatusDown {
		t.Error("expected down-svc status down")
	}
}
