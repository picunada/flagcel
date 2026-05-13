package v1

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/picunada/flagcel/internal/core"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
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
	switch {
	case errors.Is(err, core.ErrFlagNotFound):
		return ErrFlagNotFound
	case errors.Is(err, core.ErrRuleNotFound):
		return ErrRuleNotFound
	default:
		return ErrInternal
	}
}
