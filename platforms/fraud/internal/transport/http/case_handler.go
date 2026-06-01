package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	fraudcase "github.com/tikiclone/tiki/platforms/fraud/internal/case"
)

type createCaseRequest struct {
	AlertID     string  `json:"alert_id" binding:"required"`
	UserID      string  `json:"user_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	RiskScore   float64 `json:"risk_score"`
	Priority    string  `json:"priority"`
}

func (h *Handler) CreateCase(c *gin.Context) {
	var req createCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := fraudcase.PriorityMedium
	if req.Priority != "" {
		priority = fraudcase.CasePriority(req.Priority)
	}

	created, err := h.caseSvc.CreateCase(
		c.Request.Context(),
		req.AlertID,
		req.UserID,
		req.Title,
		req.Description,
		req.RiskScore,
		priority,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) ListCases(c *gin.Context) {
	status := fraudcase.CaseStatus(c.DefaultQuery("status", ""))
	priority := fraudcase.CasePriority(c.DefaultQuery("priority", ""))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	cases, total, err := h.caseSvc.ListCases(c.Request.Context(), status, priority, offset, limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cases, "total": total})
}

type updateCaseRequest struct {
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Investigator string `json:"investigator"`
	Resolution  string `json:"resolution"`
}

func (h *Handler) UpdateCase(c *gin.Context) {
	var req updateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	caseID := c.Param("id")

	if req.Investigator != "" {
		if err := h.caseSvc.AssignInvestigator(c.Request.Context(), caseID, req.Investigator); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.Priority != "" {
		if err := h.caseSvc.Escalate(c.Request.Context(), caseID, fraudcase.CasePriority(req.Priority)); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.Status != "" {
		if err := h.caseSvc.UpdateStatus(c.Request.Context(), caseID, fraudcase.CaseStatus(req.Status), req.Resolution); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	updated, err := h.caseSvc.GetCase(c.Request.Context(), caseID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}
