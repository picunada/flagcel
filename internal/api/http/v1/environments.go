package v1

import (
	"errors"
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

type EnvironmentsHandler struct {
	service *service.EnvironmentService
}

func NewEnvironmentsHandler(service *service.EnvironmentService) *EnvironmentsHandler {
	return &EnvironmentsHandler{service: service}
}

func (h *EnvironmentsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /environments", h.List)
	mux.HandleFunc("POST /environments", h.Create)
	mux.HandleFunc("GET /environments/{id}", h.Get)
	mux.HandleFunc("PUT /environments/{id}", h.Update)
	mux.HandleFunc("DELETE /environments/{id}", h.Delete)
}

func (h *EnvironmentsHandler) List(w http.ResponseWriter, r *http.Request) {
	envs, err := h.service.List(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toEnvironmentResponses(envs)); err != nil {
		WriteError(w, err)
	}
}

func (h *EnvironmentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	env, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toEnvironmentResponse(*env)); err != nil {
		WriteError(w, err)
	}
}

func (h *EnvironmentsHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[CreateEnvironmentRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	env := toCoreEnvironment("", req)
	if err := h.service.Create(r.Context(), &env); err != nil {
		WriteError(w, environmentWriteError(err))
		return
	}
	saved, err := h.service.Get(r.Context(), env.ID)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toEnvironmentResponse(*saved)); err != nil {
		WriteError(w, err)
	}
}

func (h *EnvironmentsHandler) Update(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[UpdateEnvironmentRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	env := toCoreEnvironment(r.PathValue("id"), req)
	if err := h.service.Update(r.Context(), &env); err != nil {
		WriteError(w, environmentWriteError(err))
		return
	}
	saved, err := h.service.Get(r.Context(), env.ID)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toEnvironmentResponse(*saved)); err != nil {
		WriteError(w, err)
	}
}

func (h *EnvironmentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func environmentWriteError(err error) error {
	switch {
	case errors.Is(err, core.ErrEnvironmentNotFound),
		errors.Is(err, core.ErrEnvironmentKeyTaken),
		errors.Is(err, core.ErrEnvironmentInUse),
		errors.Is(err, core.ErrDefaultEnvironment):
		return err
	default:
		return InvalidRequest(err.Error())
	}
}
