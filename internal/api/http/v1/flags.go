package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/service"
)

type FlagsHandler struct {
	service      *service.FlagService
	environments *service.EnvironmentService
}

func NewFlagsHandler(service *service.FlagService, environments *service.EnvironmentService) *FlagsHandler {
	return &FlagsHandler{
		service:      service,
		environments: environments,
	}
}

func (h *FlagsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /flags", h.GetFlags)
	mux.HandleFunc("POST /flags", h.CreateFlag)
	mux.HandleFunc("GET /flags/{key}", h.GetFlag)
	mux.HandleFunc("DELETE /flags/{key}", h.DeleteFlag)
	mux.HandleFunc("GET /flags/{key}/audit", h.GetFlagHistory)
	mux.HandleFunc("GET /environments/{environment_id}/flags", h.GetFlags)
	mux.HandleFunc("POST /environments/{environment_id}/flags", h.CreateFlag)
	mux.HandleFunc("GET /environments/{environment_id}/flags/{key}", h.GetFlag)
	mux.HandleFunc("DELETE /environments/{environment_id}/flags/{key}", h.DeleteFlag)
	mux.HandleFunc("GET /environments/{environment_id}/flags/{key}/audit", h.GetFlagHistory)
}

func (h *FlagsHandler) GetFlags(w http.ResponseWriter, r *http.Request) {
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	flags, err := h.service.GetFlags(r.Context(), environmentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagResponses(flags)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *FlagsHandler) GetFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	flag, err := h.service.GetFlag(r.Context(), environmentID, key)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagResponse(*flag)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *FlagsHandler) GetFlagHistory(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	entries, err := h.service.GetFlagHistory(r.Context(), environmentID, key)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toAuditEntryResponses(entries)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *FlagsHandler) CreateFlag(w http.ResponseWriter, r *http.Request) {
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	req, err := utils.Decode[CreateFlagRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if req.Key == "" {
		WriteError(w, InvalidRequest("key is required"))
		return
	}

	flag, err := toCoreFlag(req)
	if err != nil {
		WriteError(w, InvalidRequest(err.Error()))
		return
	}

	if err := h.service.CreateFlag(r.Context(), environmentID, &flag); err != nil {
		WriteError(w, err)
		return
	}

	saved, err := h.service.GetFlag(r.Context(), environmentID, flag.Key)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagResponse(*saved)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *FlagsHandler) DeleteFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := h.service.DeleteFlag(r.Context(), environmentID, key); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FlagsHandler) environmentID(r *http.Request) (string, error) {
	if id := r.PathValue("environment_id"); id != "" {
		return id, nil
	}
	env, err := h.environments.Default(r.Context())
	if err != nil {
		return "", err
	}
	return env.ID, nil
}
