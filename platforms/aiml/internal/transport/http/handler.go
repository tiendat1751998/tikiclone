package http

import (
	"github.com/tikiclone/tiki/platforms/aiml/internal/embeddings"
	"github.com/tikiclone/tiki/platforms/aiml/internal/experiments"
	"github.com/tikiclone/tiki/platforms/aiml/internal/featurestore"
	"github.com/tikiclone/tiki/platforms/aiml/internal/inference"
	"github.com/tikiclone/tiki/platforms/aiml/internal/modelregistry"
	"github.com/tikiclone/tiki/platforms/aiml/internal/training"
)

type Handler struct {
	featureSvc   *featurestore.Service
	modelSvc     *modelregistry.Service
	trainingSvc  *training.Service
	inferenceSvc *inference.Service
	embedSvc     *embeddings.Service
	experimentSvc *experiments.Service
}

func NewHandler(
	featureSvc *featurestore.Service,
	modelSvc *modelregistry.Service,
	trainingSvc *training.Service,
	inferenceSvc *inference.Service,
	embedSvc *embeddings.Service,
	experimentSvc *experiments.Service,
) *Handler {
	return &Handler{
		featureSvc:    featureSvc,
		modelSvc:      modelSvc,
		trainingSvc:   trainingSvc,
		inferenceSvc:  inferenceSvc,
		embedSvc:      embedSvc,
		experimentSvc: experimentSvc,
	}
}
