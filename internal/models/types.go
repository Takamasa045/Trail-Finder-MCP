package models

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
)

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Hint      string `json:"hint,omitempty"`
	Retryable bool   `json:"retryable"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type Coord struct {
	Lat float64 `json:"lat" description:"Latitude in decimal degrees"`
	Lon float64 `json:"lon" description:"Longitude in decimal degrees"`
}

func validLatLon(lat, lon float64) bool {
	return math.Abs(lat) <= 90 && math.Abs(lon) <= 180
}

// ---- Trailheads ----

type TrailheadsInput struct {
	Lat       float64  `json:"lat"`
	Lon       float64  `json:"lon"`
	RadiusM   int      `json:"radius_m,omitempty"`
	Include   []string `json:"include,omitempty"`
	AlsoWater bool     `json:"also_water,omitempty"`
	Limit     int      `json:"limit,omitempty"`
}

func (t *TrailheadsInput) Validate() error {
	if !validLatLon(t.Lat, t.Lon) {
		return errors.New("lat/lon out of range")
	}
	if t.RadiusM == 0 {
		t.RadiusM = 2000
	}
	if t.RadiusM < 50 || t.RadiusM > 20000 {
		return errors.New("radius_m must be 50..20000")
	}
	if t.Limit == 0 {
		t.Limit = 200
	}
	if t.Limit < 1 || t.Limit > 500 {
		return errors.New("limit must be 1..500")
	}
	for i, s := range t.Include {
		v := strings.ToLower(strings.TrimSpace(s))
		switch v {
		case "guidepost", "trailhead", "shelter", "pass", "entrance":
			t.Include[i] = v
		default:
			return fmt.Errorf("include supports guidepost, trailhead, shelter, pass, entrance (got %q)", s)
		}
	}
	return nil
}

type POIItem struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Name      string            `json:"name,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	Location  Coord             `json:"location"`
	DistanceM float64           `json:"distance_m,omitempty"`
	Source    string            `json:"source"`
}

type TrailheadsResponse struct {
	Center      Coord     `json:"center"`
	RadiusM     int       `json:"radius_m"`
	Items       []POIItem `json:"items"`
	Attribution string    `json:"attribution,omitempty"`
	Disclaimer  string    `json:"disclaimer,omitempty"`
}

func (r *TrailheadsResponse) WithMeta() *TrailheadsResponse {
	if r.Disclaimer == "" {
		r.Disclaimer = config.Disclaimer
	}
	return r
}

// ---- Route ----

