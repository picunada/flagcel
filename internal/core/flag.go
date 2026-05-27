package core

import "time"

type FlagConfig struct {
	Key          string    `json:"key"`
	Type         ValueType `json:"type"`
	Enabled      bool      `json:"enabled"`
	Rules        []Rule    `json:"rules"`
	DefaultValue any       `json:"default_value"`
	ContextID    *string   `json:"context_id,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Rollout    Rollout `json:"rollout"`
	Value      any     `json:"value"`
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
