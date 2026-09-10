package elevation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/geo"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func TestLookupOpenMeteoBatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "latitude=") {
			http.Error(w, "missing latitude", 400)
			return
		}
		n := len(strings.Split(r.URL.Query().Get("latitude"), ","))
		vals := make([]float64, n)
		for i := range vals {
			if i == 0 {
				vals[i] = 199
			} else {
				vals[i] = 599
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"elevation": vals})
	}))
	defer ts.Close()

	t.Setenv("ELEVATION_PROVIDER", "open-meteo")
	t.Setenv("OPENMETEO_ELEVATION_URL", ts.URL)

	vals, src, err := LookupMany(context.Background(), []geo.Point{{Lat: 35.6, Lon: 139.7}, {Lat: 35.7, Lon: 139.8}})
	if err != nil {
		t.Fatal(err)
	}
	if src != "open-meteo" || len(vals) != 2 || vals[1] != 599 {
		t.Fatalf("vals=%v src=%s", vals, src)
	}

	resp, err := Lookup(context.Background(), models.ElevationInput{Lat: 35.6, Lon: 139.7})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ElevationM != 199 {
		t.Fatalf("single=%v", resp.ElevationM)
	}
}

func TestOpenElevation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{{"elevation": 12.0}},
		})
	}))
	defer ts.Close()
	t.Setenv("ELEVATION_PROVIDER", "open-elevation")
	t.Setenv("OPEN_ELEVATION_URL", ts.URL)

	resp, err := Lookup(context.Background(), models.ElevationInput{Lat: 35, Lon: 139})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Source != "open-elevation" || resp.ElevationM != 12 {
		t.Fatalf("%+v", resp)
	}
}

func TestLookupManyEmptyAndTopoLoop(t *testing.T) {
	if _, _, err := LookupMany(context.Background(), nil); err == nil {
		t.Fatal("expected no points")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := strings.Count(r.URL.Query().Get("locations"), "|") + 1
		results := make([]map[string]any, n)
		for i := range results {
			results[i] = map[string]any{"elevation": 7.0}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
	}))
	defer ts.Close()
	t.Setenv("ELEVATION_PROVIDER", "opentopodata")
	t.Setenv("OPEN_TOPO_URL", ts.URL)
	before := RequestCount()
	vals, src, err := LookupMany(context.Background(), []geo.Point{{Lat: 1, Lon: 1}, {Lat: 2, Lon: 2}, {Lat: 3, Lon: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if src != "open-topodata" || len(vals) != 3 || vals[0] != 7 {
		t.Fatalf("vals=%v src=%s", vals, src)
	}
	if RequestCount()-before != 1 {
		t.Fatalf("expected 1 HTTP batch, got %d", RequestCount()-before)
	}
}

func TestLookupDisabled(t *testing.T) {
	t.Setenv("ELEVATION_PROVIDER", "off")
	_, err := Lookup(context.Background(), models.ElevationInput{Lat: 1, Lon: 1})
	if err == nil {
		t.Fatal("expected disabled error")
	}
}

func TestOpenTopo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{{"elevation": 42.5}},
		})
	}))
	defer ts.Close()
	t.Setenv("ELEVATION_PROVIDER", "open-topo")
	t.Setenv("OPEN_TOPO_URL", ts.URL)

	resp, err := Lookup(context.Background(), models.ElevationInput{Lat: 35, Lon: 139})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Source != "open-topodata" || resp.ElevationM != 42.5 {
		t.Fatalf("%+v", resp)
	}
}
