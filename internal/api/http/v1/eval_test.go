package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

func TestGetDefinitionsHonorsIfNoneMatch(t *testing.T) {
	svc := service.NewEvalService(nil, nil)
	handler := NewEvalHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/eval/definitions", nil)
	req = req.WithContext(context.WithValue(req.Context(), apiKeyContextKey{}, &core.APIKey{
		EnvironmentID: "env-test",
	}))
	req.Header.Set("If-None-Match", svc.DefinitionsETag())
	rec := httptest.NewRecorder()

	handler.GetDefinitions(rec, req)

	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotModified)
	}
	if got := rec.Header().Get("ETag"); got != svc.DefinitionsETag() {
		t.Fatalf("ETag = %q, want %q", got, svc.DefinitionsETag())
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("body = %q, want empty", body)
	}
}

func TestUsageSourceFromRequestUsesAPIKeyAndUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/eval/checkout-copy", nil)
	req.Header.Set("User-Agent", "flagcel-js/0.1")
	req = req.WithContext(context.WithValue(req.Context(), apiKeyContextKey{}, &core.APIKey{
		ID:            "api-key-1",
		EnvironmentID: "env-test",
	}))

	source := usageSourceFromRequest(req)

	if source.APIKeyID != "api-key-1" {
		t.Fatalf("api key id = %q, want api-key-1", source.APIKeyID)
	}
	if source.Source != "flagcel-js/0.1" {
		t.Fatalf("source = %q, want flagcel-js/0.1", source.Source)
	}
}

func TestETagMatches(t *testing.T) {
	for _, tc := range []struct {
		name   string
		header string
		etag   string
		want   bool
	}{
		{name: "exact", header: `"3"`, etag: `"3"`, want: true},
		{name: "list", header: `"1", "3"`, etag: `"3"`, want: true},
		{name: "weak", header: `W/"3"`, etag: `"3"`, want: true},
		{name: "wildcard", header: `*`, etag: `"3"`, want: true},
		{name: "mismatch", header: `"2"`, etag: `"3"`, want: false},
		{name: "empty", header: ``, etag: `"3"`, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := etagMatches(tc.header, tc.etag); got != tc.want {
				t.Fatalf("etagMatches(%q, %q) = %t, want %t", tc.header, tc.etag, got, tc.want)
			}
		})
	}
}
