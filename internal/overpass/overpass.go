package overpass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"time"

	"trail-finder-mcp/internal/config"
	"trail-finder-mcp/internal/models"
)

var httpClient = &http.Client{Timeout: 20 * time.Second}

type overpassResponse struct {
	Elements []struct {
		Type string            `json:"type"` // "node"
		ID   int64             `json:"id"`
		Lat  float64           `json:"lat"`
		Lon  float64           `json:"lon"`
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

	if len(parts) == 0 {
		return nil, fmt.Errorf("no query generated for include set")
	}

	query := fmt.Sprintf(`[out:json][timeout:25];(%s);out tags geom;`, joinParts(parts))
	reqBody := bytes.NewBufferString(query)
	req, err := http.NewRequestWithContext(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", config.UserAgent())

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
		Center:      models.Coord{Lat: in.Lat, Lon: in.Lon},
		RadiusM:     in.RadiusM,
		Items:       make([]models.POIItem, 0, len(op.Elements)),
		Attribution: "© OpenStreetMap contributors (ODbL) via Overpass API",
	}
	for _, el := range op.Elements {
		p := models.POIItem{
			ID:        fmt.Sprintf("%s/%d", el.Type, el.ID),
			Type:      classify(el.Tags),
			Tags:      el.Tags,
			Location:  models.Coord{Lat: el.Lat, Lon: el.Lon},
			Source:    "overpass/osm",
			DistanceM: distanceMeters(in.Lat, in.Lon, el.Lat, el.Lon),
		}
		if name, ok := el.Tags["name"]; ok {
			p.Name = name
		}
		resp.Items = append(resp.Items, p)
	}
	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].DistanceM < resp.Items[j].DistanceM
	})
	if limit := in.Limit; limit > 0 && len(resp.Items) > limit {
		resp.Items = resp.Items[:limit]
	}
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

func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0 // meters
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func toRadians(v float64) float64 {
	return v * math.Pi / 180
}
