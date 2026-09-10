package geocode

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
	"net/http"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/httpx"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

var (
	resultsCache = cache.New(10*time.Minute, 64)
	rateMu       sync.Mutex
	lastCall     time.Time
	httpClient   = &httpx.Client{
		HTTP:    &http.Client{Timeout: 15 * time.Second},
		Retries: 1,
		Backoff: time.Second,
	}
)

func Search(ctx context.Context, in models.GeocodeInput) (*models.GeocodeResponse, error) {
	key := fmt.Sprintf("%s|%d|%s|%s", in.Query, in.Limit, in.CountryCodes, config.DefaultLang())
	var cached models.GeocodeResponse
	if resultsCache.GetJSON(key, &cached) {
		return cached.WithMeta(), nil
	}

	if err := respectRateLimit(ctx); err != nil {
		return nil, err
	}

	endpoint, err := url.Parse(config.Env("NOMINATIM_URL", "https://nominatim.openstreetmap.org/search"))
	if err != nil {
		return nil, err
	}
	q := endpoint.Query()
	q.Set("q", in.Query)
	q.Set("format", "jsonv2")
	q.Set("limit", strconv.Itoa(in.Limit))
	q.Set("accept-language", config.DefaultLang())
	if in.CountryCodes != "" {
		q.Set("countrycodes", in.CountryCodes)
	}
	endpoint.RawQuery = q.Encode()

	var raw []struct {
		DisplayName string  `json:"display_name"`
		Name        string  `json:"name"`
		Lat         string  `json:"lat"`
		Lon         string  `json:"lon"`
		OSMType     string  `json:"osm_type"`
		OSMId       int64   `json:"osm_id"`
		Importance  float64 `json:"importance"`
	}
	if err := httpClient.GetJSON(ctx, endpoint.String(), &raw); err != nil {
		return nil, fmt.Errorf("nominatim: %w", err)
	}

	resp := &models.GeocodeResponse{
		Query:       in.Query,
		Items:       make([]models.GeocodeItem, 0, len(raw)),
		Attribution: "© OpenStreetMap contributors via Nominatim",
	}
	for _, row := range raw {
		lat, err1 := strconv.ParseFloat(row.Lat, 64)
		lon, err2 := strconv.ParseFloat(row.Lon, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		resp.Items = append(resp.Items, models.GeocodeItem{
			Name:        row.Name,
			DisplayName: row.DisplayName,
			Location:    models.Coord{Lat: lat, Lon: lon},
			OSMType:     row.OSMType,
			OSMId:       row.OSMId,
			Importance:  row.Importance,
		})
	}
	resultsCache.SetJSON(key, resp)
	return resp.WithMeta(), nil
}

func First(ctx context.Context, query string) (models.GeocodeItem, error) {
	resp, err := Search(ctx, models.GeocodeInput{Query: query, Limit: 1})
	if err != nil {
		return models.GeocodeItem{}, err
	}
	if len(resp.Items) == 0 {
		return models.GeocodeItem{}, fmt.Errorf("no geocode result for %q", query)
	}
	return resp.Items[0], nil
}

func respectRateLimit(ctx context.Context) error {
	min := time.Second
	if v := config.Env("NOMINATIM_MIN_INTERVAL", "1s"); v != "" {
		if v == "0" {
			min = 0
		} else if d, err := time.ParseDuration(v); err == nil {
			min = d
		}
	}
	if min <= 0 {
		return nil
	}
	rateMu.Lock()
	defer rateMu.Unlock()
	wait := min - time.Since(lastCall)
	if wait > 0 {
		t := time.NewTimer(wait)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
	lastCall = time.Now()
	return nil
}
