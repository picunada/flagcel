package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/engine"
	"github.com/picunada/flagcel/internal/service"
)

type EvalHandler struct {
	service *service.EvalService
}

func NewEvalHandler(s *service.EvalService) *EvalHandler {
	return &EvalHandler{service: s}
}

func (h *EvalHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /eval", h.EvaluateAll)
	mux.HandleFunc("POST /eval/{key}", h.Evaluate)
}

func (h *EvalHandler) RegisterAdmin(mux *http.ServeMux) {
	mux.HandleFunc("POST /flags/{key}/evaluate", h.EvaluateFlag)
}

func (h *EvalHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	req, err := utils.Decode[EvalRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if len(req.Context) == 0 {
		WriteError(w, InvalidRequest("context is required"))
		return
	}

	value, err := h.service.Evaluate(r.Context(), key, engine.DataContext(req.Context))
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", EvalResponse{Key: key, Value: value}); err != nil {
		WriteError(w, err)
	}
}

func (h *EvalHandler) EvaluateFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	req, err := utils.Decode[EvalRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if len(req.Context) == 0 {
		WriteError(w, InvalidRequest("context is required"))
		return
	}

	trace, err := h.service.EvaluateWithTrace(r.Context(), key, engine.DataContext(req.Context))
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toEvalTraceResponse(trace)); err != nil {
		WriteError(w, err)
	}
}

func (h *EvalHandler) EvaluateAll(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[EvalRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if len(req.Context) == 0 {
		WriteError(w, InvalidRequest("context is required"))
		return
	}

	flags, err := h.service.EvaluateAll(r.Context(), engine.DataContext(req.Context))
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", EvalAllResponse{Flags: flags}); err != nil {
		WriteError(w, err)
	}
}
