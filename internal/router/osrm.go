package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"trail-finder-mcp/internal/config"
	"trail-finder-mcp/internal/models"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

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
			Type        string      `json:"type"`
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
		Legs []osrmLeg `json:"legs"`
	} `json:"routes"`
}

func routeOSRM(ctx context.Context, in models.RouteInput) (*models.RouteResponse, error) {
	base := os.Getenv("OSRM_URL")
	if base == "" {
		base = "https://router.project-osrm.org"
	}

	includeGeometry := optionBool(in.Options, "include_geometry", true)
	includeSteps := optionBool(in.Options, "include_steps", false)
	avoidFerry := optionBool(in.Options, "avoid_ferry", false)

	coords := fmt.Sprintf("%f,%f;%f,%f", in.From.Lon, in.From.Lat, in.To.Lon, in.To.Lat)
	q := url.Values{}
	q.Set("alternatives", "false")
	if includeGeometry {
		q.Set("overview", "full")
		q.Set("geometries", "geojson")
	} else {
		q.Set("overview", "false")
	}
	if includeSteps {
		q.Set("steps", "true")
	}
	if avoidFerry {
		q.Set("exclude", "ferry")
	}
	endpoint := fmt.Sprintf("%s/route/v1/foot/%s?%s", base, coords, q.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", config.UserAgent())

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

	resp := &models.RouteResponse{
		Engine:    "osrm",
		DistanceM: r.Distance,
		DurationS: r.Duration,
	}
	if includeGeometry {
		resp.Geometry = models.GeoJSONLineString{
			Type:        r.Geometry.Type,
			Coordinates: r.Geometry.Coordinates,
		}
	}
	if includeSteps {
		resp.Steps = collectSteps(r.Legs)
	}
	return resp, nil
}

func optionBool(options map[string]any, key string, defaultVal bool) bool {
	if options == nil {
		return defaultVal
	}
	raw, ok := options[key]
	if !ok {
		return defaultVal
	}
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	case float64:
		return v != 0
	}
	return defaultVal
}

type osrmLeg struct {
	Steps []map[string]any `json:"steps"`
}

func collectSteps(legs []osrmLeg) []any {
	var out []any
	for _, leg := range legs {
		for _, step := range leg.Steps {
			out = append(out, step)
		}
	}
	return out
}
