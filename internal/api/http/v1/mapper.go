package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/picunada/flagcel/evalcore"
	"github.com/picunada/flagcel/internal/core"
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
		Description:  r.Description,
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
		ID:          r.ID,
		Description: r.Description,
		Expression:  r.Expression,
		Rollout:     toCoreRollout(r.Rollout),
		Value:       value,
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
		Description:  f.Description,
		Type:         string(f.Type),
		Enabled:      f.Enabled,
		Rules:        rules,
		DefaultValue: f.DefaultValue,
		ContextID:    f.ContextID,
		CreatedAt:    formatTime(f.CreatedAt),
		UpdatedAt:    formatTime(f.UpdatedAt),
		CreatedBy:    f.CreatedBy,
		UpdatedBy:    f.UpdatedBy,
		DeletedBy:    f.DeletedBy,
	}
}

func toRuleResponse(r core.Rule) RuleResponse {
	return RuleResponse{
		ID:          r.ID,
		Description: r.Description,
		Expression:  r.Expression,
		Rollout:     toRolloutResponse(r.Rollout),
		Value:       r.Value,
		CreatedAt:   formatTime(r.CreatedAt),
		UpdatedAt:   formatTime(r.UpdatedAt),
		CreatedBy:   r.CreatedBy,
		UpdatedBy:   r.UpdatedBy,
		DeletedBy:   r.DeletedBy,
	}
}

func toAuditEntryResponse(e core.AuditEntry) AuditEntryResponse {
	var snapshot *FlagResponse
	if e.Snapshot != nil {
		resp := toFlagResponse(*e.Snapshot)
		snapshot = &resp
	}
	return AuditEntryResponse{
		Version:    e.Version,
		Action:     e.Action,
		ActorLabel: e.ActorLabel,
		Snapshot:   snapshot,
		CreatedAt:  formatTime(e.CreatedAt),
	}
}

func toAuditEntryResponses(entries []*core.AuditEntry) []AuditEntryResponse {
	out := make([]AuditEntryResponse, len(entries))
	for i, e := range entries {
		out[i] = toAuditEntryResponse(*e)
	}
	return out
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

func toEvalTraceResponse(t evalcore.EvaluationTrace) EvalTraceResponse {
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

func toFlagUsageResponse(buckets []*core.FlagUsageBucket, events []*core.FlagUsageEvent) FlagUsageResponse {
	bucketResponses := make([]FlagUsageBucketResponse, len(buckets))
	for i, bucket := range buckets {
		bucketResponses[i] = FlagUsageBucketResponse{
			BucketStart:   formatTime(bucket.BucketStart),
			ValueType:     string(bucket.ValueType),
			Value:         bucket.Value,
			Reason:        bucket.Reason,
			MatchedRuleID: bucket.MatchedRuleID,
			APIKeyID:      bucket.APIKeyID,
			Source:        bucket.Source,
			Count:         bucket.Count,
		}
	}
	eventResponses := make([]FlagUsageEventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = FlagUsageEventResponse{
			ID:            event.ID,
			ObservedAt:    formatTime(event.ObservedAt),
			ValueType:     string(event.ValueType),
			Value:         event.Value,
			Reason:        event.Reason,
			MatchedRuleID: event.MatchedRuleID,
			APIKeyID:      event.APIKeyID,
			Source:        event.Source,
			LatencyMs:     float64(event.Latency) / float64(time.Millisecond),
			Context:       event.Context,
		}
	}
	return FlagUsageResponse{Buckets: bucketResponses, Events: eventResponses}
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
		CreatedAt:   formatTime(c.CreatedAt),
		UpdatedAt:   formatTime(c.UpdatedAt),
		CreatedBy:   c.CreatedBy,
		DeletedBy:   c.DeletedBy,
	}
}

func toContextResponses(cs []*core.ContextSchema) []ContextResponse {
	out := make([]ContextResponse, len(cs))
	for i, c := range cs {
		out[i] = toContextResponse(*c)
	}
	return out
}

func toCoreEnvironment(id string, req CreateEnvironmentRequest) core.Environment {
	return core.Environment{
		ID:          id,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}
}

func toEnvironmentResponse(env core.Environment) EnvironmentResponse {
	return EnvironmentResponse{
		ID:          env.ID,
		Key:         env.Key,
		Name:        env.Name,
		Description: env.Description,
		CreatedAt:   formatTime(env.CreatedAt),
		UpdatedAt:   formatTime(env.UpdatedAt),
		CreatedBy:   env.CreatedBy,
		DeletedBy:   env.DeletedBy,
	}
}

func toEnvironmentResponses(envs []*core.Environment) []EnvironmentResponse {
	out := make([]EnvironmentResponse, len(envs))
	for i, env := range envs {
		out[i] = toEnvironmentResponse(*env)
	}
	return out
}
