package core

import "github.com/picunada/flagcel/evalcore"

type ContextSchema = evalcore.ContextSchema
type ContextField = evalcore.ContextField
type ContextType = evalcore.ContextType

type ContextFieldReference struct {
	Path      string
	RuleCount int
}

type ContextReference struct {
	ContextID        string
	EnvironmentID    string
	EnvironmentKey   string
	FlagKey          string
	RuleCount        int
	ReferencedFields []ContextFieldReference
	Flag             FlagConfig
}

type ContextSchemaConflictError struct {
	Issues []ValidationIssue
}

func (e *ContextSchemaConflictError) Error() string {
	return "Context schema changes would invalidate referencing flags"
}

const (
	ContextTypeString    = evalcore.ContextTypeString
	ContextTypeInt       = evalcore.ContextTypeInt
	ContextTypeDouble    = evalcore.ContextTypeDouble
	ContextTypeBool      = evalcore.ContextTypeBool
	ContextTypeTimestamp = evalcore.ContextTypeTimestamp
	ContextTypeList      = evalcore.ContextTypeList
	ContextTypeMap       = evalcore.ContextTypeMap
)
