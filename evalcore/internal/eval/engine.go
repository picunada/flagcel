package eval

import (
	"fmt"
	"log/slog"

	"github.com/google/cel-go/cel"
)

type Engine struct {
	celEnv *cel.Env
}

func NewEngine(celEnv *cel.Env) *Engine {
	return &Engine{celEnv: celEnv}
}

type DataContext map[string]any

// Evaluate each rule in flag and return the matched rule value, or the flag default value.
func (e *Engine) Evaluate(flag *Flag, data DataContext) FlagValue {
	if flag == nil {
		return FlagValue{Type: ValueTypeBoolean, Value: false}
	}

	if !flag.Enabled {
		return flagValue(flag.Type, flag.DefaultValue)
	}

	for _, rule := range flag.Rules {
		matches, err := e.evaluateExpression(rule.Program, data)
		if err != nil {
			slog.Debug(fmt.Sprintf("evaluate: rule evaluation err: %s", err.Error()), "rule", rule)
			continue
		}
		if !matches {
			slog.Debug("evaluate: rule did not match", "rule", rule)
			continue
		}

		if e.bucket(flag.Key, data, rule.Rollout) {
			return flagValue(flag.Type, rule.Value)
		}
		return flagValue(flag.Type, flag.DefaultValue)
	}

	return flagValue(flag.Type, flag.DefaultValue)
}

func flagValue(valueType ValueType, value any) FlagValue {
	if valueType == "" {
		valueType = ValueTypeBoolean
	}
	if value == nil && valueType == ValueTypeBoolean {
		value = false
	}
	return FlagValue{Type: valueType, Value: value}
}

func (e *Engine) evaluateExpression(program cel.Program, data DataContext) (bool, error) {
	if program == nil {
		return true, nil
	}

	slog.Debug("evaluate: evaluating expression", "program", program, "user", data)

	out, _, err := program.Eval(map[string]any(data))
	if err != nil {
		return false, fmt.Errorf("eval: %w", err)
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("expression returned non-bool: %T", out.Value())
	}

	return result, nil
}
