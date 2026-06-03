package flagcel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/picunada/flagcel/evalcore"
)

func TestProviderFetchesDefinitionsAndEvaluates(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("Authorization = %q, want Bearer secret", got)
		}
		w.Header().Set("ETag", `"v1"`)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "success",
			"data": definitionsFixture("checkout-copy", evalcore.ValueTypeString, "control", []evalcore.Rule{
				{
					ID:         "pro",
					Expression: `user.tier == "pro"`,
					Rollout:    evalcore.Rollout{Percentage: 100},
					Value:      "pro-copy",
				},
			}),
		})
	}))
	defer server.Close()

	provider, err := NewProvider(server.URL, "secret", WithPollInterval(time.Hour))
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if err := provider.Init(openfeature.EvaluationContext{}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer provider.Shutdown()

	detail := provider.StringEvaluation(context.Background(), "checkout-copy", "fallback", openfeature.FlattenedContext{
		"user": map[string]any{"tier": "pro"},
	})
	if detail.Value != "pro-copy" {
		t.Fatalf("Value = %q, want pro-copy", detail.Value)
	}
	if detail.Reason != openfeature.TargetingMatchReason {
		t.Fatalf("Reason = %q, want %q", detail.Reason, openfeature.TargetingMatchReason)
	}
	if detail.Variant != "pro" {
		t.Fatalf("Variant = %q, want pro", detail.Variant)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

func TestProviderHonorsETagAndKeepsLastKnownDefinitions(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch requests {
		case 1:
			w.Header().Set("ETag", `"v1"`)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "success",
				"data":    definitionsFixture("enabled", evalcore.ValueTypeBoolean, true, nil),
			})
		case 2:
			if got := r.Header.Get("If-None-Match"); got != `"v1"` {
				t.Fatalf("If-None-Match = %q, want %q", got, `"v1"`)
			}
			w.Header().Set("ETag", `"v1"`)
			w.WriteHeader(http.StatusNotModified)
		default:
			http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	provider, err := NewProvider(server.URL, "", WithPollInterval(time.Hour))
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if err := provider.Init(openfeature.EvaluationContext{}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer provider.Shutdown()

	if err := provider.refresh(context.Background()); err != nil {
		t.Fatalf("refresh() 304 error = %v", err)
	}
	if err := provider.refresh(context.Background()); err == nil {
		t.Fatal("refresh() error = nil, want fetch error")
	}

	detail := provider.BooleanEvaluation(context.Background(), "enabled", false, openfeature.FlattenedContext{})
	if !detail.Value {
		t.Fatal("Value = false, want last-known true")
	}
	if detail.ProviderResolutionDetail.Error() != nil {
		t.Fatalf("ResolutionError = %v, want empty for successful last-known evaluation", detail.ProviderResolutionDetail.Error())
	}
}

func TestProviderFallsBackToDefaultWhenInitialFetchFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no definitions", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider, err := NewProvider(server.URL, "", WithPollInterval(time.Hour))
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if err := provider.Init(openfeature.EvaluationContext{}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer provider.Shutdown()

	detail := provider.BooleanEvaluation(context.Background(), "missing", true, openfeature.FlattenedContext{})
	if !detail.Value {
		t.Fatal("Value = false, want default true")
	}
	if detail.ResolutionError.Error() == "" {
		t.Fatal("ResolutionError is empty, want provider not ready error")
	}
}

func TestProviderTypeMismatchFallsBackToDefault(t *testing.T) {
	provider := &Provider{}
	if err := provider.cache.store(definitionsFixture("title", evalcore.ValueTypeString, "hello", nil), `"v1"`); err != nil {
		t.Fatalf("store() error = %v", err)
	}

	detail := provider.BooleanEvaluation(context.Background(), "title", true, openfeature.FlattenedContext{})
	if !detail.Value {
		t.Fatal("Value = false, want default true")
	}
	if detail.ResolutionError.Error() == "" {
		t.Fatal("ResolutionError is empty, want type mismatch")
	}
}

func definitionsFixture(key string, valueType evalcore.ValueType, defaultValue any, rules []evalcore.Rule) evalcore.Definitions {
	return evalcore.Definitions{
		Flags: []evalcore.FlagDefinition{
			{
				FlagConfig: evalcore.FlagConfig{
					Key:          key,
					Type:         valueType,
					Enabled:      true,
					DefaultValue: defaultValue,
					Rules:        rules,
				},
				ContextSchema: &evalcore.ContextSchema{
					Fields: []evalcore.ContextField{
						{Path: "user.tier", Type: evalcore.ContextTypeString},
					},
				},
			},
		},
	}
}
