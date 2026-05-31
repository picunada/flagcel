package core

import "github.com/picunada/flagcel/evalcore"

type ContextSchema = evalcore.ContextSchema
type ContextField = evalcore.ContextField
type ContextType = evalcore.ContextType

const (
	ContextTypeString    = evalcore.ContextTypeString
	ContextTypeInt       = evalcore.ContextTypeInt
	ContextTypeDouble    = evalcore.ContextTypeDouble
	ContextTypeBool      = evalcore.ContextTypeBool
	ContextTypeTimestamp = evalcore.ContextTypeTimestamp
	ContextTypeList      = evalcore.ContextTypeList
	ContextTypeMap       = evalcore.ContextTypeMap
)
