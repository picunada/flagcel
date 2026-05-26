package engine

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

// Evaluate each rule in flag and return bucket for first matched or return flag default value.
func (e *Engine) Evaluate(flag *Flag, data DataContext) bool {
	if flag == nil {
		return false
	}

	if !flag.Enabled {
		return flag.DefaultValue
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

		return e.bucket(flag.Key, data, rule.Rollout)
	}

	return flag.DefaultValue
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
