package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"trail-finder-mcp/internal/models"
)

var httpClient = &http.Client{ Timeout: 20 * time.Second }

func RouteFoot(ctx context.Context, in models.RouteInput) (*models.RouteResponse, error) {
	engine := "osrm"
	if in.Engine == "valhalla" && os.Getenv("VALHALLA_URL") != "" {
		// TODO: implement valhalla client; fallback to OSRM for now
		engine = "osrm"
	} else if in.Engine == "osrm" || in.Engine == "auto" {
		engine = "osrm"
	}

	switch engine {
	case "osrm":
		return routeOSRM(ctx, in)
	default:
		return nil, fmt.Errorf("engine not supported yet: %s", engine)
	}
}

type osrmResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Type        string        `json:"type"`
			Coordinates [][]float64   `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

func routeOSRM(ctx context.Context, in models.RouteInput) (*models.RouteResponse, error) {
	base := os.Getenv("OSRM_URL")
	if base == "" {
		base = "https://router.project-osrm.org"
	}
	coords := fmt.Sprintf("%f,%f;%f,%f", in.From.Lon, in.From.Lat, in.To.Lon, in.To.Lat)
	q := url.Values{}
	q.Set("alternatives", "false")
	q.Set("overview", "full")
	q.Set("geometries", "geojson")
	endpoint := fmt.Sprintf("%s/route/v1/foot/%s?%s", base, coords, q.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "trail-finder-mcp/0.1")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("osrm status=%d body=%s", res.StatusCode, string(b))
	}

	var out osrmResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Code != "Ok" || len(out.Routes) == 0 {
		return nil, fmt.Errorf("osrm route not found")
	}
	r := out.Routes[0]
	return &models.RouteResponse{
		Engine:    "osrm",
		DistanceM: r.Distance,
		DurationS: r.Duration,
		Geometry: models.GeoJSONLineString{
			Type:        r.Geometry.Type,
			Coordinates: r.Geometry.Coordinates,
		},
	}, nil
}
