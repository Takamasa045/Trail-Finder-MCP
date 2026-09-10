package models

import "testing"

func TestTrailheadsValidateDefaults(t *testing.T) {
	in := TrailheadsInput{Lat: 35.6, Lon: 139.7}
	if err := in.Validate(); err != nil {
		t.Fatal(err)
	}
	if in.RadiusM != 2000 || in.Limit != 200 {
		t.Fatalf("defaults radius=%d limit=%d", in.RadiusM, in.Limit)
	}
}

func TestTrailheadsRejectsEntranceTypos(t *testing.T) {
	in := TrailheadsInput{Lat: 35.6, Lon: 139.7, Include: []string{"waterfall"}}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestTrailheadsAcceptsJapanTypes(t *testing.T) {
	in := TrailheadsInput{Lat: 35.6, Lon: 139.7, Include: []string{"Shelter", "PASS"}}
	if err := in.Validate(); err != nil {
		t.Fatal(err)
	}
	if in.Include[0] != "shelter" || in.Include[1] != "pass" {
		t.Fatalf("normalized=%v", in.Include)
	}
}

func TestTrailheadsLatRange(t *testing.T) {
	in := TrailheadsInput{Lat: 91, Lon: 0}
	if err := in.Validate(); err == nil {
		t.Fatal("expected range error")
	}
}

func TestForecastHours(t *testing.T) {
	in := ForecastInput{Lat: 35, Lon: 139, Hours: 200}
	if err := in.Validate(); err == nil {
		t.Fatal("expected hours error")
	}
	in = ForecastInput{Lat: 35, Lon: 139}
	if err := in.Validate(); err != nil || in.Hours != 24 {
		t.Fatalf("default hours err=%v hours=%d", err, in.Hours)
	}
}

func TestGeocodeRequiresQuery(t *testing.T) {
	in := GeocodeInput{}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlaceRefAndPlan(t *testing.T) {
	in := PlanHikeInput{}
	if err := in.Validate(); err == nil {
		t.Fatal("expected missing from/to")
	}
	in = PlanHikeInput{From: PlaceRef{Query: "高尾山口駅"}, To: PlaceRef{Query: "高尾山"}}
	if err := in.Validate(); err != nil {
		t.Fatal(err)
	}
	if in.CorridorM != 150 || in.Hours != 24 {
		t.Fatalf("defaults corridor=%d hours=%d", in.CorridorM, in.Hours)
	}

	lat, lon := 35.6, 139.7
	in = PlanHikeInput{From: PlaceRef{Lat: &lat, Lon: &lon}, To: PlaceRef{Query: "高尾山"}}
	if err := in.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestElevationAndGeocodeLimits(t *testing.T) {
	if err := (&ElevationInput{Lat: 91, Lon: 0}).Validate(); err == nil {
		t.Fatal("expected elevation range error")
	}
	if err := (&ElevationInput{Lat: 35, Lon: 139}).Validate(); err != nil {
		t.Fatal(err)
	}
	g := GeocodeInput{Query: "  x  ", Limit: 99}
	if err := g.Validate(); err == nil {
		t.Fatal("expected geocode limit error")
	}
	g = GeocodeInput{Query: "  x  "}
	if err := g.Validate(); err != nil || g.Query != "x" || g.Limit != 5 {
		t.Fatalf("%+v %v", g, err)
	}
}

func TestPlanIncludeAndCoord(t *testing.T) {
	in := PlanHikeInput{From: PlaceRef{Query: "a"}, To: PlaceRef{Query: "b"}, Include: []string{"nope"}}
	if err := in.Validate(); err == nil {
		t.Fatal("expected include error")
	}
	lat := 91.0
	lon := 0.0
	p := PlaceRef{Lat: &lat, Lon: &lon}
	if _, err := p.Coord(); err == nil {
		t.Fatal("expected coord range")
	}
	resp := (&TrailheadsResponse{}).WithMeta()
	if resp.Disclaimer == "" {
		t.Fatal("disclaimer")
	}
	if (&RouteResponse{}).WithMeta().Disclaimer == "" {
		t.Fatal("route disclaimer")
	}
	if (&ElevationResponse{}).WithMeta().Disclaimer == "" {
		t.Fatal("elev disclaimer")
	}
	if (&ForecastResponse{}).WithMeta().Disclaimer == "" {
		t.Fatal("wx disclaimer")
	}
	if (&GeocodeResponse{}).WithMeta().Disclaimer == "" {
		t.Fatal("geo disclaimer")
	}
	if (&PlanHikeResponse{}).WithMeta().Disclaimer == "" {
		t.Fatal("plan disclaimer")
	}
	in2 := PlanHikeInput{From: PlaceRef{Query: "a"}, To: PlaceRef{Query: "b"}, Hours: 200}
	if err := in2.Validate(); err == nil {
		t.Fatal("hours")
	}
	in3 := PlanHikeInput{From: PlaceRef{Query: "a"}, To: PlaceRef{Query: "b"}, CorridorM: 5}
	if err := in3.Validate(); err == nil {
		t.Fatal("corridor")
	}
}

func TestRouteEngine(t *testing.T) {
	in := RouteInput{From: Coord{Lat: 1, Lon: 1}, To: Coord{Lat: 2, Lon: 2}}
	if err := in.Validate(); err != nil || in.Engine != "auto" {
		t.Fatalf("err=%v engine=%s", err, in.Engine)
	}
	in.Engine = "graphhopper"
	if err := in.Validate(); err == nil {
		t.Fatal("expected engine error")
	}
}
