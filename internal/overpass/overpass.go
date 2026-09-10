package overpass

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/geo"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/httpx"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

var resultsCache = cache.New(5*time.Minute, 64)

type overpassResponse struct {
	Elements []overpassElement `json:"elements"`
}

type overpassElement struct {
	Type   string  `json:"type"`
	ID     int64   `json:"id"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Center *struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"center"`
	Tags map[string]string `json:"tags"`
}

func (el overpassElement) coord() (lat, lon float64, ok bool) {
	if el.Type != "node" && el.Center != nil {
		return el.Center.Lat, el.Center.Lon, true
	}
	if el.Lat != 0 || el.Lon != 0 {
		return el.Lat, el.Lon, true
	}
	if el.Center != nil {
		return el.Center.Lat, el.Center.Lon, true
	}
	if el.Type == "node" {
		return el.Lat, el.Lon, true
	}
	return 0, 0, false
}

func QueryPOIs(ctx context.Context, in models.TrailheadsInput) (*models.TrailheadsResponse, error) {
	query := BuildQuery(in.Lat, in.Lon, in.RadiusM, in.Include, in.AlsoWater)
	if query == "" {
		return nil, fmt.Errorf("no query generated for include set")
	}
	items, err := runQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	center := geo.Point{Lat: in.Lat, Lon: in.Lon}
	resp := assemble(center, in.RadiusM, items, []geo.Point{center}, in.Limit)
	return resp.WithMeta(), nil
}

func QueryAlongRoute(ctx context.Context, coords []models.Coord, radiusM int, include []string, alsoWater bool, limit int) (*models.TrailheadsResponse, error) {
	if len(coords) == 0 {
		return nil, fmt.Errorf("route coordinates are required")
	}
	if radiusM <= 0 {
		radiusM = 150
	}
	points := make([]geo.Point, 0, len(coords))
	for _, c := range coords {
		points = append(points, geo.Point{Lat: c.Lat, Lon: c.Lon})
	}
	samples := geo.SampleEvery(points, 100, 48)
	parts := filterClauses(polylineAround(radiusM, samples), include, alsoWater)
	if len(parts) == 0 {
		return nil, fmt.Errorf("no query generated for include set")
	}
	query := wrapQuery(parts)
	items, err := runQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	center := samples[0]
	resp := assemble(center, radiusM, items, samples, limit)
	return resp.WithMeta(), nil
}

func BuildQuery(lat, lon float64, radiusM int, include []string, alsoWater bool) string {
	parts := nodeClauses(lat, lon, radiusM, include, alsoWater)
	if len(parts) == 0 {
		return ""
	}
	return wrapQuery(parts)
}

func wrapQuery(parts []string) string {
	return fmt.Sprintf("[out:json][timeout:20];(%s);out tags center;", strings.Join(parts, ""))
}

func pointAround(radiusM int, lat, lon float64) string {
	return fmt.Sprintf("around:%d,%.7f,%.7f", radiusM, lat, lon)
}

func polylineAround(radiusM int, samples []geo.Point) string {
	var b strings.Builder
	fmt.Fprintf(&b, "around:%d", radiusM)
	for _, p := range samples {
		fmt.Fprintf(&b, ",%.7f,%.7f", p.Lat, p.Lon)
	}
	return b.String()
}

func nodeClauses(lat, lon float64, radiusM int, include []string, alsoWater bool) []string {
	return filterClauses(pointAround(radiusM, lat, lon), include, alsoWater)
}

func filterClauses(around string, include []string, alsoWater bool) []string {
	want := map[string]bool{}
	for _, s := range include {
		want[s] = true
	}
	defaultSet := len(want) == 0
	var parts []string
	nwr := func(filter string) {
		parts = append(parts, fmt.Sprintf("nwr(%s)%s;", around, filter))
	}
	if defaultSet || want["guidepost"] {
		nwr("[information=guidepost]")
		nwr("[tourism=information][information=guidepost]")
	}
	if defaultSet || want["trailhead"] {
		nwr("[highway=trailhead]")
		nwr("[information=trailhead]")
	}
	if want["entrance"] {
		nwr("[entrance=yes]")
	}
	if want["shelter"] {
		nwr("[tourism=wilderness_hut]")
		nwr("[tourism=alpine_hut]")
		nwr("[amenity=shelter]")
	}
	if want["pass"] {
		nwr("[mountain_pass=yes]")
		nwr("[natural=saddle]")
	}
	if alsoWater {
		nwr("[amenity=drinking_water]")
		nwr("[natural=spring]")
	}
	return parts
}

func runQuery(ctx context.Context, query string) ([]models.POIItem, error) {
	var cached []models.POIItem
	if resultsCache.GetJSON(query, &cached) {
		return cached, nil
	}

	endpoint := config.Env("TRAILFINDER_OVERPASS_URL", "https://overpass-api.de/api/interpreter")
	form := url.Values{}
	form.Set("data", query)
	var op overpassResponse
	if err := httpx.PostFormJSON(ctx, endpoint, []byte(form.Encode()), &op); err != nil {
		return nil, fmt.Errorf("overpass: %w", err)
	}

	items := make([]models.POIItem, 0, len(op.Elements))
	for _, el := range op.Elements {
		lat, lon, ok := el.coord()
		if !ok {
			continue
		}
		p := models.POIItem{
			ID:       fmt.Sprintf("%s/%d", el.Type, el.ID),
			Type:     Classify(el.Tags),
			Tags:     el.Tags,
			Location: models.Coord{Lat: lat, Lon: lon},
			Source:   "overpass/osm",
		}
		if name, ok := el.Tags["name"]; ok {
			p.Name = name
		} else if name, ok := el.Tags["name:ja"]; ok {
			p.Name = name
		}
		items = append(items, p)
	}
	resultsCache.SetJSON(query, items)
	return items, nil
}

func assemble(center geo.Point, radiusM int, items []models.POIItem, distanceFrom []geo.Point, limit int) *models.TrailheadsResponse {
	seen := map[string]bool{}
	out := make([]models.POIItem, 0, len(items))
	for _, p := range items {
		if seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		p.DistanceM = geo.MinDistanceMeters(p.Location.Lat, p.Location.Lon, distanceFrom)
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].DistanceM < out[j].DistanceM
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return &models.TrailheadsResponse{
		Center:      models.Coord{Lat: center.Lat, Lon: center.Lon},
		RadiusM:     radiusM,
		Items:       out,
		Attribution: "© OpenStreetMap contributors (ODbL) via Overpass API",
	}
}

func Classify(tags map[string]string) string {
	if tags["information"] == "trailhead" || tags["highway"] == "trailhead" {
		return "trailhead"
	}
	if tags["information"] == "guidepost" {
		return "guidepost"
	}
	if tags["amenity"] == "drinking_water" {
		return "drinking_water"
	}
	if tags["natural"] == "spring" {
		return "spring"
	}
	if tags["tourism"] == "wilderness_hut" || tags["tourism"] == "alpine_hut" || tags["amenity"] == "shelter" {
		return "shelter"
	}
	if tags["mountain_pass"] == "yes" || tags["natural"] == "saddle" {
		return "pass"
	}
	if _, ok := tags["entrance"]; ok {
		return "entrance"
	}
	return "poi"
}
