package httpserver

import (
	"encoding/json"
	"net/http"

	"trail-finder-mcp/internal/models"
	"trail-finder-mcp/internal/overpass"
	"trail-finder-mcp/internal/router"
	"trail-finder-mcp/internal/elevation"
	"trail-finder-mcp/internal/weather"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/tools/trailheads", handleTrailheads)
	mux.HandleFunc("/tools/route_foot", handleRouteFoot)
	mux.HandleFunc("/tools/elevation", handleElevation)
	mux.HandleFunc("/tools/forecast", handleForecast)
	return mux
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

func handleTrailheads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in models.TrailheadsInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json: "+err.Error())
		return
	}
	if err := in.Validate(); err != nil {
		badRequest(w, err.Error())
		return
	}
	resp, err := overpass.QueryPOIs(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{
			Error: models.ErrorBody{Code: "UPSTREAM_ERROR", Message: err.Error(), Retryable: true},
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleRouteFoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in models.RouteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json: "+err.Error())
		return
	}
	if err := in.Validate(); err != nil {
		badRequest(w, err.Error())
		return
	}
	resp, err := router.RouteFoot(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{
			Error: models.ErrorBody{Code: "UPSTREAM_ERROR", Message: err.Error(), Retryable: true},
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleElevation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in models.ElevationInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json: "+err.Error())
		return
	}
	if err := in.Validate(); err != nil {
		badRequest(w, err.Error())
		return
	}
	resp, err := elevation.Lookup(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{
			Error: models.ErrorBody{Code: "UPSTREAM_ERROR", Message: err.Error(), Retryable: true},
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleForecast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in models.ForecastInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json: "+err.Error())
		return
	}
	if err := in.Validate(); err != nil {
		badRequest(w, err.Error())
		return
	}
	resp, err := weather.Forecast(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{
			Error: models.ErrorBody{Code: "UPSTREAM_ERROR", Message: err.Error(), Retryable: true},
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
