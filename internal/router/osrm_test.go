package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func TestRouteOSRM(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/route/v1/foot/") {
			http.Error(w, "bad path "+r.URL.Path, 400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 1234.5,
				"duration": 890,
				"geometry": map[string]any{
					"type":        "LineString",
					"coordinates": [][]float64{{139.27, 35.63}, {139.25, 35.62}},
				},
				"legs": []any{},
			}},
		})
	}))
	defer ts.Close()
	t.Setenv("OSRM_URL", ts.URL)
	t.Setenv("ELEVATION_PROVIDER", "off")

	resp, err := RouteFoot(context.Background(), models.RouteInput{
		From:    models.Coord{Lat: 35.63, Lon: 139.27},
		To:      models.Coord{Lat: 35.62, Lon: 139.25},
		Engine:  "auto",
		Options: map[string]any{"include_elevation": false},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Engine != "osrm" || resp.DistanceM != 1234.5 {
		t.Fatalf("%+v", resp)
	}
	if len(resp.Geometry.Coordinates) != 2 {
		t.Fatalf("geometry=%v", resp.Geometry)
	}
}

func TestValhallaRejected(t *testing.T) {
	_, err := RouteFoot(context.Background(), models.RouteInput{
		From:   models.Coord{Lat: 1, Lon: 1},
		To:     models.Coord{Lat: 2, Lon: 2},
		Engine: "valhalla",
	})
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("err=%v", err)
	}
}

func TestOptionBool(t *testing.T) {
	if optionBool(nil, "x", true) != true {
		t.Fatal("nil default")
	}
	opts := map[string]any{"x": "false", "y": 1.0, "z": true}
	if optionBool(opts, "x", true) != false {
		t.Fatal("string false")
	}
	if optionBool(opts, "y", false) != true {
		t.Fatal("float")
	}
	if optionBool(opts, "z", false) != true {
		t.Fatal("bool")
	}
}

func TestRouteStepsAndElevation(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "steps=true") {
			http.Error(w, "expected steps", 400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "Ok",
			"routes": []map[string]any{{
				"distance": 100, "duration": 80,
				"geometry": map[string]any{
					"type":        "LineString",
					"coordinates": [][]float64{{139.27, 35.63}, {139.25, 35.62}},
				},
				"legs": []map[string]any{{
					"steps": []map[string]any{{"name": "trail"}},
				}},
			}},
		})
	}))
	defer osrm.Close()
	elev := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := 1
		if s := r.URL.Query().Get("latitude"); s != "" {
			n = len(strings.Split(s, ","))
		}
		vals := make([]float64, n)
		for i := range vals {
			if n == 1 {
				vals[i] = 100
				continue
			}
			vals[i] = 100 + 80*float64(i)/float64(n-1)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"elevation": vals})
	}))
	defer elev.Close()
	t.Setenv("OSRM_URL", osrm.URL)
	t.Setenv("ELEVATION_PROVIDER", "open-meteo")
	t.Setenv("OPENMETEO_ELEVATION_URL", elev.URL)

	resp, err := RouteFoot(context.Background(), models.RouteInput{
		From:    models.Coord{Lat: 35.63, Lon: 139.27},
		To:      models.Coord{Lat: 35.62, Lon: 139.25},
		Engine:  "osrm",
		Options: map[string]any{"include_steps": true, "include_elevation": true, "avoid_ferry": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Steps) != 1 {
		t.Fatalf("steps=%v", resp.Steps)
	}
	if !resp.ElevationSampled || resp.ElevationGainM < 70 {
		t.Fatalf("gain=%v sampled=%v", resp.ElevationGainM, resp.ElevationSampled)
	}
}

func TestGeometryCoords(t *testing.T) {
	resp := models.RouteResponse{Geometry: models.GeoJSONLineString{
		Coordinates: [][]float64{{139.0, 35.0}, {139.1, 35.1}},
	}}
	got := GeometryCoords(resp)
	if len(got) != 2 || got[0].Lat != 35.0 || got[0].Lon != 139.0 {
		t.Fatalf("%+v", got)
	}
}
