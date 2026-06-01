package http

import (
	"github.com/tikiclone/tiki/platforms/live-scale/internal/cdn"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/region"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/sfu"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/stream_health"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/transcoding"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/websocket_cluster"
)

type Handler struct {
	sfu       *sfu.Service
	cdn       *cdn.Service
	cluster   *websocket_cluster.Service
	health    *stream_health.Service
	region    *region.Service
	transcode *transcoding.Service
}

func NewHandler(
	sfuSvc *sfu.Service,
	cdnSvc *cdn.Service,
	clusterSvc *websocket_cluster.Service,
	healthSvc *stream_health.Service,
	regionSvc *region.Service,
	transcodeSvc *transcoding.Service,
) *Handler {
	return &Handler{
		sfu:       sfuSvc,
		cdn:       cdnSvc,
		cluster:   clusterSvc,
		health:    healthSvc,
		region:    regionSvc,
		transcode: transcodeSvc,
	}
}
