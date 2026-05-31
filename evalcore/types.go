package evalcore

import "github.com/picunada/flagcel/evalcore/internal/eval"

type FlagConfig = eval.FlagConfig
type ValueType = eval.ValueType
type FlagValue = eval.FlagValue
type Rule = eval.Rule
type Rollout = eval.Rollout
type EvalContext = eval.EvalContext
type ContextSchema = eval.ContextSchema
type ContextField = eval.ContextField
type ContextType = eval.ContextType
type DataContext = eval.DataContext

type Flag = eval.Flag
type CompiledRule = eval.CompiledRule

type EvaluationTrace = eval.EvaluationTrace
type MatchedRuleTrace = eval.MatchedRuleTrace
type RuleEvaluationTrace = eval.RuleEvaluationTrace
type BucketTrace = eval.BucketTrace
type EvaluationResult = eval.EvaluationResult

const (
	ValueTypeBoolean = eval.ValueTypeBoolean
	ValueTypeString  = eval.ValueTypeString
	ValueTypeNumber  = eval.ValueTypeNumber
	ValueTypeJSON    = eval.ValueTypeJSON
)

const (
	ContextTypeString    = eval.ContextTypeString
	ContextTypeInt       = eval.ContextTypeInt
	ContextTypeDouble    = eval.ContextTypeDouble
	ContextTypeBool      = eval.ContextTypeBool
	ContextTypeTimestamp = eval.ContextTypeTimestamp
	ContextTypeList      = eval.ContextTypeList
	ContextTypeMap       = eval.ContextTypeMap
)
