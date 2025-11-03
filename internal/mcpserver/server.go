package mcpserver

import (
	"context"
	"errors"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"trail-finder-mcp/internal/elevation"
	"trail-finder-mcp/internal/models"
	"trail-finder-mcp/internal/overpass"
	"trail-finder-mcp/internal/router"
	"trail-finder-mcp/internal/weather"
)

const (
	serverName     = "trail-finder-mcp"
	defaultVersion = "0.1.0"
)

// Run starts the MCP server on stdio. This is the entry point used by Claude Code
// and other MCP-compatible clients.
func Run(ctx context.Context) error {
	server := NewServer()
	return server.Run(ctx, &mcp.StdioTransport{})
}

// NewServer instantiates an MCP server with all available Trail-Finder tools registered.
func NewServer() *mcp.Server {
	version := os.Getenv("TRAILFINDER_VERSION")
	if version == "" {
		version = defaultVersion
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "trailheads",
		Description: "Find nearby trailheads, guideposts, and optional water sources from OpenStreetMap (Overpass API).",
	}, handleTrailheads)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "route_foot",
		Description: "Plan a walking route between two coordinates using OSRM.",
	}, handleRouteFoot)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "elevation",
		Description: "Retrieve the elevation (meters above sea level) for a single coordinate.",
	}, handleElevation)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "forecast",
		Description: "Get short-term hourly weather forecasts from Open-Meteo for a coordinate.",
	}, handleForecast)

	return server
}

type trailheadsArgs struct {
	Lat       float64  `json:"lat" description:"Latitude in decimal degrees"`
	Lon       float64  `json:"lon" description:"Longitude in decimal degrees"`
	RadiusM   int      `json:"radius_m,omitempty" description:"Search radius in meters (default 2000)"`
	Include   []string `json:"include,omitempty" description:"Optional OSM categories to include (guidepost, trailhead)"`
	AlsoWater bool     `json:"also_water,omitempty" description:"Include nearby water sources (drinking water, springs)"`
	Limit     int      `json:"limit,omitempty" description:"Maximum number of POIs to return"`
}

type routeArgs struct {
	From    models.Coord   `json:"from" description:"Starting location (lat/lon)"`
	To      models.Coord   `json:"to" description:"Destination location (lat/lon)"`
	Engine  string         `json:"engine,omitempty" description:"Routing engine (auto, osrm, valhalla)"`
	Options map[string]any `json:"options,omitempty" description:"Optional routing flags (e.g. include_geometry)"`
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

func handleTrailheads(ctx context.Context, _ *mcp.CallToolRequest, args trailheadsArgs) (*mcp.CallToolResult, models.TrailheadsResponse, error) {
	input := models.TrailheadsInput{
		Lat:       args.Lat,
		Lon:       args.Lon,
		RadiusM:   args.RadiusM,
		Include:   args.Include,
		AlsoWater: args.AlsoWater,
		Limit:     args.Limit,
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
	input := models.RouteInput{
		From:    args.From,
		To:      args.To,
		Engine:  args.Engine,
		Options: args.Options,
	}
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
	input := models.ForecastInput{
		Lat:   args.Lat,
		Lon:   args.Lon,
		Hours: args.Hours,
	}
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
