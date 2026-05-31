package v1

import (
	"net/http"

	"github.com/picunada/flagcel/evalcore"
	"github.com/picunada/flagcel/internal/api/http/utils"
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

	value, err := h.service.Evaluate(r.Context(), key, evalcore.DataContext(req.Context))
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", EvalResponse{
		Key:       key,
		ValueType: string(value.Type),
		Value:     value.Value,
	}); err != nil {
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

	trace, err := h.service.EvaluateWithTrace(r.Context(), key, evalcore.DataContext(req.Context))
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

	flags, err := h.service.EvaluateAll(r.Context(), evalcore.DataContext(req.Context))
	if err != nil {
		WriteError(w, err)
		return
	}

	out := make(map[string]EvalFlagValueResponse, len(flags))
	for key, value := range flags {
		out[key] = toEvalFlagValueResponse(value)
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", EvalAllResponse{Flags: out}); err != nil {
		WriteError(w, err)
	}
}
