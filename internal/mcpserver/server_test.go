package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func boolPtr(v bool) *bool { return &v }

func writeElev(w http.ResponseWriter, r *http.Request, start, end float64) {
	n := 1
	if s := r.URL.Query().Get("latitude"); s != "" {
		n = len(strings.Split(s, ","))
	}
	vals := make([]float64, n)
	for i := range vals {
		if n == 1 {
			vals[i] = start
			continue
		}
		vals[i] = start + (end-start)*float64(i)/float64(n-1)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"elevation": vals})
}

func TestHandleGeocodeAndValidation(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"display_name": "X", "name": "X", "lat": "35", "lon": "139", "osm_type": "node", "osm_id": 1,
		}})
	}))
	defer up.Close()
	t.Setenv("NOMINATIM_URL", up.URL)
	t.Setenv("NOMINATIM_MIN_INTERVAL", "0")

	if _, _, err := handleGeocode(context.Background(), nil, geocodeArgs{}); err == nil {
		t.Fatal("expected validation error")
	}
	_, resp, err := handleGeocode(context.Background(), nil, geocodeArgs{Query: "mcp-geocode-x", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("%+v", resp)
	}
}

func TestHandleCoreTools(t *testing.T) {
	over := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"elements": []any{}})
	}))
	defer over.Close()
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 1, "duration": 1,
				"geometry": map[string]any{"type": "LineString", "coordinates": [][]float64{{139, 35}, {139.1, 35.1}}},
				"legs":     []any{},
			}},
		})
	}))
	defer osrm.Close()
	elev := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeElev(w, r, 10, 10)
	}))
	defer elev.Close()
	wx := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"timezone": "Asia/Tokyo",
			"hourly":   map[string]any{"time": []string{"t"}, "temperature_2m": []float64{1}, "precipitation": []float64{0}, "precipitation_probability": []float64{0}, "wind_speed_10m": []float64{1}, "wind_gusts_10m": []float64{1}, "weather_code": []int{0}},
			"daily":    map[string]any{"sunrise": []string{"s"}, "sunset": []string{"e"}},
		})
	}))
	defer wx.Close()

	t.Setenv("TRAILFINDER_OVERPASS_URL", over.URL)
	t.Setenv("OSRM_URL", osrm.URL)
	t.Setenv("ELEVATION_PROVIDER", "open-meteo")
	t.Setenv("OPENMETEO_ELEVATION_URL", elev.URL)
	t.Setenv("OPENMETEO_URL", wx.URL)

	if _, _, err := handleTrailheads(context.Background(), nil, trailheadsArgs{Lat: 91, Lon: 0}); err == nil {
		t.Fatal("expected validation")
	}
	if _, _, err := handleTrailheads(context.Background(), nil, trailheadsArgs{Lat: 35, Lon: 139, RadiusM: 500, Limit: 10}); err != nil {
		t.Fatal(err)
	}

	_, route, err := handleRouteFoot(context.Background(), nil, routeArgs{
		From:    models.Coord{Lat: 35, Lon: 139},
		To:      models.Coord{Lat: 35.1, Lon: 139.1},
		Engine:  "osrm",
		Options: map[string]any{"include_elevation": false},
	})
	if err != nil {
		t.Fatal(err)
	}
	if route.DistanceM != 1 {
		t.Fatalf("%+v", route)
	}

	if _, _, err := handleElevation(context.Background(), nil, elevationArgs{Lat: 99, Lon: 0}); err == nil {
		t.Fatal("expected elevation validation")
	}
	_, el, err := handleElevation(context.Background(), nil, elevationArgs{Lat: 35, Lon: 139})
	if err != nil {
		t.Fatal(err)
	}
	if el.ElevationM != 10 {
		t.Fatalf("%+v", el)
	}

	_, fc, err := handleForecast(context.Background(), nil, forecastArgs{Lat: 35, Lon: 139, Hours: 1})
	if err != nil {
		t.Fatal(err)
	}
	if fc.Timezone != "Asia/Tokyo" {
		t.Fatalf("%+v", fc)
	}
}

func TestHandlePlanHike(t *testing.T) {
	nominatim := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"display_name": "A", "name": "A", "lat": "35", "lon": "139", "osm_type": "node", "osm_id": 1,
		}})
	}))
	defer nominatim.Close()
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 5, "duration": 5,
				"geometry": map[string]any{"type": "LineString", "coordinates": [][]float64{{139, 35}, {139.01, 35.01}}},
				"legs":     []any{},
			}},
		})
	}))
	defer osrm.Close()
	over := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"elements": []any{}})
	}))
	defer over.Close()
	elev := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeElev(w, r, 1, 2)
	}))
	defer elev.Close()
	wx := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"timezone": "Asia/Tokyo",
			"hourly":   map[string]any{"time": []string{"t"}, "temperature_2m": []float64{0}, "precipitation": []float64{0}, "precipitation_probability": []float64{0}, "wind_speed_10m": []float64{0}, "wind_gusts_10m": []float64{0}, "weather_code": []int{0}},
			"daily":    map[string]any{"sunrise": []string{"s"}, "sunset": []string{"e"}},
		})
	}))
	defer wx.Close()

	t.Setenv("NOMINATIM_URL", nominatim.URL)
	t.Setenv("NOMINATIM_MIN_INTERVAL", "0")
	t.Setenv("OSRM_URL", osrm.URL)
	t.Setenv("TRAILFINDER_OVERPASS_URL", over.URL)
	t.Setenv("ELEVATION_PROVIDER", "open-meteo")
	t.Setenv("OPENMETEO_ELEVATION_URL", elev.URL)
	t.Setenv("OPENMETEO_URL", wx.URL)

	if _, _, err := handlePlanHike(context.Background(), nil, planArgs{}); err == nil {
		t.Fatal("expected validation")
	}
	_, resp, err := handlePlanHike(context.Background(), nil, planArgs{
		From:      models.PlaceRef{Query: "start-mcp"},
		To:        models.PlaceRef{Query: "goal-mcp"},
		AlsoWater: boolPtr(true),
		Hours:     1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Disclaimer, "参考情報") {
		t.Fatalf("disclaimer=%q", resp.Disclaimer)
	}
}
