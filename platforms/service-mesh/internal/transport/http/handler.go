package http

import (
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/discovery"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/loadbalancer"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/mtls"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/resilience"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/telemetry"
	"github.com/tikiclone/tiki/platforms/service-mesh/internal/traffic"
)

type Handler struct {
	discoverySvc *discovery.Service
	certManager  *mtls.CertManager
	trafficEng   *traffic.Engine
	lb           *loadbalancer.LoadBalancer
	executor     *resilience.Executor
	telemetry    *telemetry.InMemoryRepository
}

func NewHandler(
	ds *discovery.Service,
	cm *mtls.CertManager,
	te *traffic.Engine,
	lb *loadbalancer.LoadBalancer,
	ex *resilience.Executor,
	tel *telemetry.InMemoryRepository,
) *Handler {
	return &Handler{
		discoverySvc: ds,
		certManager:  cm,
		trafficEng:   te,
		lb:           lb,
		executor:     ex,
		telemetry:    tel,
	}
}
