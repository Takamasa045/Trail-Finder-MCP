package plan

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

func writeElevations(w http.ResponseWriter, r *http.Request, start, end float64) {
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

func TestPlanHikeFromPlaceNames(t *testing.T) {
	nominatim := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		lat, lon, name := "35.6319", "139.2694", "高尾山口駅"
		if strings.Contains(q, "高尾山") && !strings.Contains(q, "山口") {
			lat, lon, name = "35.6254", "139.2435", "高尾山"
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"display_name": name + ", 八王子市",
			"name":         name,
			"lat":          lat,
			"lon":          lon,
			"osm_type":     "node",
			"osm_id":       1,
		}})
	}))
	defer nominatim.Close()

	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 4200,
				"duration": 3600,
				"geometry": map[string]any{
					"type":        "LineString",
					"coordinates": [][]float64{{139.2694, 35.6319}, {139.2435, 35.6254}},
				},
				"legs": []any{},
			}},
		})
	}))
	defer osrm.Close()

	overpass := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"elements": []map[string]any{
				{"type": "node", "id": 9, "lat": 35.628, "lon": 139.255, "tags": map[string]string{"amenity": "drinking_water", "name": "水場"}},
			},
		})
	}))
	defer overpass.Close()

	elev := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeElevations(w, r, 200, 599)
	}))
	defer elev.Close()

	forecast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"timezone": "Asia/Tokyo",
			"hourly": map[string]any{
				"time":                      []string{"2026-09-10T08:00"},
				"temperature_2m":            []float64{18.0},
				"precipitation":             []float64{0},
				"precipitation_probability": []float64{5},
				"wind_speed_10m":            []float64{1.2},
				"wind_gusts_10m":            []float64{3.0},
				"weather_code":              []int{1},
			},
			"daily": map[string]any{
				"sunrise": []string{"2026-09-10T05:30"},
				"sunset":  []string{"2026-09-10T18:00"},
			},
		})
	}))
	defer forecast.Close()

	t.Setenv("NOMINATIM_URL", nominatim.URL)
	t.Setenv("NOMINATIM_MIN_INTERVAL", "0")
	t.Setenv("OSRM_URL", osrm.URL)
	t.Setenv("TRAILFINDER_OVERPASS_URL", overpass.URL)
	t.Setenv("ELEVATION_PROVIDER", "open-meteo")
	t.Setenv("OPENMETEO_ELEVATION_URL", elev.URL)
	t.Setenv("OPENMETEO_URL", forecast.URL)

	resp, err := Plan(context.Background(), models.PlanHikeInput{
		From:      models.PlaceRef{Query: "高尾山口駅-plan"},
		To:        models.PlaceRef{Query: "高尾山-plan"},
		AlsoWater: boolPtr(true),
		Hours:     1,
		CorridorM: 150,
		Limit:     20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.From.Name != "高尾山口駅" || resp.To.Name != "高尾山" {
		t.Fatalf("places from=%s to=%s", resp.From.Name, resp.To.Name)
	}
	if resp.Route.DistanceM != 4200 {
		t.Fatalf("distance=%v", resp.Route.DistanceM)
	}
	if !resp.Route.ElevationSampled || resp.Route.ElevationGainM < 300 {
		t.Fatalf("gain=%v sampled=%v", resp.Route.ElevationGainM, resp.Route.ElevationSampled)
	}
	if len(resp.POIs) != 1 || resp.POIs[0].Type != "drinking_water" {
		t.Fatalf("pois=%+v", resp.POIs)
	}
	if resp.Forecast.Hourly[0].TemperatureC != 18 {
		t.Fatalf("forecast=%+v", resp.Forecast)
	}
	if resp.Disclaimer == "" {
		t.Fatal("missing disclaimer")
	}
}

func TestPlanHikeNoGeometryFallsBackToEndpoints(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 10, "duration": 10,
				"geometry": map[string]any{"type": "LineString", "coordinates": [][]float64{}},
				"legs":     []any{},
			}},
		})
	}))
	defer osrm.Close()
	overpass := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"elements": []any{}})
	}))
	defer overpass.Close()
	forecast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"timezone": "Asia/Tokyo",
			"hourly":   map[string]any{"time": []string{"t"}, "temperature_2m": []float64{0}, "precipitation": []float64{0}, "precipitation_probability": []float64{0}, "wind_speed_10m": []float64{0}, "wind_gusts_10m": []float64{0}, "weather_code": []int{0}},
			"daily":    map[string]any{"sunrise": []string{"s"}, "sunset": []string{"e"}},
		})
	}))
	defer forecast.Close()
	t.Setenv("OSRM_URL", osrm.URL)
	t.Setenv("TRAILFINDER_OVERPASS_URL", overpass.URL)
	t.Setenv("ELEVATION_PROVIDER", "off")
	t.Setenv("OPENMETEO_URL", forecast.URL)

	fromLat, fromLon := 35.0, 139.0
	toLat, toLon := 35.1, 139.1
	resp, err := Plan(context.Background(), models.PlanHikeInput{
		From: models.PlaceRef{Lat: &fromLat, Lon: &fromLon},
		To:   models.PlaceRef{Lat: &toLat, Lon: &toLon},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Route.DistanceM != 10 {
		t.Fatalf("%+v", resp.Route)
	}
}

func TestPlanHikeOverpassFailureStillReturnsRoute(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 99, "duration": 10,
				"geometry": map[string]any{"type": "LineString", "coordinates": [][]float64{{139.0, 35.0}, {139.1, 35.1}}},
				"legs":     []any{},
			}},
		})
	}))
	defer osrm.Close()
	overpass := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "busy", http.StatusServiceUnavailable)
	}))
	defer overpass.Close()
	forecast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"timezone": "Asia/Tokyo",
			"hourly":   map[string]any{"time": []string{"t"}, "temperature_2m": []float64{0}, "precipitation": []float64{0}, "precipitation_probability": []float64{0}, "wind_speed_10m": []float64{0}, "wind_gusts_10m": []float64{0}, "weather_code": []int{0}},
			"daily":    map[string]any{"sunrise": []string{"s"}, "sunset": []string{"e"}},
		})
	}))
	defer forecast.Close()
	t.Setenv("OSRM_URL", osrm.URL)
	t.Setenv("TRAILFINDER_OVERPASS_URL", overpass.URL)
	t.Setenv("ELEVATION_PROVIDER", "off")
	t.Setenv("OPENMETEO_URL", forecast.URL)

	fromLat, fromLon := 35.0, 139.0
	toLat, toLon := 35.1, 139.1
	resp, err := Plan(context.Background(), models.PlanHikeInput{
		From: models.PlaceRef{Lat: &fromLat, Lon: &fromLon},
		To:   models.PlaceRef{Lat: &toLat, Lon: &toLon},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Route.DistanceM != 99 {
		t.Fatalf("route missing: %+v", resp.Route)
	}
	if len(resp.Warnings) == 0 {
		t.Fatal("expected overpass warning")
	}
}

func TestPlanHikeCoordsSkipGeocode(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 10,
				"duration": 10,
				"geometry": map[string]any{"type": "LineString", "coordinates": [][]float64{{139.0, 35.0}, {139.1, 35.1}}},
				"legs":     []any{},
			}},
		})
	}))
	defer osrm.Close()
	overpass := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"elements": []any{}})
	}))
	defer overpass.Close()
	elev := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeElevations(w, r, 1, 2)
	}))
	defer elev.Close()
	forecast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"timezone": "Asia/Tokyo",
			"hourly":   map[string]any{"time": []string{"t"}, "temperature_2m": []float64{0}, "precipitation": []float64{0}, "precipitation_probability": []float64{0}, "wind_speed_10m": []float64{0}, "wind_gusts_10m": []float64{0}, "weather_code": []int{0}},
			"daily":    map[string]any{"sunrise": []string{"s"}, "sunset": []string{"e"}},
		})
	}))
	defer forecast.Close()

	t.Setenv("OSRM_URL", osrm.URL)
	t.Setenv("TRAILFINDER_OVERPASS_URL", overpass.URL)
	t.Setenv("ELEVATION_PROVIDER", "open-meteo")
	t.Setenv("OPENMETEO_ELEVATION_URL", elev.URL)
	t.Setenv("OPENMETEO_URL", forecast.URL)

	fromLat, fromLon := 35.0, 139.0
	toLat, toLon := 35.1, 139.1
	resp, err := Plan(context.Background(), models.PlanHikeInput{
		From: models.PlaceRef{Lat: &fromLat, Lon: &fromLon},
		To:   models.PlaceRef{Lat: &toLat, Lon: &toLon},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.From.Source != "coordinates" || resp.To.Source != "coordinates" {
		t.Fatalf("sources %s %s", resp.From.Source, resp.To.Source)
	}
}
