package v1

import (
	"net/http"
	"time"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

type APIKeysHandler struct {
	service *service.AuthService
}

func NewAPIKeysHandler(s *service.AuthService) *APIKeysHandler {
	return &APIKeysHandler{service: s}
}

func (h *APIKeysHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api-keys", h.List)
	mux.HandleFunc("POST /api-keys", h.Create)
	mux.HandleFunc("DELETE /api-keys/{id}", h.Revoke)
}

func (h *APIKeysHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.service.ListAPIKeys(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}
	out := make([]APIKeyResponse, len(keys))
	for i, key := range keys {
		out[i] = toAPIKeyResponse(key)
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", out); err != nil {
		WriteError(w, err)
	}
}

func (h *APIKeysHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[CreateAPIKeyRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	created, err := h.service.CreateAPIKey(r.Context(), req.Name)
	if err != nil {
		WriteError(w, err)
		return
	}
	resp := CreateAPIKeyResponse{
		APIKeyResponse: toAPIKeyResponse(created.Key),
		Token:          created.Token,
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", resp); err != nil {
		WriteError(w, err)
	}
}

func (h *APIKeysHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RevokeAPIKey(r.Context(), r.PathValue("id")); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toUserResponse(user *core.User) *UserResponse {
	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Admin: user.Admin,
	}
}

func toAPIKeyResponse(key *core.APIKey) APIKeyResponse {
	return APIKeyResponse{
		ID:         key.ID,
		Name:       key.Name,
		Prefix:     key.Prefix,
		CreatedAt:  formatTime(key.CreatedAt),
		LastUsedAt: formatTimePtr(key.LastUsedAt),
		RevokedAt:  formatTimePtr(key.RevokedAt),
	}
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	out := formatTime(*t)
	return &out
}