type RouteInput struct {
	From    Coord          `json:"from"`
	To      Coord          `json:"to"`
	Engine  string         `json:"engine,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

func (in *RouteInput) Validate() error {
	if !validLatLon(in.From.Lat, in.From.Lon) || !validLatLon(in.To.Lat, in.To.Lon) {
		return errors.New("lat/lon out of range")
	}
	if in.Engine == "" {
		in.Engine = "auto"
	}
	in.Engine = strings.ToLower(strings.TrimSpace(in.Engine))
	switch in.Engine {
	case "auto", "osrm", "valhalla":
	default:
		return fmt.Errorf("engine must be auto, osrm, or valhalla (got %q)", in.Engine)
	}
	return nil
}

type GeoJSONLineString struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type RouteResponse struct {
	Engine           string            `json:"engine"`
	DistanceM        float64           `json:"distance_m"`
	DurationS        float64           `json:"duration_s"`
	Geometry         GeoJSONLineString `json:"geometry,omitempty"`
	Steps            []any             `json:"steps,omitempty"`
	ElevationGainM   float64           `json:"elevation_gain_m,omitempty"`
	ElevationLossM   float64           `json:"elevation_loss_m,omitempty"`
	ElevationSampled bool              `json:"elevation_sampled,omitempty"`
	Disclaimer       string            `json:"disclaimer,omitempty"`
}

func (r *RouteResponse) WithMeta() *RouteResponse {
	if r.Disclaimer == "" {
		r.Disclaimer = config.Disclaimer
	}
	return r
}

// ---- Elevation ----

type ElevationInput struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (e *ElevationInput) Validate() error {
	if !validLatLon(e.Lat, e.Lon) {
		return errors.New("lat/lon out of range")
	}
	return nil
}

type ElevationResponse struct {
	ElevationM float64 `json:"elevation_m"`
	Source     string  `json:"source"`
	Disclaimer string  `json:"disclaimer,omitempty"`
}

func (r *ElevationResponse) WithMeta() *ElevationResponse {
	if r.Disclaimer == "" {
		r.Disclaimer = config.Disclaimer
	}
	return r
}

// ---- Forecast ----

type ForecastInput struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Hours int     `json:"hours,omitempty"`
}

func (f *ForecastInput) Validate() error {
	if !validLatLon(f.Lat, f.Lon) {
		return errors.New("lat/lon out of range")
	}
	if f.Hours == 0 {
		f.Hours = 24
	}
	if f.Hours < 1 || f.Hours > 168 {
		return errors.New("hours must be 1..168")
	}
	return nil
}

type HourlyForecast struct {
	TimeISO              string  `json:"time"`
	TemperatureC         float64 `json:"temperature_c"`
	PrecipMM             float64 `json:"precip_mm"`
	PrecipProbabilityPct float64 `json:"precip_probability_pct"`
	WindMps              float64 `json:"wind_mps"`
	WindGustMps          float64 `json:"wind_gust_mps"`
	WeatherCode          int     `json:"weather_code"`
}

type ForecastResponse struct {
	Timezone    string           `json:"timezone"`
	SunriseISO  string           `json:"sunrise,omitempty"`
	SunsetISO   string           `json:"sunset,omitempty"`
	Hourly      []HourlyForecast `json:"hourly"`
	Attribution string           `json:"attribution,omitempty"`
	Disclaimer  string           `json:"disclaimer,omitempty"`
}

func (r *ForecastResponse) WithMeta() *ForecastResponse {
	if r.Disclaimer == "" {
		r.Disclaimer = config.Disclaimer
	}
	return r
}

// ---- Geocode ----

type GeocodeInput struct {
	Query        string `json:"query"`
	Limit        int    `json:"limit,omitempty"`
	CountryCodes string `json:"country_codes,omitempty"`
}

func (g *GeocodeInput) Validate() error {
	g.Query = strings.TrimSpace(g.Query)
	if g.Query == "" {
		return errors.New("query is required")
	}
	if g.Limit == 0 {
		g.Limit = 5
	}
	if g.Limit < 1 || g.Limit > 10 {
		return errors.New("limit must be 1..10")
	}
	return nil
}

type GeocodeItem struct {
	Name        string  `json:"name,omitempty"`
	DisplayName string  `json:"display_name"`
	Location    Coord   `json:"location"`
	OSMType     string  `json:"osm_type,omitempty"`
	OSMId       int64   `json:"osm_id,omitempty"`
	Importance  float64 `json:"importance,omitempty"`
}

type GeocodeResponse struct {
	Query       string        `json:"query"`
	Items       []GeocodeItem `json:"items"`
	Attribution string        `json:"attribution,omitempty"`
	Disclaimer  string        `json:"disclaimer,omitempty"`
}

func (r *GeocodeResponse) WithMeta() *GeocodeResponse {
	if r.Disclaimer == "" {
		r.Disclaimer = config.Disclaimer
	}
	return r
}

// ---- Place / plan ----

type PlaceRef struct {
	Query string   `json:"query,omitempty" description:"Place name such as 高尾山口駅 or Mt. Takao"`
	Lat   *float64 `json:"lat,omitempty"`
	Lon   *float64 `json:"lon,omitempty"`
}

func (p PlaceRef) HasQuery() bool {
	return strings.TrimSpace(p.Query) != ""
}

func (p PlaceRef) HasCoord() bool {
	return p.Lat != nil && p.Lon != nil
}

func (p PlaceRef) Coord() (Coord, error) {
	if !p.HasCoord() {
		return Coord{}, errors.New("lat/lon required")
	}
	if !validLatLon(*p.Lat, *p.Lon) {
		return Coord{}, errors.New("lat/lon out of range")
	}
	return Coord{Lat: *p.Lat, Lon: *p.Lon}, nil
}

type ResolvedPlace struct {
	Query    string `json:"query,omitempty"`
	Name     string `json:"name,omitempty"`
	Location Coord  `json:"location"`
	Source   string `json:"source"`
}

type PlanHikeInput struct {
	From      PlaceRef `json:"from"`
	To        PlaceRef `json:"to"`
	AlsoWater *bool    `json:"also_water,omitempty"`
	Include   []string `json:"include,omitempty"`
	Hours     int      `json:"hours,omitempty"`
	CorridorM int      `json:"corridor_m,omitempty"`
	Limit     int      `json:"limit,omitempty"`
}

func (in PlanHikeInput) IncludeWater() bool {
	if in.AlsoWater == nil {
		return true
	}
	return *in.AlsoWater
}

func (in *PlanHikeInput) Validate() error {
	if err := validatePlace("from", in.From); err != nil {
		return err
	}
	if err := validatePlace("to", in.To); err != nil {
		return err
	}
	if in.Hours == 0 {
		in.Hours = 24
	}
	if in.Hours < 1 || in.Hours > 168 {
		return errors.New("hours must be 1..168")
	}
	if in.CorridorM == 0 {
		in.CorridorM = 150
	}
	if in.CorridorM < 50 || in.CorridorM > 2000 {
		return errors.New("corridor_m must be 50..2000")
	}
	if in.Limit == 0 {
		in.Limit = 50
	}
	if in.Limit < 1 || in.Limit > 200 {
		return errors.New("limit must be 1..200")
	}
	for i, s := range in.Include {
		v := strings.ToLower(strings.TrimSpace(s))
		switch v {
		case "guidepost", "trailhead", "shelter", "pass", "entrance":
			in.Include[i] = v
		default:
			return fmt.Errorf("include supports guidepost, trailhead, shelter, pass, entrance (got %q)", s)
		}
	}
	return nil
}

func validatePlace(name string, p PlaceRef) error {
	if p.HasQuery() {
		return nil
	}
	if p.HasCoord() {
		_, err := p.Coord()
		return err
	}
	return fmt.Errorf("%s requires query or lat/lon", name)
}

type PlanHikeResponse struct {
	From         ResolvedPlace    `json:"from"`
	To           ResolvedPlace    `json:"to"`
	Route        RouteResponse    `json:"route"`
	POIs         []POIItem        `json:"pois"`
	Forecast     ForecastResponse `json:"forecast,omitempty"`
	ForecastGoal ForecastResponse `json:"forecast_goal,omitempty"`
	Warnings     []string         `json:"warnings,omitempty"`
	Disclaimer   string           `json:"disclaimer,omitempty"`
}

func (r *PlanHikeResponse) WithMeta() *PlanHikeResponse {
	if r.Disclaimer == "" {
		r.Disclaimer = config.Disclaimer
	}
	r.Route.WithMeta()
	r.Forecast.WithMeta()
	return r
}
