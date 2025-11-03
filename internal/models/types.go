package models

import (
	"errors"
	"math"
)

// ---- Common response ----

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Hint      string `json:"hint,omitempty"`
	Retryable bool   `json:"retryable"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ---- Trailheads ----

type TrailheadsInput struct {
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	RadiusM    int      `json:"radius_m,omitempty"`
	Include    []string `json:"include,omitempty"` // guidepost, trailhead
	AlsoWater  bool     `json:"also_water,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

func (t *TrailheadsInput) Validate() error {
	if math.Abs(t.Lat) > 90 || math.Abs(t.Lon) > 180 {
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
	return nil
}

type POIItem struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // guidepost/trailhead/drinking_water/spring
	Name       string            `json:"name,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
	Location   Coord             `json:"location"`
	DistanceM  float64           `json:"distance_m,omitempty"`
	Source     string            `json:"source"`
}

type TrailheadsResponse struct {
	Center  Coord    `json:"center"`
	RadiusM int      `json:"radius_m"`
	Items   []POIItem `json:"items"`
	Attribution string `json:"attribution,omitempty"`
}

type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// ---- Route ----

type RouteInput struct {
	From    Coord             `json:"from"`
	To      Coord             `json:"to"`
	Engine  string            `json:"engine,omitempty"` // auto|osrm|valhalla
	Options map[string]any    `json:"options,omitempty"`
}

func (in *RouteInput) Validate() error {
	if math.Abs(in.From.Lat) > 90 || math.Abs(in.From.Lon) > 180 ||
		math.Abs(in.To.Lat) > 90 || math.Abs(in.To.Lon) > 180 {
		return errors.New("lat/lon out of range")
	}
	if in.Engine == "" {
		in.Engine = "auto"
	}
	return nil
}

type GeoJSONLineString struct {
	Type        string        `json:"type"`
	Coordinates [][]float64   `json:"coordinates"` // [lon,lat]
}

type RouteResponse struct {
	Engine         string            `json:"engine"`
	DistanceM      float64           `json:"distance_m"`
	DurationS      float64           `json:"duration_s"`
	Geometry       GeoJSONLineString `json:"geometry,omitempty"`
	Steps          []any             `json:"steps,omitempty"`
	ElevationGainM float64           `json:"elevation_gain_m,omitempty"`
	ElevationLossM float64           `json:"elevation_loss_m,omitempty"`
}

// ---- Elevation ----

type ElevationInput struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (e *ElevationInput) Validate() error {
	if math.Abs(e.Lat) > 90 || math.Abs(e.Lon) > 180 {
		return errors.New("lat/lon out of range")
	}
	return nil
}

type ElevationResponse struct {
	ElevationM float64 `json:"elevation_m"`
	Source     string  `json:"source"`
}

// ---- Forecast ----

type ForecastInput struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Hours int     `json:"hours,omitempty"`
}

func (f *ForecastInput) Validate() error {
	if math.Abs(f.Lat) > 90 || math.Abs(f.Lon) > 180 {
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
	TimeISO       string  `json:"time"`
	TemperatureC  float64 `json:"temperature_c"`
	PrecipMM      float64 `json:"precip_mm"`
	WindMps       float64 `json:"wind_mps"`
}

type ForecastResponse struct {
	Timezone string           `json:"timezone"`
	Hourly   []HourlyForecast `json:"hourly"`
	Attribution string        `json:"attribution,omitempty"`
}
