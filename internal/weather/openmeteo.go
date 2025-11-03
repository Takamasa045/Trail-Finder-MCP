package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"trail-finder-mcp/internal/models"
)

var httpClient = &http.Client{ Timeout: 15 * time.Second }

func Forecast(ctx context.Context, in models.ForecastInput) (*models.ForecastResponse, error) {
	base := os.Getenv("OPENMETEO_URL")
	if base == "" {
		base = "https://api.open-meteo.com/v1/forecast"
	}
	tz := os.Getenv("DEFAULT_TZ")
	if tz == "" {
		tz = "Asia/Tokyo"
	}

	q := url.Values{}
	q.Set("latitude", fmt.Sprintf("%f", in.Lat))
	q.Set("longitude", fmt.Sprintf("%f", in.Lon))
	q.Set("hourly", "temperature_2m,precipitation,wind_speed_10m")
	q.Set("timezone", tz)

	endpoint := base + "?" + q.Encode()
	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	req.Header.Set("User-Agent", "trail-finder-mcp/0.1")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("open-meteo status=%d body=%s", res.StatusCode, string(b))
	}

	var out struct {
		Timezone string `json:"timezone"`
		Hourly struct{
			Time []string `json:"time"`
			Temperature []float64 `json:"temperature_2m"`
			Precip []float64 `json:"precipitation"`
			Wind []float64 `json:"wind_speed_10m"`
		} `json:"hourly"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	n := in.Hours
	if n > len(out.Hourly.Time) {
		n = len(out.Hourly.Time)
	}

	resp := &models.ForecastResponse{
		Timezone: out.Timezone,
		Hourly:   make([]models.HourlyForecast, 0, n),
		Attribution: "Open‑Meteo.com",
	}
	for i := 0; i < n; i++ {
		var t, p, w float64
		if i < len(out.Hourly.Temperature) { t = out.Hourly.Temperature[i] }
		if i < len(out.Hourly.Precip) { p = out.Hourly.Precip[i] }
		if i < len(out.Hourly.Wind) { w = out.Hourly.Wind[i] }
		h := models.HourlyForecast{
			TimeISO: out.Hourly.Time[i],
			TemperatureC: t,
			PrecipMM: p,
			WindMps: w,
		}
		resp.Hourly = append(resp.Hourly, h)
	}
	return resp, nil
}
