package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/shipment/internal/domain/delivery"
	"go.uber.org/zap"
)

const (
	driverGeoKey      = "drivers:locations"
	nominatimCacheKey = "nominatim:search:"
	cacheTTL          = 24 * time.Hour
)

type Config struct {
	NominatimBaseURL   string
	NominatimUserAgent string
	NominatimTimeout   time.Duration
	OSRMBaseURL        string
	OSRMTimeout        time.Duration
}

type Service struct {
	redis     *redis.Client
	httpCli   *http.Client
	cfg       Config
	logger    *zap.Logger
}

func NewService(redisClient *redis.Client, cfg Config, logger *zap.Logger) *Service {
	return &Service{
		redis: redisClient,
		httpCli: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		cfg:    cfg,
		logger: logger.Named("geo_service"),
	}
}

func (s *Service) SearchAddress(ctx context.Context, query string) ([]delivery.GeoSearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	cacheKey := nominatimCacheKey + url.QueryEscape(query)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var results []delivery.GeoSearchResult
		if json.Unmarshal([]byte(cached), &results) == nil {
			return results, nil
		}
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("addressdetails", "1")
	params.Set("limit", "10")
	params.Set("countrycodes", "vn")

	apiURL := fmt.Sprintf("%s/search?%s", s.cfg.NominatimBaseURL, params.Encode())

	reqCtx, cancel := context.WithTimeout(ctx, s.cfg.NominatimTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.cfg.NominatimUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nominatim search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("nominatim returned %d: %s", resp.StatusCode, string(body))
	}

	var raw []struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode nominatim response: %w", err)
	}

	results := make([]delivery.GeoSearchResult, 0, len(raw))
	for _, r := range raw {
		var lat, lng float64
		fmt.Sscanf(r.Lat, "%f", &lat)
		fmt.Sscanf(r.Lon, "%f", &lng)
		if lat == 0 && lng == 0 {
			continue
		}
		results = append(results, delivery.GeoSearchResult{
			Address: r.DisplayName,
			Name:    r.Name,
			Lat:     lat,
			Lng:     lng,
		})
	}

	if data, err := json.Marshal(results); err == nil {
		s.redis.Set(ctx, cacheKey, data, cacheTTL)
	}

	return results, nil
}

func (s *Service) ReverseGeocode(ctx context.Context, lat, lng float64) (*delivery.ReverseGeocodeResult, error) {
	params := url.Values{}
	params.Set("lat", fmt.Sprintf("%.6f", lat))
	params.Set("lon", fmt.Sprintf("%.6f", lng))
	params.Set("format", "json")
	params.Set("addressdetails", "1")
	params.Set("zoom", "18")

	apiURL := fmt.Sprintf("%s/reverse?%s", s.cfg.NominatimBaseURL, params.Encode())

	reqCtx, cancel := context.WithTimeout(ctx, s.cfg.NominatimTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.cfg.NominatimUserAgent)

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nominatim reverse failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim returned %d", resp.StatusCode)
	}

	var raw struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
		Address     struct {
			Road     string `json:"road"`
			City     string `json:"city"`
			District string `json:"district"`
			Suburb   string `json:"suburb"`
			Village  string `json:"village"`
			Town     string `json:"town"`
			County   string `json:"county"`
			Country  string `json:"country"`
		} `json:"address"`
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode nominatim response: %w", err)
	}

	rLat, rLng := 0.0, 0.0
	fmt.Sscanf(raw.Lat, "%f", &rLat)
	fmt.Sscanf(raw.Lon, "%f", &rLng)

	result := &delivery.ReverseGeocodeResult{
		Address: raw.DisplayName,
		Name:    raw.Name,
		Street:  raw.Address.Road,
		Country: raw.Address.Country,
		Lat:     rLat,
		Lng:     rLng,
	}
	if raw.Address.City != "" {
		result.City = raw.Address.City
	} else if raw.Address.Town != "" {
		result.City = raw.Address.Town
	}
	if raw.Address.District != "" {
		result.District = raw.Address.District
	}
	if raw.Address.Suburb != "" {
		result.Ward = raw.Address.Suburb
	}

	return result, nil
}

func (s *Service) CalculateRoute(ctx context.Context, pickupLat, pickupLng, dropoffLat, dropoffLng float64) (*delivery.RouteResult, error) {
	coords := fmt.Sprintf("%.6f,%.6f;%.6f,%.6f", pickupLng, pickupLat, dropoffLng, dropoffLat)
	apiURL := fmt.Sprintf("%s/route/v1/driving/%s?overview=full&geometries=polyline&steps=false",
		s.cfg.OSRMBaseURL, coords)

	reqCtx, cancel := context.WithTimeout(ctx, s.cfg.OSRMTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("osrm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("osrm returned %d: %s", resp.StatusCode, string(body))
	}

	var osrmResp struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry string  `json:"geometry"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return nil, fmt.Errorf("decode osrm response: %w", err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return nil, fmt.Errorf("no route found")
	}

	route := osrmResp.Routes[0]
	return &delivery.RouteResult{
		DistanceMeters:  int(route.Distance),
		DurationSeconds: int(route.Duration),
		Polyline:        route.Geometry,
	}, nil
}

// Driver location methods using Redis GEO

func (s *Service) UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64) error {
	return s.redis.GeoAdd(ctx, driverGeoKey, &redis.GeoLocation{
		Name:      driverID,
		Longitude: lng,
		Latitude:  lat,
	}).Err()
}

func (s *Service) FindNearbyDrivers(ctx context.Context, lat, lng float64, radiusKm float64, limit int) ([]delivery.NearbyDriver, error) {
	if radiusKm <= 0 {
		radiusKm = 5
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	members, err := s.redis.GeoSearch(ctx, driverGeoKey, &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     radiusKm,
		RadiusUnit: "km",
		Sort:       "ASC",
		Count:      limit,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("geosearch failed: %w", err)
	}

	drivers := make([]delivery.NearbyDriver, 0, len(members))
	for _, name := range members {
		positions, err := s.redis.GeoPos(ctx, driverGeoKey, name).Result()
		if err != nil || len(positions) == 0 || positions[0] == nil {
			continue
		}
		drivers = append(drivers, delivery.NearbyDriver{
			DriverID: name,
			Lat:      positions[0].Latitude,
			Lng:      positions[0].Longitude,
		})
	}
	return drivers, nil
}

func (s *Service) RemoveDriverLocation(ctx context.Context, driverID string) error {
	return s.redis.ZRem(ctx, driverGeoKey, driverID).Err()
}
