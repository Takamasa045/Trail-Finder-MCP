package elevation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"trail-finder-mcp/internal/config"
	"trail-finder-mcp/internal/models"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

func Lookup(ctx context.Context, in models.ElevationInput) (*models.ElevationResponse, error) {
	provider := os.Getenv("ELEVATION_PROVIDER")
	if provider == "" {
		provider = "open-elevation"
	}
	switch provider {
	case "open-topo", "open-topodata", "opentopo", "opentopodata":
		return openTopo(ctx, in)
	default:
		return openElevation(ctx, in)
	}
}

func openElevation(ctx context.Context, in models.ElevationInput) (*models.ElevationResponse, error) {
	v := url.Values{}
	v.Set("locations", fmt.Sprintf("%f,%f", in.Lat, in.Lon))
	endpoint := "https://api.open-elevation.com/api/v1/lookup?" + v.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	req.Header.Set("User-Agent", config.UserAgent())
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("open-elevation status=%d body=%s", res.StatusCode, string(b))
	}
	var out struct {
		Results []struct {
			Elevation float64 `json:"elevation"`
		} `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Results) == 0 {
		return nil, fmt.Errorf("no elevation result")
	}
	return &models.ElevationResponse{ElevationM: out.Results[0].Elevation, Source: "open-elevation"}, nil
}

func openTopo(ctx context.Context, in models.ElevationInput) (*models.ElevationResponse, error) {
	base := "https://api.opentopodata.org/v1/srtm90m"
	v := url.Values{}
	v.Set("locations", fmt.Sprintf("%f,%f", in.Lat, in.Lon))
	endpoint := base + "?" + v.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	req.Header.Set("User-Agent", config.UserAgent())
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("open-topodata status=%d body=%s", res.StatusCode, string(b))
	}
	var out struct {
		Results []struct {
			Elevation float64 `json:"elevation"`
		} `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Results) == 0 {
		return nil, fmt.Errorf("no elevation result")
	}
	return &models.ElevationResponse{ElevationM: out.Results[0].Elevation, Source: "open-topodata"}, nil
}
