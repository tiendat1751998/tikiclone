package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/analytics/internal/cohort"
)

type analyzeCohortRequest struct {
	Name             string `json:"name" binding:"required"`
	Period           string `json:"period"`
	AcquisitionField string `json:"acquisition_field"`
	TimeRange        string `json:"time_range"`
}

func (h *Handler) AnalyzeCohort(c *gin.Context) {
	var req analyzeCohortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	period := cohort.CohortPeriod(req.Period)
	if period == "" {
		period = cohort.CohortWeek
	}

	definition := &cohort.CohortDefinition{
		Name:             req.Name,
		Period:           period,
		AcquisitionField: req.AcquisitionField,
		TimeRange:        req.TimeRange,
	}

	result, err := h.cohortSvc.BuildCohort(c.Request.Context(), definition)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
