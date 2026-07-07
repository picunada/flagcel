package v1

import (
	"net/http"
	"strings"

	"github.com/picunada/flagcel/evalcore"
	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

type EvalHandler struct {
	service *service.EvalService
}

func NewEvalHandler(s *service.EvalService) *EvalHandler {
	return &EvalHandler{service: s}
}

func (h *EvalHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /eval/definitions", h.GetDefinitions)
	mux.HandleFunc("POST /eval", h.EvaluateAll)
	mux.HandleFunc("POST /eval/{key}", h.Evaluate)
}

func (h *EvalHandler) RegisterAdmin(mux *http.ServeMux) {
	mux.HandleFunc("POST /environments/{environment_id}/flags/{key}/evaluate", h.EvaluateFlag)
}

func (h *EvalHandler) GetDefinitions(w http.ResponseWriter, r *http.Request) {
	environmentID, err := environmentIDFromAPIKey(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	etag := h.service.DefinitionsETag()
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		setDefinitionsHeaders(w, etag)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	defs, etag, err := h.service.Definitions(r.Context(), environmentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	setDefinitionsHeaders(w, etag)
	if err := utils.Encode(w, r, http.StatusOK, "success", defs); err != nil {
		WriteError(w, err)
	}
}

func setDefinitionsHeaders(w http.ResponseWriter, etag string) {
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")
}

func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}
		if strings.HasPrefix(candidate, "W/") {
			candidate = strings.TrimSpace(strings.TrimPrefix(candidate, "W/"))
		}
		if candidate == etag {
			return true
		}
	}
	return false
}

func (h *EvalHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	environmentID, err := environmentIDFromAPIKey(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	req, err := utils.Decode[EvalRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if len(req.Context) == 0 {
		WriteError(w, InvalidRequest("context is required"))
		return
	}

	value, err := h.service.EvaluateWithUsage(r.Context(), environmentID, key, evalcore.DataContext(req.Context), usageSourceFromRequest(r))
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
	environmentID := r.PathValue("environment_id")
	if environmentID == "" {
		WriteError(w, InvalidRequest("environment_id is required"))
		return
	}

	req, err := utils.Decode[EvalRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if len(req.Context) == 0 {
		WriteError(w, InvalidRequest("context is required"))
		return
	}

	trace, err := h.service.EvaluateWithTrace(r.Context(), environmentID, key, evalcore.DataContext(req.Context))
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toEvalTraceResponse(trace)); err != nil {
		WriteError(w, err)
	}
}

func (h *EvalHandler) EvaluateAll(w http.ResponseWriter, r *http.Request) {
	environmentID, err := environmentIDFromAPIKey(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	req, err := utils.Decode[EvalRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if len(req.Context) == 0 {
		WriteError(w, InvalidRequest("context is required"))
		return
	}

	flags, err := h.service.EvaluateAllWithUsage(r.Context(), environmentID, evalcore.DataContext(req.Context), usageSourceFromRequest(r))
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

func usageSourceFromRequest(r *http.Request) core.FlagUsageSource {
	source := core.FlagUsageSource{Source: r.UserAgent()}
	if key, ok := apiKeyFromRequest(r); ok && key != nil {
		source.APIKeyID = key.ID
	}
	return source
}

func environmentIDFromAPIKey(r *http.Request) (string, error) {
	key, ok := apiKeyFromRequest(r)
	if !ok || key.EnvironmentID == "" {
		return "", ErrUnauthorized
	}
	return key.EnvironmentID, nil
}
