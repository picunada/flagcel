package engine

import (
	"testing"

	"github.com/picunada/flagcel/internal/core"
)

func TestEvaluateDisabledFlagReturnsDefaultValue(t *testing.T) {
	e := &Engine{}

	tests := []struct {
		name         string
		defaultValue bool
		want         bool
	}{
		{
			name:         "default false",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "default true",
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Evaluate(&Flag{
				Key:          "feature-a",
				Enabled:      false,
				DefaultValue: tt.defaultValue,
				Rules: []CompiledRule{
					{
						Rollout: Rollout{Percentage: 100, BucketBy: "id"},
					},
				},
			}, DataContext{
				"id": "user-123",
			})

			if got != tt.want {
				t.Fatalf("Evaluate() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestEvaluateConfigReturnsMatchedRuleAndBucketTrace(t *testing.T) {
	env, err := NewCELEnv()
	if err != nil {
		t.Fatalf("new cel env: %v", err)
	}
	e := NewEngine(env)

	trace := e.EvaluateConfigForContext(testFlagConfig(), testContextSchema(), DataContext{
		"user": map[string]any{
			"id":      "user-123",
			"country": "US",
		},
	})

	if trace.Reason != "matched_rule" {
		t.Fatalf("Reason = %q, want matched_rule", trace.Reason)
	}
	if trace.MatchedRule == nil || trace.MatchedRule.ID != "rule-a" {
		t.Fatalf("MatchedRule = %#v, want rule-a", trace.MatchedRule)
	}
	if trace.Bucket == nil {
		t.Fatal("expected bucket trace")
	}
	if trace.Bucket.BucketBy != "user.id" {
		t.Fatalf("BucketBy = %q, want user.id", trace.Bucket.BucketBy)
	}
	if trace.Bucket.BucketValue != "user-123" {
		t.Fatalf("BucketValue = %q, want user-123", trace.Bucket.BucketValue)
	}
	if trace.Bucket.BucketNumber == nil {
		t.Fatal("expected bucket number")
	}
}

func TestEvaluateConfigReturnsCELError(t *testing.T) {
	env, err := NewCELEnv()
	if err != nil {
		t.Fatalf("new cel env: %v", err)
	}
	e := NewEngine(env)

	cfg := testFlagConfig()
	cfg.Rules[0].Expression = `user.country ==`

	trace := e.EvaluateConfig(cfg, DataContext{
		"user": map[string]any{"country": "US"},
	})

	if trace.Reason != "cel_error" {
		t.Fatalf("Reason = %q, want cel_error", trace.Reason)
	}
	if trace.Error == "" {
		t.Fatal("expected CEL error")
	}
	if len(trace.RuleResults) != 1 || trace.RuleResults[0].Error == "" {
		t.Fatalf("RuleResults = %#v, want first rule error", trace.RuleResults)
	}
}

func TestCompileFlagForContextDeclaresContextRoots(t *testing.T) {
	env, err := NewCELEnv()
	if err != nil {
		t.Fatalf("new cel env: %v", err)
	}
	e := NewEngine(env)

	cfg := core.FlagConfig{
		Key:          "feature-a",
		Enabled:      true,
		DefaultValue: false,
		Rules: []core.Rule{
			{
				ID:         "rule-a",
				Expression: `request.path == "/checkout"`,
				Rollout:    core.Rollout{Percentage: 100},
			},
		},
	}

	if _, err := e.CompileFlag(cfg.Key, cfg); err == nil {
		t.Fatal("expected base env compile to reject undeclared request root")
	}

	compiled, err := e.CompileFlagForContext(cfg.Key, cfg, &core.ContextSchema{
		ID: "context-a",
		Fields: []core.ContextField{
			{Path: "request.path", Type: core.ContextTypeString},
		},
	})
	if err != nil {
		t.Fatalf("compile for context: %v", err)
	}

	got := e.Evaluate(compiled, DataContext{
		"request": map[string]any{"path": "/checkout"},
	})
	if !got {
		t.Fatal("Evaluate() = false, want true")
	}
}

func testFlagConfig() core.FlagConfig {
	return core.FlagConfig{
		Key:          "feature-a",
		Enabled:      true,
		DefaultValue: false,
		Rules: []core.Rule{
			{
				ID:         "rule-a",
				Expression: `user.country == "US"`,
				Rollout: core.Rollout{
					Percentage: 100,
					BucketBy:   "user.id",
				},
			},
		},
	}
}

func testContextSchema() *core.ContextSchema {
	return &core.ContextSchema{
		ID: "context-a",
		Fields: []core.ContextField{
			{Path: "user.id", Type: core.ContextTypeString},
			{Path: "user.country", Type: core.ContextTypeString},
		},
	}
}
