package engine

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/picunada/flagcel/internal/core"
)

func (e *Engine) CompileFlag(key string, config core.FlagConfig) (*Flag, error) {
	return e.compileFlag(key, config, e.celEnv)
}

func (e *Engine) CompileFlagForContext(key string, config core.FlagConfig, schema *core.ContextSchema) (*Flag, error) {
	env, err := NewCELEnvForContext(schema)
	if err != nil {
		return nil, err
	}
	return e.compileFlag(key, config, env)
}

func (e *Engine) compileFlag(key string, config core.FlagConfig, env *cel.Env) (*Flag, error) {
	flag := &Flag{
		Key:          key,
		Type:         config.Type,
		Enabled:      config.Enabled,
		DefaultValue: config.DefaultValue,
		Rules:        make([]CompiledRule, 0, len(config.Rules)),
	}
	if flag.Type == "" {
		flag.Type = core.ValueTypeBoolean
	}
	if flag.DefaultValue == nil && flag.Type == core.ValueTypeBoolean {
		flag.DefaultValue = false
	}

	for i, r := range config.Rules {
		program, err := e.compileExpressionWithEnv(env, r.Expression)
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
		flag.Rules = append(flag.Rules, CompiledRule{
			ID:      r.ID,
			Source:  r.Expression,
			Program: program,
			Rollout: Rollout(r.Rollout),
			Value:   ruleValue(r, flag.Type),
		})
	}

	return flag, nil
}

func ruleValue(rule core.Rule, valueType core.ValueType) any {
	if rule.Value == nil && valueType == core.ValueTypeBoolean {
		return true
	}
	return rule.Value
}

func (e *Engine) compileExpression(source string) (cel.Program, error) {
	return e.compileExpressionWithEnv(e.celEnv, source)
}

func (e *Engine) compileExpressionWithEnv(env *cel.Env, source string) (cel.Program, error) {
	if source == "" {
		return nil, nil
	}

	ast, issues := env.Compile(source)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("compile: %w", issues.Err())
	}

	if ast.OutputType() != cel.BoolType {
		return nil, fmt.Errorf(
			"expression must return bool type, got %s",
			ast.OutputType().String(),
		)
	}

	return env.Program(
		ast,
		cel.EvalOptions(cel.OptOptimize),
		cel.InterruptCheckFrequency(100),
	)
}
