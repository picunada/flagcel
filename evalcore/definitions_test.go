package evalcore

import "testing"

func TestLoadEvaluatorEvaluatesCompiledDefinitions(t *testing.T) {
	evaluator, err := Load(Definitions{
		Flags: []FlagDefinition{
			{
				FlagConfig: FlagConfig{
					Key:          "checkout-copy",
					Type:         ValueTypeString,
					Enabled:      true,
					DefaultValue: "control",
					Rules: []Rule{
						{
							ID:         "variant-a",
							Expression: `user.tier == "pro"`,
							Rollout:    Rollout{Percentage: 100, BucketBy: "user.id"},
							Value:      "pro-copy",
						},
					},
				},
				ContextSchema: &ContextSchema{
					ID: "user-context",
					Fields: []ContextField{
						{Path: "user.id", Type: ContextTypeString},
						{Path: "user.tier", Type: ContextTypeString},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	result := evaluator.Evaluate("checkout-copy", DataContext{
		"user": map[string]any{"id": "user-123", "tier": "pro"},
	})

	if result.ValueType != ValueTypeString || result.Value != "pro-copy" {
		t.Fatalf("Evaluate() = %#v, want string pro-copy", result)
	}
	if result.Reason != "matched_rule" || result.Variant != "variant-a" || result.Error != "" {
		t.Fatalf("Evaluate() = %#v, want matched variant-a without error", result)
	}
}

func TestLoadEvaluatorReturnsResultWithoutTrace(t *testing.T) {
	evaluator, err := Load(Definitions{
		Flags: []FlagDefinition{
			{
				FlagConfig: FlagConfig{
					Key:          "checkout-copy",
					Type:         ValueTypeString,
					Enabled:      true,
					DefaultValue: "control",
					Rules: []Rule{
						{
							ID:         "variant-a",
							Expression: `user.tier == "pro"`,
							Rollout:    Rollout{Percentage: 0, BucketBy: "user.id"},
							Value:      "pro-copy",
						},
					},
				},
				ContextSchema: &ContextSchema{
					Fields: []ContextField{
						{Path: "user.id", Type: ContextTypeString},
						{Path: "user.tier", Type: ContextTypeString},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	result := evaluator.Evaluate("checkout-copy", DataContext{
		"user": map[string]any{"id": "user-123", "tier": "pro"},
	})

	if result.ValueType != ValueTypeString || result.Value != "control" {
		t.Fatalf("Evaluate() = %#v, want default string control", result)
	}
	if result.Reason != "matched_rule" || result.Variant != "variant-a" || result.Error != "" {
		t.Fatalf("Evaluate() = %#v, want matched variant-a without error", result)
	}
}

func TestLoadEvaluatorReturnsCompileError(t *testing.T) {
	_, err := Load(Definitions{
		Flags: []FlagDefinition{
			{
				FlagConfig: FlagConfig{
					Key:          "bad-flag",
					Enabled:      true,
					DefaultValue: false,
					Rules: []Rule{
						{ID: "bad-rule", Expression: `user.tier ==`, Rollout: Rollout{Percentage: 100}},
					},
				},
				ContextSchema: &ContextSchema{
					Fields: []ContextField{{Path: "user.tier", Type: ContextTypeString}},
				},
			},
		},
	})
	if err == nil {
		t.Fatal("Load() error = nil, want compile error")
	}
}

func TestEvaluationResultIncludesRuleCELError(t *testing.T) {
	evaluator, err := Load(Definitions{
		Flags: []FlagDefinition{
			{
				FlagConfig: FlagConfig{
					Key:          "age-gate",
					Enabled:      true,
					DefaultValue: false,
					Rules: []Rule{
						{
							ID:         "adult",
							Expression: `user.age > 18`,
							Rollout:    Rollout{Percentage: 100},
							Value:      true,
						},
					},
				},
				ContextSchema: &ContextSchema{
					Fields: []ContextField{{Path: "user.age", Type: ContextTypeString}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	result := evaluator.Evaluate("age-gate", DataContext{
		"user": map[string]any{"age": "not-a-number"},
	})

	if result.Error == "" {
		t.Fatalf("Evaluate() = %#v, want CEL rule error", result)
	}
	if result.ValueType != ValueTypeBoolean || result.Value != false {
		t.Fatalf("Evaluate() = %#v, want default false", result)
	}
}

func TestLoadEvaluatorSupportsTopLevelScalarContextFields(t *testing.T) {
	evaluator, err := Load(Definitions{
		Flags: []FlagDefinition{
			{
				FlagConfig: FlagConfig{
					Key:          "openfeature-targeting",
					Type:         ValueTypeBoolean,
					Enabled:      true,
					DefaultValue: false,
					Rules: []Rule{
						{
							ID:         "targeting-key",
							Expression: `targetingKey == "user-123"`,
							Rollout:    Rollout{Percentage: 100},
							Value:      true,
						},
					},
				},
				ContextSchema: &ContextSchema{
					Fields: []ContextField{{Path: "targetingKey", Type: ContextTypeString}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	result := evaluator.Evaluate("openfeature-targeting", DataContext{
		"targetingKey": "user-123",
	})

	if result.ValueType != ValueTypeBoolean || result.Value != true {
		t.Fatalf("Evaluate() = %#v, want boolean true", result)
	}
	if result.Reason != "matched_rule" || result.Variant != "targeting-key" || result.Error != "" {
		t.Fatalf("Evaluate() = %#v, want matched targeting-key without error", result)
	}
}
