package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/discovery"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/mtls"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/resilience"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/traffic"
)

func (h *Handler) RegisterService(c *gin.Context) {
	var inst discovery.ServiceInstance
	if err := c.ShouldBindJSON(&inst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.discoverySvc.Register(c.Request.Context(), &inst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inst)
}

func (h *Handler) Heartbeat(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.discoverySvc.Heartbeat(c.Request.Context(), req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) ListServices(c *gin.Context) {
	services, err := h.discoverySvc.ListServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services)
}

func (h *Handler) DiscoverServices(c *gin.Context) {
	name := c.Query("name")
	region := c.Query("region")
	zone := c.Query("zone")
	instances, err := h.discoverySvc.Discover(c.Request.Context(), name, region, zone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, instances)
}

func (h *Handler) CreateCA(c *gin.Context) {
	var req struct {
		Organization string `json:"organization"`
		CommonName   string `json:"common_name"`
		ValidityDays int    `json:"validity_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ValidityDays <= 0 {
		req.ValidityDays = 365
	}
	ca, err := mtls.NewCertificateAuthority(req.Organization, req.CommonName, req.ValidityDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ca.ListCerts())
}

func (h *Handler) IssueCertificate(c *gin.Context) {
	var req struct {
		ServiceName  string `json:"service_name"`
		CommonName   string `json:"common_name"`
		Organization string `json:"organization"`
		ValidityDays int    `json:"validity_days"`
		IsServer     bool   `json:"is_server"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ValidityDays <= 0 {
		req.ValidityDays = 365
	}
	cert, err := h.certManager.IssueCert(c.Request.Context(), req.ServiceName, req.CommonName, req.Organization, req.ValidityDays, req.IsServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cert)
}

func (h *Handler) RenewCertificate(c *gin.Context) {
	var req struct {
		CertID       string `json:"cert_id"`
		ValidityDays int    `json:"validity_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ValidityDays <= 0 {
		req.ValidityDays = 365
	}
	cert, err := h.certManager.RenewCert(c.Request.Context(), req.CertID, req.ValidityDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cert)
}

func (h *Handler) RevokeCertificate(c *gin.Context) {
	var req struct {
		CertID string `json:"cert_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.certManager.RevokeCert(c.Request.Context(), req.CertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

func (h *Handler) ListCertificates(c *gin.Context) {
	certs := h.certManager.ListCerts(c.Request.Context())
	c.JSON(http.StatusOK, certs)
}

func (h *Handler) VerifyCertificate(c *gin.Context) {
	var req struct {
		CertID string `json:"cert_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.certManager.VerifyCert(c.Request.Context(), req.CertID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "valid"})
}

func (h *Handler) CreateTrafficRule(c *gin.Context) {
	var rule traffic.TrafficRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.trafficEng.CreateRule(c.Request.Context(), &rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

func (h *Handler) ListTrafficRules(c *gin.Context) {
	rules, err := h.trafficEng.ListRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *Handler) EvaluateRoute(c *gin.Context) {
	var req struct {
		Source      string            `json:"source_service"`
		Destination string            `json:"destination_service"`
		Headers     map[string]string `json:"headers"`
		Path        string            `json:"path"`
		Method      string            `json:"method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.trafficEng.EvaluateRoute(c.Request.Context(), req.Source, req.Destination, req.Headers, req.Path, req.Method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *Handler) GetNextInstance(c *gin.Context) {
	var req struct {
		SourceIP string `json:"source_ip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inst, err := h.lb.NextInstance(c.Request.Context(), req.SourceIP)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inst)
}

func (h *Handler) ExecuteWithRetry(c *gin.Context) {
	var req struct {
		Policy resilience.RetryPolicy `json:"policy"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.executor.ExecuteWithRetry(c.Request.Context(), req.Policy, func(ctx context.Context) error {
		return nil
	})
	c.JSON(http.StatusOK, gin.H{"executed": err == nil, "error": func() string {
		if err != nil {
			return err.Error()
		}
		return ""
	}()})
}

func (h *Handler) ExecuteWithBulkhead(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.executor.ExecuteWithBulkhead(c.Request.Context(), req.Name, func(ctx context.Context) error {
		return nil
	})
	c.JSON(http.StatusOK, gin.H{"executed": err == nil, "error": func() string {
		if err != nil {
			return err.Error()
		}
		return ""
	}()})
}

func (h *Handler) ListCircuitBreakers(c *gin.Context) {
	cbs := h.executor.ListCircuitBreakers()
	type stateInfo struct {
		Name  string `json:"name"`
		State string `json:"state"`
	}
	var result []stateInfo
	for _, cb := range cbs {
		state := "closed"
		switch cb.State() {
		case resilience.StateOpen:
			state = "open"
		case resilience.StateHalfOpen:
			state = "half_open"
		}
		result = append(result, stateInfo{Name: cb.Name(), State: state})
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) RecordCall(c *gin.Context) {
	var req struct {
		Source      string  `json:"source"`
		Destination string  `json:"destination"`
		DurationMs  float64 `json:"duration_ms"`
		Status      string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.telemetry.RecordCall(c.Request.Context(), req.Source, req.Destination, req.DurationMs, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

func (h *Handler) GetServiceGraph(c *gin.Context) {
	graph, err := h.telemetry.GetServiceGraph(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, graph)
}

func (h *Handler) GetTraces(c *gin.Context) {
	service := c.Query("service")
	traces, err := h.telemetry.GetTraces(c.Request.Context(), service, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, traces)
}
