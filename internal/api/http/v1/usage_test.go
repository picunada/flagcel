package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUsageQueryFromRequestDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/environments/env-test/flags/checkout/usage", nil)
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	query := usageQueryFromRequest(req, "env-test", "checkout", now)

	if query.EnvironmentID != "env-test" {
		t.Fatalf("environment = %q, want env-test", query.EnvironmentID)
	}
	if query.FlagKey != "checkout" {
		t.Fatalf("flag key = %q, want checkout", query.FlagKey)
	}
	if !query.Since.Equal(now.Add(-24 * time.Hour)) {
		t.Fatalf("since = %s, want %s", query.Since, now.Add(-24*time.Hour))
	}
	if query.Limit != 50 {
		t.Fatalf("limit = %d, want 50", query.Limit)
	}
}
