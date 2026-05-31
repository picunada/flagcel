package engine

import (
	"github.com/google/cel-go/cel"
	"github.com/picunada/flagcel/evalcore"
)

type DataContext = evalcore.DataContext
type Flag = evalcore.Flag
type CompiledRule = evalcore.CompiledRule
type Rollout = evalcore.Rollout
type EvaluationTrace = evalcore.EvaluationTrace
type MatchedRuleTrace = evalcore.MatchedRuleTrace
type RuleEvaluationTrace = evalcore.RuleEvaluationTrace
type BucketTrace = evalcore.BucketTrace
type EvaluationResult = evalcore.EvaluationResult

type Engine struct {
	inner *evalcore.Engine
}

func NewCELEnv() (*cel.Env, error) {
	return evalcore.NewCELEnv()
}

func NewCELEnvForContext(schema *evalcore.ContextSchema) (*cel.Env, error) {
	return evalcore.NewCELEnvForContext(schema)
}

func NewEngine(celEnv *cel.Env) *Engine {
	return &Engine{inner: evalcore.NewEngine(celEnv)}
}

func (e *Engine) core() *evalcore.Engine {
	if e == nil || e.inner == nil {
		return evalcore.NewEngine(nil)
	}
	return e.inner
}

func (e *Engine) Evaluate(flag *Flag, data DataContext) evalcore.FlagValue {
	return e.core().Evaluate(flag, data)
}

func (e *Engine) EvaluateTrace(flag *Flag, data DataContext) EvaluationTrace {
	return e.core().EvaluateTrace(flag, data)
}

func (e *Engine) CompileFlag(key string, config evalcore.FlagConfig) (*Flag, error) {
	return e.core().CompileFlag(key, config)
}

func (e *Engine) CompileFlagForContext(key string, config evalcore.FlagConfig, schema *evalcore.ContextSchema) (*Flag, error) {
	return e.core().CompileFlagForContext(key, config, schema)
}

func (e *Engine) EvaluateConfig(config evalcore.FlagConfig, data DataContext) EvaluationTrace {
	return e.core().EvaluateConfig(config, data)
}

func (e *Engine) EvaluateConfigForContext(config evalcore.FlagConfig, schema *evalcore.ContextSchema, data DataContext) EvaluationTrace {
	return e.core().EvaluateConfigForContext(config, schema, data)
}

func (e *Engine) bucket(flagKey string, user DataContext, rollout Rollout) bool {
	return evalcore.Bucket(flagKey, user, rollout)
}
