package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/elevation"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/geocode"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/overpass"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/plan"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/router"
	"github.com/Takamasa045/Trail-Finder-MCP/internal/weather"
)

const maxRequestBodyBytes int64 = 1 << 20

func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/tools/trailheads", handleTool(func(ctx context.Context, in models.TrailheadsInput) (*models.TrailheadsResponse, error) {
		return overpass.QueryPOIs(ctx, in)
	}))
	mux.HandleFunc("/tools/route_foot", handleTool(func(ctx context.Context, in models.RouteInput) (*models.RouteResponse, error) {
		return router.RouteFoot(ctx, in)
	}))
	mux.HandleFunc("/tools/elevation", handleTool(func(ctx context.Context, in models.ElevationInput) (*models.ElevationResponse, error) {
		return elevation.Lookup(ctx, in)
	}))
	mux.HandleFunc("/tools/forecast", handleTool(func(ctx context.Context, in models.ForecastInput) (*models.ForecastResponse, error) {
		return weather.Forecast(ctx, in)
	}))
	mux.HandleFunc("/tools/geocode", handleTool(func(ctx context.Context, in models.GeocodeInput) (*models.GeocodeResponse, error) {
		return geocode.Search(ctx, in)
	}))
	mux.HandleFunc("/tools/plan_hike", handleTool(func(ctx context.Context, in models.PlanHikeInput) (*models.PlanHikeResponse, error) {
		return plan.Plan(ctx, in)
	}))
	return mux
}

func handleTool[In any, Out any](fn func(context.Context, In) (*Out, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		defer r.Body.Close()
		var in In
		if err := decodeJSONBody(r.Body, &in); err != nil {
			badRequest(w, "invalid json: "+err.Error())
			return
		}
		if v, ok := any(&in).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				badRequest(w, err.Error())
				return
			}
		}
		resp, err := fn(r.Context(), in)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, models.ErrorResponse{
				Error: models.ErrorBody{Code: "UPSTREAM_ERROR", Message: err.Error(), Retryable: true},
			})
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
		Error: models.ErrorBody{Code: "BAD_REQUEST", Message: msg, Retryable: false},
	})
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
		Error: models.ErrorBody{
			Code:      "METHOD_NOT_ALLOWED",
			Message:   "only " + allowed + " is supported on this endpoint",
			Retryable: false,
		},
	})
}

func decodeJSONBody(body io.Reader, dst any) error {
	dec := json.NewDecoder(io.LimitReader(body, maxRequestBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON object")
		}
		return err
	}
	return nil
}
