package v1

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/picunada/flagcel/internal/core"
)

type APIError struct {
	Status  int                    `json:"-"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details []core.ValidationIssue `json:"details,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

var (
	ErrFlagNotFound = &APIError{
		Status:  http.StatusNotFound,
		Code:    "FLAG_NOT_FOUND",
		Message: "Flag not found",
	}
	ErrRuleNotFound = &APIError{
		Status:  http.StatusNotFound,
		Code:    "RULE_NOT_FOUND",
		Message: "Rule not found",
	}
	ErrContextNotFound = &APIError{
		Status:  http.StatusNotFound,
		Code:    "CONTEXT_NOT_FOUND",
		Message: "Context not found",
	}
	ErrContextNameTaken = &APIError{
		Status:  http.StatusConflict,
		Code:    "CONTEXT_NAME_TAKEN",
		Message: "Context name already in use",
	}
	ErrUnauthorized = &APIError{
		Status:  http.StatusUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: "Authentication required",
	}
	ErrInvalidCredentials = &APIError{
		Status:  http.StatusUnauthorized,
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid email or password",
	}
	ErrForbidden = &APIError{
		Status:  http.StatusForbidden,
		Code:    "FORBIDDEN",
		Message: "Access denied",
	}
	ErrAPIKeyNotFound = &APIError{
		Status:  http.StatusNotFound,
		Code:    "API_KEY_NOT_FOUND",
		Message: "API key not found",
	}
	ErrAuthNotConfigured = &APIError{
		Status:  http.StatusServiceUnavailable,
		Code:    "AUTH_NOT_CONFIGURED",
		Message: "Auth is not configured",
	}
	ErrInternal = &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
	}
)

func InvalidRequest(message string) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    "INVALID_REQUEST",
		Message: message,
	}
}

func BadRequest(message string) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

type errorEnvelope struct {
	Error *APIError `json:"error"`
}

func WriteError(w http.ResponseWriter, err error) {
	apiErr := toAPIError(err)
	if apiErr.Status >= 500 {
		slog.Error("api error", "code", apiErr.Code, "err", err)
	} else {
		slog.Debug("api error", "code", apiErr.Code, "err", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: apiErr})
}

func toAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	var validationErr *core.ValidationError
	if errors.As(err, &validationErr) {
		return &APIError{
			Status:  http.StatusUnprocessableEntity,
			Code:    "RULE_VALIDATION_FAILED",
			Message: validationErr.Error(),
			Details: validationErr.Issues,
		}
	}
	switch {
	case errors.Is(err, core.ErrFlagNotFound):
		return ErrFlagNotFound
	case errors.Is(err, core.ErrRuleNotFound):
		return ErrRuleNotFound
	case errors.Is(err, core.ErrContextNotFound):
		return ErrContextNotFound
	case errors.Is(err, core.ErrContextNameTaken):
		return ErrContextNameTaken
	case errors.Is(err, core.ErrUserNotAllowed):
		return ErrForbidden
	case errors.Is(err, core.ErrInvalidCredentials):
		return ErrInvalidCredentials
	case errors.Is(err, core.ErrSessionNotFound):
		return ErrUnauthorized
	case errors.Is(err, core.ErrAPIKeyNotFound):
		return ErrAPIKeyNotFound
	case errors.Is(err, core.ErrAuthNotConfigured):
		return ErrAuthNotConfigured
	default:
		return ErrInternal
	}
}
