package core

import "time"

type FlagConfig struct {
	Key          string    `json:"key"`
	Enabled      bool      `json:"enabled"`
	Rules        []Rule    `json:"rules"`
	DefaultValue bool      `json:"default_value"`
	ContextID    *string   `json:"context_id,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Rule struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Rollout    Rollout `json:"rollout"`
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
