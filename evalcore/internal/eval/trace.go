package eval

type EvaluationTrace struct {
	Key          string
	Type         ValueType
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

type EvaluationResult struct {
	Key       string    `json:"key,omitempty"`
	Value     any       `json:"value"`
	ValueType ValueType `json:"value_type"`
	Reason    string    `json:"reason"`
	Variant   string    `json:"variant,omitempty"`
	Error     string    `json:"error,omitempty"`
}

func (t EvaluationTrace) Result() EvaluationResult {
	result := EvaluationResult{
		Key:       t.Key,
		Value:     t.Value,
		ValueType: t.Type,
		Reason:    t.Reason,
		Error:     t.Error,
	}
	if t.MatchedRule != nil {
		result.Variant = t.MatchedRule.ID
	}
	if result.Error == "" {
		for _, rule := range t.RuleResults {
			if rule.Error != "" {
				result.Error = rule.Error
				break
			}
		}
	}
	return result
}

func (e *Engine) EvaluateConfig(config FlagConfig, data DataContext) EvaluationTrace {
	return e.EvaluateConfigForContext(config, nil, data)
}

func (e *Engine) EvaluateConfigForContext(config FlagConfig, schema *ContextSchema, data DataContext) EvaluationTrace {
	trace := EvaluationTrace{
		Key:          config.Key,
		Type:         config.Type,
		Enabled:      config.Enabled,
		DefaultValue: config.DefaultValue,
		Value:        config.DefaultValue,
		RuleResults:  make([]RuleEvaluationTrace, 0, len(config.Rules)),
	}
	if trace.Type == "" {
		trace.Type = ValueTypeBoolean
	}
	if trace.DefaultValue == nil && trace.Type == ValueTypeBoolean {
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

		bucket := e.bucketDetails(config.Key, data, rule.Rollout)
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

func (e *Engine) EvaluateTrace(flag *Flag, data DataContext) EvaluationTrace {
	if flag == nil {
		return EvaluationTrace{
			Type:         ValueTypeBoolean,
			DefaultValue: false,
			Value:        false,
			Reason:       "not_found",
			Error:        "flag not found",
		}
	}

	trace := EvaluationTrace{
		Key:          flag.Key,
		Type:         flag.Type,
		Enabled:      flag.Enabled,
		DefaultValue: flag.DefaultValue,
		Value:        flag.DefaultValue,
		RuleResults:  make([]RuleEvaluationTrace, 0, len(flag.Rules)),
	}
	if trace.Type == "" {
		trace.Type = ValueTypeBoolean
	}
	if trace.DefaultValue == nil && trace.Type == ValueTypeBoolean {
		trace.DefaultValue = false
		trace.Value = false
	}

	if !flag.Enabled {
		trace.Reason = "disabled"
		return trace
	}

	for i, rule := range flag.Rules {
		result := RuleEvaluationTrace{
			ID:         rule.ID,
			Index:      i,
			Expression: rule.Source,
			Value:      rule.Value,
		}

		matches, err := e.evaluateExpression(rule.Program, data)
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

		bucket := e.bucketDetails(flag.Key, data, rule.Rollout)
		trace.Bucket = &bucket
		trace.MatchedRule = &MatchedRuleTrace{
			ID:         rule.ID,
			Index:      i,
			Expression: rule.Source,
			Value:      rule.Value,
		}
		if bucket.InRollout {
			trace.Value = rule.Value
		}
		trace.Reason = "matched_rule"
		return trace
	}

	trace.Reason = "default_no_match"
	return trace
}
