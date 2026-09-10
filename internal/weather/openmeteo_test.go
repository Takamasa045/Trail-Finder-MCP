package weather

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func TestForecastRequestsMetersPerSecond(t *testing.T) {
	var rawQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"timezone": "Asia/Tokyo",
			"hourly": map[string]any{
				"time":                      []string{"2026-09-10T00:00", "2026-09-10T01:00"},
				"temperature_2m":            []float64{12.5, 11.0},
				"precipitation":             []float64{0.2, 1.0},
				"precipitation_probability": []float64{10, 40},
				"wind_speed_10m":            []float64{2.5, 3.0},
				"wind_gusts_10m":            []float64{5.0, 8.0},
				"weather_code":              []int{1, 61},
			},
			"daily": map[string]any{
				"sunrise": []string{"2026-09-10T05:30"},
				"sunset":  []string{"2026-09-10T18:00"},
			},
		})
	}))
	defer ts.Close()

	t.Setenv("OPENMETEO_URL", ts.URL)
	t.Setenv("DEFAULT_TZ", "Asia/Tokyo")

	resp, err := Forecast(context.Background(), models.ForecastInput{Lat: 35.6, Lon: 139.7, Hours: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rawQuery, "wind_speed_unit=ms") {
		t.Fatalf("missing wind_speed_unit=ms in %s", rawQuery)
	}
	if !strings.Contains(rawQuery, "forecast_hours=2") {
		t.Fatalf("missing forecast_hours in %s", rawQuery)
	}
	if !strings.Contains(rawQuery, "precipitation_probability") {
		t.Fatalf("missing precip probability in %s", rawQuery)
	}
	if resp.SunriseISO != "2026-09-10T05:30" || resp.SunsetISO != "2026-09-10T18:00" {
		t.Fatalf("sun %+v", resp)
	}
	if len(resp.Hourly) != 2 {
		t.Fatalf("hourly=%d", len(resp.Hourly))
	}
	if resp.Hourly[1].WindMps != 3.0 || resp.Hourly[1].WeatherCode != 61 {
		t.Fatalf("hour1=%+v", resp.Hourly[1])
	}
}
