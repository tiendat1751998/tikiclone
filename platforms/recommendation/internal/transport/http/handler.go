package http

import (
	"github.com/tikiclone/tiki/platforms/recommendation/internal/collaborative"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/events"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/recommender"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/trending"
)

type Handler struct {
	recommender recommender.Service
	trending    trending.Service
	collab      collaborative.Service
	publisher   events.Publisher
}

func NewHandler(
	recSvc recommender.Service,
	trendingSvc trending.Service,
	collabSvc collaborative.Service,
	pub events.Publisher,
) *Handler {
	return &Handler{
		recommender: recSvc,
		trending:    trendingSvc,
		collab:      collabSvc,
		publisher:   pub,
	}
}
