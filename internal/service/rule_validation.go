package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/cel-go/cel"
	celast "github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	celtypes "github.com/google/cel-go/common/types"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/engine"
)

func normalizeRule(rule core.Rule) core.Rule {
	rule.Expression = strings.TrimSpace(rule.Expression)
	rule.Rollout.BucketBy = strings.TrimSpace(rule.Rollout.BucketBy)
	return rule
}

func validateRule(rule core.Rule, schema *core.ContextSchema) error {
	rule = normalizeRule(rule)
	issues := make([]core.ValidationIssue, 0)

	if rule.Expression == "" {
		issues = append(issues, core.ValidationIssue{
			Code:    core.ValidationIssueParseError,
			Field:   "expression",
			Message: "Expression is required.",
		})
	}

	if rule.Rollout.Percentage < 0 || rule.Rollout.Percentage > 100 {
		issues = append(issues, core.ValidationIssue{
			Code:    core.ValidationIssueInvalidRollout,
			Field:   "rollout.percentage",
			Message: "Rollout percentage must be between 0 and 100.",
		})
	}

	if rule.Rollout.Percentage > 0 && rule.Rollout.Percentage < 100 && rule.Rollout.BucketBy == "" {
		issues = append(issues, core.ValidationIssue{
			Code:    core.ValidationIssueMissingBucket,
			Field:   "rollout.bucket_by",
			Message: "Bucket field is required for partial rollouts.",
		})
	}
	if rule.Rollout.BucketBy != "" && schema != nil && !schemaAllowsBucketPath(schema, rule.Rollout.BucketBy) {
		issues = append(issues, core.ValidationIssue{
			Code:    core.ValidationIssueMissingBucket,
			Field:   "rollout.bucket_by",
			Path:    rule.Rollout.BucketBy,
			Message: fmt.Sprintf("Bucket field %q is not defined in context %q.", rule.Rollout.BucketBy, schema.Name),
		})
	}

	if rule.Expression != "" {
		env, err := engine.NewCELEnvForContext(schema)
		if err != nil {
			return fmt.Errorf("rule validation: build CEL env: %w", err)
		}

		parsed, parseIssues := env.Parse(rule.Expression)
		if parseIssues != nil && parseIssues.Err() != nil {
			issues = append(issues, core.ValidationIssue{
				Code:    core.ValidationIssueParseError,
				Field:   "expression",
				Message: fmt.Sprintf("Expression parse error: %s", parseIssues.Err()),
			})
		} else {
			for _, path := range unknownExpressionPaths(parsed, schema) {
				issues = append(issues, core.ValidationIssue{
					Code:    core.ValidationIssueUnknownField,
					Field:   "expression",
					Path:    path,
					Message: fmt.Sprintf("Unknown field %q.", path),
				})
			}

			checked, checkIssues := env.Check(parsed)
			if checkIssues != nil && checkIssues.Err() != nil {
				issueCode := core.ValidationIssueParseError
				if strings.Contains(checkIssues.Err().Error(), "undeclared reference to") {
					issueCode = core.ValidationIssueUnknownField
				}
				issues = append(issues, core.ValidationIssue{
					Code:    issueCode,
					Field:   "expression",
					Message: fmt.Sprintf("Expression compile error: %s", checkIssues.Err()),
				})
			} else if checked.OutputType() != cel.BoolType {
				issues = append(issues, core.ValidationIssue{
					Code:    core.ValidationIssueNonBoolExpression,
					Field:   "expression",
					Message: fmt.Sprintf("Expression must return bool, got %s.", checked.OutputType().String()),
				})
			}
		}
	}

	if len(issues) > 0 {
		return &core.ValidationError{
			Message: "Rule validation failed",
			Issues:  dedupeIssues(issues),
		}
	}
	return nil
}

func unknownExpressionPaths(parsed *cel.Ast, schema *core.ContextSchema) []string {
	if parsed == nil || schema == nil || len(schema.Fields) == 0 {
		return nil
	}

	roots := schemaRoots(schema)
	unknown := map[string]struct{}{}
	celast.PostOrderVisit(parsed.NativeRep().Expr(), celast.NewExprVisitor(func(expr celast.Expr) {
		parts, ok := expressionPathParts(expr)
		if !ok || len(parts) < 2 {
			return
		}
		if _, ok := roots[parts[0]]; !ok {
			return
		}
		path := strings.Join(parts, ".")
		if !schemaAllowsPath(schema, path) {
			unknown[path] = struct{}{}
		}
	}))

	out := make([]string, 0, len(unknown))
	for path := range unknown {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func expressionPathParts(expr celast.Expr) ([]string, bool) {
	switch expr.Kind() {
	case celast.IdentKind:
		name := expr.AsIdent()
		if name == "" {
			return nil, false
		}
		return []string{name}, true
	case celast.SelectKind:
		selectExpr := expr.AsSelect()
		parts, ok := expressionPathParts(selectExpr.Operand())
		if !ok {
			return nil, false
		}
		return append(parts, selectExpr.FieldName()), true
	case celast.CallKind:
		call := expr.AsCall()
		if call.FunctionName() != operators.Index && call.FunctionName() != operators.OptIndex {
			return nil, false
		}
		args := call.Args()
		if len(args) != 2 {
			return nil, false
		}
		parts, ok := expressionPathParts(args[0])
		if !ok {
			return nil, false
		}
		key, ok := stringLiteral(args[1])
		if !ok || key == "" {
			return nil, false
		}
		return append(parts, key), true
	default:
		return nil, false
	}
}

func stringLiteral(expr celast.Expr) (string, bool) {
	if expr.Kind() != celast.LiteralKind {
		return "", false
	}
	val, ok := expr.AsLiteral().(celtypes.String)
	if !ok {
		return "", false
	}
	return string(val), true
}

func schemaAllowsPath(schema *core.ContextSchema, path string) bool {
	for _, field := range schema.Fields {
		if field.Path == path || strings.HasPrefix(field.Path, path+".") {
			return true
		}
		if field.Type == core.ContextTypeMap && strings.HasPrefix(path, field.Path+".") {
			return true
		}
	}
	return false
}

func schemaAllowsBucketPath(schema *core.ContextSchema, path string) bool {
	for _, field := range schema.Fields {
		if field.Path == path {
			return true
		}
		if field.Type == core.ContextTypeMap && strings.HasPrefix(path, field.Path+".") {
			return true
		}
	}
	return false
}

func schemaRoots(schema *core.ContextSchema) map[string]struct{} {
	roots := make(map[string]struct{}, len(schema.Fields))
	for _, field := range schema.Fields {
		root, _, _ := strings.Cut(field.Path, ".")
		if root != "" {
			roots[root] = struct{}{}
		}
	}
	return roots
}

func dedupeIssues(issues []core.ValidationIssue) []core.ValidationIssue {
	seen := make(map[string]struct{}, len(issues))
	out := make([]core.ValidationIssue, 0, len(issues))
	for _, issue := range issues {
		key := issue.Code + "\x00" + issue.Field + "\x00" + issue.Path + "\x00" + issue.Message
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, issue)
	}
	return out
}
