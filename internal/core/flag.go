package core

import "github.com/picunada/flagcel/evalcore"

type FlagConfig = evalcore.FlagConfig
type ValueType = evalcore.ValueType
type FlagValue = evalcore.FlagValue
type Rule = evalcore.Rule
type Rollout = evalcore.Rollout
type EvalContext = evalcore.EvalContext

const (
	ValueTypeBoolean = evalcore.ValueTypeBoolean
	ValueTypeString  = evalcore.ValueTypeString
	ValueTypeNumber  = evalcore.ValueTypeNumber
	ValueTypeJSON    = evalcore.ValueTypeJSON
)
