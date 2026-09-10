package geocode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func TestSearchParsesNominatim(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "高尾山口駅" {
			http.Error(w, "bad q", 400)
			return
		}
		if r.URL.Query().Get("format") != "jsonv2" {
			http.Error(w, "bad format", 400)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"display_name": "高尾山口駅, 八王子市, 東京都",
			"name":         "高尾山口駅",
			"lat":          "35.6319",
			"lon":          "139.2694",
			"osm_type":     "node",
			"osm_id":       123,
			"importance":   0.7,
		}})
	}))
	defer ts.Close()

	t.Setenv("NOMINATIM_URL", ts.URL)
	t.Setenv("NOMINATIM_MIN_INTERVAL", "0")

	resp, err := Search(context.Background(), models.GeocodeInput{Query: "高尾山口駅", Limit: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Location.Lat != 35.6319 {
		t.Fatalf("%+v", resp.Items)
	}

	item, err := First(context.Background(), "高尾山口駅")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(item.DisplayName, "高尾山口") {
		t.Fatalf("%+v", item)
	}
}

func TestSearchCountryCodes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("countrycodes") != "jp" {
			http.Error(w, "missing countrycodes", 400)
			return
		}
		_, _ = w.Write([]byte(`[{"display_name":"Y","name":"Y","lat":"36","lon":"138","osm_type":"way","osm_id":2}]`))
	}))
	defer ts.Close()
	t.Setenv("NOMINATIM_URL", ts.URL)
	t.Setenv("NOMINATIM_MIN_INTERVAL", "1ms")

	resp, err := Search(context.Background(), models.GeocodeInput{Query: "country-test", Limit: 1, CountryCodes: "jp"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Name != "Y" {
		t.Fatalf("%+v", resp.Items)
	}
}

func TestFirstEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()
	t.Setenv("NOMINATIM_URL", ts.URL)
	t.Setenv("NOMINATIM_MIN_INTERVAL", "0")

	_, err := First(context.Background(), "nowhere-xyz")
	if err == nil {
		t.Fatal("expected error")
	}
}
