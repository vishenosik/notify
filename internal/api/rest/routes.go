package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	interfaces "github.com/vishenosik/notify/internal/app/interface"
)

func Routes(
	healthUseCase interfaces.HealthUseCase,
) http.Handler {
	r := chi.NewRouter()
	r.Method(http.MethodGet, "/healthz", HealthCheckHandler(healthUseCase))
	r.Method(http.MethodGet, "/version", BuildInfoHandler())
	return r
}
