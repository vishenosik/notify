package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/vishenosik/notify/internal/api/rest/mapper"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/config"
)

const healthCheckPingTimeout = 2 * time.Second

func HealthCheckHandler(healthUseCase interfaces.HealthUseCase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := healthUseCase.Check(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}

		if verbose := r.URL.Query().Get("v") == "1"; verbose {
			render.JSON(w, r, mapper.ToHealthCheckResponseVerbose(resp))
			return
		}

		render.JSON(w, r, mapper.ToHealthCheckResponse(resp))
	})
}

func BuildInfoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, mapper.NewBuildInfoResponse(config.GetBuildInfo()))
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
