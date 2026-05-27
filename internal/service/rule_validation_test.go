package service

import (
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
