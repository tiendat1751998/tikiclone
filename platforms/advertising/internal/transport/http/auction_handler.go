package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/advertising/internal/bidding"
	"github.com/tikiclone/tiki/platforms/advertising/internal/metrics"
)

type auctionRequest struct {
	UserID  string           `json:"user_id"`
	Context bidding.BidContext `json:"context"`
	MaxBid  float64          `json:"max_bid"`
}

type auctionResponse struct {
	Winner      *bidding.BidResponse   `json:"winner"`
	SecondPrice float64                `json:"second_price"`
	AllBids     []bidding.BidResponse  `json:"all_bids"`
}

func (h *Handler) RunAuction(c *gin.Context) {
	var req auctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	metrics.AuctionsTotal.Inc()

	result, err := h.biddingSvc.RunAuction(c.Request.Context(), &bidding.BidRequest{
		UserID:  req.UserID,
		Context: req.Context,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, auctionResponse{
		Winner:      result.Winner,
		SecondPrice: result.SecondPrice,
		AllBids:     result.AllBids,
	})
}
