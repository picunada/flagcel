package v1

import (
	"fmt"
	"log/slog"
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
	mux.HandleFunc("GET /", h.GetFlags)
	mux.HandleFunc("POST /", h.CreateFlag)
	mux.HandleFunc("DELETE /", h.DeleteFlag)
}

func (h *FlagsHandler) GetFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.service.GetFlags(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err)
		slog.Error(fmt.Sprintf("failed to get flags: %v", err))
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagResponses(flags)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *FlagsHandler) CreateFlag(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[CreateFlagRequest](r)

	if err != nil {
		utils.Error(w, http.StatusUnprocessableEntity, err)
		slog.Error(fmt.Sprintf("failed to validate create flag: %v", err))
		return
	}

	flag := toCoreFlag(req)

	if err := h.service.CreateFlag(r.Context(), &flag); err != nil {
		utils.Error(w, http.StatusInternalServerError, err)
		slog.Error(fmt.Sprintf("failed to create flag: %v", err))
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagResponse(flag)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *FlagsHandler) DeleteFlag(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("key")
	if id == "" {
		utils.Error(w, http.StatusBadRequest, fmt.Errorf("id is required"))
		slog.Error("id is required")
		return
	}

	if err := h.service.DeleteFlag(r.Context(), id); err != nil {
		utils.Error(w, http.StatusInternalServerError, err)
		slog.Error(fmt.Sprintf("failed to delete flag: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
