package engine

import (
	"encoding/json"
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

			if got.Type != core.ValueTypeBoolean || got.Value != tt.want {
				t.Fatalf("Evaluate() = %#v, want boolean %t", got, tt.want)
			}
		})
	}
}

func TestEvaluateReturnsTypedDefaultsAndRuleValues(t *testing.T) {
	env, err := NewCELEnv()
	if err != nil {
		t.Fatalf("new cel env: %v", err)
	}
	e := NewEngine(env)

	tests := []struct {
		name         string
		valueType    core.ValueType
		defaultValue any
		ruleValue    any
		want         any
	}{
		{name: "boolean rule value", valueType: core.ValueTypeBoolean, defaultValue: false, ruleValue: true, want: true},
		{name: "string rule value", valueType: core.ValueTypeString, defaultValue: "control", ruleValue: "variant-a", want: "variant-a"},
		{name: "number rule value", valueType: core.ValueTypeNumber, defaultValue: float64(0), ruleValue: float64(42.5), want: float64(42.5)},
		{name: "json rule value", valueType: core.ValueTypeJSON, defaultValue: map[string]any{"name": "control"}, ruleValue: map[string]any{"name": "variant"}, want: map[string]any{"name": "variant"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := e.CompileFlag("feature-a", core.FlagConfig{
				Key:          "feature-a",
				Type:         tt.valueType,
				Enabled:      true,
				DefaultValue: tt.defaultValue,
				Rules: []core.Rule{
					{
						Expression: `true`,
						Rollout:    core.Rollout{Percentage: 100},
						Value:      tt.ruleValue,
					},
				},
			})
			if err != nil {
				t.Fatalf("compile flag: %v", err)
			}

			got := e.Evaluate(compiled, DataContext{})
			if got.Type != tt.valueType {
				t.Fatalf("Type = %q, want %q", got.Type, tt.valueType)
			}
			if !valuesEqual(got.Value, tt.want) {
				t.Fatalf("Value = %#v, want %#v", got.Value, tt.want)
			}
		})
	}
}

func TestEvaluateMatchedRuleOutsideRolloutReturnsDefault(t *testing.T) {
	env, err := NewCELEnv()
	if err != nil {
		t.Fatalf("new cel env: %v", err)
	}
	e := NewEngine(env)

	compiled, err := e.CompileFlag("feature-a", core.FlagConfig{
		Key:          "feature-a",
		Type:         core.ValueTypeString,
		Enabled:      true,
		DefaultValue: "control",
		Rules: []core.Rule{
			{
				Expression: `true`,
				Rollout:    core.Rollout{Percentage: 0},
				Value:      "variant-a",
			},
		},
	})
	if err != nil {
		t.Fatalf("compile flag: %v", err)
	}

	got := e.Evaluate(compiled, DataContext{})
	if got.Type != core.ValueTypeString || got.Value != "control" {
		t.Fatalf("Evaluate() = %#v, want string control", got)
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
	if got.Type != core.ValueTypeBoolean || got.Value != true {
		t.Fatalf("Evaluate() = %#v, want boolean true", got)
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

func valuesEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
