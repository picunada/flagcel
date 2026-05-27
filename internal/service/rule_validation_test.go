package service

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/picunada/flagcel/internal/core"
)

func TestValidateRuleAcceptsValidRule(t *testing.T) {
	err := validateRule(core.Rule{
		Expression: `user.country == "US"`,
		Rollout: core.Rollout{
			Percentage: 50,
			BucketBy:   "user.id",
		},
	}, testRuleContextSchema())
	if err != nil {
		t.Fatalf("validateRule() error = %v", err)
	}
}

func TestValidateFlagValueTypes(t *testing.T) {
	tests := []struct {
		name string
		flag core.FlagConfig
	}{
		{
			name: "boolean",
			flag: core.FlagConfig{
				Type:         core.ValueTypeBoolean,
				DefaultValue: false,
				Rules: []core.Rule{{
					Expression: `user.country == "US"`,
					Rollout:    core.Rollout{Percentage: 100},
					Value:      true,
				}},
			},
		},
		{
			name: "string",
			flag: core.FlagConfig{
				Type:         core.ValueTypeString,
				DefaultValue: "control",
				Rules: []core.Rule{{
					Expression: `user.country == "US"`,
					Rollout:    core.Rollout{Percentage: 100},
					Value:      "variant",
				}},
			},
		},
		{
			name: "number",
			flag: core.FlagConfig{
				Type:         core.ValueTypeNumber,
				DefaultValue: json.Number("1.5"),
				Rules: []core.Rule{{
					Expression: `user.country == "US"`,
					Rollout:    core.Rollout{Percentage: 100},
					Value:      json.Number("2"),
				}},
			},
		},
		{
			name: "json",
			flag: core.FlagConfig{
				Type:         core.ValueTypeJSON,
				DefaultValue: map[string]any{"name": "control"},
				Rules: []core.Rule{{
					Expression: `user.country == "US"`,
					Rollout:    core.Rollout{Percentage: 100},
					Value:      map[string]any{"name": "variant"},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateFlag(tt.flag, testRuleContextSchema()); err != nil {
				t.Fatalf("validateFlag() error = %v", err)
			}
		})
	}
}

func TestValidateFlagRejectsMismatchedValues(t *testing.T) {
	flag := core.FlagConfig{
		Type:         core.ValueTypeNumber,
		DefaultValue: "not-a-number",
		Rules: []core.Rule{{
			Expression: `user.country == "US"`,
			Rollout:    core.Rollout{Percentage: 100},
			Value:      true,
		}},
	}

	err := validateFlag(flag, testRuleContextSchema())
	if err == nil {
		t.Fatal("validateFlag() error = nil, want validation error")
	}
	var validationErr *core.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("validateFlag() error = %T, want *core.ValidationError", err)
	}
	if !hasIssue(validationErr.Issues, core.ValidationIssueInvalidValue) {
		t.Fatalf("issues = %#v, want invalid value", validationErr.Issues)
	}
}

func TestValidateRuleReturnsStructuredIssues(t *testing.T) {
	tests := []struct {
		name string
		rule core.Rule
		want string
	}{
		{
			name: "parse error",
			rule: core.Rule{
				Expression: `user.country ==`,
				Rollout:    core.Rollout{Percentage: 100},
			},
			want: core.ValidationIssueParseError,
		},
		{
			name: "unknown expression field",
			rule: core.Rule{
				Expression: `user.plan == "pro"`,
				Rollout:    core.Rollout{Percentage: 100},
			},
			want: core.ValidationIssueUnknownField,
		},
		{
			name: "non bool expression",
			rule: core.Rule{
				Expression: `user.country`,
				Rollout:    core.Rollout{Percentage: 100},
			},
			want: core.ValidationIssueNonBoolExpression,
		},
		{
			name: "invalid rollout",
			rule: core.Rule{
				Expression: `user.country == "US"`,
				Rollout:    core.Rollout{Percentage: 101},
			},
			want: core.ValidationIssueInvalidRollout,
		},
		{
			name: "missing bucket for partial rollout",
			rule: core.Rule{
				Expression: `user.country == "US"`,
				Rollout:    core.Rollout{Percentage: 10},
			},
			want: core.ValidationIssueMissingBucket,
		},
		{
			name: "bucket field missing from schema",
			rule: core.Rule{
				Expression: `user.country == "US"`,
				Rollout: core.Rollout{
					Percentage: 100,
					BucketBy:   "user.missing",
				},
			},
			want: core.ValidationIssueMissingBucket,
		},
		{
			name: "bucket field must be concrete",
			rule: core.Rule{
				Expression: `user.country == "US"`,
				Rollout: core.Rollout{
					Percentage: 100,
					BucketBy:   "user",
				},
			},
			want: core.ValidationIssueMissingBucket,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRule(tt.rule, testRuleContextSchema())
			if err == nil {
				t.Fatal("validateRule() error = nil, want validation error")
			}
			var validationErr *core.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("validateRule() error = %T, want *core.ValidationError", err)
			}
			if !hasIssue(validationErr.Issues, tt.want) {
				t.Fatalf("issues = %#v, want code %q", validationErr.Issues, tt.want)
			}
		})
	}
}

func hasIssue(issues []core.ValidationIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func testRuleContextSchema() *core.ContextSchema {
	return &core.ContextSchema{
		Name: "user request",
		Fields: []core.ContextField{
			{Path: "user.id", Type: core.ContextTypeString},
			{Path: "user.country", Type: core.ContextTypeString},
			{Path: "request.path", Type: core.ContextTypeString},
		},
	}
}
