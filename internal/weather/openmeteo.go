package weather

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/httpx"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
)

func Forecast(ctx context.Context, in models.ForecastInput) (*models.ForecastResponse, error) {
	base := config.Env("OPENMETEO_URL", "https://api.open-meteo.com/v1/forecast")
	q := url.Values{}
	q.Set("latitude", fmt.Sprintf("%f", in.Lat))
	q.Set("longitude", fmt.Sprintf("%f", in.Lon))
	q.Set("hourly", "temperature_2m,precipitation,precipitation_probability,wind_speed_10m,wind_gusts_10m,weather_code")
	q.Set("daily", "sunrise,sunset")
	q.Set("timezone", config.DefaultTZ())
	q.Set("wind_speed_unit", "ms")
	q.Set("forecast_hours", fmt.Sprintf("%d", in.Hours))

	var out struct {
		Timezone string `json:"timezone"`
		Hourly   struct {
			Time        []string  `json:"time"`
			Temperature []float64 `json:"temperature_2m"`
			Precip      []float64 `json:"precipitation"`
			PrecipProb  []float64 `json:"precipitation_probability"`
			Wind        []float64 `json:"wind_speed_10m"`
			Gust        []float64 `json:"wind_gusts_10m"`
			Code        []int     `json:"weather_code"`
		} `json:"hourly"`
		Daily struct {
			Sunrise []string `json:"sunrise"`
			Sunset  []string `json:"sunset"`
		} `json:"daily"`
	}
	if err := httpx.GetJSON(ctx, base+"?"+q.Encode(), &out); err != nil {
		return nil, fmt.Errorf("open-meteo: %w", err)
	}

	hours := in.Hours
	if hours <= 0 {
		hours = 24
	}
	n := hours
	if n > len(out.Hourly.Time) {
		n = len(out.Hourly.Time)
	}

	resp := &models.ForecastResponse{
		Timezone:    out.Timezone,
		Hourly:      make([]models.HourlyForecast, 0, n),
		Attribution: "Open-Meteo.com",
	}
	if len(out.Daily.Sunrise) > 0 {
		resp.SunriseISO = out.Daily.Sunrise[0]
	}
	if len(out.Daily.Sunset) > 0 {
		resp.SunsetISO = out.Daily.Sunset[0]
	}
	for i := 0; i < n; i++ {
		h := models.HourlyForecast{TimeISO: out.Hourly.Time[i]}
		h.TemperatureC = at(out.Hourly.Temperature, i)
		h.PrecipMM = at(out.Hourly.Precip, i)
		h.PrecipProbabilityPct = at(out.Hourly.PrecipProb, i)
		h.WindMps = at(out.Hourly.Wind, i)
		h.WindGustMps = at(out.Hourly.Gust, i)
		if i < len(out.Hourly.Code) {
			h.WeatherCode = out.Hourly.Code[i]
		}
		resp.Hourly = append(resp.Hourly, h)
	}
	return resp.WithMeta(), nil
}

func at(vals []float64, i int) float64 {
	if i < len(vals) {
		return vals[i]
	}
	return 0
}
