package router

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/elevation"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/geo"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/httpx"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func RouteFoot(ctx context.Context, in models.RouteInput) (*models.RouteResponse, error) {
	engine := strings.ToLower(strings.TrimSpace(in.Engine))
	if engine == "" || engine == "auto" {
		engine = "osrm"
	}
	if engine == "valhalla" {
		return nil, fmt.Errorf("valhalla is not implemented; use engine=osrm or auto")
	}
	if engine != "osrm" {
		return nil, fmt.Errorf("engine not supported: %s", engine)
	}
	resp, err := routeOSRM(ctx, in)
	if err != nil {
		return nil, err
	}
	if optionBool(in.Options, "include_elevation", true) && len(resp.Geometry.Coordinates) >= 2 {
		enrichElevation(ctx, resp)
	}
	return resp.WithMeta(), nil
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
	base := strings.TrimRight(config.Env("OSRM_URL", "https://router.project-osrm.org"), "/")

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

	var out osrmResponse
	if err := httpx.GetJSON(ctx, endpoint, &out); err != nil {
		return nil, fmt.Errorf("osrm: %w", err)
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

func enrichElevation(ctx context.Context, resp *models.RouteResponse) {
	pts := geometryPoints(resp.Geometry.Coordinates)
	if len(pts) < 2 {
		return
	}
	samples := geo.SampleEvery(pts, 100, 40)
	vals, _, err := elevation.LookupMany(ctx, samples)
	if err != nil || len(vals) != len(samples) {
		return
	}
	gain, loss := geo.ElevationGainLoss(vals)
	resp.ElevationGainM = gain
	resp.ElevationLossM = loss
	resp.ElevationSampled = true
}

func geometryPoints(coords [][]float64) []geo.Point {
	out := make([]geo.Point, 0, len(coords))
	for _, c := range coords {
		if len(c) < 2 {
			continue
		}
		out = append(out, geo.Point{Lat: c[1], Lon: c[0]})
	}
	return out
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

func GeometryCoords(resp models.RouteResponse) []models.Coord {
	pts := geometryPoints(resp.Geometry.Coordinates)
	out := make([]models.Coord, len(pts))
	for i, p := range pts {
		out[i] = models.Coord{Lat: p.Lat, Lon: p.Lon}
	}
	return out
}
