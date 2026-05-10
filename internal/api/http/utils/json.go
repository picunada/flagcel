package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Encode[T any](w http.ResponseWriter, r *http.Request, status int, message string, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(SuccessResponse{
		Message: message,
		Data:    v,
	})
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}

func Error(w http.ResponseWriter, rerr int, err error) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(ErrorResponse{Error: err.Error()})
	if err != nil {
		http.Error(w, err.Error(), rerr)
		return
	}
	w.Write(data)
}
