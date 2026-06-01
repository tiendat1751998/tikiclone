package http

import (
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/bulkindexer"
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/coordinator"
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/monitoring"
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/pipeline"
	"github.com/tikiclone/tiki/platforms/search-indexing/internal/synonyms"
)

type Handler struct {
	coordinator coordinator.Service
	bulkindexer bulkindexer.Service
	pipeline    pipeline.Service
	synonyms    synonyms.Service
	monitoring  monitoring.Service
}

func NewHandler(
	coordinator coordinator.Service,
	bulkindexer bulkindexer.Service,
	pipeline pipeline.Service,
	synonyms synonyms.Service,
	monitoring monitoring.Service,
) *Handler {
	return &Handler{
		coordinator: coordinator,
		bulkindexer: bulkindexer,
		pipeline:    pipeline,
		synonyms:    synonyms,
		monitoring:  monitoring,
	}
}
