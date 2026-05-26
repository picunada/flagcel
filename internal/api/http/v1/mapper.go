package v1

import (
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/engine"
)

func toCoreFlag(r CreateFlagRequest) core.FlagConfig {
	rules := make([]core.Rule, len(r.Rules))
	for i, rule := range r.Rules {
		rules[i] = toCoreRule(rule)
	}
	return core.FlagConfig{
		Key:          r.Key,
		Enabled:      r.Enabled,
		Rules:        rules,
		DefaultValue: r.DefaultValue,
		ContextID:    r.ContextID,
	}
}

func toCoreRule(r CreateRuleRequest) core.Rule {
	return core.Rule{
		Expression: r.Expression,
		Rollout:    toCoreRollout(r.Rollout),
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
			Matched:    result.Matched,
			Error:      result.Error,
		}
	}
	return out
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
