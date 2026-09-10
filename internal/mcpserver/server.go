package mcpserver

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/elevation"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/geocode"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/overpass"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/plan"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/router"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/weather"
)

const serverName = "trail-finder-mcp"

func Run(ctx context.Context) error {
	return NewServer().Run(ctx, &mcp.StdioTransport{})
}

func NewServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: config.Version(),
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "geocode",
		Description: "Geocode a place name (Japanese or English, e.g. 高尾山口駅 or Mt. Takao) to coordinates via OpenStreetMap Nominatim. Use this when the user gives a station or mountain name instead of lat/lon.",
	}, handleGeocode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "trailheads",
		Description: "Find nearby trailheads, guideposts, huts, passes, and optional water sources from OpenStreetMap (Overpass API). Prefers Japan-relevant tags such as highway=trailhead.",
	}, handleTrailheads)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "route_foot",
		Description: "Plan a walking route between two coordinates using OSRM. Includes geometry and elevation gain/loss when available. engine=valhalla is not implemented.",
	}, handleRouteFoot)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "elevation",
		Description: "Retrieve elevation in meters for a single coordinate (default provider: Open-Meteo).",
	}, handleElevation)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "forecast",
		Description: "Get an hourly hiking forecast from Open-Meteo: temperature, precipitation, rain probability, wind (m/s), gusts, weather code, sunrise and sunset.",
	}, handleForecast)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "plan_hike",
		Description: "Plan a hike from place names or coordinates: geocode, walking route, POIs along the route (trailheads/guideposts/huts/water), elevation gain/loss, and forecasts at start and goal. Water is included unless also_water=false. If both query and lat/lon are set, query wins.",
	}, handlePlanHike)

	return server
}

type trailheadsArgs struct {
	Lat       float64  `json:"lat" description:"Latitude in decimal degrees"`
	Lon       float64  `json:"lon" description:"Longitude in decimal degrees"`
	RadiusM   int      `json:"radius_m,omitempty" description:"Search radius in meters (default 2000)"`
	Include   []string `json:"include,omitempty" description:"OSM categories: guidepost, trailhead, shelter, pass, entrance"`
	AlsoWater bool     `json:"also_water,omitempty" description:"Include drinking water and springs"`
	Limit     int      `json:"limit,omitempty" description:"Maximum number of POIs to return"`
}

type routeArgs struct {
	From    models.Coord   `json:"from" description:"Starting location (lat/lon)"`
	To      models.Coord   `json:"to" description:"Destination location (lat/lon)"`
	Engine  string         `json:"engine,omitempty" description:"Routing engine (auto or osrm). valhalla is not implemented."`
	Options map[string]any `json:"options,omitempty" description:"Optional flags: include_geometry, include_steps, avoid_ferry, include_elevation"`
}

type elevationArgs struct {
	Lat float64 `json:"lat" description:"Latitude in decimal degrees"`
	Lon float64 `json:"lon" description:"Longitude in decimal degrees"`
}

type forecastArgs struct {
	Lat   float64 `json:"lat" description:"Latitude in decimal degrees"`
	Lon   float64 `json:"lon" description:"Longitude in decimal degrees"`
	Hours int     `json:"hours,omitempty" description:"Number of forecast hours (default 24, max 168)"`
}

type geocodeArgs struct {
	Query        string `json:"query" description:"Place name to geocode"`
	Limit        int    `json:"limit,omitempty" description:"Max results (default 5, max 10)"`
	CountryCodes string `json:"country_codes,omitempty" description:"Optional Nominatim countrycodes, e.g. jp"`
}

type planArgs struct {
	From      models.PlaceRef `json:"from" description:"Start: {query} or {lat,lon}"`
	To        models.PlaceRef `json:"to" description:"Goal: {query} or {lat,lon}"`
	AlsoWater *bool           `json:"also_water,omitempty" description:"Include water sources along the route (default true)"`
	Include   []string        `json:"include,omitempty" description:"POI types along the route"`
	Hours     int             `json:"hours,omitempty" description:"Forecast hours from the start (default 24)"`
	CorridorM int             `json:"corridor_m,omitempty" description:"Search corridor around the route in meters (default 150)"`
	Limit     int             `json:"limit,omitempty" description:"Max POIs along the route"`
}

