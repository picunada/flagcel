package service

import (
	"encoding/json"
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

func normalizeFlag(flag core.FlagConfig) core.FlagConfig {
	if flag.Type == "" {
		flag.Type = core.ValueTypeBoolean
	}
	if flag.Type == core.ValueTypeBoolean && flag.DefaultValue == nil {
		flag.DefaultValue = false
	}
	for i, rule := range flag.Rules {
		flag.Rules[i] = normalizeRule(rule)
	}
	return flag
}

func validateFlag(flag core.FlagConfig, schema *core.ContextSchema) error {
	flag = normalizeFlag(flag)
	issues := make([]core.ValidationIssue, 0)

	if !isSupportedValueType(flag.Type) {
		issues = append(issues, core.ValidationIssue{
			Code:    core.ValidationIssueInvalidValueType,
			Field:   "type",
			Message: "Flag type must be one of boolean, string, number, or json.",
		})
	}
	if issue := validateValueIssue(flag.Type, flag.DefaultValue, "default_value"); issue != nil {
		issues = append(issues, *issue)
	}

	for i, rule := range flag.Rules {
		if issue := validateValueIssue(flag.Type, ruleValueForValidation(rule, flag.Type), fmt.Sprintf("rules[%d].value", i)); issue != nil {
			issues = append(issues, *issue)
		}
		if err := validateRule(rule, schema); err != nil {
			if validationErr, ok := err.(*core.ValidationError); ok {
				for _, issue := range validationErr.Issues {
					issue.Field = fmt.Sprintf("rules[%d].%s", i, issue.Field)
					issues = append(issues, issue)
				}
			} else {
				return err
			}
		}
	}

	if len(issues) > 0 {
		return &core.ValidationError{
			Message: "Flag validation failed",
			Issues:  dedupeIssues(issues),
		}
	}
	return nil
}

func validateRuleValue(rule core.Rule, valueType core.ValueType) error {
	if valueType == "" {
		valueType = core.ValueTypeBoolean
	}
	if issue := validateValueIssue(valueType, ruleValueForValidation(rule, valueType), "value"); issue != nil {
		return &core.ValidationError{
			Message: "Rule validation failed",
			Issues:  []core.ValidationIssue{*issue},
		}
	}
	return nil
}

func isSupportedValueType(valueType core.ValueType) bool {
	switch valueType {
	case core.ValueTypeBoolean, core.ValueTypeString, core.ValueTypeNumber, core.ValueTypeJSON:
		return true
	default:
		return false
	}
}

func ruleValueForValidation(rule core.Rule, valueType core.ValueType) any {
	if rule.Value == nil && valueType == core.ValueTypeBoolean {
		return true
	}
	return rule.Value
}

func validateValueIssue(valueType core.ValueType, value any, field string) *core.ValidationIssue {
	if valueType == "" {
		valueType = core.ValueTypeBoolean
	}
	if !isSupportedValueType(valueType) {
		return nil
	}

	ok := false
	switch valueType {
	case core.ValueTypeBoolean:
		_, ok = value.(bool)
	case core.ValueTypeString:
		_, ok = value.(string)
	case core.ValueTypeNumber:
		ok = isJSONNumber(value)
	case core.ValueTypeJSON:
		ok = true
	}
	if ok {
		return nil
	}
	return &core.ValidationIssue{
		Code:    core.ValidationIssueInvalidValue,
		Field:   field,
		Message: fmt.Sprintf("%s must match flag type %q.", field, valueType),
	}
}

func isJSONNumber(value any) bool {
	switch v := value.(type) {
	case json.Number:
		_, err := v.Float64()
		return err == nil
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
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
