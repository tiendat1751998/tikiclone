package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

type searchRequest struct {
	Query     string  `json:"query"`
	Category  string  `json:"category"`
	Brand     string  `json:"brand"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	MinRating float64 `json:"min_rating"`
	SortBy    string  `json:"sort_by"`
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
}

func (h *Handler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if q := c.Query("q"); q != "" {
			req.Query = q
			req.Page = parseInt(c.Query("page"), 1)
			req.Limit = parseInt(c.Query("limit"), 20)
			req.Category = c.Query("category")
			req.SortBy = c.Query("sort_by")
			req.MinPrice = parseFloat(c.Query("min_price"), 0)
			req.MaxPrice = parseFloat(c.Query("max_price"), 0)
			req.MinRating = parseFloat(c.Query("min_rating"), 0)
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
	}

	q := search.SearchQuery{
		Query:     req.Query,
		Category:  req.Category,
		Brand:     req.Brand,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		MinRating: req.MinRating,
		SortBy:    req.SortBy,
		Page:      req.Page,
		Limit:     req.Limit,
	}

	result, err := h.search.Search(c.Request.Context(), q)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) FacetedSearch(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	q := search.SearchQuery{
		Query:     req.Query,
		Category:  req.Category,
		Brand:     req.Brand,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		MinRating: req.MinRating,
		SortBy:    req.SortBy,
		Page:      req.Page,
		Limit:     req.Limit,
	}

	result, err := h.search.FacetedSearch(c.Request.Context(), q)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return def
}
