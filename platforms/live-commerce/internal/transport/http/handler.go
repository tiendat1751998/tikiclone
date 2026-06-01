package http

import (
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/application"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
)

type Handler struct {
	service *application.LiveCommerceService
}

func NewHandler(s *application.LiveCommerceService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateLivestream(c *gin.Context) {
	var req struct {
		SellerID    string     `json:"seller_id" binding:"required"`
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description"`
		CoverURL    string     `json:"cover_url"`
		Category    string     `json:"category"`
		Tags        []string   `json:"tags"`
		ScheduledAt *time.Time `json:"scheduled_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	ls, err := h.service.CreateLivestream(c.Request.Context(), req.SellerID, req.Title, req.Description, req.CoverURL, req.Category, req.Tags, req.ScheduledAt)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ls)
}

func (h *Handler) GetLivestream(c *gin.Context) {
	ls, err := h.service.GetLivestream(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, ls)
}

func (h *Handler) StartLivestream(c *gin.Context) {
	if err := h.service.StartLivestream(c.Request.Context(), c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "livestream started"})
}

func (h *Handler) EndLivestream(c *gin.Context) {
	if err := h.service.EndLivestream(c.Request.Context(), c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "livestream ended"})
}

func (h *Handler) ListActiveLivestreams(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	streams, total, err := h.service.ListActiveLivestreams(c.Request.Context(), offset, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": streams, "total": total, "offset": offset, "limit": limit})
}

func (h *Handler) ListSellerLivestreams(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sellerID := c.Param("seller_id")
	streams, total, err := h.service.ListSellerLivestreams(c.Request.Context(), sellerID, offset, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": streams, "total": total, "offset": offset, "limit": limit})
}

func (h *Handler) SendMessage(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Username string `json:"username" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	msg, err := h.service.SendChatMessage(c.Request.Context(), c.Param("room_id"), req.UserID, req.Username, req.Content)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h *Handler) SendReaction(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Type   string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	if err := h.service.SendReaction(c.Request.Context(), c.Param("room_id"), req.UserID, req.Type); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reaction sent"})
}

func (h *Handler) SendGift(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Username string `json:"username" binding:"required"`
		GiftType string `json:"gift_type" binding:"required"`
		Amount   int64  `json:"amount" binding:"required"`
		Currency string `json:"currency" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	if err := h.service.SendGift(c.Request.Context(), c.Param("room_id"), req.UserID, req.Username, req.GiftType, req.Amount, req.Currency); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "gift sent"})
}

func (h *Handler) PinProduct(c *gin.Context) {
	var req struct {
		ProductID   string `json:"product_id" binding:"required"`
		ProductName string `json:"product_name" binding:"required"`
		Price       int64  `json:"price" binding:"required"`
		ImageURL    string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	pp, err := h.service.PinProduct(c.Request.Context(), c.Param("id"), req.ProductID, req.ProductName, req.Price, req.ImageURL)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, pp)
}

func (h *Handler) UnpinProduct(c *gin.Context) {
	if err := h.service.UnpinProduct(c.Request.Context(), c.Param("id"), c.Param("product_id")); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product unpinned"})
}

func (h *Handler) GetPinnedProducts(c *gin.Context) {
	products, err := h.service.GetPinnedProducts(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *Handler) GetViewerCount(c *gin.Context) {
	count, err := h.service.GetViewerCount(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"viewer_count": count})
}

func (h *Handler) GetReactionSummary(c *gin.Context) {
	summary, err := h.service.GetReactionSummary(c.Request.Context(), c.Param("room_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) GetGiftLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	board, err := h.service.GetGiftLeaderboard(c.Request.Context(), c.Param("room_id"), limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": board})
}

func (h *Handler) GetChatHistory(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	messages, total, err := h.service.GetChatHistory(c.Request.Context(), c.Param("room_id"), offset, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": messages, "total": total, "offset": offset, "limit": limit})
}

func (h *Handler) ModerateAction(c *gin.Context) {
	var req struct {
		RoomID      string `json:"room_id" binding:"required"`
		UserID      string `json:"user_id" binding:"required"`
		Action      string `json:"action" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		ModeratedBy string `json:"moderated_by" binding:"required"`
		DurationSec int64  `json:"duration_sec"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	if err := h.service.ModerateAction(c.Request.Context(), req.RoomID, req.UserID, req.Action, req.Reason, req.ModeratedBy, req.DurationSec); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "action applied"})
}

func (h *Handler) GetTrendingLivestreams(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	trending := h.service.GetTrendingLivestreams(c.Request.Context(), limit)
	c.JSON(http.StatusOK, gin.H{"data": trending})
}

func (h *Handler) WSAuth(c *gin.Context) {
	h.service.HandleViewerJoined(c.Request.Context(), c.Param("room_id"), c.GetString("user_id"))
	c.Next()
}

var errorStatusMap = map[string]int{
	domain.ErrLivestreamNotFound.Error(): http.StatusNotFound,
	domain.ErrInvalidLiveState.Error():   http.StatusConflict,
	domain.ErrRoomNotFound.Error():       http.StatusNotFound,
	domain.ErrAlreadyEnded.Error():       http.StatusConflict,
	domain.ErrNotLive.Error():            http.StatusConflict,
	domain.ErrUserMuted.Error():          http.StatusForbidden,
	domain.ErrUserBanned.Error():         http.StatusForbidden,
	domain.ErrInvalidReaction.Error():    http.StatusBadRequest,
	domain.ErrMessageTooLong.Error():     http.StatusBadRequest,
}

func handleError(c *gin.Context, err error) {
	if code, ok := errorStatusMap[err.Error()]; ok {
		c.AbortWithStatusJSON(code, gin.H{"error_code": err.Error(), "message": err.Error()})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": "An unexpected error occurred"})
}
