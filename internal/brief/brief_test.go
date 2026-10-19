package brief

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/sources"
)

func point(lat, lon float64) models.PlaceRef { return models.PlaceRef{Lat: &lat, Lon: &lon} }
func input() Input {
	return Input{Subject: "Test mountain", Mountain: point(36, 138), Routes: []RouteInput{{Name: "south", Start: point(35.99, 138), OfficialURLs: []string{"https://authority.example/south"}}, {Name: "north", Start: point(36.01, 138), OfficialURLs: []string{"https://authority.example/north"}}}, InformationURLs: []string{"https://authority.example/2019/10/17/old"}, AsOf: "2026-10-19", Hours: 2}
}
func fixtures() dependencies {
	return dependencies{
		now: func() time.Time { return time.Date(2026, 10, 19, 3, 0, 0, 0, time.UTC) },
		search: func(context.Context, models.GeocodeInput) (*models.GeocodeResponse, error) {
			return nil, fmt.Errorf("search unavailable")
		},
		elevation: func(_ context.Context, in models.ElevationInput) (*models.ElevationResponse, error) {
			return &models.ElevationResponse{ElevationM: 1000 + (in.Lat-35.99)*10000, Source: "test DEM"}, nil
		},
		forecast: func(_ context.Context, in models.ForecastInput) (*models.ForecastResponse, error) {
			return &models.ForecastResponse{RequestedLocation: models.Coord{Lat: in.Lat, Lon: in.Lon}, ModelLocation: &models.Coord{Lat: 36, Lon: 138}, RetrievedAt: "2026-10-19T03:00:00Z", ModelRunAt: "not supplied by upstream", Timezone: "Asia/Tokyo", Hourly: []models.HourlyForecast{{TimeISO: "2026-10-19T12:00", TemperatureC: 5, PrecipMM: 0, WindMps: 2}, {TimeISO: "2026-10-19T13:00", TemperatureC: 4, PrecipMM: 1, WindMps: 3, WindGustMps: 5}}}, nil
		},
		route: func(context.Context, models.RouteInput) (*models.RouteResponse, error) {
			return nil, fmt.Errorf("osrm: route snapped too far")
		},
		fetch: func(_ context.Context, u, subject string, asOf time.Time) sources.Document {
			return sources.Document{URL: u, Status: "fetched", Title: subject, Published: &sources.Date{Value: "2019-10-17", Origin: "metadata"}, Freshness: "older_than_14_days", CheckLinks: []string{u}, Excerpts: []sources.Excerpt{{SourceURL: u, Text: "distance 8 km return", Topic: "distance", Freshness: "older_than_14_days"}, {SourceURL: u, Text: "closure after 2019 typhoon", Topic: "access", Freshness: "older_than_14_days"}}}
		},
	}
}
func TestComparisonRetainsWeatherAndEvidenceAfterRoutingFailure(t *testing.T) {
	out, err := compare(context.Background(), input(), fixtures())
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Routes) != 2 || len(out.Sources) != 3 {
		t.Fatalf("%+v", out)
	}
	for _, r := range out.Routes {
		if r.RoutingStatus != "unavailable" || r.RouteEstimate != nil || !strings.Contains(r.RoutingReason, "snapped") {
			t.Fatalf("route=%+v", r)
		}
		if r.Conditions.Summary.PrecipTotalMM != 1 || r.Conditions.Summary.TemperatureMinC != 4 || r.Conditions.Forecast.ModelLocation.Lat != 36 {
			t.Fatalf("conditions=%+v", r.Conditions)
		}
		if len(r.Conditions.Hourly) != 0 {
			t.Fatal("hourly rows must be opt in")
		}
		if r.OfficialRouteEvidence["distance"].Status != "source_excerpts" || len(r.OfficialRouteEvidence["distance"].Excerpts) != 1 {
			t.Fatalf("evidence=%+v", r.OfficialRouteEvidence)
		}
		if r.Foliage.Status != "unconfirmed" || r.Access.Status != "unconfirmed" {
			t.Fatal("old data confirmed current conditions")
		}
		if r.PointElevationDifferenceM == nil {
			t.Fatal("missing estimated point difference")
		}
	}
	if out.SummitConditions == nil || out.SummitConditions.Summary == nil {
		t.Fatal("summit weather lost")
	}
	if !strings.Contains(Markdown(out), "south") || !strings.Contains(Markdown(out), "unconfirmed") {
		t.Fatal("missing readable result")
	}
}
func TestAmbiguousNamesRequireSelectionButSourcesSurvive(t *testing.T) {
	in := input()
	in.Mountain = models.PlaceRef{Query: "Shared mountain"}
	d := fixtures()
	d.search = func(context.Context, models.GeocodeInput) (*models.GeocodeResponse, error) {
		return &models.GeocodeResponse{Items: []models.GeocodeItem{{Name: "Shared mountain", Location: models.Coord{Lat: 36, Lon: 138}}, {Name: "Shared mountain", Location: models.Coord{Lat: 40, Lon: 140}}}}, nil
	}
	d.route = func(context.Context, models.RouteInput) (*models.RouteResponse, error) {
		t.Error("must not route ambiguous name")
		return nil, nil
	}
	out, err := compare(context.Background(), in, d)
	if err != nil {
		t.Fatal(err)
	}
	if out.Mountain.Status != "needs_selection" || len(out.Mountain.Candidates) != 2 || out.SummitConditions != nil {
		t.Fatalf("%+v", out.Mountain)
	}
	if out.Routes[0].Conditions == nil || len(out.Sources) == 0 {
		t.Fatal("unrelated useful data lost")
	}
}
func TestFuzzySingleMatchAndFetchFailure(t *testing.T) {
	in := input()
	in.Routes[0].Start = models.PlaceRef{Query: "Intended place"}
	d := fixtures()
	d.search = func(context.Context, models.GeocodeInput) (*models.GeocodeResponse, error) {
		return &models.GeocodeResponse{Items: []models.GeocodeItem{{Name: "Other place"}}}, nil
	}
	d.fetch = func(_ context.Context, u, s string, date time.Time) sources.Document {
		return sources.Document{URL: u, Status: "unavailable", Error: "HTTP 503"}
	}
	out, err := compare(context.Background(), in, d)
	if err != nil {
		t.Fatal(err)
	}
	if out.Routes[0].Start.Status != "needs_selection" || out.Routes[0].Conditions != nil {
		t.Fatal("fuzzy candidate used")
	}
	if out.Routes[0].Access.Status != "unconfirmed" || len(out.Routes[0].Access.CheckURLs) == 0 {
		t.Fatal("missing verification links after fetch failure")
	}
}
func TestRecentSourceIsNotLiveStatus(t *testing.T) {
	e := evidence("foliage", []sources.Document{{URL: "https://authority.example/", Status: "fetched", Excerpts: []sources.Excerpt{{Topic: "foliage", Text: "Color report", Freshness: "within_14_days"}}}}, true)
	if e.Status != "unconfirmed" || !strings.Contains(e.Reason, "Recent dated") {
		t.Fatalf("%+v", e)
	}
}
func TestValidation(t *testing.T) {
	for _, alter := range []func(*Input){
		func(in *Input) { in.Subject = "" }, func(in *Input) { in.Routes = nil }, func(in *Input) { in.Hours = 999 }, func(in *Input) { in.AsOf = "10/19" },
		func(in *Input) { in.Routes[1].Name = in.Routes[0].Name }, func(in *Input) { in.Mountain = models.PlaceRef{} },
		func(in *Input) { in.InformationURLs = []string{"http://127.0.0.1/"} }, func(in *Input) { in.Mountain.Query = "query and coordinates" },
	} {
		in := input()
		alter(&in)
		if err := in.Validate(); err == nil {
			t.Fatalf("invalid input accepted: %+v", in)
		}
	}
}
