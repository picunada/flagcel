package evalcore

import (
	"github.com/google/cel-go/cel"
	"github.com/picunada/flagcel/evalcore/internal/eval"
)

type Engine struct {
	inner *eval.Engine
}

func NewCELEnv() (*cel.Env, error) {
	return eval.NewCELEnv()
}

func NewCELEnvForContext(schema *ContextSchema) (*cel.Env, error) {
	return eval.NewCELEnvForContext(schema)
}

func NewEngine(celEnv *cel.Env) *Engine {
	return &Engine{inner: eval.NewEngine(celEnv)}
}

func (e *Engine) core() *eval.Engine {
	if e == nil || e.inner == nil {
		return eval.NewEngine(nil)
	}
	return e.inner
}

func (e *Engine) Evaluate(flag *Flag, data DataContext) FlagValue {
	return e.core().Evaluate(flag, data)
}

func (e *Engine) EvaluateTrace(flag *Flag, data DataContext) EvaluationTrace {
	return e.core().EvaluateTrace(flag, data)
}

func (e *Engine) CompileFlag(key string, config FlagConfig) (*Flag, error) {
	return e.core().CompileFlag(key, config)
}

func (e *Engine) CompileFlagForContext(key string, config FlagConfig, schema *ContextSchema) (*Flag, error) {
	return e.core().CompileFlagForContext(key, config, schema)
}

func (e *Engine) EvaluateConfig(config FlagConfig, data DataContext) EvaluationTrace {
	return e.core().EvaluateConfig(config, data)
}

func (e *Engine) EvaluateConfigForContext(config FlagConfig, schema *ContextSchema, data DataContext) EvaluationTrace {
	return e.core().EvaluateConfigForContext(config, schema, data)
}

func Bucket(flagKey string, user DataContext, rollout Rollout) bool {
	return eval.Bucket(flagKey, user, rollout)
}
