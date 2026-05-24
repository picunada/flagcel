package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

type ContextsHandler struct {
	service *service.ContextService
}

func NewContextsHandler(s *service.ContextService) *ContextsHandler {
	return &ContextsHandler{service: s}
}

func (h *ContextsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /contexts", h.ListContexts)
	mux.HandleFunc("POST /contexts", h.CreateContext)
	mux.HandleFunc("GET /contexts/{id}", h.GetContext)
	mux.HandleFunc("PUT /contexts/{id}", h.UpdateContext)
	mux.HandleFunc("DELETE /contexts/{id}", h.DeleteContext)
}

func (h *ContextsHandler) ListContexts(w http.ResponseWriter, r *http.Request) {
	cs, err := h.service.ListContexts(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toContextResponses(cs)); err != nil {
		WriteError(w, err)
	}
}

func (h *ContextsHandler) GetContext(w http.ResponseWriter, r *http.Request) {
	c, err := h.service.GetContext(r.Context(), r.PathValue("id"))
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toContextResponse(*c)); err != nil {
		WriteError(w, err)
	}
}

func (h *ContextsHandler) CreateContext(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[CreateContextRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		WriteError(w, InvalidRequest("name is required"))
		return
	}
	if err := validateFields(req.Fields); err != nil {
		WriteError(w, err)
		return
	}
	c := toCoreContext("", req.Name, req.Description, req.Fields)
	saved, err := h.service.CreateContext(r.Context(), &c)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toContextResponse(*saved)); err != nil {
		WriteError(w, err)
	}
}

func (h *ContextsHandler) UpdateContext(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	req, err := utils.Decode[UpdateContextRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		WriteError(w, InvalidRequest("name is required"))
		return
	}
	if err := validateFields(req.Fields); err != nil {
		WriteError(w, err)
		return
	}
	c := toCoreContext(id, req.Name, req.Description, req.Fields)
	saved, err := h.service.UpdateContext(r.Context(), &c)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toContextResponse(*saved)); err != nil {
		WriteError(w, err)
	}
}

func (h *ContextsHandler) DeleteContext(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteContext(r.Context(), r.PathValue("id")); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateFields(fields []ContextFieldDTO) *APIError {
	for _, f := range fields {
		if f.Path == "" {
			continue
		}
		if !core.ContextType(f.Type).Valid() {
			return InvalidRequest("invalid field type: " + f.Type)
		}
	}
	return nil
}
