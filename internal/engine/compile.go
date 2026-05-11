package engine

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/picunada/flagcel/internal/core"
)

func (e *Engine) CompileFlag(key string, config core.FlagConfig) (*Flag, error) {
	flag := &Flag{
		Key:          key,
		Enabled:      config.Enabled,
		DefaultValue: config.DefaultValue,
		Rules:        make([]CompiledRule, 0, len(config.Rules)),
	}

	for i, r := range config.Rules {
		program, err := e.compileExpression(r.Expression)
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
		flag.Rules = append(flag.Rules, CompiledRule{
			Source:  r.Expression,
			Program: program,
			Rollout: Rollout(r.Rollout),
		})
	}

	return flag, nil
}

func (e *Engine) compileExpression(source string) (cel.Program, error) {
	if source == "" {
		return nil, nil
	}

	ast, issues := e.celEnv.Compile(source)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("compile: %w", issues.Err())
	}

	if ast.OutputType() != cel.BoolType {
		return nil, fmt.Errorf(
			"expression must return bool type, got %s",
			ast.OutputType().String(),
		)
	}

	return e.celEnv.Program(
		ast,
		cel.EvalOptions(cel.OptOptimize),
		cel.InterruptCheckFrequency(100),
	)
}
