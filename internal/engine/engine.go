package engine

import (
	"fmt"

	"github.com/google/cel-go/cel"
)

type Engine struct {
	celEnv *cel.Env
}

type UserContext map[string]any

// Evaluate each rule in flag and return bucket for first matched or return flag default value.
func (e *Engine) Evaluate(flag *Flag, user UserContext) bool {
	if flag == nil {
		return false
	}

	if !flag.Enabled {
		return false
	}

	for _, rule := range flag.Rules {
		matches, err := e.evaluateExpression(rule.Program, user)
		if err != nil {
			continue
		}
		if !matches {
			continue
		}

		return e.bucket(flag.Key, user, rule.Rollout)
	}

	return flag.DefaultValue
}

func (e *Engine) evaluateExpression(program cel.Program, user UserContext) (bool, error) {
	if program == nil {
		return true, nil
	}

	out, _, err := program.Eval(map[string]any{
		"user": user,
	})
	if err != nil {
		return false, fmt.Errorf("eval: %w", err)
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("expression returned non-bool: %T", out.Value())
	}

	return result, nil
}