func handleTrailheads(ctx context.Context, _ *mcp.CallToolRequest, args trailheadsArgs) (*mcp.CallToolResult, models.TrailheadsResponse, error) {
	input := models.TrailheadsInput{
		Lat: args.Lat, Lon: args.Lon, RadiusM: args.RadiusM,
		Include: args.Include, AlsoWater: args.AlsoWater, Limit: args.Limit,
	}
	if err := input.Validate(); err != nil {
		return nil, models.TrailheadsResponse{}, err
	}
	resp, err := overpass.QueryPOIs(ctx, input)
	if err != nil {
		return nil, models.TrailheadsResponse{}, err
	}
	if resp == nil {
		return nil, models.TrailheadsResponse{}, errors.New("empty response from overpass")
	}
	return nil, *resp, nil
}

func handleRouteFoot(ctx context.Context, _ *mcp.CallToolRequest, args routeArgs) (*mcp.CallToolResult, models.RouteResponse, error) {
	input := models.RouteInput{From: args.From, To: args.To, Engine: args.Engine, Options: args.Options}
	if err := input.Validate(); err != nil {
		return nil, models.RouteResponse{}, err
	}
	resp, err := router.RouteFoot(ctx, input)
	if err != nil {
		return nil, models.RouteResponse{}, err
	}
	if resp == nil {
		return nil, models.RouteResponse{}, errors.New("empty response from routing engine")
	}
	return nil, *resp, nil
}

func handleElevation(ctx context.Context, _ *mcp.CallToolRequest, args elevationArgs) (*mcp.CallToolResult, models.ElevationResponse, error) {
	input := models.ElevationInput{Lat: args.Lat, Lon: args.Lon}
	if err := input.Validate(); err != nil {
		return nil, models.ElevationResponse{}, err
	}
	resp, err := elevation.Lookup(ctx, input)
	if err != nil {
		return nil, models.ElevationResponse{}, err
	}
	if resp == nil {
		return nil, models.ElevationResponse{}, errors.New("empty response from elevation provider")
	}
	return nil, *resp, nil
}

func handleForecast(ctx context.Context, _ *mcp.CallToolRequest, args forecastArgs) (*mcp.CallToolResult, models.ForecastResponse, error) {
	input := models.ForecastInput{Lat: args.Lat, Lon: args.Lon, Hours: args.Hours}
	if err := input.Validate(); err != nil {
		return nil, models.ForecastResponse{}, err
	}
	resp, err := weather.Forecast(ctx, input)
	if err != nil {
		return nil, models.ForecastResponse{}, err
	}
	if resp == nil {
		return nil, models.ForecastResponse{}, errors.New("empty response from weather provider")
	}
	return nil, *resp, nil
}

func handleGeocode(ctx context.Context, _ *mcp.CallToolRequest, args geocodeArgs) (*mcp.CallToolResult, models.GeocodeResponse, error) {
	input := models.GeocodeInput{Query: args.Query, Limit: args.Limit, CountryCodes: args.CountryCodes}
	if err := input.Validate(); err != nil {
		return nil, models.GeocodeResponse{}, err
	}
	resp, err := geocode.Search(ctx, input)
	if err != nil {
		return nil, models.GeocodeResponse{}, err
	}
	if resp == nil {
		return nil, models.GeocodeResponse{}, errors.New("empty response from geocoder")
	}
	return nil, *resp, nil
}

func handlePlanHike(ctx context.Context, _ *mcp.CallToolRequest, args planArgs) (*mcp.CallToolResult, models.PlanHikeResponse, error) {
	input := models.PlanHikeInput{
		From: args.From, To: args.To, AlsoWater: args.AlsoWater,
		Include: args.Include, Hours: args.Hours, CorridorM: args.CorridorM, Limit: args.Limit,
	}
	if err := input.Validate(); err != nil {
		return nil, models.PlanHikeResponse{}, err
	}
	resp, err := plan.Plan(ctx, input)
	if err != nil {
		return nil, models.PlanHikeResponse{}, err
	}
	if resp == nil {
		return nil, models.PlanHikeResponse{}, errors.New("empty hike plan")
	}
	return nil, *resp, nil
}
