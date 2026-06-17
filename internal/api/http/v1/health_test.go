package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	router := NewRouter(&Handlers{
		Health: NewHealthHandler(nil),
		Auth:   &AuthHandler{},
	})

	for _, path := range []string{"/healthz", "/api/v1/healthz"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}

			var body struct {
				Message string         `json:"message"`
				Data    HealthResponse `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Message != "success" || body.Data.Status != "ok" {
				t.Fatalf("body = %+v, want success status ok", body)
			}
		})
	}
}

func TestReadinessRoutes(t *testing.T) {
	router := NewRouter(&Handlers{
		Health: NewHealthHandler(readinessCheckerFunc(func(context.Context) error {
			return nil
		})),
		Auth: &AuthHandler{},
	})

	for _, path := range []string{"/readyz", "/api/v1/readyz"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			var body struct {
				Message string         `json:"message"`
				Data    HealthResponse `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Message != "success" || body.Data.Status != "ok" {
				t.Fatalf("body = %+v, want success status ok", body)
			}
		})
	}
}

func TestReadinessRouteReturnsUnavailable(t *testing.T) {
	router := NewRouter(&Handlers{
		Health: NewHealthHandler(readinessCheckerFunc(func(context.Context) error {
			return errors.New("db unavailable")
		})),
		Auth: &AuthHandler{},
	})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error == nil || body.Error.Code != ErrNotReady.Code {
		t.Fatalf("error = %+v, want %s", body.Error, ErrNotReady.Code)
	}
}

type readinessCheckerFunc func(context.Context) error

func (f readinessCheckerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}
