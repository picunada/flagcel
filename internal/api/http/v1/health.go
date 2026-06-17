package v1

import (
	"context"
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
)

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	readiness ReadinessChecker
}

type HealthResponse struct {
	Status string `json:"status"`
}

func NewHealthHandler(readiness ReadinessChecker) *HealthHandler {
	return &HealthHandler{readiness: readiness}
}

func (h *HealthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Check)
	mux.HandleFunc("GET /readyz", h.Ready)
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	if err := utils.Encode(w, r, http.StatusOK, "success", HealthResponse{Status: "ok"}); err != nil {
		WriteError(w, err)
	}
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.readiness == nil || h.readiness.Ping(r.Context()) != nil {
		WriteError(w, ErrNotReady)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", HealthResponse{Status: "ok"}); err != nil {
		WriteError(w, err)
	}
}
