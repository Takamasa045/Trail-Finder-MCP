package overpass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"trail-finder-mcp/internal/models"
)

var httpClient = &http.Client{ Timeout: 20 * time.Second }

type overpassResponse struct {
	Elements []struct {
		Type string `json:"type"` // "node"
		ID   int64  `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Tags map[string]string `json:"tags"`
	} `json:"elements"`
}

func QueryPOIs(ctx context.Context, in models.TrailheadsInput) (*models.TrailheadsResponse, error) {
	url := os.Getenv("TRAILFINDER_OVERPASS_URL")
	if url == "" {
		url = "https://overpass-api.de/api/interpreter"
	}
	// Build query parts
	var parts []string
	include := map[string]bool{}
	for _, s := range in.Include {
		include[s] = true
	}
	if len(include) == 0 || include["guidepost"] {
		parts = append(parts, fmt.Sprintf("node(around:%d,%.7f,%.7f)[information=guidepost];", in.RadiusM, in.Lat, in.Lon))
	}
	if len(include) == 0 || include["trailhead"] {
		parts = append(parts, fmt.Sprintf("node(around:%d,%.7f,%.7f)[information=trailhead];", in.RadiusM, in.Lat, in.Lon))
		parts = append(parts, fmt.Sprintf("node(around:%d,%.7f,%.7f)[entrance=yes];", in.RadiusM, in.Lat, in.Lon))
	}
	if in.AlsoWater {
		parts = append(parts, fmt.Sprintf("node(around:%d,%.7f,%.7f)[amenity=drinking_water];", in.RadiusM, in.Lat, in.Lon))
		parts = append(parts, fmt.Sprintf("node(around:%d,%.7f,%.7f)[natural=spring];", in.RadiusM, in.Lat, in.Lon))
	}

	query := fmt.Sprintf(`[out:json][timeout:25];(%s);out tags geom;`, joinParts(parts))
	reqBody := bytes.NewBufferString(query)
	req, err := http.NewRequestWithContext(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "trail-finder-mcp/0.1 (+contact)")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("overpass status=%d body=%s", res.StatusCode, string(b))
	}
	var op overpassResponse
	if err := json.NewDecoder(res.Body).Decode(&op); err != nil {
		return nil, err
	}

	resp := &models.TrailheadsResponse{
		Center: models.Coord{Lat: in.Lat, Lon: in.Lon},
		RadiusM: in.RadiusM,
		Items: make([]models.POIItem, 0, len(op.Elements)),
		Attribution: "© OpenStreetMap contributors (ODbL) via Overpass API",
	}
	for _, el := range op.Elements {
		p := models.POIItem{
			ID: fmt.Sprintf("%s/%d", el.Type, el.ID),
			Type: classify(el.Tags),
			Tags: el.Tags,
			Location: models.Coord{Lat: el.Lat, Lon: el.Lon},
			Source: "overpass/osm",
		}
		if name, ok := el.Tags["name"]; ok {
			p.Name = name
		}
		resp.Items = append(resp.Items, p)
	}
	// Optional: trim and sort by rough distance
	if len(resp.Items) > in.Limit {
		resp.Items = resp.Items[:in.Limit]
	}
	sort.Slice(resp.Items, func(i, j int) bool {
		di := abs(resp.Items[i].Location.Lat - in.Lat) + abs(resp.Items[i].Location.Lon - in.Lon)
		dj := abs(resp.Items[j].Location.Lat - in.Lat) + abs(resp.Items[j].Location.Lon - in.Lon)
		return di < dj
	})
	return resp, nil
}

func classify(tags map[string]string) string {
	if v := tags["information"]; v == "trailhead" {
		return "trailhead"
	}
	if v := tags["information"]; v == "guidepost" {
		return "guidepost"
	}
	if v := tags["amenity"]; v == "drinking_water" {
		return "drinking_water"
	}
	if v := tags["natural"]; v == "spring" {
		return "spring"
	}
	if _, ok := tags["entrance"]; ok {
		return "trailhead"
	}
	return "poi"
}

func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var b bytes.Buffer
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
