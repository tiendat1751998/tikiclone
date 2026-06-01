package delivery

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/shipment/internal/domain/delivery"
	"github.com/tikiclone/tiki/services/shipment/internal/infrastructure/geo"
	"go.uber.org/zap"
)

type GeoHandler struct {
	geoService *geo.Service
	logger     *zap.Logger
}

func NewGeoHandler(geoService *geo.Service, logger *zap.Logger) *GeoHandler {
	return &GeoHandler{geoService: geoService, logger: logger.Named("geo_handler")}
}

func (h *GeoHandler) SearchAddress(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}
	if len(query) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too long (max 200 chars)"})
		return
	}

	results, err := h.geoService.SearchAddress(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("address search failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "address search failed"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *GeoHandler) ReverseGeocode(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng parameters are required"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat value"})
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng value"})
		return
	}

	result, err := h.geoService.ReverseGeocode(c.Request.Context(), lat, lng)
	if err != nil {
		h.logger.Error("reverse geocode failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reverse geocoding failed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *GeoHandler) CalculateRoute(c *gin.Context) {
	var req struct {
		PickupLat  float64 `json:"pickup_lat" binding:"required"`
		PickupLng  float64 `json:"pickup_lng" binding:"required"`
		DropoffLat float64 `json:"dropoff_lat" binding:"required"`
		DropoffLng float64 `json:"dropoff_lng" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	result, err := h.geoService.CalculateRoute(c.Request.Context(), req.PickupLat, req.PickupLng, req.DropoffLat, req.DropoffLng)
	if err != nil {
		h.logger.Error("route calculation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "route calculation failed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *GeoHandler) UpdateDriverLocation(c *gin.Context) {
	var req delivery.DriverLocationUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if err := h.geoService.UpdateDriverLocation(c.Request.Context(), req.DriverID, req.Lat, req.Lng); err != nil {
		h.logger.Error("failed to update driver location", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *GeoHandler) FindNearbyDrivers(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.DefaultQuery("radius", "5")
	limitStr := c.DefaultQuery("limit", "20")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat value"})
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng value"})
		return
	}

	radius, _ := strconv.ParseFloat(radiusStr, 64)
	limit, _ := strconv.Atoi(limitStr)

	drivers, err := h.geoService.FindNearbyDrivers(c.Request.Context(), lat, lng, radius, limit)
	if err != nil {
		h.logger.Error("nearby driver search failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search nearby drivers"})
		return
	}

	c.JSON(http.StatusOK, drivers)
}
