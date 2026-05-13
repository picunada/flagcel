package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/service"
)

type FlagsHandler struct {
	service *service.FlagService
}

func NewFlagsHandler(service *service.FlagService) *FlagsHandler {
	return &FlagsHandler{
		service: service,
	}
}

func (h *FlagsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /flags", h.GetFlags)
	mux.HandleFunc("POST /flags", h.CreateFlag)
	mux.HandleFunc("GET /flags/{key}", h.GetFlag)
	mux.HandleFunc("DELETE /flags/{key}", h.DeleteFlag)
}

func (h *FlagsHandler) GetFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.service.GetFlags(r.Context())
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

	flag, err := h.service.GetFlag(r.Context(), key)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagResponse(*flag)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *FlagsHandler) CreateFlag(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[CreateFlagRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if req.Key == "" {
		WriteError(w, InvalidRequest("key is required"))
		return
	}

	flag := toCoreFlag(req)

	if err := h.service.CreateFlag(r.Context(), &flag); err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagResponse(flag)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *FlagsHandler) DeleteFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if err := h.service.DeleteFlag(r.Context(), key); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
