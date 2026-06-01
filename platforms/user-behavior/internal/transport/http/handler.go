package http
import ("net/http"; "github.com/gin-gonic/gin"; "github.com/tikiclone/tiki/platforms/user-behavior/internal/application"; "github.com/tikiclone/tiki/platforms/user-behavior/internal/domain"; "github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "go.opentelemetry.io/otel"; "go.uber.org/zap")

type Handler struct { service *application.BehaviorService }
func NewHandler(s *application.BehaviorService) *Handler { return &Handler{service: s} }

func (h *Handler) IngestEvent(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-user-behavior").Start(c.Request.Context(), "http.ingest")
	var event domain.ClickEvent
	if err := c.ShouldBindJSON(&event); err != nil { c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()}); return }
	if err := h.service.IngestEvent(ctx, &event); err != nil { handleError(c, err); return }
	c.JSON(http.StatusAccepted, gin.H{"id": event.ID})
}

func (h *Handler) IngestBatch(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-user-behavior").Start(c.Request.Context(), "http.ingest_batch")
	var events []*domain.ClickEvent
	if err := c.ShouldBindJSON(&events); err != nil { c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()}); return }
	count, err := h.service.IngestBatch(ctx, events)
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusAccepted, gin.H{"ingested": count})
}

func (h *Handler) GetTrending(c *gin.Context) {
	ctx, _ := otel.Tracer("tiki-user-behavior").Start(c.Request.Context(), "http.trending")
	products, err := h.service.GetTrendingProducts(ctx, 20)
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusOK, gin.H{"trending": products})
}

var errorStatusMap = map[error]int{domain.ErrEventValidation: http.StatusBadRequest}
func handleError(c *gin.Context, err error) {
	for e, s := range errorStatusMap {
		if err.Error() == e.Error() { c.AbortWithStatusJSON(s, gin.H{"error_code": e.Error(), "message": err.Error()}); return }
	}
	observability.LogWithTrace(c.Request.Context()).Error("unhandled error", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": "An unexpected error occurred"})
}
