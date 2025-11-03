package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"trail-finder-mcp/internal/elevation"
	"trail-finder-mcp/internal/models"
	"trail-finder-mcp/internal/overpass"
	"trail-finder-mcp/internal/router"
	"trail-finder-mcp/internal/weather"
)

const maxRequestBodyBytes int64 = 1 << 20

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

func handleTrailheads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	defer r.Body.Close()
	var in models.TrailheadsInput
	if err := decodeJSONBody(r.Body, &in); err != nil {
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
		methodNotAllowed(w, http.MethodPost)
		return
	}
	defer r.Body.Close()
	var in models.RouteInput
	if err := decodeJSONBody(r.Body, &in); err != nil {
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
		methodNotAllowed(w, http.MethodPost)
		return
	}
	defer r.Body.Close()
	var in models.ElevationInput
	if err := decodeJSONBody(r.Body, &in); err != nil {
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
		methodNotAllowed(w, http.MethodPost)
		return
	}
	defer r.Body.Close()
	var in models.ForecastInput
	if err := decodeJSONBody(r.Body, &in); err != nil {
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
