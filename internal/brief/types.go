// Package brief combines caller-selected sources with location-aware forecasts.
// It never turns a routed OSM line into an official hiking route.
package brief

import (
	"fmt"
	"strings"
	"time"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/sources"
)

type RouteInput struct {
	Name         string          `json:"name" jsonschema:"Distinct route label, e.g. south trail or north trail"`
	Start        models.PlaceRef `json:"start" jsonschema:"Trailhead coordinates or exact place name; ambiguous results need coordinate selection"`
	OfficialURLs []string        `json:"official_urls,omitempty" jsonschema:"Up to 3 public authority/operator route pages selected by the caller. Authority is not independently certified."`
}

type Input struct {
	Mountain        models.PlaceRef `json:"mountain" jsonschema:"Summit coordinates or exact place name; prefer coordinates after inspecting geocode results"`
	Routes          []RouteInput    `json:"routes" jsonschema:"1 to 4 named route starts to compare; sources and weather remain available if routing fails"`
	InformationURLs []string        `json:"information_urls,omitempty" jsonschema:"Up to 4 public authority/operator foliage or access notices selected with your web search. No web search is performed by this tool."`
	Subject         string          `json:"subject" jsonschema:"Mountain name used to scope excerpts on multi-mountain pages"`
	AsOf            string          `json:"as_of,omitempty" jsonschema:"Date to assess source age, YYYY-MM-DD; defaults to today in DEFAULT_TZ. Weather is always the current forecast, not a historical reconstruction."`
	IncludeHourly   bool            `json:"include_hourly,omitempty" jsonschema:"Include hourly rows, default false; summaries cover all returned hours"`
	Hours           int             `json:"hours,omitempty" jsonschema:"Forecast hours, default 24, range 1..168"`
}

func (in *Input) Validate() error {
	in.Subject = strings.TrimSpace(in.Subject)
	if in.Subject == "" || len([]rune(in.Subject)) > 100 {
		return fmt.Errorf("subject is required (1..100 characters)")
	}
	if len(in.Routes) < 1 || len(in.Routes) > 4 {
		return fmt.Errorf("routes must contain 1..4 starts")
	}
	if in.Hours == 0 {
		in.Hours = 24
	}
	if in.Hours < 1 || in.Hours > 168 {
		return fmt.Errorf("hours must be 1..168")
	}
	if in.AsOf != "" {
		if _, err := time.Parse("2006-01-02", in.AsOf); err != nil {
			return fmt.Errorf("as_of must be YYYY-MM-DD")
		}
	}
	places := []models.PlaceRef{in.Mountain}
	names := map[string]bool{}
	urls := append([]string{}, in.InformationURLs...)
	if len(in.InformationURLs) > 4 {
		return fmt.Errorf("information_urls supports at most 4 URLs")
	}
	for _, r := range in.Routes {
		name := strings.TrimSpace(r.Name)
		if name == "" || len([]rune(name)) > 100 || names[name] {
			return fmt.Errorf("route names must be distinct, 1..100 characters")
		}
		names[name] = true
		if len(r.OfficialURLs) > 3 {
			return fmt.Errorf("official_urls supports at most 3 URLs per route")
		}
		urls = append(urls, r.OfficialURLs...)
		places = append(places, r.Start)
	}
	for _, p := range places {
		if p.HasQuery() {
			if p.HasCoord() {
				return fmt.Errorf("use either query or coordinates, not both")
			}
			if len([]rune(p.Query)) > 200 {
				return fmt.Errorf("query too long")
			}
			continue
		}
		if _, err := p.Coord(); err != nil {
			return fmt.Errorf("mountain/start: %w", err)
		}
	}
	if len(unique(urls)) > 8 {
		return fmt.Errorf("at most 8 distinct source URLs per call")
	}
	for _, u := range urls {
		if err := sources.ValidateURL(u); err != nil {
			return err
		}
	}
	return nil
}

type LocationResult struct {
	Status     string                `json:"status"` // resolved, needs_selection, unavailable
	Place      *models.ResolvedPlace `json:"place,omitempty"`
	Candidates []models.GeocodeItem  `json:"candidates,omitempty"`
	Reason     string                `json:"reason,omitempty"`
}

type WeatherSummary struct {
	From                  string  `json:"from"`
	Through               string  `json:"through"`
	TemperatureMinC       float64 `json:"temperature_min_c"`
	TemperatureMaxC       float64 `json:"temperature_max_c"`
	PrecipTotalMM         float64 `json:"precip_total_mm"`
	RainProbabilityMaxPct float64 `json:"rain_probability_max_pct"`
	WindMaxMPS            float64 `json:"wind_max_mps"`
	GustMaxMPS            float64 `json:"gust_max_mps"`
}

type ForecastInfo struct {
	RequestedLocation models.Coord  `json:"requested_location"`
	ModelLocation     *models.Coord `json:"model_location,omitempty"`
	ModelElevationM   *float64      `json:"model_elevation_m,omitempty"`
	RetrievedAt       string        `json:"retrieved_at"`
	SourceURL         string        `json:"source_url"`
	ModelRunAt        string        `json:"model_run_at"`
	Timezone          string        `json:"timezone"`
	Sunrise           string        `json:"sunrise,omitempty"`
	Sunset            string        `json:"sunset,omitempty"`
}

type PointConditions struct {
	Hourly []models.HourlyForecast `json:"hourly,omitempty"`

	Status            string                    `json:"status"`
	RequestedLocation models.Coord              `json:"requested_location"`
	TerrainElevation  *models.ElevationResponse `json:"terrain_elevation_estimate,omitempty"`
	Forecast          *ForecastInfo             `json:"forecast,omitempty"`
	Summary           *WeatherSummary           `json:"summary,omitempty"`
	Warnings          []string                  `json:"warnings,omitempty"`
}

type TopicEvidence struct {
	Status    string            `json:"status"` // source_excerpts or unconfirmed; never guarantees current conditions
	Excerpts  []sources.Excerpt `json:"excerpts"`
	CheckURLs []string          `json:"check_urls"`
	Reason    string            `json:"reason"`
}

type RouteResult struct {
	Name                      string                   `json:"name"`
	Start                     LocationResult           `json:"start"`
	Conditions                *PointConditions         `json:"conditions,omitempty"`
	RouteEstimate             *models.RouteResponse    `json:"route_estimate,omitempty"`
	RoutingStatus             string                   `json:"routing_status"`
	RoutingReason             string                   `json:"routing_reason,omitempty"`
	PointElevationDifferenceM *float64                 `json:"point_elevation_difference_m,omitempty"`
	OfficialRouteEvidence     map[string]TopicEvidence `json:"official_route_evidence"`
	Foliage                   TopicEvidence            `json:"foliage"`
	Access                    TopicEvidence            `json:"access"`
}

type Response struct {
	Information      map[string]TopicEvidence `json:"information"`
	Subject          string                   `json:"subject"`
	AsOf             string                   `json:"as_of"`
	RetrievedAt      string                   `json:"retrieved_at"`
	Mountain         LocationResult           `json:"mountain"`
	SummitConditions *PointConditions         `json:"summit_conditions,omitempty"`
	Routes           []RouteResult            `json:"routes"`
	Sources          []sources.Document       `json:"sources"`
	Limitations      []string                 `json:"limitations"`
}
