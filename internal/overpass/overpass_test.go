package overpass

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func TestBuildQueryJapanDefaults(t *testing.T) {
	q := BuildQuery(35.63, 139.26, 2000, nil, true)
	for _, want := range []string{
		"nwr(",
		"highway=trailhead",
		"information=guidepost",
		"amenity=drinking_water",
		"natural=spring",
	} {
		if !strings.Contains(q, want) {
			t.Fatalf("query missing %s:\n%s", want, q)
		}
	}
	if strings.Contains(q, "entrance=yes") {
		t.Fatal("default query should not include noisy entrance=yes")
	}
}

func TestBuildQueryShelterAndEntrance(t *testing.T) {
	q := BuildQuery(35.63, 139.26, 1000, []string{"shelter", "entrance"}, false)
	if !strings.Contains(q, "tourism=wilderness_hut") || !strings.Contains(q, "entrance=yes") {
		t.Fatalf("query=%s", q)
	}
	if strings.Contains(q, "highway=trailhead") {
		t.Fatal("explicit include should not add default trailhead")
	}
}

func TestClassify(t *testing.T) {
	cases := map[string]map[string]string{
		"trailhead":      {"highway": "trailhead"},
		"guidepost":      {"information": "guidepost"},
		"drinking_water": {"amenity": "drinking_water"},
		"shelter":        {"tourism": "alpine_hut"},
		"pass":           {"natural": "saddle"},
		"entrance":       {"entrance": "yes"},
	}
	for want, tags := range cases {
		if got := Classify(tags); got != want {
			t.Fatalf("Classify(%v)=%s want %s", tags, got, want)
		}
	}
}

func TestQueryPOIsParsesAndSorts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "data=") {
			http.Error(w, "expected form body", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"elements": []map[string]any{
				{"type": "node", "id": 2, "lat": 35.64, "lon": 139.28, "tags": map[string]string{"highway": "trailhead", "name": "遠い登山口"}},
				{"type": "node", "id": 1, "lat": 35.631, "lon": 139.270, "tags": map[string]string{"information": "guidepost", "name": "近い道標"}},
			},
		})
	}))
	defer ts.Close()

	t.Setenv("TRAILFINDER_OVERPASS_URL", ts.URL)

	resp, err := QueryPOIs(context.Background(), models.TrailheadsInput{
		Lat: 35.631, Lon: 139.270, RadiusM: 2000, AlsoWater: true, Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items=%d", len(resp.Items))
	}
	if resp.Items[0].Name != "近い道標" {
		t.Fatalf("expected nearest first, got %q", resp.Items[0].Name)
	}
	if resp.Items[0].Type != "guidepost" || resp.Items[1].Type != "trailhead" {
		t.Fatalf("types %s %s", resp.Items[0].Type, resp.Items[1].Type)
	}
	if resp.Disclaimer == "" {
		t.Fatal("missing disclaimer")
	}
}

func TestQueryAlongRoute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"elements": []map[string]any{
				{"type": "node", "id": 1, "lat": 35.63, "lon": 139.27, "tags": map[string]string{"highway": "trailhead", "name": "登山口"}},
				{"type": "node", "id": 1, "lat": 35.63, "lon": 139.27, "tags": map[string]string{"highway": "trailhead", "name": "登山口"}},
			},
		})
	}))
	defer ts.Close()
	t.Setenv("TRAILFINDER_OVERPASS_URL", ts.URL)

	resp, err := QueryAlongRoute(context.Background(), []models.Coord{
		{Lat: 35.63, Lon: 139.27},
		{Lat: 35.62, Lon: 139.25},
	}, 120, []string{"trailhead"}, true, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("dedupe items=%d", len(resp.Items))
	}
	if _, err := QueryAlongRoute(context.Background(), nil, 120, nil, false, 10); err == nil {
		t.Fatal("expected empty coords error")
	}
}

func TestQueryPOIsWayCenter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"elements": []map[string]any{
				{
					"type": "way", "id": 77,
					"center": map[string]any{"lat": 35.7, "lon": 137.8},
					"tags":   map[string]string{"tourism": "wilderness_hut", "name": "小屋"},
				},
			},
		})
	}))
	defer ts.Close()
	t.Setenv("TRAILFINDER_OVERPASS_URL", ts.URL)
	resp, err := QueryPOIs(context.Background(), models.TrailheadsInput{
		Lat: 35.7, Lon: 137.8, RadiusM: 1000, Include: []string{"shelter"}, Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Type != "shelter" {
		t.Fatalf("%+v", resp.Items)
	}
	if resp.Items[0].Location.Lat != 35.7 || resp.Items[0].ID != "way/77" {
		t.Fatalf("%+v", resp.Items[0])
	}
}
