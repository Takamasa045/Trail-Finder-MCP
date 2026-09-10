package elevation

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/geo"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/httpx"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

var requestCount atomic.Int64

func RequestCount() int64 { return requestCount.Load() }

func Lookup(ctx context.Context, in models.ElevationInput) (*models.ElevationResponse, error) {
	provider := strings.ToLower(config.Env("ELEVATION_PROVIDER", "open-meteo"))
	if provider == "off" || provider == "none" {
		return nil, fmt.Errorf("elevation provider disabled")
	}
	vals, source, err := lookupMany(ctx, provider, []geo.Point{{Lat: in.Lat, Lon: in.Lon}})
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, fmt.Errorf("no elevation result")
	}
	return (&models.ElevationResponse{ElevationM: vals[0], Source: source}).WithMeta(), nil
}

func LookupMany(ctx context.Context, points []geo.Point) ([]float64, string, error) {
	provider := strings.ToLower(config.Env("ELEVATION_PROVIDER", "open-meteo"))
	if provider == "off" || provider == "none" {
		return nil, "", fmt.Errorf("elevation provider disabled")
	}
	if len(points) == 0 {
		return nil, "", fmt.Errorf("no points")
	}
	return lookupMany(ctx, provider, points)
}

func lookupMany(ctx context.Context, provider string, points []geo.Point) ([]float64, string, error) {
	requestCount.Add(1)
	switch provider {
	case "open-elevation":
		return openElevation(ctx, points)
	case "open-topo", "open-topodata", "opentopo", "opentopodata":
		return openTopo(ctx, points)
	default:
		return openMeteo(ctx, points)
	}
}

func joinLocations(points []geo.Point) string {
	parts := make([]string, len(points))
	for i, p := range points {
		parts[i] = fmt.Sprintf("%f,%f", p.Lat, p.Lon)
	}
	return strings.Join(parts, "|")
}

func openMeteo(ctx context.Context, points []geo.Point) ([]float64, string, error) {
	lats := make([]string, len(points))
	lons := make([]string, len(points))
	for i, p := range points {
		lats[i] = fmt.Sprintf("%f", p.Lat)
		lons[i] = fmt.Sprintf("%f", p.Lon)
	}
	base := config.Env("OPENMETEO_ELEVATION_URL", "https://api.open-meteo.com/v1/elevation")
	q := url.Values{}
	q.Set("latitude", strings.Join(lats, ","))
	q.Set("longitude", strings.Join(lons, ","))
	var out struct {
		Elevation []float64 `json:"elevation"`
	}
	if err := httpx.GetJSON(ctx, base+"?"+q.Encode(), &out); err != nil {
		return nil, "", fmt.Errorf("open-meteo elevation: %w", err)
	}
	if len(out.Elevation) != len(points) {
		return nil, "", fmt.Errorf("open-meteo elevation: got %d values, want %d", len(out.Elevation), len(points))
	}
	return out.Elevation, "open-meteo", nil
}

func openElevation(ctx context.Context, points []geo.Point) ([]float64, string, error) {
	v := url.Values{}
	v.Set("locations", joinLocations(points))
	endpoint := "https://api.open-elevation.com/api/v1/lookup?" + v.Encode()
	if base := config.Env("OPEN_ELEVATION_URL", ""); base != "" {
		endpoint = strings.TrimRight(base, "/") + "/api/v1/lookup?" + v.Encode()
	}
	var out struct {
		Results []struct {
			Elevation float64 `json:"elevation"`
		} `json:"results"`
	}
	if err := httpx.GetJSON(ctx, endpoint, &out); err != nil {
		return nil, "", fmt.Errorf("open-elevation: %w", err)
	}
	if len(out.Results) != len(points) {
		return nil, "", fmt.Errorf("open-elevation: got %d values, want %d", len(out.Results), len(points))
	}
	vals := make([]float64, len(out.Results))
	for i, r := range out.Results {
		vals[i] = r.Elevation
	}
	return vals, "open-elevation", nil
}

func openTopo(ctx context.Context, points []geo.Point) ([]float64, string, error) {
	base := config.Env("OPEN_TOPO_URL", "https://api.opentopodata.org/v1/srtm90m")
	v := url.Values{}
	v.Set("locations", joinLocations(points))
	var out struct {
		Results []struct {
			Elevation float64 `json:"elevation"`
		} `json:"results"`
	}
	if err := httpx.GetJSON(ctx, base+"?"+v.Encode(), &out); err != nil {
		return nil, "", fmt.Errorf("open-topodata: %w", err)
	}
	if len(out.Results) != len(points) {
		return nil, "", fmt.Errorf("open-topodata: got %d values, want %d", len(out.Results), len(points))
	}
	vals := make([]float64, len(out.Results))
	for i, r := range out.Results {
		vals[i] = r.Elevation
	}
	return vals, "open-topodata", nil
}
