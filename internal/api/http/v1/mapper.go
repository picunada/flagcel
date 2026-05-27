package v1

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/engine"
)

func toCoreFlag(r CreateFlagRequest) (core.FlagConfig, error) {
	rules := make([]core.Rule, len(r.Rules))
	for i, rule := range r.Rules {
		coreRule, err := toCoreRule(rule)
		if err != nil {
			return core.FlagConfig{}, fmt.Errorf("rules[%d].%w", i, err)
		}
		rules[i] = coreRule
	}
	valueType := core.ValueType(r.Type)
	if valueType == "" {
		valueType = core.ValueTypeBoolean
	}
	if valueType == core.ValueTypeJSON && r.DefaultValue == nil {
		return core.FlagConfig{}, fmt.Errorf("default_value: required for json flags")
	}
	defaultValue, err := decodeValue(r.DefaultValue, defaultValueFallback(valueType))
	if err != nil {
		return core.FlagConfig{}, fmt.Errorf("default_value: %w", err)
	}
	return core.FlagConfig{
		Key:          r.Key,
		Type:         valueType,
		Enabled:      r.Enabled,
		Rules:        rules,
		DefaultValue: defaultValue,
		ContextID:    r.ContextID,
	}, nil
}

func toCoreRule(r CreateRuleRequest) (core.Rule, error) {
	value, err := decodeValue(r.Value, true)
	if err != nil {
		return core.Rule{}, fmt.Errorf("value: %w", err)
	}
	return core.Rule{
		Expression: r.Expression,
		Rollout:    toCoreRollout(r.Rollout),
		Value:      value,
	}, nil
}

func decodeValue(raw json.RawMessage, fallback any) (any, error) {
	if raw == nil {
		return fallback, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func defaultValueFallback(valueType core.ValueType) any {
	switch valueType {
	case core.ValueTypeString:
		return ""
	case core.ValueTypeNumber:
		return json.Number("0")
	case core.ValueTypeJSON:
		return nil
	case core.ValueTypeBoolean:
		fallthrough
	default:
		return false
	}
}

func toCoreRollout(r RolloutResponse) core.Rollout {
	return core.Rollout{
		Percentage: r.Percentage,
		BucketBy:   r.BucketBy,
	}
}

func toFlagResponse(f core.FlagConfig) FlagResponse {
	rules := make([]RuleResponse, len(f.Rules))
	for i, rule := range f.Rules {
		rules[i] = toRuleResponse(rule)
	}
	return FlagResponse{
		Key:          f.Key,
		Type:         string(f.Type),
		Enabled:      f.Enabled,
		Rules:        rules,
		DefaultValue: f.DefaultValue,
		ContextID:    f.ContextID,
		UpdatedAt:    formatTime(f.UpdatedAt),
	}
}

func toRuleResponse(r core.Rule) RuleResponse {
	return RuleResponse{
		ID:         r.ID,
		Expression: r.Expression,
		Rollout:    toRolloutResponse(r.Rollout),
		Value:      r.Value,
	}
}

func toRolloutResponse(r core.Rollout) RolloutResponse {
	return RolloutResponse{
		Percentage: r.Percentage,
		BucketBy:   r.BucketBy,
	}
}

func toFlagResponses(flags []*core.FlagConfig) []FlagResponse {
	out := make([]FlagResponse, len(flags))
	for i, f := range flags {
		out[i] = toFlagResponse(*f)
	}
	return out
}

func toEvalTraceResponse(t engine.EvaluationTrace) EvalTraceResponse {
	out := EvalTraceResponse{
		Key:          t.Key,
		ValueType:    string(t.Type),
		Enabled:      t.Enabled,
		DefaultValue: t.DefaultValue,
		Value:        t.Value,
		Reason:       t.Reason,
		Error:        t.Error,
		RuleResults:  make([]EvalRuleResultResponse, len(t.RuleResults)),
	}

	if t.MatchedRule != nil {
		out.MatchedRule = &EvalMatchedRuleResponse{
			ID:         t.MatchedRule.ID,
			Index:      t.MatchedRule.Index,
			Expression: t.MatchedRule.Expression,
			Value:      t.MatchedRule.Value,
		}
	}
	if t.Bucket != nil {
		out.Bucket = &EvalBucketResponse{
			BucketBy:     t.Bucket.BucketBy,
			BucketValue:  t.Bucket.BucketValue,
			BucketNumber: t.Bucket.BucketNumber,
			Percentage:   t.Bucket.Percentage,
			InRollout:    t.Bucket.InRollout,
			Missing:      t.Bucket.Missing,
		}
	}
	for i, result := range t.RuleResults {
		out.RuleResults[i] = EvalRuleResultResponse{
			ID:         result.ID,
			Index:      result.Index,
			Expression: result.Expression,
			Value:      result.Value,
			Matched:    result.Matched,
			Error:      result.Error,
		}
	}
	return out
}

func toEvalFlagValueResponse(v core.FlagValue) EvalFlagValueResponse {
	return EvalFlagValueResponse{
		ValueType: string(v.Type),
		Value:     v.Value,
	}
}

func toCoreContext(id string, name, description string, fields []ContextFieldDTO) core.ContextSchema {
	coreFields := make([]core.ContextField, len(fields))
	for i, f := range fields {
		coreFields[i] = core.ContextField{
			Path: f.Path,
			Type: core.ContextType(f.Type),
		}
	}
	return core.ContextSchema{
		ID:          id,
		Name:        name,
		Description: description,
		Fields:      coreFields,
	}
}

func toContextResponse(c core.ContextSchema) ContextResponse {
	fields := make([]ContextFieldDTO, len(c.Fields))
	for i, f := range c.Fields {
		fields[i] = ContextFieldDTO{
			Path: f.Path,
			Type: string(f.Type),
		}
	}
	return ContextResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Fields:      fields,
	}
}

func toContextResponses(cs []*core.ContextSchema) []ContextResponse {
	out := make([]ContextResponse, len(cs))
	for i, c := range cs {
		out[i] = toContextResponse(*c)
	}
	return out
}
