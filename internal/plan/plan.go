package plan

import (
	"context"
	"fmt"
	"strings"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/geocode"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/overpass"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/router"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/weather"
)

func Plan(ctx context.Context, in models.PlanHikeInput) (*models.PlanHikeResponse, error) {
	from, err := resolve(ctx, in.From)
	if err != nil {
		return nil, fmt.Errorf("from: %w", err)
	}
	to, err := resolve(ctx, in.To)
	if err != nil {
		return nil, fmt.Errorf("to: %w", err)
	}

	route, err := router.RouteFoot(ctx, models.RouteInput{
		From:    from.Location,
		To:      to.Location,
		Engine:  "osrm",
		Options: map[string]any{"include_geometry": true, "include_elevation": true},
	})
	if err != nil {
		return nil, err
	}

	out := &models.PlanHikeResponse{
		From:  from,
		To:    to,
		Route: *route,
		POIs:  []models.POIItem{},
	}
	if !route.ElevationSampled {
		out.Warnings = append(out.Warnings, "elevation profile unavailable; gain/loss omitted")
	}

	include := in.Include
	if len(include) == 0 {
		include = []string{"guidepost", "trailhead", "shelter", "pass"}
	}
	coords := router.GeometryCoords(*route)
	if len(coords) == 0 {
		coords = []models.Coord{from.Location, to.Location}
	}
	pois, err := overpass.QueryAlongRoute(ctx, coords, in.CorridorM, include, in.IncludeWater(), in.Limit)
	if err != nil {
		out.Warnings = append(out.Warnings, "POI search failed: "+err.Error())
	} else if pois != nil {
		out.POIs = pois.Items
	}

	startWx, err := weather.Forecast(ctx, models.ForecastInput{
		Lat: from.Location.Lat, Lon: from.Location.Lon, Hours: in.Hours,
	})
	if err != nil {
		out.Warnings = append(out.Warnings, "start forecast failed: "+err.Error())
	} else if startWx != nil {
		out.Forecast = *startWx
	}

	goalWx, err := weather.Forecast(ctx, models.ForecastInput{
		Lat: to.Location.Lat, Lon: to.Location.Lon, Hours: in.Hours,
	})
	if err != nil {
		out.Warnings = append(out.Warnings, "goal forecast failed: "+err.Error())
	} else if goalWx != nil {
		out.ForecastGoal = *goalWx
	}

	return out.WithMeta(), nil
}

func resolve(ctx context.Context, p models.PlaceRef) (models.ResolvedPlace, error) {
	if p.HasQuery() {
		item, err := geocode.First(ctx, strings.TrimSpace(p.Query))
		if err != nil {
			return models.ResolvedPlace{}, err
		}
		name := item.Name
		if name == "" {
			name = item.DisplayName
		}
		return models.ResolvedPlace{
			Query:    p.Query,
			Name:     name,
			Location: item.Location,
			Source:   "nominatim",
		}, nil
	}
	c, err := p.Coord()
	if err != nil {
		return models.ResolvedPlace{}, err
	}
	return models.ResolvedPlace{
		Location: c,
		Source:   "coordinates",
	}, nil
}
