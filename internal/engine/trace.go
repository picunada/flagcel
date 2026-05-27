package engine

import "github.com/picunada/flagcel/internal/core"

type EvaluationTrace struct {
	Key          string
	Type         core.ValueType
	Enabled      bool
	DefaultValue any
	Value        any
	Reason       string
	Error        string
	MatchedRule  *MatchedRuleTrace
	Bucket       *BucketTrace
	RuleResults  []RuleEvaluationTrace
}

type MatchedRuleTrace struct {
	ID         string
	Index      int
	Expression string
	Value      any
}

type RuleEvaluationTrace struct {
	ID         string
	Index      int
	Expression string
	Value      any
	Matched    bool
	Error      string
}

type BucketTrace struct {
	BucketBy     string
	BucketValue  string
	BucketNumber *int
	Percentage   int
	InRollout    bool
	Missing      bool
}

func (e *Engine) EvaluateConfig(config core.FlagConfig, data DataContext) EvaluationTrace {
	return e.EvaluateConfigForContext(config, nil, data)
}

func (e *Engine) EvaluateConfigForContext(config core.FlagConfig, schema *core.ContextSchema, data DataContext) EvaluationTrace {
	trace := EvaluationTrace{
		Key:          config.Key,
		Type:         config.Type,
		Enabled:      config.Enabled,
		DefaultValue: config.DefaultValue,
		Value:        config.DefaultValue,
		RuleResults:  make([]RuleEvaluationTrace, 0, len(config.Rules)),
	}
	if trace.Type == "" {
		trace.Type = core.ValueTypeBoolean
	}
	if trace.DefaultValue == nil && trace.Type == core.ValueTypeBoolean {
		trace.DefaultValue = false
		trace.Value = false
	}

	if !config.Enabled {
		trace.Reason = "disabled"
		return trace
	}

	env, err := NewCELEnvForContext(schema)
	if err != nil {
		trace.Error = err.Error()
		trace.Reason = "cel_error"
		return trace
	}

	for i, rule := range config.Rules {
		result := RuleEvaluationTrace{
			ID:         rule.ID,
			Index:      i,
			Expression: rule.Expression,
			Value:      ruleValue(rule, trace.Type),
		}

		program, err := e.compileExpressionWithEnv(env, rule.Expression)
		if err != nil {
			result.Error = err.Error()
			trace.Error = err.Error()
			trace.Reason = "cel_error"
			trace.RuleResults = append(trace.RuleResults, result)
			return trace
		}

		matches, err := e.evaluateExpression(program, data)
		if err != nil {
			result.Error = err.Error()
			trace.RuleResults = append(trace.RuleResults, result)
			continue
		}

		result.Matched = matches
		trace.RuleResults = append(trace.RuleResults, result)
		if !matches {
			continue
		}

		rollout := Rollout(rule.Rollout)
		bucket := e.bucketDetails(config.Key, data, rollout)
		trace.Bucket = &bucket
		trace.MatchedRule = &MatchedRuleTrace{
			ID:         rule.ID,
			Index:      i,
			Expression: rule.Expression,
			Value:      ruleValue(rule, trace.Type),
		}
		if bucket.InRollout {
			trace.Value = ruleValue(rule, trace.Type)
		}
		trace.Reason = "matched_rule"
		return trace
	}

	trace.Reason = "default_no_match"
	return trace
}
