package evalcore

import "fmt"

type Definitions struct {
	Flags    []FlagDefinition `json:"flags"`
	Contexts []ContextSchema  `json:"contexts,omitempty"`
}

type FlagDefinition struct {
	FlagConfig
	ContextSchema *ContextSchema `json:"context_schema,omitempty"`
}

type Evaluator struct {
	engine *Engine
	flags  map[string]*Flag
	keys   []string
}

func Load(defs Definitions) (*Evaluator, error) {
	env, err := NewCELEnv()
	if err != nil {
		return nil, err
	}
	engine := NewEngine(env)
	contexts := make(map[string]*ContextSchema, len(defs.Contexts))
	for i := range defs.Contexts {
		context := defs.Contexts[i]
		if context.ID != "" {
			contexts[context.ID] = &context
		}
	}

	evaluator := &Evaluator{
		engine: engine,
		flags:  make(map[string]*Flag, len(defs.Flags)),
		keys:   make([]string, 0, len(defs.Flags)),
	}
	for i, def := range defs.Flags {
		schema := def.ContextSchema
		if schema == nil && def.ContextID != nil {
			schema = contexts[*def.ContextID]
		}
		flag, err := engine.CompileFlagForContext(def.Key, def.FlagConfig, schema)
		if err != nil {
			return nil, fmt.Errorf("flags[%d] %q: %w", i, def.Key, err)
		}
		evaluator.flags[flag.Key] = flag
		evaluator.keys = append(evaluator.keys, flag.Key)
	}
	return evaluator, nil
}

func (e *Evaluator) Evaluate(key string, ctx DataContext) EvaluationResult {
	if e == nil {
		return missingFlagResult(key)
	}
	flag, ok := e.flags[key]
	if !ok {
		return missingFlagResult(key)
	}
	return e.engine.EvaluateTrace(flag, ctx).Result()
}

func (e *Evaluator) EvaluateAll(ctx DataContext) map[string]EvaluationResult {
	if e == nil {
		return map[string]EvaluationResult{}
	}
	out := make(map[string]EvaluationResult, len(e.flags))
	for _, key := range e.keys {
		out[key] = e.Evaluate(key, ctx)
	}
	return out
}

func missingFlagResult(key string) EvaluationResult {
	return EvaluationResult{
		Key:       key,
		Value:     false,
		ValueType: ValueTypeBoolean,
		Reason:    "not_found",
		Error:     "flag not found",
	}
}
