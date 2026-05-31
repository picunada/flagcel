package eval

import "time"

type FlagConfig struct {
	Key          string    `json:"key"`
	Description  string    `json:"description,omitempty"`
	Type         ValueType `json:"type"`
	Enabled      bool      `json:"enabled"`
	Rules        []Rule    `json:"rules"`
	DefaultValue any       `json:"default_value"`
	ContextID    *string   `json:"context_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedBy    *string   `json:"created_by,omitempty"`
	DeletedBy    *string   `json:"deleted_by,omitempty"`
}

type ValueType string

const (
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeString  ValueType = "string"
	ValueTypeNumber  ValueType = "number"
	ValueTypeJSON    ValueType = "json"
)

type FlagValue struct {
	Type  ValueType `json:"value_type"`
	Value any       `json:"value"`
}

type Rule struct {
	ID          string    `json:"id"`
	Description string    `json:"description,omitempty"`
	Expression  string    `json:"expression"`
	Rollout     Rollout   `json:"rollout"`
	Value       any       `json:"value"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	DeletedBy   *string   `json:"deleted_by,omitempty"`
}

type Rollout struct {
	Percentage int    `json:"percentage"`
	BucketBy   string `json:"bucket_by,omitempty"`
}

type EvalContext struct {
	User    map[string]any `json:"user"`
	Request map[string]any `json:"request"`
	Time    *time.Time     `json:"time"`
}

type ContextSchema struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Fields      []ContextField `json:"fields"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedBy   *string        `json:"created_by,omitempty"`
	DeletedBy   *string        `json:"deleted_by,omitempty"`
}

type ContextField struct {
	Path string      `json:"path"`
	Type ContextType `json:"type"`
}

type ContextType string

const (
	ContextTypeString    ContextType = "string"
	ContextTypeInt       ContextType = "int"
	ContextTypeDouble    ContextType = "double"
	ContextTypeBool      ContextType = "bool"
	ContextTypeTimestamp ContextType = "timestamp"
	ContextTypeList      ContextType = "list"
	ContextTypeMap       ContextType = "map"
)

func (t ContextType) Valid() bool {
	switch t {
	case ContextTypeString, ContextTypeInt, ContextTypeDouble,
		ContextTypeBool, ContextTypeTimestamp, ContextTypeList, ContextTypeMap:
		return true
	}
	return false
}
